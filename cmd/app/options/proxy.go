package options

import (
	"fmt"

	"github.com/spf13/pflag"
)

// ProxyOptions runs a kubezoo proxy server
type ProxyOptions struct {
	// ca, cert and key file to build secure connection between kubezoo and upstream cluster
	ProxyClientCAFile   string
	ProxyClientCertFile string
	ProxyClientKeyFile  string

	ProxyClientQPS   float32
	ProxyClientBurst int

	ClientCAFile          string
	ClientCAKeyFile       string
	UpstreamMaster        string
	ServiceAccountKeyFile string
	BindAddress           string
	SecurePort            int
}

// NewProxyOptions creates a new ProxyOptions object
func NewProxyOptions() *ProxyOptions {
	return &ProxyOptions{
		ProxyClientQPS:   1000,
		ProxyClientBurst: 2000,
		SecurePort:       6443,
	}
}

func (o *ProxyOptions) AddFlags(fs *pflag.FlagSet) {
	if o == nil {
		return
	}
	fs.StringVar(&o.ProxyClientCAFile, "proxy-client-ca-file", o.ProxyClientCAFile, "proxy client ca file to verify upstream cluster apiserver.")
	fs.StringVar(&o.ProxyClientCertFile, "proxy-client-cert-file", o.ProxyClientCertFile, "proxy client cert file to prove the identity of kubezoo proxy "+
		"server to upstream cluster apiserver.")
	fs.StringVar(&o.ProxyClientKeyFile, "proxy-client-key-file", o.ProxyClientKeyFile, "proxy client key file to prove the identity of kubezoo proxy "+
		"server to upstream cluster apiserver.")
	fs.Float32Var(&o.ProxyClientQPS, "proxy-client-qps", o.ProxyClientQPS,
		fmt.Sprintf("the maximum QPS to the upstream cluster apiserver, default to %v", o.ProxyClientQPS))
	fs.IntVar(&o.ProxyClientBurst, "proxy-client-burst", o.ProxyClientBurst,
		fmt.Sprintf("the maximun burst for thorttle to the upstream cluster apiserver, default to %v", o.ProxyClientBurst))
	fs.StringVar(&o.UpstreamMaster, "proxy-upstream-master", o.UpstreamMaster, "upstream apiserver master address")
	fs.StringVar(&o.BindAddress, "proxy-bind-address", o.BindAddress, "The server address of the tenants' kubeconfig file, N.B. this address should be a valid server address of the client-ca-file.")
	fs.IntVar(&o.SecurePort, "proxy-secure-port", o.SecurePort, "The port on which the kubezoo used to serve HTTPS with authentication and authorization.")
	fs.StringVar(&o.ClientCAKeyFile, "client-ca-key-file", o.ClientCAKeyFile, "Filename containing a PEM-encoded RSA or ECDSA private key used to sign tenant certificates.")
	return
}

func (o *ProxyOptions) Validate() []error {
	if o == nil {
		return nil
	}

	errors := []error{}

	if len(o.ProxyClientCAFile) == 0 {
		errors = append(errors, fmt.Errorf("--proxy-client-ca-file cannot be empty"))
	}
	if len(o.ProxyClientKeyFile) == 0 {
		errors = append(errors, fmt.Errorf("--proxy-client-key-file cannot be empty"))
	}
	if len(o.ProxyClientCertFile) == 0 {
		errors = append(errors, fmt.Errorf("--proxy-client-cert-file cannot be empty"))
	}
	if len(o.ClientCAKeyFile) == 0 {
		errors = append(errors, fmt.Errorf("--client-ca-key-file cannot be empty"))
	}
	if len(o.ClientCAFile) == 0 {
		errors = append(errors, fmt.Errorf("--client-ca-file cannot be empty"))
	}
	if len(o.UpstreamMaster) == 0 {
		errors = append(errors, fmt.Errorf("--proxy-upstream-master cannot be empty"))
	}
	if len(o.BindAddress) == 0 {
		errors = append(errors, fmt.Errorf("--proxy-bind-address cannot be empty"))
	}
	if o.SecurePort < 1 || o.SecurePort > 65535 {
		errors = append(errors, fmt.Errorf("--proxy-secure-port %v must be between 1 and 65535, inclusive. It cannot be turned off with 0", o.SecurePort))
	}
	return errors
}
