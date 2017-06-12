/*
Copyright 2014 The Kubernetes Authors.

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

package priority

import (
	"testing"
	"k8s.io/kubernetes/pkg/api"
	admission "k8s.io/apiserver/pkg/admission"
	"fmt"
)

func TestPriorityMapping(t *testing.T) {
	handler := ComputePriority()
	expectedSystemPriority :=  int32(100000)
	expectedDefaultPriority :=  int32(0)
	tests := []struct {
		description  string
		requestedPod api.Pod
		expectedPod  api.Pod
	}{
		{
			description: "pod has a priorityClassName set with system, so it should have highest priority",
			requestedPod: api.Pod{
				Spec: api.PodSpec{
					PriorityClassName: "system",
				},
			},
			expectedPod: api.Pod{
				Spec: api.PodSpec{
					PriorityClassName: "system",
					Priority: &expectedSystemPriority,
				},
			},
		},
		{
			description: "pod has a priorityClassName set with system, so it should have highest priority",
			requestedPod: api.Pod{
				Spec: api.PodSpec{
					PriorityClassName: "SYSTEM",
				},
			},
			expectedPod: api.Pod{
				Spec: api.PodSpec{
					PriorityClassName: "SYSTEM",
					Priority: &expectedSystemPriority,
				},
			},
		},
		{
			description: "pod has a priorityClassName which is not present in default allowable , so it should have zero priority",
			requestedPod: api.Pod{
				Spec: api.PodSpec{
					PriorityClassName: "SYSTEM1",
				},
			},
			expectedPod: api.Pod{
				Spec: api.PodSpec{
					PriorityClassName: "SYSTEM1",
					Priority: &expectedDefaultPriority,
				},
			},
		},
		{
			description: "pod has no priorityClassName set, so it will have zero priority",
			requestedPod: api.Pod{
			},
			expectedPod: api.Pod{
				Spec: api.PodSpec{
					Priority: &expectedDefaultPriority,
				},
			},
		},

	}
	for _, test := range tests {
		err := handler.Admit(admission.NewAttributesRecord(&test.requestedPod, nil, api.Kind("Pod").WithVersion("version"), "foo", "name", api.Resource("pods").WithVersion("version"), "", "ignored", nil))
		if err != nil {
			t.Errorf("[%s]: unexpected error %v for pod %+v", test.description, err, test.requestedPod)
		}
		if *test.expectedPod.Spec.Priority != *test.requestedPod.Spec.Priority {
			t.Errorf("Didn't expect an error for test %q", test.description)
		}
	}
}

func TestHandles(t *testing.T) {
	handler := ComputePriority()
	tests := map[admission.Operation]bool{
		admission.Update:  true,
		admission.Create:  true,
		admission.Delete:  false,
		admission.Connect: false,
	}
	for op, expected := range tests {
		result := handler.Handles(op)
		if result != expected {
			t.Errorf("Unexpected result for operation %s: %v\n", op, result)
		}
	}
}

