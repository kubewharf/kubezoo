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
	"encoding/json"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	schedulingv1 "k8s.io/api/scheduling/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog"

	tenantv1alpha1 "github.com/kubewharf/kubezoo/pkg/apis/tenant/v1alpha1"
	"github.com/kubewharf/kubezoo/pkg/util"
)

var _ = Describe("Tenant controller", func() {
	var tenant *tenantv1alpha1.Tenant
	BeforeEach(func() {
		var err error
		tenant, err = controlPlaneClient.TenantV1alpha1().Tenants().Create(ctx, &tenantv1alpha1.Tenant{
			ObjectMeta: metav1.ObjectMeta{
				Name: testTenantName,
			},
		}, metav1.CreateOptions{})
		if errors.IsAlreadyExists(err) {
			err = nil
		}
		Expect(err).NotTo(HaveOccurred())
		time.Sleep(1 * time.Second)
	})

	It("init resource for tenant", func() {
		systemNamespaces := []string{metav1.NamespaceSystem, metav1.NamespacePublic, corev1.NamespaceNodeLease, corev1.NamespaceDefault}
		for _, systemNamespace := range systemNamespaces {
			name := tenant.Name + "-" + systemNamespace
			_, err := upstreamClient.CoreV1().Namespaces().Get(ctx, name, metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())
		}

		clusterRoles := []string{"cluster-admin", "admin"}
		for _, clusterRole := range clusterRoles {
			name := tenant.Name + "-" + clusterRole
			_, err := upstreamClient.RbacV1().ClusterRoles().Get(ctx, name, metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())
		}

		clusterRoleBindings := []string{"cluster-admin"}
		for _, clusterRoleBinding := range clusterRoleBindings {
			name := tenant.Name + "-" + clusterRoleBinding
			_, err := upstreamClient.RbacV1().ClusterRoleBindings().Get(ctx, name, metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())
		}
		var err error
		tenant, err = controlPlaneClient.TenantV1alpha1().Tenants().Get(ctx, tenant.Name, metav1.GetOptions{})
		Expect(err).NotTo(HaveOccurred())
		var kubeconfig string
		if len(tenant.Annotations) > 0 {
			kubeconfig = tenant.Annotations[util.AnnotationTenantKubeConfigBase64]
		}
		Expect(kubeconfig).NotTo(BeEmpty())
	})

	It("create native cluster-scoped resources", func() {
		var err error

		_, err = upstreamClient.SchedulingV1().PriorityClasses().Create(ctx, &schedulingv1.PriorityClass{
			ObjectMeta: metav1.ObjectMeta{
				Name: testTenantName + "-" + testPriorityClassName,
			},
			Value:         1000000,
			GlobalDefault: false,
		}, metav1.CreateOptions{})
		if errors.IsAlreadyExists(err) {
			err = nil
		}
		Expect(err).NotTo(HaveOccurred())

		time.Sleep(1 * time.Second)

		_, err = upstreamClient.SchedulingV1().PriorityClasses().Get(ctx, testTenantName+"-"+testPriorityClassName, metav1.GetOptions{})
		Expect(err).NotTo(HaveOccurred())
	})

	It("create CRD", func() {
		var err error

		testCRD := &apiextensionsv1.CustomResourceDefinition{
			ObjectMeta: metav1.ObjectMeta{
				Name: crdPlural + "." + testTenantName + "-" + crdGroup,
			},
			Spec: apiextensionsv1.CustomResourceDefinitionSpec{
				Group: testTenantName + "-" + crdGroup,
				Scope: apiextensionsv1.ClusterScoped,
				Names: apiextensionsv1.CustomResourceDefinitionNames{
					Plural:   crdPlural,
					Kind:     "Foo",
					ListKind: "FooList",
					Singular: "foo",
				},
				Versions: []apiextensionsv1.CustomResourceDefinitionVersion{
					{
						Name:    "v1",
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

		_, err = crdClient.ApiextensionsV1().CustomResourceDefinitions().Create(ctx, testCRD, metav1.CreateOptions{})
		if errors.IsAlreadyExists(err) {
			err = nil
		}
		Expect(err).NotTo(HaveOccurred())

		time.Sleep(1 * time.Second)

		_, err = crdClient.ApiextensionsV1().CustomResourceDefinitions().Get(ctx, crdPlural+"."+testTenantName+"-"+crdGroup, metav1.GetOptions{})
		Expect(err).NotTo(HaveOccurred())
	})

	It("create CR", func() {
		var err error
		crJson := `{"apiVersion":"kubezoo-controller-test-a.com/v1","kind":"Foo","metadata":{"name":"kubezoo-controller-test-my-foo"},"spec":{"a":"b"}}`

		obj := &unstructured.Unstructured{}
		err = json.Unmarshal([]byte(crJson), &obj.Object)
		Expect(err).NotTo(HaveOccurred())
		Expect(obj.Object).NotTo(BeEmpty())

		gvr := schema.GroupVersionResource{
			Group:    "kubezoo-controller-test-a.com",
			Version:  "v1",
			Resource: "foos",
		}

		_, err = dynamicClient.Resource(gvr).Create(ctx, obj, metav1.CreateOptions{})
		if errors.IsAlreadyExists(err) {
			err = nil
		}
		Expect(err).NotTo(HaveOccurred())

		time.Sleep(1 * time.Second)

		_, err = dynamicClient.Resource(gvr).Get(ctx, "kubezoo-controller-test-my-foo", metav1.GetOptions{})
		Expect(err).NotTo(HaveOccurred())
	})

	It("delete tenant", func() {
		err := controlPlaneClient.TenantV1alpha1().Tenants().Delete(ctx, testTenantName, metav1.DeleteOptions{})
		Expect(err).NotTo(HaveOccurred())
		time.Sleep(2 * time.Second)

		systemNamespaces := []string{metav1.NamespaceSystem, metav1.NamespacePublic, corev1.NamespaceNodeLease, corev1.NamespaceDefault}
		for _, systemNamespace := range systemNamespaces {
			name := testTenantName + "-" + systemNamespace
			obj, err := upstreamClient.CoreV1().Namespaces().Get(ctx, name, metav1.GetOptions{})
			if err == nil {
				Expect(obj.DeletionTimestamp).NotTo(BeNil())
			} else {
				Expect(errors.IsNotFound(err)).To(BeTrue())
			}
		}

		clusterRoles := []string{"cluster-admin", "admin"}
		for _, clusterRole := range clusterRoles {
			name := testTenantName + "-" + clusterRole
			obj, err := upstreamClient.RbacV1().ClusterRoles().Get(ctx, name, metav1.GetOptions{})
			if err == nil {
				klog.Warningf("obj name %s, deletionTime %v", name, obj.DeletionTimestamp)
				Expect(obj.DeletionTimestamp).NotTo(BeNil())
			} else {
				Expect(errors.IsNotFound(err)).To(BeTrue())
			}
		}

		clusterRoleBindings := []string{"cluster-admin"}
		for _, clusterRoleBinding := range clusterRoleBindings {
			name := testTenantName + "-" + clusterRoleBinding
			obj, err := upstreamClient.RbacV1().ClusterRoleBindings().Get(ctx, name, metav1.GetOptions{})
			if err == nil {
				Expect(obj.DeletionTimestamp).NotTo(BeNil())
			} else {
				Expect(errors.IsNotFound(err)).To(BeTrue())
			}
		}

		if obj, err := upstreamClient.SchedulingV1().PriorityClasses().Get(ctx, testTenantName+"-"+testPriorityClassName, metav1.GetOptions{}); err == nil {
			Expect(obj.DeletionTimestamp).NotTo(BeNil())
		} else {
			Expect(errors.IsNotFound(err)).To(BeTrue())
		}

		if obj, err := crdClient.ApiextensionsV1().CustomResourceDefinitions().Get(ctx, crdPlural+"."+testTenantName+"-"+crdGroup, metav1.GetOptions{}); err == nil {
			Expect(obj.GetDeletionTimestamp()).NotTo(BeNil())
		} else {
			Expect(errors.IsNotFound(err)).To(BeTrue())
		}

		if obj, err := dynamicClient.Resource(schema.GroupVersionResource{
			Group:    "kubezoo-controller-test-a.com",
			Version:  "v1",
			Resource: "foos",
		}).Get(ctx, "kubezoo-controller-test-my-foo", metav1.GetOptions{}); err == nil {
			Expect(obj.GetDeletionTimestamp()).NotTo(BeNil())
		} else {
			Expect(errors.IsNotFound(err)).To(BeTrue())
		}
	})
})
