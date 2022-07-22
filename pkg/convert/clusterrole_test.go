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
	"fmt"
	"testing"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	rbacinternal "k8s.io/kubernetes/pkg/apis/rbac"

	"github.com/kubewharf/kubezoo/pkg/util"
)

// TestClusterRoleTransformerForward tests the forward method of the
// ClusterRoleTransformer.
func TestClusterRoleTransformerForward(t *testing.T) {
	tenant := "111111"
	CRDGroup := util.AddTenantIDPrefix(tenant, "kubezoo.io")
	CRDVersion := "v1"
	CRDPlural := "foos"
	FullCRDName := CRDPlural + "." + CRDGroup
	originName := "mycr"
	cr := rbacinternal.ClusterRole{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterRole",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: originName,
		},
		Rules: []rbacinternal.PolicyRule{
			{
				Verbs:     []string{"get"},
				APIGroups: []string{""},
				Resources: []string{"secrets"},
			},
			{
				Verbs:     []string{"get"},
				APIGroups: []string{util.TrimTenantIDPrefix(tenant, CRDGroup)},
				Resources: []string{CRDPlural},
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
				Name: FullCRDName,
			},
			Spec: apiextensionsv1.CustomResourceDefinitionSpec{
				Group: CRDGroup,
				Versions: []apiextensionsv1.CustomResourceDefinitionVersion{
					{
						Name: CRDVersion,
					},
				},
				Scope: apiextensionsv1.NamespaceScoped,
				Names: apiextensionsv1.CustomResourceDefinitionNames{
					Plural: CRDPlural,
					Kind:   "Foo",
				},
			},
		}
		return []*apiextensionsv1.CustomResourceDefinition{&crd}, nil
	}

	c := NewClusterRoleTransformer(fakeListTenantCRDsFunc)
	_, err := c.Forward(&cr, tenant)
	if err != nil {
		t.Errorf("Failed to forward with err %s", err)
	}

	fmt.Printf("%v", cr)

	secret := cr.Rules[0].Resources
	if len(secret) != 1 && secret[0] != "secrets" {
		t.Errorf("Unexpected resource")
	}

	crd := cr.Rules[1].APIGroups
	fmt.Printf("%v", crd)
	if len(crd) != 1 || crd[0] != CRDGroup {
		t.Errorf("Unexpected resource")
	}
}

// TestClusterRoleTransformerBackward tests the backward method of the
// ClusterRoleTransformer.
func TestClusterRoleTransformerBackward(t *testing.T) {
	tenant := "111111"
	CRDGroup := util.AddTenantIDPrefix(tenant, "kubezoo.io")
	CRDVersion := "v1"
	CRDPlural := "foos"
	FullCRDName := CRDPlural + "." + CRDGroup
	originName := "mycr"
	cr := rbacinternal.ClusterRole{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterRole",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: originName,
		},
		Rules: []rbacinternal.PolicyRule{
			{
				Verbs:     []string{"get"},
				APIGroups: []string{""},
				Resources: []string{"secrets"},
			},
			{
				Verbs:     []string{"get"},
				APIGroups: []string{CRDGroup},
				Resources: []string{CRDPlural},
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
				Name: FullCRDName,
			},
			Spec: apiextensionsv1.CustomResourceDefinitionSpec{
				Group: CRDGroup,
				Versions: []apiextensionsv1.CustomResourceDefinitionVersion{
					{
						Name: CRDVersion,
					},
				},
				Scope: apiextensionsv1.NamespaceScoped,
				Names: apiextensionsv1.CustomResourceDefinitionNames{
					Plural: CRDPlural,
					Kind:   "Foo",
				},
			},
		}
		return []*apiextensionsv1.CustomResourceDefinition{&crd}, nil
	}

	c := NewClusterRoleTransformer(fakeListTenantCRDsFunc)
	_, err := c.Backward(&cr, tenant)
	if err != nil {
		t.Errorf("Failed to forward with err %s", err)
	}

	fmt.Printf("%v", cr)

	secret := cr.Rules[0].Resources
	if len(secret) != 1 && secret[0] != "secrets" {
		t.Errorf("Unexpected resource")
	}

	crd := cr.Rules[1].APIGroups
	fmt.Printf("%v", crd)
	if len(crd) != 1 || crd[0] != util.TrimTenantIDPrefix(tenant, CRDGroup) {
		t.Errorf("Unexpected resource")
	}
}
