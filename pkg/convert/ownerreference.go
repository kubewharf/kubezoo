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
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/kubewharf/kubezoo/pkg/util"
)

// OwnerReferenceTransformer transforms tenant owner reference to/from upstream owner reference
type OwnerReferenceTransformer interface {
	// Forward transforms tenant owner reference to upstream owner reference
	Forward(or *metav1.OwnerReference, tenantID string) (*metav1.OwnerReference, error)
	// Backward transforms upstream owner reference to tenant owner reference
	Backward(or *metav1.OwnerReference, tenantID string) (*metav1.OwnerReference, error)
}

type ownerReferenceTransformer struct {
	checkGroupKind util.CheckGroupKindFunc
}

// NewOwnerReferenceTransformer returns a new OwnerReferenceTransformer
func NewOwnerReferenceTransformer(checkGroupKind util.CheckGroupKindFunc) OwnerReferenceTransformer {
	return &ownerReferenceTransformer{checkGroupKind: checkGroupKind}
}

// Forward transforms tenant owner reference to upstream owner reference.
func (t *ownerReferenceTransformer) Forward(or *metav1.OwnerReference, tenantID string) (*metav1.OwnerReference, error) {
	if or == nil {
		return nil, nil
	}
	gv, err := schema.ParseGroupVersion(or.APIVersion)
	if err != nil {
		return nil, err
	}

	namespaced, customResourceGroup, err := t.checkGroupKind(gv.Group, or.Kind, tenantID, true)
	if err != nil {
		return nil, err
	}
	if !namespaced && len(or.Name) != 0 {
		or.Name = util.AddTenantIDPrefix(tenantID, or.Name)
	}
	if customResourceGroup && len(or.APIVersion) != 0 {
		or.APIVersion = util.AddTenantIDPrefix(tenantID, or.APIVersion)
	}
	return or, nil
}

// Backward transforms upstream owner reference to tenant owner reference
func (t *ownerReferenceTransformer) Backward(or *metav1.OwnerReference, tenantID string) (*metav1.OwnerReference, error) {
	if or == nil {
		return nil, nil
	}
	gv, err := schema.ParseGroupVersion(or.APIVersion)
	if err != nil {
		return nil, err
	}

	namespaced, customResourceGroup, err := t.checkGroupKind(gv.Group, or.Kind, tenantID, true)
	if err != nil {
		return nil, err
	}
	if !namespaced {
		if !strings.HasPrefix(or.Name, tenantID) {
			return nil, fmt.Errorf("ownerReference: %+v, name: %s must have tenantID prefix: %s", or, or.Name, tenantID)
		}
		or.Name = util.TrimTenantIDPrefix(tenantID, or.Name)
	}
	if customResourceGroup {
		if !strings.HasPrefix(or.APIVersion, tenantID) {
			return nil, fmt.Errorf("ownerReference: %+v, apiVersion: %s must have tenantID prefix: %s", or, or.Name, tenantID)
		}
		or.APIVersion = util.TrimTenantIDPrefix(tenantID, or.APIVersion)
	}
	return or, nil
}
