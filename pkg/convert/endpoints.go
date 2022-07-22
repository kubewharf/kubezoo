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
	"github.com/pkg/errors"

	"k8s.io/apimachinery/pkg/runtime"
	internal "k8s.io/kubernetes/pkg/apis/core"
)

// EndpointsTransformer implements the transformation between
// client and upstream server for Endpoints resource.
type EndpointsTransformer struct {
	objectRefTransformer ObjectReferenceTransformer
}

var _ ObjectTransformer = &EndpointsTransformer{}

// NewEndpointsTransformer initiates a EndpointsTransformer which implements the
// ObjectTransformer interfaces.
func NewEndpointsTransformer(ort ObjectReferenceTransformer) ObjectTransformer {
	return &EndpointsTransformer{
		objectRefTransformer: ort,
	}
}

// Forward transforms tenant object reference to upstream object reference.
func (t *EndpointsTransformer) Forward(obj runtime.Object, tenantID string) (runtime.Object, error) {
	endpoints, ok := obj.(*internal.Endpoints)
	if !ok {
		return nil, errors.Errorf("failed to assert runtime object to internal version of endpoints")
	}

	for i := range endpoints.Subsets {
		for j := range endpoints.Subsets[i].Addresses {
			target, err := t.objectRefTransformer.Forward(endpoints.Subsets[i].Addresses[j].TargetRef, tenantID)
			if err != nil {
				return nil, err
			}
			endpoints.Subsets[i].Addresses[j].TargetRef = target
		}
		for j := range endpoints.Subsets[i].NotReadyAddresses {
			target, err := t.objectRefTransformer.Forward(endpoints.Subsets[i].NotReadyAddresses[j].TargetRef, tenantID)
			if err != nil {
				return nil, err
			}
			endpoints.Subsets[i].NotReadyAddresses[j].TargetRef = target
		}
	}
	return endpoints, nil
}

// Backward transforms upstream object reference to tenant object reference.
func (t *EndpointsTransformer) Backward(obj runtime.Object, tenantID string) (runtime.Object, error) {
	endpoints, ok := obj.(*internal.Endpoints)
	if !ok {
		return nil, errors.Errorf("failed to assert runtime object to internal version of endpoints")
	}
	for i := range endpoints.Subsets {
		for j := range endpoints.Subsets[i].Addresses {
			target, err := t.objectRefTransformer.Backward(endpoints.Subsets[i].Addresses[j].TargetRef, tenantID)
			if err != nil {
				return nil, err
			}
			endpoints.Subsets[i].Addresses[j].TargetRef = target
		}
		for j := range endpoints.Subsets[i].NotReadyAddresses {
			target, err := t.objectRefTransformer.Backward(endpoints.Subsets[i].NotReadyAddresses[j].TargetRef, tenantID)
			if err != nil {
				return nil, err
			}
			endpoints.Subsets[i].NotReadyAddresses[j].TargetRef = target
		}
	}
	return endpoints, nil
}
