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
	storageinternal "k8s.io/kubernetes/pkg/apis/storage"

	"github.com/kubewharf/kubezoo/pkg/util"
)

// VolumeAttachmentTransformer implements the transformation between client
// and upstream server for VolumeAttachment resource.
type VolumeAttachmentTransformer struct{}

var _ ObjectTransformer = &VolumeAttachmentTransformer{}

// NewVolumeAttachmentTransformer initiates a VolumeAttachmentTransformer
// which implements the ObjectTransformer interfaces.
func NewVolumeAttachmentTransformer() ObjectTransformer {
	return &VolumeAttachmentTransformer{}
}

// Forward transforms tenant object reference to upstream object reference.
func (v *VolumeAttachmentTransformer) Forward(obj runtime.Object, tenantID string) (runtime.Object, error) {
	volumeAttachment, ok := obj.(*storageinternal.VolumeAttachment)
	if !ok {
		return nil, errors.Errorf("fail to assert the runtime object to the internal version of volumeattachment")
	}

	if volumeAttachment.Spec.Source.PersistentVolumeName != nil && len(*volumeAttachment.Spec.Source.PersistentVolumeName) > 0 {
		*volumeAttachment.Spec.Source.PersistentVolumeName = util.AddTenantIDPrefix(tenantID, *volumeAttachment.Spec.Source.PersistentVolumeName)
	}
	return volumeAttachment, nil
}

// Backward transforms upstream object reference to tenant object reference.
func (v *VolumeAttachmentTransformer) Backward(obj runtime.Object, tenantID string) (runtime.Object, error) {
	volumeAttachment, ok := obj.(*storageinternal.VolumeAttachment)
	if !ok {
		return nil, errors.Errorf("fail to assert the runtime object to the internal version of volumeattachment")
	}

	if volumeAttachment.Spec.Source.PersistentVolumeName != nil && len(*volumeAttachment.Spec.Source.PersistentVolumeName) > 0 {
		if !strings.HasPrefix(*volumeAttachment.Spec.Source.PersistentVolumeName, tenantID) {
			return nil, errors.Errorf("invalid pv name %s in volume attachment %s, tenant id is %s", *volumeAttachment.Spec.Source.PersistentVolumeName, volumeAttachment.Name, tenantID)
		}
		*volumeAttachment.Spec.Source.PersistentVolumeName = util.TrimTenantIDPrefix(tenantID, *volumeAttachment.Spec.Source.PersistentVolumeName)
	}
	return volumeAttachment, nil
}
