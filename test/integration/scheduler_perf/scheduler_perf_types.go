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

package benchmark

// High Level Configuration for all predicates and priorities.
type schedulerPerfConfig struct {
	NodeAffinity *nodeAffinity
	TaintsAndTolerations *taintsAndTolerations
}

// nodeAffinity priority configuration details.
type nodeAffinity struct {
	numGroups       int    // number of Node-Pod sets with Pods NodeAffinity matching given Nodes.
	nodeAffinityKey string // Node Selection Key.
}


// nodeAffinity priority configuration details.
type taintsAndTolerations struct {
	taintsCount int // number of taints to be added to each node.
	tolerationsCount int // number of tolerations to be added to each pod.
}


