# Developer Guide

There's a `Makefile` in the root folder. Here are some common options:

Build KubeZoo binary `kube-zoo`
```bash
make build
```

Build for specific architecture. (`amd64`,`arm`,`arm64`)
```bash
GOOS=linux GOARCH=arm64 make build
```

Build all docker images for all supported architectures.
```bash
make release
```

Build all docker images for specific architecture.
```bash
make release ARCH=arm64
```

Build kubezoo-e2e-test binary to test KubeZoo.
```bash
make e2e
```

Parameters to debug KubeZoo binary locally
```bash
--allow-privileged=true
--apiserver-count=1
--cors-allowed-origins=.*
--delete-collection-workers=1
--etcd-prefix=/zoo
--etcd-servers=http://localhost:2379 # run etcd server locally
--event-ttl=1h0m0s
--logtostderr=true
--max-requests-inflight=1002
--service-cluster-ip-range=192.168.0.1/16
--service-node-port-range=20000-32767
--storage-backend=etcd3
--authorization-mode=AlwaysAllow
--client-ca-file=/path_to_client_ca_file
--client-ca-key-file=/path_to_client_ca_key_file
--tls-cert-file=/path_to_tls_cert_file
--tls-private-key-file=/path_to_tls_private_key_file
--service-account-key-file=/path_to_upstream_service_account_key_file
--service-account-issuer=foo
--service-account-signing-key-file=/path_to_upstream_service_account_signing_key_file
--proxy-client-cert-file=/path_to_proxy_client_cert_file
--proxy-client-key-file=/path_to_proxy_client_key_file
--proxy-client-ca-file=/path_to_proxy_client_ca_file
--request-timeout=10m
--watch-cache=true
--proxy-upstream-master=https://127.0.0.1:49329 # run upstream cluster with kind
--service-account-lookup=false
--proxy-bind-address=127.0.0.1
--proxy-secure-port=6443
--api-audiences=foo
```