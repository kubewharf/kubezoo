module github.com/kubewharf/kubezoo

go 1.14

require (
	github.com/davecgh/go-spew v1.1.1
	github.com/go-openapi/spec v0.19.5
	github.com/go-openapi/strfmt v0.19.3
	github.com/go-openapi/validate v0.19.5
	github.com/go-test/deep v1.0.8
	github.com/gogo/protobuf v1.3.2
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.19.0
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v1.3.0
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.7.0
	k8s.io/api v0.21.2
	k8s.io/apiextensions-apiserver v0.18.2
	k8s.io/apimachinery v0.21.2
	k8s.io/apiserver v0.21.2
	k8s.io/client-go v0.21.2
	k8s.io/cloud-provider v0.18.10
	k8s.io/component-base v0.21.2
	k8s.io/klog v1.0.0
	k8s.io/kube-aggregator v0.0.0
	k8s.io/kube-openapi v0.0.0-20210305001622-591a79e4bda7
	k8s.io/kubernetes v0.0.0-00010101000000-000000000000
	k8s.io/utils v0.0.0-20200324210504-a9aa75ae1b89
	sigs.k8s.io/apiserver-runtime v1.0.2
	sigs.k8s.io/controller-runtime v0.6.0
)

replace (
	github.com/googleapis/gax-go/v2 => github.com/googleapis/gax-go/v2 v2.0.4
	google.golang.org/api => google.golang.org/api v0.14.0
	google.golang.org/genproto => google.golang.org/genproto v0.0.0-20190819201941-24fa4b261c55
	google.golang.org/grpc => google.golang.org/grpc v1.26.0
	k8s.io/api => k8s.io/api v0.18.10
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.18.10
	k8s.io/apimachinery => k8s.io/apimachinery v0.18.11-rc.0
	k8s.io/apiserver => k8s.io/apiserver v0.18.10
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.18.10
	k8s.io/client-go => k8s.io/client-go v0.18.10
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.18.10
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.18.10
	k8s.io/code-generator => k8s.io/code-generator v0.18.18-rc.0
	k8s.io/component-base => k8s.io/component-base v0.18.10
	k8s.io/cri-api => k8s.io/cri-api v0.18.18-rc.0
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.18.10
	k8s.io/gengo => k8s.io/gengo v0.0.0-20200114144118-36b2048a9120
	k8s.io/heapster => k8s.io/heapster v1.2.0-beta.1
	k8s.io/klog => k8s.io/klog v1.0.0
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.18.10
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.18.10
	k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20200410145947-61e04a5be9a6
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.18.10
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.18.10
	k8s.io/kubectl => k8s.io/kubectl v0.18.10
	k8s.io/kubelet => k8s.io/kubelet v0.18.10
	k8s.io/kubernetes => k8s.io/kubernetes v1.18.10
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.18.10
	k8s.io/metrics => k8s.io/metrics v0.18.10
	k8s.io/repo-infra => k8s.io/repo-infra v0.0.1-alpha.1
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.18.10
	k8s.io/system-validators => k8s.io/system-validators v1.0.4
	k8s.io/utils => k8s.io/utils v0.0.0-20200324210504-a9aa75ae1b89
	sigs.k8s.io/controller-runtime => sigs.k8s.io/controller-runtime v0.6.0
)
