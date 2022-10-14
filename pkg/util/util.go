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

package util

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	extensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsv1b1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	v1 "k8s.io/apiextensions-apiserver/pkg/client/listers/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metainternalversion "k8s.io/apimachinery/pkg/apis/meta/internalversion"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/apiserver/pkg/endpoints/request"

	//"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog"
)

const (
	TenantIDSeparator = "-"
	// TODO(renjingsi): move this to tenant apis and add some validations
	TenantIDLength = 6
	TenantIDKey    = "tenant"
)

// AddTenantIDPrefix add tenantId as the prefix.
func AddTenantIDPrefix(tenantID, input string) string {
	return tenantID + TenantIDSeparator + input
}

// TrimTenantIDPrefix removes tenantId prefix.
func TrimTenantIDPrefix(tenantID, input string) string {
	return strings.TrimPrefix(input, tenantID+TenantIDSeparator)
}

var invalidPrefixedNamespaceErr = fmt.Errorf("TenantID prefixed namespace must be in the form %s",
	AddTenantIDPrefix("tenantID", "namespace"))

var (
	errInvalidTenantNameLength = "tenant ID must be exactly 6 characters"
	errNonRfc1123TenantName    = "tenant ID must contain only alphanumeric characters or -, and must not start with a -"
)

func ValidateTenantName(tenantId string) *string {
	if len(tenantId) != TenantIDLength {
		return &errInvalidTenantNameLength
	}

	// RFC 1123 validation
	// The last character can be dash because tenant ID is always joined with the actual namespace name,
	// so the trailing dash will only lead to names like `tenant--default`
	for i, char := range tenantId {
		if 'a' <= char && char <= 'z' {
			continue
		}
		if '0' <= char && char <= '9' {
			continue
		}
		if i > 0 && char == '-' {
			continue
		}
		return &errNonRfc1123TenantName
	}

	return nil
}

// GetTenantIDFromNamespace get the tenantId from the prefix of namespace.
func GetTenantIDFromNamespace(namespace string) (string, error) {
	if len(namespace) < TenantIDLength+2 {
		return "", invalidPrefixedNamespaceErr
	}
	if namespace[TenantIDLength] != '-' {
		return "", invalidPrefixedNamespaceErr
	}
	return namespace[:TenantIDLength], nil
}

// AddTenantIDToUserInfo add the tenantId to the extra of userinfo.
func AddTenantIDToUserInfo(tenantID string, info user.Info) user.Info {
	extra := info.GetExtra()
	if extra == nil {
		extra = map[string][]string{}
	}
	extra[TenantIDKey] = []string{tenantID}
	return &user.DefaultInfo{
		Name:   info.GetName(),
		UID:    info.GetUID(),
		Groups: info.GetGroups(),
		Extra:  extra,
	}
}

// UpstreamObjectBelongsToTenant returns true if object belongs to tenant according to tenantID.
func UpstreamObjectBelongsToTenant(obj runtime.Object, tenantID string, isNamespaceScoped bool) bool {
	accessor, err := meta.Accessor(obj)
	if err != nil {
		klog.Errorf("failed to get accessor for object: %+v", obj)
		return false
	}

	// name of crd object is in the form: <plural>.<tenantID>-<group>
	if IsCRDObject(obj) {
		parts := strings.SplitN(accessor.GetName(), ".", 2)
		if len(parts) < 2 {
			return false
		}
		return strings.HasPrefix(parts[1], tenantID+"-")
	}

	// Todo: renjs, temporarily expose nodes for tenants to pass Conformance test
	t, err := meta.TypeAccessor(obj)
	if err != nil {
		klog.Errorf("failed to get type accessor for object: %+v", obj)
		return false
	}
	if t.GetAPIVersion() == "v1" && t.GetKind() == "Node" {
		return true
	}

	// non-crd object is namespace scoped
	if isNamespaceScoped {
		return strings.HasPrefix(accessor.GetNamespace(), tenantID+"-")
	}
	// non-crd object is cluster scoped
	return strings.HasPrefix(accessor.GetName(), tenantID+"-")
}

// IsCRDObject checks whether the input obj is a CRD object or not.
func IsCRDObject(obj runtime.Object) bool {
	switch v := obj.(type) {
	case *unstructured.Unstructured:
		t, err := meta.TypeAccessor(obj)
		if err != nil {
			klog.Errorf("failed to get type for object: %+v", v)
			return false
		}
		if (t.GetAPIVersion() == "apiextensions.k8s.io/v1" || t.GetAPIVersion() == "apiextensions.k8s.io/v1beta1") &&
			t.GetKind() == "CustomResourceDefinition" {
			return true
		}
	case *apiextensionsv1.CustomResourceDefinition:
		return true
	case *apiextensionsv1b1.CustomResourceDefinition:
		return true
	}
	return false
}

// ConvertCRDNameToUpstream convert the name of CRD with adding tenantId prefix in group.
func ConvertCRDNameToUpstream(name, tenantID string) string {
	parts := strings.SplitN(name, ".", 2)
	if len(parts) < 2 {
		klog.Errorf("invalid crd name: %s", name)
		return name
	}
	plural, group := parts[0], parts[1]
	return fmt.Sprintf("%s.%s-%s", plural, tenantID, group)
}

// ConvertTenantObjectNameToUpstream convert the object to upstream object by adding tenantId prefix.
func ConvertTenantObjectNameToUpstream(name, tenantID string, gvk schema.GroupVersionKind) string {
	if (gvk.GroupVersion().String() == "apiextensions.k8s.io/v1" || gvk.GroupVersion().String() == "apiextensions.k8s.io/v1beta1") &&
		gvk.Kind == "CustomResourceDefinition" {
		return ConvertCRDNameToUpstream(name, tenantID)
	}
	return AddTenantIDPrefix(tenantID, name)
}

// ConvertUpstreamApiGroupToTenant convert upstream the apigroup to tenant by trimming the tenantId prefix.
func ConvertUpstreamApiGroupToTenant(tenantID string, apiGroup *metav1.APIGroup) {
	if apiGroup == nil {
		return
	}
	apiGroup.Name = TrimTenantIDPrefix(tenantID, apiGroup.Name)
	for i := range apiGroup.Versions {
		apiGroup.Versions[i].GroupVersion = TrimTenantIDPrefix(tenantID, apiGroup.Versions[i].GroupVersion)
	}
	apiGroup.PreferredVersion.GroupVersion = TrimTenantIDPrefix(tenantID, apiGroup.PreferredVersion.GroupVersion)
}

// ConvertUpstreamResourceListToTenant convert upstream resource list to tenant by trimming the tenantId prefix.
func ConvertUpstreamResourceListToTenant(tenantID string, resourceList *metav1.APIResourceList) {
	if resourceList == nil {
		return
	}
	resourceList.GroupVersion = TrimTenantIDPrefix(tenantID, resourceList.GroupVersion)
	for i := range resourceList.APIResources {
		resourceList.APIResources[i].Group = TrimTenantIDPrefix(tenantID, resourceList.APIResources[i].Group)
	}
}

// GetUnstructured return Unstructured for any given kubernetes type.
func GetUnstructured(resource interface{}) (*unstructured.Unstructured, error) {
	content, err := json.Marshal(resource)
	if err != nil {
		return nil, err
	}
	unstructuredResource := &unstructured.Unstructured{}
	err = unstructuredResource.UnmarshalJSON(content)
	if err != nil {
		return nil, err
	}
	return unstructuredResource, nil
}

var groupKindNamespaced = map[metav1.GroupKind]bool{
	{Group: "", Kind: "Binding"}:                                                    true,
	{Group: "", Kind: "ComponentStatus"}:                                            false,
	{Group: "", Kind: "ConfigMap"}:                                                  true,
	{Group: "", Kind: "Endpoints"}:                                                  true,
	{Group: "", Kind: "Event"}:                                                      true,
	{Group: "", Kind: "LimitRange"}:                                                 true,
	{Group: "", Kind: "Namespace"}:                                                  false,
	{Group: "", Kind: "Node"}:                                                       false,
	{Group: "", Kind: "PersistentVolumeClaim"}:                                      true,
	{Group: "", Kind: "PersistentVolume"}:                                           false,
	{Group: "", Kind: "Pod"}:                                                        true,
	{Group: "", Kind: "PodTemplate"}:                                                true,
	{Group: "", Kind: "ReplicationController"}:                                      true,
	{Group: "", Kind: "ResourceQuota"}:                                              true,
	{Group: "", Kind: "Secret"}:                                                     true,
	{Group: "", Kind: "ServiceAccount"}:                                             true,
	{Group: "", Kind: "Service"}:                                                    true,
	{Group: "admissionregistration.k8s.io", Kind: "MutatingWebhookConfiguration"}:   false,
	{Group: "admissionregistration.k8s.io", Kind: "ValidatingWebhookConfiguration"}: false,
	{Group: "apiextensions.k8s.io", Kind: "CustomResourceDefinition"}:               false,
	{Group: "apps", Kind: "ControllerRevision"}:                                     true,
	{Group: "apps", Kind: "DaemonSet"}:                                              true,
	{Group: "apps", Kind: "Deployment"}:                                             true,
	{Group: "apps", Kind: "ReplicaSet"}:                                             true,
	{Group: "apps", Kind: "StatefulSet"}:                                            true,
	{Group: "authentication.k8s.io", Kind: "TokenReview"}:                           false,
	{Group: "authorization.k8s.io", Kind: "LocalSubjectAccessReview"}:               true,
	{Group: "authorization.k8s.io", Kind: "SelfSubjectAccessReview"}:                false,
	{Group: "authorization.k8s.io", Kind: "SelfSubjectRulesReview"}:                 false,
	{Group: "authorization.k8s.io", Kind: "SubjectAccessReview"}:                    false,
	{Group: "autoscaling", Kind: "HorizontalPodAutoscaler"}:                         true,
	{Group: "autoscaling", Kind: "Scale"}:                                           true,
	{Group: "batch", Kind: "CronJob"}:                                               true,
	{Group: "batch", Kind: "Job"}:                                                   true,
	{Group: "certificates.k8s.io", Kind: "CertificateSigningRequest"}:               false,
	{Group: "coordination.k8s.io", Kind: "Lease"}:                                   true,
	{Group: "discovery.k8s.io", Kind: "EndpointSlice"}:                              true,
	{Group: "events.k8s.io", Kind: "Event"}:                                         true,
	{Group: "extensions", Kind: "Ingress"}:                                          true,
	{Group: "networking.k8s.io", Kind: "IngressClass"}:                              false,
	{Group: "networking.k8s.io", Kind: "Ingress"}:                                   true,
	{Group: "networking.k8s.io", Kind: "NetworkPolicy"}:                             true,
	{Group: "node.k8s.io", Kind: "RuntimeClass"}:                                    false,
	{Group: "policy", Kind: "PodDisruptionBudget"}:                                  true,
	{Group: "policy", Kind: "PodSecurityPolicy"}:                                    false,
	{Group: "rbac.authorization.k8s.io", Kind: "ClusterRoleBinding"}:                false,
	{Group: "rbac.authorization.k8s.io", Kind: "ClusterRole"}:                       false,
	{Group: "rbac.authorization.k8s.io", Kind: "RoleBinding"}:                       true,
	{Group: "rbac.authorization.k8s.io", Kind: "Role"}:                              true,
	{Group: "scheduling.k8s.io", Kind: "PriorityClass"}:                             false,
	{Group: "storage.k8s.io", Kind: "CSIDriver"}:                                    false,
	{Group: "storage.k8s.io", Kind: "CSINode"}:                                      false,
	{Group: "storage.k8s.io", Kind: "StorageClass"}:                                 false,
	{Group: "storage.k8s.io", Kind: "VolumeAttachment"}:                             false,
}

// IsGroupKindNamespaced check the kind is namespace scoped or not.
func IsGroupKindNamespaced(kind metav1.GroupKind) (bool, error) {
	namespaced, ok := groupKindNamespaced[kind]
	if !ok {
		return false, fmt.Errorf("unrecognized kind: %+v", kind)
	}
	return namespaced, nil
}

// TenantIDFrom returns tenantID from ctx.
func TenantIDFrom(ctx context.Context) string {
	tenantExtra := "tenant"
	user, ok := request.UserFrom(ctx)
	if !ok {
		return ""
	}
	if len(user.GetExtra()[tenantExtra]) > 0 {
		return user.GetExtra()[tenantExtra][0]
	}
	return ""
}

// ListCRDsForTenant returns the CRDs belonged to the tenant.
func ListCRDsForTenant(tenantID string, crdLister v1.CustomResourceDefinitionLister) ([]*extensionsv1.CustomResourceDefinition, error) {
	crdList, err := crdLister.List(labels.Everything())
	if err != nil {
		return nil, err
	}
	tenantCRDs := make([]*extensionsv1.CustomResourceDefinition, 0, len(crdList))
	for _, crd := range crdList {
		if UpstreamObjectBelongsToTenant(crd, tenantID, false) {
			tenantCRDs = append(tenantCRDs, crd)
		}
	}
	return tenantCRDs, nil
}

// CheckGroupKindFunc returns whether resource of the group/kind is namespaced and whether it is custom resource group for the tenant.
type CheckGroupKindFunc func(group, kind, tenantID string, isTenantObject bool) (namespaced, customResourceGroup bool, err error)

// NewCheckGroupKindFunc returns a check function to check the group/kind type.
func NewCheckGroupKindFunc(crdLister v1.CustomResourceDefinitionLister) CheckGroupKindFunc {
	return func(group, kind, tenantID string, isTenantObject bool) (namespaced, customResourceGroup bool, err error) {
		// native group/kind
		namespaced, err = IsGroupKindNamespaced(metav1.GroupKind{Group: group, Kind: kind})
		if err == nil {
			return namespaced, false, nil
		}

		crdList, err := ListCRDsForTenant(tenantID, crdLister)
		if err != nil {
			return false, false, err
		}

		// tenant crd group/kind
		if isTenantObject {
			group = AddTenantIDPrefix(tenantID, group)
		}
		for _, crd := range crdList {
			if crd.Spec.Group == group && crd.Spec.Names.Kind == kind {
				return crd.Spec.Scope == extensionsv1.NamespaceScoped, true, nil
			}
		}
		// TODO: temporary fix for system crd
		if isTenantObject {
			// try to find system crd
			group = TrimTenantIDPrefix(tenantID, group)
			allCRDs, err := crdLister.List(labels.Everything())
			if err != nil {
				return false, false, err
			}
			for _, crd := range allCRDs {
				if crd.Spec.Group == group && crd.Spec.Names.Kind == kind {
					// system crds should be treated as custom resource for tenant
					return crd.Spec.Scope == extensionsv1.NamespaceScoped, false, nil
				}
			}
		}

		return false, false, fmt.Errorf("unregistered crd group: %s, kind: %s", group, kind)
	}
}

// TenantFrom returns the value of the tenant info on the ctx.
func TenantFrom(ctx context.Context) (string, bool) {
	tenantExtra := "tenant"
	user, ok := request.UserFrom(ctx)
	if !ok {
		return "", false
	}
	if len(user.GetExtra()[tenantExtra]) > 0 {
		return user.GetExtra()[tenantExtra][0], true
	}
	return "", false
}

// ConvertInternalListOptions converts internal versions to v1 version.
func ConvertInternalListOptions(ctx context.Context, options *metainternalversion.ListOptions, tenantID string) (*metav1.ListOptions, error) {
	var err error
	out := &metav1.ListOptions{}
	if options.FieldSelector != nil {
		fn := func(label, value string) (string, string, error) {
			if label == "involvedObject.namespace" && value != "" && tenantID != "" {
				value = tenantID + "-" + value
			}
			return label, value, nil
		}
		if options.FieldSelector, err = options.FieldSelector.Transform(fn); err != nil {
			err = errors.NewBadRequest(err.Error())
			return nil, err
		}
	}
	if err = metainternalversion.Convert_internalversion_ListOptions_To_v1_ListOptions(options, out, nil); err != nil {
		return nil, err
	}
	return out, nil
}

// FilterUnstructuredList filter the unstructures not belonged to the tenant
func FilterUnstructuredList(utdList *unstructured.UnstructuredList, tenantID string, isNamespaceScoped bool) *unstructured.UnstructuredList {
	filtered := &unstructured.UnstructuredList{
		Object: utdList.Object,
		Items:  make([]unstructured.Unstructured, 0),
	}
	for i := range utdList.Items {
		if UpstreamObjectBelongsToTenant(&utdList.Items[i], tenantID, isNamespaceScoped) {
			filtered.Items = append(filtered.Items, utdList.Items[i])
		}
	}
	return filtered
}

type FakeCRDLister struct {
	Crds []*apiextensionsv1.CustomResourceDefinition
}

func (l *FakeCRDLister) List(selector labels.Selector) (ret []*apiextensionsv1.CustomResourceDefinition, err error) {
	return l.Crds, nil
}

func (l *FakeCRDLister) Get(name string) (*apiextensionsv1.CustomResourceDefinition, error) {
	return nil, nil
}

// GetGVR returns the corresponding GVR for the given APIResource.
func GetGVR(rsrc metav1.APIResource) schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    rsrc.Group,
		Version:  rsrc.Version,
		Resource: rsrc.Name,
	}
}

// IsCRD checks if the given APIResource is the CRD.
func IsCRD(r metav1.APIResource) bool {
	return r.Group == "apiextensions.k8s.io" &&
		r.Name == "customresourcedefinitions"
}

// FlattenResourceLists flattens the given nested list and return a list of resources.
func FlattenResourceLists(resourceLists []*metav1.APIResourceList) (ret []metav1.APIResource) {
	for _, resourceList := range resourceLists {
		for _, resource := range resourceList.APIResources {
			group, version := splitGV(resourceList.GroupVersion)
			resource.Group = group
			resource.Version = version
			ret = append(ret, resource)
		}
	}
	return
}

// splitGV splits the GroupVersion string (group/version).
func splitGV(GV string) (group, version string) {
	tks := strings.Split(GV, "/")
	if len(tks) < 1 {
		return
	}
	// i.e., v1 -> core/v1
	version = tks[0]
	if len(tks) < 2 {
		return
	}
	group, version = tks[0], tks[1]
	return
}

// ContainString checks if the slice contains the string.
func ContainString(sli []string, s string) bool {
	for _, ts := range sli {
		if ts == s {
			return true
		}
	}
	return false
}

// RemoveString removes the string from the slice, if found.
func RemoveString(sli []string, s string) (ret []string) {
	for i, ts := range sli {
		if ts == s {
			sli = append(sli[:i], sli[i+1:]...)
		}
	}
	ret = sli
	return
}
