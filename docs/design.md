# Design Document

## Introduction

KubeZoo implements the multi-tenancy function of the control plane based on the core concept of `protocol conversion`.
By adding the unique identifier of the tenant to the "name/namespace" and other fields of the resource, KubeZoo solves 
the problem of naming conflicts between resources with the same name of different tenants in the same upstream Kubernetes cluster.

![img](./img/design-ideas.png)

## Architecture Overview

As shown in the figure below, KubeZoo is an independently deployed service. It consists of a stateless kubezoo-server 
and etcd, with good horizontal scalability. KubeZoo provides a unified Kubernetes access portal for tenants, and forwards
requests from tenants to the upstream Kubernetes cluster after protocol translation, so the upstream cluster actually 
complete the resource expression. The control plane of the upstream Kubernetes mainly includes Master and Etcd. We recommend 
implementing the data plane by elastic containers (such as AWS Fargate, Aliyun ECI, etc.), to have stronger network and
storage isolation functions, such as VPC, etc.

![img](./img/architecture-overview.png)

- KubeZoo: consists of the stateless kubezoo-server and Etcd.
- K8S Master：
    - Master：apiserver / scheduler / controller-manager / Etcd
    - Virtual Kubelet: connect the control plane and the data plane, connect the elastic container services of different
      public cloud vendors, and finally complete the expression of resources such as Pod/Service in the Master;
- Container Instance Service：Public cloud elastic container services, such as AWS Fargate, Aliyun ECI, etc.

## Tenant Management

KubeZoo has a built-in Tenant object that describes basic tenant information. The name is a mandatory field, globally 
unique and fixed 6-bit string (including characters or numbers). It can theoretically manage 2176782336 tenants (36 ^ 6).
The Tenant object is stored in KubeZoo's etcd in the following format：
````
apiVersion: tenant.kubezoo.io/v1alpha1
kind: Tenant
metadata:
name: "foofoo"
annotations:
  ...... # add schema for tenant(optional)
spec:
  id: 0
````

KubeZoo provides the function of certificate issuance, and administrators have the ability to manage the life cycle of 
tenant. Whenever the administrator creates a tenant, an X509 certificate is issued for the tenant. The certificate contains 
the tenant's information, such as name, etc., and writes annotations. At the same time, the built-in namespace, rbac, etc. 
of each tenant are synchronized to upstream Kubernetes.

````
apiVersion: tenant.kubezoo.io/v1alpha1
kind: Tenant
metadata:
  name: "foofoo"
  annotations:
    kubezoo.io/tenant.kubeconfig.base64: YXBpVmVy...ExRbz0K
    ......
spec:
  id: 0
status: {}
````

Whenever an administrator deletes a tenant, the tenant resource recovery is triggered. And KubeZoo removes all the tenant's
resources of Kubernetes upstream, cleans up meta information on the KubeZoo side. Since tenant lifecycle management is 
essentially the management of tenant object meta-information, certificate issuance and resource synchronization. The process 
is simple and does not require the creation of physical Master/Etcd and computing resource pools, so KubeZoo has lightweight,
second-level massive tenant lifecycle management function.

## Protocol Conversion

### Namespace Scope Resource

Kubernetes has about 40+ namespace scoped resources, such as deployment / statefulset / pod / configmap, etc. By associating
tenant information in the namespace field of each resource, the multi-tenancy function of namespace scope resources is achieved.

![img](./img/namespace-scope-resource.png)

### Cluster Scope Resource

Kubernetes has about 20+ cluster scope resources, such as pv / namespace / storageclass, etc. By associating tenant 
information in the name, the multi-tenancy function of cluster scope resources is achieved.

![img](./img/cluster-scope-resource.png)

### Custom Resource

Custom Resource Definition (CRD) is a special cluster scope resource, whose name consists of group and plural, and we 
choose to associate tenant information in the group prefix.

![img](./img/custom-resource.png)
