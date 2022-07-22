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
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/kubewharf/kubezoo/pkg/common"
	"github.com/kubewharf/kubezoo/pkg/util"
)

// DefaultConvertor implements the transformation between
// client and upstream server for generic default resource.
type DefaultConvertor struct {
	ownerRefTransformer OwnerReferenceTransformer
}

var _ common.ObjectConvertor = &DefaultConvertor{}

// NewDefaultConvertor initiates a DefaultConvertor which implements the
// ObjectConvertor interfaces.
func NewDefaultConvertor(ort OwnerReferenceTransformer) common.ObjectConvertor {
	return &DefaultConvertor{
		ownerRefTransformer: ort,
	}
}

// ConvertTenantObjectToUpstreamObject convert the tenant object to
// upstream object.
func (c *DefaultConvertor) ConvertTenantObjectToUpstreamObject(obj runtime.Object, tenantID string, isNamespaceScoped bool) error {
	accessor, err := meta.Accessor(obj)
	if err != nil {
		return err
	}
	if isNamespaceScoped && accessor.GetNamespace() != "" {
		prefixed := util.AddTenantIDPrefix(tenantID, accessor.GetNamespace())
		accessor.SetNamespace(prefixed)
	} else if !isNamespaceScoped && accessor.GetName() != "" {
		prefixed := util.AddTenantIDPrefix(tenantID, accessor.GetName())
		accessor.SetName(prefixed)
	}
	ownerReferences := accessor.GetOwnerReferences()
	for i := range ownerReferences {
		target, err := c.ownerRefTransformer.Forward(&ownerReferences[i], tenantID)
		if err != nil {
			return err
		}
		ownerReferences[i] = *target
	}
	accessor.SetOwnerReferences(ownerReferences)
	return nil
}

// ConvertUpstreamObjectToTenantObject convert the upstream object to
// tenant object.
func (c *DefaultConvertor) ConvertUpstreamObjectToTenantObject(obj runtime.Object, tenantID string, isNamespaceScoped bool) error {
	accessor, err := meta.Accessor(obj)
	if err != nil {
		return err
	}
	// some object may be valid to have empty name and namespace, such as TokenReview
	if isNamespaceScoped {
		namespace := accessor.GetNamespace()
		trimmed := util.TrimTenantIDPrefix(tenantID, namespace)
		accessor.SetNamespace(trimmed)
	} else {
		name := accessor.GetName()
		trimmed := util.TrimTenantIDPrefix(tenantID, name)
		accessor.SetName(trimmed)
	}
	ownerReferences := accessor.GetOwnerReferences()
	for i := range ownerReferences {
		target, err := c.ownerRefTransformer.Backward(&ownerReferences[i], tenantID)
		if err != nil {
			return err
		}
		ownerReferences[i] = *target
	}
	accessor.SetOwnerReferences(ownerReferences)
	return nil
}
