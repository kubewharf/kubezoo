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
	disinternal "k8s.io/kubernetes/pkg/apis/discovery"
)

// EndpointSliceTransformer implements the transformation between
// client and upstream server for EndpointSlice resource.
type EndpointSliceTransformer struct {
	objectRefTransformer ObjectReferenceTransformer
}

var _ ObjectTransformer = &EventTransformer{}

// NewEndpointSliceTransformer initiates a EndpointSliceTransformer which
// implements the ObjectTransformer interfaces.
func NewEndpointSliceTransformer(ort ObjectReferenceTransformer) ObjectTransformer {
	return &EndpointSliceTransformer{
		objectRefTransformer: ort,
	}
}

// Forward transforms tenant object reference to upstream object reference.
func (t *EndpointSliceTransformer) Forward(obj runtime.Object, tenantID string) (runtime.Object, error) {
	endpointSlice, ok := obj.(*disinternal.EndpointSlice)
	if !ok {
		return nil, errors.Errorf("fail to assert the runtime object to the internal version of endpointslice")
	}

	for i := range endpointSlice.Endpoints {
		target, err := t.objectRefTransformer.Forward(endpointSlice.Endpoints[i].TargetRef, tenantID)
		if err != nil {
			return nil, err
		}
		endpointSlice.Endpoints[i].TargetRef = target
	}
	return endpointSlice, nil
}

// Backward transforms upstream object reference to tenant object reference.
func (t *EndpointSliceTransformer) Backward(obj runtime.Object, tenantID string) (runtime.Object, error) {
	endpointSlice, ok := obj.(*disinternal.EndpointSlice)
	if !ok {
		return nil, errors.Errorf("fail to assert the runtime object to the internal version of endpointslice")
	}

	for i := range endpointSlice.Endpoints {
		target, err := t.objectRefTransformer.Backward(endpointSlice.Endpoints[i].TargetRef, tenantID)
		if err != nil {
			return nil, err
		}
		endpointSlice.Endpoints[i].TargetRef = target
	}
	return endpointSlice, nil
}
