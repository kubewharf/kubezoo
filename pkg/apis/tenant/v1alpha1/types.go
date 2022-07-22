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
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/apiserver-runtime/pkg/builder/resource"
	"sigs.k8s.io/apiserver-runtime/pkg/builder/resource/resourcestrategy"
)

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Tenant struct {
	metav1.TypeMeta `json:",inline"`
	// `metadata` is the standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	// `spec` is the specification of the desired behavior of a flow-schema.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#spec-and-status
	// +optional
	Spec TenantSpec `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
	// `status` is the current status of a flow-schema.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#spec-and-status
	// +optional
	Status TenantStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// TenantList is a list of Tenant objects.
type TenantList struct {
	metav1.TypeMeta `json:",inline"`
	// `metadata` is the standard list metadata.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata
	// +optional
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// `items` is a list of tenant
	// +listType=atomic
	Items []Tenant `json:"items" protobuf:"bytes,2,rep,name=items"`
}

// TenantSpec describes how the proxy-rule's specification looks like.
type TenantSpec struct {
	ID int32 `json:"id" protobuf:"varint,1,name=id"`
}

// TenantStatus represents the current state of a rule.
type TenantStatus struct {
	// Current state of tenant.
	Online bool `json:"online,omitempty" protobuf:"bytes,1,name=online"`
}

var _ resource.Object = &Tenant{}
var _ resourcestrategy.Validater = &Tenant{}

// GetObjectMeta returns the object metadata of tenant.
func (t *Tenant) GetObjectMeta() *metav1.ObjectMeta {
	return &t.ObjectMeta
}

// NamespaceScoped returns whether the tenant is namespace scoped or not.
func (t *Tenant) NamespaceScoped() bool {
	return false
}

// New returns a tenant object.
func (t *Tenant) New() runtime.Object {
	return &Tenant{}
}

// NewList returns a list of tenant objects.
func (t *Tenant) NewList() runtime.Object {
	return &TenantList{}
}

// GetGroupVersionResource returns the group version resource of tenant.
func (t *Tenant) GetGroupVersionResource() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    "tenant.kubezoo.io",
		Version:  "v1alpha1",
		Resource: "tenants",
	}
}

// IsStorageVersion returns whether tenant is storage version or not.
func (t *Tenant) IsStorageVersion() bool {
	return false
}

// Validate does some validation.
func (t *Tenant) Validate(ctx context.Context) field.ErrorList {
	return nil
}

var _ resource.ObjectList = &TenantList{}

// GetListMeta returns list metadata.
func (in *TenantList) GetListMeta() *metav1.ListMeta {
	return &in.ListMeta
}

// SubResourceName returns subresource name.
func (in TenantStatus) SubResourceName() string {
	return "status"
}

// Xuchen implements ObjectWithStatusSubResource interface.
var _ resource.ObjectWithStatusSubResource = &Tenant{}

// GetStatus returns the sub resource of status.
func (in *Tenant) GetStatus() resource.StatusSubResource {
	return in.Status
}

// XuchenStatus{} implements StatusSubResource interface.
var _ resource.StatusSubResource = &TenantStatus{}

// CopyTo the status.
func (in TenantStatus) CopyTo(parent resource.ObjectWithStatusSubResource) {
	parent.(*Tenant).Status = in
}
