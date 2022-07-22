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

	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	internal "k8s.io/kubernetes/pkg/apis/core"
)

// TestPVCTranformerForward tests the forward method of the PVCTranformer.
func TestPVCTranformerForward(t *testing.T) {
	scName := "my-sc"
	volumeMode := internal.PersistentVolumeFilesystem

	cases := []struct {
		name   string
		tenant string
		in     internal.PersistentVolumeClaim
		want   internal.PersistentVolumeClaim
	}{
		{
			name:   "test forward pvc",
			tenant: "111111",
			in: internal.PersistentVolumeClaim{
				TypeMeta: metav1.TypeMeta{
					Kind:       "PersistentVolumeClaim",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "my-pvc",
				},
				Spec: internal.PersistentVolumeClaimSpec{
					AccessModes: []internal.PersistentVolumeAccessMode{
						internal.ReadWriteOnce,
					},
					Resources: internal.ResourceRequirements{
						Requests: internal.ResourceList{
							internal.ResourceStorage: resource.MustParse("20Gi"),
						},
					},
					StorageClassName: &scName,
					VolumeMode:       &volumeMode,
					VolumeName:       "pv-1",
				},
			},
			want: internal.PersistentVolumeClaim{
				TypeMeta: metav1.TypeMeta{
					Kind:       "PersistentVolumeClaim",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "my-pvc",
				},
				Spec: internal.PersistentVolumeClaimSpec{
					AccessModes: []internal.PersistentVolumeAccessMode{
						internal.ReadWriteOnce,
					},
					Resources: internal.ResourceRequirements{
						Requests: internal.ResourceList{
							internal.ResourceStorage: resource.MustParse("20Gi"),
						},
					},
					StorageClassName: &scName,
					VolumeMode:       &volumeMode,
					VolumeName:       "111111-pv-1",
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			e := NewPVCTransformer()
			if _, err := e.Forward(&c.in, c.tenant); err != nil {
				t.Fatalf("failed to forward pvc, err: %+v", err)
			}
			if !reflect.DeepEqual(c.in, c.want) {
				t.Errorf("out: %+v, want: %+v", c.in, c.want)
			}
		})
	}
}

// TestPVCTranformerBackward tests the backward method of the PVCTranformer.
func TestPVCTranformerBackward(t *testing.T) {
	scName := "my-sc"
	volumeMode := internal.PersistentVolumeFilesystem

	cases := []struct {
		name   string
		tenant string
		in     internal.PersistentVolumeClaim
		want   internal.PersistentVolumeClaim
	}{
		{
			name:   "test forward pvc",
			tenant: "111111",
			in: internal.PersistentVolumeClaim{
				TypeMeta: metav1.TypeMeta{
					Kind:       "PersistentVolumeClaim",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "my-pvc",
				},
				Spec: internal.PersistentVolumeClaimSpec{
					AccessModes: []internal.PersistentVolumeAccessMode{
						internal.ReadWriteOnce,
					},
					Resources: internal.ResourceRequirements{
						Requests: internal.ResourceList{
							internal.ResourceStorage: resource.MustParse("20Gi"),
						},
					},
					StorageClassName: &scName,
					VolumeMode:       &volumeMode,
					VolumeName:       "111111-pv-1",
				},
			},
			want: internal.PersistentVolumeClaim{
				TypeMeta: metav1.TypeMeta{
					Kind:       "PersistentVolumeClaim",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "my-pvc",
				},
				Spec: internal.PersistentVolumeClaimSpec{
					AccessModes: []internal.PersistentVolumeAccessMode{
						internal.ReadWriteOnce,
					},
					Resources: internal.ResourceRequirements{
						Requests: internal.ResourceList{
							internal.ResourceStorage: resource.MustParse("20Gi"),
						},
					},
					StorageClassName: &scName,
					VolumeMode:       &volumeMode,
					VolumeName:       "pv-1",
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			e := NewPVCTransformer()
			if _, err := e.Backward(&c.in, c.tenant); err != nil {
				t.Fatalf("failed to backward pvc, err: %+v", err)
			}
			if !reflect.DeepEqual(c.in, c.want) {
				t.Errorf("out: %+v, want: %+v", c.in, c.want)
			}
		})
	}
}
