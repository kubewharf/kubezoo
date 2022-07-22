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

package filters

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/apiserver/pkg/endpoints/request"

	"github.com/kubewharf/kubezoo/pkg/util"
)

type fakeDiscoveryProxy struct {
	tenantID string
	group    string
	version  string

	apiGroupList    *metav1.APIGroupList
	apiGroup        *metav1.APIGroup
	apiResourceList *metav1.APIResourceList
}

func (dp *fakeDiscoveryProxy) ServerGroups(tenantID string) (*metav1.APIGroupList, error) {
	if tenantID == dp.tenantID {
		return dp.apiGroupList, nil
	}
	return nil, nil
}

func (dp *fakeDiscoveryProxy) ServerVersionsForGroup(tenantID, group string) (*metav1.APIGroup, error) {
	if tenantID == dp.tenantID && group == dp.group {
		return dp.apiGroup, nil
	}
	return nil, nil
}

func (dp *fakeDiscoveryProxy) ServerResourcesForGroupVersion(tenantID, group, version string) (*metav1.APIResourceList, error) {
	if tenantID == dp.tenantID && group == dp.group && version == dp.version {
		return dp.apiResourceList, nil
	}
	return nil, nil
}

// TestWithDiscoveryProxy checks some methods about discovery.
func TestWithDiscoveryProxy(t *testing.T) {
	tenantID := "demo01"
	group := "extensions"
	version := "v1beta1"
	apiGroupList := &metav1.APIGroupList{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "APIGroupList",
		},
		Groups: []metav1.APIGroup{
			{
				Name: "extensions",
				Versions: []metav1.GroupVersionForDiscovery{
					{
						GroupVersion: "extensions/v1beta1",
						Version:      "v1beta1",
					},
				},
			},
		},
	}
	apiGroup := &metav1.APIGroup{
		Name: "extensions",
		Versions: []metav1.GroupVersionForDiscovery{
			{
				GroupVersion: "extensions/v1beta1",
				Version:      "v1beta1",
			},
		},
	}
	apiResourceList := &metav1.APIResourceList{
		GroupVersion: "extensions/v1beta1",
		APIResources: []metav1.APIResource{
			{
				Name:       "ingresses",
				Namespaced: true,
			},
		},
	}

	proxy := &fakeDiscoveryProxy{
		tenantID:        tenantID,
		group:           group,
		version:         version,
		apiGroupList:    apiGroupList,
		apiResourceList: apiResourceList,
		apiGroup:        apiGroup,
	}
	discovery := WithDiscoveryProxy(nil, proxy)
	requestInfo := &request.RequestInfo{
		Verb: "get",
		Path: "/apis",
	}
	userInfo := util.AddTenantIDToUserInfo(tenantID, &user.DefaultInfo{})
	ctx := request.WithUser(context.Background(), userInfo)
	ctx = request.WithRequestInfo(ctx, requestInfo)

	// test GET /apis
	getGroupsReq := (&http.Request{URL: &url.URL{Path: "/apis"}}).WithContext(ctx)
	getGroupsResp := httptest.NewRecorder()
	discovery.ServeHTTP(getGroupsResp, getGroupsReq)
	actual := getGroupsResp.Body.Bytes()
	expected, _ := json.Marshal(apiGroupList)
	assert.Equal(t, expected, actual)

	// test GET /apis/extensions
	getVersionsReq := (&http.Request{URL: &url.URL{Path: "/apis/extensions"}}).WithContext(ctx)
	getVersionsResp := httptest.NewRecorder()
	discovery.ServeHTTP(getVersionsResp, getVersionsReq)
	actual = getVersionsResp.Body.Bytes()
	expected, _ = json.Marshal(apiGroup)
	assert.Equal(t, expected, actual)

	// test GET /apis/extensions/v1beta1
	getResourceReq := (&http.Request{URL: &url.URL{Path: "/apis/extensions/v1beta1"}}).WithContext(ctx)
	getResourceResp := httptest.NewRecorder()
	discovery.ServeHTTP(getResourceResp, getResourceReq)
	actual = getResourceResp.Body.Bytes()
	expected, _ = json.Marshal(apiResourceList)
	assert.Equal(t, expected, actual)
}
