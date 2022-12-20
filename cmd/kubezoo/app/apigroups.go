package app

import (
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	admissionregistrationv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	appsapiv1 "k8s.io/api/apps/v1"
	appsv1beta1 "k8s.io/api/apps/v1beta1"
	appsv1beta2 "k8s.io/api/apps/v1beta2"
	authenticationv1 "k8s.io/api/authentication/v1"
	authenticationv1beta1 "k8s.io/api/authentication/v1beta1"
	authorizationv1 "k8s.io/api/authorization/v1"
	authorizationv1beta1 "k8s.io/api/authorization/v1beta1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	autoscalingv2beta1 "k8s.io/api/autoscaling/v2beta1"
	autoscalingv2beta2 "k8s.io/api/autoscaling/v2beta2"
	batchapiv1 "k8s.io/api/batch/v1"
	batchapiv1beta1 "k8s.io/api/batch/v1beta1"
	certificatesv1beta1 "k8s.io/api/certificates/v1beta1"
	coordinationv1 "k8s.io/api/coordination/v1"
	coordinationv1beta1 "k8s.io/api/coordination/v1beta1"
	coreapiv1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	discoveryv1beta1 "k8s.io/api/discovery/v1beta1"
	eventsv1 "k8s.io/api/events/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	networkingv1 "k8s.io/api/networking/v1"
	networkingv1beta1 "k8s.io/api/networking/v1beta1"
	nodev1 "k8s.io/api/node/v1"
	policyv1 "k8s.io/api/policy/v1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	rbacv1 "k8s.io/api/rbac/v1"
	rbacv1alpha1 "k8s.io/api/rbac/v1alpha1"
	rbacv1beta1 "k8s.io/api/rbac/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apiserver/pkg/registry/rest"
	"k8s.io/kubernetes/pkg/apis/admissionregistration"
	"k8s.io/kubernetes/pkg/apis/apps"
	"k8s.io/kubernetes/pkg/apis/authentication"
	"k8s.io/kubernetes/pkg/apis/authorization"
	"k8s.io/kubernetes/pkg/apis/autoscaling"
	"k8s.io/kubernetes/pkg/apis/batch"
	"k8s.io/kubernetes/pkg/apis/certificates"
	"k8s.io/kubernetes/pkg/apis/coordination"
	"k8s.io/kubernetes/pkg/apis/core"
	"k8s.io/kubernetes/pkg/apis/discovery"
	"k8s.io/kubernetes/pkg/apis/networking"
	"k8s.io/kubernetes/pkg/apis/node"
	"k8s.io/kubernetes/pkg/apis/policy"
	"k8s.io/kubernetes/pkg/apis/rbac"

	"github.com/kubewharf/kubezoo/pkg/common"
)

var legacyGroup = common.APIGroupConfig{
	Group: coreapiv1.GroupName,
	StorageConfigs: map[string]map[string]*common.StorageConfig{
		"v1": {
			"pods": {
				Kind: coreapiv1.
					SchemeGroupVersion.WithKind("Pod"),
				Resource:        "pods",
				ShortNames:      []string{"po"},
				NamespaceScoped: true,
				NewFunc: func() runtime.Object {
					return &core.Pod{}
				},
				NewListFunc: func() runtime.Object {
					return &core.PodList{}
				},
			},
			"pods/attach": {
				Kind: coreapiv1.
					SchemeGroupVersion.WithKind("Pod"),
				Resource:        "pods",
				Subresource:     "attach",
				ShortNames:      []string{"po"},
				NamespaceScoped: true,
				NewFunc: func() runtime.Object {
					return &core.Pod{}
				},
			},
			"pods/status": {
				Kind: coreapiv1.
					SchemeGroupVersion.WithKind("Pod"),
				Resource:        "pods",
				Subresource:     "status",
				ShortNames:      []string{"po"},
				NamespaceScoped: true,
				NewFunc: func() runtime.Object {
					return &core.Pod{}
				},
			},
			"pods/log": {
				IsConnecter: true,
			},
			"pods/exec": {
				IsConnecter: true,
			},
			"pods/eviction": {
				Kind: coreapiv1.
					SchemeGroupVersion.WithKind("Pod"),
				Resource:        "pods",
				Subresource:     "eviction",
				ShortNames:      []string{"po"},
				NamespaceScoped: true,
				NewFunc: func() runtime.Object {
					return &core.Pod{}
				},
			},
			"pods/portforward": {
				Kind: coreapiv1.
					SchemeGroupVersion.WithKind("Pod"),
				Resource:        "pods",
				Subresource:     "portforward",
				ShortNames:      []string{"po"},
				NamespaceScoped: true,
				NewFunc: func() runtime.Object {
					return &core.Pod{}
				},
			},
			"pods/proxy": {
				Kind: coreapiv1.
					SchemeGroupVersion.WithKind("Pod"),
				Resource:        "pods",
				Subresource:     "proxy",
				ShortNames:      []string{"po"},
				NamespaceScoped: true,
				NewFunc: func() runtime.Object {
					return &core.Pod{}
				},
			},
			"pods/binding": {
				Kind: coreapiv1.
					SchemeGroupVersion.WithKind("Pod"),
				Resource:        "pods",
				Subresource:     "binding",
				ShortNames:      []string{"po"},
				NamespaceScoped: true,
				NewFunc: func() runtime.Object {
					return &core.Pod{}
				},
			},
			"pods/ephemeralcontainers": {
				Kind: coreapiv1.
					SchemeGroupVersion.WithKind("Pod"),
				Resource:        "pods",
				Subresource:     "ephemeralcontainers",
				ShortNames:      []string{"po"},
				NamespaceScoped: true,
				NewFunc: func() runtime.Object {
					return &core.Pod{}
				},
			},
			"bindings": {
				Kind: coreapiv1.
					SchemeGroupVersion.WithKind("Binding"),
				Resource:        "bindings",
				NamespaceScoped: true,
				NewFunc: func() runtime.Object {
					return &core.Binding{}
				},
			},
			"podtemplates": {
				Kind: coreapiv1.
					SchemeGroupVersion.WithKind("PodTemplate"),
				Resource:        "podtemplates",
				ShortNames:      []string{},
				NamespaceScoped: true,
				NewFunc: func() runtime.Object {
					return &core.PodTemplate{}
				},
				NewListFunc: func() runtime.Object {
					return &core.PodTemplateList{}
				},
			},

			"replicationcontrollers": {
				Kind: coreapiv1.
					SchemeGroupVersion.WithKind("ReplicationController"),
				Resource:        "replicationcontrollers",
				ShortNames:      []string{"rc"},
				NamespaceScoped: true,
				NewFunc: func() runtime.Object {
					return &core.ReplicationController{}
				},
				NewListFunc: func() runtime.Object {
					return &core.ReplicationControllerList{}
				},
			},
			"replicationcontrollers/status": {
				Kind: coreapiv1.
					SchemeGroupVersion.WithKind("ReplicationController"),
				Resource:        "replicationcontrollers",
				Subresource:     "status",
				ShortNames:      []string{"rc"},
				NamespaceScoped: true,
				NewFunc: func() runtime.Object {
					return &core.ReplicationController{}
				},
			},
			"replicationcontrollers/scale": {
				Kind: coreapiv1.
					SchemeGroupVersion.WithKind("ReplicationController"),
				Resource:        "replicationcontrollers",
				Subresource:     "scale",
				ShortNames:      []string{"rc"},
				NamespaceScoped: true,
				NewFunc: func() runtime.Object {
					return &autoscaling.Scale{}
				},
				GroupVersionKindFunc: groupVersionKindForScale,
			},

			"services": {
				Kind: coreapiv1.
					SchemeGroupVersion.WithKind("Service"),
				Resource:        "services",
				ShortNames:      []string{"svc"},
				NamespaceScoped: true,
				NewFunc: func() runtime.Object {
					return &core.Service{}
				},
				NewListFunc: func() runtime.Object {
					return &core.ServiceList{}
				},
			},
			"services/proxy": {
				Kind: coreapiv1.
					SchemeGroupVersion.WithKind("Service"),
				Resource:        "services",
				Subresource:     "proxy",
				ShortNames:      []string{"svc"},
				NamespaceScoped: true,
				NewFunc: func() runtime.Object {
					return &core.Service{}
				},
			},
			"services/status": {
				Kind: coreapiv1.
					SchemeGroupVersion.WithKind("Service"),
				Resource:        "services",
				Subresource:     "status",
				ShortNames:      []string{"svc"},
				NamespaceScoped: true,
				NewFunc: func() runtime.Object {
					return &core.Service{}
				},
			},

			"endpoints": {
				Kind: coreapiv1.
					SchemeGroupVersion.WithKind("Endpoints"),
				Resource:        "endpoints",
				ShortNames:      []string{"ep"},
				NamespaceScoped: true,
				NewFunc: func() runtime.Object {
					return &core.Endpoints{}
				},
				NewListFunc: func() runtime.Object {
					return &core.EndpointsList{}
				},
			},

			"nodes": {
				Kind: coreapiv1.
					SchemeGroupVersion.WithKind("Node"),
				Resource:        "nodes",
				ShortNames:      []string{"no"},
				NamespaceScoped: false,
				NewFunc: func() runtime.Object {
					return &core.Node{}
				},
				NewListFunc: func() runtime.Object {
					return &core.NodeList{}
				},
			},
			"nodes/status": {
				Kind: coreapiv1.
					SchemeGroupVersion.WithKind("Node"),
				Resource:        "nodes",
				Subresource:     "status",
				ShortNames:      []string{"no"},
				NamespaceScoped: false,
				NewFunc: func() runtime.Object {
					return &core.Node{}
				},
			},
			"nodes/proxy": {
				Kind: coreapiv1.
					SchemeGroupVersion.WithKind("Node"),
				Resource:        "nodes",
				Subresource:     "proxy",
				ShortNames:      []string{"no"},
				NamespaceScoped: false,
				NewFunc: func() runtime.Object {
					return &core.Node{}
				},
			},

			"events": {
				Kind: coreapiv1.
					SchemeGroupVersion.WithKind("Event"),
				Resource:        "events",
				ShortNames:      []string{"ev"},
				NamespaceScoped: true,
				NewFunc: func() runtime.Object {
					return &core.Event{}
				},
				NewListFunc: func() runtime.Object {
					return &core.EventList{}
				},
			},

			"limitranges": {
				Kind: coreapiv1.
					SchemeGroupVersion.WithKind("LimitRange"),
				Resource:        "limitranges",
				ShortNames:      []string{"limits"},
				NamespaceScoped: true,
				NewFunc: func() runtime.Object {
					return &core.LimitRange{}
				},
				NewListFunc: func() runtime.Object {
					return &core.LimitRangeList{}
				},
				TableConvertor: rest.NewDefaultTableConvertor(core.Resource("limitranges")),
			},

			"resourcequotas": {
				Kind: coreapiv1.
					SchemeGroupVersion.WithKind("ResourceQuota"),
				Resource:        "resourcequotas",
				ShortNames:      []string{"quota"},
				NamespaceScoped: true,
				NewFunc: func() runtime.Object {
					return &core.ResourceQuota{}
				},
				NewListFunc: func() runtime.Object {
					return &core.ResourceQuotaList{}
				},
			},
			"resourcequotas/status": {
				Kind: coreapiv1.
					SchemeGroupVersion.WithKind("ResourceQuota"),
				Resource:        "resourcequotas",
				Subresource:     "status",
				ShortNames:      []string{"quota"},
				NamespaceScoped: true,
				NewFunc: func() runtime.Object {
					return &core.ResourceQuota{}
				},
			},

			"namespaces": {
				Kind: coreapiv1.
					SchemeGroupVersion.WithKind("Namespace"),
				Resource:        "namespaces",
				ShortNames:      []string{"ns"},
				NamespaceScoped: false,
				NewFunc: func() runtime.Object {
					return &core.Namespace{}
				},
				NewListFunc: func() runtime.Object {
					return &core.NamespaceList{}
				},
			},

			"namespaces/status": {
				Kind: coreapiv1.
					SchemeGroupVersion.WithKind("Namespace"),
				Resource:        "namespaces",
				Subresource:     "status",
				ShortNames:      []string{"ns"},
				NamespaceScoped: false,
				NewFunc: func() runtime.Object {
					return &core.Namespace{}
				},
			},

			"namespaces/finalize": {
				Kind: coreapiv1.
					SchemeGroupVersion.WithKind("Namespace"),
				Resource:        "namespaces",
				Subresource:     "finalize",
				ShortNames:      []string{"ns"},
				NamespaceScoped: false,
				NewFunc: func() runtime.Object {
					return &core.Namespace{}
				},
			},

			"secrets": {
				Kind: coreapiv1.
					SchemeGroupVersion.WithKind("Secret"),
				Resource:        "secrets",
				ShortNames:      []string{},
				NamespaceScoped: true,
				NewFunc: func() runtime.Object {
					return &core.Secret{}
				},
				NewListFunc: func() runtime.Object {
					return &core.SecretList{}
				},
			},
			"serviceaccounts": {
				Kind: coreapiv1.
					SchemeGroupVersion.WithKind("ServiceAccount"),
				Resource:        "serviceaccounts",
				ShortNames:      []string{"sa"},
				NamespaceScoped: true,
				NewFunc: func() runtime.Object {
					return &core.ServiceAccount{}
				},
				NewListFunc: func() runtime.Object {
					return &core.ServiceAccountList{}
				},
			},

			"persistentvolumes": {
				Kind: coreapiv1.
					SchemeGroupVersion.WithKind("PersistentVolume"),
				Resource:        "persistentvolumes",
				ShortNames:      []string{"pv"},
				NamespaceScoped: false,
				NewFunc: func() runtime.Object {
					return &core.PersistentVolume{}
				},
				NewListFunc: func() runtime.Object {
					return &core.PersistentVolumeList{}
				},
			},
			"persistentvolumes/status": {
				Kind: coreapiv1.
					SchemeGroupVersion.WithKind("PersistentVolume"),
				Resource:        "persistentvolumes",
				Subresource:     "status",
				ShortNames:      []string{"pv"},
				NamespaceScoped: false,
				NewFunc: func() runtime.Object {
					return &core.PersistentVolume{}
				},
			},
			"persistentvolumeclaims": {
				Kind: coreapiv1.
					SchemeGroupVersion.WithKind("PersistentVolumeClaim"),
				Resource:        "persistentvolumeclaims",
				ShortNames:      []string{"pvc"},
				NamespaceScoped: true,
				NewFunc: func() runtime.Object {
					return &core.PersistentVolumeClaim{}
				},
				NewListFunc: func() runtime.Object {
					return &core.PersistentVolumeClaimList{}
				},
			},
			"configmaps": {
				Kind: coreapiv1.
					SchemeGroupVersion.WithKind("ConfigMap"),
				Resource:        "configmaps",
				ShortNames:      []string{"cm"},
				NamespaceScoped: true,
				NewFunc: func() runtime.Object {
					return &core.ConfigMap{}
				},
				NewListFunc: func() runtime.Object {
					return &core.ConfigMapList{}
				},
			},
			"serviceaccounts/token": {
				Kind: coreapiv1.
					SchemeGroupVersion.WithKind("ServiceAccount"),
				Resource:        "serviceaccounts",
				Subresource:     "token",
				ShortNames:      []string{"sa"},
				NamespaceScoped: true,
				NewFunc: func() runtime.Object {
					return &core.ServiceAccount{}
				},
			},

			"componentstatuses": {
				Kind: coreapiv1.
					SchemeGroupVersion.WithKind("ComponentStatus"),
				Resource:        "componentstatuses",
				ShortNames:      []string{"cs"},
				NamespaceScoped: false,
				NewFunc: func() runtime.Object {
					return &core.ComponentStatus{}
				},
				NewListFunc: func() runtime.Object {
					return &core.ComponentStatusList{}
				},
			},
		},
	},
}

var nonLegacyGroups = []common.APIGroupConfig{
	{
		// group: apps
		// ref: k8s.io/kubernetes/pkg/registry/apps/rest/storage_apps.go
		appsapiv1.GroupName,
		map[string]map[string]*common.StorageConfig{
			"v1": {
				// deployments
				"deployments": {
					Kind:            appsapiv1.SchemeGroupVersion.WithKind("Deployment"),
					Resource:        "deployments",
					ShortNames:      []string{"deploy"},
					NamespaceScoped: true,
					NewFunc:         func() runtime.Object { return &apps.Deployment{} },
					NewListFunc:     func() runtime.Object { return &apps.DeploymentList{} },
				},
				"deployments/status": {
					Kind:            appsapiv1.SchemeGroupVersion.WithKind("Deployment"),
					Resource:        "deployments",
					Subresource:     "status",
					NamespaceScoped: true,
					NewFunc:         func() runtime.Object { return &apps.Deployment{} },
				},
				"deployments/scale": {
					Kind:                 appsapiv1.SchemeGroupVersion.WithKind("Deployment"),
					Resource:             "deployments",
					Subresource:          "scale",
					NamespaceScoped:      true,
					NewFunc:              func() runtime.Object { return &autoscaling.Scale{} },
					GroupVersionKindFunc: groupVersionKindForScale,
				},

				// statefulsets
				"statefulsets": {
					Kind:            appsapiv1.SchemeGroupVersion.WithKind("StatefulSet"),
					Resource:        "statefulsets",
					ShortNames:      []string{"sts"},
					NamespaceScoped: true,
					NewFunc:         func() runtime.Object { return &apps.StatefulSet{} },
					NewListFunc:     func() runtime.Object { return &apps.StatefulSetList{} },
				},
				"statefulsets/status": {
					Kind:            appsapiv1.SchemeGroupVersion.WithKind("StatefulSet"),
					Resource:        "statefulsets",
					Subresource:     "status",
					NamespaceScoped: true,
					NewFunc:         func() runtime.Object { return &apps.StatefulSet{} },
				},
				"statefulsets/scale": {
					Kind:                 appsapiv1.SchemeGroupVersion.WithKind("StatefulSet"),
					Resource:             "statefulsets",
					Subresource:          "scale",
					NamespaceScoped:      true,
					NewFunc:              func() runtime.Object { return &autoscaling.Scale{} },
					GroupVersionKindFunc: groupVersionKindForScale,
				},

				// daemonsets
				"daemonsets": {
					Kind:            appsapiv1.SchemeGroupVersion.WithKind("DaemonSet"),
					Resource:        "daemonsets",
					ShortNames:      []string{"ds"},
					NamespaceScoped: true,
					NewFunc:         func() runtime.Object { return &apps.DaemonSet{} },
					NewListFunc:     func() runtime.Object { return &apps.DaemonSetList{} },
				},
				"daemonsets/status": {
					Kind:            appsapiv1.SchemeGroupVersion.WithKind("DaemonSet"),
					Resource:        "daemonsets",
					ShortNames:      []string{"ds"},
					Subresource:     "status",
					NamespaceScoped: true,
					NewFunc:         func() runtime.Object { return &apps.DaemonSet{} },
				},

				// replicasets
				"replicasets": {
					Kind:            appsapiv1.SchemeGroupVersion.WithKind("ReplicaSet"),
					Resource:        "replicasets",
					ShortNames:      []string{"rs"},
					NamespaceScoped: true,
					NewFunc:         func() runtime.Object { return &apps.ReplicaSet{} },
					NewListFunc:     func() runtime.Object { return &apps.ReplicaSetList{} },
				},
				"replicasets/status": {
					Kind:            appsapiv1.SchemeGroupVersion.WithKind("ReplicaSet"),
					Resource:        "replicasets",
					Subresource:     "status",
					NamespaceScoped: true,
					NewFunc:         func() runtime.Object { return &apps.ReplicaSet{} },
				},
				"replicasets/scale": {
					Kind:                 appsapiv1.SchemeGroupVersion.WithKind("ReplicaSet"),
					Resource:             "replicasets",
					Subresource:          "scale",
					NamespaceScoped:      true,
					NewFunc:              func() runtime.Object { return &autoscaling.Scale{} },
					GroupVersionKindFunc: groupVersionKindForScale,
				},

				// controllerrevisions
				"controllerrevisions": {
					Kind:            appsapiv1.SchemeGroupVersion.WithKind("ControllerRevision"),
					Resource:        "controllerrevisions",
					NamespaceScoped: true,
					NewFunc:         func() runtime.Object { return &apps.ControllerRevision{} },
					NewListFunc:     func() runtime.Object { return &apps.ControllerRevisionList{} },
				},
			},
		},
	},

	{
		authenticationv1.GroupName,
		map[string]map[string]*common.StorageConfig{
			"v1": {
				"tokenreviews": {
					Kind:            authenticationv1.SchemeGroupVersion.WithKind("TokenReview"),
					Resource:        "tokenreviews",
					NamespaceScoped: false,
					NewFunc:         func() runtime.Object { return &authentication.TokenReview{} },
				},
			},
			"v1beta1": {
				"tokenreviews": {
					Kind:            authenticationv1beta1.SchemeGroupVersion.WithKind("TokenReview"),
					Resource:        "tokenreviews",
					NamespaceScoped: false,
					NewFunc:         func() runtime.Object { return &authentication.TokenReview{} },
				},
			},
		},
	},

	{
		batchapiv1.GroupName,
		map[string]map[string]*common.StorageConfig{
			"v1": {
				"jobs": {
					Kind:            batchapiv1.SchemeGroupVersion.WithKind("Job"),
					Resource:        "jobs",
					NamespaceScoped: true,
					NewFunc:         func() runtime.Object { return &batch.Job{} },
					NewListFunc:     func() runtime.Object { return &batch.JobList{} },
				},
				"jobs/status": {
					Kind:            batchapiv1.SchemeGroupVersion.WithKind("Job"),
					Resource:        "jobs",
					Subresource:     "status",
					NamespaceScoped: true,
					NewFunc:         func() runtime.Object { return &batch.Job{} },
				},
				"cronjobs": {
					Kind:            batchapiv1.SchemeGroupVersion.WithKind("CronJob"),
					Resource:        "cronjobs",
					NamespaceScoped: true,
					NewFunc:         func() runtime.Object { return &batch.CronJob{} },
					NewListFunc:     func() runtime.Object { return &batch.CronJobList{} },
				},
				"cronjobs/status": {
					Kind:            batchapiv1.SchemeGroupVersion.WithKind("CronJob"),
					Resource:        "cronjobs",
					Subresource:     "status",
					NamespaceScoped: true,
					NewFunc:         func() runtime.Object { return &batch.CronJob{} },
				},
			},
			"v1beta1": {
				"cronjobs": {
					Kind:            batchapiv1beta1.SchemeGroupVersion.WithKind("CronJob"),
					Resource:        "cronjobs",
					ShortNames:      []string{"cj"},
					NamespaceScoped: true,
					NewFunc:         func() runtime.Object { return &batch.CronJob{} },
					NewListFunc:     func() runtime.Object { return &batch.CronJobList{} },
				},
				"cronjobs/status": {
					Kind:            batchapiv1beta1.SchemeGroupVersion.WithKind("CronJob"),
					Resource:        "cronjobs",
					Subresource:     "status",
					NamespaceScoped: true,
					NewFunc:         func() runtime.Object { return &batch.CronJob{} },
				},
			},
		},
	},

	{
		admissionregistrationv1.GroupName,
		map[string]map[string]*common.StorageConfig{
			"v1": {
				"validatingwebhookconfigurations": {
					Kind:            admissionregistrationv1.SchemeGroupVersion.WithKind("ValidatingWebhookConfiguration"),
					Resource:        "validatingwebhookconfigurations",
					NamespaceScoped: false,
					NewFunc:         func() runtime.Object { return &admissionregistration.ValidatingWebhookConfiguration{} },
					NewListFunc:     func() runtime.Object { return &admissionregistration.ValidatingWebhookConfigurationList{} },
				},
				"mutatingwebhookconfigurations": {
					Kind:            admissionregistrationv1.SchemeGroupVersion.WithKind("MutatingWebhookConfiguration"),
					Resource:        "mutatingwebhookconfigurations",
					NamespaceScoped: false,
					NewFunc:         func() runtime.Object { return &admissionregistration.MutatingWebhookConfiguration{} },
					NewListFunc:     func() runtime.Object { return &admissionregistration.MutatingWebhookConfigurationList{} },
				},
			},
			"v1beta1": {
				"validatingwebhookconfigurations": {
					Kind:            admissionregistrationv1beta1.SchemeGroupVersion.WithKind("ValidatingWebhookConfiguration"),
					Resource:        "validatingwebhookconfigurations",
					NamespaceScoped: false,
					NewFunc:         func() runtime.Object { return &admissionregistration.ValidatingWebhookConfiguration{} },
					NewListFunc:     func() runtime.Object { return &admissionregistration.ValidatingWebhookConfigurationList{} },
				},
				"mutatingwebhookconfigurations": {
					Kind:            admissionregistrationv1beta1.SchemeGroupVersion.WithKind("MutatingWebhookConfiguration"),
					Resource:        "mutatingwebhookconfigurations",
					NamespaceScoped: false,
					NewFunc:         func() runtime.Object { return &admissionregistration.MutatingWebhookConfiguration{} },
					NewListFunc:     func() runtime.Object { return &admissionregistration.MutatingWebhookConfigurationList{} },
				},
			},
		},
	},

	{
		eventsv1.GroupName,
		map[string]map[string]*common.StorageConfig{
			"v1": {
				"events": {
					Kind:            eventsv1.SchemeGroupVersion.WithKind("Event"),
					Resource:        "events",
					NamespaceScoped: true,
					NewFunc:         func() runtime.Object { return &eventsv1.Event{} },
					NewListFunc:     func() runtime.Object { return &eventsv1.EventList{} },
				},
			},
		},
	},

	{
		rbacv1.GroupName,
		map[string]map[string]*common.StorageConfig{
			"v1": {
				"roles": {
					Kind:            rbacv1.SchemeGroupVersion.WithKind("Role"),
					Resource:        "roles",
					NamespaceScoped: true,
					NewFunc:         func() runtime.Object { return &rbac.Role{} },
					NewListFunc:     func() runtime.Object { return &rbac.RoleList{} },
					TableConvertor:  rest.NewDefaultTableConvertor(rbac.Resource("roles")),
				},
				"rolebindings": {
					Kind:            rbacv1.SchemeGroupVersion.WithKind("RoleBinding"),
					Resource:        "rolebindings",
					NamespaceScoped: true,
					NewFunc:         func() runtime.Object { return &rbac.RoleBinding{} },
					NewListFunc:     func() runtime.Object { return &rbac.RoleBindingList{} },
				},
				"clusterroles": {
					Kind:            rbacv1.SchemeGroupVersion.WithKind("ClusterRole"),
					Resource:        "clusterroles",
					NamespaceScoped: false,
					NewFunc:         func() runtime.Object { return &rbac.ClusterRole{} },
					NewListFunc:     func() runtime.Object { return &rbac.ClusterRoleList{} },
					TableConvertor:  rest.NewDefaultTableConvertor(rbac.Resource("clusterroles")),
				},
				"clusterrolebindings": {
					Kind:            rbacv1.SchemeGroupVersion.WithKind("ClusterRoleBinding"),
					Resource:        "clusterrolebindings",
					NamespaceScoped: false,
					NewFunc:         func() runtime.Object { return &rbac.ClusterRoleBinding{} },
					NewListFunc:     func() runtime.Object { return &rbac.ClusterRoleBindingList{} },
				},
			},
			"v1alpha1": {
				"roles": {
					Kind:            rbacv1alpha1.SchemeGroupVersion.WithKind("Role"),
					Resource:        "roles",
					NamespaceScoped: true,
					NewFunc:         func() runtime.Object { return &rbac.Role{} },
					NewListFunc:     func() runtime.Object { return &rbac.RoleList{} },
					TableConvertor:  rest.NewDefaultTableConvertor(rbac.Resource("roles")),
				},
				"rolebindings": {
					Kind:            rbacv1alpha1.SchemeGroupVersion.WithKind("RoleBinding"),
					Resource:        "rolebindings",
					NamespaceScoped: true,
					NewFunc:         func() runtime.Object { return &rbac.RoleBinding{} },
					NewListFunc:     func() runtime.Object { return &rbac.RoleBindingList{} },
				},
				"clusterroles": {
					Kind:            rbacv1alpha1.SchemeGroupVersion.WithKind("ClusterRole"),
					Resource:        "clusterroles",
					NamespaceScoped: false,
					NewFunc:         func() runtime.Object { return &rbac.ClusterRole{} },
					NewListFunc:     func() runtime.Object { return &rbac.ClusterRoleList{} },
					TableConvertor:  rest.NewDefaultTableConvertor(rbac.Resource("clusterroles")),
				},
				"clusterrolebindings": {
					Kind:            rbacv1alpha1.SchemeGroupVersion.WithKind("ClusterRoleBinding"),
					Resource:        "clusterrolebindings",
					NamespaceScoped: false,
					NewFunc:         func() runtime.Object { return &rbac.ClusterRoleBinding{} },
					NewListFunc:     func() runtime.Object { return &rbac.ClusterRoleBindingList{} },
				},
			},
			"v1beta1": {
				"roles": {
					Kind:            rbacv1beta1.SchemeGroupVersion.WithKind("Role"),
					Resource:        "roles",
					NamespaceScoped: true,
					NewFunc:         func() runtime.Object { return &rbac.Role{} },
					NewListFunc:     func() runtime.Object { return &rbac.RoleList{} },
					TableConvertor:  rest.NewDefaultTableConvertor(rbac.Resource("roles")),
				},
				"rolebindings": {
					Kind:            rbacv1beta1.SchemeGroupVersion.WithKind("RoleBinding"),
					Resource:        "rolebindings",
					NamespaceScoped: true,
					NewFunc:         func() runtime.Object { return &rbac.RoleBinding{} },
					NewListFunc:     func() runtime.Object { return &rbac.RoleBindingList{} },
				},
				"clusterroles": {
					Kind:            rbacv1beta1.SchemeGroupVersion.WithKind("ClusterRole"),
					Resource:        "clusterroles",
					NamespaceScoped: false,
					NewFunc:         func() runtime.Object { return &rbac.ClusterRole{} },
					NewListFunc:     func() runtime.Object { return &rbac.ClusterRoleList{} },
				},
				"clusterrolebindings": {
					Kind:            rbacv1beta1.SchemeGroupVersion.WithKind("ClusterRoleBinding"),
					Resource:        "clusterrolebindings",
					NamespaceScoped: false,
					NewFunc:         func() runtime.Object { return &rbac.ClusterRoleBinding{} },
					NewListFunc:     func() runtime.Object { return &rbac.ClusterRoleBindingList{} },
				},
			},
		},
	},

	{
		policyv1beta1.GroupName,
		map[string]map[string]*common.StorageConfig{
			"v1beta1": {
				"poddisruptionbudgets": {
					Kind:            policyv1beta1.SchemeGroupVersion.WithKind("PodDisruptionBudget"),
					Resource:        "poddisruptionbudgets",
					NamespaceScoped: true,
					ShortNames:      []string{"pdb"},
					NewFunc:         func() runtime.Object { return &policy.PodDisruptionBudget{} },
					NewListFunc:     func() runtime.Object { return &policy.PodDisruptionBudgetList{} },
				},
				"podsecuritypolicies": {
					Kind:            policyv1beta1.SchemeGroupVersion.WithKind("PodSecurityPolicy"),
					Resource:        "podsecuritypolicies",
					NamespaceScoped: false,
					ShortNames:      []string{"psp"},
					NewFunc:         func() runtime.Object { return &policy.PodSecurityPolicy{} },
					NewListFunc:     func() runtime.Object { return &policy.PodSecurityPolicyList{} },
				},
			},
			"v1": {
				"poddisruptionbudgets": {
					Kind:            policyv1.SchemeGroupVersion.WithKind("PodDisruptionBudget"),
					Resource:        "poddisruptionbudgets",
					NamespaceScoped: true,
					ShortNames:      []string{"pdb"},
					NewFunc:         func() runtime.Object { return &policy.PodDisruptionBudget{} },
					NewListFunc:     func() runtime.Object { return &policy.PodDisruptionBudgetList{} },
				},
				"poddisruptionbudgets/status": {
					Kind:            policyv1.SchemeGroupVersion.WithKind("PodDisruptionBudget"),
					Resource:        "poddisruptionbudgets",
					Subresource:     "status",
					NamespaceScoped: true,
					ShortNames:      []string{"pdb"},
					NewFunc:         func() runtime.Object { return &policy.PodDisruptionBudget{} },
				},
			},
		},
	},

	{
		networkingv1.GroupName,
		map[string]map[string]*common.StorageConfig{
			"v1": {
				"networkpolicies": {
					Kind:            networkingv1.SchemeGroupVersion.WithKind("NetworkPolicy"),
					Resource:        "networkpolicies",
					NamespaceScoped: true,
					ShortNames:      []string{"netpol"},
					NewFunc:         func() runtime.Object { return &networking.NetworkPolicy{} },
					NewListFunc:     func() runtime.Object { return &networking.NetworkPolicyList{} },
				},
				"ingresses": {
					Kind:            networkingv1.SchemeGroupVersion.WithKind("Ingress"),
					Resource:        "ingresses",
					NamespaceScoped: true,
					ShortNames:      []string{"ing"},
					NewFunc:         func() runtime.Object { return &networking.Ingress{} },
					NewListFunc:     func() runtime.Object { return &networking.IngressList{} },
				},
				"ingresses/status": {
					Kind:            networkingv1.SchemeGroupVersion.WithKind("Ingress"),
					Resource:        "ingresses",
					Subresource:     "status",
					NamespaceScoped: true,
					ShortNames:      []string{"ing"},
					NewFunc:         func() runtime.Object { return &networking.Ingress{} },
				},
				"ingressclasses": {
					Kind:            networkingv1.SchemeGroupVersion.WithKind("IngressClass"),
					Resource:        "ingressclasses",
					NamespaceScoped: false,
					NewFunc:         func() runtime.Object { return &networking.IngressClass{} },
					NewListFunc:     func() runtime.Object { return &networking.IngressClassList{} },
				},
			},
			"v1beta1": {
				"ingresses": {
					Kind:            networkingv1beta1.SchemeGroupVersion.WithKind("Ingress"),
					Resource:        "ingresses",
					NamespaceScoped: true,
					ShortNames:      []string{"ing"},
					NewFunc:         func() runtime.Object { return &networking.Ingress{} },
					NewListFunc:     func() runtime.Object { return &networking.IngressList{} },
				},
				"ingresses/status": {
					Kind:            networkingv1beta1.SchemeGroupVersion.WithKind("Ingress"),
					Resource:        "ingresses",
					Subresource:     "status",
					NamespaceScoped: true,
					ShortNames:      []string{"ing"},
					NewFunc:         func() runtime.Object { return &networking.Ingress{} },
				},
				"ingressclasses": {
					Kind:            networkingv1beta1.SchemeGroupVersion.WithKind("IngressClass"),
					Resource:        "ingressclasses",
					NamespaceScoped: false,
					NewFunc:         func() runtime.Object { return &networking.IngressClass{} },
					NewListFunc:     func() runtime.Object { return &networking.IngressClassList{} },
				},
			},
		},
	},

	{
		extensionsv1beta1.GroupName,
		map[string]map[string]*common.StorageConfig{
			"v1beta1": {
				"ingresses": {
					Kind:            extensionsv1beta1.SchemeGroupVersion.WithKind("Ingress"),
					Resource:        "ingresses",
					NamespaceScoped: true,
					ShortNames:      []string{"ing"},
					NewFunc:         func() runtime.Object { return &networking.Ingress{} },
					NewListFunc:     func() runtime.Object { return &networking.IngressList{} },
				},
				"ingresses/status": {
					Kind:            extensionsv1beta1.SchemeGroupVersion.WithKind("Ingress"),
					Resource:        "ingresses",
					Subresource:     "status",
					NamespaceScoped: true,
					ShortNames:      []string{"ing"},
					NewFunc:         func() runtime.Object { return &networking.Ingress{} },
				},
			},
		},
	},

	{
		discoveryv1beta1.GroupName,
		map[string]map[string]*common.StorageConfig{
			"v1beta1": {
				"endpointslices": {
					Kind:            discoveryv1beta1.SchemeGroupVersion.WithKind("EndpointSlice"),
					Resource:        "endpointslices",
					NamespaceScoped: true,
					NewFunc:         func() runtime.Object { return &discovery.EndpointSlice{} },
					NewListFunc:     func() runtime.Object { return &discovery.EndpointSliceList{} },
				},
			},
			"v1": {
				"endpointslices": {
					Kind:            discoveryv1.SchemeGroupVersion.WithKind("EndpointSlice"),
					Resource:        "endpointslices",
					NamespaceScoped: true,
					NewFunc:         func() runtime.Object { return &discovery.EndpointSlice{} },
					NewListFunc:     func() runtime.Object { return &discovery.EndpointSliceList{} },
				},
			},
		},
	},

	{
		coordinationv1.GroupName,
		map[string]map[string]*common.StorageConfig{
			"v1": {
				"leases": {
					Kind:            coordinationv1.SchemeGroupVersion.WithKind("Lease"),
					Resource:        "leases",
					NamespaceScoped: true,
					NewFunc:         func() runtime.Object { return &coordination.Lease{} },
					NewListFunc:     func() runtime.Object { return &coordination.LeaseList{} },
				},
			},
			"v1beta1": {
				"leases": {
					Kind:            coordinationv1beta1.SchemeGroupVersion.WithKind("Lease"),
					Resource:        "leases",
					NamespaceScoped: true,
					NewFunc:         func() runtime.Object { return &coordination.Lease{} },
					NewListFunc:     func() runtime.Object { return &coordination.LeaseList{} },
				},
			},
		},
	},

	{
		certificatesv1beta1.GroupName,
		map[string]map[string]*common.StorageConfig{
			"v1beta1": {
				"certificatesigningrequests": {
					Kind:            certificatesv1beta1.SchemeGroupVersion.WithKind("CertificateSigningRequest"),
					Resource:        "certificatesigningrequests",
					NamespaceScoped: false,
					NewFunc:         func() runtime.Object { return &certificates.CertificateSigningRequest{} },
					NewListFunc:     func() runtime.Object { return &certificates.CertificateSigningRequestList{} },
				},
				"certificatesigningrequests/status": {
					Kind:            certificatesv1beta1.SchemeGroupVersion.WithKind("CertificateSigningRequest"),
					Resource:        "certificatesigningrequests",
					Subresource:     "status",
					NamespaceScoped: false,
					NewFunc:         func() runtime.Object { return &certificates.CertificateSigningRequest{} },
				},
				"certificatesigningrequests/approval": {
					Kind:            certificatesv1beta1.SchemeGroupVersion.WithKind("CertificateSigningRequest"),
					Resource:        "certificatesigningrequests",
					Subresource:     "approval",
					NamespaceScoped: false,
					NewFunc:         func() runtime.Object { return &certificates.CertificateSigningRequest{} },
				},
			},
		},
	},

	{
		autoscalingv1.GroupName,
		map[string]map[string]*common.StorageConfig{
			"v1": {
				"horizontalpodautoscalers": {
					Kind:            autoscalingv1.SchemeGroupVersion.WithKind("HorizontalPodAutoscaler"),
					Resource:        "horizontalpodautoscalers",
					NamespaceScoped: true,
					ShortNames:      []string{"hpa"},
					NewFunc:         func() runtime.Object { return &autoscaling.HorizontalPodAutoscaler{} },
					NewListFunc:     func() runtime.Object { return &autoscaling.HorizontalPodAutoscalerList{} },
				},
				"horizontalpodautoscalers/status": {
					Kind:            autoscalingv1.SchemeGroupVersion.WithKind("HorizontalPodAutoscaler"),
					Resource:        "horizontalpodautoscalers",
					Subresource:     "status",
					NamespaceScoped: true,
					ShortNames:      []string{"hpa"},
					NewFunc:         func() runtime.Object { return &autoscaling.HorizontalPodAutoscaler{} },
				},
			},
			"v2beta1": {
				"horizontalpodautoscalers": {
					Kind:            autoscalingv2beta1.SchemeGroupVersion.WithKind("HorizontalPodAutoscaler"),
					Resource:        "horizontalpodautoscalers",
					NamespaceScoped: true,
					ShortNames:      []string{"hpa"},
					NewFunc:         func() runtime.Object { return &autoscaling.HorizontalPodAutoscaler{} },
					NewListFunc:     func() runtime.Object { return &autoscaling.HorizontalPodAutoscalerList{} },
				},
				"horizontalpodautoscalers/status": {
					Kind:            autoscalingv2beta1.SchemeGroupVersion.WithKind("HorizontalPodAutoscaler"),
					Resource:        "horizontalpodautoscalers",
					Subresource:     "status",
					NamespaceScoped: true,
					ShortNames:      []string{"hpa"},
					NewFunc:         func() runtime.Object { return &autoscaling.HorizontalPodAutoscaler{} },
				},
			},
			"v2beta2": {
				"horizontalpodautoscalers": {
					Kind:            autoscalingv2beta2.SchemeGroupVersion.WithKind("HorizontalPodAutoscaler"),
					Resource:        "horizontalpodautoscalers",
					NamespaceScoped: true,
					ShortNames:      []string{"hpa"},
					NewFunc:         func() runtime.Object { return &autoscaling.HorizontalPodAutoscaler{} },
					NewListFunc:     func() runtime.Object { return &autoscaling.HorizontalPodAutoscalerList{} },
				},
				"horizontalpodautoscalers/status": {
					Kind:            autoscalingv2beta2.SchemeGroupVersion.WithKind("HorizontalPodAutoscaler"),
					Resource:        "horizontalpodautoscalers",
					Subresource:     "status",
					NamespaceScoped: true,
					ShortNames:      []string{"hpa"},
					NewFunc:         func() runtime.Object { return &autoscaling.HorizontalPodAutoscaler{} },
				},
			},
			"v2": {
				"horizontalpodautoscalers": {
					Kind:            autoscalingv2.SchemeGroupVersion.WithKind("HorizontalPodAutoscaler"),
					Resource:        "horizontalpodautoscalers",
					NamespaceScoped: true,
					ShortNames:      []string{"hpa"},
					NewFunc:         func() runtime.Object { return &autoscaling.HorizontalPodAutoscaler{} },
					NewListFunc:     func() runtime.Object { return &autoscaling.HorizontalPodAutoscalerList{} },
				},
				"horizontalpodautoscalers/status": {
					Kind:            autoscalingv2.SchemeGroupVersion.WithKind("HorizontalPodAutoscaler"),
					Resource:        "horizontalpodautoscalers",
					Subresource:     "status",
					NamespaceScoped: true,
					ShortNames:      []string{"hpa"},
					NewFunc:         func() runtime.Object { return &autoscaling.HorizontalPodAutoscaler{} },
				},
			},
		},
	},

	{
		authorizationv1.GroupName,
		map[string]map[string]*common.StorageConfig{
			"v1": {
				"subjectaccessreviews": {
					Kind:            authorizationv1.SchemeGroupVersion.WithKind("SubjectAccessReview"),
					Resource:        "subjectaccessreviews",
					NamespaceScoped: false,
					NewFunc:         func() runtime.Object { return &authorization.SubjectAccessReview{} },
				},
				"selfsubjectaccessreviews": {
					Kind:            authorizationv1.SchemeGroupVersion.WithKind("SelfSubjectAccessReview"),
					Resource:        "selfsubjectaccessreviews",
					NamespaceScoped: false,
					NewFunc:         func() runtime.Object { return &authorization.SelfSubjectAccessReview{} },
				},
				"localsubjectaccessreviews": {
					Kind:            authorizationv1.SchemeGroupVersion.WithKind("LocalSubjectAccessReview"),
					Resource:        "localsubjectaccessreviews",
					NamespaceScoped: true,
					NewFunc:         func() runtime.Object { return &authorization.LocalSubjectAccessReview{} },
				},
				"selfsubjectrulesreviews": {
					Kind:            authorizationv1.SchemeGroupVersion.WithKind("SelfSubjectRulesReview"),
					Resource:        "selfsubjectrulesreviews",
					NamespaceScoped: false,
					NewFunc:         func() runtime.Object { return &authorization.SelfSubjectRulesReview{} },
				},
			},
			"v1beta1": {
				"subjectaccessreviews": {
					Kind:            authorizationv1beta1.SchemeGroupVersion.WithKind("SubjectAccessReview"),
					Resource:        "subjectaccessreviews",
					NamespaceScoped: false,
					NewFunc:         func() runtime.Object { return &authorization.SubjectAccessReview{} },
				},
				"selfsubjectaccessreviews": {
					Kind:            authorizationv1beta1.SchemeGroupVersion.WithKind("SelfSubjectAccessReview"),
					Resource:        "selfsubjectaccessreviews",
					NamespaceScoped: false,
					NewFunc:         func() runtime.Object { return &authorization.SelfSubjectAccessReview{} },
				},
				"localsubjectaccessreviews": {
					Kind:            authorizationv1beta1.SchemeGroupVersion.WithKind("LocalSubjectAccessReview"),
					Resource:        "localsubjectaccessreviews",
					NamespaceScoped: true,
					NewFunc:         func() runtime.Object { return &authorization.LocalSubjectAccessReview{} },
				},
				"selfsubjectrulesreviews": {
					Kind:            authorizationv1beta1.SchemeGroupVersion.WithKind("SelfSubjectRulesReview"),
					Resource:        "selfsubjectrulesreviews",
					NamespaceScoped: false,
					NewFunc:         func() runtime.Object { return &authorization.SelfSubjectRulesReview{} },
				},
			},
		},
	},

	{
		nodev1.GroupName,
		map[string]map[string]*common.StorageConfig{
			"v1": {
				"runtimeclasses": {
					Kind:            nodev1.SchemeGroupVersion.WithKind("RuntimeClass"),
					Resource:        "runtimeclasses",
					NamespaceScoped: false,
					NewFunc: func() runtime.Object {
						return &node.RuntimeClass{}
					},
					NewListFunc: func() runtime.Object {
						return &node.RuntimeClassList{}
					},
				},
			},
		},
	},

	// the following kinds should not be available to serverless kubernetes users, so the api configs are skipped.
	// group: storage.k8s.io
	// kinds: CSIDriver, CSINode, StorageClass, VolumeAttachment

	// group: scheduling.k8s.io
	// kinds: PriorityClass

	// group: node.k8s.io
	// kinds: RuntimeClass
}

func groupVersionKindForScale(containingGV schema.GroupVersion) schema.GroupVersionKind {
	switch containingGV {
	case extensionsv1beta1.SchemeGroupVersion:
		return extensionsv1beta1.SchemeGroupVersion.WithKind("Scale")
	case appsv1beta1.SchemeGroupVersion:
		return appsv1beta1.SchemeGroupVersion.WithKind("Scale")
	case appsv1beta2.SchemeGroupVersion:
		return appsv1beta2.SchemeGroupVersion.WithKind("Scale")
	default:
		return autoscalingv1.SchemeGroupVersion.WithKind("Scale")
	}
}
