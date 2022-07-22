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
	"k8s.io/klog"

	"github.com/kubewharf/kubezoo/pkg/common"
)

// ObjectTransformer transforms tenant object to/from upstream object
// NOTE for both Forward and Backward, the first input and ouput are either
// unstructured or internal object.
type ObjectTransformer interface {
	// Forward transforms the tenant object to the upstream object.
	Forward(obj runtime.Object, tenantID string) (runtime.Object, error)
	// Backward transforms the upstream object to the tenant object.
	Backward(obj runtime.Object, tenantID string) (runtime.Object, error)
}

// CrossReferenceConvertor implements the interfaces of ObjectConvertor and
// ObjectTransformer to support the cross references cases such as
// PersistenceVolume and PersistenceVolumeClaim.
type CrossReferenceConvertor struct {
	defaultConverter  common.ObjectConvertor
	objectTransformer ObjectTransformer
}

var _ common.ObjectConvertor = &CrossReferenceConvertor{}

// convert spec.volumeName field
func NewCrossReferenceConverter(c common.ObjectConvertor, objectTransformer ObjectTransformer) common.ObjectConvertor {
	return &CrossReferenceConvertor{defaultConverter: c, objectTransformer: objectTransformer}
}

// ConvertTenantObjectToUpstreamObject convert the tenant object to
// upstream object.
func (c *CrossReferenceConvertor) ConvertTenantObjectToUpstreamObject(obj runtime.Object, tenantID string, isNamespaceScoped bool) error {
	err := c.defaultConverter.ConvertTenantObjectToUpstreamObject(obj, tenantID, isNamespaceScoped)
	if err != nil {
		klog.Errorf("fail to convert tenant object to upstream object: %v", err)
		return err
	}

	// convert logic
	obj, err = c.objectTransformer.Forward(obj, tenantID)
	if err != nil {
		klog.Errorf("fail to convert tenant object to upstream object: %v", err)
		return err
	}

	return nil
}

// ConvertUpstreamObjectToTenantObject convert the upstream object to
// tenant object.
func (c *CrossReferenceConvertor) ConvertUpstreamObjectToTenantObject(obj runtime.Object, tenantID string, isNamespaceScoped bool) error {
	err := c.defaultConverter.ConvertUpstreamObjectToTenantObject(obj, tenantID, isNamespaceScoped)
	if err != nil {
		klog.Errorf("fail to convert upstream object to upstream object: %v", err)
		return err
	}

	// convert logic
	obj, err = c.objectTransformer.Backward(obj, tenantID)
	if err != nil {
		klog.Errorf("fail to convert upstream object to upstream object: %v", err)
		return err
	}

	return nil
}
