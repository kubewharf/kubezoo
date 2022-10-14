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

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE}")/../.." && pwd -P)"
CONTEXT_NAME=$(kubectl config current-context)
if [ ${CONTEXT_NAME::5} != "kind-" ]; then
    echo "Current kubectl context is not a kind cluster" >&2
    exit 1
fi
KIND_SERVER="$(yq eval '.clusters.[]|select(.name=="'${CONTEXT_NAME}'")|.cluster.server' ~/.kube/config)"
readonly TEMP_DIR=$(pwd -P)/_output/pki
readonly UPSTREAM_DIR=$TEMP_DIR/upstream
readonly KUBEZOO_DIR=$TEMP_DIR/kubezoo

get_upstream_pki_kind() {
    [ -z $TEMP_DIR ] || mkdir -p $TEMP_DIR

    local kind_cluster_name=${CONTEXT_NAME:5}
    local kind_docker=$(docker ps --filter "name=${kind_cluster_name}-control-plane" --format "{{.ID}}")

    [ -z $UPSTREAM_DIR ] || mkdir -p $UPSTREAM_DIR
    docker cp $kind_docker:/etc/kubernetes/pki/sa.pub $UPSTREAM_DIR/sa.pub
    docker cp $kind_docker:/etc/kubernetes/pki/sa.key $UPSTREAM_DIR/sa.key
    docker cp $kind_docker:/etc/kubernetes/pki/apiserver.crt $UPSTREAM_DIR/apiserver.crt
    docker cp $kind_docker:/etc/kubernetes/pki/apiserver.key $UPSTREAM_DIR/apiserver.key
    yq eval '.users.[]|select(.name=="'${CONTEXT_NAME}'")|.user.client-certificate-data' ~/.kube/config | base64 \
        --decode >$UPSTREAM_DIR/client.crt
    yq eval '.users.[]|select(.name=="'${CONTEXT_NAME}'")|.user.client-key-data' ~/.kube/config | base64 \
        --decode >$UPSTREAM_DIR/client-key.crt
    yq eval '.clusters.[]|select(.name=="'${CONTEXT_NAME}'")|.cluster.certificate-authority-data' \
        ~/.kube/config | base64 --decode >$UPSTREAM_DIR/ca.crt
}

gen_kubezoo_pki() {
    [ -z $KUBEZOO_DIR ] || mkdir -p $KUBEZOO_DIR
    gen_ca
    gen_admin_cert
    gen_kubernetes_cert
}

gen_kubezoo_quota_pki() {
    [ -z $KUBEZOO_DIR ] || mkdir -p $KUBEZOO_DIR

}

gen_ca() {
    cat >$KUBEZOO_DIR/ca-config.json <<EOF
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

    cat >$KUBEZOO_DIR/ca-csr.json <<EOF
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

    cat >$KUBEZOO_DIR/kubernetes-csr.json <<EOF
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
    cat >$KUBEZOO_DIR/admin-csr.json <<EOF
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

base64file() {
    input=${1}
    if base64 --help | grep GNU >/dev/null; then
        # gnu base64
        base64 -w 0 "${input}"
    elif base64 --help | grep "input" >/dev/null; then
        base64 -b 0 -i "${input}"
    fi
}

gen_quota_webhook_cert() {

    QUOTA_WEBHOOK_HOSTNAMES=kubezoo-cluster-resource-quota.default,kubezoo-cluster-resource-quota.default.svc

    cat >$KUBEZOO_DIR/quota-csr.json <<EOF
{
  "CN": "kubezoo-cluster-resource-quota",
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
        -hostname=127.0.0.1,${QUOTA_WEBHOOK_HOSTNAMES} \
        -profile=kubernetes \
        $KUBEZOO_DIR/quota-csr.json | cfssljson -bare quota-webhook

    cd -

    caBase64="$(base64file ${KUBEZOO_DIR}/ca.pem)"
    mkdir -p _output/setup
    sed "s/{caBundle}/${caBase64}/g" config/setup/quota.tmpl.yaml >_output/setup/quota.yaml
}

create_pki_secret() {
    if kubectl get secret kubezoo-pki; then
        kubectl delete secret kubezoo-pki
    fi
    kubectl create secret generic kubezoo-pki \
        --from-file=ca-key.pem=$KUBEZOO_DIR/ca-key.pem \
        --from-file=ca.pem=$KUBEZOO_DIR/ca.pem \
        --from-file=kubernetes-key.pem=$KUBEZOO_DIR/kubernetes-key.pem \
        --from-file=kubernetes.pem=$KUBEZOO_DIR/kubernetes.pem

    if kubectl get secret upstream-pki; then
        kubectl delete secret upstream-pki
    fi
    kubectl create secret generic upstream-pki \
        --from-file=sa.pub=$UPSTREAM_DIR/sa.pub \
        --from-file=sa.key=$UPSTREAM_DIR/sa.key \
        --from-file=client.crt=$UPSTREAM_DIR/client.crt \
        --from-file=client-key.crt=$UPSTREAM_DIR/client-key.crt \
        --from-file=ca.crt=$UPSTREAM_DIR/ca.crt

    if kubectl get secret quota-webhook-pki; then
        kubectl delete secret quota-webhook-pki
    fi
    kubectl create secret tls quota-webhook-pki \
        --key=$KUBEZOO_DIR/quota-webhook-key.pem \
        --cert=$KUBEZOO_DIR/quota-webhook.pem
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

kubezoo_parametes="

--allow-privileged=true \
--apiserver-count=1 \
--cors-allowed-origins=.* \
--delete-collection-workers=1 \
--etcd-prefix=/zoo \
--etcd-servers=http://localhost:2379 \
--event-ttl=1h0m0s \
--logtostderr=true \
--max-requests-inflight=1002 \
--service-cluster-ip-range=192.168.0.1/16 \
--service-node-port-range=20000-32767 \
--storage-backend=etcd3 \
--authorization-mode=AlwaysAllow \
--client-ca-file=$KUBEZOO_DIR/ca.pem \
--client-ca-key-file=$KUBEZOO_DIR/ca-key.pem \
--tls-cert-file=$KUBEZOO_DIR/kubernetes.pem \
--tls-private-key-file=$KUBEZOO_DIR/kubernetes-key.pem \
--service-account-key-file=$UPSTREAM_DIR/sa.pub \
--service-account-issuer=foo \
--service-account-signing-key-file=$UPSTREAM_DIR/sa.key \
--proxy-client-cert-file=$UPSTREAM_DIR/client.crt \
--proxy-client-key-file=$UPSTREAM_DIR/client-key.crt \
--proxy-client-ca-file=$UPSTREAM_DIR/ca.crt \
--request-timeout=10m \
--watch-cache=true \
--proxy-upstream-master=$KIND_SERVER \
--service-account-lookup=false \
--proxy-bind-address=127.0.0.1 \
--proxy-secure-port=6443 \
--api-audiences=foo
"

print_kubezoo_parameters() {
    echo "$kubezoo_parametes"
}

gen_pki_setup_ctx() {
    get_upstream_pki_kind
    gen_kubezoo_pki
    gen_quota_webhook_cert
    create_pki_secret
    set_context
}

gen_pki_setup_ctx_print_parameters() {
    get_upstream_pki_kind
    gen_kubezoo_pki
    set_context
    print_kubezoo_parameters
}

"$@"
