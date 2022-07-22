/*
Copyright 2022 The KubeZoo Authors.

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

package controller

import (
	"context"
	"net"
	"os"
	"path"
	"strconv"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/cert"
	"sigs.k8s.io/controller-runtime/pkg/envtest"

	tenantv1alpha1 "github.com/kubewharf/kubezoo/pkg/apis/tenant/v1alpha1"
	tenantclientset "github.com/kubewharf/kubezoo/pkg/generated/clientset/versioned"
	tenantinformer "github.com/kubewharf/kubezoo/pkg/generated/informers/externalversions/tenant/v1alpha1"
)

var (
	controlPlaneTestEnv *envtest.Environment
	controlPlaneCfg     *rest.Config
	upstreamCfg         *rest.Config
	controlPlaneClient  tenantclientset.Interface
	upstreamClient      kubernetes.Interface
	ctx                 context.Context
	cancel              context.CancelFunc

	testTenantName = "kubezoo-controller-test"
)

func TestTenantController(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "tenant controller suite")
}

// Setup the kubezoo controller.
var _ = BeforeSuite(func() {
	ctx, cancel = context.WithCancel(context.TODO())
	By("bootstrapping test environment")
	var err error

	// create control plane env
	controlPlaneTestEnv = &envtest.Environment{
		CRDs: []runtime.Object{tenantCRD},
	}

	// start control plane
	controlPlaneCfg, err = controlPlaneTestEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(controlPlaneCfg).NotTo(BeNil())

	err = tenantv1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	// create control plane client
	controlPlaneClient, err = tenantclientset.NewForConfig(controlPlaneCfg)
	Expect(err).NotTo(HaveOccurred())
	Expect(controlPlaneClient).NotTo(BeNil())

	// set up upstream cluster env
	upstreamCfg = controlPlaneCfg
	// create upstream client
	upstreamClient, err = kubernetes.NewForConfig(upstreamCfg)
	Expect(err).NotTo(HaveOccurred())
	Expect(upstreamClient).NotTo(BeNil())

	// generate client ca key and cert
	tempDir, err := os.MkdirTemp("", "kubezoo")
	Expect(err).NotTo(HaveOccurred())
	Expect(tempDir).NotTo(BeEmpty())

	// generate client ca key and cert for upstream cluster
	host, port, err := net.SplitHostPort(upstreamCfg.Host)
	Expect(err).NotTo(HaveOccurred())
	portInt, err := strconv.Atoi(port)
	Expect(err).NotTo(HaveOccurred())
	_, _, err = cert.GenerateSelfSignedCertKeyWithFixtures(host, nil, nil, tempDir)
	Expect(err).NotTo(HaveOccurred())
	clientCAKey := path.Join(tempDir, host+"__.key")
	clientCACert := path.Join(tempDir, host+"__.crt")

	informer := tenantinformer.NewTenantInformer(controlPlaneClient, 0, cache.Indexers{})

	go func() {
		defer GinkgoRecover()
		Run(ctx.Done(),
			informer,
			controlPlaneClient.TenantV1alpha1(),
			upstreamClient,
			clientCACert,
			clientCAKey,
			host,
			portInt,
		)
	}()
}, 60)

// Tearing down the test environment
var _ = AfterSuite(func() {
	cancel()
	By("tearing down the test environment")
	err := controlPlaneTestEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})

// use v1beta1 to bypass the openapi schema check
var tenantCRD = &apiextensionsv1beta1.CustomResourceDefinition{
	ObjectMeta: metav1.ObjectMeta{
		Name: "tenants." + tenantv1alpha1.SchemeGroupVersion.Group,
	},
	Spec: apiextensionsv1beta1.CustomResourceDefinitionSpec{
		Group: tenantv1alpha1.SchemeGroupVersion.Group,
		Names: apiextensionsv1beta1.CustomResourceDefinitionNames{
			Kind:     "Tenant",
			ListKind: "TenantList",
			Plural:   "tenants",
			Singular: "tenant",
		},
		Scope: apiextensionsv1beta1.ClusterScoped,
		Versions: []apiextensionsv1beta1.CustomResourceDefinitionVersion{
			{
				Name:    "v1alpha1",
				Served:  true,
				Storage: true,
			},
		},
	},
}
