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
	"testing"

	crdinternal "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kubewharf/kubezoo/pkg/util"
)

// TestCRDConvertorConvertTenantObjectToUpstreamObject tests the
// ConvertTenantObjectToUpstreamObject methods of CRDConvertor.
func TestCRDConvertorConvertTenantObjectToUpstreamObject(t *testing.T) {

	tenant := "111111"
	CRDPlural := "foos"
	CRDGroup := "a.com"
	CRDVersion := "v1"
	FullCRDName := CRDPlural + "." + CRDGroup
	c := NewCRDConvertor(NewOwnerReferenceTransformer(checkGroupKind))

	testCases := map[string]struct {
		crd       crdinternal.CustomResourceDefinition
		expectErr bool
	}{
		"This is a normal crd": {
			crd: crdinternal.CustomResourceDefinition{
				TypeMeta: metav1.TypeMeta{
					Kind:       "CustomResourceDefinition",
					APIVersion: "apiextensions/v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: FullCRDName,
				},
				Spec: crdinternal.CustomResourceDefinitionSpec{
					Group:   CRDGroup,
					Version: CRDVersion,
					Scope:   crdinternal.NamespaceScoped,
					Names: crdinternal.CustomResourceDefinitionNames{
						Plural: CRDPlural,
						Kind:   "Foo",
					},
				},
			},
			expectErr: false,
		},
		"This is a error crd": {
			crd: crdinternal.CustomResourceDefinition{
				TypeMeta: metav1.TypeMeta{
					Kind:       "CustomResourceDefinition",
					APIVersion: "apiextensions/v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: FullCRDName + "badname",
				},
				Spec: crdinternal.CustomResourceDefinitionSpec{
					Group:   CRDGroup,
					Version: CRDVersion,
					Scope:   crdinternal.NamespaceScoped,
					Names: crdinternal.CustomResourceDefinitionNames{
						Plural: CRDPlural,
						Kind:   "Foo",
					},
				},
			},
			expectErr: true,
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			err := c.ConvertTenantObjectToUpstreamObject(&testCase.crd, tenant, true)
			if testCase.expectErr {
				if err == nil {
					t.Errorf("Expect error")
				} else {
					return
				}
			}
			if err != nil {
				t.Errorf("Failed ConvertTenantObjectToUpstreamObject with err %s", err)
			}
			if testCase.crd.Spec.Group != tenant+util.TenantIDSeparator+CRDGroup {
				t.Errorf("Unexpected group.")
			}
			if testCase.crd.Name != CRDPlural+"."+tenant+util.TenantIDSeparator+CRDGroup {
				t.Errorf("Unexpected group.")
			}
		})
	}
}

// TestCRDConvertorConvertUpstreamObjectToTenantObject tests the
// ConvertUpstreamObjectToTenantObject methods of CRDConvertor.
func TestCRDConvertorConvertUpstreamObjectToTenantObject(t *testing.T) {
	tenant := "111111"
	CRDPlural := "foos"
	CRDGroup := tenant + util.TenantIDSeparator + "a.com"
	CRDVersion := "v1"
	FullCRDName := CRDPlural + "." + CRDGroup
	c := NewCRDConvertor(NewOwnerReferenceTransformer(checkGroupKind))

	testCases := map[string]struct {
		crd       crdinternal.CustomResourceDefinition
		expectErr bool
	}{
		"This is a normal crd": {
			crd: crdinternal.CustomResourceDefinition{
				TypeMeta: metav1.TypeMeta{
					Kind:       "CustomResourceDefinition",
					APIVersion: "apiextensions/v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: FullCRDName,
				},
				Spec: crdinternal.CustomResourceDefinitionSpec{
					Group:   CRDGroup,
					Version: CRDVersion,
					Scope:   crdinternal.NamespaceScoped,
					Names: crdinternal.CustomResourceDefinitionNames{
						Plural: CRDPlural,
						Kind:   "Foo",
					},
				},
			},
			expectErr: false,
		},
		"This is a err crd": {
			crd: crdinternal.CustomResourceDefinition{
				TypeMeta: metav1.TypeMeta{
					Kind:       "CustomResourceDefinition",
					APIVersion: "apiextensions/v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: FullCRDName,
				},
				Spec: crdinternal.CustomResourceDefinitionSpec{
					Group:   "bad" + CRDGroup,
					Version: CRDVersion,
					Scope:   crdinternal.NamespaceScoped,
					Names: crdinternal.CustomResourceDefinitionNames{
						Plural: CRDPlural,
						Kind:   "Foo",
					},
				},
			},
			expectErr: true,
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			err := c.ConvertUpstreamObjectToTenantObject(&testCase.crd, tenant, true)
			if testCase.expectErr {
				if err == nil {
					t.Errorf("Expect error")
				} else {
					return
				}
			}
			if err != nil {
				t.Errorf("Failed ConvertTenantObjectToUpstreamObject with err %s", err)
			}
			if tenant+util.TenantIDSeparator+testCase.crd.Spec.Group != CRDGroup {
				t.Errorf("Unexpected group.")
			}
			if testCase.crd.Name != CRDPlural+"."+util.TrimTenantIDPrefix(tenant, CRDGroup) {
				t.Errorf("Unexpected group.")
			}
		})
	}
}
