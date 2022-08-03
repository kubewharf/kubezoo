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

package common

import (
	"net/http"
	"net/url"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/kubewharf/kubezoo/pkg/dynamic"
)

type APIGroupConfig struct {
	Group string
	// the map is version to resource to storage config
	StorageConfigs map[string]map[string]*StorageConfig
}

type StorageConfig struct {
	Kind        schema.GroupVersionKind
	Resource    string
	Subresource string
	ShortNames  []string

	NamespaceScoped bool

	IsCustomResource bool

	IsConnecter bool

	// NewFunc returns a new instance of the type this registry returns for a
	// GET of a single object, e.g.:
	//
	// curl GET /apis/group/version/namespaces/my-ns/myresource/name-of-object
	NewFunc func() runtime.Object

	// NewListFunc returns a new list of the type this registry; it is the
	// type returned when the resource is listed, e.g.:
	//
	// curl GET /apis/group/version/namespaces/my-ns/myresource
	NewListFunc func() runtime.Object

	// dynamic client is used to communicate with upstream cluster
	DynamicClient dynamic.Interface

	Convertor ObjectConvertor

	ProxyTransport       http.RoundTripper
	UpstreamMaster       *url.URL
	GroupVersionKindFunc GroupVersionKindFunc
}

type GroupVersionKindFunc func(containingGV schema.GroupVersion) schema.GroupVersionKind
