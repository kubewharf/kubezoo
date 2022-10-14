## Quick Start 

This tutorial introduces how to setup KubeZoo on Kubernetes. KubeZoo should work with any 
standard Kubernetes cluster out of box, but for demo purposes, in this tutorial, we will use 
a `kind` cluster as the upstream cluster.

### Prerequisites

Please install the latest version of 
- [kind](https://kind.sigs.k8s.io/docs/user/quick-start/#installation)
- [kubectl](https://kubernetes.io/docs/tasks/tools/)
- [yq](https://github.com/mikefarah/yq#install)
- [cfssl](https://github.com/cloudflare/cfssl#installation)

### Create `kind` cluster and run kubezoo on it

Run the following command to create local kubezoo enviroment

```console
make local-up
```

When all processes are ready, kubezoo will run on local port 6443, make sure that port is not occupied by another application and you will see the following output
```
Export kubezoo server to 6443
Forwarding from 127.0.0.1:6443 -> 6443
Forwarding from [::1]:6443 -> 6443
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
default              kubezoo-0                                    1/1     Running   0          34m
default              kubezoo-etcd-0                               1/1     Running   0          34m
default              test                                         1/1     Running   0          2m41s
...
```

[rfc1123-label]: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#dns-label-names
