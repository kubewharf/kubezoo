## Quick Start 

This tutorial introduces how to setup KubeZoo on Kubernetes. KubeZoo should work with any 
standard Kubernetes cluster out of box, but for demo purposes, in this tutorial, we will use 
a `kind` cluster as the upstream cluster.

### Prerequisites

Please install the latest version of [kind](https://kind.sigs.k8s.io/docs/user/quick-start/#installation) 
and [kubectl](https://kubernetes.io/docs/tasks/tools/)

### Create `kind` cluster and set kubectl context

```console
kind create cluster
kubectl config use-context kind-kind
```

### Create secrets storing required certificates/keys

We need two sets of certificates/keys to run KubeZoo: 
1. certificates/keys used to secure the communication between tenant clients 
and KubeZoo, which will be generated locally and store in the 
secret `kubezoo-pki`; 
2. certificates/keys used to secure the communication between KubeZoo and the 
upstream cluster, which will need to be fetched from the upstream cluster and 
store in the secret `upstream-pki`

```console
$ bash $KUBEZOO_PATH/hack/lib/gen_pki.sh
```

Upon success, you should be able to see the two secrets

```console
$ kubectl get secrets
NAME                  TYPE                                  DATA   AGE
default-token-r5kd8   kubernetes.io/service-account-token   3      5h54m
kubezoo-pki           Opaque                                4      5h37m
upstream-pki          Opaque                                4      5h41m
```

This script will also setup a `kubectl` context for accessing `KubeZoo`
```console
$ kubectl config get-contexts
CURRENT   NAME        CLUSTER     AUTHINFO    NAMESPACE
*         kind-kind   kind-kind   kind-kind
          zoo         zoo         zoo-admin
```

### Set up KubeZoo and Etcd

KubeZoo can share the Etcd cluster with the upstream cluster, but in this demo, 
we will create an independent Etcd cluster for KubeZoo. We have put all 
required yaml in `$KUBEZOO_PATH/config/setup/all_in_one.yaml`

```console
kubectl apply -f $KUBEZOO_PATH/config/setup/all_in_one.yaml
```

if everything goes as expected, we should see the related resources are ready.

```console
$ kubectl get all
NAME                 READY   STATUS    RESTARTS   AGE
pod/kubezoo-0        1/1     Running   0          4h58m
pod/kubezoo-etcd-0   1/1     Running   0          4h58m

NAME                   TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)          AGE
service/kubernetes     ClusterIP   *               <none>        443/TCP          6h21m
service/kubezoo        NodePort    *               <none>        6443:30485/TCP   4h58m
service/kubezoo-etcd   ClusterIP   None            <none>        <none>           4h58m

NAME                            READY   AGE
statefulset.apps/kubezoo        1/1     4h58m
statefulset.apps/kubezoo-etcd   1/1     4h58m
```

### Export the KubeZoo service

```console
$ kubectl port-forward svc/kubezoo 6443:6443
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
