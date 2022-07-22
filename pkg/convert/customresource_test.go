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

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/kubewharf/kubezoo/pkg/util"
)

// TestCustomResourceTransformerForward tests the forward method of the
// CustomResourceTransformer.
func TestCustomResourceTransformerForward(t *testing.T) {
	tenant := "111111"
	apiVersion := "mygroup/v1"

	un := unstructured.Unstructured{}
	un.SetAPIVersion(apiVersion)
	un.SetKind("Mykind")
	un.SetNamespace("mynamespace")
	un.SetName("myname")

	c := NewCustomResourceTransformer()
	obj, err := c.Forward(&un, tenant)
	if err != nil {
		t.Errorf("Failed to forward with err %s", err)
	}
	newUn, ok := obj.(*unstructured.Unstructured)
	if !ok {
		t.Errorf("Failed to convert to unstructred.")
	}
	if newUn.GetAPIVersion() != tenant+util.TenantIDSeparator+apiVersion {
		t.Errorf("Unexpected api version.")
	}
}

// TestCustomResourceTransformerBackward tests the backward method of the
// CustomResourceTransformer.
func TestCustomResourceTransformerBackward(t *testing.T) {
	tenant := "111111"
	apiVersion := tenant + util.TenantIDSeparator + "mygroup/v1"

	un := unstructured.Unstructured{}
	un.SetAPIVersion(apiVersion)
	un.SetKind("Mykind")
	un.SetNamespace("mynamespace")
	un.SetName("myname")

	c := NewCustomResourceTransformer()
	obj, err := c.Backward(&un, tenant)
	if err != nil {
		t.Errorf("Failed to forward with err %s", err)
	}
	newUn, ok := obj.(*unstructured.Unstructured)
	if !ok {
		t.Errorf("Failed to convert to unstructred.")
	}
	if tenant+util.TenantIDSeparator+newUn.GetAPIVersion() != apiVersion {
		t.Errorf("Unexpected api version.")
	}
}
