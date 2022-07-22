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
	"testing"

	v1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/apiserver/pkg/endpoints/request"
)

// TestGetTenantIDFromNamespace tests the GetTenantIDFromNamespace function.
func TestGetTenantIDFromNamespace(t *testing.T) {
	tenantId := "111111"
	ns, err := GetTenantIDFromNamespace("xxx")
	if ns != "" || err == nil {
		t.Errorf("'xxx' ns should return empty tenant with err")
	}

	ns, err = GetTenantIDFromNamespace("xxxxxxxxxxxxxx")
	if ns != "" || err == nil {
		t.Errorf("'xxxxxxxxxxxxxx' ns should return empty tenant with err")
	}

	ns, err = GetTenantIDFromNamespace(tenantId + "-myns")
	if ns != tenantId || err != nil {
		t.Errorf("'111111-myns' ns should return '111111' ns success")
	}
}

// TestAddTenantIDToUserInfo tests the AddTenantIDToUserInfo function.
func TestAddTenantIDToUserInfo(t *testing.T) {
	tenantId := "111111"
	defaultInfo := user.DefaultInfo{
		Name: "foo",
	}

	info := AddTenantIDToUserInfo(tenantId, &defaultInfo)
	if info == nil {
		t.Errorf("user info should not be nil")
	}
	extra := info.GetExtra()
	if v, ok := extra[TenantIDKey]; !ok {
		t.Errorf("TenantIDKey %s should be found", TenantIDKey)
	} else {
		if len(v) != 1 {
			t.Errorf("extra length should be one")
			if v[0] != tenantId {
				t.Errorf("tenant id is not same, found %s", v[0])
			}
		}
	}
}

// TestConvertTenantObjectNameToUpstream checks the converting from tenant object to upstream object.
func TestConvertTenantObjectNameToUpstream(t *testing.T) {
	tenantId := "111111"
	tests := []struct {
		name     string
		tenantId string
		gvk      schema.GroupVersionKind
		expect   string
	}{
		{
			name:     "myname",
			tenantId: tenantId,
			gvk: schema.GroupVersionKind{
				Group:   "",
				Version: "v1",
				Kind:    "PersistenceVolume",
			},
			expect: tenantId + "-myname",
		},
		{
			name:     "plural.group",
			tenantId: tenantId,
			gvk: schema.GroupVersionKind{
				Group:   "apiextensions.k8s.io",
				Version: "v1",
				Kind:    "CustomResourceDefinition",
			},
			expect: "plural.111111-group",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := ConvertTenantObjectNameToUpstream(test.name, test.tenantId, test.gvk)
			if got != test.expect {
				t.Errorf("unexpected name, got %s, want %s", got, test.expect)
			}
		})
	}
}

// TestUpstreamObjectBelongsToTenant tests the UpstreamObjectBelongsToTenant function.
func TestUpstreamObjectBelongsToTenant(t *testing.T) {
	tenantId := "111111"
	CRDPlural := "foos"
	CRDGroup := tenantId + "-" + "a.com"
	CRDVersion := "v1"
	FullCRDName := CRDPlural + "." + CRDGroup

	tests := []struct {
		name              string
		obj               runtime.Object
		tenantId          string
		isNamespaceScoped bool
		expect            bool
	}{
		{
			name: "test namespaced upstream object belongs to the tenant",
			obj: &v1.Pod{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Pod",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "myname",
					Namespace: tenantId + "-myns",
				},
			},
			tenantId:          tenantId,
			isNamespaceScoped: true,
			expect:            true,
		},
		{
			name: "test namespaced upstream object not belongs to the tenant",
			obj: &v1.Pod{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Pod",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "myname",
					Namespace: "myns",
				},
			},
			tenantId:          tenantId,
			isNamespaceScoped: true,
			expect:            false,
		},
		{
			name: "test cluster scope upstream object belongs to the tenant",
			obj: &v1.PersistentVolume{
				TypeMeta: metav1.TypeMeta{
					Kind:       "PersistentVolume",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      tenantId + "-myname",
					Namespace: "myns",
				},
			},
			tenantId:          tenantId,
			isNamespaceScoped: false,
			expect:            true,
		},
		{
			name: "test cluster scope upstream object not belongs to the tenant",
			obj: &v1.PersistentVolume{
				TypeMeta: metav1.TypeMeta{
					Kind:       "PersistentVolume",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "myname",
					Namespace: "myns",
				},
			},
			tenantId:          tenantId,
			isNamespaceScoped: false,
			expect:            false,
		},
		{
			name: "test crd upstream object belongs to the tenant",
			obj: &apiextensionsv1.CustomResourceDefinition{
				TypeMeta: metav1.TypeMeta{
					Kind:       "CustomResourceDefinition",
					APIVersion: "apiextensions/v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: FullCRDName,
				},
				Spec: apiextensionsv1.CustomResourceDefinitionSpec{
					Group: CRDGroup,
					Versions: []apiextensionsv1.CustomResourceDefinitionVersion{
						{
							Name: CRDVersion,
						},
					},
					Scope: apiextensionsv1.NamespaceScoped,
					Names: apiextensionsv1.CustomResourceDefinitionNames{
						Plural: CRDPlural,
						Kind:   "Foo",
					},
				},
			},
			tenantId:          tenantId,
			isNamespaceScoped: false,
			expect:            true,
		},
		{
			name: "test crd upstream object not belongs to the tenant",
			obj: &apiextensionsv1.CustomResourceDefinition{
				TypeMeta: metav1.TypeMeta{
					Kind:       "CustomResourceDefinition",
					APIVersion: "apiextensions/v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "xxxx.xxxx",
				},
				Spec: apiextensionsv1.CustomResourceDefinitionSpec{
					Group: CRDGroup,
					Versions: []apiextensionsv1.CustomResourceDefinitionVersion{
						{
							Name: CRDVersion,
						},
					},
					Scope: apiextensionsv1.NamespaceScoped,
					Names: apiextensionsv1.CustomResourceDefinitionNames{
						Plural: CRDPlural,
						Kind:   "Foo",
					},
				},
			},
			tenantId:          tenantId,
			isNamespaceScoped: false,
			expect:            false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := UpstreamObjectBelongsToTenant(test.obj, test.tenantId, test.isNamespaceScoped)
			if got != test.expect {
				t.Errorf("test case %s unexpect, got %v, want %v", test.name, got, test.expect)
			}
		})
	}
}

// TestConvertUpstreamApiGroupToTenant tests the ConvertUpstreamApiGroupToTenant method.
func TestConvertUpstreamApiGroupToTenant(t *testing.T) {
	tenantId := "111111"
	name := "myapigroup"
	groupVersion := "mygroupversion"
	apiGroup := &metav1.APIGroup{
		Name: tenantId + "-" + name,
		Versions: []metav1.GroupVersionForDiscovery{
			{
				GroupVersion: tenantId + "-" + groupVersion,
			},
		},
		PreferredVersion: metav1.GroupVersionForDiscovery{
			GroupVersion: tenantId + "-" + groupVersion,
		},
	}

	ConvertUpstreamApiGroupToTenant(tenantId, apiGroup)
	if apiGroup.Name != name {
		t.Errorf("unexpected name, got %s, want %s", apiGroup.Name, name)
	}
	if apiGroup.PreferredVersion.GroupVersion != groupVersion {
		t.Errorf("unexpectd perferred version, got %s, want %s", apiGroup.PreferredVersion.GroupVersion, groupVersion)
	}
	if apiGroup.Versions[0].GroupVersion != groupVersion {
		t.Errorf("unexpectd version, got %s, want %s", apiGroup.Versions[0].GroupVersion, groupVersion)
	}
}

// TestConvertUpstreamResourceListToTenant tests the ConvertUpstreamResourceListToTenant method.
func TestConvertUpstreamResourceListToTenant(t *testing.T) {
	tenantId := "111111"
	groupVersion := "mygroupversion"
	group := "mygroup"
	resourceList := &metav1.APIResourceList{
		GroupVersion: tenantId + "-" + groupVersion,
		APIResources: []metav1.APIResource{
			{
				Group: tenantId + "-" + group,
			},
		},
	}

	ConvertUpstreamResourceListToTenant(tenantId, resourceList)
	if resourceList.GroupVersion != groupVersion {
		t.Errorf("unexpectd version, got %s, want %s", resourceList.GroupVersion, groupVersion)
	}
	if resourceList.APIResources[0].Group != group {
		t.Errorf("unexpectd group, got %s, want %s", resourceList.APIResources[0].Group, group)
	}
}

// TestTenantIDFrom checks getting the tenantId from context.
func TestTenantIDFrom(t *testing.T) {

	tests := []struct {
		c      context.Context
		expect string
		isGot  bool
		name   string
	}{
		{
			name: "get tenant from context",
			c: request.WithUser(context.TODO(), &user.DefaultInfo{
				Extra: map[string][]string{
					"tenant": []string{"111111"},
				},
			}),
			expect: "111111",
			isGot:  true,
		},
		{
			name: "get empty tenant from context",
			c: request.WithUser(context.TODO(), &user.DefaultInfo{
				Extra: map[string][]string{},
			}),
			expect: "",
			isGot:  false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := TenantIDFrom(test.c)
			if got != test.expect {
				t.Errorf("unexpectd tenant, got %s, want %s", got, test.expect)
			}
			got, isGot := TenantFrom(test.c)
			if got != test.expect {
				t.Errorf("unexpectd tenant, got %s, want %s", got, test.expect)
			}
			if isGot != test.isGot {
				t.Errorf("unexpectd expectGot, got %v, want %v", isGot, test.isGot)
			}
		})
	}
}

// TestNewCheckGroupKindFunc tests the group kink checking function.
func TestNewCheckGroupKindFunc(t *testing.T) {
	tenantId := "111111"
	CRDGroup := tenantId + "-" + "kubezoo.io"
	CRDVersion := "v1"
	CRDPlural := "foos"
	FullCRDName := CRDPlural + "." + CRDGroup

	crds := []*apiextensionsv1.CustomResourceDefinition{
		{
			TypeMeta: metav1.TypeMeta{
				Kind:       "CustomResourceDefinition",
				APIVersion: "apiextensions/v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: FullCRDName,
			},
			Spec: apiextensionsv1.CustomResourceDefinitionSpec{
				Group: CRDGroup,
				Versions: []apiextensionsv1.CustomResourceDefinitionVersion{
					{
						Name: CRDVersion,
					},
				},
				Scope: apiextensionsv1.NamespaceScoped,
				Names: apiextensionsv1.CustomResourceDefinitionNames{
					Plural: CRDPlural,
					Kind:   "Foo",
				},
			},
		},
	}

	crdLister := FakeCRDLister{
		crds,
	}

	f := NewCheckGroupKindFunc(&crdLister)

	tests := []struct {
		name                      string
		group                     string
		kind                      string
		isTenantObject            bool
		expectNamespaced          bool
		expectCustomResourceGroup bool
		expectErrorNil            bool
	}{
		{
			name:                      "check pod resource",
			group:                     "",
			kind:                      "Pod",
			isTenantObject:            false,
			expectNamespaced:          true,
			expectCustomResourceGroup: false,
			expectErrorNil:            true,
		},
		{
			name:                      "check persistence volume resource",
			group:                     "",
			kind:                      "PersistentVolume",
			isTenantObject:            false,
			expectNamespaced:          false,
			expectCustomResourceGroup: false,
			expectErrorNil:            true,
		},
		{
			name:                      "check crd resource",
			group:                     "kubezoo.io",
			kind:                      "Foo",
			isTenantObject:            true,
			expectNamespaced:          true,
			expectCustomResourceGroup: true,
			expectErrorNil:            true,
		},
		{
			name:                      "check crd resource not exists",
			group:                     "kubezoo1.io",
			kind:                      "Foo1",
			isTenantObject:            true,
			expectNamespaced:          false,
			expectCustomResourceGroup: false,
			expectErrorNil:            false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gotNamespaced, gotCustomResourceGroup, err := f(test.group, test.kind, tenantId, test.isTenantObject)
			if test.expectErrorNil && err != nil {
				t.Errorf("unexpected err %s", err)
			}
			if !test.expectErrorNil && err == nil {
				t.Errorf("expected err")
			}
			if gotNamespaced != test.expectNamespaced {
				t.Errorf("unexpected namespaced got %v, want %v", gotNamespaced, test.expectNamespaced)
			}
			if gotCustomResourceGroup != test.expectCustomResourceGroup {
				t.Errorf("unexpected CustomResourceGroup got %v, want %v", gotCustomResourceGroup, test.expectCustomResourceGroup)
			}
		})
	}
}
