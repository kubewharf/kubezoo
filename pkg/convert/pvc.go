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

	"github.com/kubewharf/kubezoo/pkg/util"

	"github.com/pkg/errors"

	"k8s.io/apimachinery/pkg/runtime"
	internal "k8s.io/kubernetes/pkg/apis/core"
)

// PVCTranformer implements the transformation between client and
// upstream server for PersistenceVolumeClaim resource.
type PVCTranformer struct{}

var _ ObjectTransformer = &PVCTranformer{}

// NewPVCTransformer initiates a PVCTranformer which implements
// the ObjectTransformer interfaces.
func NewPVCTransformer() ObjectTransformer {
	return &PVCTranformer{}
}

// Forward transforms tenant object reference to upstream object reference.
func (v *PVCTranformer) Forward(obj runtime.Object, tenantID string) (runtime.Object, error) {
	pvc, ok := obj.(*internal.PersistentVolumeClaim)
	if !ok {
		return nil, errors.Errorf("fail to assert the runtime object to the internal version of persistentvolumeclaim")
	}
	if len(pvc.Spec.VolumeName) > 0 {
		pvc.Spec.VolumeName = util.AddTenantIDPrefix(tenantID, pvc.Spec.VolumeName)
	}

	return pvc, nil
}

// Backward transforms upstream object reference to tenant object reference.
func (v *PVCTranformer) Backward(obj runtime.Object, tenantID string) (runtime.Object, error) {
	pvc, ok := obj.(*internal.PersistentVolumeClaim)
	if !ok {
		return nil, errors.Errorf("fail to assert the runtime object to the internal version of persistentvolumeclaim")
	}
	if len(pvc.Spec.VolumeName) > 0 {
		if !strings.HasPrefix(pvc.Spec.VolumeName, tenantID) {
			return nil, errors.Errorf("invalid pv name %s in pvc %s, tenant id is %s", pvc.Spec.VolumeName, pvc.Spec.VolumeName, tenantID)
		}
		pvc.Spec.VolumeName = util.TrimTenantIDPrefix(tenantID, pvc.Spec.VolumeName)
	}

	return pvc, nil
}
