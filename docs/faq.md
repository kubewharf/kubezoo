# FAQ

- Does kubezoo have any other restrictions except for not supporting daemonset resources?

> Kubezoo supports most of the resources such as pod, deployment and statefulset by default, while restrict the cluster sharing resources such as daemonset and node. The reason is that when multiple tenants share a cluster, no tenant is expected to sense and manipulate nodes(including daemonset) for security and isolation purpose.

- Does kubezoo support RBAC for tenants?

> Yes, kubezoo impersonates tenant identities through the impersonate mechanism, so the RBAC API is consistent with native API of kubernetes.

- Does CRD share across the tenants?

> KubeZoo divides CRD into two categories: one is tenant CRD which among tenants are completely isolated. The other is system CRD in a public cloud scenario, which will be handled by the same controller in the backend cluster. System CRD can be configured with a special policy to ensure that they are available to one or more tenants who can create objects.

- What if pods of different tenants are deployed on the same Node and their performance affects each other?

> In the public cloud scenario, we could implement the data plane through some services such as elastic container instance with higher isolation to ensure the complete isolation of computing, storage, and network resources.

- Does kubezoo need a dedicated kubectl?

> No, kubezoo supports the full kubernetes API, so each tenant could use kubectl exactly the same way as a single cluster. 

- What are the advantages and disadvantages of kubezoo and kubernetes's HNC?

> The HNC solution implements a hierarchical namespace structure that is still evolving and has not yet become the standard API of kubernetes yet. The advantage of kubezoo is that it provides the standard kubernetes API. In other words, if HNC will be supported by the standard kubernetes API in the future, then every tenant of kubezoo will be able to use HNC.

- What is the landing scene of kubezoo?

> From the perspective of private cloud, many small service resources have small demands. However, if a cluster is independently maintained for these small services, the operation and resource costs are high. Therefore, private cloud has a clear scenario. In the public cloud scenario, most tenants need only a lit resources, so the construction of serverless kubernetes based on kubezoo has the advantages of high efficiency and low cost.