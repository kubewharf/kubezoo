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

import v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

// CustomGroupResourcesMap records the existence of all custom api group and resources for a tenant
// the first key is api group and the second key is resource name
type CustomGroupResourcesMap map[string]map[string]*v1.CustomResourceDefinition

// NewCustomGroupResourcesMap return a CRD map.
func NewCustomGroupResourcesMap(crdList []*v1.CustomResourceDefinition) CustomGroupResourcesMap {
	grm := make(map[string]map[string]*v1.CustomResourceDefinition)
	for _, crd := range crdList {
		_, ok := grm[crd.Spec.Group]
		if !ok {
			grm[crd.Spec.Group] = make(map[string]*v1.CustomResourceDefinition)
		}
		grm[crd.Spec.Group][crd.Spec.Names.Plural] = crd
	}
	return grm
}

// HasGroup checks the map contains the api group or not.
func (grm CustomGroupResourcesMap) HasGroup(apiGroup string) bool {
	return grm[apiGroup] != nil
}

// HasResource checks the map contains the resource or not.
func (grm CustomGroupResourcesMap) HasResource(resourceName string) bool {
	for _, resources := range grm {
		if resources[resourceName] != nil {
			return true
		}
	}
	return false
}

// HasGroupResource checks the map contains the group resource or not.
func (grm CustomGroupResourcesMap) HasGroupResource(apiGroup, resourceName string) bool {
	return grm[apiGroup][resourceName] != nil
}

// HasGroupVersion checks the map contains the group version or not.
func (grm CustomGroupResourcesMap) HasGroupVersion(apiGroup, version string) bool {
	for _, crd := range grm[apiGroup] {
		if crd == nil {
			continue
		}
		for i := range crd.Spec.Versions {
			if crd.Spec.Versions[i].Name == version {
				return true
			}
		}
	}
	return false
}

// HasGroupVersionResource checks the map contains the group version resource or not.
func (grm CustomGroupResourcesMap) HasGroupVersionResource(apiGroup, version, resourceName string) bool {
	crd := grm.GetCRD(apiGroup, resourceName)
	if crd == nil {
		return false
	}
	for i := range crd.Spec.Versions {
		if crd.Spec.Versions[i].Name == version {
			return true
		}
	}
	return false
}

// GetCRD return the CRD by APIGroup and resource name.
func (grm CustomGroupResourcesMap) GetCRD(apiGroup, resourceName string) *v1.CustomResourceDefinition {
	return grm[apiGroup][resourceName]
}
