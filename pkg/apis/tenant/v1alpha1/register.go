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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var SchemeGroupVersion = schema.GroupVersion{
	Group:   "tenant.kubezoo.io",
	Version: "v1alpha1",
}
var AddToScheme = func(scheme *runtime.Scheme) error {
	metav1.AddToGroupVersion(scheme, schema.GroupVersion{
		Group:   "tenant.kubezoo.io",
		Version: "v1alpha1",
	})
	scheme.AddKnownTypes(schema.GroupVersion{
		Group:   "tenant.kubezoo.io",
		Version: "v1alpha1",
	}, &Tenant{}, &TenantList{})

	scheme.AddKnownTypes(schema.GroupVersion{
		Group:   "tenant.kubezoo.io",
		Version: runtime.APIVersionInternal,
	}, &Tenant{}, &TenantList{})

	return nil
}

// Resource takes an unqualified resource and returns a Group qualified GroupResource
func Resource(resource string) schema.GroupResource {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}
