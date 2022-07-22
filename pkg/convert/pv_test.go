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

// TestPVTranformerForward tests the forward method of the PVTranformer.
func TestPVTranformerForward(t *testing.T) {
	cases := []struct {
		name   string
		tenant string
		in     internal.PersistentVolume
		want   internal.PersistentVolume
	}{
		{
			name:   "test forward pv",
			tenant: "111111",
			in: internal.PersistentVolume{
				TypeMeta: metav1.TypeMeta{
					Kind:       "PersistentVolume",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "my-pv",
				},
				Spec: internal.PersistentVolumeSpec{
					Capacity: internal.ResourceList{
						internal.ResourceStorage: resource.MustParse("10Gi"),
					},
					ClaimRef: &internal.ObjectReference{
						Namespace: "default",
						Name:      "pvc-2",
					},
				},
			},
			want: internal.PersistentVolume{
				TypeMeta: metav1.TypeMeta{
					Kind:       "PersistentVolume",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "my-pv",
				},
				Spec: internal.PersistentVolumeSpec{
					Capacity: internal.ResourceList{
						internal.ResourceStorage: resource.MustParse("10Gi"),
					},
					ClaimRef: &internal.ObjectReference{
						Namespace: "111111-default",
						Name:      "pvc-2",
					},
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			e := NewPVTransformer()
			if _, err := e.Forward(&c.in, c.tenant); err != nil {
				t.Fatalf("failed to forward pv, err: %+v", err)
			}
			if !reflect.DeepEqual(c.in, c.want) {
				t.Errorf("out: %+v, want: %+v", c.in, c.want)
			}
		})
	}
}

// TestPVTranformerBackward tests the backward method of the PVTranformer.
func TestPVTranformerBackward(t *testing.T) {
	cases := []struct {
		name   string
		tenant string
		in     internal.PersistentVolume
		want   internal.PersistentVolume
	}{
		{
			name:   "test backward pv",
			tenant: "111111",
			in: internal.PersistentVolume{
				TypeMeta: metav1.TypeMeta{
					Kind:       "PersistentVolume",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "my-pv",
				},
				Spec: internal.PersistentVolumeSpec{
					Capacity: internal.ResourceList{
						internal.ResourceStorage: resource.MustParse("10Gi"),
					},
					ClaimRef: &internal.ObjectReference{
						Namespace: "111111-default",
						Name:      "pvc-2",
					},
				},
			},
			want: internal.PersistentVolume{
				TypeMeta: metav1.TypeMeta{
					Kind:       "PersistentVolume",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "my-pv",
				},
				Spec: internal.PersistentVolumeSpec{
					Capacity: internal.ResourceList{
						internal.ResourceStorage: resource.MustParse("10Gi"),
					},
					ClaimRef: &internal.ObjectReference{
						Namespace: "default",
						Name:      "pvc-2",
					},
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			e := NewPVTransformer()
			if _, err := e.Backward(&c.in, c.tenant); err != nil {
				t.Fatalf("failed to backward pv, err: %+v", err)
			}
			if !reflect.DeepEqual(c.in, c.want) {
				t.Errorf("out: %+v, want: %+v", c.in, c.want)
			}
		})
	}
}
