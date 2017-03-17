/*
Copyright 2016 The Kubernetes Authors.

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

package nodetaint

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/admission"
	"k8s.io/kubernetes/pkg/api"
)

// ensures the hard PodAntiAffinity is denied if it defines TopologyKey other than kubernetes.io/hostname.
func TestOtherResources(t *testing.T) {
	namespace := "testnamespace"
	name := "testname"
	node := &api.Node{
		ObjectMeta: metav1.ObjectMeta{Name: name},
	}
	tests := []struct {
		name        string
		kind        string
		resource    string
		subresource string
		object      runtime.Object
		expectError bool
	}{
		{
			name:     "non-pod resource",
			kind:     "Foo",
			resource: "foos",
			object:   node,
		},
		{
			name:        "pod subresource",
			kind:        "Node",
			resource:    "nodes",
			subresource: "eviction",
			object:      node,
		},
		{
			name:        "non-pod object",
			kind:        "Node",
			resource:    "pods",
			object:      node,
			expectError: false,
		},
	}

	for _, tc := range tests {
		handler := &plugin{}

		err := handler.Admit(admission.NewAttributesRecord(tc.object, nil, api.Kind(tc.kind).WithVersion("version"), namespace, name, api.Resource(tc.resource).WithVersion("version"), tc.subresource, admission.Create, nil))

		if tc.expectError {
			if err == nil {
				t.Errorf("%s: unexpected nil error", tc.name)
			}
			continue
		}

		if err != nil {
			t.Errorf("%s: unexpected error: %v", tc.name, err)
			continue
		}
	}
}
