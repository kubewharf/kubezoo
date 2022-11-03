module github.com/kubewharf/kubezoo

go 1.14

require (
	github.com/davecgh/go-spew v1.1.1
	github.com/go-openapi/spec v0.19.5
	github.com/go-test/deep v1.0.8
	github.com/gogo/protobuf v1.3.2
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/gnostic v0.5.7-v3refs
	github.com/onsi/ginkgo v1.16.1
	github.com/onsi/gomega v1.11.0
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v1.4.0
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.7.0
	k8s.io/api v0.24.0
	k8s.io/apiextensions-apiserver v0.24.0
	k8s.io/apimachinery v0.24.4-rc.0
	k8s.io/apiserver v0.24.0
	k8s.io/client-go v0.24.0
	k8s.io/cluster-bootstrap v0.24.0 // indirect
	k8s.io/component-base v0.24.0
	k8s.io/component-helpers v0.24.0
	k8s.io/klog v1.0.0
	k8s.io/kube-aggregator v0.24.0
	k8s.io/kube-openapi v0.0.0-20220328201542-3ee0da9b0b42
	k8s.io/kubelet v0.24.0 // indirect
	k8s.io/kubernetes v1.24.0
	k8s.io/legacy-cloud-providers v0.24.0 // indirect
	k8s.io/mount-utils v0.24.4-rc.0 // indirect
	k8s.io/pod-security-admission v0.24.0 // indirect
	k8s.io/utils v0.0.0-20220210201930-3a6ce19ff2f9
	sigs.k8s.io/apiserver-runtime v1.0.2
	sigs.k8s.io/controller-runtime v0.6.0
	sigs.k8s.io/structured-merge-diff/v4 v4.2.1
)

replace (
	github.com/gregjones/httpcache => github.com/gregjones/httpcache v0.0.0-20180305231024-9cad4c3443a7
	github.com/imdario/mergo => github.com/imdario/mergo v0.3.5
	github.com/kisielk/errcheck => github.com/kisielk/errcheck v1.5.0
	github.com/mohae/deepcopy => github.com/mohae/deepcopy v0.0.0-20170603005431-491d3605edfb
	github.com/olekukonko/tablewriter => github.com/olekukonko/tablewriter v0.0.4
	github.com/rogpeppe/go-internal => github.com/rogpeppe/go-internal v1.3.0
	gopkg.in/gcfg.v1 => gopkg.in/gcfg.v1 v1.2.0
	gopkg.in/warnings.v0 => gopkg.in/warnings.v0 v0.1.1
	honnef.co/go/tools => honnef.co/go/tools v0.0.1-2020.1.4
	k8s.io/api => k8s.io/api v0.24.0
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.24.0
	k8s.io/apimachinery => k8s.io/apimachinery v0.24.4-rc.0
	k8s.io/apiserver => k8s.io/apiserver v0.24.0
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.24.0
	k8s.io/client-go => k8s.io/client-go v0.24.0
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.24.0
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.24.0
	k8s.io/code-generator => k8s.io/code-generator v0.24.4-rc.0
	k8s.io/component-base => k8s.io/component-base v0.24.0
	k8s.io/component-helpers => k8s.io/component-helpers v0.24.0
	k8s.io/controller-manager => k8s.io/controller-manager v0.24.0
	k8s.io/cri-api => k8s.io/cri-api v0.25.0-alpha.0
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.24.0
	k8s.io/gengo => k8s.io/gengo v0.0.0-20211129171323-c02415ce4185
	k8s.io/klog => k8s.io/klog v1.0.0
	k8s.io/klog/v2 => k8s.io/klog/v2 v2.60.1
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.24.0
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.24.0
	k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20220328201542-3ee0da9b0b42
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.24.0
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.24.0
	k8s.io/kubectl => k8s.io/kubectl v0.24.0
	k8s.io/kubelet => k8s.io/kubelet v0.24.0
	k8s.io/kubernetes => k8s.io/kubernetes v1.24.0
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.24.0
	k8s.io/metrics => k8s.io/metrics v0.24.0
	k8s.io/mount-utils => k8s.io/mount-utils v0.24.4-rc.0
	k8s.io/pod-security-admission => k8s.io/pod-security-admission v0.24.0
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.24.0
	k8s.io/system-validators => k8s.io/system-validators v1.7.0
	k8s.io/utils => k8s.io/utils v0.0.0-20220210201930-3a6ce19ff2f9
)
