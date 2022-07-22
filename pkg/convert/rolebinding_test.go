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
	"reflect"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	rbacinternal "k8s.io/kubernetes/pkg/apis/rbac"
)

// TestRoleBindingTransformerForward tests the forward method of the
// RoleBindingTransformer.
func TestRoleBindingTransformerForward(t *testing.T) {
	cases := []struct {
		name   string
		tenant string
		in     rbacinternal.RoleBinding
		want   rbacinternal.RoleBinding
	}{
		{
			name:   "test forward rolebinding which references a role",
			tenant: "111111",
			in: rbacinternal.RoleBinding{
				TypeMeta: metav1.TypeMeta{
					Kind:       "RoleBinding",
					APIVersion: "rbac.authorization.k8s.io/v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-rolebinding",
					Namespace: "default",
				},
				Subjects: []rbacinternal.Subject{
					{
						Kind:     rbacinternal.UserKind,
						APIGroup: "rbac.authorization.k8s.io",
						Name:     "my-user",
					},
					{
						Kind:      rbacinternal.GroupKind,
						APIGroup:  "rbac.authorization.k8s.io",
						Name:      "system:serviceaccounts:mygroup",
						Namespace: "my-ns",
					},
					{
						Kind:      rbacinternal.ServiceAccountKind,
						APIGroup:  "rbac.authorization.k8s.io",
						Name:      "my-sa",
						Namespace: "my-ns",
					},
				},
				RoleRef: rbacinternal.RoleRef{
					Kind:     "Role",
					Name:     "my-role",
					APIGroup: "rbac.authorization.k8s.io",
				},
			},
			want: rbacinternal.RoleBinding{
				TypeMeta: metav1.TypeMeta{
					Kind:       "RoleBinding",
					APIVersion: "rbac.authorization.k8s.io/v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-rolebinding",
					Namespace: "default",
				},
				Subjects: []rbacinternal.Subject{
					{
						Kind:     rbacinternal.UserKind,
						APIGroup: "rbac.authorization.k8s.io",
						Name:     "111111-my-user",
					},
					{
						Kind:      rbacinternal.GroupKind,
						APIGroup:  "rbac.authorization.k8s.io",
						Name:      "system:serviceaccounts:111111-mygroup",
						Namespace: "my-ns",
					},
					{
						Kind:      rbacinternal.ServiceAccountKind,
						APIGroup:  "rbac.authorization.k8s.io",
						Name:      "my-sa",
						Namespace: "111111-my-ns",
					},
				},
				RoleRef: rbacinternal.RoleRef{
					Kind:     "Role",
					Name:     "my-role",
					APIGroup: "rbac.authorization.k8s.io",
				},
			},
		},
		{
			name:   "test forward rolebinding which references a clusterrole",
			tenant: "111111",
			in: rbacinternal.RoleBinding{
				TypeMeta: metav1.TypeMeta{
					Kind:       "RoleBinding",
					APIVersion: "rbac.authorization.k8s.io/v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-rolebinding",
					Namespace: "default",
				},
				Subjects: []rbacinternal.Subject{
					{
						Kind:     rbacinternal.UserKind,
						APIGroup: "rbac.authorization.k8s.io",
						Name:     "my-user",
					},
					{
						Kind:      rbacinternal.GroupKind,
						APIGroup:  "rbac.authorization.k8s.io",
						Name:      "system:serviceaccounts:mygroup",
						Namespace: "my-ns",
					},
					{
						Kind:      rbacinternal.ServiceAccountKind,
						APIGroup:  "rbac.authorization.k8s.io",
						Name:      "my-sa",
						Namespace: "my-ns",
					},
				},
				RoleRef: rbacinternal.RoleRef{
					Kind:     "ClusterRole",
					Name:     "my-clusterrole",
					APIGroup: "rbac.authorization.k8s.io",
				},
			},
			want: rbacinternal.RoleBinding{
				TypeMeta: metav1.TypeMeta{
					Kind:       "RoleBinding",
					APIVersion: "rbac.authorization.k8s.io/v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-rolebinding",
					Namespace: "default",
				},
				Subjects: []rbacinternal.Subject{
					{
						Kind:     rbacinternal.UserKind,
						APIGroup: "rbac.authorization.k8s.io",
						Name:     "111111-my-user",
					},
					{
						Kind:      rbacinternal.GroupKind,
						APIGroup:  "rbac.authorization.k8s.io",
						Name:      "system:serviceaccounts:111111-mygroup",
						Namespace: "my-ns",
					},
					{
						Kind:      rbacinternal.ServiceAccountKind,
						APIGroup:  "rbac.authorization.k8s.io",
						Name:      "my-sa",
						Namespace: "111111-my-ns",
					},
				},
				RoleRef: rbacinternal.RoleRef{
					Kind:     "ClusterRole",
					Name:     "111111-my-clusterrole",
					APIGroup: "rbac.authorization.k8s.io",
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			e := NewRoleBindingTransformer()
			if _, err := e.Forward(&c.in, c.tenant); err != nil {
				t.Fatalf("failed to forward rolebinding, err: %+v", err)
			}
			if !reflect.DeepEqual(c.in, c.want) {
				t.Errorf("out: %+v, want: %+v", c.in, c.want)
			}
		})
	}
}

// TestRoleBindingTransformerBackward tests the backward method of the
// RoleBindingTransformer.
func TestRoleBindingTransformerBackward(t *testing.T) {
	cases := []struct {
		name   string
		tenant string
		in     rbacinternal.RoleBinding
		want   rbacinternal.RoleBinding
	}{
		{
			name:   "test backward rolebinding which references a role",
			tenant: "111111",
			in: rbacinternal.RoleBinding{
				TypeMeta: metav1.TypeMeta{
					Kind:       "RoleBinding",
					APIVersion: "rbac.authorization.k8s.io/v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-rolebinding",
					Namespace: "default",
				},
				Subjects: []rbacinternal.Subject{
					{
						Kind:     rbacinternal.UserKind,
						APIGroup: "rbac.authorization.k8s.io",
						Name:     "111111-my-user",
					},
					{
						Kind:      rbacinternal.GroupKind,
						APIGroup:  "rbac.authorization.k8s.io",
						Name:      "system:serviceaccounts:111111-mygroup",
						Namespace: "my-ns",
					},
					{
						Kind:      rbacinternal.ServiceAccountKind,
						APIGroup:  "rbac.authorization.k8s.io",
						Name:      "my-sa",
						Namespace: "111111-my-ns",
					},
				},
				RoleRef: rbacinternal.RoleRef{
					Kind:     "Role",
					Name:     "my-role",
					APIGroup: "rbac.authorization.k8s.io",
				},
			},
			want: rbacinternal.RoleBinding{
				TypeMeta: metav1.TypeMeta{
					Kind:       "RoleBinding",
					APIVersion: "rbac.authorization.k8s.io/v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-rolebinding",
					Namespace: "default",
				},
				Subjects: []rbacinternal.Subject{
					{
						Kind:     rbacinternal.UserKind,
						APIGroup: "rbac.authorization.k8s.io",
						Name:     "my-user",
					},
					{
						Kind:      rbacinternal.GroupKind,
						APIGroup:  "rbac.authorization.k8s.io",
						Name:      "system:serviceaccounts:mygroup",
						Namespace: "my-ns",
					},
					{
						Kind:      rbacinternal.ServiceAccountKind,
						APIGroup:  "rbac.authorization.k8s.io",
						Name:      "my-sa",
						Namespace: "my-ns",
					},
				},
				RoleRef: rbacinternal.RoleRef{
					Kind:     "Role",
					Name:     "my-role",
					APIGroup: "rbac.authorization.k8s.io",
				},
			},
		},
		{
			name:   "test backward rolebinding which references a clusterrole",
			tenant: "111111",
			in: rbacinternal.RoleBinding{
				TypeMeta: metav1.TypeMeta{
					Kind:       "RoleBinding",
					APIVersion: "rbac.authorization.k8s.io/v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-rolebinding",
					Namespace: "default",
				},
				Subjects: []rbacinternal.Subject{
					{
						Kind:     rbacinternal.UserKind,
						APIGroup: "rbac.authorization.k8s.io",
						Name:     "111111-my-user",
					},
					{
						Kind:      rbacinternal.GroupKind,
						APIGroup:  "rbac.authorization.k8s.io",
						Name:      "system:serviceaccounts:111111-mygroup",
						Namespace: "my-ns",
					},
					{
						Kind:      rbacinternal.ServiceAccountKind,
						APIGroup:  "rbac.authorization.k8s.io",
						Name:      "my-sa",
						Namespace: "111111-my-ns",
					},
				},
				RoleRef: rbacinternal.RoleRef{
					Kind:     "ClusterRole",
					Name:     "111111-my-clusterrole",
					APIGroup: "rbac.authorization.k8s.io",
				},
			},
			want: rbacinternal.RoleBinding{
				TypeMeta: metav1.TypeMeta{
					Kind:       "RoleBinding",
					APIVersion: "rbac.authorization.k8s.io/v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-rolebinding",
					Namespace: "default",
				},
				Subjects: []rbacinternal.Subject{
					{
						Kind:     rbacinternal.UserKind,
						APIGroup: "rbac.authorization.k8s.io",
						Name:     "my-user",
					},
					{
						Kind:      rbacinternal.GroupKind,
						APIGroup:  "rbac.authorization.k8s.io",
						Name:      "system:serviceaccounts:mygroup",
						Namespace: "my-ns",
					},
					{
						Kind:      rbacinternal.ServiceAccountKind,
						APIGroup:  "rbac.authorization.k8s.io",
						Name:      "my-sa",
						Namespace: "my-ns",
					},
				},
				RoleRef: rbacinternal.RoleRef{
					Kind:     "ClusterRole",
					Name:     "my-clusterrole",
					APIGroup: "rbac.authorization.k8s.io",
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			e := NewRoleBindingTransformer()
			if _, err := e.Backward(&c.in, c.tenant); err != nil {
				t.Fatalf("failed to backward rolebinding, err: %+v", err)
			}
			if !reflect.DeepEqual(c.in, c.want) {
				t.Errorf("out: %+v, want: %+v", c.in, c.want)
			}
		})
	}
}
