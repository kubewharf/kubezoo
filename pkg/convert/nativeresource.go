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
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/kubewharf/kubezoo/pkg/common"
)

// nativeObjectConvertor implements the transformation between
// client and upstream server for native resource.
type nativeObjectConvertor struct {
	defaultConvertor       common.ObjectConvertor
	nativeKindToConvertors map[schema.GroupKind]common.ObjectConvertor
}

// NewNativeObjectConvertor initiates a nativeObjectConvertor which implements
// the ObjectConvertor interfaces.
func NewNativeObjectConvertor(defaultConvertor common.ObjectConvertor,
	kindToConvertors map[schema.GroupKind]common.ObjectConvertor) common.ObjectConvertor {
	return &nativeObjectConvertor{
		defaultConvertor:       defaultConvertor,
		nativeKindToConvertors: kindToConvertors,
	}
}

// ConvertTenantObjectToUpstreamObject convert the tenant object to
// upstream object.
func (c *nativeObjectConvertor) ConvertTenantObjectToUpstreamObject(obj runtime.Object, tenantID string, isNamespaceScoped bool) error {
	convertor := c.nativeKindToConvertors[obj.GetObjectKind().GroupVersionKind().GroupKind()]
	if convertor == nil {
		convertor = c.defaultConvertor
	}
	return convertor.ConvertTenantObjectToUpstreamObject(obj, tenantID, isNamespaceScoped)
}

// ConvertUpstreamObjectToTenantObject convert the upstream object to
// tenant object.
func (c *nativeObjectConvertor) ConvertUpstreamObjectToTenantObject(obj runtime.Object, tenantID string, isNamespaceScoped bool) error {
	convertor := c.nativeKindToConvertors[obj.GetObjectKind().GroupVersionKind().GroupKind()]
	if convertor == nil {
		convertor = c.defaultConvertor
	}
	return convertor.ConvertUpstreamObjectToTenantObject(obj, tenantID, isNamespaceScoped)
}
