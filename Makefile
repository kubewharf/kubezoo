# Copyright 2022 The KubeZoo Authors.
# 
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
# 
#     http://www.apache.org/licenses/LICENSE-2.0
# 
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

TARGET_PLATFORMS ?= linux/amd64
IMAGE_REPO ?= kubezoo
IMAGE_TAG ?= $(shell git describe --abbrev=0 --tags --always)
GIT_COMMIT = $(shell git rev-parse HEAD)

ifeq ($(shell git tag --points-at ${GIT_COMMIT}),)
GIT_VERSION=$(IMAGE_TAG)-$(shell echo ${GIT_COMMIT} | cut -c 1-7)
else
GIT_VERSION=$(IMAGE_TAG)
endif

ifneq ($(IMAGE_TAG), $(shell git describe --abbrev=0 --tags --always))
GIT_VERSION=$(IMAGE_TAG)
endif

DOCKER_BUILD_ARGS = --build-arg GIT_VERSION=${GIT_VERSION}

# ENVTEST_K8S_VERSION refers to the version of kubebuilder assets to be downloaded by envtest binary.
ENVTEST_K8S_VERSION = 1.19

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

ENVTEST = $(shell pwd)/bin/setup-envtest
KUBECODEGEN = $(shell pwd)/bin/kube-codegen

# go-get-tool will 'go get' any package $2 and install it to $1.
PROJECT_DIR := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))
define go-get-tool
@[ -f $(1) ] || { \
set -xe ;\
TMP_DIR=$$(mktemp -d) ;\
cd $$TMP_DIR ;\
go mod init tmp ;\
echo "Downloading $(2)" ;\
go get $(2) ;\
cd $(shell go list -f '{{ .Dir }}' -m $(2)) ;\
GOBIN=$(PROJECT_DIR)/bin go install $(3) ;\
rm -rf $$TMP_DIR ;\
}
endef

.PHONY: envtest
envtest: ## Download envtest-setup locally if necessary.
	$(call go-get-tool,$(ENVTEST),sigs.k8s.io/controller-runtime/tools/setup-envtest@latest,./)

.PHONY: kube-codegen
kube-codegen:
	$(call go-get-tool,$(KUBECODEGEN),github.com/zoumo/kube-codegen@v0.2.0,./cmd/kube-codegen)


.PHONY: e2e
e2e: envtest
	KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) -p path)" go test ./pkg/controller -coverprofile cover.out

.PHONY: all
all: test build

.PHONY: fmt
fmt: ## Format the code.
	go fmt ./pkg/... ./cmd/...

.PHONY: vet
vet: ## Exam the code and report suspicious constructs.
	go vet ./pkg/... ./cmd/...

test: fmt vet

.PHONY: build
build:
	bash hack/make-rules/build.sh $(WHAT)

.PHONY: clean
clean:
	-rm -rf _output	

.PHONY: docker-build
docker-build: ## Build the kubezoo container image. Please make sure there are at least 6GB free memory for docker daemon.
	docker buildx build --no-cache --load ${DOCKER_BUILD_ARGS} --platform ${TARGET_PLATFORMS} -f hack/dockerfiles/kubezoo.Dockerfile . -t ${IMAGE_REPO}/kubezoo:${GIT_VERSION}
	docker buildx build --no-cache --load ${DOCKER_BUILD_ARGS} --platform ${TARGET_PLATFORMS} -f hack/dockerfiles/clusterresourcequota.Dockerfile . -t ${IMAGE_REPO}/clusterresourcequota:${GIT_VERSION}

.PHONY: docker-push
docker-push: ## Build and push the kubezoo container image.
	docker buildx rm kubezoo-container-builder || true
	docker buildx create --use --name=kubezoo-container-builder
	docker buildx build --no-cache --push ${DOCKER_BUILD_ARGS} --platform ${TARGET_PLATFORMS} -f hack/dockerfiles/kubezoo.Dockerfile . -t ${IMAGE_REPO}/kubezoo:${GIT_VERSION}
	docker buildx build --no-cache --push ${DOCKER_BUILD_ARGS} --platform ${TARGET_PLATFORMS} -f hack/dockerfiles/clusterresourcequota.Dockerfile . -t ${IMAGE_REPO}/clusterresourcequota:${GIT_VERSION}


.PHONY: local-up
local-up: ## Setup kubezoo locally on a kind cluster
	bash hack/make-rules/local_up.sh

code-gen: kube-codegen
	@kube-codegen code-gen \
		--generators deepcopy,protobuf,openap,crd \
		--go-header-file hack/boilerplate.go.txt \
		--client-path pkg/generated \
		--apis-module github.com/kubewharf/kubezoo \
		--apis-path pkg/apis 

client-gen: 
	@kube-codegen client-gen  \
	--go-header-file hack/boilerplate.go.txt \
	--client-path pkg/generated \
	--apis-module github.com/kubewharf/kubezoo \
	--apis-path pkg/apis \
	--clientset-dir=clientset/versioned \
	--informers-dir=informers/externalversions

