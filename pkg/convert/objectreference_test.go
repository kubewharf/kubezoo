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
	"testing"

	"github.com/go-test/deep"
	internal "k8s.io/kubernetes/pkg/apis/core"
)

var (
	tenantObjectReferences = []*internal.ObjectReference{
		&internal.ObjectReference{
			APIVersion:      "apps/v1",
			Kind:            "Deployment",
			Namespace:       "default",
			Name:            "dp-92a152bd69",
			UID:             "bd081382-4015-11eb-9d00-b8599fe021ac",
			ResourceVersion: "sdfefef",
		},
		&internal.ObjectReference{
			APIVersion:      "v1",
			Kind:            "Namespace",
			Name:            "default",
			UID:             "da0f337b-98ca-446d-96d6-25dc3ff1ddcc",
			ResourceVersion: "sdfdefef",
		},
	}
	upstreamObjectReferences = []*internal.ObjectReference{
		&internal.ObjectReference{
			APIVersion:      "apps/v1",
			Kind:            "Deployment",
			Namespace:       "test01-default",
			Name:            "dp-92a152bd69",
			UID:             "bd081382-4015-11eb-9d00-b8599fe021ac",
			ResourceVersion: "sdfefef",
		},
		&internal.ObjectReference{
			APIVersion:      "v1",
			Kind:            "Namespace",
			Name:            "test01-default",
			UID:             "da0f337b-98ca-446d-96d6-25dc3ff1ddcc",
			ResourceVersion: "sdfdefef",
		},
	}
)

// TestObjectReferenceTransformer_Forward tests the forward method of the
// ObjectReferenceTransformer.
func TestObjectReferenceTransformer_Forward(t *testing.T) {
	transformer := NewObjectReferenceTransformer(checkGroupKind)
	for i := range tenantObjectReferences {
		got, err := transformer.Forward(tenantObjectReferences[i].DeepCopy(), tenantID)
		if err != nil {
			t.Errorf("failed to transform tenant object reference: %+v, err: %v", *tenantObjectReferences[i], err)
		}
		if diff := deep.Equal(got, upstreamObjectReferences[i]); diff != nil {
			t.Error(diff)
		}
	}
}

// TestObjectReferenceTransformer_Backward tests the backward method of the
// ObjectReferenceTransformer.
func TestObjectReferenceTransformer_Backward(t *testing.T) {
	transformer := NewObjectReferenceTransformer(checkGroupKind)
	for i := range upstreamObjectReferences {
		got, err := transformer.Backward(upstreamObjectReferences[i].DeepCopy(), tenantID)
		if err != nil {
			t.Errorf("failed to transform upstream object reference: %+v, err: %v", *upstreamObjectReferences[i], err)
		}
		if diff := deep.Equal(got, tenantObjectReferences[i]); diff != nil {
			t.Error(diff)
		}
	}
}
