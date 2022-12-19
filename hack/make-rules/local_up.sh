#!/usr/bin/env bash

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

set -eu

ZOO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd -P)"
source "${ZOO_ROOT}/hack/lib/init.sh"

readonly REQUIRED_CMD=(
    go
    docker
    kubectl
    kind
)

readonly LOCAL_UP_IMAGE_TAG="v0.2.0"
readonly LOCAL_ARCH=$(go env GOHOSTARCH)
readonly LOCAL_OS=$(go env GOHOSTOS)
readonly CLUSTER_NAME="kubezoo-e2e-test"
readonly KIND_KUBECONFIG=${KIND_KUBECONFIG:-${HOME}/.kube/config}

cleanup() {
    rm -rf $ZOO_ROOT/_output
    if kind get clusters | grep "${CLUSTER_NAME}"; then
        kubectl --context "kind-${CLUSTER_NAME}" delete statefulset --all
        kubectl --context "kind-${CLUSTER_NAME}" delete deployment --all
        kubectl --context "kind-${CLUSTER_NAME}" delete validatingwebhookconfigurations --all
    fi
}

cleanup_on_err() {
    if [[ $? -ne 0 ]]; then
        cleanup
    fi
}

preflight() {
    echo "Preflight Check..."
    for bin in "${REQUIRED_CMD[@]}"; do
        command -v ${bin} >/dev/null 2>&1 || (echo "$bin is not installed" && exit 0)
    done
}

local_up() {
    echo "Creating the kind cluster $CLUSTER_NAME..."
    if kind get clusters | grep "${CLUSTER_NAME}"; then
        cleanup
    else
        kind create cluster --name "${CLUSTER_NAME}"
    fi
    kubectl config use-context "kind-${CLUSTER_NAME}"

    echo "Generating PKI and context..."
    bash "${ZOO_ROOT}"/hack/lib/gen_pki.sh gen_pki_setup_ctx

    # echo "Loading image on $CLUSTER_NAME..."
    docker pull kubezoo/kubezoo:"${LOCAL_UP_IMAGE_TAG}"
    docker pull kubezoo/clusterresourcequota:"${LOCAL_UP_IMAGE_TAG}"
    docker tag kubezoo/kubezoo:"${LOCAL_UP_IMAGE_TAG}" kubezoo/kubezoo:local-up
    docker tag kubezoo/clusterresourcequota:"${LOCAL_UP_IMAGE_TAG}" kubezoo/clusterresourcequota:local-up

    echo "Setting up ClusterResourceQuota on $CLUSTER_NAME..."
    # run quota controller and webhook
    kubectl --context "kind-${CLUSTER_NAME}" apply -f $ZOO_ROOT/_output/setup/quota.yaml
    while ! kubectl --context "kind-${CLUSTER_NAME}" get clusterresourcequota; do
        echo ">> clusterresourcequota is not ready, sleep 1s"
        sleep 1s
    done

    echo "Setting up kubezoo on $CLUSTER_NAME..."
    kubectl apply -f $ZOO_ROOT/config/setup/all_in_one.yaml

    while ! (kubectl --context "kind-${CLUSTER_NAME}" get pods kubezoo-0 | grep "Running"); do
        echo ">> wait for kubezoo server running"
        sleep 1s
    done

    echo "Export kubezoo server to 6443"
    kubectl --context "kind-${CLUSTER_NAME}" port-forward svc/kubezoo 6443:6443

}

preflight
local_up
