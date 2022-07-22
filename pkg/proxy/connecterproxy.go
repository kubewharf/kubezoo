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
	"bufio"
	"context"
	"net"
	"net/http"
	"net/url"
	"strings"

	authenticationv1 "k8s.io/api/authentication/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/httpstream"
	"k8s.io/apimachinery/pkg/util/proxy"
	apirequest "k8s.io/apiserver/pkg/endpoints/request"
	"k8s.io/apiserver/pkg/registry/rest"
	api "k8s.io/kubernetes/pkg/apis/core"

	"github.com/kubewharf/kubezoo/pkg/util"
)

var upgradeableMethods = []string{"GET", "POST"}

const (
	LogSubresource = "log"
	Namespace      = "namespaces"
)

type ConnecterProxy struct {
	transport      http.RoundTripper
	upstreamMaster *url.URL
}

var _ = rest.Connecter(&ConnecterProxy{})

func (cp *ConnecterProxy) New() runtime.Object {
	return &api.Pod{}
}

// Connect proxy the connection to the upstream server if shoud proxy.
func (cp *ConnecterProxy) Connect(ctx context.Context, id string, options runtime.Object, r rest.Responder) (http.Handler, error) {
	// TODO (chao.zheng): validate the input args?
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		requestInfo, ok := apirequest.RequestInfoFrom(req.Context())
		if !ok {
			http.Error(w, "invalid request info", http.StatusInternalServerError)
			return
		}

		// connect is either a pods/log subresource request or
		// a upgrade stream request
		if requestInfo.Subresource != LogSubresource &&
			!httpstream.IsUpgradeRequest(req) {
			http.Error(w, "connect resource request is not upgrade request",
				http.StatusInternalServerError)
			return
		}

		cp.connect(req, w)
	}), nil
}

// connect implement the converting of path and redirect the connection to upstream server.
func (cp *ConnecterProxy) connect(req *http.Request, w http.ResponseWriter) {
	// extract tenant from context
	tenant, ok := util.TenantFrom(req.Context())
	if !ok {
		http.Error(w, "invalid tenant info", http.StatusInternalServerError)
		return
	}
	// extract userInfo from context
	userInfo, ok := apirequest.UserFrom(req.Context())
	if !ok {
		http.Error(w, "no User found in context", http.StatusInternalServerError)
		return
	}
	u := *req.URL
	var namespaceTransformed bool

	// transform namespace for pod
	paths := strings.Split(u.Path, "/")
	for i, p := range paths {
		if p == Namespace && i+1 < len(paths) {
			paths[i+1] = util.AddTenantIDPrefix(tenant, paths[i+1])
			namespaceTransformed = true
			break
		}
	}
	if !namespaceTransformed {
		http.Error(w, "namespace not contained in proxy connect url",
			http.StatusInternalServerError)
		return
	}
	u.Path = strings.Join(paths, "/")

	// set proxy upstream server
	u.Host = cp.upstreamMaster.Host
	u.Scheme = cp.upstreamMaster.Scheme

	// need transform request host also, or it will keep original request host. upstream apiserver can't handle correctly
	req.Host = cp.upstreamMaster.Host

	// set impersonate header
	if req.Header == nil {
		req.Header = make(map[string][]string)
	}
	req.Header[authenticationv1.ImpersonateUserHeader] = []string{userInfo.GetName()}
	req.Header[authenticationv1.ImpersonateGroupHeader] = userInfo.GetGroups()

	// decorate response writer to enable metrics
	delegate := &ResponseWriterDelegator{ResponseWriter: w}
	_, cn := w.(http.CloseNotifier)
	_, fl := w.(http.Flusher)
	_, hj := w.(http.Hijacker)
	var rw http.ResponseWriter
	if cn && fl && hj {
		rw = &FancyResponseWriterDelegator{delegate}
	} else {
		rw = delegate
	}
	// proxy logic
	proxyHandler := proxy.NewUpgradeAwareHandler(&u, cp.transport, false, false, &responder{w: w})
	proxyHandler.ServeHTTP(rw, req)

}

// ResponseWriterDelegator interface wraps http.ResponseWriter to additionally record content-length, status-code, etc.
type ResponseWriterDelegator struct {
	http.ResponseWriter

	status      int
	written     int64
	wroteHeader bool
}

// WriteHeader sends an HTTP response header with the provided status code.
func (r *ResponseWriterDelegator) WriteHeader(code int) {
	r.status = code
	r.wroteHeader = true
	r.ResponseWriter.WriteHeader(code)
}

// Write writes the data to the connection as part of an HTTP reply.
func (r *ResponseWriterDelegator) Write(b []byte) (int, error) {
	if !r.wroteHeader {
		r.WriteHeader(http.StatusOK)
	}
	n, err := r.ResponseWriter.Write(b)
	r.written += int64(n)
	return n, err
}

// Status return the http response status.
func (r *ResponseWriterDelegator) Status() int {
	return r.status
}

// ContentLength return the length of http response content.
func (r *ResponseWriterDelegator) ContentLength() int {
	return int(r.written)
}

type FancyResponseWriterDelegator struct {
	*ResponseWriterDelegator
}

// CloseNotify returns a channel that receives at most a
// single value (true) when the client connection has gone
// away.
func (f *FancyResponseWriterDelegator) CloseNotify() <-chan bool {
	return f.ResponseWriter.(http.CloseNotifier).CloseNotify()
}

// Flush sends any buffered data to the client.
func (f *FancyResponseWriterDelegator) Flush() {
	f.ResponseWriter.(http.Flusher).Flush()
}

// Hijack lets the caller take over the connection.
func (f *FancyResponseWriterDelegator) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return f.ResponseWriter.(http.Hijacker).Hijack()
}

// responder implements rest.Responder for assisting a connector in writing
// objects or errors.
type responder struct {
	w http.ResponseWriter
}

func (r *responder) Error(_ http.ResponseWriter, _ *http.Request, err error) {
	http.Error(r.w, err.Error(), http.StatusInternalServerError)
}

func (cp *ConnecterProxy) NewConnectOptions() (runtime.Object, bool, string) {
	return &api.PodExecOptions{}, false, ""
}

func (cp *ConnecterProxy) ConnectMethods() []string {
	return upgradeableMethods
}

func NewConnecterProxy(transport http.RoundTripper, upstreamMaster *url.URL) (rest.Storage, error) {
	return &ConnecterProxy{
		transport:      transport,
		upstreamMaster: upstreamMaster,
	}, nil
}
