# KubeZoo 2.0 Conformance Test

## Run Conformance Test

Steps to run the conformance test.

```
$ export NUM_NODES=0
$ export KUBE_TEST_REPO_LIST={path of repo_list.yml}
$ export KUBECONFIG={your kubernetes config path}

$ ./e2e.test --provider=skeleton  --ginkgo.focus="\[Conformance\]" -num-nodes=0
```

Note, use a repo_list config file to accelerate the speed to download images within private registry.

## Report

Here is the conformance test report at 2022/05/17 with about 70% success ratio.

```
Summarizing 86 Failures:

[Fail] [sig-api-machinery] Garbage collector [It] should orphan pods created by rc if delete options say so [Conformance]
/Users/wsfdl/go/src/k8s.io/kubernetes/_output/local/go/src/k8s.io/kubernetes/test/e2e/apimachinery/garbage_collector.go:56

[Fail] [sig-api-machinery] Watchers [It] should observe add, update, and delete watch notifications on configmaps [Conformance]
/Users/wsfdl/go/src/k8s.io/kubernetes/_output/local/go/src/k8s.io/kubernetes/test/e2e/apimachinery/watch.go:418

[Fail] [sig-network] Networking [BeforeEach] Granular Checks: Pods should function for intra-pod communication: udp [NodeConformance] [Conformance]
/Users/wsfdl/go/src/k8s.io/kubernetes/_output/local/go/src/k8s.io/kubernetes/test/e2e/framework/framework.go:240

[Fail] [sig-node] Downward API [It] should provide pod name, namespace and IP address as env vars [NodeConformance] [Conformance]
/Users/wsfdl/go/src/k8s.io/kubernetes/_output/local/go/src/k8s.io/kubernetes/test/e2e/framework/util.go:782

[Fail] [sig-api-machinery] AdmissionWebhook [Privileged:ClusterAdmin] [It] should be able to deny attaching pod [Conformance]
/Users/wsfdl/go/src/k8s.io/kubernetes/_output/local/go/src/k8s.io/kubernetes/test/e2e/apimachinery/webhook.go:961

[Fail] [sig-network] Services [It] should be able to change the type from NodePort to ExternalName [Conformance]
/Users/wsfdl/go/src/k8s.io/kubernetes/_output/local/go/src/k8s.io/kubernetes/test/e2e/network/service.go:1735

[Fail] [sig-network] DNS [It] should provide DNS for the cluster  [Conformance]
/Users/wsfdl/go/src/k8s.io/kubernetes/_output/local/go/src/k8s.io/kubernetes/test/e2e/network/dns_common.go:538

[Fail] [k8s.io] [sig-node] NoExecuteTaintManager Multiple Pods [Serial] [It] evicts pods with minTolerationSeconds [Disruptive] [Conformance]
/Users/wsfdl/go/src/k8s.io/kubernetes/_output/local/go/src/k8s.io/kubernetes/test/e2e/node/taints.go:431

[Fail] [sig-api-machinery] CustomResourceConversionWebhook [Privileged:ClusterAdmin] [It] should be able to convert from CR v1 to CR v2 [Conformance]
/Users/wsfdl/go/src/k8s.io/kubernetes/_output/local/go/src/k8s.io/kubernetes/test/e2e/apimachinery/crd_conversion_webhook.go:496

[Fail] [sig-api-machinery] AdmissionWebhook [Privileged:ClusterAdmin] [It] should mutate custom resource with different stored version [Conformance]
/Users/wsfdl/go/src/k8s.io/kubernetes/_output/local/go/src/k8s.io/kubernetes/test/e2e/apimachinery/webhook.go:1826

[Fail] [sig-apps] ReplicationController [It] should release no longer matching pods [Conformance]
/Users/wsfdl/go/src/k8s.io/kubernetes/_output/local/go/src/k8s.io/kubernetes/test/e2e/apps/rc.go:350

[Fail] [k8s.io] [sig-node] NoExecuteTaintManager Single Pod [Serial] [It] removing taint cancels eviction [Disruptive] [Conformance]
/Users/wsfdl/go/src/k8s.io/kubernetes/_output/local/go/src/k8s.io/kubernetes/test/e2e/framework/util.go:1074

[Fail] [sig-api-machinery] AdmissionWebhook [Privileged:ClusterAdmin] [It] should deny crd creation [Conformance]
/Users/wsfdl/go/src/k8s.io/kubernetes/_output/local/go/src/k8s.io/kubernetes/test/e2e/apimachinery/webhook.go:2059

[Fail] [sig-apps] ReplicaSet [It] should adopt matching pods on creation and release no longer matching pods [Conformance]
/Users/wsfdl/go/src/k8s.io/kubernetes/_output/local/go/src/k8s.io/kubernetes/test/e2e/apps/replica_set.go:343

[Fail] [sig-network] Services [It] should be able to change the type from ClusterIP to ExternalName [Conformance]
/Users/wsfdl/go/src/k8s.io/kubernetes/_output/local/go/src/k8s.io/kubernetes/test/e2e/network/service.go:1716

[Fail] [sig-apps] Daemon set [Serial] [BeforeEach] should retry creating failed daemon pods [Conformance]
/Users/wsfdl/go/src/k8s.io/kubernetes/_output/local/go/src/k8s.io/kubernetes/test/e2e/apps/daemon_set.go:145

[Fail] [sig-network] Networking Granular Checks: Pods [It] should function for intra-pod communication: http [NodeConformance] [Conformance]
/Users/wsfdl/go/src/k8s.io/kubernetes/_output/local/go/src/k8s.io/kubernetes/test/e2e/framework/network/utils.go:652

[Fail] [sig-api-machinery] AdmissionWebhook [Privileged:ClusterAdmin] [It] should be able to deny pod and configmap creation [Conformance]
/Users/wsfdl/go/src/k8s.io/kubernetes/_output/local/go/src/k8s.io/kubernetes/test/e2e/apimachinery/webhook.go:909

[Fail] [sig-scheduling] SchedulerPredicates [Serial] [BeforeEach] validates that there is no conflict between pods with same hostPort but different hostIP and protocol [Conformance]
/Users/wsfdl/go/src/k8s.io/kubernetes/_output/local/go/src/k8s.io/kubernetes/test/e2e/scheduling/predicates.go:109

[Fail] [sig-cli] Kubectl client Update Demo [It] should create and stop a replication controller  [Conformance]
/Users/wsfdl/go/src/k8s.io/kubernetes/_output/local/go/src/k8s.io/kubernetes/test/e2e/kubectl/kubectl.go:2118

[Fail] [sig-api-machinery] Garbage collector [It] should not delete dependents that have both valid owner and owner that's waiting for dependents to be deleted [Conformance]
/Users/wsfdl/go/src/k8s.io/kubernetes/_output/local/go/src/k8s.io/kubernetes/test/e2e/apimachinery/garbage_collector.go:56

[Fail] [sig-api-machinery] CustomResourcePublishOpenAPI [Privileged:ClusterAdmin] [It] updates the published spec when one version gets renamed [Conformance]
/Users/wsfdl/go/src/k8s.io/kubernetes/_output/local/go/src/k8s.io/kubernetes/test/e2e/apimachinery/crd_publish_openapi.go:384

[Fail] [sig-network] Proxy version v1 [It] should proxy through a service and a pod  [Conformance]
/Users/wsfdl/go/src/k8s.io/kubernetes/_output/local/go/src/k8s.io/kubernetes/test/e2e/network/proxy.go:260

[Fail] [sig-network] DNS [It] should provide DNS for pods for Hostname [LinuxOnly] [Conformance]
/Users/wsfdl/go/src/k8s.io/kubernetes/_output/local/go/src/k8s.io/kubernetes/test/e2e/network/dns_common.go:538

[Panic!] [sig-api-machinery] CustomResourceDefinition resources [Privileged:ClusterAdmin] Simple CustomResourceDefinition [It] getting/updating/patching custom resource definition status sub-resource works  [Conformance]
/Users/wsfdl/go/src/k8s.io/kubernetes/_output/local/go/src/k8s.io/kubernetes/vendor/k8s.io/apimachinery/pkg/util/runtime/runtime.go:55

[Fail] [sig-api-machinery] AdmissionWebhook [Privileged:ClusterAdmin] [It] should be able to deny custom resource creation, update and deletion [Conformance]
/Users/wsfdl/go/src/k8s.io/kubernetes/_output/local/go/src/k8s.io/kubernetes/test/e2e/apimachinery/webhook.go:1749

[Fail] [sig-storage] EmptyDir wrapper volumes [It] should not cause race condition when used for configmaps [Serial] [Conformance]
/Users/wsfdl/go/src/k8s.io/kubernetes/_output/local/go/src/k8s.io/kubernetes/test/e2e/storage/empty_dir_wrapper.go:353

[Fail] [sig-api-machinery] Aggregator [It] Should be able to support the 1.17 Sample API Server using the current Aggregator [Conformance]
/Users/wsfdl/go/src/k8s.io/kubernetes/_output/local/go/src/k8s.io/kubernetes/test/e2e/apimachinery/aggregator.go:401

[Fail] [sig-network] Services [It] should be able to create a functioning NodePort service [Conformance]
/Users/wsfdl/go/src/k8s.io/kubernetes/_output/local/go/src/k8s.io/kubernetes/test/e2e/network/service.go:1158

[Fail] [sig-api-machinery] AdmissionWebhook [Privileged:ClusterAdmin] [It] should mutate pod and apply defaults after mutation [Conformance]
/Users/wsfdl/go/src/k8s.io/kubernetes/_output/local/go/src/k8s.io/kubernetes/test/e2e/apimachinery/webhook.go:1055

[Fail] [sig-api-machinery] CustomResourcePublishOpenAPI [Privileged:ClusterAdmin] [It] works for multiple CRDs of different groups [Conformance]
/Users/wsfdl/go/src/k8s.io/kubernetes/_output/local/go/src/k8s.io/kubernetes/test/e2e/apimachinery/crd_publish_openapi.go:269

[Fail] [sig-apps] Daemon set [Serial] [BeforeEach] should run and stop simple daemon [Conformance]
/Users/wsfdl/go/src/k8s.io/kubernetes/_output/local/go/src/k8s.io/kubernetes/test/e2e/apps/daemon_set.go:145

[Fail] [k8s.io] InitContainer [NodeConformance] [It] should not start app containers if init containers fail on a RestartAlways pod [Conformance]
/Users/wsfdl/go/src/k8s.io/kubernetes/_output/local/go/src/k8s.io/kubernetes/test/e2e/common/init_container.go:399

[Fail] [sig-apps] ReplicationController [It] should serve a basic image on each replica with a public image  [Conformance]
/Users/wsfdl/go/src/k8s.io/kubernetes/_output/local/go/src/k8s.io/kubernetes/test/e2e/apps/rc.go:171

[Fail] [sig-apps] StatefulSet [k8s.io] Basic StatefulSet functionality [StatefulSetBasic] [It] Should recreate evicted statefulset [Conformance]
/Users/wsfdl/go/src/k8s.io/kubernetes/_output/local/go/src/k8s.io/kubernetes/test/e2e/apps/statefulset.go:704

[Fail] [sig-api-machinery] CustomResourceConversionWebhook [Privileged:ClusterAdmin] [It] should be able to convert a non homogeneous list of CRs [Conformance]
/Users/wsfdl/go/src/k8s.io/kubernetes/_output/local/go/src/k8s.io/kubernetes/test/e2e/apimachinery/crd_conversion_webhook.go:496

[Fail] [sig-api-machinery] AdmissionWebhook [Privileged:ClusterAdmin] [It] listing validating webhooks should work [Conformance]
/Users/wsfdl/go/src/k8s.io/kubernetes/_output/local/go/src/k8s.io/kubernetes/test/e2e/apimachinery/webhook.go:606

[Fail] [sig-network] Proxy version v1 [It] should proxy logs on node using proxy subresource  [Conformance]
/Users/wsfdl/go/src/k8s.io/kubernetes/_output/local/go/src/k8s.io/kubernetes/test/e2e/network/proxy.go:296

[Fail] [sig-scheduling] SchedulerPredicates [Serial] [BeforeEach] validates that there exists conflict between pods with same hostPort and protocol but one using 0.0.0.0 hostIP [Conformance]
/Users/wsfdl/go/src/k8s.io/kubernetes/_output/local/go/src/k8s.io/kubernetes/test/e2e/scheduling/predicates.go:109

[Fail] [sig-api-machinery] CustomResourcePublishOpenAPI [Privileged:ClusterAdmin] [It] works for CRD preserving unknown fields at the schema root [Conformance]
/Users/wsfdl/go/src/k8s.io/kubernetes/_output/local/go/src/k8s.io/kubernetes/test/e2e/apimachinery/crd_publish_openapi.go:186

[Fail] [sig-network] DNS [It] should provide DNS for pods for Subdomain [Conformance]
/Users/wsfdl/go/src/k8s.io/kubernetes/_output/local/go/src/k8s.io/kubernetes/test/e2e/network/dns_common.go:538

[Panic!] [sig-cli] Kubectl client Kubectl describe [It] should check if kubectl describe prints relevant information for rc and pods  [Conformance]
/usr/local/Cellar/go/1.15.2/libexec/src/runtime/panic.go:88

[Fail] [sig-scheduling] SchedulerPredicates [Serial] [BeforeEach] validates that NodeSelector is respected if matching  [Conformance]
/Users/wsfdl/go/src/k8s.io/kubernetes/_output/local/go/src/k8s.io/kubernetes/test/e2e/scheduling/predicates.go:109

[Fail] [sig-network] Service endpoints latency [It] should not be very high  [Conformance]
/Users/wsfdl/go/src/k8s.io/kubernetes/_output/local/go/src/k8s.io/kubernetes/test/e2e/network/service_latency.go:129

[Fail] [sig-network] DNS [It] should resolve DNS of partial qualified names for services [LinuxOnly] [Conformance]
/Users/wsfdl/go/src/k8s.io/kubernetes/_output/local/go/src/k8s.io/kubernetes/test/e2e/network/dns_common.go:538

[Fail] [sig-api-machinery] AdmissionWebhook [Privileged:ClusterAdmin] [It] should mutate configmap [Conformance]
/Users/wsfdl/go/src/k8s.io/kubernetes/_output/local/go/src/k8s.io/kubernetes/test/e2e/apimachinery/webhook.go:988

[Fail] [sig-api-machinery] CustomResourcePublishOpenAPI [Privileged:ClusterAdmin] [It] works for multiple CRDs of same group and version but different kinds [Conformance]
/Users/wsfdl/go/src/k8s.io/kubernetes/_output/local/go/src/k8s.io/kubernetes/test/e2e/apimachinery/crd_publish_openapi.go:350

[Fail] [sig-api-machinery] AdmissionWebhook [Privileged:ClusterAdmin] [It] should unconditionally reject operations on fail closed webhook [Conformance]
/Users/wsfdl/go/src/k8s.io/kubernetes/_output/local/go/src/k8s.io/kubernetes/test/e2e/apimachinery/webhook.go:1275

[Fail] [sig-apps] Deployment [It] deployment should support proportional scaling [Conformance]
/Users/wsfdl/go/src/k8s.io/kubernetes/_output/local/go/src/k8s.io/kubernetes/test/e2e/apps/deployment.go:721

[Fail] [sig-network] Proxy version v1 [It] should proxy logs on node with explicit kubelet port using proxy subresource  [Conformance]
/Users/wsfdl/go/src/k8s.io/kubernetes/_output/local/go/src/k8s.io/kubernetes/test/e2e/network/proxy.go:296

[Fail] [sig-network] DNS [It] should provide DNS for services  [Conformance]
/Users/wsfdl/go/src/k8s.io/kubernetes/_output/local/go/src/k8s.io/kubernetes/test/e2e/network/dns_common.go:538

[Fail] [k8s.io] Pods [It] should be submitted and removed [NodeConformance] [Conformance]
/Users/wsfdl/go/src/k8s.io/kubernetes/_output/local/go/src/k8s.io/kubernetes/test/e2e/common/pods.go:322

[Fail] [sig-apps] Daemon set [Serial] [BeforeEach] should run and stop complex daemon [Conformance]
/Users/wsfdl/go/src/k8s.io/kubernetes/_output/local/go/src/k8s.io/kubernetes/test/e2e/apps/daemon_set.go:145

[Fail] [sig-network] DNS [It] should provide DNS for ExternalName services [Conformance]
/Users/wsfdl/go/src/k8s.io/kubernetes/_output/local/go/src/k8s.io/kubernetes/test/e2e/network/dns_common.go:538

[Fail] [sig-cli] Kubectl client Guestbook application [It] should create and stop a working application  [Conformance]
/Users/wsfdl/go/src/k8s.io/kubernetes/_output/local/go/src/k8s.io/kubernetes/test/e2e/kubectl/kubectl.go:1858

[Fail] [sig-apps] Daemon set [Serial] [BeforeEach] should update pod when spec was updated and update strategy is RollingUpdate [Conformance]
/Users/wsfdl/go/src/k8s.io/kubernetes/_output/local/go/src/k8s.io/kubernetes/test/e2e/apps/daemon_set.go:145

[Fail] [sig-api-machinery] AdmissionWebhook [Privileged:ClusterAdmin] [It] should mutate custom resource [Conformance]
/Users/wsfdl/go/src/k8s.io/kubernetes/_output/local/go/src/k8s.io/kubernetes/test/e2e/apimachinery/webhook.go:1826

[Fail] [sig-api-machinery] CustomResourceDefinition Watch [Privileged:ClusterAdmin] CustomResourceDefinition Watch [It] watch on custom resource definition objects [Conformance]
/Users/wsfdl/go/src/k8s.io/kubernetes/_output/local/go/src/k8s.io/kubernetes/test/e2e/apimachinery/crd_watch.go:69

[Fail] [sig-api-machinery] AdmissionWebhook [Privileged:ClusterAdmin] [It] should not be able to mutate or prevent deletion of webhook configuration objects [Conformance]
/Users/wsfdl/go/src/k8s.io/kubernetes/_output/local/go/src/k8s.io/kubernetes/test/e2e/apimachinery/webhook.go:1361

[Fail] [sig-api-machinery] CustomResourcePublishOpenAPI [Privileged:ClusterAdmin] [It] works for CRD preserving unknown fields in an embedded object [Conformance]
/Users/wsfdl/go/src/k8s.io/kubernetes/_output/local/go/src/k8s.io/kubernetes/test/e2e/apimachinery/crd_publish_openapi.go:228

[Fail] [sig-scheduling] SchedulerPredicates [Serial] [BeforeEach] validates resource limits of pods that are allowed to run  [Conformance]
/Users/wsfdl/go/src/k8s.io/kubernetes/_output/local/go/src/k8s.io/kubernetes/test/e2e/scheduling/predicates.go:109

[Fail] [sig-api-machinery] AdmissionWebhook [Privileged:ClusterAdmin] [It] should mutate custom resource with pruning [Conformance]
/Users/wsfdl/go/src/k8s.io/kubernetes/_output/local/go/src/k8s.io/kubernetes/test/e2e/apimachinery/webhook.go:1826

[Fail] [sig-network] DNS [It] should provide /etc/hosts entries for the cluster [LinuxOnly] [Conformance]
/Users/wsfdl/go/src/k8s.io/kubernetes/_output/local/go/src/k8s.io/kubernetes/test/e2e/network/dns_common.go:538

[Fail] [sig-api-machinery] AdmissionWebhook [Privileged:ClusterAdmin] [It] listing mutating webhooks should work [Conformance]
/Users/wsfdl/go/src/k8s.io/kubernetes/_output/local/go/src/k8s.io/kubernetes/test/e2e/apimachinery/webhook.go:680

[Fail] [sig-api-machinery] Watchers [It] should observe an object deletion if it stops meeting the requirements of the selector [Conformance]
/Users/wsfdl/go/src/k8s.io/kubernetes/_output/local/go/src/k8s.io/kubernetes/test/e2e/apimachinery/watch.go:284

[Fail] [sig-network] Services [It] should be able to change the type from ExternalName to NodePort [Conformance]
/Users/wsfdl/go/src/k8s.io/kubernetes/_output/local/go/src/k8s.io/kubernetes/test/e2e/network/service.go:1670

[Fail] [sig-api-machinery] CustomResourcePublishOpenAPI [Privileged:ClusterAdmin] [It] removes definition from spec when one version gets changed to not be served [Conformance]
/Users/wsfdl/go/src/k8s.io/kubernetes/_output/local/go/src/k8s.io/kubernetes/test/e2e/apimachinery/crd_publish_openapi.go:435

[Fail] [k8s.io] [sig-node] PreStop [It] should call prestop when killing a pod  [Conformance]
/Users/wsfdl/go/src/k8s.io/kubernetes/_output/local/go/src/k8s.io/kubernetes/test/e2e/node/pre_stop.go:165

[Fail] [sig-network] Networking Granular Checks: Pods [It] should function for node-pod communication: http [LinuxOnly] [NodeConformance] [Conformance]
/Users/wsfdl/go/src/k8s.io/kubernetes/_output/local/go/src/k8s.io/kubernetes/test/e2e/framework/network/utils.go:652

[Fail] [sig-apps] Daemon set [Serial] [BeforeEach] should rollback without unnecessary restarts [Conformance]
/Users/wsfdl/go/src/k8s.io/kubernetes/_output/local/go/src/k8s.io/kubernetes/test/e2e/apps/daemon_set.go:145

[Fail] [sig-api-machinery] CustomResourcePublishOpenAPI [Privileged:ClusterAdmin] [It] works for multiple CRDs of same group but different versions [Conformance]
/Users/wsfdl/go/src/k8s.io/kubernetes/_output/local/go/src/k8s.io/kubernetes/test/e2e/apimachinery/crd_publish_openapi.go:302

[Fail] [sig-network] Networking Granular Checks: Pods [It] should function for node-pod communication: udp [LinuxOnly] [NodeConformance] [Conformance]
/Users/wsfdl/go/src/k8s.io/kubernetes/_output/local/go/src/k8s.io/kubernetes/test/e2e/framework/network/utils.go:652

[Fail] [sig-apps] StatefulSet [k8s.io] Basic StatefulSet functionality [StatefulSetBasic] [It] should have a working scale subresource [Conformance]
/Users/wsfdl/go/src/k8s.io/kubernetes/_output/local/go/src/k8s.io/kubernetes/test/e2e/apps/statefulset.go:806

[Fail] [sig-api-machinery] AdmissionWebhook [Privileged:ClusterAdmin] [It] patching/updating a mutating webhook should work [Conformance]
/Users/wsfdl/go/src/k8s.io/kubernetes/_output/local/go/src/k8s.io/kubernetes/test/e2e/apimachinery/webhook.go:527

[Fail] [sig-auth] ServiceAccounts [It] should mount an API token into pods  [Conformance]
/Users/wsfdl/go/src/k8s.io/kubernetes/_output/local/go/src/k8s.io/kubernetes/test/e2e/auth/service_accounts.go:240

[Fail] [sig-api-machinery] CustomResourcePublishOpenAPI [Privileged:ClusterAdmin] [It] works for CRD with validation schema [Conformance]
/Users/wsfdl/go/src/k8s.io/kubernetes/_output/local/go/src/k8s.io/kubernetes/test/e2e/apimachinery/crd_publish_openapi.go:68

[Fail] [sig-api-machinery] AdmissionWebhook [Privileged:ClusterAdmin] [It] patching/updating a validating webhook should work [Conformance]
/Users/wsfdl/go/src/k8s.io/kubernetes/_output/local/go/src/k8s.io/kubernetes/test/e2e/apimachinery/webhook.go:432

[Fail] [sig-apps] ReplicaSet [It] should serve a basic image on each replica with a public image  [Conformance]
/Users/wsfdl/go/src/k8s.io/kubernetes/_output/local/go/src/k8s.io/kubernetes/test/e2e/apps/replica_set.go:173

[Fail] [sig-api-machinery] Watchers [It] should receive events on concurrent watches in same order [Conformance]
/Users/wsfdl/go/src/k8s.io/kubernetes/_output/local/go/src/k8s.io/kubernetes/test/e2e/apimachinery/watch.go:439

[Fail] [sig-scheduling] SchedulerPredicates [Serial] [BeforeEach] validates that NodeSelector is respected if not matching  [Conformance]
/Users/wsfdl/go/src/k8s.io/kubernetes/_output/local/go/src/k8s.io/kubernetes/test/e2e/scheduling/predicates.go:109

[Fail] [sig-api-machinery] Garbage collector [It] should keep the rc around until all its pods are deleted if the deleteOptions says so [Conformance]
/Users/wsfdl/go/src/k8s.io/kubernetes/_output/local/go/src/k8s.io/kubernetes/test/e2e/apimachinery/garbage_collector.go:56

[Fail] [sig-api-machinery] CustomResourcePublishOpenAPI [Privileged:ClusterAdmin] [It] works for CRD without validation schema [Conformance]
/Users/wsfdl/go/src/k8s.io/kubernetes/_output/local/go/src/k8s.io/kubernetes/test/e2e/apimachinery/crd_publish_openapi.go:145

[Fail] [sig-api-machinery] Servers with support for Table transformation [BeforeEach] should return a 406 for a backend which does not implement metadata [Conformance]
/Users/wsfdl/go/src/k8s.io/kubernetes/_output/local/go/src/k8s.io/kubernetes/test/e2e/framework/skipper/skipper.go:226

[Fail] [sig-api-machinery] AdmissionWebhook [Privileged:ClusterAdmin] [It] should honor timeout [Conformance]
/Users/wsfdl/go/src/k8s.io/kubernetes/_output/local/go/src/k8s.io/kubernetes/test/e2e/apimachinery/webhook.go:2188

[Fail] [sig-cli] Kubectl client Update Demo [It] should scale a replication controller  [Conformance]
/Users/wsfdl/go/src/k8s.io/kubernetes/_output/local/go/src/k8s.io/kubernetes/test/e2e/kubectl/kubectl.go:2118

[Fail] [sig-api-machinery] CustomResourceDefinition resources [Privileged:ClusterAdmin] [It] custom resource defaulting for requests and from storage works  [Conformance]
/Users/wsfdl/go/src/k8s.io/kubernetes/_output/local/go/src/k8s.io/kubernetes/test/e2e/apimachinery/custom_resource_definition.go:304

Ran 277 of 4993 Specs in 10610.033 seconds
FAIL! -- 191 Passed | 86 Failed | 0 Pending | 4716 Skipped
```
