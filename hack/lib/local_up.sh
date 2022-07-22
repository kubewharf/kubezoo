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

cleanup() {
    rm -rf $ZOO_ROOT/_output
    kind delete clusters $CLUSTER_NAME
}

cleanup_on_err() {
    if [[ $? -ne 0 ]]; then
        cleanup
    fi
}

preflight() {
    echo "Preflight Check..."
    for bin in "${REQUIRED_CMD[@]}"; do
        command -v ${bin} > /dev/null 2>&1 || \
            echo "$bin is not installed"; exit 0
    done
}

local_up_openyurt() {
    echo "Creating the kind cluster $CLUSTER_NAME..."
    kind create cluster --name $CLUSTER_NAME 
    echo "Generating PKI and context..."
    gen_pki_setup_ctx
    echo "Setting up kubezoo on $CLUSTER_NAME..."
    kubectl apply -f $ZOO_ROOT/config/setup/all_in_one.yaml 
}
