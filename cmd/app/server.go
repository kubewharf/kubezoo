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
	stdx509 "crypto/x509"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/kubewharf/kubezoo/pkg/controller"
	"github.com/kubewharf/kubezoo/pkg/generated/clientset/versioned"
	"github.com/kubewharf/kubezoo/pkg/generated/informers/externalversions"
	"github.com/spf13/cobra"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	externalinformer "k8s.io/apiextensions-apiserver/pkg/client/informers/externalversions"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	util_net "k8s.io/apimachinery/pkg/util/net"
	"k8s.io/apimachinery/pkg/util/sets"
	utilwait "k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apiserver/pkg/authentication/authenticator"
	"k8s.io/apiserver/pkg/authentication/request/union"
	"k8s.io/apiserver/pkg/authentication/request/x509"
	"k8s.io/apiserver/pkg/authentication/user"
	genericapifilters "k8s.io/apiserver/pkg/endpoints/filters"
	"k8s.io/apiserver/pkg/registry/generic"
	"k8s.io/apiserver/pkg/server"
	genericapiserver "k8s.io/apiserver/pkg/server"
	"k8s.io/apiserver/pkg/server/filters"
	genericfilters "k8s.io/apiserver/pkg/server/filters"
	serveroptions "k8s.io/apiserver/pkg/server/options"
	serverstorage "k8s.io/apiserver/pkg/server/storage"
	"k8s.io/apiserver/pkg/storage/etcd3/preflight"
	"k8s.io/apiserver/pkg/util/webhook"
	clidiscovery "k8s.io/client-go/discovery"
	clientgoinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/keyutil"
	cliflag "k8s.io/component-base/cli/flag"
	utilflag "k8s.io/component-base/cli/flag"
	"k8s.io/component-base/cli/globalflag"
	"k8s.io/component-base/metrics"
	_ "k8s.io/component-base/metrics/prometheus/workqueue" // for workqueue metric registration
	"k8s.io/component-base/term"
	"k8s.io/component-base/version"
	"k8s.io/component-base/version/verflag"
	"k8s.io/klog"
	aggregatorapiserver "k8s.io/kube-aggregator/pkg/apiserver"
	"k8s.io/kubernetes/pkg/api/legacyscheme"
	"k8s.io/kubernetes/pkg/capabilities"
	master "k8s.io/kubernetes/pkg/controlplane"
	"k8s.io/kubernetes/pkg/controlplane/reconcilers"
	"k8s.io/kubernetes/pkg/kubeapiserver"
	kubeauthenticator "k8s.io/kubernetes/pkg/kubeapiserver/authenticator"
	kubeoptions "k8s.io/kubernetes/pkg/kubeapiserver/options"
	"k8s.io/kubernetes/pkg/routes"
	"k8s.io/kubernetes/pkg/serviceaccount"

	"github.com/kubewharf/kubezoo/cmd/app/options"
	_ "github.com/kubewharf/kubezoo/pkg/apis/tenant/install"
	"github.com/kubewharf/kubezoo/pkg/common"
	"github.com/kubewharf/kubezoo/pkg/convert"
	"github.com/kubewharf/kubezoo/pkg/dynamic"
	tenantfilters "github.com/kubewharf/kubezoo/pkg/filters"
	"github.com/kubewharf/kubezoo/pkg/proxy"
	tenantrest "github.com/kubewharf/kubezoo/pkg/rest"
	"github.com/kubewharf/kubezoo/pkg/util"
)

const (
	etcdRetryLimit    = 60
	etcdRetryInterval = 1 * time.Second
)

// NewAPIServerCommand creates a *cobra.Command object with default parameters
func NewAPIServerCommand() *cobra.Command {
	s := options.NewServerRunOptions()
	cmd := &cobra.Command{
		Use: "kube-zoo",
		Long: `The Kubernetes API server validates and configures data
for the api objects which include pods, services, replicationcontrollers, and
others. The API Server services REST operations and provides the frontend to the
cluster's shared state through which all other components interact.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			verflag.PrintAndExitIfRequested()
			utilflag.PrintFlags(cmd.Flags())

			// set default options
			completedOptions, err := Complete(s)
			if err != nil {
				return err
			}

			// validate options
			if errs := completedOptions.Validate(); len(errs) != 0 {
				return utilerrors.NewAggregate(errs)
			}

			return Run(completedOptions, genericapiserver.SetupSignalHandler())
		},
	}

	fs := cmd.Flags()
	namedFlagSets := s.Flags()
	verflag.AddFlags(namedFlagSets.FlagSet("global"))
	globalflag.AddGlobalFlags(namedFlagSets.FlagSet("global"), cmd.Name())
	options.AddCustomGlobalFlags(namedFlagSets.FlagSet("generic"))
	for _, f := range namedFlagSets.FlagSets {
		fs.AddFlagSet(f)
	}

	usageFmt := "Usage:\n  %s\n"
	cols, _, _ := term.TerminalSize(cmd.OutOrStdout())
	cmd.SetUsageFunc(func(cmd *cobra.Command) error {
		fmt.Fprintf(cmd.OutOrStderr(), usageFmt, cmd.UseLine())
		cliflag.PrintSections(cmd.OutOrStderr(), namedFlagSets, cols)
		return nil
	})
	cmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		fmt.Fprintf(cmd.OutOrStdout(), "%s\n\n"+usageFmt, cmd.Long, cmd.UseLine())
		cliflag.PrintSections(cmd.OutOrStdout(), namedFlagSets, cols)
	})

	return cmd
}

// Run runs the specified APIServer.  This should never exit.
func Run(completeOptions completedServerRunOptions, stopCh <-chan struct{}) error {
	// To help debugging, immediately log version
	klog.Infof("Version: %+v", version.Get())

	server, err := CreateServerChain(completeOptions, stopCh)
	if err != nil {
		return err
	}

	prepared := server.GenericAPIServer.PrepareRun()

	return prepared.Run(stopCh)
}

// CreateServerChain creates the apiservers connected via delegation.
func CreateServerChain(completedOptions completedServerRunOptions, stopCh <-chan struct{}) (*master.Instance, error) {
	kubeAPIServerConfig, _, serviceResolver, proxyConfig, controlPlaneConfig, err := CreateKubeAPIServerConfig(completedOptions)
	if err != nil {
		return nil, err
	}

	// If additional API servers are added, they should be gated.
	apiExtensionsConfig, err := createAPIExtensionsConfig(*kubeAPIServerConfig.GenericConfig, kubeAPIServerConfig.ExtraConfig.VersionedInformers, completedOptions.ServerRunOptions, completedOptions.MasterCount,
		serviceResolver, webhook.NewDefaultAuthenticationInfoResolverWrapper(nil, kubeAPIServerConfig.GenericConfig.EgressSelector, kubeAPIServerConfig.GenericConfig.LoopbackClientConfig,
			kubeAPIServerConfig.GenericConfig.TracerProvider))
	if err != nil {
		return nil, err
	}
	apiExtensionsServer, err := createKubeAPIExtensionsServer(apiExtensionsConfig, genericapiserver.NewEmptyDelegate(), proxyConfig)
	if err != nil {
		return nil, err
	}

	kubeZooServer, err := CreateKubeZooServer(kubeAPIServerConfig, apiExtensionsServer.GenericAPIServer, proxyConfig, controlPlaneConfig)
	if err != nil {
		return nil, err
	}
	return kubeZooServer, nil
}

// InstallLegacyAPI will install the legacy APIs if they are enabled.
func InstallLegacyAPI(m *master.Instance,
	apiResourceConfigSource serverstorage.APIResourceConfigSource,
	c *master.CompletedConfig,
	restOptionsGetter generic.RESTOptionsGetter,
	legacyConfig common.APIGroupConfig) error {
	legacyProviders, err := proxy.NewRESTStorageProviders(legacyConfig)
	if err != nil {
		return err
	}

	restStorageBuilder := legacyProviders[0]
	groupName := restStorageBuilder.GroupName()
	if !apiResourceConfigSource.AnyResourceForGroupEnabled(groupName) {
		klog.V(1).Infof("Skipping disabled API group %q.", groupName)
		return nil
	}
	apiGroupInfo, err := restStorageBuilder.NewRESTStorage(
		apiResourceConfigSource, restOptionsGetter)
	if err != nil {
		return fmt.Errorf("problem initializing API group %q : %v",
			groupName, err)
	}
	klog.V(1).Infof("Enabling API group %q.", groupName)

	if postHookProvider, ok := restStorageBuilder.(genericapiserver.PostStartHookProvider); ok {
		name, hook, err := postHookProvider.PostStartHook()
		if err != nil {
			klog.Fatalf("Error building PostStartHook: %v", err)
		}
		m.GenericAPIServer.AddPostStartHookOrDie(name, hook)
	}

	if err := m.GenericAPIServer.InstallLegacyAPIGroup(
		genericapiserver.DefaultLegacyAPIPrefix, &apiGroupInfo); err != nil {
		return fmt.Errorf("error in registering group versions: %v", err)
	}
	return nil
}

// CreateKubeZooServer creates and wires a workable kube-zoo-apiserver
func CreateKubeZooServer(kubeAPIServerConfig *master.Config,
	delegateAPIServer genericapiserver.DelegationTarget,
	proxyConfig *ProxyConfig,
	controlPlaneConfig *ControlPlaneConfig) (*master.Instance, error) {
	c := kubeAPIServerConfig.Complete()
	// disable admission
	c.GenericConfig.AdmissionControl = nil
	s, err := c.GenericConfig.New("kube-zoo-server", delegateAPIServer)
	if err != nil {
		return nil, err
	}

	if c.ExtraConfig.EnableLogsSupport {
		routes.Logs{}.Install(s.Handler.GoRestfulContainer)
	}
	m := &master.Instance{
		GenericAPIServer:          s,
		ClusterAuthenticationInfo: c.ExtraConfig.ClusterAuthenticationInfo,
	}

	proxyConfig.ApplyToGroup(&legacyGroup)
	if err := InstallLegacyAPI(
		m, kubeAPIServerConfig.ExtraConfig.APIResourceConfigSource,
		&c, c.GenericConfig.RESTOptionsGetter, legacyGroup); err != nil {
		return nil, err
	}

	for i := range nonLegacyGroups {
		proxyConfig.ApplyToGroup(&nonLegacyGroups[i])
	}
	providers, err := proxy.NewRESTStorageProviders(nonLegacyGroups...)
	if err != nil {
		return nil, err
	}

	providers = append(providers, tenantrest.RESTStorageProvider{})

	if err := m.InstallAPIs(kubeAPIServerConfig.ExtraConfig.APIResourceConfigSource, kubeAPIServerConfig.GenericConfig.RESTOptionsGetter, providers...); err != nil {
		return nil, err
	}

	m.GenericAPIServer.AddPostStartHookOrDie("start-tenant-controller", func(context genericapiserver.PostStartHookContext) error {
		go controller.Run(make(chan struct{}),
			controlPlaneConfig.tenantInformers.Tenant().V1alpha1().Tenants().Informer(),
			controlPlaneConfig.tenantClient.TenantV1alpha1(),
			proxyConfig.typedClientSet,
			proxyConfig.clientCAFile,
			proxyConfig.clientCAKeyFile,
			proxyConfig.proxyBindAddress,
			proxyConfig.proxySecurePort)
		return nil
	})
	m.GenericAPIServer.AddPostStartHookOrDie("tenant-informer-synced", func(context genericapiserver.PostStartHookContext) error {
		return utilwait.PollImmediateUntil(100*time.Millisecond, func() (bool, error) {
			return controlPlaneConfig.tenantInformers.Tenant().V1alpha1().Tenants().Informer().HasSynced(), nil
		}, context.StopCh)
	})

	return m, nil
}

// CreateKubeAPIServerConfig creates all the resources for running the API server, but runs none of them
func CreateKubeAPIServerConfig(
	s completedServerRunOptions,
) (
	*master.Config,
	*genericapiserver.DeprecatedInsecureServingInfo,
	aggregatorapiserver.ServiceResolver,
	*ProxyConfig,
	*ControlPlaneConfig,
	error,
) {
	genericConfig, insecureServingInfo, serviceResolver, _, storageFactory, proxyConfig, controlPlaneConfig, err := buildGenericConfig(s.ServerRunOptions)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	schemes := []string{"http", "https"}
	etcdEndpoint := s.Etcd.StorageConfig.Transport.ServerList[0]
	for _, scheme := range schemes {
		etcdEndpoint = strings.TrimPrefix(etcdEndpoint, scheme+"://")
	}

	if _, port, err := net.SplitHostPort(etcdEndpoint); err == nil && port != "0" && len(port) != 0 {
		etcdConnection := preflight.EtcdConnection{ServerList: s.Etcd.StorageConfig.Transport.ServerList}
		if err := utilwait.PollImmediate(etcdRetryInterval, etcdRetryLimit*etcdRetryInterval, etcdConnection.CheckEtcdServers); err != nil {
			return nil, nil, nil, nil, nil, fmt.Errorf("error waiting for etcd connection: %v", err)
		}

	}

	capabilities.Initialize(capabilities.Capabilities{
		AllowPrivileged: s.AllowPrivileged,
		PrivilegedSources: capabilities.PrivilegedSources{
			HostNetworkSources: []string{},
			HostPIDSources:     []string{},
			HostIPCSources:     []string{},
		},
		PerConnectionBandwidthLimitBytesPerSec: s.MaxConnectionBytesPerSec,
	})

	if len(s.ShowHiddenMetricsForVersion) > 0 {
		metrics.SetShowHidden()
	}

	serviceIPRange, apiServerServiceIP, err := master.ServiceIPRange(s.PrimaryServiceClusterIPRange)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	// defaults to empty range and ip
	var secondaryServiceIPRange net.IPNet
	// process secondary range only if provided by user
	if s.SecondaryServiceClusterIPRange.IP != nil {
		secondaryServiceIPRange, _, err = master.ServiceIPRange(s.SecondaryServiceClusterIPRange)
		if err != nil {
			return nil, nil, nil, nil, nil, err
		}
	}

	config := &master.Config{
		GenericConfig: genericConfig,
		ExtraConfig: master.ExtraConfig{
			APIResourceConfigSource: storageFactory.APIResourceConfigSource,
			StorageFactory:          storageFactory,
			EventTTL:                s.EventTTL,
			KubeletClientConfig:     s.KubeletConfig,
			EnableLogsSupport:       s.EnableLogsHandler,
			ServiceIPRange:          serviceIPRange,
			APIServerServiceIP:      apiServerServiceIP,
			SecondaryServiceIPRange: secondaryServiceIPRange,

			APIServerServicePort: 443,

			ServiceNodePortRange:      s.ServiceNodePortRange,
			KubernetesServiceNodePort: s.KubernetesServiceNodePort,

			EndpointReconcilerType: reconcilers.Type(s.EndpointReconcilerType),
			MasterCount:            s.MasterCount,

			ServiceAccountIssuer:        s.ServiceAccountIssuer,
			ServiceAccountMaxExpiration: s.ServiceAccountTokenMaxExpiration,
		},
	}

	clientCAProvider, err := s.Authentication.ClientCert.GetClientCAContentProvider()
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	config.ExtraConfig.ClusterAuthenticationInfo.ClientCA = clientCAProvider

	requestHeaderConfig, err := s.Authentication.RequestHeader.ToAuthenticationRequestHeaderConfig()
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	if requestHeaderConfig != nil {
		config.ExtraConfig.ClusterAuthenticationInfo.RequestHeaderCA = requestHeaderConfig.CAContentProvider
		config.ExtraConfig.ClusterAuthenticationInfo.RequestHeaderAllowedNames = requestHeaderConfig.AllowedClientNames
		config.ExtraConfig.ClusterAuthenticationInfo.RequestHeaderExtraHeaderPrefixes = requestHeaderConfig.ExtraHeaderPrefixes
		config.ExtraConfig.ClusterAuthenticationInfo.RequestHeaderGroupHeaders = requestHeaderConfig.GroupHeaders
		config.ExtraConfig.ClusterAuthenticationInfo.RequestHeaderUsernameHeaders = requestHeaderConfig.UsernameHeaders
	}

	// Load the public keys.
	var pubKeys []interface{}
	for _, f := range s.Authentication.ServiceAccounts.KeyFiles {
		keys, err := keyutil.PublicKeysFromFile(f)
		if err != nil {
			return nil, nil, nil, nil, nil, fmt.Errorf("failed to parse key file %q: %v", f, err)
		}
		pubKeys = append(pubKeys, keys...)
	}
	// Plumb the required metadata through ExtraConfig.
	config.ExtraConfig.ServiceAccountIssuerURL = s.Authentication.ServiceAccounts.Issuers[0]
	config.ExtraConfig.ServiceAccountJWKSURI = s.Authentication.ServiceAccounts.JWKSURI
	config.ExtraConfig.ServiceAccountPublicKeys = pubKeys

	return config, insecureServingInfo, serviceResolver, proxyConfig, controlPlaneConfig, nil
}

type ControlPlaneConfig struct {
	tenantClient    versioned.Interface
	tenantInformers externalversions.SharedInformerFactory
}

func buildControlPlaneConfig(loopConfig *rest.Config) (*ControlPlaneConfig, error) {
	tenantClient, err := versioned.NewForConfig(loopConfig)
	if err != nil {
		return nil, err
	}
	tenantInformers := externalversions.NewSharedInformerFactory(tenantClient, 5*time.Minute)
	return &ControlPlaneConfig{
		tenantClient:    tenantClient,
		tenantInformers: tenantInformers,
	}, nil
}

type ProxyConfig struct {
	dynamicClient   dynamic.Interface
	discoveryClient *clidiscovery.DiscoveryClient
	crdClient       *apiextensions.Clientset
	typedClientSet  kubernetes.Interface

	crdInformers externalinformer.SharedInformerFactory

	nativeConvertor common.ObjectConvertor
	customConvertor common.ObjectConvertor

	proxyTransport http.RoundTripper
	upstreamMaster *url.URL

	proxyBindAddress string
	proxySecurePort  int

	clientCAFile    string
	clientCAKeyFile string
}

func (c *ProxyConfig) ApplyToGroup(group *common.APIGroupConfig) {
	for version := range group.StorageConfigs {
		for resource := range group.StorageConfigs[version] {
			c.ApplyToStorage(group.StorageConfigs[version][resource])
		}
	}
}

func (c *ProxyConfig) ApplyToStorage(config *common.StorageConfig) {
	config.DynamicClient = c.dynamicClient
	config.ProxyTransport = c.proxyTransport
	config.UpstreamMaster = c.upstreamMaster
	if config.IsCustomResource {
		config.Convertor = c.customConvertor
	} else {
		config.Convertor = c.nativeConvertor
	}
}

func buildProxyConfig(o *options.ProxyOptions) (*ProxyConfig, error) {
	upstreamConfig, err := clientcmd.BuildConfigFromFlags(o.UpstreamMaster, "")
	if err != nil {
		return nil, err
	}
	upstreamConfig.CAFile = o.ProxyClientCAFile
	upstreamConfig.KeyFile = o.ProxyClientKeyFile
	upstreamConfig.CertFile = o.ProxyClientCertFile
	upstreamConfig.QPS = o.ProxyClientQPS
	upstreamConfig.Burst = o.ProxyClientBurst
	dynamicClient, err := dynamic.NewForConfig(upstreamConfig)
	if err != nil {
		return nil, err
	}
	discoveryClient, err := clidiscovery.NewDiscoveryClientForConfig(upstreamConfig)
	if err != nil {
		return nil, err
	}
	crdClient, err := apiextensions.NewForConfig(upstreamConfig)
	if err != nil {
		return nil, err
	}
	typedClientSet, err := kubernetes.NewForConfig(upstreamConfig)
	if err != nil {
		return nil, err
	}

	crdInformers := externalinformer.NewSharedInformerFactory(crdClient, 5*time.Minute)
	crdLister := crdInformers.Apiextensions().V1().CustomResourceDefinitions().Lister()
	checkGroupKind := util.NewCheckGroupKindFunc(crdLister)
	listTenantCRDs := convert.ListTenantCRDsFunc(func(tenantID string) ([]*apiextensionsv1.CustomResourceDefinition, error) {
		return util.ListCRDsForTenant(tenantID, crdLister)
	})
	nativeConvertor, customConvertor := convert.InitConvertors(checkGroupKind, listTenantCRDs)

	// construct transport for connect proxy round trip
	proxyTransport, err := rest.TransportFor(upstreamConfig)
	if err != nil {
		return nil, err
	}
	tlsConfig, err := util_net.TLSClientConfig(proxyTransport)
	if err == nil && tlsConfig != nil {
		// since http2 doesn't support websocket, we need to disable http2 when using websocket
		if supportsHTTP11(tlsConfig.NextProtos) {
			tlsConfig.NextProtos = []string{"http/1.1"}
		}
	}
	upstreamMaster, err := url.Parse(o.UpstreamMaster)
	if err != nil {
		return nil, err
	}

	return &ProxyConfig{
		dynamicClient:    dynamicClient,
		discoveryClient:  discoveryClient,
		crdClient:        crdClient,
		typedClientSet:   typedClientSet,
		crdInformers:     crdInformers,
		nativeConvertor:  nativeConvertor,
		customConvertor:  customConvertor,
		proxyTransport:   proxyTransport,
		upstreamMaster:   upstreamMaster,
		proxyBindAddress: o.BindAddress,
		proxySecurePort:  o.SecurePort,
		clientCAFile:     o.ClientCAFile,
		clientCAKeyFile:  o.ClientCAKeyFile,
	}, nil
}

// copy from https://github.com/kubernetes/apimachinery/blob/master/pkg/util/proxy/dial.go.
func supportsHTTP11(nextProtos []string) bool {
	if len(nextProtos) == 0 {
		return true
	}
	for _, proto := range nextProtos {
		if proto == "http/1.1" {
			return true
		}
	}
	return false
}

// BuildGenericConfig takes the master server options and produces the genericapiserver.Config associated with it
func buildGenericConfig(
	s *options.ServerRunOptions,
) (
	genericConfig *genericapiserver.Config,
	insecureServingInfo *genericapiserver.DeprecatedInsecureServingInfo,
	serviceResolver aggregatorapiserver.ServiceResolver,
	admissionPostStartHook genericapiserver.PostStartHookFunc,
	storageFactory *serverstorage.DefaultStorageFactory,
	proxyConfig *ProxyConfig,
	controlPlaneConfig *ControlPlaneConfig,
	lastErr error,
) {
	genericConfig = genericapiserver.NewConfig(legacyscheme.Codecs)
	// install resource config without any resource
	genericConfig.MergedResourceConfig = serverstorage.NewResourceConfig()

	proxyConfig, lastErr = buildProxyConfig(s.Proxy)
	if lastErr != nil {
		return
	}

	var discoveryProxy proxy.DiscoveryProxy
	discoveryProxy, lastErr = proxy.NewDiscoveryProxy(proxyConfig.discoveryClient,
		proxyConfig.crdInformers.Apiextensions().V1().CustomResourceDefinitions().Lister())
	if lastErr != nil {
		return
	}
	genericConfig.BuildHandlerChainFunc = NewBuildHandlerChanFunc(discoveryProxy)

	if lastErr = s.GenericServerRunOptions.ApplyTo(genericConfig); lastErr != nil {
		return
	}

	if lastErr = s.SecureServing.ApplyTo(&genericConfig.SecureServing, &genericConfig.LoopbackClientConfig); lastErr != nil {
		return
	}
	if lastErr = s.Features.ApplyTo(genericConfig); lastErr != nil {
		return
	}
	if lastErr = s.APIEnablement.ApplyTo(genericConfig, master.DefaultAPIResourceConfigSource(), legacyscheme.Scheme); lastErr != nil {
		return
	}
	// enable kubezoo
	genericConfig.MergedResourceConfig.EnableVersions(tenantrest.SchemeGroupVersion)
	if lastErr = s.EgressSelector.ApplyTo(genericConfig); lastErr != nil {
		return
	}

	// todo: generate open api definitions
	//genericConfig.OpenAPIConfig = genericapiserver.DefaultOpenAPIConfig(generatedopenapi.GetOpenAPIDefinitions, openapinamer.NewDefinitionNamer(legacyscheme.Scheme, extensionsapiserver.Scheme, aggregatorscheme.Scheme))
	//genericConfig.OpenAPIConfig.Info.Title = "Kubernetes"
	genericConfig.LongRunningFunc = filters.BasicLongRunningRequestCheck(
		sets.NewString("watch", "proxy"),
		sets.NewString("attach", "exec", "proxy", "log", "portforward"),
	)

	kubeVersion := version.Get()
	genericConfig.Version = &kubeVersion

	storageFactoryConfig := kubeapiserver.NewStorageFactoryConfig()
	storageFactoryConfig.APIResourceConfig = genericConfig.MergedResourceConfig
	completedStorageFactoryConfig, err := storageFactoryConfig.Complete(s.Etcd)
	if err != nil {
		lastErr = err
		return
	}
	storageFactory, lastErr = completedStorageFactoryConfig.New()
	if lastErr != nil {
		return
	}

	if genericConfig.EgressSelector != nil {
		storageFactory.StorageConfig.Transport.EgressLookup = genericConfig.EgressSelector.Lookup
	}
	if lastErr = s.Etcd.ApplyWithStorageFactoryTo(storageFactory, genericConfig); lastErr != nil {
		return
	}

	// Use protobufs for self-communication.
	// Since not every generic apiserver has to support protobufs, we
	// cannot default to it in generic apiserver and need to explicitly
	// set it in kube-apiserver.
	genericConfig.LoopbackClientConfig.ContentConfig.ContentType = "application/vnd.kubernetes.protobuf"
	// Disable compression for self-communication, since we are going to be
	// on a fast local network
	genericConfig.LoopbackClientConfig.DisableCompression = true

	controlPlaneConfig, lastErr = buildControlPlaneConfig(genericConfig.LoopbackClientConfig)
	if lastErr != nil {
		return
	}

	if lastErr = applyAuthenticationOptions(s.Authentication, genericConfig); lastErr != nil {
		return
	}

	ac, _ := s.Authentication.ToAuthenticationConfig()
	if ac.ClientCAContentProvider != nil {
		// append the authentication handler that will extract tenant ID from
		// the x509 certificate
		ta := x509.NewDynamic(ac.ClientCAContentProvider.VerifyOptions,
			CommonNameUserConversion)
		genericConfig.Authentication.Authenticator = union.New(ta,
			genericConfig.Authentication.Authenticator)
	}

	genericConfig.Authorization.Authorizer, genericConfig.RuleResolver, err = s.Authorization.ToAuthorizationConfig(nil).New()
	if err != nil {
		lastErr = fmt.Errorf("invalid authorization config: %v", err)
		return
	}
	return
}

func applyAuthenticationOptions(o *kubeoptions.BuiltInAuthenticationOptions, genericConfig *server.Config) error {
	authenticatorConfig, err := o.ToAuthenticationConfig()
	if err != nil {
		return err
	}

	authInfo := &genericConfig.Authentication
	secureServing := genericConfig.SecureServing
	//openAPIConfig := genericConfig.OpenAPIConfig
	if authenticatorConfig.ClientCAContentProvider != nil {
		if err = authInfo.ApplyClientCert(authenticatorConfig.ClientCAContentProvider, secureServing); err != nil {
			return fmt.Errorf("unable to load client CA file: %v", err)
		}
	}
	if authenticatorConfig.RequestHeaderConfig != nil && authenticatorConfig.RequestHeaderConfig.CAContentProvider != nil {
		if err = authInfo.ApplyClientCert(authenticatorConfig.RequestHeaderConfig.CAContentProvider, secureServing); err != nil {
			return fmt.Errorf("unable to load client CA file: %v", err)
		}
	}

	authInfo.APIAudiences = o.APIAudiences
	if o.ServiceAccounts != nil && len(o.ServiceAccounts.Issuers) > 0 && o.ServiceAccounts.Issuers[0] != "" && len(o.APIAudiences) == 0 {
		authInfo.APIAudiences = authenticator.Audiences{o.ServiceAccounts.Issuers[0]}
	}
	authInfo.Authenticator, _, err = authenticatorConfig.New()
	return err
}

// completedServerRunOptions is a private wrapper that enforces a call of Complete() before Run can be invoked.
type completedServerRunOptions struct {
	*options.ServerRunOptions
}

// Complete set default ServerRunOptions.
// Should be called after kube-apiserver flags parsed.
func Complete(s *options.ServerRunOptions) (completedServerRunOptions, error) {
	var options completedServerRunOptions
	// set defaults
	if err := s.GenericServerRunOptions.DefaultAdvertiseAddress(s.SecureServing.SecureServingOptions); err != nil {
		return options, err
	}
	if err := s.GenericServerRunOptions.DefaultAdvertiseAddress(s.SecureServing.SecureServingOptions); err != nil {
		return options, err
	}

	// process s.ServiceClusterIPRange from list to Primary and Secondary
	// we process secondary only if provided by user
	apiServerServiceIP, primaryServiceIPRange, secondaryServiceIPRange, err := getServiceIPAndRanges(s.ServiceClusterIPRanges)
	if err != nil {
		return options, err
	}
	s.PrimaryServiceClusterIPRange = primaryServiceIPRange
	s.SecondaryServiceClusterIPRange = secondaryServiceIPRange

	if err := s.SecureServing.MaybeDefaultWithSelfSignedCerts(s.GenericServerRunOptions.AdvertiseAddress.String(), []string{"kubernetes.default.svc", "kubernetes.default", "kubernetes"}, []net.IP{apiServerServiceIP}); err != nil {
		return options, fmt.Errorf("error creating self-signed certificates: %v", err)
	}

	if len(s.GenericServerRunOptions.ExternalHost) == 0 {
		if len(s.GenericServerRunOptions.AdvertiseAddress) > 0 {
			s.GenericServerRunOptions.ExternalHost = s.GenericServerRunOptions.AdvertiseAddress.String()
		} else {
			if hostname, err := os.Hostname(); err == nil {
				s.GenericServerRunOptions.ExternalHost = hostname
			} else {
				return options, fmt.Errorf("error finding host name: %v", err)
			}
		}
		klog.Infof("external host was not specified, using %v", s.GenericServerRunOptions.ExternalHost)
	}

	s.Authentication.ApplyAuthorization(s.Authorization)

	// Use (ServiceAccountSigningKeyFile != "") as a proxy to the user enabling
	// TokenRequest functionality. This defaulting was convenient, but messed up
	// a lot of people when they rotated their serving cert with no idea it was
	// connected to their service account keys. We are taking this opportunity to
	// remove this problematic defaulting.
	if s.ServiceAccountSigningKeyFile == "" {
		// Default to the private server key for service account token signing
		if len(s.Authentication.ServiceAccounts.KeyFiles) == 0 && s.SecureServing.ServerCert.CertKey.KeyFile != "" {
			if kubeauthenticator.IsValidServiceAccountKeyFile(s.SecureServing.ServerCert.CertKey.KeyFile) {
				s.Authentication.ServiceAccounts.KeyFiles = []string{s.SecureServing.ServerCert.CertKey.KeyFile}
			} else {
				klog.Warning("No TLS key provided, service account token authentication disabled")
			}
		}
	}

	if s.ServiceAccountSigningKeyFile != "" && len(s.Authentication.ServiceAccounts.Issuers) > 0 &&
		s.Authentication.ServiceAccounts.Issuers[0] != "" {
		sk, err := keyutil.PrivateKeyFromFile(s.ServiceAccountSigningKeyFile)
		if err != nil {
			return options, fmt.Errorf("failed to parse service-account-issuer-key-file: %v", err)
		}
		if s.Authentication.ServiceAccounts.MaxExpiration != 0 {
			lowBound := time.Hour
			upBound := time.Duration(1<<32) * time.Second
			if s.Authentication.ServiceAccounts.MaxExpiration < lowBound ||
				s.Authentication.ServiceAccounts.MaxExpiration > upBound {
				return options, fmt.Errorf("the serviceaccount max expiration must be between 1 hour to 2^32 seconds")
			}
		}

		s.ServiceAccountIssuer, err = serviceaccount.JWTTokenGenerator(s.Authentication.ServiceAccounts.Issuers[0], sk)
		if err != nil {
			return options, fmt.Errorf("failed to build token generator: %v", err)
		}
		s.ServiceAccountTokenMaxExpiration = s.Authentication.ServiceAccounts.MaxExpiration
	}

	if s.Etcd.EnableWatchCache {
		sizes := kubeapiserver.DefaultWatchCacheSizes()
		if userSpecified, err := serveroptions.ParseWatchCacheSizes(s.Etcd.WatchCacheSizes); err == nil {
			for resource, size := range userSpecified {
				sizes[resource] = size
			}
		}
		s.Etcd.WatchCacheSizes, err = serveroptions.WriteWatchCacheSizes(sizes)
		if err != nil {
			return options, err
		}
	}

	if s.APIEnablement.RuntimeConfig != nil {
		for key, value := range s.APIEnablement.RuntimeConfig {
			if key == "v1" || strings.HasPrefix(key, "v1/") ||
				key == "api/v1" || strings.HasPrefix(key, "api/v1/") {
				delete(s.APIEnablement.RuntimeConfig, key)
				s.APIEnablement.RuntimeConfig["/v1"] = value
			}
			if key == "api/legacy" {
				delete(s.APIEnablement.RuntimeConfig, key)
			}
		}
	}
	options.ServerRunOptions = s

	// currently, we use the same ca and ca-key files for tenant and admin
	options.Proxy.ClientCAFile = options.Authentication.ClientCert.ClientCA
	return options, nil
}

func buildServiceResolver(enabledAggregatorRouting bool, hostname string, informer clientgoinformers.SharedInformerFactory) webhook.ServiceResolver {
	var serviceResolver webhook.ServiceResolver
	if enabledAggregatorRouting {
		serviceResolver = aggregatorapiserver.NewEndpointServiceResolver(
			informer.Core().V1().Services().Lister(),
			informer.Core().V1().Endpoints().Lister(),
		)
	} else {
		serviceResolver = aggregatorapiserver.NewClusterIPServiceResolver(
			informer.Core().V1().Services().Lister(),
		)
	}
	// resolve kubernetes.default.svc locally
	if localHost, err := url.Parse(hostname); err == nil {
		serviceResolver = aggregatorapiserver.NewLoopbackServiceResolver(serviceResolver, localHost)
	}
	return serviceResolver
}

func getServiceIPAndRanges(serviceClusterIPRanges string) (net.IP, net.IPNet, net.IPNet, error) {
	serviceClusterIPRangeList := []string{}
	if serviceClusterIPRanges != "" {
		serviceClusterIPRangeList = strings.Split(serviceClusterIPRanges, ",")
	}

	var apiServerServiceIP net.IP
	var primaryServiceIPRange net.IPNet
	var secondaryServiceIPRange net.IPNet
	var err error
	// nothing provided by user, use default range (only applies to the Primary)
	if len(serviceClusterIPRangeList) == 0 {
		var primaryServiceClusterCIDR net.IPNet
		primaryServiceIPRange, apiServerServiceIP, err = master.ServiceIPRange(primaryServiceClusterCIDR)
		if err != nil {
			return net.IP{}, net.IPNet{}, net.IPNet{}, fmt.Errorf("error determining service IP ranges: %v", err)
		}
		return apiServerServiceIP, primaryServiceIPRange, net.IPNet{}, nil
	}

	if len(serviceClusterIPRangeList) > 0 {
		_, primaryServiceClusterCIDR, err := net.ParseCIDR(serviceClusterIPRangeList[0])
		if err != nil {
			return net.IP{}, net.IPNet{}, net.IPNet{}, fmt.Errorf("service-cluster-ip-range[0] is not a valid cidr")
		}

		primaryServiceIPRange, apiServerServiceIP, err = master.ServiceIPRange(*(primaryServiceClusterCIDR))
		if err != nil {
			return net.IP{}, net.IPNet{}, net.IPNet{}, fmt.Errorf("error determining service IP ranges for primary service cidr: %v", err)
		}
	}

	// user provided at least two entries
	// note: validation asserts that the list is max of two dual stack entries
	if len(serviceClusterIPRangeList) > 1 {
		_, secondaryServiceClusterCIDR, err := net.ParseCIDR(serviceClusterIPRangeList[1])
		if err != nil {
			return net.IP{}, net.IPNet{}, net.IPNet{}, fmt.Errorf("service-cluster-ip-range[1] is not an ip net")
		}
		secondaryServiceIPRange = *secondaryServiceClusterCIDR
	}
	return apiServerServiceIP, primaryServiceIPRange, secondaryServiceIPRange, nil
}

func NewBuildHandlerChanFunc(discoveryProxy proxy.DiscoveryProxy) func(apiHandler http.Handler, c *server.Config) (secure http.Handler) {
	return func(handler http.Handler, c *genericapiserver.Config) (secure http.Handler) {
		failedHandler := genericapifilters.Unauthorized(c.Serializer)
		handler = tenantfilters.WithDiscoveryProxy(handler, discoveryProxy)
		handler = tenantfilters.WithTenantInfo(handler)
		handler = genericapifilters.WithAuthentication(handler, c.Authentication.Authenticator, failedHandler, c.Authentication.APIAudiences)
		handler = genericfilters.WithCORS(handler, c.CorsAllowedOriginList, nil, nil, nil, "true")
		handler = genericfilters.WithTimeoutForNonLongRunningRequests(handler, c.LongRunningFunc)
		handler = genericfilters.WithWaitGroup(handler, c.LongRunningFunc, c.HandlerChainWaitGroup)
		handler = genericapifilters.WithRequestInfo(handler, c.RequestInfoResolver)
		if c.SecureServing != nil && !c.SecureServing.DisableHTTP2 && c.GoawayChance > 0 {
			handler = genericfilters.WithProbabilisticGoaway(handler, c.GoawayChance)
		}
		handler = genericapifilters.WithCacheControl(handler)
		handler = genericfilters.WithPanicRecovery(handler, c.RequestInfoResolver)
		return handler
	}
}

var CommonNameUserConversion = x509.UserConversionFunc(func(chain []*stdx509.Certificate) (*authenticator.Response, bool, error) {
	if len(chain[0].Subject.CommonName) == 0 {
		return nil, false, nil
	}

	OrganizationalUnit := chain[0].Subject.OrganizationalUnit
	CommonName := chain[0].Subject.CommonName

	u := user.DefaultInfo{
		Name:   CommonName,
		Groups: chain[0].Subject.Organization,
	}
	tenantIDLength := 6
	if len(OrganizationalUnit) > 0 {
		if len(OrganizationalUnit[0]) == tenantIDLength && len(CommonName) > tenantIDLength {
			if OrganizationalUnit[0] == CommonName[:tenantIDLength] && CommonName[tenantIDLength] == '-' {
				tenantName := OrganizationalUnit[0]
				u.Extra = map[string][]string{"tenant": []string{tenantName}}
			}
		}
	}

	return &authenticator.Response{
		User: &u,
	}, true, nil
})
