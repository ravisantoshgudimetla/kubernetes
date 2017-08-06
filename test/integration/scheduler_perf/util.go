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
	"net/http"
	"net/http/httptest"
	"bufio"
	"os"
	"strings"
	"strconv"

	"github.com/golang/glog"
	clientv1core "k8s.io/client-go/kubernetes/typed/core/v1"
	clientv1 "k8s.io/client-go/pkg/api/v1"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/v1"
	"k8s.io/kubernetes/pkg/client/clientset_generated/clientset"
	informers "k8s.io/kubernetes/pkg/client/informers/informers_generated/externalversions"
	"k8s.io/kubernetes/plugin/pkg/scheduler"
	_ "k8s.io/kubernetes/plugin/pkg/scheduler/algorithmprovider"
	"k8s.io/kubernetes/plugin/pkg/scheduler/factory"
	"k8s.io/kubernetes/test/integration/framework"

)

// mustSetupScheduler starts the following components:
// - k8s api server (a.k.a. master)
// - scheduler
// It returns scheduler config factory and destroyFunc which should be used to
// remove resources after finished.
// Notes on rate limiter:
//   - client rate limit is set to 5000.
func mustSetupScheduler() (schedulerConfigurator scheduler.Configurator, destroyFunc func()) {

	h := &framework.MasterHolder{Initialized: make(chan struct{})}
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		<-h.Initialized
		h.M.GenericAPIServer.Handler.ServeHTTP(w, req)
	}))

	framework.RunAMasterUsingServer(framework.NewIntegrationTestMasterConfig(), s, h)

	clientSet := clientset.NewForConfigOrDie(&restclient.Config{
		Host:          s.URL,
		ContentConfig: restclient.ContentConfig{GroupVersion: &api.Registry.GroupOrDie(v1.GroupName).GroupVersion},
		QPS:           5000.0,
		Burst:         5000,
	})

	informerFactory := informers.NewSharedInformerFactory(clientSet, 0)

	schedulerConfigurator = factory.NewConfigFactory(
		v1.DefaultSchedulerName,
		clientSet,
		informerFactory.Core().V1().Nodes(),
		informerFactory.Core().V1().Pods(),
		informerFactory.Core().V1().PersistentVolumes(),
		informerFactory.Core().V1().PersistentVolumeClaims(),
		informerFactory.Core().V1().ReplicationControllers(),
		informerFactory.Extensions().V1beta1().ReplicaSets(),
		informerFactory.Apps().V1beta1().StatefulSets(),
		informerFactory.Core().V1().Services(),
		v1.DefaultHardPodAffinitySymmetricWeight,
	)

	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartRecordingToSink(&clientv1core.EventSinkImpl{Interface: clientv1core.New(clientSet.Core().RESTClient()).Events("")})

	sched, err := scheduler.NewFromConfigurator(schedulerConfigurator, func(conf *scheduler.Config) {
		conf.Recorder = eventBroadcaster.NewRecorder(api.Scheme, clientv1.EventSource{Component: "scheduler"})
	})
	if err != nil {
		glog.Fatalf("Error creating scheduler: %v", err)
	}

	stop := make(chan struct{})
	informerFactory.Start(stop)

	sched.Run()

	destroyFunc = func() {
		glog.Infof("destroying")
		sched.StopEverything()
		close(stop)
		s.Close()
		glog.Infof("destroyed")
	}
	return
}



type podInfo struct {
	podName   string
	startTime float64
	endTime   float64
	memory    string
	cpu       string
}


var noOfEntries int64

func getEntriesInFile() error {
	// Replace it with os.getCwd() and append string.
	file, err := os.Open("/home/ravig/Projects/Personal/Golang_Practice/readFileUpdated.txt")
	if err != nil {
		return err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)

	noOfEntries = 1
	for scanner.Scan() {
		noOfEntries++
	}
	return nil
}


// To hold all the pods read from file.
// TODO: As of now, reads a maximum of 3000, need to change it to read from whole values.
var podInfoList = make([]podInfo, 3000)

// To hold all the times available in file.
// TODO: As of now, reads a maximum of 3000, need to change it to read from whole values.
var timeList = make([]float64, 3000)

// The approach seems terrible as we are doing multiple reads of same file in each function. Ideal case would be to read
// once from file and fill all the datastructures we need.
func readFromFile() error {
	// Replace it with os.getCwd() and append string.
	file, err := os.Open("/home/ravig/Projects/Personal/Golang_Practice/readFileUpdated.txt")
	if err != nil {
		return err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)

	count := 0
	cpuDefault := "0m"
	memDefault := "0m"
	for scanner.Scan() {
		tokens := strings.Split(scanner.Text(), ",")
		podInfoList[count].podName = "podName" + strconv.Itoa(count)
		startTime, _ := strconv.ParseFloat(strings.Split(tokens[0], "=")[1],0)
		podInfoList[count].startTime = startTime
		endTime, _ := strconv.ParseFloat(strings.Split(tokens[1], "=")[1],0)
		podInfoList[count].endTime = endTime
		podInfoList[count].memory = strings.Split(tokens[2], "=")[1]
		podInfoList[count].cpu = strings.Split(tokens[3], "=")[1]
		if podInfoList[count].cpu == "" {
			podInfoList[count].cpu = cpuDefault
		}
		if podInfoList[count].memory == "" {
			podInfoList[count].memory = memDefault
		}
		updateTimeList(podInfoList[count].startTime, podInfoList[count].endTime)
		count++
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}

// exists checks if the item exists in the slice.
func exists(time float64) bool {
	for _, v := range timeList {
		if v == time {
			return true
		}
	}
	return false
}

// updateTimeList updates the global timelist. This could be improved further.
func updateTimeList(startTime float64, endTime float64) {
	if !exists(startTime) {
		timeList = append(timeList, startTime)
	}
	if !exists(endTime) {
		timeList = append(timeList, endTime)
	}
	return
}

// getPodsToCreate gets all the pods to create at the given time t1.
func getPodsToCreate(time float64) []podInfo {
	var podsToCreate []podInfo
	for _, v := range podInfoList {
		if v.startTime == time {
			podsToCreate = append(podsToCreate, v)
		}
	}
	return podsToCreate
}

// getPodsToDelete deletes the pods at the given time t2.
func getPodsToDelete(time float64) []podInfo{
	var podsToDelete []podInfo
	for _, v := range podInfoList {
		if v.endTime == time {
			podsToDelete = append(podsToDelete, v)
		}
	}
	return podsToDelete

}

