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
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
		for _, clusterrole := range clusterRoles {
			name := tenant.Name + "-" + clusterrole
			_, err := upstreamClient.RbacV1().ClusterRoles().Get(ctx, name, metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())
		}

		clusterRoleBindings := []string{"cluster-admin"}
		for _, clusterclusterRoleBinding := range clusterRoleBindings {
			name := tenant.Name + "-" + clusterclusterRoleBinding
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

	It("delete tenant", func() {
		err := controlPlaneClient.TenantV1alpha1().Tenants().Delete(ctx, testTenantName, metav1.DeleteOptions{})
		klog.Warningf("delete tenant")
		Expect(err).NotTo(HaveOccurred())
		time.Sleep(2 * time.Second)
		ns, _ := upstreamClient.RbacV1().ClusterRoles().List(ctx, metav1.ListOptions{})
		for _, n := range ns.Items {
			klog.Warningf("objects name %v, DeleteTime %v", n.Name, n.DeletionTimestamp)
		}

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
		for _, clusterrole := range clusterRoles {
			name := testTenantName + "-" + clusterrole
			obj, err := upstreamClient.RbacV1().ClusterRoles().Get(ctx, name, metav1.GetOptions{})
			if err == nil {
				klog.Warningf("obj name %s, deletionTime %v", name, obj.DeletionTimestamp)
				Expect(obj.DeletionTimestamp).NotTo(BeNil())
			} else {
				Expect(errors.IsNotFound(err)).To(BeTrue())
			}
		}

		clusterRoleBindings := []string{"cluster-admin"}
		for _, clusterclusterRoleBinding := range clusterRoleBindings {
			name := testTenantName + "-" + clusterclusterRoleBinding
			obj, err := upstreamClient.RbacV1().ClusterRoleBindings().Get(ctx, name, metav1.GetOptions{})
			if err == nil {
				Expect(obj.DeletionTimestamp).NotTo(BeNil())
			} else {
				Expect(errors.IsNotFound(err)).To(BeTrue())
			}
		}
	})
})
