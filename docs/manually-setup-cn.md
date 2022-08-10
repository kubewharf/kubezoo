## 快速开始

本文介绍如何在 Kubernetes 集群上部署 KubeZoo. 虽然 KubeZoo 可以对接任何标准的 Kubernetes 集群，但是作为样例,
本文会以 `kind` 集群作为上游集群.

### 前置条件

请安装最新版本的 [kind](https://kind.sigs.k8s.io/docs/user/quick-start/#installation) 
和 [kubectl](https://kubernetes.io/docs/tasks/tools/)

### 创建 `kind` 集群并配置 kubectl 上下文 

```console
kind create cluster
kubectl config use-context kind-kind
```

### 创建用于存储证书和秘钥的 secrets


我们需要两套证书/秘钥用来运行 KubeZoo:
1. 用于加密租户客户端和 KubeZoo 之间流量的证书和秘钥，本地生成后存储在 secret `kubezoo-pki` 中；
2. 用于加密 KubeZoo 和上游集群之间流量的证书和秘钥，这部分证书/秘钥需要从上游集群获取，
   然后存储在 secret  `upstream-pki` 中.

```console
$ bash $KUBEZOO_PATH/hack/lib/gen_pki.sh
```

如果上述命令成功运行，你将看到下面两个 secret:

```console
$ kubectl get secrets
NAME                  TYPE                                  DATA   AGE
default-token-r5kd8   kubernetes.io/service-account-token   3      5h54m
kubezoo-pki           Opaque                                4      5h37m
upstream-pki          Opaque                                4      5h41m
```

上述脚本也会创建一个 `kubectl` 上下文用于访问 `KubeZoo`
```console
$ kubectl config get-contexts
CURRENT   NAME        CLUSTER     AUTHINFO    NAMESPACE
*         kind-kind   kind-kind   kind-kind
          zoo         zoo         zoo-admin
```

### 搭建 KubeZoo 和 Etcd

KubeZoo 可以和上游集群共享 Etcd 集群，但是在本样例中，我们将会为 KubeZoo 创建一个独立的 Etcd 集群.
集群配置都被放在 yaml 文件中：`$KUBEZOO_PATH/config/setup/all_in_one.yaml`

```console
kubectl apply -f $KUBEZOO_PATH/config/setup/all_in_one.yaml
```

如果一切顺利，我们会看到相关的资源已经就绪.

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

### 暴露 KubeZoo service

```console
$ kubectl port-forward svc/kubezoo 6443:6443
Forwarding from 127.0.0.1:6443 -> 6443
Forwarding from [::1]:6443 -> 6443
```

### 访问 KubeZoo!

```console
$ kubectl api-resources --context zoo
NAME                              SHORTNAMES   APIVERSION                        NAMESPACED   KIND
...
tenants                                        tenant.kubezoo.io/v1alpha1       false        Tenant
```

### 创建一个租户

```console
$ kubectl apply -f config/setup/sample_tenant.yaml --context zoo
tenant.tenant.kubezoo.io/111111 created
```

### 获取租户的 kubeconfigs 文件

```console
$ kubectl get tenant 111111 --context zoo -o jsonpath='{.metadata.annotations.kubezoo\.io\/tenant\.kubeconfig\.base64}' | base64 --decode > 111111.kubeconfig
```

### 以租户的身份创建一个 pod

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

以租户的身份获取 pod

```console
$ kubectl get po --kubeconfig 111111.kubeconfig
NAME   READY   STATUS    RESTARTS   AGE
test   1/1     Running   0          44s
```

以集群管理员的身份获取 pod

```console
$ kubectl get po -A
NAMESPACE            NAME                                         READY   STATUS    RESTARTS   AGE
111111-default       test                                         1/1     Running   0          2m28s
default              kubezoo-0                                    1/1     Running   0          34m
default              kubezoo-etcd-0                               1/1     Running   0          34m
default              test                                         1/1     Running   0          2m41s
...
```
