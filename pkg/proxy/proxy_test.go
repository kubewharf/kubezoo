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
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kubewharf/kubezoo/pkg/common"
	"github.com/kubewharf/kubezoo/pkg/dynamic"
	"github.com/kubewharf/kubezoo/pkg/util"

	"github.com/stretchr/testify/assert"

	appsapiv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metainternalversion "k8s.io/apimachinery/pkg/apis/meta/internalversion"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/apiserver/pkg/endpoints/request"
	"k8s.io/apiserver/pkg/registry/rest"
	restclient "k8s.io/client-go/rest"
	"k8s.io/kubernetes/pkg/apis/apps"
	"k8s.io/kubernetes/pkg/printers"
	printersinternal "k8s.io/kubernetes/pkg/printers/internalversion"
	printerstorage "k8s.io/kubernetes/pkg/printers/storage"
)

// TestNewTenantProxy tests the NewTenantProxy method.
func TestNewTenantProxy(t *testing.T) {
	invalidConfig := common.StorageConfig{}
	_, err := NewTenantProxy(invalidConfig)
	assert.Error(t, err)

	config := common.StorageConfig{
		NewFunc: func() runtime.Object {
			return nil
		},
	}
	_, err = NewTenantProxy(config)
	assert.NoError(t, err)
}

type fakeConvertor struct{}

func (f *fakeConvertor) ConvertTenantObjectToUpstreamObject(obj runtime.Object, tenantID string, isNamespaceScoped bool) error {
	accessor, err := meta.Accessor(obj)
	if err != nil {
		return err
	}
	if isNamespaceScoped && accessor.GetNamespace() != "" {
		prefixed := util.AddTenantIDPrefix(tenantID, accessor.GetNamespace())
		accessor.SetNamespace(prefixed)
	} else if !isNamespaceScoped && accessor.GetName() != "" {
		prefixed := util.AddTenantIDPrefix(tenantID, accessor.GetName())
		accessor.SetName(prefixed)
	}
	return nil
}

func (f *fakeConvertor) ConvertUpstreamObjectToTenantObject(obj runtime.Object, tenantID string, isNamespaceScoped bool) error {
	accessor, err := meta.Accessor(obj)
	if err != nil {
		return err
	}
	if isNamespaceScoped {
		namespace := accessor.GetNamespace()
		trimmed := util.TrimTenantIDPrefix(tenantID, namespace)
		accessor.SetNamespace(trimmed)
	} else {
		name := accessor.GetName()
		trimmed := util.TrimTenantIDPrefix(tenantID, name)
		accessor.SetName(trimmed)
	}
	return nil
}

func tenantContext(tenantID string, requestInfo *request.RequestInfo) context.Context {
	userInfo := util.AddTenantIDToUserInfo(tenantID, &user.DefaultInfo{})
	ctx := request.WithUser(context.Background(), userInfo)
	return request.WithRequestInfo(ctx, requestInfo)
}

// TestTenantProxy_Get tests the Get method for TenantProxy.
func TestTenantProxy_Get(t *testing.T) {
	tenantID := "test01"
	tenantNamespace := "default"
	upstreamNamespace := util.AddTenantIDPrefix(tenantID, tenantNamespace)
	deploymentName := "foo"
	upstreamDeployment := appsapiv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: upstreamNamespace,
			Name:      deploymentName,
		},
	}

	fakeUpstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == fmt.Sprintf("/apis/apps/v1/namespaces/%s/deployments/%s", upstreamNamespace, deploymentName) {
			deployment, err := json.Marshal(upstreamDeployment)
			assert.NoError(t, err)
			w.Write(deployment)
		} else {
			t.Errorf("unexpected url: %v", r.URL.Path)
		}
	}))
	defer fakeUpstream.Close()
	client := dynamic.NewForConfigOrDie(&restclient.Config{Host: fakeUpstream.URL})
	config := common.StorageConfig{
		Kind:            appsapiv1.SchemeGroupVersion.WithKind("Deployment"),
		Resource:        "deployments",
		ShortNames:      []string{"deploy"},
		NamespaceScoped: true,
		NewFunc:         func() runtime.Object { return &apps.Deployment{} },
		NewListFunc:     func() runtime.Object { return &apps.DeploymentList{} },
		DynamicClient:   client,
		Convertor:       &fakeConvertor{},
	}
	proxy, err := NewTenantProxy(config)
	assert.NoError(t, err)
	getter, ok := proxy.(rest.Getter)
	if !ok {
		t.Errorf("tenant proxy should implement rest.Getter")
	}

	ctx := tenantContext(tenantID, &request.RequestInfo{
		Verb:      "get",
		Namespace: tenantNamespace,
	})
	obj, err := getter.Get(ctx, deploymentName, &metav1.GetOptions{})
	assert.NoError(t, err)
	accessor, err := meta.Accessor(obj)
	assert.NoError(t, err)
	assert.Equal(t, tenantNamespace, accessor.GetNamespace())
}

func TestTenantProxyWithListerList(t *testing.T) {
	tenantID := "test01"
	tenantNamespace := "default"
	upstreamNamespace := util.AddTenantIDPrefix(tenantID, tenantNamespace)
	deploymentName1 := "foo1"
	deploymentName2 := "foo2"
	upstreamDeployment1 := appsapiv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: upstreamNamespace,
			Name:      deploymentName1,
		},
	}
	upstreamDeployment2 := appsapiv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: upstreamNamespace,
			Name:      deploymentName2,
		},
	}
	upstreamDeploymentList := appsapiv1.DeploymentList{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "List",
		},
		ListMeta: metav1.ListMeta{},
		Items:    []appsapiv1.Deployment{upstreamDeployment1, upstreamDeployment2},
	}

	fakeUpstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == fmt.Sprintf("/apis/apps/v1/namespaces/%s/deployments", upstreamNamespace) {
			deployment, err := json.Marshal(upstreamDeploymentList)
			assert.NoError(t, err)
			w.Write(deployment)
		} else {
			t.Errorf("unexpected url: %v", r.URL.Path)
		}
	}))
	defer fakeUpstream.Close()
	client := dynamic.NewForConfigOrDie(&restclient.Config{Host: fakeUpstream.URL})
	config := common.StorageConfig{
		Kind:            appsapiv1.SchemeGroupVersion.WithKind("Deployment"),
		Resource:        "deployments",
		ShortNames:      []string{"deploy"},
		NamespaceScoped: true,
		NewFunc:         func() runtime.Object { return &apps.Deployment{} },
		NewListFunc:     func() runtime.Object { return &apps.DeploymentList{} },
		DynamicClient:   client,
		Convertor:       &fakeConvertor{},
	}
	proxy, err := NewTenantProxy(config)
	assert.NoError(t, err)
	lister, ok := proxy.(rest.Lister)
	if !ok {
		t.Errorf("tenant proxy should implement rest.Lister")
	}

	ctx := tenantContext(tenantID, &request.RequestInfo{
		Verb:      "list",
		Namespace: tenantNamespace,
	})
	obj, err := lister.List(ctx, &metainternalversion.ListOptions{})
	assert.NoError(t, err)
	deploymentList := obj.(*apps.DeploymentList)
	for _, d := range deploymentList.Items {
		assert.Equal(t, tenantNamespace, d.Namespace)
	}
}

func TestTenantProxyCreate(t *testing.T) {
	tenantID := "test01"
	tenantNamespace := "default"
	upstreamNamespace := util.AddTenantIDPrefix(tenantID, tenantNamespace)
	deploymentName := "foo"
	tenantDeployment := appsapiv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: tenantNamespace,
			Name:      deploymentName,
		},
	}
	upstreamDeployment := appsapiv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: upstreamNamespace,
			Name:      deploymentName,
		},
	}

	fakeUpstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == fmt.Sprintf("/apis/apps/v1/namespaces/%s/deployments", upstreamNamespace) {
			deployment, err := json.Marshal(upstreamDeployment)
			assert.NoError(t, err)
			w.Write(deployment)
		} else {
			t.Errorf("unexpected url: %v", r.URL.Path)
		}
	}))
	defer fakeUpstream.Close()
	client := dynamic.NewForConfigOrDie(&restclient.Config{Host: fakeUpstream.URL})
	config := common.StorageConfig{
		Kind:            appsapiv1.SchemeGroupVersion.WithKind("Deployment"),
		Resource:        "deployments",
		ShortNames:      []string{"deploy"},
		NamespaceScoped: true,
		NewFunc:         func() runtime.Object { return &apps.Deployment{} },
		NewListFunc:     func() runtime.Object { return &apps.DeploymentList{} },
		DynamicClient:   client,
		Convertor:       &fakeConvertor{},
	}
	proxy, err := NewTenantProxy(config)
	assert.NoError(t, err)
	creater, ok := proxy.(rest.Creater)
	if !ok {
		t.Errorf("tenant proxy should implement rest.Creater")
	}

	ctx := tenantContext(tenantID, &request.RequestInfo{
		Verb:      "create",
		Namespace: tenantNamespace,
	})
	fakeFunc := func(ctx context.Context, obj runtime.Object) error {
		return nil
	}
	obj, err := creater.Create(ctx, &tenantDeployment, fakeFunc, &metav1.CreateOptions{})
	assert.NoError(t, err)
	accessor, err := meta.Accessor(obj)
	assert.NoError(t, err)
	assert.Equal(t, tenantNamespace, accessor.GetNamespace())
}

func TestTenantProxyUpdate(t *testing.T) {
	tenantID := "test01"
	tenantNamespace := "default"
	upstreamNamespace := util.AddTenantIDPrefix(tenantID, tenantNamespace)
	deploymentName := "foo"
	tenantDeployment := appsapiv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: tenantNamespace,
			Name:      deploymentName,
		},
	}
	upstreamDeployment := appsapiv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: upstreamNamespace,
			Name:      deploymentName,
		},
	}

	fakeUpstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == fmt.Sprintf("/apis/apps/v1/namespaces/%s/deployments/%s", upstreamNamespace, deploymentName) {
			deployment, err := json.Marshal(upstreamDeployment)
			assert.NoError(t, err)
			w.Write(deployment)
		} else {
			t.Errorf("unexpected url: %v", r.URL.Path)
		}
	}))
	defer fakeUpstream.Close()
	client := dynamic.NewForConfigOrDie(&restclient.Config{Host: fakeUpstream.URL})
	config := common.StorageConfig{
		Kind:            appsapiv1.SchemeGroupVersion.WithKind("Deployment"),
		Resource:        "deployments",
		ShortNames:      []string{"deploy"},
		NamespaceScoped: true,
		NewFunc:         func() runtime.Object { return &apps.Deployment{} },
		NewListFunc:     func() runtime.Object { return &apps.DeploymentList{} },
		DynamicClient:   client,
		Convertor:       &fakeConvertor{},
	}
	proxy, err := NewTenantProxy(config)
	assert.NoError(t, err)
	updater, ok := proxy.(rest.Updater)
	if !ok {
		t.Errorf("tenant proxy should implement rest.Updater")
	}

	ctx := tenantContext(tenantID, &request.RequestInfo{
		Verb:      "update",
		Namespace: tenantNamespace,
	})
	fakeValidateObjectFunc := func(ctx context.Context, obj runtime.Object) error {
		return nil
	}
	fakeValidateObjectUpdateFunc := func(ctx context.Context, obj, old runtime.Object) error {
		return nil
	}

	obj, created, err := updater.Update(ctx, deploymentName, rest.DefaultUpdatedObjectInfo(&tenantDeployment), fakeValidateObjectFunc, fakeValidateObjectUpdateFunc, false, &metav1.UpdateOptions{})
	assert.NoError(t, err)
	assert.Equal(t, false, created)
	accessor, err := meta.Accessor(obj)
	assert.NoError(t, err)
	assert.Equal(t, tenantNamespace, accessor.GetNamespace())
}

func TestTenantProxyDelete(t *testing.T) {
	tenantID := "test01"
	tenantNamespace := "default"
	upstreamNamespace := util.AddTenantIDPrefix(tenantID, tenantNamespace)
	deploymentName := "foo"
	upstreamDeployment := appsapiv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: upstreamNamespace,
			Name:      deploymentName,
		},
	}

	fakeUpstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == fmt.Sprintf("/apis/apps/v1/namespaces/%s/deployments/%s", upstreamNamespace, deploymentName) {
			deployment, err := json.Marshal(upstreamDeployment)
			assert.NoError(t, err)
			w.Write(deployment)
		} else {
			t.Errorf("unexpected url: %v", r.URL.Path)
		}
	}))
	defer fakeUpstream.Close()
	client := dynamic.NewForConfigOrDie(&restclient.Config{Host: fakeUpstream.URL})
	config := common.StorageConfig{
		Kind:            appsapiv1.SchemeGroupVersion.WithKind("Deployment"),
		Resource:        "deployments",
		ShortNames:      []string{"deploy"},
		NamespaceScoped: true,
		NewFunc:         func() runtime.Object { return &apps.Deployment{} },
		NewListFunc:     func() runtime.Object { return &apps.DeploymentList{} },
		DynamicClient:   client,
		Convertor:       &fakeConvertor{},
	}
	proxy, err := NewTenantProxy(config)
	assert.NoError(t, err)
	deleter, ok := proxy.(rest.GracefulDeleter)
	if !ok {
		t.Errorf("tenant proxy should implement rest.Getter")
	}

	ctx := tenantContext(tenantID, &request.RequestInfo{
		Verb:      "delete",
		Namespace: tenantNamespace,
	})

	fakeFunc := func(ctx context.Context, obj runtime.Object) error {
		return nil
	}

	obj, deleted, err := deleter.Delete(ctx, deploymentName, fakeFunc, &metav1.DeleteOptions{})
	assert.NoError(t, err)
	accessor, err := meta.Accessor(obj)
	assert.NoError(t, err)
	assert.Equal(t, deleted, true)
	assert.Equal(t, tenantNamespace, accessor.GetNamespace())
}

func TestTenantProxyDeleteCollection(t *testing.T) {
	tenantID := "test01"
	tenantNamespace := "default"
	upstreamNamespace := util.AddTenantIDPrefix(tenantID, tenantNamespace)
	deploymentName1 := "foo1"
	deploymentName2 := "foo2"
	upstreamDeployment1 := appsapiv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: upstreamNamespace,
			Name:      deploymentName1,
		},
	}

	upstreamDeployment2 := appsapiv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: upstreamNamespace,
			Name:      deploymentName2,
		},
	}

	upstreamDeploymentList := appsapiv1.DeploymentList{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "List",
		},
		ListMeta: metav1.ListMeta{},
		Items:    []appsapiv1.Deployment{upstreamDeployment1, upstreamDeployment2},
	}

	fakeUpstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == fmt.Sprintf("/apis/apps/v1/namespaces/%s/deployments/%s", upstreamNamespace, deploymentName1) {
			deployment, err := json.Marshal(upstreamDeployment1)
			assert.NoError(t, err)
			w.Write(deployment)
		} else if r.URL.Path == fmt.Sprintf("/apis/apps/v1/namespaces/%s/deployments/%s", upstreamNamespace, deploymentName2) {
			deployment, err := json.Marshal(upstreamDeployment2)
			assert.NoError(t, err)
			w.Write(deployment)
		} else if r.URL.Path == fmt.Sprintf("/apis/apps/v1/namespaces/%s/deployments", upstreamNamespace) {
			deploymentList, err := json.Marshal(upstreamDeploymentList)
			assert.NoError(t, err)
			w.Write(deploymentList)
		} else {
			t.Errorf("unexpected url: %v", r.URL.Path)
		}
	}))
	defer fakeUpstream.Close()
	client := dynamic.NewForConfigOrDie(&restclient.Config{Host: fakeUpstream.URL})
	config := common.StorageConfig{
		Kind:            appsapiv1.SchemeGroupVersion.WithKind("Deployment"),
		Resource:        "deployments",
		ShortNames:      []string{"deploy"},
		NamespaceScoped: true,
		NewFunc:         func() runtime.Object { return &apps.Deployment{} },
		NewListFunc:     func() runtime.Object { return &apps.DeploymentList{} },
		DynamicClient:   client,
		Convertor:       &fakeConvertor{},
	}
	proxy, err := NewTenantProxy(config)
	assert.NoError(t, err)
	collectionDeleter, ok := proxy.(rest.CollectionDeleter)
	if !ok {
		t.Errorf("tenant proxy should implement rest.Getter")
	}

	ctx := tenantContext(tenantID, &request.RequestInfo{
		Verb:      "delete",
		Namespace: tenantNamespace,
	})

	fakeFunc := func(ctx context.Context, obj runtime.Object) error {
		return nil
	}

	listOptions := metainternalversion.ListOptions{
		LabelSelector: labels.Everything(),
	}

	obj, err := collectionDeleter.DeleteCollection(ctx, fakeFunc, &metav1.DeleteOptions{}, &listOptions)

	assert.NoError(t, err)
	deploymentList := obj.(*apps.DeploymentList)
	for _, d := range deploymentList.Items {
		assert.Equal(t, tenantNamespace, d.Namespace)
	}
}

func TestTenantProxyWatch(t *testing.T) {
	tenantID := "test01"
	tenantNamespace := "default"
	upstreamNamespace := util.AddTenantIDPrefix(tenantID, tenantNamespace)
	deploymentName := "foo1"

	upstreamDeployment := appsapiv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: upstreamNamespace,
			Name:      deploymentName,
		},
	}

	client := dynamic.NewForConfigOrDie(&restclient.Config{Host: ""})
	tc := printerstorage.TableConvertor{TableGenerator: printers.NewTableGenerator().With(printersinternal.AddHandlers)}

	proxy := &tenantProxy{
		kind:             appsapiv1.SchemeGroupVersion.WithKind("Deployment"),
		namespaceScoped:  true,
		isCustomResource: false,
		resource:         "deployments",
		shortNames:       []string{"deploy"},
		newFunc:          func() runtime.Object { return &apps.Deployment{} },
		newListFunc:      func() runtime.Object { return &apps.DeploymentList{} },
		dynamicClient:    client,
		convertor:        &fakeConvertor{},
		tableConvertor:   tc,
	}

	fakeWatcher := watch.NewFake()
	w, err := newProxyWatch(fakeWatcher, proxy, tenantID)
	if err != nil {
		t.Errorf("failed to new proxy watch")
	}

	obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&upstreamDeployment)
	if err != nil {
		t.Errorf("failed to convert to unstructured")
	}
	un := &unstructured.Unstructured{Object: obj}

	fakeWatcher.Add(un)
	event := <-w.ResultChan()
	if event.Type != watch.Added {
		t.Errorf("unexpected event type.")
	}

	accessor, err := meta.Accessor(event.Object)
	assert.NoError(t, err)
	assert.Equal(t, tenantNamespace, accessor.GetNamespace())
}
