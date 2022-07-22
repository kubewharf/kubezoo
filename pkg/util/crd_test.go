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

package util

import (
	"testing"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TestCRD mainly tests the methods of CustomGroupResourcesMap.
func TestCRD(t *testing.T) {

	tenant := "111111"
	CRDPlural := "foos"
	CRDGroup := tenant + "-" + "a.com"
	CRDVersion := "v1"
	FullCRDName := CRDPlural + "." + CRDGroup

	crdList := []*apiextensionsv1.CustomResourceDefinition{
		{
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
		},
	}

	m := NewCustomGroupResourcesMap(crdList)

	if !m.HasGroup(CRDGroup) {
		t.Errorf("group %s not found, expected found", CRDGroup)
	}

	if !m.HasResource(CRDPlural) {
		t.Errorf("group %s not found, expected found", CRDPlural)
	}

	if !m.HasGroupResource(CRDGroup, CRDPlural) {
		t.Errorf("group resource %s/%s not found, expected found", CRDGroup, CRDPlural)
	}

	if !m.HasGroupVersion(CRDGroup, CRDVersion) {
		t.Errorf("group version %s/%s not found, expected found", CRDGroup, CRDVersion)
	}

	if !m.HasGroupVersionResource(CRDGroup, CRDVersion, CRDPlural) {
		t.Errorf("group version %s/%s/%s not found, expected found", CRDGroup, CRDVersion, CRDVersion)
	}

	crd := m.GetCRD(CRDGroup, CRDPlural)
	if crd == nil {
		t.Errorf("crd should not be nil.")
	}
}
