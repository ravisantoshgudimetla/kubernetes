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

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apiserver/pkg/admission"
	"k8s.io/kubernetes/pkg/api"
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
	priorityMap map[string]*int32
}

// ComputePrior creates a new instance of the LimitPodHardAntiAffinityTopology admission controller
func ComputePriority() admission.Interface {
	return &priorityConfig{
		Handler: admission.NewHandler(admission.Create, admission.Update),
		priorityMap: initializePriorities(),
	}
}

func initializePriorities() map[string]*int32{
	var priorityMap = make(map[string]*int32)
	// Have to create a var for every priority as golang doesn't allow to take address of numeric constants.
	systemPriority := int32(100000)
	priorityMap["system"] = &systemPriority
	return priorityMap
}

// Admit will populate the priority based on the PriorityClass field.
func (p *priorityConfig) Admit(attributes admission.Attributes) (error) {
	// Ignore all calls to subresources or resources other than pods.
	if len(attributes.GetSubresource()) != 0 || attributes.GetResource().GroupResource() != api.Resource("pods") {
		return nil
	}
	pod, ok := attributes.GetObject().(*api.Pod)
	if !ok {
		return apierrors.NewBadRequest("Resource was marked with kind Pod but was unable to be converted")
	}
	priorityClass := pod.Spec.PriorityClassName
	pod.Spec.Priority = p.priorityMap[priorityClass]
	return nil
}
