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

package priority

import (
	"io"
	"strings"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apiserver/pkg/admission"
	"k8s.io/kubernetes/pkg/api"
	//"fmt"
)

// Register registers a plugin
func Register(plugins *admission.Plugins) {
	plugins.Register("priorityConfig", func(config io.Reader) (admission.Interface, error) {
		return ComputePriority(), nil
	})
}

// priorityConfig contains the client used by the admission controller
type priorityConfig struct {
	*admission.Handler
	//TODO: Check if this map should be present here. There should be an option to update and delete keys from this map.
	priorityMap map[string]*int32
}

// ComputePriority creates a new instance of the priorityConfig admission controller
func ComputePriority() admission.Interface {
	return &priorityConfig{
		Handler:     admission.NewHandler(admission.Create, admission.Update),
		priorityMap: initializePriorities(),
	}
}

// initializePriorities initializes the priorities for various priorities.
// TODO: This will be getPriorities once we know where the priorityMap has to be placed.
func initializePriorities() map[string]*int32 {
	var priorityMap = make(map[string]*int32)
	maxInt := 2147483647
	systemKeyword := "system"
	// Have to create a var for every priorityClassName as golang doesn't allow to take address of numeric constants.
	systemPriority := int32(maxInt)
	priorityMap[systemKeyword] = &systemPriority
	return priorityMap
}

// Admit will only admit pods with valid or default priorityClassNames in PriorityClassName field.
func (p *priorityConfig) Admit(attributes admission.Attributes) error {
	// Ignore all calls to subresources or resources other than pods.
	if len(attributes.GetSubresource()) != 0 || attributes.GetResource().GroupResource() != api.Resource("pods") {
		return nil
	}
	//fmt.Println(attributes.)
	pod, ok := attributes.GetObject().(*api.Pod)
	if !ok {
		return apierrors.NewBadRequest("Resource was marked with kind Pod but was unable to be converted")
	}
	priorityClassName := pod.Spec.PriorityClassName
	if len(priorityClassName) == 0 {
		// No priorityClass specified. Set the default priority.
		defaultPriority := int32(0)
		pod.Spec.Priority = &defaultPriority
		return nil
	}
	priorityClass := strings.ToLower(priorityClassName)
	priority, ok := p.priorityMap[priorityClass]
	if !ok {
		return apierrors.NewBadRequest("Pod is specified with a priorityClass that does not exist")
	} else {
		pod.Spec.Priority = priority
	}
	return nil
}
