---
title: Garbage Collection of Tenant Resources
authors:
  - "@caohe"
reviewers:
  - "@Silverglass"
  - "@zoumo"
  - "@SOF3"
creation-date: 2022-10-30
last-updated: 2022-10-30
status: provisional
---

# Garbage Collection of Tenant Resources

## Table of Contents

<!-- toc -->
- [Summary](#summary)
- [Motivation](#motivation)
    - [Goals](#goals)
    - [Non-Goals/Future Work](#non-goalsfuture-work)
- [Proposal](#proposal)
    - [User Stories](#user-stories)
        - [Story 1](#story-1)
        - [Story 2](#story-2)
        - [Story 3](#story-3)
    - [API](#api)
    - [Design Details](#design-details)
- [Alternatives](#alternatives)
- [References](#references)
<!-- /toc -->

## Summary

When a tenant is deleted, all resources related to the tenant should be cleaned up, also known as garbage collection. In the current implementation of KubeZoo, only some kinds of resources of the tenant are cleaned up, which can lead to cluster-scoped resource leaks.

Therefore, this proposal introduces an enhanced garbage collection mechanism to avoid leakage of tenant resources.

## Motivation

### Goals

- Improve the garbage collection mechanism of tenant resources to ensure that all kinds of resources are not leaked

### Non-Goals/Future Work

- When a tenant creates a resource through KubeZoo, add a label to it. Therefore, we can use label selectors to filter the resources of a tenant to improve performance when performing garbage collection. This optimization will be designed and implemented in the follow-up.

## Proposal

### User Stories

#### Story 1

Tenants can create some native cluster-scoped resources through KubeZoo. For example,

- tenants can manage storage by manually creating PersistentVolumes;

- tenants can define some specific PriorityClasses to represent the priority of a Pod when queuing and preempting;

- Tenants can build custom admission webhooks (not yet supported) and configure them through MutatingWebhookConfigurations or ValidatingWebhookConfigurations.

When a tenant is deleted, the tenant's native cluster-scoped resources should be cleaned up.

#### Story 2

Tenants can create some CRDs through KubeZoo to extend their API definitions. When a tenant is deleted, the CRDs defined by the tenant should also be cleaned up.

#### Story 3

Tenants can create some cluster-scoped CRs through KubeZoo, which may be of the following two types:

- tenant-defined CRDs
- system CRDs predefined by the admin

When a tenant is deleted, CRs of the above kinds should also be cleaned up.

## API

Introduce a finalizer for the Tenant object:

```yaml
apiVersion: tenant.kubezoo.io/v1alpha1
kind: Tenant
metadata:
  name: "111111"
  finalizers:
  - kubezoo.io/tenant
spec:
  id: 111111
```

### Design Details

Modify the Tenant Controller's processing logic for Tenant events:

- When a `CREATE` event is watched, add a `kubezoo.io/tenant` finalizer for the Tenant object.

- When an `UPDATE` event is watched, if the Tenant has a `DeletionTimestamp` and a `kubezoo.io/tenant` finalizer,
  
    1. List all cluster-scoped GVRs via DiscoveryClient, and divide them into two groups according to whether they are CRD or not.
       
    2. Traverse the GVR list of CRDs. For each GVR, check whether the CRD belongs to this tenant according to the prefix of its `Group`. If so, delete it via DynamicClient.

        > In this step, CRs associated with CRDs belonging to this tenant will be deleted in cascade. This avoids the overhead of listing all CRs in step (iii) and improves performance.
    
    3. Traverse the list of non-CRD GVRs. For each GVR, list all objects of this kind via DynamicClient. For each object, check whether it belongs to this tenant according to the prefix of its `Name`. If so, delete it via DynamicClient.
        > - In this step, the tenant's namespaces will be cleaned up, which triggers the Namespace Controller to clean up the tenant's Namespaced resources.
        > - In this step, the tenant's CRs associated with system CRDs are also cleaned up.

    4. Remove the `kubezoo.io/tenant` finalizer from the Tenant object.
    
## Alternatives

- **Delete tenant's cluster-scoped resources in cascade through the OwnerReferences mechanism**

    - When a `CREATE` event of a Tenant object is watched, Tenant Controller will
        - Create a dummy cluster-scoped object for the tenant in the upstream cluster, and record the `UID` of this dummy object in the Tenant object.
    
        - Add a `kubezoo.io/tenant` finalizer for the Tenant object.

    - When a tenant creates a cluster-scoped resource through KubeZoo, the proxy adds an `OwnerReference` (pointing to the dummy object) to it before creating it in the upstream cluster.
    
    - When a tenant queries cluster-scoped resources through KubeZoo, the proxy removes the `OwnerReference` added by KubeZoo and returns it to the tenant.

    - When a `DELETE` event of a Tenant object is watched, if the Tenant has a `DeletionTimestamp` and a `kubezoo.io/tenant` finalizer, Tenant Controller will
        
        - Delete the dummy object in the upstream cluster, which triggers the Garbage Collector to clean up the tenant's cluster-scoped resources.
    
        - Remove the `kubezoo.io/tenant` finalizer from the Tenant object. 
          
        This alternative has some flaws:
      
        - More intrusive: We need to modify the Tenant Controller, protocol conversion rules, and the definition of Tenant object. Besides, we also need to create additional objects in the upstream cluster.
    
        - High risk of accidental deletion: All cluster-scoped resources of a tenant are associated with a single dummy object. If the dummy object is deleted by mistake, all the resources of the tenant will be deleted in cascade, which has a large impact.

## References

- [Garbage Collection](https://kubernetes.io/docs/concepts/architecture/garbage-collection/)

- [Use Cascading Deletion in a Cluster](https://kubernetes.io/docs/tasks/administer-cluster/use-cascading-deletion/)

- [Owners and Dependents](https://kubernetes.io/docs/concepts/overview/working-with-objects/owners-dependents/)

- [Finalizers](https://kubernetes.io/docs/concepts/overview/working-with-objects/finalizers/)

- [Using Finalizers to Control Deletion](https://kubernetes.io/blog/2021/05/14/using-finalizers-to-control-deletion/)

- [Kubernetes Garbage Collection](https://medium.com/@bharatnc/kubernetes-garbage-collection-781223f03c17)

- [OwnerReference Resource Field](https://github.com/kubernetes/enhancements/tree/master/keps/sig-api-machinery/2336-OwnerReference-resource-field)
