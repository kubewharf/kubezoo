/*
Copyright 2022 The KubeZoo Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"fmt"
	"strconv"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	typedappsv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	typedrbacv1 "k8s.io/client-go/kubernetes/typed/rbac/v1"
	"k8s.io/klog"
	"k8s.io/utils/pointer"

	tenantv1a1 "github.com/kubewharf/kubezoo/pkg/apis/tenant/v1alpha1"
	"github.com/kubewharf/kubezoo/pkg/util"
)

const LastSyncGeneration = "kubezoo.io/last-sync-generation"

// syncTenantStack syncs the tenant service stack, including coredns.
func syncCoredns(
	ctx context.Context,
	coreClient typedcorev1.CoreV1Interface,
	appsClient typedappsv1.AppsV1Interface,
	rbacClient typedrbacv1.RbacV1Interface,
	tenant *tenantv1a1.Tenant,
	zooHost string,
	zooPort int,
) error {
	if err := syncCorednsConfigMap(ctx, coreClient, tenant); err != nil {
		return fmt.Errorf("sync coredns configmap: %w", err)
	}

	if err := syncCorednsClusterRole(ctx, rbacClient, tenant); err != nil {
		return fmt.Errorf("sync coredns clusterrole:")
	}

	if err := syncCorednsServiceAccount(ctx, coreClient, tenant); err != nil {
		return fmt.Errorf("sync coredns configmap: %w", err)
	}

	if err := syncCorednsDeploy(ctx, appsClient, tenant, zooHost, zooPort); err != nil {
		return fmt.Errorf("sync coredns deployment: %w", err)
	}

	return nil
}

func syncCorednsConfigMap(ctx context.Context, coreClient typedcorev1.CoreV1Interface, tenant *tenantv1a1.Tenant) error {
	cmClient := coreClient.ConfigMaps(util.AddTenantIDPrefix(tenant.Name, corednsNamespace(tenant)))
	cmName := corednsConfigMapName(tenant)
	cm, err := cmClient.Get(ctx, cmName, metav1.GetOptions{ResourceVersion: "0"})
	if err != nil {
		if !errors.IsNotFound(err) {
			return fmt.Errorf("get current configmap: %w", err)
		}

		cm = renderCorednsConfigMap(tenant)
		cm, err = cmClient.Create(ctx, cm, metav1.CreateOptions{FieldValidation: "Strict"})
		if err != nil {
			return fmt.Errorf("create coredns configmap: %w", err)
		}
	} else {
		if !cm.DeletionTimestamp.IsZero() {
			klog.Warningf("coredns configmap for tenant %q is terminating, but kubezoo will recreate it", tenant.Name)
			return nil
		}

		generationString := cm.Annotations[LastSyncGeneration]

		var needUpdate bool
		if lastSyncGeneration, err := strconv.ParseInt(generationString, 10, 64); err != nil {
			klog.Warningf("coredns deployment for tenant %q contains an invalid last-sync-generation, force overwriting")
			needUpdate = true
		} else {
			needUpdate = lastSyncGeneration != tenant.Generation
		}

		if needUpdate {
			oldResourceVersion := cm.ResourceVersion
			cm = renderCorednsConfigMap(tenant)
			cm.ResourceVersion = oldResourceVersion

			cm, err = cmClient.Update(ctx, cm, metav1.UpdateOptions{FieldValidation: "Strict"})
			if err != nil {
				return fmt.Errorf("update coredns serviceaccount: %w", err)
			}
		}
	}

	return nil
}

func syncCorednsServiceAccount(ctx context.Context, coreClient typedcorev1.CoreV1Interface, tenant *tenantv1a1.Tenant) error {
	saClient := coreClient.ServiceAccounts(util.AddTenantIDPrefix(tenant.Name, corednsNamespace(tenant)))
	saName := corednsServiceAccountName(tenant)
	sa, err := saClient.Get(ctx, saName, metav1.GetOptions{ResourceVersion: "0"})
	if err != nil {
		if !errors.IsNotFound(err) {
			return fmt.Errorf("get current serviceaccount: %w", err)
		}

		sa = renderCorednsServiceAccount(tenant)
		sa, err = saClient.Create(ctx, sa, metav1.CreateOptions{FieldValidation: "Strict"})
		if err != nil {
			return fmt.Errorf("create coredns serviceaccount: %w", err)
		}
	} else {
		if !sa.DeletionTimestamp.IsZero() {
			klog.Warningf("coredns serviceaccount for tenant %q is terminating, but kubezoo will recreate it", tenant.Name)
			return nil
		}

		generationString := sa.Annotations[LastSyncGeneration]

		var needUpdate bool
		if lastSyncGeneration, err := strconv.ParseInt(generationString, 10, 64); err != nil {
			klog.Warningf("coredns deployment for tenant %q contains an invalid last-sync-generation, force overwriting")
			needUpdate = true
		} else {
			needUpdate = lastSyncGeneration != tenant.Generation
		}

		if needUpdate {
			oldResourceVersion := sa.ResourceVersion
			sa = renderCorednsServiceAccount(tenant)
			sa.ResourceVersion = oldResourceVersion

			sa, err = saClient.Update(ctx, sa, metav1.UpdateOptions{FieldValidation: "Strict"})
			if err != nil {
				return fmt.Errorf("update coredns serviceaccount: %w", err)
			}
		}
	}

	return nil
}

func syncCorednsDeploy(
	ctx context.Context,
	appsClient typedappsv1.AppsV1Interface,
	tenant *tenantv1a1.Tenant,
	zooHost string,
	zooPort int,
) error {
	deployClient := appsClient.Deployments(util.AddTenantIDPrefix(tenant.Name, corednsNamespace(tenant)))
	deployName := corednsDeployName(tenant)
	deploy, err := deployClient.Get(ctx, deployName, metav1.GetOptions{ResourceVersion: "0"})
	if err != nil {
		if !errors.IsNotFound(err) {
			return fmt.Errorf("get current deploy: %w", err)
		}

		deploy = renderCorednsDeploy(tenant, zooHost, zooPort)
		deploy, err = deployClient.Create(ctx, deploy, metav1.CreateOptions{FieldValidation: "Strict"})
		if err != nil {
			return fmt.Errorf("create coredns deployment: %w", err)
		}
	} else {
		if !deploy.DeletionTimestamp.IsZero() {
			klog.Warningf("coredns deployment for tenant %q is terminating, but kubezoo will recreate it", tenant.Name)
			return nil
		}

		generationString := deploy.Annotations[LastSyncGeneration]

		var needUpdate bool
		if lastSyncGeneration, err := strconv.ParseInt(generationString, 10, 64); err != nil {
			klog.Warningf("coredns deployment for tenant %q contains an invalid last-sync-generation, force overwriting")
			needUpdate = true
		} else {
			needUpdate = lastSyncGeneration != tenant.Generation
		}

		if needUpdate {
			oldResourceVersion := deploy.ResourceVersion
			deploy = renderCorednsDeploy(tenant, zooHost, zooPort)
			deploy.ResourceVersion = oldResourceVersion

			deploy, err = deployClient.Update(ctx, deploy, metav1.UpdateOptions{FieldValidation: "Strict"})
			if err != nil {
				return fmt.Errorf("update coredns deployment: %w", err)
			}
		}
	}

	return nil
}

func renderCorednsConfigMap(tenant *tenantv1a1.Tenant) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: util.AddTenantIDPrefix(tenant.Name, corednsNamespace(tenant)),
			Name:      corednsConfigMapName(tenant),
			Labels:    corednsLabels(),
			Annotations: map[string]string{
				LastSyncGeneration: fmt.Sprint(tenant.Generation),
			},
		},
		Data: map[string]string{
			"Corefile": renderCorefile(tenant),
		},
	}
}

func renderCorefile(tenant *tenantv1a1.Tenant) string {
	return fmt.Sprintf(".:53 {\n" +
		"errors\n" +
		"health {\n" +
		"  lameduck 5s\n" +
		"}\n" +
		"ready\n" +
		"kubernetes cluster.local in-addr.arpa ip6.arpa {\n" +
		"  pods insecure\n" +
		"  fallthrough in-addr.arpa ip6.arpa\n" +
		"  ttl 30\n" +
		"}\n" +
		"prometheus :9153\n" +
		"forward . /etc/resolv.conf {\n" +
		"  max_concurrent 1000\n" +
		"}\n" +
		"cache 30\n" +
		"loop\n" +
		"reload\n" +
		"loadbalance\n" +
		"}")
}

func renderCorednsServiceAccount(tenant *tenantv1a1.Tenant) *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: util.AddTenantIDPrefix(tenant.Name, corednsNamespace(tenant)),
			Name:      corednsServiceAccountName(tenant),
			Labels:    corednsLabels(),
			Annotations: map[string]string{
				LastSyncGeneration: fmt.Sprint(tenant.Generation),
			},
		},
	}
}

func renderCorednsDeploy(tenant *tenantv1a1.Tenant, zooHost string, zooPort int) *appsv1.Deployment {
	const corefileVolumeName = "corefile"

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: util.AddTenantIDPrefix(tenant.Name, corednsNamespace(tenant)),
			Name:      corednsDeployName(tenant),
			Labels:    corednsLabels(),
			Annotations: map[string]string{
				LastSyncGeneration: fmt.Sprint(tenant.Generation),
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: pointer.Int32(corednsReplicas(tenant)),
			Selector: corednsLabelSelector(),
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: corednsLabels(),
				},
				Spec: corev1.PodSpec{
					Volumes: []corev1.Volume{
						{
							Name: corefileVolumeName,
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: corednsConfigMapName(tenant),
									},
									DefaultMode: pointer.Int32(420),
									Items: []corev1.KeyToPath{
										{Key: "Corefile", Path: "Corefile"},
									},
								},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name:            "coredns",
							Image:           corednsImage(),
							ImagePullPolicy: corev1.PullIfNotPresent,
							Args:            []string{"-conf", "/etc/coredns/Corefile"},
							Env: []corev1.EnvVar{
								{Name: "KUBERNETES_SERVICE_HOST", Value: zooHost},
								{Name: "KUBERNETES_SERVICE_PORT", Value: zooPort},
							},
							Ports: []corev1.ContainerPort{
								{ContainerPort: 53, Name: "dns", Protocol: corev1.ProtocolUDP},
								{ContainerPort: 53, Name: "dns-tcp", Protocol: corev1.ProtocolTCP},
								{ContainerPort: 9153, Name: "metrics", Protocol: corev1.ProtocolTCP},
							},
							Resources: corednsResourceRequirements(tenant),
							VolumeMounts: []corev1.VolumeMount{
								{MountPath: "/etc/coredns", Name: corefileVolumeName, ReadOnly: true},
							},
							LivenessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path:   "/health",
										Port:   intstr.FromInt(8080),
										Scheme: corev1.URISchemeHTTP,
									},
								},
								InitialDelaySeconds: 60,
								TimeoutSeconds:      5,
								PeriodSeconds:       10,
								SuccessThreshold:    1,
								FailureThreshold:    5,
							},
							ReadinessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path:   "/ready",
										Port:   intstr.FromInt(8181),
										Scheme: corev1.URISchemeHTTP,
									},
								},
								TimeoutSeconds:   1,
								PeriodSeconds:    10,
								SuccessThreshold: 1,
								FailureThreshold: 3,
							},
							SecurityContext: &corev1.SecurityContext{
								Capabilities: &corev1.Capabilities{
									Add:  []corev1.Capability{"NET_BIND_SERVICE"},
									Drop: []corev1.Capability{"all"},
								},
							},
						},
					},
					DNSPolicy:          corev1.DNSDefault, // do not use cluster DNS
					ServiceAccountName: corednsServiceAccountName(tenant),
					Tolerations:        []corev1.Toleration{
						// TODO: do we need to add special tenant tolerations like CriticalAddonsOnly?
					},
					// TODO: do we need tenant-level priority classes?
				},
			},
		},
	}
}

func corednsLabels() map[string]string {
	return map[string]string{
		"app.kubernetes.io/name":       "coredns",
		"app.kubernetes.io/component":  "tenant-stack",
		"app.kubernetes.io/managed-by": "kubezoo",
	}
}
func corednsLabelSelector() *metav1.LabelSelector {
	return &metav1.LabelSelector{
		MatchLabels: map[string]string{
			"app.kubernetes.io/name":      "coredns",
			"app.kubernetes.io/component": "tenant-stack",
		},
	}
}

func corednsDeployName(tenant *tenantv1a1.Tenant) string         { return "coredns" }
func corednsNamespace(tenant *tenantv1a1.Tenant) string          { return "kube-system" }
func corednsConfigMapName(tenant *tenantv1a1.Tenant) string      { return "coredns" }
func corednsServiceAccountName(tenant *tenantv1a1.Tenant) string { return "coredns" }
func corednsImage() string                                       { return "registry.k8s.io/coredns/coredns:v1.9.3" }

func corednsReplicas(tenant *tenantv1a1.Tenant) int32 {
	return 1 // TODO read from tenant
}
func corednsResourceRequirements(tenant *tenantv1a1.Tenant) corev1.ResourceRequirements {
	return corev1.ResourceRequirements{
		// TODO read from tenant
	}
}
