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

set -xue

readonly TEMP_DIR=$(mktemp -d)
readonly UPSTREAM_DIR=$TEMP_DIR/upstream
readonly KUBEZOO_DIR=$TEMP_DIR/kubezoo

get_upstream_pki_kind() {
    local context_name=$(kubectl config current-context)
    if [ ${context_name::5} != "kind-" ]; then
        echo "Current kubectl context is not a kind cluster" >&2
        exit 1
    fi
    local kind_cluster_name=${context_name:5}
    local kind_docker=$(docker ps --filter "name=${kind_cluster_name}-control-plane" --format "{{.ID}}")

    [ -z $UPSTREAM_DIR ] || mkdir -p $UPSTREAM_DIR
    docker cp $kind_docker:/etc/kubernetes/pki/sa.pub $UPSTREAM_DIR/sa.pub
    docker cp $kind_docker:/etc/kubernetes/pki/apiserver.crt $UPSTREAM_DIR/apiserver.crt
    docker cp $kind_docker:/etc/kubernetes/pki/apiserver.key $UPSTREAM_DIR/apiserver.key
    yq eval '.users.[]|select(.name=="'${context_name}'")|.user.client-certificate-data' ~/.kube/config | base64 \
        --decode > $UPSTREAM_DIR/client.crt
    yq eval '.users.[]|select(.name=="'${context_name}'")|.user.client-key-data' ~/.kube/config | base64 \
        --decode > $UPSTREAM_DIR/client-key.crt
    yq eval '.clusters.[]|select(.name=="'${context_name}'")|.cluster.certificate-authority-data' \
        ~/.kube/config | base64 --decode > $UPSTREAM_DIR/ca.crt
}

gen_kubezoo_pki() {
    [ -z $KUBEZOO_DIR ] || mkdir -p $KUBEZOO_DIR
    gen_ca
    gen_admin_cert
    gen_kubernetes_cert
}

gen_ca(){
    cat > $KUBEZOO_DIR/ca-config.json <<EOF
{
  "signing": {
    "default": {
      "expiry": "8760h"
    },
    "profiles": {
      "kubernetes": {
        "usages": ["signing", "key encipherment", "server auth", "client auth"],
        "expiry": "8760h"
      }
    }
  }
}
EOF

    cat > $KUBEZOO_DIR/ca-csr.json <<EOF
{
  "CN": "Kubernetes",
  "key": {
    "algo": "rsa",
    "size": 2048
  },
  "names": [
    {
      "C": "US",
      "L": "Sunnyvale",
      "O": "KubeZoo",
      "OU": "CA",
      "ST": "CA"
    }
  ]
}
EOF
    cd $KUBEZOO_DIR
    cfssl gencert -initca $KUBEZOO_DIR/ca-csr.json | cfssljson -bare ca
    cd -
}

gen_kubernetes_cert() {

    KUBERNETES_HOSTNAMES=kubernetes,kubernetes.default,kubernetes.default.svc,kubernetes.default.svc.cluster,kubernetes.svc.cluster.local,localhost,host.minikube.internal

    cat > $KUBEZOO_DIR/kubernetes-csr.json <<EOF
{
  "CN": "kubernetes",
  "key": {
    "algo": "rsa",
    "size": 2048
  },
  "names": [
    {
      "C": "US",
      "L": "Sunnyvale",
      "O": "KubeZoo",
      "OU": "KubeZoo",
      "ST": "CA"
    }
  ]
}
EOF
    cd $KUBEZOO_DIR
    cfssl gencert \
      -ca=$KUBEZOO_DIR/ca.pem \
      -ca-key=$KUBEZOO_DIR/ca-key.pem \
      -config=$KUBEZOO_DIR/ca-config.json \
      -hostname=127.0.0.1,${KUBERNETES_HOSTNAMES} \
      -profile=kubernetes \
      $KUBEZOO_DIR/kubernetes-csr.json | cfssljson -bare kubernetes
    cd -
}

gen_admin_cert() {
    cat > $KUBEZOO_DIR/admin-csr.json <<EOF
{
  "CN": "admin",
  "key": {
    "algo": "rsa",
    "size": 2048
  },
  "names": [
    {
      "C": "US",
      "L": "Sunnyvale",
      "O": "system:masters",
      "OU": "KubeZoo",
      "ST": "CA"
    }
  ]
}
EOF
    cd $KUBEZOO_DIR
    cfssl gencert \
      -ca=$KUBEZOO_DIR/ca.pem \
      -ca-key=$KUBEZOO_DIR/ca-key.pem \
      -config=$KUBEZOO_DIR/ca-config.json \
      -profile=kubernetes \
      $KUBEZOO_DIR/admin-csr.json | cfssljson -bare admin
    cd -
}

create_pki_secret() {
    kubectl create secret generic kubezoo-pki \
        --from-file=ca-key.pem=$KUBEZOO_DIR/ca-key.pem \
        --from-file=ca.pem=$KUBEZOO_DIR/ca.pem \
        --from-file=kubernetes-key.pem=$KUBEZOO_DIR/kubernetes-key.pem \
        --from-file=kubernetes.pem=$KUBEZOO_DIR/kubernetes.pem \
        && 
    kubectl create secret generic upstream-pki \
        --from-file=sa.pub=$UPSTREAM_DIR/sa.pub \
        --from-file=client.crt=$UPSTREAM_DIR/client.crt \
        --from-file=client-key.crt=$UPSTREAM_DIR/client-key.crt \
        --from-file=ca.crt=$UPSTREAM_DIR/ca.crt
}

set_context() {
    kubectl config set-cluster zoo \
      --certificate-authority=$KUBEZOO_DIR/ca.pem \
      --embed-certs=true \
      --server=https://127.0.0.1:6443
    kubectl config set-credentials zoo-admin \
      --client-certificate=$KUBEZOO_DIR/admin.pem \
      --client-key=$KUBEZOO_DIR/admin-key.pem \
      --embed-certs=true
    kubectl config set-context zoo \
      --cluster=zoo \
      --user=zoo-admin
}

gen_pki_setup_ctx() {
    get_upstream_pki_kind
    gen_kubezoo_pki
    create_pki_secret
    set_context
}

gen_pki_setup_ctx
