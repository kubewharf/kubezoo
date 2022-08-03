# FAQ

- KubeZoo 除不支持 DaemonSet 资源外，还有其他的限制吗？

> KubeZoo 默认支持 Pod、Deployment、Statefulset 等绝大部分资源，但是限制对 Daemonset 和 Node 等集群共享资源的支持，原因是若多个租户共享一个集群时，出于安全和隔离的要求，不期望任何一个租户感知和操作节点(包括 Daemonset)。

- KubeZoo 支持租户的 RBAC 吗？

> 支持，KubeZoo 通过 impersonate 机制模拟租户身份，故 RBAC API 与原生集群是一致的。

- 不同租户创建的 CRD 能共用吗？

> KubeZoo 把 CRD 分为两类： 一种是租户级别的 CRD，各个租户之间的 CRD 是完全隔离的。另一种是在公有云场景下提供系统级别的 CRD，在后端集群上会由同一个 Controller 进行处理。系统级别的 CRD 可以配置为一种特殊的策略，保证它对于某一个或某一些租户是开放的，这些租户可以创建系统级别的 CRD 的对象。

- 不同租户的 Pod 部署到相同的 Node 上，性能互相影响怎么办？

> 在公有云场景下，Pod 可能会通过一些隔离性更高的服务，如弹性容器实例等进行数据面的实现，进而保证计算、存储和网络等资源的彻底隔离。

- KubeZoo 可以采用 kubectl 命令行吗，和原生是否有区别？

> 没有区别。KubeZoo 可以支持完整的 Kubernetes 的 API 视图，故每一个租户用 Kubectl 跟单集群的方式完全一样。唯一的不同是 KubeZoo 会为租户单独签发证书，发送 Kubeconfig，用户只需要指定正确的 Kubeconfig 即可。

- KubeZoo 和 Kubernetes 自己的多租户方案 HNC 比较有哪些优势和不足呢？

> HNC 方案实现了一种层级化的 namespace 的结构，目前还在演进当中的，尚未成为 Kubernetes 的标准 API。KubeZoo 的优势在于它可以提供标准的 Kubernetes API。换而言之，若以后标准 Kubernetes API 也支持 HNC，那 KubeZoo 的每个租户也能使用 HNC，相当于 KubeZoo 是 HNC 能力的一个超集。

- KubeZoo 的实际落地场景？

> 从私有云视角来看，很多小的业务资源量诉求小，但若为这些小业务各自独立维护一个集群，则运维和资源成本高，故在私有云具备明确的场景；在公有云场景下，绝大部分的租户资源体量小，基于 KubeZoo 构建 serverless K8S 具备高效、底成本的优势。