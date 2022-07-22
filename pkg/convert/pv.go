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
	"k8s.io/apimachinery/pkg/runtime"
	internal "k8s.io/kubernetes/pkg/apis/core"

	"github.com/kubewharf/kubezoo/pkg/util"
)

// PVTranformer implements the transformation between client and
// upstream server for PersistenceVolume resource.
type PVTranformer struct{}

var _ ObjectTransformer = &PVTranformer{}

// NewPVTransformer initiates a PVTranformer which implements
// the ObjectTransformer interfaces.
func NewPVTransformer() ObjectTransformer {
	return &PVTranformer{}
}

// Forward transforms tenant object reference to upstream object reference.
func (v *PVTranformer) Forward(obj runtime.Object, tenantID string) (runtime.Object, error) {
	// convert logic
	pv, ok := obj.(*internal.PersistentVolume)
	if !ok {
		return nil, errors.Errorf("fail to assert the runtime object to the internal version of persistentvolume")
	}

	if pv.Spec.ClaimRef != nil && len(pv.Spec.ClaimRef.Namespace) > 0 {
		pv.Spec.ClaimRef.Namespace = util.AddTenantIDPrefix(tenantID, pv.Spec.ClaimRef.Namespace)
	}
	return pv, nil
}

// Backward transforms upstream object reference to tenant object reference.
func (v *PVTranformer) Backward(obj runtime.Object, tenantID string) (runtime.Object, error) {
	pv, ok := obj.(*internal.PersistentVolume)
	if !ok {
		return nil, errors.Errorf("fail to assert the runtime object to the internal version of persistentvolume")
	}

	if pv.Spec.ClaimRef != nil && len(pv.Spec.ClaimRef.Namespace) > 0 {
		if !strings.HasPrefix(pv.Spec.ClaimRef.Namespace, tenantID) {
			return nil, errors.Errorf("invalid namespace %s in pv %s claim ref, tenant id is %s", pv.Spec.ClaimRef.Namespace, pv.Name, tenantID)
		}
		pv.Spec.ClaimRef.Namespace = util.TrimTenantIDPrefix(tenantID, pv.Spec.ClaimRef.Namespace)
	}

	return pv, nil
}
