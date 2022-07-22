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
	"strings"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/kubewharf/kubezoo/pkg/util"
)

// CustomResourceTransformer implements the transformation between
// client and upstream server for custom resource.
type CustomResourceTransformer struct{}

var _ ObjectTransformer = &CustomResourceTransformer{}

// NewCustomResourceTransformer initiates a CustomResourceTransformer
// which implements the ObjectTransformer interfaces.
func NewCustomResourceTransformer() ObjectTransformer {
	return &CustomResourceTransformer{}
}

// Forward transforms tenant object reference to upstream object reference.
func (t *CustomResourceTransformer) Forward(obj runtime.Object, tenantID string) (runtime.Object, error) {
	u, ok := obj.(*unstructured.Unstructured)
	if !ok {
		return nil, errors.New("fail to assert runtime object to the unstructured object")
	}
	groupVersion := u.GetAPIVersion()
	u.SetAPIVersion(util.AddTenantIDPrefix(tenantID, groupVersion))
	return u, nil
}

// Backward transforms upstream object reference to tenant object reference.
func (t *CustomResourceTransformer) Backward(obj runtime.Object, tenantID string) (runtime.Object, error) {
	u, ok := obj.(*unstructured.Unstructured)
	if !ok {
		return nil, errors.New("fail to assert runtime object to the unstructured object")
	}
	groupVersion := u.GetAPIVersion()
	if !strings.HasPrefix(groupVersion, tenantID) {
		return nil, errors.Errorf("invalid apiVersion %s in cr %s, tenant id is %s", groupVersion, u.GetName(), tenantID)
	}
	u.SetAPIVersion(util.TrimTenantIDPrefix(tenantID, groupVersion))
	return u, nil
}
