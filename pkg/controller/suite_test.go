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
	"strings"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	testTenantName         = "kubezoo-controller-test"
	xPreserveUnknownFields = true
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
		CRDs: []*apiextensionsv1.CustomResourceDefinition{tenantCRD},
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

	// trim the prefix "https://" and the suffix "/"
	// because the Host in a kubeconfig is like "https://127.0.0.1:6443/"
	// and the parameter of func net.SplitHostPort() should be a domain name, an IPv4 or IPv6 address with port only, not a URL.
	upstreamCfg.Host = strings.TrimPrefix(upstreamCfg.Host, "https://")
	upstreamCfg.Host = strings.TrimSuffix(upstreamCfg.Host, "/")

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

var tenantCRD = &apiextensionsv1.CustomResourceDefinition{
	ObjectMeta: metav1.ObjectMeta{
		Name: "tenants." + tenantv1alpha1.SchemeGroupVersion.Group,
	},
	Spec: apiextensionsv1.CustomResourceDefinitionSpec{
		Group: tenantv1alpha1.SchemeGroupVersion.Group,
		Names: apiextensionsv1.CustomResourceDefinitionNames{
			Kind:     "Tenant",
			ListKind: "TenantList",
			Plural:   "tenants",
			Singular: "tenant",
		},
		Scope: apiextensionsv1.ClusterScoped,
		Versions: []apiextensionsv1.CustomResourceDefinitionVersion{
			{
				Name:    "v1alpha1",
				Served:  true,
				Storage: true,
				Schema: &apiextensionsv1.CustomResourceValidation{
					OpenAPIV3Schema: &apiextensionsv1.JSONSchemaProps{
						Properties: map[string]apiextensionsv1.JSONSchemaProps{
							"apiVersion": {
								Type: "string",
							},
							"kind": {
								Type: "string",
							},
							"metadata": {
								Type: "object",
							},
							"spec": {
								Type:                   "object",
								XPreserveUnknownFields: &xPreserveUnknownFields,
							},
							"status": {
								Type:                   "object",
								XPreserveUnknownFields: &xPreserveUnknownFields,
							},
						},
						Required: []string{
							"metadata",
							"spec",
						},
						Type: "object",
					},
				},
			},
		},
	},
}
