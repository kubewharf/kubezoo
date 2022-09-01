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

package proxy

import (
	"context"
	"fmt"

	v1 "k8s.io/apiextensions-apiserver/pkg/client/listers/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/discovery"

	openapi_v2 "github.com/google/gnostic/openapiv2"

	"github.com/kubewharf/kubezoo/pkg/util"
)

type DiscoveryProxy interface {
	// ServerGroups returns the supported groups for tenant, with information like supported versions and the
	// preferred version.
	ServerGroups(tenantID string) (*metav1.APIGroupList, error)
	// ServerVersionsForGroup returns the supported versions and the preferred version of a group for tenant.
	ServerVersionsForGroup(tenantID, group string) (*metav1.APIGroup, error)
	// ServerResourcesForGroupVersion returns the supported resources for a group and version for tenant.
	ServerResourcesForGroupVersion(tenantID, group, version string) (*metav1.APIResourceList, error)
	// ServerVersion retrieves and parses the server's version (git version).
	ServerVersion() (*version.Info, error)
	// OpenAPISchema fetches the open api schema using a rest client and parses the proto.
	OpenAPISchema() (*openapi_v2.Document, error)
}

// discoveryProxy implements the DiscoveryProxy interface
type discoveryProxy struct {
	// discoveryClient discover server-supported API groups,
	// versions and resources from upstream cluster.
	discoveryClient *discovery.DiscoveryClient
	// crdLister helps list CustomResourceDefinitions from upstream cluster.
	crdLister v1.CustomResourceDefinitionLister
}

func NewDiscoveryProxy(discoveryClient *discovery.DiscoveryClient,
	crdLister v1.CustomResourceDefinitionLister) (DiscoveryProxy, error) {
	if discoveryClient == nil {
		return nil, fmt.Errorf("discoveryClient is nil")
	}
	if crdLister == nil {
		return nil, fmt.Errorf("crdLister is nil")
	}
	return &discoveryProxy{discoveryClient: discoveryClient, crdLister: crdLister}, nil
}

// ServerGroups returns the supported groups for tenant, with information like supported versions and the
// preferred version.
func (dp *discoveryProxy) ServerGroups(tenantID string) (*metav1.APIGroupList, error) {
	crds, err := util.ListCRDsForTenant(tenantID, dp.crdLister)
	if err != nil {
		return nil, err
	}
	grm := util.NewCustomGroupResourcesMap(crds)
	groupList, err := dp.discoveryClient.ServerGroups()
	if err != nil {
		return nil, err
	}
	return filterAPIGroupList(groupList, grm, tenantID), nil
}

// filterAPIGroupList filter the apigroup according to the tenantId prefix.
func filterAPIGroupList(apiGroupList *metav1.APIGroupList, grm util.CustomGroupResourcesMap, tenantID string) *metav1.APIGroupList {
	if apiGroupList == nil {
		return nil
	}
	filtered := &metav1.APIGroupList{
		TypeMeta: apiGroupList.TypeMeta,
		Groups:   make([]metav1.APIGroup, 0, len(apiGroupList.Groups)),
	}

	for i := range apiGroupList.Groups {
		groupName := apiGroupList.Groups[i].Name
		// exclude the groupVersions exposed at /api
		if groupName == "" {
			continue
		}
		// native groups
		if nativeScheme.IsGroupRegistered(groupName) {
			filtered.Groups = append(filtered.Groups, apiGroupList.Groups[i])
			continue
		}
		// custom group for tenant
		if grm.HasGroup(groupName) {
			util.ConvertUpstreamApiGroupToTenant(tenantID, &apiGroupList.Groups[i])
			filtered.Groups = append(filtered.Groups, apiGroupList.Groups[i])
			continue
		}
	}
	return filtered
}

// ServerVersionsForGroup returns the supported versions and the preferred version of a group for tenant.
func (dp *discoveryProxy) ServerVersionsForGroup(tenantID, group string) (*metav1.APIGroup, error) {
	crds, err := util.ListCRDsForTenant(tenantID, dp.crdLister)
	if err != nil {
		return nil, err
	}
	grm := util.NewCustomGroupResourcesMap(crds)
	customResourceUpstreamGroup := util.AddTenantIDPrefix(tenantID, group)
	if grm.HasGroup(customResourceUpstreamGroup) {
		group = customResourceUpstreamGroup
	}

	g := &metav1.APIGroup{}
	if err := dp.discoveryClient.RESTClient().Get().AbsPath("/apis/" + group).Do(context.TODO()).Into(g); err != nil {
		return nil, err
	}
	util.ConvertUpstreamApiGroupToTenant(tenantID, g)
	return g, nil
}

// ServerResourcesForGroupVersion returns the supported resources for a group and version for tenant.
func (dp *discoveryProxy) ServerResourcesForGroupVersion(tenantID, group, version string) (*metav1.APIResourceList, error) {
	crds, err := util.ListCRDsForTenant(tenantID, dp.crdLister)
	if err != nil {
		return nil, err
	}
	grm := util.NewCustomGroupResourcesMap(crds)
	customResourceUpstreamGroup := util.AddTenantIDPrefix(tenantID, group)
	if grm.HasGroupVersion(customResourceUpstreamGroup, version) {
		group = customResourceUpstreamGroup
	}
	resourceList, err := dp.discoveryClient.ServerResourcesForGroupVersion(group + "/" + version)
	if err != nil {
		return nil, err
	}
	util.ConvertUpstreamResourceListToTenant(tenantID, resourceList)
	return resourceList, nil
}

// ServerVersion retrieves and parses the server's version (git version).
func (dp *discoveryProxy) ServerVersion() (*version.Info, error) {
	return dp.discoveryClient.ServerVersion()
}

func (dp *discoveryProxy) OpenAPISchema() (*openapi_v2.Document, error) {
	return dp.discoveryClient.OpenAPISchema()
}
