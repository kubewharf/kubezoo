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
)

// TestEndpointsTransformerForward tests the forward method of the
// EndpointsTransformer.
func TestEndpointsTransformerForward(t *testing.T) {
	nodeName := "node-1"

	cases := []struct {
		name   string
		tenant string
		in     internal.Endpoints
		want   internal.Endpoints
	}{
		{
			name:   "test forward ep",
			tenant: "111111",
			in: internal.Endpoints{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Endpoints",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "my-ep",
				},
				Subsets: []internal.EndpointSubset{
					{
						Addresses: []internal.EndpointAddress{
							{
								IP:       "0.0.0.1",
								NodeName: &nodeName,
								TargetRef: &internal.ObjectReference{
									Kind:      "Pod",
									Namespace: "default",
									Name:      "pod-1",
								},
							},
						},
						NotReadyAddresses: []internal.EndpointAddress{
							{
								IP:       "0.0.0.2",
								NodeName: &nodeName,
								TargetRef: &internal.ObjectReference{
									Kind:      "Pod",
									Namespace: "default",
									Name:      "pod-2",
								},
							},
						},
						Ports: []internal.EndpointPort{
							{
								Port:     80,
								Protocol: internal.ProtocolTCP,
							},
						},
					},
				},
			},
			want: internal.Endpoints{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Endpoints",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "my-ep",
				},
				Subsets: []internal.EndpointSubset{
					{
						Addresses: []internal.EndpointAddress{
							{
								IP:       "0.0.0.1",
								NodeName: &nodeName,
								TargetRef: &internal.ObjectReference{
									Kind:      "Pod",
									Namespace: "111111-default",
									Name:      "pod-1",
								},
							},
						},
						NotReadyAddresses: []internal.EndpointAddress{
							{
								IP:       "0.0.0.2",
								NodeName: &nodeName,
								TargetRef: &internal.ObjectReference{
									Kind:      "Pod",
									Namespace: "111111-default",
									Name:      "pod-2",
								},
							},
						},
						Ports: []internal.EndpointPort{
							{
								Port:     80,
								Protocol: internal.ProtocolTCP,
							},
						},
					},
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			e := NewEndpointsTransformer(NewObjectReferenceTransformer(checkGroupKind))
			if _, err := e.Forward(&c.in, c.tenant); err != nil {
				t.Fatalf("failed to forward ep, err: %+v", err)
			}
			if !reflect.DeepEqual(c.in, c.want) {
				t.Errorf("out: %+v, want: %+v", c.in, c.want)
			}
		})
	}
}

// TestEndpointsTransformerBackward tests the backward method of the
// EndpointsTransformer.
func TestEndpointsTransformerBackward(t *testing.T) {
	nodeName := "node-1"

	cases := []struct {
		name   string
		tenant string
		in     internal.Endpoints
		want   internal.Endpoints
	}{
		{
			name:   "test backward ep",
			tenant: "111111",
			in: internal.Endpoints{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Endpoints",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "my-ep",
				},
				Subsets: []internal.EndpointSubset{
					{
						Addresses: []internal.EndpointAddress{
							{
								IP:       "0.0.0.1",
								NodeName: &nodeName,
								TargetRef: &internal.ObjectReference{
									Kind:      "Pod",
									Namespace: "111111-default",
									Name:      "pod-1",
								},
							},
						},
						NotReadyAddresses: []internal.EndpointAddress{
							{
								IP:       "0.0.0.2",
								NodeName: &nodeName,
								TargetRef: &internal.ObjectReference{
									Kind:      "Pod",
									Namespace: "111111-default",
									Name:      "pod-2",
								},
							},
						},
						Ports: []internal.EndpointPort{
							{
								Port:     80,
								Protocol: internal.ProtocolTCP,
							},
						},
					},
				},
			},
			want: internal.Endpoints{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Endpoints",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "my-ep",
				},
				Subsets: []internal.EndpointSubset{
					{
						Addresses: []internal.EndpointAddress{
							{
								IP:       "0.0.0.1",
								NodeName: &nodeName,
								TargetRef: &internal.ObjectReference{
									Kind:      "Pod",
									Namespace: "default",
									Name:      "pod-1",
								},
							},
						},
						NotReadyAddresses: []internal.EndpointAddress{
							{
								IP:       "0.0.0.2",
								NodeName: &nodeName,
								TargetRef: &internal.ObjectReference{
									Kind:      "Pod",
									Namespace: "default",
									Name:      "pod-2",
								},
							},
						},
						Ports: []internal.EndpointPort{
							{
								Port:     80,
								Protocol: internal.ProtocolTCP,
							},
						},
					},
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			e := NewEndpointsTransformer(NewObjectReferenceTransformer(checkGroupKind))
			if _, err := e.Backward(&c.in, c.tenant); err != nil {
				t.Fatalf("failed to backward ep, err: %+v", err)
			}
			if !reflect.DeepEqual(c.in, c.want) {
				t.Errorf("out: %+v, want: %+v", c.in, c.want)
			}
		})
	}
}
