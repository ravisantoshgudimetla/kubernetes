/*
Copyright 2017 The Kubernetes Authors.

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
	"encoding/json"
	"fmt"
	"io"

	"github.com/golang/glog"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apiserver/pkg/admission"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/v1"
	clientset "k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset"
	informers "k8s.io/kubernetes/pkg/client/informers/informers_generated/internalversion"
	"k8s.io/kubernetes/pkg/util/tolerations"
	pluginapi "k8s.io/kubernetes/plugin/pkg/admission/podtolerationrestriction/apis/podtolerationrestriction"
)

// Register registers a plugin
func Register(plugins *admission.Plugins) {
	plugins.Register("PodTolerationRestriction", func(config io.Reader) (admission.Interface, error) {
		pluginConfig, err := loadConfiguration(config)
		if err != nil {
			return nil, err
		}
		return NewPodTolerationsPlugin(pluginConfig), nil
	})
}

type podTolerationsPlugin struct {
	*admission.Handler
}

// This plugin first verifies any conflict between a pod's tolerations and
// its namespace's tolerations, and rejects the pod if there's a conflict.
// If there's no conflict, the pod's tolerations are merged with its namespace's
// toleration. Resulting pod's tolerations are verified against its namespace's
// whitelist of tolerations. If the verification is successful, the pod is admitted
// otherwise rejected. If a namespace does not have associated default or whitelist
// of tolerations, then cluster level default or whitelist of tolerations are used
// instead if specified. Tolerations to a namespace are assigned via
// scheduler.alpha.kubernetes.io/defaultTolerations and scheduler.alpha.kubernetes.io/tolerationsWhitelist
// annotations keys.
func (p *podTolerationsPlugin) Admit(a admission.Attributes) error {
	resource := a.GetResource().GroupResource()
	if resource != api.Resource("pods") {
		return nil
	}
	if a.GetSubresource() != "" {
		// only run the checks below on pods proper and not subresources
		return nil
	}

	obj := a.GetObject()
	pod, ok := obj.(*api.Pod)
	if !ok {
		glog.Errorf("expected pod but got %s", a.GetKind().Kind)
		return nil
	}

	if !p.WaitForReady() {
		return admission.NewForbidden(a, fmt.Errorf("not yet ready to handle request"))
	}


	return nil

}

func NewPodTolerationsPlugin(pluginConfig *pluginapi.Configuration) *podTolerationsPlugin {
	return &podTolerationsPlugin{
		Handler:      admission.NewHandler(admission.Create, admission.Update),
	}
}


func (p *podTolerationsPlugin) Validate() error {
	if p.namespaceLister == nil {
		return fmt.Errorf("missing namespaceLister")
	}
	if p.client == nil {
		return fmt.Errorf("missing client")
	}
	return nil
}


