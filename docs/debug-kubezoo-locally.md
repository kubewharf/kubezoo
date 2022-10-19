## Debug Locally

This tutorial introduces how to debug KubeZoo locally, with a `kind` cluster as the upstream cluster.

### Prerequisites

Please install the latest version of [kind](https://kind.sigs.k8s.io/docs/user/quick-start/#installation), [kubectl](https://kubernetes.io/docs/tasks/tools/), [yq](https://github.com/mikefarah/yq#install), [etcd](https://etcd.io/docs/v3.4/install/) and [cfssl](https://github.com/cloudflare/cfssl#installation)

### Create `kind` cluster and set kubectl context

```console
kind create cluster
kubectl config use-context kind-kind
```

### Create secrets storing required certificates/keys

We need two sets of certificates/keys to run KubeZoo:
1. certificates/keys used to secure the communication between tenant clients
   and KubeZoo, which will be generated locally and store in the
   subdirectory `_output/kubezoo` under KubeZoo root directory;
2. certificates/keys used to secure the communication between KubeZoo and the
   upstream cluster, which will need to be fetched from the upstream cluster and
   store in the secret `_output/kubezoo` under KubeZoo root directory

```console
$ bash $KUBEZOO_PATH/hack/lib/gen_pki.sh gen_pki_setup_ctx_print_parameters
```

Upon success, the parameters to run KubeZoo are printed

```console
--allow-privileged=true
--apiserver-count=1
--cors-allowed-origins=.*
--delete-collection-workers=1
--etcd-prefix=/zoo
...
```

This script will also setup a `kubectl` context for accessing `KubeZoo`
```console
$ kubectl config get-contexts
CURRENT   NAME        CLUSTER     AUTHINFO    NAMESPACE
*         kind-kind   kind-kind   kind-kind
          zoo         zoo         zoo-admin
```

### Launch `etcd` locally
```console
$ etcd
{"level":"info","ts":"2021-09-17T09:19:32.783-0400","caller":"etcdmain/etcd.go:72","msg":... }
⋮
```

### Run `KubeZoo` with printed parameters
```console
kubezoo --allow-privileged=true
--apiserver-count=1
--cors-allowed-origins=.*
--delete-collection-workers=1
--etcd-prefix=/zoo
...
```

### Connect to the KubeZoo!

```console
$ kubectl api-resources --context zoo
NAME                              SHORTNAMES   APIVERSION                        NAMESPACED   KIND
...
tenants                                        tenant.kubezoo.io/v1alpha1       false        Tenant
```

### Create a tenant

```console
$ kubectl apply -f config/setup/sample_tenant.yaml --context zoo
tenant.tenant.kubezoo.io/111111 created
```

The tenant name must be a valid 6-character [RFC 1123][rfc1123-label] DNS label prefix (`[A-Za-z0-9][A-Za-z0-9\-]{5}`).

### Get the kubeconfigs of the tenant

```console
$ kubectl get tenant 111111 --context zoo -o jsonpath='{.metadata.annotations.kubezoo\.io\/tenant\.kubeconfig\.base64}' | base64 --decode > 111111.kubeconfig
```

### Create a pod as the tenant

```console
$ kubectl apply --kubeconfig 111111.kubeconfig -f- <<EOF
apiVersion: v1
kind: Pod
metadata:
  name: test
spec:
  containers:
  - name: test
    image: busybox
    command:
    - tail
    args:
    - -f
    - /dev/null
EOF
pod/test created
```

Get the pod as the tenant

```console
$ kubectl get po --kubeconfig 111111.kubeconfig
NAME   READY   STATUS    RESTARTS   AGE
test   1/1     Running   0          44s
```

Get the pod as the cluster administrator

```console
$ kubectl get po -A
NAMESPACE            NAME                                         READY   STATUS    RESTARTS   AGE
111111-default       test                                         1/1     Running   0          2m28s
...
```

[rfc1123-label]: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#dns-label-names
