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
	internal "k8s.io/kubernetes/pkg/apis/core"
	disinternal "k8s.io/kubernetes/pkg/apis/discovery"
)

// TestEndpointSliceTransformerForward tests the forward method of the
// EndpointSliceTransformer.
func TestEndpointSliceTransformerForward(t *testing.T) {
	portName := "http"
	protocol := internal.ProtocolTCP
	var port int32 = 80
	condition := true
	hostName := "pod-1"

	cases := []struct {
		name   string
		tenant string
		in     disinternal.EndpointSlice
		want   disinternal.EndpointSlice
	}{
		{
			name:   "test forward epslice",
			tenant: "111111",
			in: disinternal.EndpointSlice{
				TypeMeta: metav1.TypeMeta{
					Kind:       "EndpointSlice",
					APIVersion: "discovery.k8s.io/v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "my-epslice",
				},
				AddressType: disinternal.AddressTypeIPv4,
				Ports: []disinternal.EndpointPort{
					{
						Name:     &portName,
						Protocol: &protocol,
						Port:     &port,
					},
				},
				Endpoints: []disinternal.Endpoint{
					{
						Addresses: []string{
							"0.0.0.1",
						},
						Conditions: disinternal.EndpointConditions{
							Ready: &condition,
						},
						Hostname: &hostName,
						TargetRef: &internal.ObjectReference{
							Kind:      "Pod",
							Namespace: "default",
							Name:      "pod-1",
						},
					},
				},
			},
			want: disinternal.EndpointSlice{
				TypeMeta: metav1.TypeMeta{
					Kind:       "EndpointSlice",
					APIVersion: "discovery.k8s.io/v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "my-epslice",
				},
				AddressType: disinternal.AddressTypeIPv4,
				Ports: []disinternal.EndpointPort{
					{
						Name:     &portName,
						Protocol: &protocol,
						Port:     &port,
					},
				},
				Endpoints: []disinternal.Endpoint{
					{
						Addresses: []string{
							"0.0.0.1",
						},
						Conditions: disinternal.EndpointConditions{
							Ready: &condition,
						},
						Hostname: &hostName,
						TargetRef: &internal.ObjectReference{
							Kind:      "Pod",
							Namespace: "111111-default",
							Name:      "pod-1",
						},
					},
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			e := NewEndpointSliceTransformer(NewObjectReferenceTransformer(checkGroupKind))
			if _, err := e.Forward(&c.in, c.tenant); err != nil {
				t.Fatalf("failed to forward epslice, err: %+v", err)
			}
			if !reflect.DeepEqual(c.in, c.want) {
				t.Errorf("out: %+v, want: %+v", c.in, c.want)
			}
		})
	}
}

// TestEndpointSliceTransformerBackward tests the backward method of the
// EndpointSliceTransformer.
func TestEndpointSliceTransformerBackward(t *testing.T) {
	portName := "http"
	protocol := internal.ProtocolTCP
	var port int32 = 80
	condition := true
	hostName := "pod-1"

	cases := []struct {
		name   string
		tenant string
		in     disinternal.EndpointSlice
		want   disinternal.EndpointSlice
	}{
		{
			name:   "test forward epslice",
			tenant: "111111",
			in: disinternal.EndpointSlice{
				TypeMeta: metav1.TypeMeta{
					Kind:       "EndpointSlice",
					APIVersion: "discovery.k8s.io/v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "my-epslice",
				},
				AddressType: disinternal.AddressTypeIPv4,
				Ports: []disinternal.EndpointPort{
					{
						Name:     &portName,
						Protocol: &protocol,
						Port:     &port,
					},
				},
				Endpoints: []disinternal.Endpoint{
					{
						Addresses: []string{
							"0.0.0.1",
						},
						Conditions: disinternal.EndpointConditions{
							Ready: &condition,
						},
						Hostname: &hostName,
						TargetRef: &internal.ObjectReference{
							Kind:      "Pod",
							Namespace: "111111-default",
							Name:      "pod-1",
						},
					},
				},
			},
			want: disinternal.EndpointSlice{
				TypeMeta: metav1.TypeMeta{
					Kind:       "EndpointSlice",
					APIVersion: "discovery.k8s.io/v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "my-epslice",
				},
				AddressType: disinternal.AddressTypeIPv4,
				Ports: []disinternal.EndpointPort{
					{
						Name:     &portName,
						Protocol: &protocol,
						Port:     &port,
					},
				},
				Endpoints: []disinternal.Endpoint{
					{
						Addresses: []string{
							"0.0.0.1",
						},
						Conditions: disinternal.EndpointConditions{
							Ready: &condition,
						},
						Hostname: &hostName,
						TargetRef: &internal.ObjectReference{
							Kind:      "Pod",
							Namespace: "default",
							Name:      "pod-1",
						},
					},
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			e := NewEndpointSliceTransformer(NewObjectReferenceTransformer(checkGroupKind))
			if _, err := e.Backward(&c.in, c.tenant); err != nil {
				t.Fatalf("failed to backward epslice, err: %+v", err)
			}
			if !reflect.DeepEqual(c.in, c.want) {
				t.Errorf("out: %+v, want: %+v", c.in, c.want)
			}
		})
	}
}
