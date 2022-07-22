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

package convert

import (
	"strings"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	sa "k8s.io/apiserver/pkg/authentication/serviceaccount"
	rbacinternal "k8s.io/kubernetes/pkg/apis/rbac"

	"github.com/kubewharf/kubezoo/pkg/util"
)

// TestClusterRoleBindingTransformerForward tests the forward method of the
// ClusterRoleBindingTransformer.
func TestClusterRoleBindingTransformerForward(t *testing.T) {
	tenant := "111111"
	originName := "mycrb"
	userName := "myuser"
	groupName := "system:serviceaccounts:mygroup"
	originalNamespace := "myns"
	serviceAccountName := "mysa"
	clusterRoleName := "mycr"

	crb := rbacinternal.ClusterRoleBinding{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterRoleBinding",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: originName,
		},
		Subjects: []rbacinternal.Subject{
			{
				Kind:     rbacinternal.UserKind,
				APIGroup: "rbac.authorization.k8s.io",
				Name:     userName,
			},
			{
				Kind:      rbacinternal.GroupKind,
				APIGroup:  "rbac.authorization.k8s.io",
				Name:      groupName,
				Namespace: originalNamespace,
			},
			{
				Kind:      rbacinternal.ServiceAccountKind,
				APIGroup:  "rbac.authorization.k8s.io",
				Name:      serviceAccountName,
				Namespace: originalNamespace,
			},
		},
		RoleRef: rbacinternal.RoleRef{
			Kind:     "ClusterRole",
			Name:     clusterRoleName,
			APIGroup: "rbac.authorization.k8s.io",
		},
	}

	c := NewClusterRoleBindingTransformer()
	c.Forward(&crb, tenant)

	if crb.RoleRef.Name != tenant+util.TenantIDSeparator+clusterRoleName {
		t.Errorf("Unexpected cluster role name.")
	}

	for _, subject := range crb.Subjects {
		switch subject.Kind {
		case rbacinternal.UserKind:
			if tenant+util.TenantIDSeparator+userName != subject.Name {
				t.Errorf("Unexpected user")
			}
		case rbacinternal.ServiceAccountKind:
			if tenant+util.TenantIDSeparator+originalNamespace != subject.Namespace {
				t.Errorf("Unexpected namespace")
			}
		case rbacinternal.GroupKind:
			namespace := strings.TrimPrefix(groupName, sa.ServiceAccountGroupPrefix)
			if subject.Name != sa.ServiceAccountGroupPrefix+util.AddTenantIDPrefix(tenant, namespace) {
				t.Errorf("Unexpected group name")
			}
		}
	}
}

// TestClusterRoleBindingTransformerBackward tests the backward method of the
// ClusterRoleBindingTransformer.
func TestClusterRoleBindingTransformerBackward(t *testing.T) {
	tenant := "111111"
	originName := "mycrb"
	userName := "myuser"
	groupName := "system:serviceaccounts:mygroup"
	originalNamespace := "myns"
	serviceAccountName := "mysa"
	clusterRoleName := "mycr"

	crb := rbacinternal.ClusterRoleBinding{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterRoleBinding",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: tenant + util.TenantIDSeparator + originName,
		},
		Subjects: []rbacinternal.Subject{
			{
				Kind:     rbacinternal.UserKind,
				APIGroup: "rbac.authorization.k8s.io",
				Name:     tenant + util.TenantIDSeparator + userName,
			},
			{
				Kind:      rbacinternal.GroupKind,
				APIGroup:  "rbac.authorization.k8s.io",
				Name:      sa.ServiceAccountGroupPrefix + util.AddTenantIDPrefix(tenant, (strings.TrimPrefix(groupName, sa.ServiceAccountGroupPrefix))),
				Namespace: originalNamespace,
			},
			{
				Kind:      rbacinternal.ServiceAccountKind,
				APIGroup:  "rbac.authorization.k8s.io",
				Name:      serviceAccountName,
				Namespace: tenant + util.TenantIDSeparator + originalNamespace,
			},
		},
		RoleRef: rbacinternal.RoleRef{
			Kind:     "ClusterRole",
			Name:     clusterRoleName,
			APIGroup: "rbac.authorization.k8s.io",
		},
	}

	c := NewClusterRoleBindingTransformer()
	c.Backward(&crb, tenant)

	if crb.RoleRef.Name != clusterRoleName {
		t.Errorf("Unexpected cluster role name.")
	}

	for _, subject := range crb.Subjects {
		switch subject.Kind {
		case rbacinternal.UserKind:
			if userName != subject.Name {
				t.Errorf("Unexpected user")
			}
		case rbacinternal.ServiceAccountKind:
			if originalNamespace != subject.Namespace {
				t.Errorf("Unexpected namespace")
			}
		case rbacinternal.GroupKind:
			if subject.Name != groupName {
				t.Errorf("Unexpected group name")
			}
		}
	}
}
