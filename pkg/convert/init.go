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
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/kubewharf/kubezoo/pkg/common"
	"github.com/kubewharf/kubezoo/pkg/util"
)

// InitConvertors initialize native convertor and custom convertor
func InitConvertors(checkGroupKind util.CheckGroupKindFunc, listTenantCRDs ListTenantCRDsFunc) (nativeConvertor, customConvertor common.ObjectConvertor) {
	ownerReferenceTransformer := NewOwnerReferenceTransformer(checkGroupKind)
	objectReferenceTransformer := NewObjectReferenceTransformer(checkGroupKind)
	defaultConvertor := NewDefaultConvertor(ownerReferenceTransformer)
	nopeConvertor := NewNopeConvertor()

	nativeKindToConvertors := map[schema.GroupKind]common.ObjectConvertor{
		schema.GroupKind{
			Group: "",
			Kind:  "Endpoints",
		}: NewCrossReferenceConverter(defaultConvertor, NewEndpointsTransformer(objectReferenceTransformer)),
		schema.GroupKind{
			Group: "discovery.k8s.io",
			Kind:  "EndpointSlice",
		}: NewCrossReferenceConverter(defaultConvertor, NewEndpointSliceTransformer(objectReferenceTransformer)),
		schema.GroupKind{
			Group: "",
			Kind:  "Event",
		}: NewCrossReferenceConverter(defaultConvertor, NewEventTransformer(objectReferenceTransformer)),
		schema.GroupKind{
			Group: "apiextensions.k8s.io",
			Kind:  "CustomResourceDefinition",
		}: NewCRDConvertor(ownerReferenceTransformer),
		schema.GroupKind{
			Group: "",
			Kind:  "PersistentVolumeClaim",
		}: defaultConvertor,
		schema.GroupKind{
			Group: "",
			Kind:  "PersistentVolume",
		}: nopeConvertor,
		schema.GroupKind{
			Group: "storage.k8s.io",
			Kind:  "VolumeAttachment",
		}: NewCrossReferenceConverter(defaultConvertor, NewVolumeAttachmentTransformer()),
		schema.GroupKind{
			Group: "rbac.authorization.k8s.io",
			Kind:  "ClusterRole",
		}: NewCrossReferenceConverter(defaultConvertor, NewClusterRoleTransformer(listTenantCRDs)),
		schema.GroupKind{
			Group: "rbac.authorization.k8s.io",
			Kind:  "ClusterRoleBinding",
		}: NewCrossReferenceConverter(defaultConvertor, NewClusterRoleBindingTransformer()),
		schema.GroupKind{
			Group: "rbac.authorization.k8s.io",
			Kind:  "Role",
		}: NewCrossReferenceConverter(defaultConvertor, NewRoleTransformer(listTenantCRDs)),
		schema.GroupKind{
			Group: "rbac.authorization.k8s.io",
			Kind:  "RoleBinding",
		}: NewCrossReferenceConverter(defaultConvertor, NewRoleBindingTransformer()),
		schema.GroupKind{
			Group: "authentication.k8s.io",
			Kind:  "TokenReview",
		}: NewCrossReferenceConverter(defaultConvertor, NewTokenReviewTransformer()),

		// resources with nope convertor:
		schema.GroupKind{
			Group: "scheduling.k8s.io",
			Kind:  "PriorityClass",
		}: nopeConvertor,
		schema.GroupKind{
			Group: "policy",
			Kind:  "PodSecurityPolicy",
		}: nopeConvertor,
	}
	nativeConvertor = NewNativeObjectConvertor(defaultConvertor, nativeKindToConvertors)
	customConvertor = NewCrossReferenceConverter(defaultConvertor, NewCustomResourceTransformer())
	return nativeConvertor, customConvertor
}
