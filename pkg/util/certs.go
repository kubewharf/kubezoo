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
	"crypto"
	cryptorand "crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math"
	"math/big"
	"net"
	"time"

	"github.com/pkg/errors"
)

const (
	AnnotationTenantKubeConfigBase64 = "kubezoo.io/tenant.kubeconfig.base64"
	KubeZooClusterName               = "kube-zoo"

	RsaKeySize = 2048
	// CertificateValidity defines the validity, i.e., 10 Years, for all the signed certificates.
	CertificateValidity = time.Hour * 24 * 365 * 10
)

// Config contains the basic fields required for creating a certificate
type Config struct {
	CommonName         string
	Organization       []string
	OrganizationalUnit []string
	AltNames           AltNames
	Usages             []x509.ExtKeyUsage
}

// AltNames contains the domain names and IP addresses that will be added
// to the API Server's x509 certificate SubAltNames field. The values will
// be passed directly to the x509.Certificate object.
type AltNames struct {
	DNSNames []string
	IPs      []net.IP
}

// EncodeCertPEM returns PEM-endcoded certificate data.
func EncodeCertPEM(cert *x509.Certificate) []byte {
	block := pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert.Raw,
	}
	return pem.EncodeToMemory(&block)
}

// EncodePrivateKeyPEM returns PEM-encoded private key data.
func EncodePrivateKeyPEM(key *rsa.PrivateKey) []byte {
	block := pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}

	return pem.EncodeToMemory(&block)
}

// NewTenantCertAndKey creates new certificate and key for the denoted tenant.
func NewTenantCertAndKey(caFile, caKeyFile, tenantID string) (*x509.Certificate, *rsa.PrivateKey, error) {
	// load ca, ca-key from files
	tlsCert, err := tls.LoadX509KeyPair(caFile, caKeyFile)
	if err != nil {
		return nil, nil, err
	}

	key, ok := tlsCert.PrivateKey.(crypto.Signer)
	if !ok {
		return nil, nil, fmt.Errorf("private key is not crypto.Signer")
	}
	cert, err := x509.ParseCertificate(tlsCert.Certificate[0])
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse cert: %v", err)
	}
	// generate the certificate config
	config := &Config{
		OrganizationalUnit: []string{tenantID},
		CommonName:         tenantID + "-admin",
		Usages:             []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
	}

	return NewCertAndKey(cert, key, config)
}

// NewCertAndKey creates new certificate and key by passing the certificate authority certificate and key.
func NewCertAndKey(caCert *x509.Certificate, caKey crypto.Signer, config *Config) (*x509.Certificate, *rsa.PrivateKey, error) {
	key, err := NewPrivateKey()
	if err != nil {
		return nil, nil, errors.Wrap(err, "unable to create private key")
	}

	cert, err := NewSignedCert(config, key, caCert, caKey)
	if err != nil {
		return nil, nil, errors.Wrap(err, "unable to sign certificate")
	}

	return cert, key, nil
}

// NewPrivateKey creates an RSA private key.
func NewPrivateKey() (*rsa.PrivateKey, error) {
	return rsa.GenerateKey(cryptorand.Reader, RsaKeySize)
}

// NewSignedCert creates a signed certificate using the given CA certificate and key.
func NewSignedCert(cfg *Config, key crypto.Signer, caCert *x509.Certificate, caKey crypto.Signer) (*x509.Certificate, error) {
	serial, err := cryptorand.Int(cryptorand.Reader, new(big.Int).SetInt64(math.MaxInt64))
	if err != nil {
		return nil, err
	}
	if len(cfg.CommonName) == 0 {
		return nil, errors.New("must specify a CommonName")
	}
	if len(cfg.Usages) == 0 {
		return nil, errors.New("must specify at least one ExtKeyUsage")
	}
	// OrganizationalUnit is required by Tenant authentication
	if len(cfg.OrganizationalUnit) == 0 {
		return nil, errors.New("must specify a OrganizationalUnit")
	}

	certTmpl := x509.Certificate{
		Subject: pkix.Name{
			CommonName:         cfg.CommonName,
			Organization:       cfg.Organization,
			OrganizationalUnit: cfg.OrganizationalUnit,
		},
		DNSNames:     cfg.AltNames.DNSNames,
		IPAddresses:  cfg.AltNames.IPs,
		SerialNumber: serial,
		NotBefore:    caCert.NotBefore,
		NotAfter:     time.Now().Add(CertificateValidity).UTC(),
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  cfg.Usages,
	}
	certDERBytes, err := x509.CreateCertificate(cryptorand.Reader, &certTmpl, caCert, key.Public(), caKey)
	if err != nil {
		return nil, err
	}
	return x509.ParseCertificate(certDERBytes)
}
