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
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apiserver/pkg/registry/generic"
	"k8s.io/apiserver/pkg/registry/rest"
	genericapiserver "k8s.io/apiserver/pkg/server"
	serverstorage "k8s.io/apiserver/pkg/server/storage"
	"k8s.io/kubernetes/pkg/api/legacyscheme"
	"k8s.io/kubernetes/pkg/master"

	"github.com/kubewharf/kubezoo/pkg/common"
)

type RESTStorageProvider struct {
	apiGroupConfig common.APIGroupConfig
}

// NewRESTStorage returns a rest storage.
func (r RESTStorageProvider) NewRESTStorage(apiResourceConfigSource serverstorage.APIResourceConfigSource, restOptionsGetter generic.RESTOptionsGetter) (genericapiserver.APIGroupInfo, bool, error) {
	scheme := legacyscheme.Scheme
	apiextensionsv1.AddToScheme(scheme)
	apiextensions.AddToScheme(scheme)
	apiGroupInfo := genericapiserver.NewDefaultAPIGroupInfo(r.apiGroupConfig.Group, scheme, runtime.NewParameterCodec(scheme), serializer.NewCodecFactory(scheme))
	for version, resources := range r.apiGroupConfig.StorageConfigs {
		storage := map[string]rest.Storage{}
		for resource, config := range resources {
			ps, err := NewTenantProxy(*config)
			if err != nil {
				return genericapiserver.APIGroupInfo{}, false, err
			}
			storage[resource] = ps
		}
		apiGroupInfo.VersionedResourcesStorageMap[version] = storage
	}
	return apiGroupInfo, true, nil
}

// NewStoragesForGV returns a rest storage for group version.
func NewStoragesForGV(cfgs map[string]*common.StorageConfig) (
	map[string]rest.Storage, error) {
	storage := map[string]rest.Storage{}
	for resource, config := range cfgs {
		proxy, err := NewTenantProxy(*config)
		if err != nil {
			return nil, err
		}
		storage[resource] = proxy
	}
	return storage, nil
}

// GroupName returns the group name of the api group.
func (r RESTStorageProvider) GroupName() string {
	return r.apiGroupConfig.Group
}

// NewRESTStorageProviders returns providers of rest storage.
func NewRESTStorageProviders(apiGroupConfigs ...common.APIGroupConfig) ([]master.RESTStorageProvider, error) {
	providers := make([]master.RESTStorageProvider, 0, len(apiGroupConfigs))
	for _, config := range apiGroupConfigs {
		providers = append(providers, RESTStorageProvider{config})
	}
	return providers, nil
}
