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
	"net/http"
	"net/http/httptest"
	"testing"

	"k8s.io/apiserver/pkg/endpoints/request"
	"k8s.io/kubernetes/pkg/serviceaccount"

	"github.com/kubewharf/kubezoo/pkg/util"
)

// TestWithTenantInfo tests the method WithTenantInfo.
func TestWithTenantInfo(t *testing.T) {
	tenantID := "demo01"
	serviceAccountInfo := serviceaccount.ServiceAccountInfo{
		Name:      "foo",
		Namespace: util.AddTenantIDPrefix(tenantID, "foo"),
		UID:       "aa458e95-3a0d-11e7-a8bf-0cc47a944282",
		PodName:   "bar",
		PodUID:    "aa458e95-3a0d-11e7-a8bf-0cc47a944141",
	}
	req := &http.Request{}
	req = req.WithContext(request.WithUser(req.Context(), serviceAccountInfo.UserInfo()))

	success := make(chan struct{})
	tenant := WithTenantInfo(
		http.HandlerFunc(func(_ http.ResponseWriter, req *http.Request) {
			user, ok := request.UserFrom(req.Context())
			if !ok {
				t.Fatalf("no user info found")
			}
			tenantIDs := user.GetExtra()[util.TenantIDKey]
			if len(tenantIDs) != 1 || tenantIDs[0] != tenantID {
				t.Fatalf("tenantID slice should be [%s], but got: %v", tenantID, tenantIDs)
			}
			close(success)
		}))

	tenant.ServeHTTP(httptest.NewRecorder(), req)
	<-success
}
