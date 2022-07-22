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
