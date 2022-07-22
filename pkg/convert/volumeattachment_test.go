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
	"reflect"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	storageinternal "k8s.io/kubernetes/pkg/apis/storage"
)

// TestVolumeAttachmentTransformerForward tests the forward method of the
// VolumeAttachmentTransformer.
func TestVolumeAttachmentTransformerForward(t *testing.T) {
	pvName := "my-pv"
	prefixedPvName := "111111-my-pv"

	cases := []struct {
		name   string
		tenant string
		in     storageinternal.VolumeAttachment
		want   storageinternal.VolumeAttachment
	}{
		{
			name:   "test forward VolumeAttachment",
			tenant: "111111",
			in: storageinternal.VolumeAttachment{
				TypeMeta: metav1.TypeMeta{
					Kind:       "VolumeAttachment",
					APIVersion: "storage.k8s.io/v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "my-volume-attachment",
				},
				Spec: storageinternal.VolumeAttachmentSpec{
					Attacher: "my-csi.kubezoo.io",
					Source: storageinternal.VolumeAttachmentSource{
						PersistentVolumeName: &pvName,
					},
					NodeName: "node-1",
				},
			},
			want: storageinternal.VolumeAttachment{
				TypeMeta: metav1.TypeMeta{
					Kind:       "VolumeAttachment",
					APIVersion: "storage.k8s.io/v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "my-volume-attachment",
				},
				Spec: storageinternal.VolumeAttachmentSpec{
					Attacher: "my-csi.kubezoo.io",
					Source: storageinternal.VolumeAttachmentSource{
						PersistentVolumeName: &prefixedPvName,
					},
					NodeName: "node-1",
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			e := NewVolumeAttachmentTransformer()
			if _, err := e.Forward(&c.in, c.tenant); err != nil {
				t.Fatalf("failed to forward VolumeAttachment, err: %+v", err)
			}
			if !reflect.DeepEqual(c.in, c.want) {
				t.Errorf("out: %+v, want: %+v", c.in, c.want)
			}
		})
	}
}

// TestVolumeAttachmentTransformerBackward tests the backward method of the
// VolumeAttachmentTransformer.
func TestVolumeAttachmentTransformerBackward(t *testing.T) {
	pvName := "my-pv"
	prefixedPvName := "111111-my-pv"

	cases := []struct {
		name   string
		tenant string
		in     storageinternal.VolumeAttachment
		want   storageinternal.VolumeAttachment
	}{
		{
			name:   "test backward VolumeAttachment",
			tenant: "111111",
			in: storageinternal.VolumeAttachment{
				TypeMeta: metav1.TypeMeta{
					Kind:       "VolumeAttachment",
					APIVersion: "storage.k8s.io/v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "my-volume-attachment",
				},
				Spec: storageinternal.VolumeAttachmentSpec{
					Attacher: "my-csi.kubezoo.io",
					Source: storageinternal.VolumeAttachmentSource{
						PersistentVolumeName: &prefixedPvName,
					},
					NodeName: "node-1",
				},
			},
			want: storageinternal.VolumeAttachment{
				TypeMeta: metav1.TypeMeta{
					Kind:       "VolumeAttachment",
					APIVersion: "storage.k8s.io/v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "my-volume-attachment",
				},
				Spec: storageinternal.VolumeAttachmentSpec{
					Attacher: "my-csi.kubezoo.io",
					Source: storageinternal.VolumeAttachmentSource{
						PersistentVolumeName: &pvName,
					},
					NodeName: "node-1",
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			e := NewVolumeAttachmentTransformer()
			if _, err := e.Backward(&c.in, c.tenant); err != nil {
				t.Fatalf("failed to backward VolumeAttachment, err: %+v", err)
			}
			if !reflect.DeepEqual(c.in, c.want) {
				t.Errorf("out: %+v, want: %+v", c.in, c.want)
			}
		})
	}
}
