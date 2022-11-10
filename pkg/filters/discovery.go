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
	"encoding/json"
	"net/http"
	"strings"

	"github.com/golang/protobuf/proto"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apiserver/pkg/endpoints/request"

	"github.com/kubewharf/kubezoo/pkg/proxy"
	"github.com/kubewharf/kubezoo/pkg/util"
)

const (
	// protobuf mime type
	openAPIV2mimePb = "application/com.github.proto-openapi.spec.v2@v1.0+protobuf"
)

// WithDiscoveryProxy creates an http handler that proxy tenant discovery request
func WithDiscoveryProxy(handler http.Handler, discoveryProxy proxy.DiscoveryProxy) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestInfo, _ := request.RequestInfoFrom(r.Context())
		tenantID := util.TenantIDFrom(r.Context())
		if tenantID == "" || !isDiscoveryRequest(requestInfo) {
			handler.ServeHTTP(w, r)
			return
		}

		path := strings.Trim(r.URL.Path, "/")
		parts := strings.Split(path, "/")
		if len(parts) == 1 && parts[0] == "apis" {
			// path: /apis
			groups, err := discoveryProxy.ServerGroups(tenantID)
			if err != nil {
				responseDiscoveryError(w, err)
				return
			}
			responseJson(w, groups)
			return
		} else if len(parts) == 2 && parts[0] == "apis" {
			// path: /apis/{group}
			versions, err := discoveryProxy.ServerVersionsForGroup(tenantID, parts[1])
			if err != nil {
				responseDiscoveryError(w, err)
				return
			}
			responseJson(w, versions)
			return
		} else if len(parts) == 3 && parts[0] == "apis" {
			// path: /apis/{group}/{version}
			resources, err := discoveryProxy.ServerResourcesForGroupVersion(tenantID, parts[1], parts[2])
			if err != nil {
				responseDiscoveryError(w, err)
				return
			}
			responseJson(w, resources)
			return
		} else if len(parts) == 1 && parts[0] == "version" {
			// path: /version
			version, err := discoveryProxy.ServerVersion()
			if err != nil {
				responseDiscoveryError(w, err)
				return
			}
			responseJson(w, version)
			return
		} else if path == "openapi/v2" || path == "swagger-2.0.0.pb-v1" {
			if r.Header.Get("Accept") == openAPIV2mimePb {
				doc, err := discoveryProxy.OpenAPISchema()
				if err != nil {
					responseDiscoveryError(w, err)
					return
				}
				docJson, err := json.Marshal(doc)
				if err != nil {
					responseDiscoveryError(w, err)
					return
				}
				newJson := strings.ReplaceAll(string(docJson), tenantID+"-", "")
				if err := json.Unmarshal([]byte(newJson), doc); err != nil {
					responseDiscoveryError(w, err)
					return
				}

				bytes, err := proto.Marshal(doc)
				if err != nil {
					responseDiscoveryError(w, err)
					return
				}
				w.Write(bytes)
				return
			}

			swagger, err := discoveryProxy.GetSwagger()
			if err != nil {
				responseDiscoveryError(w, err)
				return
			}
			for key, schema := range swagger.SwaggerProps.Definitions {
				if strings.Contains(key, tenantID+"-") {
					delete(swagger.SwaggerProps.Definitions, key)
					newKey := strings.ReplaceAll(key, tenantID+"-", "")
					swagger.SwaggerProps.Definitions[newKey] = schema
				}
			}
			responseJson(w, swagger)
			return
		}
	})
}

// responseDiscoveryError returns StatusInternalServerError response.
func responseDiscoveryError(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}
	code := http.StatusInternalServerError
	msg := err.Error()
	statusErr, ok := err.(*errors.StatusError)
	if ok && statusErr.Status().Code != 0 {
		code = int(statusErr.ErrStatus.Code)
	}
	w.WriteHeader(code)
	w.Write([]byte(msg))
}

// responseJson marshal the body and write to the connection
// as part of an HTTP reply.
func responseJson(w http.ResponseWriter, v interface{}) {
	js, err := json.Marshal(v)
	if err != nil {
		responseDiscoveryError(w, err)
		return
	}
	w.Write(js)
}

// isDiscoveryRequest checks the request is discovery request or not.
func isDiscoveryRequest(requestInfo *request.RequestInfo) bool {
	if requestInfo.IsResourceRequest || requestInfo.Verb != "get" || requestInfo.Path == "/api" {
		return false
	}
	// todo(renjingsi): handle /api
	if strings.HasPrefix(requestInfo.Path, "/apis") {
		return true
	}
	switch requestInfo.Path {
	case "/version", "/openapi/v2", "/swagger-2.0.0.pb-v1":
		return true
	}
	return false
}
