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
	"github.com/go-test/deep"
	"github.com/kubewharf/kubezoo/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

// checkGroupKind checks whether NATIVE group/kind is namespaced, which means customResourceGroup is always false.
// It is only used for build unit tests
func checkGroupKind(group, kind, tenantID string, isTenantObject bool) (namespaced, customResourceGroup bool, err error) {
	namespaced, err = util.IsGroupKindNamespaced(metav1.GroupKind{Group: group, Kind: kind})
	return
}

var (
	tenantID              = "test01"
	tenantOwnerReferences = []*metav1.OwnerReference{
		&metav1.OwnerReference{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
			Name:       "dp-92a152bd69",
			UID:        "bd081382-4015-11eb-9d00-b8599fe021ac",
		},
		&metav1.OwnerReference{
			APIVersion: "v1",
			Kind:       "Namespace",
			Name:       "default",
			UID:        "da0f337b-98ca-446d-96d6-25dc3ff1ddcc",
		},
	}
	upstreamOwnerReferences = []*metav1.OwnerReference{
		&metav1.OwnerReference{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
			Name:       "dp-92a152bd69",
			UID:        "bd081382-4015-11eb-9d00-b8599fe021ac",
		},
		&metav1.OwnerReference{
			APIVersion: "v1",
			Kind:       "Namespace",
			Name:       "test01-default",
			UID:        "da0f337b-98ca-446d-96d6-25dc3ff1ddcc",
		},
	}
)

// TestOwnerReferenceTransformer_Forward tests the forward method of the
// OwnerReferenceTransformer.
func TestOwnerReferenceTransformer_Forward(t *testing.T) {
	transformer := NewOwnerReferenceTransformer(checkGroupKind)
	for i := range tenantOwnerReferences {
		got, err := transformer.Forward(tenantOwnerReferences[i].DeepCopy(), tenantID)
		if err != nil {
			t.Errorf("failed to transform tenant owner reference: %+v, err: %v", *tenantOwnerReferences[i], err)
		}
		if diff := deep.Equal(got, upstreamOwnerReferences[i]); diff != nil {
			t.Error(diff)
		}
	}
}

// TestOwnerReferenceTransformer_Backward tests the backward method of the
// OwnerReferenceTransformer.
func TestOwnerReferenceTransformer_Backward(t *testing.T) {
	transformer := NewOwnerReferenceTransformer(checkGroupKind)
	for i := range upstreamOwnerReferences {
		got, err := transformer.Backward(upstreamOwnerReferences[i].DeepCopy(), tenantID)
		if err != nil {
			t.Errorf("failed to transform upstream owner reference: %+v, err: %v", *upstreamOwnerReferences[i], err)
		}
		if diff := deep.Equal(got, tenantOwnerReferences[i]); diff != nil {
			t.Error(diff)
		}
	}
}
