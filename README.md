# KubeZoo - Gateway Service for Kubernetes <br /> Multi-tenancy

English | [简体中文](./README.zh.md)


## Overview

KubeZoo is a lightweight gateway service that leverages the existing namespace model 
and add multi-tenancy capability to existing Kubernetes. KubeZoo provides 
view-level isolation among tenants by capturing and transforming the requests and responses.
Please refer to [design doc](./docs/design.md) for details.

<div align="center">
  <!--[if IE]>
    <img src="docs/img/kubezoo-overview.png" width=80% title="KubeZoo Overview" loading="eager" />
  <![endif]-->
  <picture>
    <source srcset="docs/img/kubezoo-overview-dark.png" width=80% title="KubeZoo Overview" media="(prefers-color-scheme: dark)">
    <img src="docs/img/kubezoo-overview.png" width=80% title="KubeZoo Overview" loading="eager" />
  </picture>
</div>

## Why KubeZoo

There exists [three common multi-tenancy](https://kubernetes.io/blog/2021/04/15/three-tenancy-models-for-kubernetes/) models for Kubernetes, 
i.e., Namespace as a Service (NaaS), Cluster as a Service (CaaS), Control Planes as a service (CPaaS). Each of them can be applied to address different 
use cases. However, our cases have some specific requirements and constraints that can not be met by the existing models, 
* ***Many Small Tenants*** - there usually exist hundreds of tenants who only need to run small batch workloads containing few pods for tens of minutes.
* ***Short Turnaround Time*** - users/tenants are usually impatient, who desire to have their service to be ready in minutes.
* ***Tight Manpower*** - managing thousands of clusters/control-planes can be labour-intensive and infeasible for medium-sized dev team.

To address these cases, we present a new tenancy model, i.e., **Kubernetes API as a Service (KAaaS)**, which provides competent isolation with 
negligible overheads and operation costs. KubeZoo implements this model with all tenants sharing both the control-plane and data-plane, which is 
suitable for the scenarios where thousands of small tenants need to share an underlying Kubernetes cluster.

<div align="center">
  <img src="docs/img/comparison.png" width=80% title="Comparison of Different Solutions">
</div>

For more details lease refer [FAQ](./docs/faq.md).

## Prerequisites

Please check the [resource and system requirements](./docs/resource-and-system-requirements.md) before installing KubeZoo.

## Getting started

KubeZoo supports Kubernetes versions up to 1.24. Using higher Kubernetes versions may cause
compatibility issues. KubeZoo can be installed using any of the following methods:

| Methods                     | Instruction                                | Estimated time |
| --------------------------- | ------------------------------------------ | -------------- |
| Deploy KubeZoo from scratch | [Deploy KubeZoo](./docs/manually-setup.md) | < 2 minutes    |

## Community

### Contributing

If you are willing to be a contributor for the KubeZoo project, please refer to our [CONTRIBUTING](CONTRIBUTING.md) document for details.
We have also prepared a developer [guide](./docs/developer-guide.md) to help the code contributors.

### Contact

If you have any questions or want to contribute, you are welcome to communicate most things via GitHub issues or pull requests.
Or Contact to [Maintainers](./MAINTAINERS.md)


## License

KubeZoo is under the Apache 2.0 license. See the [LICENSE](LICENSE) file for details.
Certain implementations in KubeZoo rely on the existing code from Kubernetes and the credits go to the original Kubernetes authors.
