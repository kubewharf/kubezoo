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

set -x

# project_info generates the project information and values
# for 'ldflags -X' option.
project_info() {
    PROJECT_INFO_PKG=${ZOO_MOD}/pkg/projectinfo
    echo "-X ${PROJECT_INFO_PKG}.projectPrefix=${PROJECT_PREFIX}"
    echo "-X ${PROJECT_INFO_PKG}.labelPrefix=${LABEL_PREFIX}"
    echo "-X ${PROJECT_INFO_PKG}.gitVersion=${GIT_VERSION}"
    echo "-X ${PROJECT_INFO_PKG}.gitCommit=${GIT_COMMIT}"
    echo "-X ${PROJECT_INFO_PKG}.buildDate=${BUILD_DATE}"
}

# build_binaries build the binary locally.
build_binaries() {
    local goflags goldflags gcflags
    goldflags="${GOLDFLAGS:--s -w $(project_info)}"
    gcflags="${GOGCFLAGS:-}"
    goflags=${GOFLAGS:-}

    local arg
    local targets=()

    for arg; do
        if [[ "$arg" == -* ]]; then
            # args starting with a dash are flags for go.
            goflags+=("$arg")
        else
            targets+=("$arg")
        fi
    done

    # target_bin_dir contains the GOOS and GOARCH
    # eg: ./_output/bin/darwin/arm64/
    local target_bin_dir=$ZOO_LOCAL_BIN_DIR/$(go env GOOS)/$(go env GOARCH)

    rm -rf $target_bin_dir
    mkdir -p $target_bin_dir
    cd $target_bin_dir

    if [[ ${#targets[*]} == 0 ]]; then
        targets=(kubezoo clusterresourcequota)
    fi

    for target in "${targets[@]}"; do
        echo "Building $target"
        go build -o $target \
            -ldflags "${goldflags:-}" \
            -gcflags "${gcflags:-}" \
            $goflags $ZOO_ROOT/cmd/${target}
    done

}
