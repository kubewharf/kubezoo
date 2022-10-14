/*
Copyright 2020 The Bytedance Authors.

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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ClusterResourceQuotaSpec defines the desired state of ClusterResourceQuota
type ClusterResourceQuotaSpec struct {
	corev1.ResourceQuotaSpec `json:",inline" protobuf:"bytes,1,opt,name=resourceQuotaSpec"`
	// A label query over a set of resources, in this case namespaces.
	// +optional
	NamepsaceSelector *metav1.LabelSelector `json:"namespaceSelector,omitempty" protobuf:"bytes,2,opt,name=namespaceSelector"`
	// namespaces specifies which namespaces the cluster resource quota applies to.
	// +optional
	Namespaces []string `json:"namespaces,omitempty" protobuf:"bytes,3,rep,name=namespaces"`
}

// ClusterResourceQuotaStatus defines the observed state of ClusterResourceQuota
type ClusterResourceQuotaStatus struct {
	corev1.ResourceQuotaStatus `json:",inline" protobuf:"bytes,1,opt,name=resourceQuotaStatus"`
}

// +genclient
// +genclient:nonNamespaced
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:storageversion

// ClusterResourceQuota is the Schema for the clusterresourcequota API
type ClusterResourceQuota struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Spec   ClusterResourceQuotaSpec   `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
	Status ClusterResourceQuotaStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true

// ClusterResourceQuotaList contains a list of ClusterResourceQuota
type ClusterResourceQuotaList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Items           []ClusterResourceQuota `json:"items" protobuf:"bytes,2,rep,name=items"`
}
