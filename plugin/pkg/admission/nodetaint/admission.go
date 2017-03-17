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

package nodetaint

import (
	"io"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apiserver/pkg/admission"
	"k8s.io/kubernetes/pkg/api"
	corelisters "k8s.io/kubernetes/pkg/client/listers/core/internalversion"
)

func init() {
	admission.RegisterPlugin("RegisterTaintedNodeByDefault", func(config io.Reader) (admission.Interface, error) {
		return RegisterTaintedNodeOnly(), nil
	})
}

// plugin contains the client used by the admission controller
type plugin struct {
	*admission.Handler
	nodeLister corelisters.NodeLister
}

// RegisterTaintedNodeOnly creates a new instance of the RegisterTaintedNodeByDefault admission controller
func RegisterTaintedNodeOnly() admission.Interface {
	return &plugin{
		Handler: admission.NewHandler(admission.Create, admission.Update),
	}
}

// Admit will deny any node that is not having atleast one taint by default.
func (p *plugin) Admit(attributes admission.Attributes) error {
	// Ignore all calls to subresources or resources other than nodes.
	if len(attributes.GetSubresource()) != 0 || attributes.GetResource().GroupResource() != api.Resource("nodes") {
		return nil
	}
	node, ok := attributes.GetObject().(*api.Node)
	if !ok{
		return apierrors.NewBadRequest("Resource was marked with kind Node but was unable to be converted")
	}
	nodepod, err := p.nodeLister.Get(node.Name)
	if err != nil {
		if len(node.Spec.Taints) == 0  {
			return apierrors.NewBadRequest("Node is supposed to have atleast one taint by default to separate")}
	}
	if len(nodepod.Spec.Taints) == 0 {
		return apierrors.NewBadRequest("Node is supposed to have atleast one taint by default to separate")
	}
	return nil
}

