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

// Package app does all of the work necessary to create a Kubernetes
// APIServer by binding together the API, master and APIServer infrastructure.
// It can be configured and called directly or via the hyperkube framework.
package app

import (
	"net/http"
	"time"

	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/install"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextensionsapiserver "k8s.io/apiextensions-apiserver/pkg/apiserver"
	apiextensionsoptions "k8s.io/apiextensions-apiserver/pkg/cmd/server/options"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apiserver/pkg/features"
	"k8s.io/apiserver/pkg/registry/rest"
	genericapiserver "k8s.io/apiserver/pkg/server"
	genericoptions "k8s.io/apiserver/pkg/server/options"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	"k8s.io/apiserver/pkg/util/webhook"
	kubeexternalinformers "k8s.io/client-go/informers"

	"github.com/kubewharf/kubezoo/cmd/app/options"
	"github.com/kubewharf/kubezoo/pkg/common"
	"github.com/kubewharf/kubezoo/pkg/proxy"
)

var (
	Scheme = runtime.NewScheme()
	Codecs = serializer.NewCodecFactory(Scheme)

	// if you modify this, make sure you update the crEncoder
	unversionedVersion = schema.GroupVersion{Group: "", Version: "v1"}
	unversionedTypes   = []runtime.Object{
		&metav1.Status{},
		&metav1.WatchEvent{},
		&metav1.APIVersions{},
		&metav1.APIGroupList{},
		&metav1.APIGroup{},
		&metav1.APIResourceList{},
	}
)

func init() {
	install.Install(Scheme)

	// we need to add the options to empty v1
	metav1.AddToGroupVersion(Scheme, schema.GroupVersion{Group: "", Version: "v1"})

	Scheme.AddUnversionedTypes(unversionedVersion, unversionedTypes...)
}

var v1StorageConfig = map[string]*common.StorageConfig{
	"customresourcedefinitions": {
		Kind: v1.SchemeGroupVersion.
			WithKind("CustomResourceDefinition"),
		Resource:        "customresourcedefinitions",
		ShortNames:      []string{"crd", "crds"},
		NamespaceScoped: false,
		NewFunc: func() runtime.Object {
			return &apiextensions.CustomResourceDefinition{}
		},
		NewListFunc: func() runtime.Object {
			return &apiextensions.CustomResourceDefinitionList{}
		},
		TableConvertor: rest.NewDefaultTableConvertor(apiextensions.Resource("customresourcedefinitions")),
	},
	"customresourcedefinitions/status": {
		Kind: v1.SchemeGroupVersion.
			WithKind("CustomResourceDefinition"),
		Resource:        "customresourcedefinitions",
		ShortNames:      []string{"crd", "crds"},
		Subresource:     "status",
		NamespaceScoped: false,
		NewFunc: func() runtime.Object {
			return &apiextensions.CustomResourceDefinition{}
		},
	},
}

var v1beta1StorageConfig = map[string]*common.StorageConfig{
	"customresourcedefinitions": {
		Kind: v1beta1.SchemeGroupVersion.
			WithKind("CustomResourceDefinition"),
		Resource:        "customresourcedefinitions",
		ShortNames:      []string{"crd", "crds"},
		NamespaceScoped: false,
		NewFunc: func() runtime.Object {
			return &apiextensions.CustomResourceDefinition{}
		},
		NewListFunc: func() runtime.Object {
			return &apiextensions.CustomResourceDefinitionList{}
		},
		TableConvertor: rest.NewDefaultTableConvertor(apiextensions.Resource("customresourcedefinitions")),
	},
	"customresourcedefinitions/status": {
		Kind: v1beta1.SchemeGroupVersion.
			WithKind("CustomResourceDefinition"),
		Resource:        "customresourcedefinitions",
		ShortNames:      []string{"crd", "crds"},
		Subresource:     "status",
		NamespaceScoped: false,
		NewFunc: func() runtime.Object {
			return &apiextensions.CustomResourceDefinition{}
		},
	},
}

func createAPIExtensionsConfig(
	kubeAPIServerConfig genericapiserver.Config,
	externalInformers kubeexternalinformers.SharedInformerFactory,
	commandOptions *options.ServerRunOptions,
	masterCount int,
	serviceResolver webhook.ServiceResolver,
	authResolverWrapper webhook.AuthenticationInfoResolverWrapper,
) (*apiextensionsapiserver.Config, error) {
	// make a shallow copy to let us twiddle a few things
	// most of the config actually remains the same.  We only need to mess with a couple items related to the particulars of the apiextensions
	genericConfig := kubeAPIServerConfig
	genericConfig.PostStartHooks = map[string]genericapiserver.PostStartHookConfigEntry{}
	genericConfig.RESTOptionsGetter = nil

	// override genericConfig.AdmissionControl with apiextensions' scheme,
	// because apiextentions apiserver should use its own scheme to convert resources.
	//err := commandOptions.Admission.ApplyTo(
	//	&genericConfig,
	//	externalInformers,
	//	genericConfig.LoopbackClientConfig,
	//	feature.DefaultFeatureGate,
	//	pluginInitializers...)
	//if err != nil {
	//	return nil, err
	//}

	// copy the etcd options so we don't mutate originals.
	etcdOptions := *commandOptions.Etcd
	etcdOptions.StorageConfig.Paging = utilfeature.DefaultFeatureGate.Enabled(features.APIListChunking)
	etcdOptions.StorageConfig.Codec = apiextensionsapiserver.Codecs.LegacyCodec(v1beta1.SchemeGroupVersion, v1.SchemeGroupVersion)
	etcdOptions.StorageConfig.EncodeVersioner = runtime.NewMultiGroupVersioner(v1beta1.SchemeGroupVersion, schema.GroupKind{Group: v1beta1.GroupName})
	genericConfig.RESTOptionsGetter = &genericoptions.SimpleRestOptionsFactory{Options: etcdOptions}

	// override MergedResourceConfig with apiextensions defaults and registry
	if err := commandOptions.APIEnablement.ApplyTo(
		&genericConfig,
		apiextensionsapiserver.DefaultAPIResourceConfigSource(),
		apiextensionsapiserver.Scheme); err != nil {
		return nil, err
	}

	apiextensionsConfig := &apiextensionsapiserver.Config{
		GenericConfig: &genericapiserver.RecommendedConfig{
			Config:                genericConfig,
			SharedInformerFactory: externalInformers,
		},
		ExtraConfig: apiextensionsapiserver.ExtraConfig{
			CRDRESTOptionsGetter: apiextensionsoptions.NewCRDRESTOptionsGetter(etcdOptions),
			MasterCount:          masterCount,
			AuthResolverWrapper:  authResolverWrapper,
			ServiceResolver:      serviceResolver,
		},
	}

	// we need to clear the poststarthooks so we don't add them multiple times to all the servers (that fails)
	apiextensionsConfig.GenericConfig.PostStartHooks = map[string]genericapiserver.PostStartHookConfigEntry{}

	return apiextensionsConfig, nil
}

func createKubeAPIExtensionsServer(apiextensionsConfig *apiextensionsapiserver.Config, delegateAPIServer genericapiserver.DelegationTarget, proxyConfig *ProxyConfig) (*apiextensionsapiserver.CustomResourceDefinitions, error) {
	c := apiextensionsConfig.Complete()
	genericServer, err := c.GenericConfig.New("kube-zoo-apiextensions-server", delegateAPIServer)
	if err != nil {
		return nil, err
	}

	s := &apiextensionsapiserver.CustomResourceDefinitions{
		GenericAPIServer: genericServer,
		Informers:        proxyConfig.crdInformers,
	}

	apiGroupInfo := genericapiserver.NewDefaultAPIGroupInfo(apiextensions.GroupName, Scheme, metav1.ParameterCodec, Codecs)

	for resource := range v1beta1StorageConfig {
		proxyConfig.ApplyToStorage(v1beta1StorageConfig[resource])
	}
	storage, err := proxy.NewStoragesForGV(v1beta1StorageConfig)
	if err != nil {
		return nil, err
	}
	apiGroupInfo.VersionedResourcesStorageMap[v1beta1.SchemeGroupVersion.Version] = storage

	for resource := range v1StorageConfig {
		proxyConfig.ApplyToStorage(v1StorageConfig[resource])
	}
	storage, err = proxy.NewStoragesForGV(v1StorageConfig)
	if err != nil {
		return nil, err
	}
	apiGroupInfo.VersionedResourcesStorageMap[v1.SchemeGroupVersion.Version] = storage

	if err := s.GenericAPIServer.InstallAPIGroup(&apiGroupInfo); err != nil {
		return nil, err
	}

	delegateHandler := delegateAPIServer.UnprotectedHandler()
	if delegateHandler == nil {
		delegateHandler = http.NotFoundHandler()
	}

	crdHandler, err := NewCustomResourceDefinitionHandler(
		s.Informers.Apiextensions().V1().CustomResourceDefinitions(),
		delegateHandler,
		c.ExtraConfig.CRDRESTOptionsGetter,
		c.GenericConfig.AdmissionControl,
		c.ExtraConfig.ServiceResolver,
		c.ExtraConfig.AuthResolverWrapper,
		c.ExtraConfig.MasterCount,
		s.GenericAPIServer.Authorizer,
		c.GenericConfig.RequestTimeout,
		time.Duration(c.GenericConfig.MinRequestTimeout)*time.Second,
		apiGroupInfo.StaticOpenAPISpec,
		c.GenericConfig.MaxRequestBodyBytes,
		proxyConfig,
	)
	if err != nil {
		return nil, err
	}
	s.GenericAPIServer.Handler.NonGoRestfulMux.Handle("/apis", crdHandler)
	s.GenericAPIServer.Handler.NonGoRestfulMux.HandlePrefix("/apis/", crdHandler)

	s.GenericAPIServer.AddPostStartHookOrDie("start-apiextensions-informers", func(context genericapiserver.PostStartHookContext) error {
		s.Informers.Start(context.StopCh)
		return nil
	})

	// we don't want to report healthy until we can handle all CRDs that have already been registered.  Waiting for the informer
	// to sync makes sure that the lister will be valid before we begin.  There may still be races for CRDs added after startup,
	// but we won't go healthy until we can handle the ones already present.
	s.GenericAPIServer.AddPostStartHookOrDie("upstream-crd-informer-synced", func(context genericapiserver.PostStartHookContext) error {
		return wait.PollImmediateUntil(100*time.Millisecond, func() (bool, error) {
			return s.Informers.Apiextensions().V1().CustomResourceDefinitions().Informer().HasSynced(), nil
		}, context.StopCh)
	})

	return s, nil
}
