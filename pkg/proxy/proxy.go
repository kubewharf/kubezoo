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
	"encoding/json"
	"fmt"
	"github.com/kubewharf/kubezoo/pkg/proxy/pod"
	"strings"

	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metainternalversion "k8s.io/apimachinery/pkg/apis/meta/internalversion"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/apiserver/pkg/endpoints/request"
	apirequest "k8s.io/apiserver/pkg/endpoints/request"
	genericregistry "k8s.io/apiserver/pkg/registry/generic/registry"
	"k8s.io/apiserver/pkg/registry/rest"
	"k8s.io/kubernetes/pkg/printers"
	printersinternal "k8s.io/kubernetes/pkg/printers/internalversion"
	printerstorage "k8s.io/kubernetes/pkg/printers/storage"

	"github.com/kubewharf/kubezoo/pkg/common"
	"github.com/kubewharf/kubezoo/pkg/dynamic"
	"github.com/kubewharf/kubezoo/pkg/util"
)

// tenantProxyWithLister implements StandardStorage
var _ = rest.StandardStorage(&tenantProxyWithLister{})

// tenantProxy implements the converting between tenant object and upstream object.
type tenantProxy struct {
	// refactored tenant proxy
	kind             schema.GroupVersionKind
	convertor        common.ObjectConvertor
	namespaceScoped  bool
	resource         string
	subresource      string
	shortNames       []string
	isCustomResource bool

	// NewFunc returns a new instance of the type this registry returns for a
	// GET of a single object
	newFunc func() runtime.Object

	// NewListFunc returns a new list of the type this registry; it is the
	// type returned when the resource is listed
	newListFunc func() runtime.Object

	// TableConvertor is an optional interface for transforming items or lists
	// of items into tabular output. If unset, the default will be used.
	tableConvertor rest.TableConvertor

	// dynamic client is used to communicate with upstream cluster
	dynamicClient dynamic.Interface

	groupVersionKindFunc common.GroupVersionKindFunc
}

// tenantProxyWithLister is a wrapper of tenantProxy, it exposes Lister interface to enable installation of List method
// it also exposes TableConvertor interface to convert list to table
type tenantProxyWithLister struct {
	tenantProxy
}

func (p *tenantProxyWithLister) NewList() runtime.Object {
	return p.newList()
}

func (p *tenantProxyWithLister) List(ctx context.Context, options *metainternalversion.ListOptions) (runtime.Object, error) {
	return p.list(ctx, options)
}

func (tp *tenantProxyWithLister) ConvertToTable(ctx context.Context, object runtime.Object, tableOptions runtime.Object) (*metav1.Table, error) {
	if tp.tableConvertor == nil {
		return nil, fmt.Errorf("tableConvertor is nil")
	}
	return tp.tableConvertor.ConvertToTable(ctx, object, tableOptions)
}

func (tc *tenantProxy) NamespaceScoped() bool {
	return tc.namespaceScoped
}

func (tc *tenantProxy) ShortNames() []string {
	return tc.shortNames
}

// NewTenantProxy returns the tenant proxy which implements the storage intefaces.
func NewTenantProxy(config common.StorageConfig) (rest.Storage, error) {
	if config.IsConnecter {
		return NewConnecterProxy(config.ProxyTransport, config.UpstreamMaster)
	}
	if (config.Resource == "pods" || config.Resource == "services" || config.Resource == "nodes") && config.Subresource == "proxy" {
		return pod.NewProxyREST(config.ProxyTransport, config.UpstreamMaster)
	}

	if config.NewFunc == nil && config.NewListFunc == nil {
		return nil, fmt.Errorf("both NewFunc and NewListFunc is nil")
	}
	if config.Subresource != "" && config.NewListFunc != nil {
		return nil, fmt.Errorf("subresource (%s:%s) should not have list method", config.Resource, config.Subresource)
	}

	tc := rest.NewDefaultTableConvertor(apiextensions.Resource("customresourcedefinitions"))
	if !isCustomResourceDefinition(config.Kind) {
		tc = printerstorage.TableConvertor{TableGenerator: printers.NewTableGenerator().With(printersinternal.AddHandlers)}
	}

	proxy := &tenantProxy{
		kind:                 config.Kind,
		namespaceScoped:      config.NamespaceScoped,
		isCustomResource:     config.IsCustomResource,
		resource:             config.Resource,
		subresource:          config.Subresource,
		shortNames:           config.ShortNames,
		newFunc:              config.NewFunc,
		newListFunc:          config.NewListFunc,
		dynamicClient:        config.DynamicClient,
		convertor:            config.Convertor,
		groupVersionKindFunc: config.GroupVersionKindFunc,
		tableConvertor:       tc,
	}
	if config.NewListFunc == nil {
		return proxy, nil
	}
	return &tenantProxyWithLister{*proxy}, nil
}

// isCustomResourceDefinition checks whether the kind is a CRD or not.
func isCustomResourceDefinition(kind schema.GroupVersionKind) bool {
	return kind.GroupKind() == apiextensionsv1.Kind("CustomResourceDefinition")
}

func (tp *tenantProxy) GroupVersionKind(containingGV schema.GroupVersion) schema.GroupVersionKind {
	if tp.groupVersionKindFunc == nil {
		return tp.kind
	}
	return tp.groupVersionKindFunc(containingGV)
}

// getClient returns a dynamic client.
func (tp *tenantProxy) getClient(ctx context.Context) (dynamic.ResourceInterface, error) {
	tenantID, ok := util.TenantFrom(ctx)
	if !ok {
		return nil, fmt.Errorf("tanentID doesn't exist in context")
	}
	requestInfo, ok := apirequest.RequestInfoFrom(ctx)
	if !ok {
		return nil, fmt.Errorf("missing requestInfo")
	}
	var client dynamic.ResourceInterface
	gv := tp.kind.GroupVersion()
	if tp.isCustomResource {
		gv.Group = util.AddTenantIDPrefix(tenantID, gv.Group)
	}
	client = tp.dynamicClient.Resource(gv.WithResource(tp.resource))
	if tp.namespaceScoped && len(requestInfo.Namespace) != 0 {
		namespace := util.AddTenantIDPrefix(tenantID, requestInfo.Namespace)
		client = tp.dynamicClient.Resource(gv.WithResource(tp.resource)).Namespace(namespace)
	}
	return client, nil
}

// convertUnstructuredToOutput convert the unstructured to runtime object.
func (tp *tenantProxy) convertUnstructuredToOutput(utd *unstructured.Unstructured, output runtime.Object) error {
	if o, ok := output.(*unstructured.Unstructured); ok {
		o.SetUnstructuredContent(utd.UnstructuredContent())
		return nil
	}

	kind := tp.GroupVersionKind(tp.kind.GroupVersion())
	original, err := nativeScheme.New(kind)
	if err != nil {
		return err
	}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(utd.Object, original); err != nil {
		return err
	}
	if err := nativeScheme.Convert(original, output, context.TODO()); err != nil {
		return err
	}
	return nil
}

// convertUnstructuredListToOutput convert a unstructured list to runtime object.
func (tp *tenantProxy) convertUnstructuredListToOutput(utdList *unstructured.UnstructuredList, output runtime.Object) error {
	if o, ok := output.(*unstructured.UnstructuredList); ok {
		o.SetUnstructuredContent(utdList.UnstructuredContent())
		return nil
	}

	origin, err := nativeScheme.New(tp.kind.GroupVersion().WithKind(tp.kind.Kind + "List"))
	if err != nil {
		return err
	}

	js, err := utdList.MarshalJSON()
	if err != nil {
		return err
	}

	if err := json.Unmarshal(js, &origin); err != nil {
		return err
	}
	if err := nativeScheme.Convert(origin, output, context.TODO()); err != nil {
		return err
	}
	return nil
}

// Get finds a resource in the upstream cluster by name and returns it.
// Although it can return an arbitrary error value, IsNotFound(err) is true for the
// returned error value err when the specified resource is not found.
func (tp *tenantProxy) Get(ctx context.Context, name string, options *metav1.GetOptions) (runtime.Object, error) {
	if tp.newFunc == nil {
		return nil, fmt.Errorf("newFunc is nil")
	}
	tenantID, ok := util.TenantFrom(ctx)
	if !ok {
		return nil, fmt.Errorf("tanentID doesn't exist in context")
	}

	client, err := tp.getClient(ctx)
	if err != nil {
		return nil, err
	}
	var utd *unstructured.Unstructured
	// todo: renjingsi, expose node to pass Conformance test
	if !tp.namespaceScoped && tp.kind.Kind != "Node" {
		name = util.ConvertTenantObjectNameToUpstream(name, tenantID, tp.kind)
	}
	if subResource := tp.subresource; subResource != "" {
		utd, err = client.Get(ctx, name, *options, subResource)
	} else {
		utd, err = client.Get(ctx, name, *options)
	}
	if err != nil {
		return nil, util.TrimTenantIDFromError(err, tenantID)
	}

	// convert unstructured object to internal for non CRD resources
	output := tp.New()
	if err := tp.convertUnstructuredToOutput(utd, output); err != nil {
		return nil, err
	}
	if err := tp.convertUpstreamObjectToTenantObject(output, tenantID); err != nil {
		return nil, err
	}

	return output, nil
}

// New returns an empty object that can be used with Update after request data has been put into it.
func (tp *tenantProxy) New() runtime.Object {
	if tp.newFunc == nil {
		return nil
	}
	return tp.newFunc()
}

// newList returns an empty object that can be used with the List call.
func (tp *tenantProxy) newList() runtime.Object {
	return tp.newListFunc()
}

// Update finds a resource in the storage and updates it. Some implementations
// may allow updates creates the object - they should set the created boolean
// to true.
func (tp *tenantProxy) Update(ctx context.Context, name string, objInfo rest.UpdatedObjectInfo,
	_ rest.ValidateObjectFunc, _ rest.ValidateObjectUpdateFunc,
	forceAllowCreate bool, options *metav1.UpdateOptions) (runtime.Object, bool, error) {
	if tp.newFunc == nil {
		return nil, false, fmt.Errorf("newFunc is nil")
	}

	requestInfo, ok := request.RequestInfoFrom(ctx)
	if !ok {
		return nil, false, fmt.Errorf("missing requestInfo")
	}
	if requestInfo.Verb == "patch" {
		return tp.guaranteedUpdate(ctx, name, objInfo, options)
	}

	original, err := tp.Get(ctx, name, &metav1.GetOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return nil, false, err
	}
	if errors.IsNotFound(err) && !forceAllowCreate {
		return nil, false, err
	}

	obj, err := objInfo.UpdatedObject(ctx, original)
	if err != nil {
		return nil, false, err
	}
	return tp.update(ctx, obj, options)
}

// update convert the tenant object to upstream object before updating
// to the upstream server, and then convert the response to tenant object.
func (tp *tenantProxy) update(ctx context.Context, obj runtime.Object, options *metav1.UpdateOptions) (runtime.Object, bool, error) {
	tenantID, ok := util.TenantFrom(ctx)
	if !ok {
		return nil, false, fmt.Errorf("missing tenantID in context")
	}

	// 1. convert the internal version of tenant object to upstream object
	if err := tp.convertTenantObjectToUpstreamObject(obj, tenantID); err != nil {
		return nil, false, err
	}

	// 2. convert the internal obj to unstructured
	utd, err := tp.convertInternalObjectToUnstructuredObject(obj)
	if err != nil {
		return nil, false, err
	}

	// 3. call update api
	var (
		got     *unstructured.Unstructured
		created bool
	)
	client, err := tp.getClient(ctx)
	if err != nil {
		return nil, false, err
	}
	if subresource := tp.subresource; subresource == "" {
		got, created, err = client.Update(ctx, utd, *options)
	} else if subresource == "status" {
		got, err = client.UpdateStatus(ctx, utd, *options)
	} else {
		got, created, err = client.Update(ctx, utd, *options, subresource)
	}
	if err != nil {
		return nil, false, util.TrimTenantIDFromError(err, tenantID)
	}

	// 4. convert got to output
	output := tp.New()
	if err := tp.convertUnstructuredToOutput(got, output); err != nil {
		return nil, false, err
	}

	// 5. convert got to tenant
	if err := tp.convertUpstreamObjectToTenantObject(output, tenantID); err != nil {
		return nil, false, err
	}

	return output, created, nil
}

func (tp *tenantProxy) convertInternalObjectToVersionedObject(obj runtime.Object) (runtime.Object, error) {
	kind := tp.GroupVersionKind(tp.kind.GroupVersion())
	return runtime.UnsafeObjectConvertor(nativeScheme).ConvertToVersion(obj, kind.GroupVersion())
}

func (tp *tenantProxy) convertInternalObjectToUnstructuredObject(obj runtime.Object) (*unstructured.Unstructured, error) {
	if utd, ok := obj.(*unstructured.Unstructured); ok {
		// for custom resource, the internal object is already unstructured, so skip converting
		return utd, nil
	}
	versioned, err := tp.convertInternalObjectToVersionedObject(obj)
	if err != nil {
		return nil, err
	}
	utdObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(versioned)
	if err != nil {
		return nil, err
	}
	return &unstructured.Unstructured{Object: utdObj}, nil
}

// Create creates a new version of a resource.
func (tp *tenantProxy) Create(ctx context.Context, obj runtime.Object, _ rest.ValidateObjectFunc, options *metav1.CreateOptions) (runtime.Object, error) {
	if tp.newFunc == nil {
		return nil, fmt.Errorf("newFunc is nil")
	}
	tenantID, ok := util.TenantFrom(ctx)
	if !ok {
		return nil, fmt.Errorf("tanentID doesn't exist in context")
	}

	// 1. convert the internal version of tenant object to upstream object
	if err := tp.convertTenantObjectToUpstreamObject(obj, tenantID); err != nil {
		return nil, err
	}

	// 2. convert the internal obj to unstructured
	utd, err := tp.convertInternalObjectToUnstructuredObject(obj)
	if err != nil {
		return nil, err
	}

	// 3. call create api
	var got *unstructured.Unstructured
	client, err := tp.getClient(ctx)
	if err != nil {
		return nil, err
	}
	if subresource := tp.subresource; subresource == "" {
		got, err = client.Create(ctx, utd, *options)
	} else {
		got, err = client.Create(ctx, utd, *options, subresource)
	}
	if err != nil {
		return nil, util.TrimTenantIDFromError(err, tenantID)
	}

	// 4. convert the got(if it is not an unstructured object) to internal
	output := tp.New()
	if err := tp.convertUnstructuredToOutput(got, output); err != nil {
		return nil, err
	}

	// 5. convert the internal object to tenant
	if err := tp.convertUpstreamObjectToTenantObject(output, tenantID); err != nil {
		return nil, err
	}

	return output, nil
}

// Delete convert the tenant object to upstream object before deleting
// to the upstream server, and then convert the response to tenant object.
func (tp *tenantProxy) Delete(ctx context.Context, name string, _ rest.ValidateObjectFunc, options *metav1.DeleteOptions) (runtime.Object, bool, error) {
	if tp.newFunc == nil {
		return nil, false, fmt.Errorf("newFunc is nil")
	}
	tenantID, ok := util.TenantFrom(ctx)
	if !ok {
		return nil, false, fmt.Errorf("tanentID doesn't exist in context")
	}

	if !tp.namespaceScoped {
		name = util.ConvertTenantObjectNameToUpstream(name, tenantID, tp.kind)
	}
	var (
		got     *unstructured.Unstructured
		deleted bool
		err     error
	)
	client, err := tp.getClient(ctx)
	if subresource := tp.subresource; subresource == "" {
		got, deleted, err = client.Delete(ctx, name, *options)
	} else {
		got, deleted, err = client.Delete(ctx, name, *options, subresource)
	}
	if err != nil {
		return nil, deleted, util.TrimTenantIDFromError(err, tenantID)
	}

	if got.GetAPIVersion() == "v1" && got.GroupVersionKind().Kind == "Status" {
		// if we get status, probably we are handling cr object
		status := &metav1.Status{}
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(got.Object, status); err != nil {
			return nil, deleted, util.TrimTenantIDFromError(err, tenantID)
		}
		ret := util.TrimTenantIDFromStatus(*status, tenantID)
		return &ret, deleted, nil
	}

	output := tp.New()
	if err := tp.convertUnstructuredToOutput(got, output); err != nil {
		return nil, deleted, err
	}

	if err := tp.convertUpstreamObjectToTenantObject(output, tenantID); err != nil {
		return nil, deleted, err
	}

	return output, deleted, nil
}

// list convert the tenant object to upstream object before list from
// the upstream server, and then convert the response to tenant object.
func (tp *tenantProxy) list(ctx context.Context, options *metainternalversion.ListOptions) (runtime.Object, error) {
	if tp.newListFunc == nil {
		return nil, fmt.Errorf("newListFunc is nil")
	}
	tenantID, ok := util.TenantFrom(ctx)
	if !ok {
		return nil, fmt.Errorf("tanentID doesn't exist in context %v", ctx)
	}

	proxyOptions, err := util.ConvertInternalListOptions(ctx, options, tenantID)
	if err != nil {
		return nil, err
	}
	client, err := tp.getClient(ctx)
	if err != nil {
		return nil, err
	}
	utdList, err := client.List(ctx, *proxyOptions)
	if err != nil {
		return nil, util.TrimTenantIDFromError(err, tenantID)
	}

	utdList = util.FilterUnstructuredList(utdList, tenantID, tp.namespaceScoped)

	// convert internal/unstructured list item one by one
	for i := range utdList.Items {
		// convert each item of the unstructured list to internal version for non-CRD resources
		oupObj := tp.New()
		if err := tp.convertUnstructuredToOutput(&utdList.Items[i], oupObj); err != nil {
			return nil, err
		}
		// convert to tenant
		if err := tp.convertUpstreamObjectToTenantObject(oupObj, tenantID); err != nil {
			return nil, err
		}
		// convert it back to unstructured and put it back to the unstructured list
		utd, err := tp.convertInternalObjectToUnstructuredObject(oupObj)
		if err != nil {
			return nil, err
		}
		utdList.Items[i] = *utd
	}

	if tp.isCustomResource {
		utdList.SetAPIVersion(util.TrimTenantIDPrefix(tenantID, utdList.GetAPIVersion()))
	}

	// convert the entire unstructured list to internal version of list for non-CRD resources
	oupList := tp.newList()
	if err := tp.convertUnstructuredListToOutput(utdList, oupList); err != nil {
		return nil, err
	}

	return oupList, nil
}

// DeleteCollection convert the tenant object to upstream object before listing
// from the upstream server, and then delete the item one by one according to the list.
func (tp *tenantProxy) DeleteCollection(ctx context.Context, _ rest.ValidateObjectFunc, options *metav1.DeleteOptions, listOptions *metainternalversion.ListOptions) (runtime.Object, error) {
	if tp.newListFunc == nil {
		return nil, fmt.Errorf("newListFunc is nil")
	}
	tenantID, ok := util.TenantFrom(ctx)
	if !ok {
		return nil, fmt.Errorf("tanentID doesn't exist in context")
	}

	proxyListOptions, err := util.ConvertInternalListOptions(ctx, listOptions, tenantID)
	if err != nil {
		return nil, err
	}
	client, err := tp.getClient(ctx)
	if err != nil {
		return nil, err
	}
	utdList, err := client.List(ctx, *proxyListOptions)
	if err != nil {
		return nil, util.TrimTenantIDFromError(err, tenantID)
	}
	utdList = util.FilterUnstructuredList(utdList, tenantID, tp.namespaceScoped)
	for i := range utdList.Items {
		name := utdList.Items[i].GetName()
		_, _, err = client.Delete(ctx, name, *options)
		if err != nil {
			return nil, util.TrimTenantIDFromError(err, tenantID)
		}
	}

	// convert upstream object to tenant object one at a time
	for i := range utdList.Items {
		// convert each item of the unstructured list to internal version for non-CRD resources
		oupObj := tp.New()
		if err := tp.convertUnstructuredToOutput(&utdList.Items[i], oupObj); err != nil {
			return nil, err
		}
		// convert to tenant
		if err := tp.convertUpstreamObjectToTenantObject(oupObj, tenantID); err != nil {
			return nil, err
		}
		// convert it back to unstructured and put it back to the unstructured list
		utd, err := tp.convertInternalObjectToUnstructuredObject(oupObj)
		if err != nil {
			return nil, err
		}
		utdList.Items[i] = *utd
	}

	//if tp.isCustomResource {
	//	utdList.SetAPIVersion(util.TrimTenantIDPrefix(tp.tenantID, utdList.GetAPIVersion()))
	//}

	// convert the entire unstructured list to internal version of list for non-CRD resources
	oupList := tp.newList()
	if err := tp.convertUnstructuredListToOutput(utdList, oupList); err != nil {
		return nil, err
	}

	return oupList, nil
}

// Watch return a proxy watch if need proxy.
func (tp *tenantProxy) Watch(ctx context.Context, options *metainternalversion.ListOptions) (watch.Interface, error) {
	if tp.newListFunc == nil {
		return nil, fmt.Errorf("newListFunc is nil")
	}
	tenantID, ok := util.TenantFrom(ctx)
	if !ok {
		return nil, fmt.Errorf("tanentID doesn't exist in context")
	}

	proxyOpt, err := util.ConvertInternalListOptions(ctx, options, tenantID)
	if err != nil {
		return nil, err
	}
	client, err := tp.getClient(ctx)
	if err != nil {
		return nil, err
	}
	w, err := client.Watch(ctx, *proxyOpt)
	if err != nil {
		return nil, util.TrimTenantIDFromError(err, tenantID)
	}
	return newProxyWatch(w, tp, tenantID)
}

// convertTenantObjectToUpstreamObject converts tenant object to upstream object.
func (tp *tenantProxy) convertTenantObjectToUpstreamObject(obj runtime.Object, tenantID string) error {
	// if obj is of type unstructured, it should be custom resource, whose apiVersion is prefixed with tenant id
	// (eg: 888888-stable.example.com), leave trimming of tenant id prefix to custom convertor
	if _, ok := obj.(*unstructured.Unstructured); !ok {
		// GVK for internal type object is always empty, set it with the right kind so that we can pick a convertor for it
		obj.GetObjectKind().SetGroupVersionKind(tp.kind)
	}
	return tp.convertor.ConvertTenantObjectToUpstreamObject(obj, tenantID, tp.namespaceScoped)
}

// convertUpstreamObjectToTenantObject converts upstream object to tenant object.
func (tp *tenantProxy) convertUpstreamObjectToTenantObject(obj runtime.Object, tenantID string) error {
	// if obj is of type unstructured, it should be custom resource, whose apiVersion is prefixed with tenant id
	// (eg: 888888-stable.example.com), leave trimming of tenant id prefix to custom convertor
	if _, ok := obj.(*unstructured.Unstructured); !ok {
		// GVK for internal type object is always empty, set it with the right kind so that we can pick a convertor for it
		obj.GetObjectKind().SetGroupVersionKind(tp.kind)
	}
	return tp.convertor.ConvertUpstreamObjectToTenantObject(obj, tenantID, tp.namespaceScoped)
}

// guaranteedUpdate ensures a guaranteed updating.
func (tp *tenantProxy) guaranteedUpdate(ctx context.Context, name string,
	objInfo rest.UpdatedObjectInfo, options *metav1.UpdateOptions,
) (runtime.Object, bool, error) {
	for {
		original, err := tp.Get(ctx, name, &metav1.GetOptions{})
		if err != nil && !errors.IsNotFound(err) {
			return nil, false, err
		}
		if errors.IsNotFound(err) {
			if err := runtime.SetZeroValue(original); err != nil {
				return nil, false, err
			}
		}

		if err := checkPreconditions(objInfo.Preconditions(), original); err != nil {
			return nil, false, err
		}
		updated, err := objInfo.UpdatedObject(ctx, original)
		if err != nil {
			return nil, false, err
		}

		got, created, err := tp.update(ctx, updated, options)
		if errors.IsConflict(err) && strings.Contains(err.Error(), genericregistry.OptimisticLockErrorMsg) {
			// retry update on optimistic lock conflict
			continue
		}
		if err != nil {
			return nil, false, err
		}
		return got, created, nil
	}
}

// checkPreconditions checks the precondition for guarantee updating.
func checkPreconditions(preconditions *metav1.Preconditions, obj runtime.Object) error {
	if preconditions == nil {
		return nil
	}
	objMeta, err := meta.Accessor(obj)
	if err != nil {
		return fmt.Errorf("can't enforce preconditions %v on un-introspectable object %v, got error: %v", *preconditions, obj, err)
	}
	if preconditions.UID != nil && *preconditions.UID != objMeta.GetUID() {
		return fmt.Errorf("precondition failed: UID in precondition: %v, UID in object meta: %v", *preconditions.UID, objMeta.GetUID())
	}
	if preconditions.ResourceVersion != nil && *preconditions.ResourceVersion != objMeta.GetResourceVersion() {
		return fmt.Errorf("Precondition failed: ResourceVersion in precondition: %v, ResourceVersion in object meta: %v", *preconditions.ResourceVersion, objMeta.GetResourceVersion())
	}
	return nil
}
