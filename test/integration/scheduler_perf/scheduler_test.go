/*
Copyright 2015 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package benchmark

import (
	"github.com/golang/glog"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/kubernetes/pkg/api/v1"
	"k8s.io/kubernetes/pkg/client/clientset_generated/clientset"
	"k8s.io/kubernetes/plugin/pkg/scheduler"
	testutils "k8s.io/kubernetes/test/utils"
	"testing"
	"time"
	"fmt"
	"sort"
)


var (
	basePodTemplate = &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "sched-perf-pod-",
		},
		// TODO: this needs to be configurable.
		Spec: testutils.MakePodSpec(),
	}
	baseNodeTemplate = &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "sample-node-",
		},
		Spec: v1.NodeSpec{
			// TODO: investigate why this is needed.
			ExternalID: "foo",
		},
		Status: v1.NodeStatus{
			Capacity: v1.ResourceList{
				v1.ResourcePods:   *resource.NewQuantity(110, resource.DecimalSI),
				v1.ResourceCPU:    resource.MustParse("4"),
				v1.ResourceMemory: resource.MustParse("32Gi"),
			},
			Phase: v1.NodeRunning,
			Conditions: []v1.NodeCondition{
				{Type: v1.NodeReady, Status: v1.ConditionTrue},
			},
		},
	}
)
// TestSchedule100Node3KPods schedules 3k pods on 100 nodes.
func TestSchedule100Node3KPods(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping because we want to run short tests")
	}
	// As of now, getting entries.
	// TODO: Create a struct to hold file related values.
	/*if err := getEntriesInFile(); err != nil {
		t.Errorf("Error counting entries in file")
	}*/
	config := getBaseConfig(100)
	//writePodAndNodeTopologyToConfig(config)
	if err := readFromFile(); err != nil {
		t.Errorf("Error reading from file")
	}
	schedulePods(config)
}

// testConfig contains the some input parameters needed for running test-suite
type testConfig struct {
	// Note: We don't need numPods, numNodes anymore in this struct but keeping them for backward compatibility
	numNodes                  int
	nodePreparer              testutils.TestNodePreparer
	schedulerSupportFunctions scheduler.Configurator
	destroyFunc               func()
}

// getBaseConfig returns baseConfig after initializing number of nodes and pods.
// We have to function for backward compatibility. We can combine this into baseConfig.
// TODO: Remove this function once the backward compatibility is not needed.
func getBaseConfig(nodes int) *testConfig {
	schedulerConfigFactory, destroyFunc := mustSetupScheduler()
	return &testConfig{
		schedulerSupportFunctions: schedulerConfigFactory,
		destroyFunc:               destroyFunc,
		numNodes:                  nodes,
	}
}

func createPod(cs clientset.Interface, pod podInfo) (*v1.Pod, error) {
	podToBeCreated := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: pod.podName,
			Name: pod.podName,
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{{
				Name:  "pause",
				Image: "kubernetes/pause",
				Ports: []v1.ContainerPort{{ContainerPort: 80}},
				Resources: v1.ResourceRequirements{
					Limits: v1.ResourceList{
						v1.ResourceCPU:    resource.MustParse(pod.cpu),
						v1.ResourceMemory: resource.MustParse(pod.memory),
					},
					Requests: v1.ResourceList{
						v1.ResourceCPU:    resource.MustParse(pod.cpu),
						v1.ResourceMemory: resource.MustParse(pod.memory),
					},
				},
			}},
		},
	}
	return cs.Core().Pods("test-").Create(podToBeCreated)

}

func deletePod(cs clientset.Interface, pod podInfo) {
	// test- is the namespace name.
	options := &metav1.DeleteOptions{}
	cs.Core().Pods("test-").Delete(pod.podName, options)
}


func generateNodes(config *testConfig) {
	for i := 0; i < config.numNodes; i++ {
		config.schedulerSupportFunctions.GetClient().Core().Nodes().Create(baseNodeTemplate)

	}
	return
}

func checkInScheduled(podCreated *v1.Pod, scheduledPodList []*v1.Pod) bool{
	for _, scheduledPod:= range scheduledPodList {
		if scheduledPod.ObjectMeta.Name == podCreated.ObjectMeta.Name {
			return true
		}
	}
	return false
}

// schedulePods schedules specific number of pods on specific number of nodes.
// This is used to learn the scheduling throughput on various
// sizes of cluster and changes as more and more pods are scheduled.
// It won't stop until all pods are scheduled.
// It returns the minimum of throughput over whole run.
func schedulePods(config *testConfig) {
	defer config.destroyFunc()
	generateNodes(config)
	cs := config.schedulerSupportFunctions.GetClient()
	sort.Float64s(timeList)
	for _, timeNow := range timeList {
		// Checking if the value is not equal to 0.
		if timeNow != float64(0) {
			podsToDelete := getPodsToDelete(timeNow)
			podsToCreate := getPodsToCreate(timeNow)
			// Both pod deletion and creation could be parallelized(individually).
			// First delete all the pods that are not needed.
			for _, pod := range podsToDelete {
				deletePod(cs, pod)
			}
			var podList []*v1.Pod
			// Create all the pods needed for this iteration.
			for _, pod := range podsToCreate {
				// If there is no error for pod creation add it to
				// pod list.
				pod, err := createPod(cs, pod)
				if err != nil {
					podList = append(podList, pod)
				}
			}
			for {
				start := time.Now()
				scheduled, err := config.schedulerSupportFunctions.GetScheduledPodLister().List(labels.Everything())
				if err != nil {
					glog.Fatalf("%v", err)
				}
				allPodsScheduled := true
				// Check if all the pods got scheduled.
				// TODO: This is an expensive operation. Need to work around this.
				for _, podCreated := range scheduled {
					allPodsScheduled = allPodsScheduled && checkInScheduled(podCreated, scheduled)
				}
				// If yes, break this loop.
				if allPodsScheduled {
					fmt.Printf("At: %f, PodsCreated: %d, PodsDelete: %d, TimeTaken: %v\n", timeNow, len(podsToCreate), len(podsToDelete), time.Since(start))
					break
				}


			}
		}
	}
	return
}
