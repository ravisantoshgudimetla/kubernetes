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

package informers

import (
	"reflect"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
	extensions "k8s.io/kubernetes/pkg/apis/extensions/v1beta1"
	"k8s.io/kubernetes/pkg/client/legacylisters"
	"time"
)


// DaemonSetInformer is type of SharedIndexInformer which watches and lists all pods.
// Interface provides constructor for informer and lister for pods
type DaemonSetInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() *listers.StoreToDaemonSetLister
}

type daemonSetInformer struct {
	*sharedInformerFactory
}

func getCacheType(f *sharedInformerFactory, informerType reflect.Type)  (cache.ListFunc, cache.WatchFunc, interface{}, time.Duration){
	switch informerType{
	case &extensions.DaemonSet{}:
		return func(options metav1.ListOptions) (runtime.Object, error) {
			return f.client.Extensions().DaemonSets(metav1.NamespaceAll).List(options)
		}, func(options metav1.ListOptions) (watch.Interface, error) {
			return f.client.Extensions().DaemonSets(metav1.NamespaceAll).Watch(options)

		}, &extensions.DaemonSet{}, f.defaultResync
	case &extensions.Deployment{}:
		return  func(options metav1.ListOptions) (runtime.Object, error) {
				return f.client.Extensions().Deployments(metav1.NamespaceAll).List(options)
		}, func(options metav1.ListOptions) (watch.Interface, error) {
				return f.client.Extensions().ReplicaSets(metav1.NamespaceAll).Watch(options)
		}, &extensions.ReplicaSet{}, f.defaultResync
	}
	return
}

func getInformer(f *sharedInformerFactory, informerType reflect.Type) cache.SharedIndexInformer {
	f.lock.Lock()
	defer f.lock.Unlock()
	informer, exists := f.informers[informerType]
	if exists {
		return informer
	}
	listfun, watchfunc, objtype, resync := getCacheType(f, informerType)
	informer = cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: listfun,
			WatchFunc: watchfunc,
		},
		objtype,
		resync,
		cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc},
	)
	f.informers[informerType] = informer
	return informer
}

func (f *daemonSetInformer) Informer() cache.SharedIndexInformer {
	informerType := reflect.TypeOf(&extensions.DaemonSet{})
	informer := getInformer(f, informerType)
	return informer
}

func (f *daemonSetInformer) Lister() *listers.StoreToDaemonSetLister {
	informer := f.Informer()
	return &listers.StoreToDaemonSetLister{Store: informer.GetIndexer()}
}

// DeploymentInformer is a type of SharedIndexInformer which watches and lists all deployments.
type DeploymentInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() *listers.StoreToDeploymentLister
}

type deploymentInformer struct {
	*sharedInformerFactory
}

func (f *deploymentInformer) Informer() cache.SharedIndexInformer {
	informerType := reflect.TypeOf(&extensions.Deployment{})
	informer := getInformer(f, informerType)
	return informer
}

func (f *deploymentInformer) Lister() *listers.StoreToDeploymentLister {
	informer := f.Informer()
	return &listers.StoreToDeploymentLister{Indexer: informer.GetIndexer()}
}

// ReplicaSetInformer is a type of SharedIndexInformer which watches and lists all replicasets.
type ReplicaSetInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() *listers.StoreToReplicaSetLister
}

type replicaSetInformer struct {
	*sharedInformerFactory
}

func (f *replicaSetInformer) Informer() cache.SharedIndexInformer {
	informerType := reflect.TypeOf(&extensions.ReplicaSet{})
	informer := getInformer(f, informerType)
	return informer
}

func (f *replicaSetInformer) Lister() *listers.StoreToReplicaSetLister {
	informer := f.Informer()
	return &listers.StoreToReplicaSetLister{Indexer: informer.GetIndexer()}
}
