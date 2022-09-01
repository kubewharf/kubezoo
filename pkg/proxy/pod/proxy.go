package pod

import (
	"context"
	"github.com/kubewharf/kubezoo/pkg/util"
	authenticationv1 "k8s.io/api/authentication/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/proxy"
	apirequest "k8s.io/apiserver/pkg/endpoints/request"
	"k8s.io/apiserver/pkg/registry/rest"
	api "k8s.io/kubernetes/pkg/apis/core"
	"net/http"
	"net/url"
	"strings"
)

const (
	LogSubresource = "log"
	Namespace      = "namespaces"
)

// Proxy implements the proxy subresource for a Pod
type ProxyREST struct {
	transport      http.RoundTripper
	upstreamMaster *url.URL
}

func NewProxyREST(transport http.RoundTripper, upstreamMaster *url.URL) (rest.Storage, error) {
	return &ProxyREST{
		transport:      transport,
		upstreamMaster: upstreamMaster,
	}, nil
}

// Implement Connecter
var _ = rest.Connecter(&ProxyREST{})

var proxyMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}

// New returns an empty podProxyOptions object.
func (r *ProxyREST) New() runtime.Object {
	return &api.PodProxyOptions{}
}

// ConnectMethods returns the list of HTTP methods that can be proxied
func (r *ProxyREST) ConnectMethods() []string {
	return proxyMethods
}

// NewConnectOptions returns versioned resource that represents proxy parameters
func (r *ProxyREST) NewConnectOptions() (runtime.Object, bool, string) {
	return &api.PodProxyOptions{}, true, "path"
}

// Connect returns a handler for the pod proxy
func (r *ProxyREST) Connect(ctx context.Context, id string, opts runtime.Object, responder rest.Responder) (http.Handler, error) {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
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
		//var namespaceTransformed bool

		// transform namespace for pod
		paths := strings.Split(u.Path, "/")
		for i, p := range paths {
			if p == Namespace && i+1 < len(paths) {
				paths[i+1] = util.AddTenantIDPrefix(tenant, paths[i+1])
				//namespaceTransformed = true
				break
			}
		}
		//if !namespaceTransformed {
		//	http.Error(w, "namespace not contained in proxy connect url",
		//		http.StatusInternalServerError)
		//	return
		//}

		// to prevent upstream 301
		lastIndex := len(paths) - 1
		if lastIndex >= 0 && paths[lastIndex] == "proxy" {
			paths[lastIndex] = "proxy/"
		}

		u.Path = strings.Join(paths, "/")

		// set proxy upstream server
		u.Host = r.upstreamMaster.Host
		u.Scheme = r.upstreamMaster.Scheme

		// need transform request host also, or it will keep original request host. upstream apiserver can't handle correctly
		req.Host = r.upstreamMaster.Host

		// set impersonate header
		if req.Header == nil {
			req.Header = make(map[string][]string)
		}
		req.Header[authenticationv1.ImpersonateUserHeader] = []string{userInfo.GetName()}
		req.Header[authenticationv1.ImpersonateGroupHeader] = userInfo.GetGroups()

		//proxyOpts, ok := opts.(*api.PodProxyOptions)
		//if !ok {
		//	http.Error(w, "invalid tenant info", http.StatusInternalServerError)
		//	return
		//}
		//u.Path = net.JoinPreservingTrailingSlash(u.Path, proxyOpts.Path)

		// proxy logic
		proxyHandler := proxy.NewUpgradeAwareHandler(&u, r.transport, false, false, &responder_{w: w})
		proxyHandler.ServeHTTP(w, req)
	}), nil
}

// responder implements rest.Responder for assisting a connector in writing
// objects or errors.
type responder_ struct {
	w http.ResponseWriter
}

func (r *responder_) Error(_ http.ResponseWriter, _ *http.Request, err error) {
	http.Error(r.w, err.Error(), http.StatusInternalServerError)
}
