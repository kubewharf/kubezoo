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

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	rbacinternal "k8s.io/kubernetes/pkg/apis/rbac"
)

// TestRoleTransformerForward tests the forward method of the
// RoleTransformer.
func TestRoleTransformerForward(t *testing.T) {
	cases := []struct {
		name   string
		tenant string
		in     rbacinternal.Role
		want   rbacinternal.Role
	}{
		{
			name:   "test forward role",
			tenant: "111111",
			in: rbacinternal.Role{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Role",
					APIVersion: "rbac.authorization.k8s.io/v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "my-role",
				},
				Rules: []rbacinternal.PolicyRule{
					{
						Verbs:         []string{"get"},
						APIGroups:     []string{""},
						Resources:     []string{"secrets"},
						ResourceNames: []string{},
					},
					{
						Verbs:         []string{"get"},
						APIGroups:     []string{"kubezoo.io"},
						Resources:     []string{"foos"},
						ResourceNames: []string{},
					},
				},
			},
			want: rbacinternal.Role{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Role",
					APIVersion: "rbac.authorization.k8s.io/v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "my-role",
				},
				Rules: []rbacinternal.PolicyRule{
					{
						Verbs:         []string{"get"},
						APIGroups:     []string{""},
						Resources:     []string{"secrets"},
						ResourceNames: []string{},
					},
					{
						Verbs:         []string{"get"},
						APIGroups:     []string{"111111-kubezoo.io"},
						Resources:     []string{"foos"},
						ResourceNames: []string{},
					},
				},
			},
		},
	}

	fakeListTenantCRDsFunc := func(tenantID string) ([]*apiextensionsv1.CustomResourceDefinition, error) {
		crd := apiextensionsv1.CustomResourceDefinition{
			TypeMeta: metav1.TypeMeta{
				Kind:       "CustomResourceDefinition",
				APIVersion: "apiextensions/v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "foos.111111-kubezoo.io",
			},
			Spec: apiextensionsv1.CustomResourceDefinitionSpec{
				Group: "111111-kubezoo.io",
				Versions: []apiextensionsv1.CustomResourceDefinitionVersion{
					{
						Name: "v1",
					},
				},
				Scope: apiextensionsv1.NamespaceScoped,
				Names: apiextensionsv1.CustomResourceDefinitionNames{
					Plural: "foos",
					Kind:   "Foo",
				},
			},
		}
		return []*apiextensionsv1.CustomResourceDefinition{&crd}, nil
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			e := NewRoleTransformer(fakeListTenantCRDsFunc)
			if _, err := e.Forward(&c.in, c.tenant); err != nil {
				t.Fatalf("failed to forward role, err: %+v", err)
			}
			if !reflect.DeepEqual(c.in, c.want) {
				t.Errorf("out: %+v, want: %+v", c.in, c.want)
			}
		})
	}
}

// TestRoleTransformerBackward tests the backward method of the
// RoleTransformer.
func TestRoleTransformerBackward(t *testing.T) {
	cases := []struct {
		name   string
		tenant string
		in     rbacinternal.Role
		want   rbacinternal.Role
	}{
		{
			name:   "test backward role",
			tenant: "111111",
			in: rbacinternal.Role{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Role",
					APIVersion: "rbac.authorization.k8s.io/v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "my-role",
				},
				Rules: []rbacinternal.PolicyRule{
					{
						Verbs:         []string{"get"},
						APIGroups:     []string{""},
						Resources:     []string{"secrets"},
						ResourceNames: []string{},
					},
					{
						Verbs:         []string{"get"},
						APIGroups:     []string{"111111-kubezoo.io"},
						Resources:     []string{"foos"},
						ResourceNames: []string{},
					},
				},
			},
			want: rbacinternal.Role{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Role",
					APIVersion: "rbac.authorization.k8s.io/v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "my-role",
				},
				Rules: []rbacinternal.PolicyRule{
					{
						Verbs:         []string{"get"},
						APIGroups:     []string{""},
						Resources:     []string{"secrets"},
						ResourceNames: []string{},
					},
					{
						Verbs:         []string{"get"},
						APIGroups:     []string{"kubezoo.io"},
						Resources:     []string{"foos"},
						ResourceNames: []string{},
					},
				},
			},
		},
	}

	fakeListTenantCRDsFunc := func(tenantID string) ([]*apiextensionsv1.CustomResourceDefinition, error) {
		crd := apiextensionsv1.CustomResourceDefinition{
			TypeMeta: metav1.TypeMeta{
				Kind:       "CustomResourceDefinition",
				APIVersion: "apiextensions/v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "foos.111111-kubezoo.io",
			},
			Spec: apiextensionsv1.CustomResourceDefinitionSpec{
				Group: "111111-kubezoo.io",
				Versions: []apiextensionsv1.CustomResourceDefinitionVersion{
					{
						Name: "v1",
					},
				},
				Scope: apiextensionsv1.NamespaceScoped,
				Names: apiextensionsv1.CustomResourceDefinitionNames{
					Plural: "foos",
					Kind:   "Foo",
				},
			},
		}
		return []*apiextensionsv1.CustomResourceDefinition{&crd}, nil
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			e := NewRoleTransformer(fakeListTenantCRDsFunc)
			if _, err := e.Backward(&c.in, c.tenant); err != nil {
				t.Fatalf("failed to backward role, err: %+v", err)
			}
			if !reflect.DeepEqual(c.in, c.want) {
				t.Errorf("out: %+v, want: %+v", c.in, c.want)
			}
		})
	}
}
