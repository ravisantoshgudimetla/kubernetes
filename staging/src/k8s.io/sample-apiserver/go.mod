// This is a generated file. Do not edit directly.

module k8s.io/sample-apiserver

go 1.16

require (
	github.com/dustin/go-humanize v1.0.0 // indirect
	github.com/go-openapi/spec v0.19.3
	github.com/google/gofuzz v1.1.0
	github.com/spf13/cobra v1.1.1
	go.etcd.io/bbolt v1.3.5 // indirect
	google.golang.org/genproto v0.0.0-20201110150050-8816d57aaa9a // indirect
	k8s.io/apimachinery v0.20.0
	k8s.io/apiserver v0.20.0
	k8s.io/client-go v0.20.0
	k8s.io/code-generator v0.20.0
	k8s.io/component-base v0.20.0
	k8s.io/klog/v2 v2.4.0
	k8s.io/kube-openapi v0.0.0-20210305001622-591a79e4bda7
)

replace (
	github.com/go-bindata/go-bindata => github.com/go-bindata/go-bindata v3.1.1+incompatible
	github.com/imdario/mergo => github.com/imdario/mergo v0.3.5
	github.com/mattn/go-colorable => github.com/mattn/go-colorable v0.0.9
	github.com/onsi/ginkgo => github.com/openshift/ginkgo v4.5.0-origin.1+incompatible
	github.com/robfig/cron => github.com/robfig/cron v1.1.0
	go.uber.org/multierr => go.uber.org/multierr v1.1.0
	gopkg.in/yaml.v2 => gopkg.in/yaml.v2 v2.2.8
	k8s.io/api => ../api
	k8s.io/apiextensions-apiserver => ../apiextensions-apiserver
	k8s.io/apimachinery => ../apimachinery
	k8s.io/apiserver => ../apiserver
	k8s.io/client-go => ../client-go
	k8s.io/code-generator => ../code-generator
	k8s.io/component-base => ../component-base
	k8s.io/kube-aggregator => ../kube-aggregator
	k8s.io/sample-apiserver => ../sample-apiserver
)
