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
)

// EventTransformer implements the transformation between
// client and upstream server for Event resource.
type EventTransformer struct {
	objectRefTransformer ObjectReferenceTransformer
}

var _ ObjectTransformer = &EventTransformer{}

// NewEventTransformer initiates a EventTransformer which implements
// the ObjectTransformer interfaces.
func NewEventTransformer(ort ObjectReferenceTransformer) ObjectTransformer {
	return &EventTransformer{
		objectRefTransformer: ort,
	}
}

// Forward transforms tenant object reference to upstream object reference.
func (t *EventTransformer) Forward(obj runtime.Object, tenantID string) (runtime.Object, error) {
	event, ok := obj.(*internal.Event)
	if !ok {
		return nil, errors.Errorf("fail to assert the runtime object to the internal version of event")
	}

	target, err := t.objectRefTransformer.Forward(&event.InvolvedObject, tenantID)
	if err != nil {
		return nil, err
	}
	event.InvolvedObject = *target
	return event, nil
}

// Backward transforms upstream object reference to tenant object reference.
func (t *EventTransformer) Backward(obj runtime.Object, tenantID string) (runtime.Object, error) {
	event, ok := obj.(*internal.Event)
	if !ok {
		return nil, errors.Errorf("fail to assert the runtime object to the internal version of event")
	}

	target, err := t.objectRefTransformer.Backward(&event.InvolvedObject, tenantID)
	if err != nil {
		return nil, err
	}
	event.InvolvedObject = *target
	event.Message = strings.ReplaceAll(event.Message, tenantID+"-", "")
	return event, nil
}
