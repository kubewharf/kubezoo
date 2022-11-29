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
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes"
	typedappsv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	typedrbacv1 "k8s.io/client-go/kubernetes/typed/rbac/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/retry"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/component-helpers/auth/rbac/reconciliation"
	"k8s.io/klog"
	"k8s.io/kubernetes/pkg/apis/rbac"
	rbacv1helpers "k8s.io/kubernetes/pkg/apis/rbac/v1"

	tenantv1a1 "github.com/kubewharf/kubezoo/pkg/apis/tenant/v1alpha1"
	"github.com/kubewharf/kubezoo/pkg/dynamic"
	tenantclient "github.com/kubewharf/kubezoo/pkg/generated/clientset/versioned/typed/tenant/v1alpha1"
	tenantlister "github.com/kubewharf/kubezoo/pkg/generated/listers/tenant/v1alpha1"
	"github.com/kubewharf/kubezoo/pkg/util"
)

type EventType int

const (
	Add = iota
	Update
	Delete
)

const (
	maxRetries         = 10
	tenantFinalizerKey = "kubezoo.io/tenant"
	verbList           = "list"
	verbDelete         = "delete"
)

// Event indicate the informerEvent
type Event struct {
	tenantId  string
	eventType EventType
}

// TenantController take responsibility for tenant management including
// the tenant object, tenant certificate and the tenant's k8s resources.
type TenantController struct {
	queue                   workqueue.RateLimitingInterface
	tenantInformer          cache.SharedIndexInformer
	tenantLister            tenantlister.TenantLister
	tenantClient            tenantclient.TenantV1alpha1Interface
	upstreamAppsClient      typedappsv1.AppsV1Interface
	upstreamCoreClient      typedcorev1.CoreV1Interface
	upstreamRbacClient      typedrbacv1.RbacV1Interface
	upstreamDiscoveryClient *discovery.DiscoveryClient
	upstreamDynamicClient   dynamic.Interface
	upstreamCRDClient       *apiextensions.Clientset
	clientCAFile            string
	clientCAKeyFile         string
	kubeZooBindAddress      string
	kubeZooSecurePort       int
}

func setupEventHandler(tc *TenantController) {
	tc.tenantInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			tenantId, err := cache.MetaNamespaceKeyFunc(obj)
			if err == nil {
				tc.queue.Add(Event{tenantId: tenantId, eventType: Add})
			}
		},
		UpdateFunc: func(_, new interface{}) {
			tenantId, err := cache.MetaNamespaceKeyFunc(new)
			if err == nil {
				tc.queue.Add(Event{tenantId: tenantId, eventType: Update})
			}
		},
		DeleteFunc: func(obj interface{}) {
			tenantId, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err == nil {
				tc.queue.Add(Event{tenantId: tenantId, eventType: Delete})
			}
		},
	})
}

// Run starts the tenant controller
func Run(
	stopCh <-chan struct{},
	ti cache.SharedIndexInformer,
	tenantClient tenantclient.TenantV1alpha1Interface,
	k8sClient kubernetes.Interface,
	discoveryClient *discovery.DiscoveryClient,
	dynamicClient dynamic.Interface,
	crdClient *apiextensions.Clientset,
	clientCAFile, clientCAKeyFile string,
	kubeZooBindAddress string, kubeZooSecurePort int,
) {
	tc := &TenantController{
		queue:                   workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter()),
		tenantInformer:          ti,
		tenantLister:            tenantlister.NewTenantLister(ti.GetIndexer()),
		tenantClient:            tenantClient,
		upstreamCoreClient:      k8sClient.CoreV1(),
		upstreamAppsClient:      k8sClient.AppsV1(),
		upstreamRbacClient:      k8sClient.RbacV1(),
		upstreamDiscoveryClient: discoveryClient,
		upstreamDynamicClient:   dynamicClient,
		upstreamCRDClient:       crdClient,
		clientCAFile:            clientCAFile,
		clientCAKeyFile:         clientCAKeyFile,
		kubeZooBindAddress:      kubeZooBindAddress,
		kubeZooSecurePort:       kubeZooSecurePort,
	}
	setupEventHandler(tc)

	defer utilruntime.HandleCrash()
	defer tc.queue.ShutDown()

	klog.V(4).Info("Starting Tenant Controller")

	go tc.tenantInformer.Run(stopCh)

	if !cache.WaitForCacheSync(stopCh, tc.HasSynced) {
		utilruntime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
		return
	}

	klog.V(4).Info("Tenant controller synced and ready")
	wait.Until(tc.runWorker, time.Second, stopCh)
}

// HasSynced returns true if the shared informer's store has been
// informed by at least one full LIST of the authoritative state
// of the informer's object collection.
func (tc *TenantController) HasSynced() bool {
	return tc.tenantInformer.HasSynced()
}

// runWorker start to process the events.
func (tc *TenantController) runWorker() {
	for tc.processNextItem() {
	}
}

// processNextItem gets event from queue and process it.
func (tc *TenantController) processNextItem() bool {
	ctx := context.TODO()

	newEvent, quit := tc.queue.Get()

	if quit {
		return false
	}
	defer tc.queue.Done(newEvent)
	err := tc.processItem(ctx, newEvent.(Event))
	if err == nil {
		// No error, reset the ratelimit counters
		tc.queue.Forget(newEvent)
	} else if tc.queue.NumRequeues(newEvent) < maxRetries {
		klog.Errorf("Error processing %s (will retry): %v", newEvent.(Event).tenantId, err)
		tc.queue.AddRateLimited(newEvent)
	} else {
		// err != nil and too many retries
		klog.Errorf("Error processing %s (giving up): %v", newEvent.(Event).tenantId, err)
		tc.queue.Forget(newEvent)
		utilruntime.HandleError(err)
	}

	return true
}

// processItem processes the event according the event type.
func (tc *TenantController) processItem(ctx context.Context, e Event) error {
	// process events based on its type
	switch e.eventType {
	case Add, Update:
		return tc.onTenantAddOrUpdate(ctx, e.tenantId)
	case Delete:
		klog.Warningf("deleting tenant %v", e.tenantId)
		return nil
	}
	return nil
}

// onTenantAddOrUpdate handles the ADD or UPDATE event of a Tenant.
func (tc *TenantController) onTenantAddOrUpdate(ctx context.Context, tenantId string) error {
	tenant, err := tc.tenantClient.Tenants().Get(context.TODO(), tenantId, metav1.GetOptions{})
	if err != nil {
		return err
	}

	if tenant.ObjectMeta.DeletionTimestamp.IsZero() {
		if !util.ContainString(tenant.ObjectMeta.Finalizers, tenantFinalizerKey) {
			tenant.ObjectMeta.Finalizers = append(tenant.ObjectMeta.Finalizers, tenantFinalizerKey)
			if _, err := tc.tenantClient.Tenants().Update(context.TODO(), tenant, metav1.UpdateOptions{}); err != nil {
				return err
			}
		}
	} else {
		if util.ContainString(tenant.ObjectMeta.Finalizers, tenantFinalizerKey) {
			if err = tc.deleteResources(tenantId); err != nil {
				return err
			}
			tenant.ObjectMeta.Finalizers = util.RemoveString(tenant.ObjectMeta.Finalizers, tenantFinalizerKey)
			if _, err = tc.tenantClient.Tenants().Update(context.TODO(), tenant, metav1.UpdateOptions{}); err != nil {
				if apierrors.IsNotFound(err) {
					return nil
				}
				return err
			}
		}
		return nil
	}

	return tc.syncResources(ctx, tenantId, tc.kubeZooBindAddress, tc.kubeZooSecurePort)
}

// deleteResources deletes resources belonging to the tenant from the upstream cluster.
func (tc *TenantController) deleteResources(tenantId string) error {
	klog.V(4).Infof("delete resources for tenant %s", tenantId)

	clusterScopedResources, err := tc.getClusterScopedResources()
	if err != nil {
		return err
	}
	if err := tc.deleteCRDs(tenantId); err != nil {
		return err
	}
	nonCRDResources := tc.filterCRDs(clusterScopedResources)
	if err := tc.deleteNonCRDClusterScopedResources(tenantId, nonCRDResources); err != nil {
		return err
	}

	klog.Infof("deleted resources for tenant %s", tenantId)
	return nil
}

// genClusterScopedResourceList generates the list of the cluster-scoped resources.
func (tc *TenantController) getClusterScopedResources() ([]metav1.APIResource, error) {
	_, apiResourceLists, err := tc.upstreamDiscoveryClient.ServerGroupsAndResources()
	if err != nil {
		return nil, err
	}

	clusterScopedLists := discovery.FilteredBy(
		discovery.ResourcePredicateFunc(
			func(gv string, r *metav1.APIResource) bool {
				return !r.Namespaced
			},
		), apiResourceLists)

	return util.FlattenResourceLists(clusterScopedLists), nil
}

// deleteCRDs delete all crds belong to the tenant.
// NOTE: when a CRD is deleted, all its associated CRs will be removed automatically.
func (tc *TenantController) deleteCRDs(tenantId string) error {
	crdList, err := tc.upstreamCRDClient.ApiextensionsV1().CustomResourceDefinitions().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return err
	}

	for _, crd := range crdList.Items {
		if strings.HasPrefix(crd.Spec.Group, tenantId) {
			if err = tc.upstreamCRDClient.ApiextensionsV1().CustomResourceDefinitions().Delete(context.TODO(), crd.GetName(), metav1.DeleteOptions{}); err != nil {
				if apierrors.IsNotFound(err) {
					continue
				}
				return err
			}
			klog.Infof("delete crd(%s) for tenant %s", crd.GetName(), tenantId)
			continue
		}
		klog.V(4).Infof("crd(%s) does not belong to tenant %s", crd.GetName(), tenantId)
	}

	return nil
}

// deleteNonCRDClusterScopedResources delete all non-crd resources belong to the tenant.
func (tc *TenantController) deleteNonCRDClusterScopedResources(tenantId string, nonCRDAPIResources []metav1.APIResource) error {
	for _, apiResource := range nonCRDAPIResources {
		if !util.ContainString(apiResource.Verbs, verbList) || !util.ContainString(apiResource.Verbs, verbDelete) {
			continue
		}

		gvr := util.GetGVR(apiResource)
		rClient := tc.upstreamDynamicClient.Resource(gvr)

		resourceList, err := rClient.List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				continue
			}
			return err
		}

		for _, resource := range resourceList.Items {
			if strings.HasPrefix(resource.GetName(), tenantId) {
				if _, _, err = rClient.Delete(context.TODO(), resource.GetName(), metav1.DeleteOptions{}); err != nil {
					if apierrors.IsNotFound(err) {
						continue
					}
					return err
				}
				klog.Infof("delete cluster-scoped resource (%s) for tenant %s", resource.GetName(), tenantId)
			} else {
				klog.V(4).Infof("cluster-scoped resource (%s) does not belong to tenant %s", resource.GetName(), tenantId)
			}
		}
	}

	return nil
}

// filterCRDs split cluster-scoped resources to CRDs and non-CRDs.
func (tc *TenantController) filterCRDs(clusterScopedAPIResources []metav1.APIResource) []metav1.APIResource {
	nonCRDs := make([]metav1.APIResource, 0, len(clusterScopedAPIResources))
	for _, resource := range clusterScopedAPIResources {
		if !util.IsCRD(resource) {
			nonCRDs = append(nonCRDs, resource)
		}
	}

	return nonCRDs
}

// syncResources sync system resources to the upstream cluster when new tenant is being created.
func (tc *TenantController) syncResources(ctx context.Context, tenantId string, zooHost string, zooPort int) error {
	tenant, err := tc.tenantLister.Get(tenantId)
	if err != nil {
		return fmt.Errorf("get tenant from lister: %w", err)
	}

	if tc.tenantClient == nil || tc.upstreamCoreClient == nil || tc.upstreamRbacClient == nil {
		return errors.New("Skip synchronize namespaces or RBAC resources since nil client.")
	}

	klog.V(4).Infof("Sync system resources for tenant %s", tenantId)
	if err := syncNamespaces(tc.upstreamCoreClient, tenantId); err != nil {
		return fmt.Errorf("sync namespace: %w", err)
	}
	if err := syncClusterRoles(tc.upstreamCoreClient, tc.upstreamRbacClient, tenantId); err != nil {
		return fmt.Errorf("sync cluster role: %w", err)
	}
	if err := syncClusterRoleBindings(tc.upstreamCoreClient, tc.upstreamRbacClient, tenantId); err != nil {
		return fmt.Errorf("sync cluster role binding: %w", err)
	}
	if err := syncTenantStack(ctx, tc.upstreamCoreClient, tc.upstreamAppsClient, tenant, zooHost, zooPort); err != nil {
		return fmt.Errorf("sync tenant stack: %w", err)
	}
	if err := genCertAndKubeconfig(tc.tenantClient, tenantId, tc.tenantLister, tc.clientCAFile, tc.clientCAKeyFile, tc.kubeZooBindAddress, tc.kubeZooSecurePort); err != nil {
		return fmt.Errorf("gen cert and kubeconfig: %w", err)
	}
	return nil
}

// syncNamespaces synchronize the system namespaces to upstream cluster.
func syncNamespaces(coreClient typedcorev1.CoreV1Interface, tenantId string) error {
	systemNamespaces := []string{metav1.NamespaceSystem, metav1.NamespacePublic, corev1.NamespaceNodeLease, corev1.NamespaceDefault}

	for _, systemNamespace := range systemNamespaces {
		namespace := tenantId + "-" + systemNamespace

		if _, err := coreClient.Namespaces().Get(context.TODO(), namespace, metav1.GetOptions{}); err != nil {
			newNamespace := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name:      namespace,
					Namespace: "",
				},
			}

			_, err = coreClient.Namespaces().Create(context.TODO(), newNamespace, metav1.CreateOptions{})
			if err != nil && !apierrors.IsAlreadyExists(err) {
				klog.Warningf("Failed to create the tenant namespace %s with error %v", namespace, err)
				return err
			}
		}
	}
	return nil
}

// syncClusterRoles synchronize the cluster roles to upstream cluster.
func syncClusterRoles(coreClient typedcorev1.CoreV1Interface, rbacClient typedrbacv1.RbacV1Interface, tenantId string) error {
	if _, err := rbacClient.ClusterRoles().List(context.TODO(), metav1.ListOptions{ResourceVersion: "0"}); err != nil {
		klog.Warningf("Failed to list the clusterroles %s with error %v", tenantId, err)
		return err
	}

	clusterRoles := []rbacv1.ClusterRole{
		{
			// a "root" role which can do absolutely anything
			ObjectMeta: metav1.ObjectMeta{Name: tenantId + "-" + "cluster-admin"},
			Rules: []rbacv1.PolicyRule{
				rbacv1helpers.NewRule("*").Groups("*").Resources("*").RuleOrDie(),
				rbacv1helpers.NewRule("*").URLs("*").RuleOrDie(),
			},
		},
		{
			// a role for a namespace level admin.  It is `edit` plus the power to grant permissions to other users.
			ObjectMeta: metav1.ObjectMeta{Name: tenantId + "-" + "admin"},
			AggregationRule: &rbacv1.AggregationRule{
				ClusterRoleSelectors: []metav1.LabelSelector{
					{MatchLabels: map[string]string{"rbac.authorization.k8s.io/aggregate-to-admin": "true"}},
				},
			},
		},
	}

	for _, clusterRole := range clusterRoles {
		opts := reconciliation.ReconcileRoleOptions{
			Role:    reconciliation.ClusterRoleRuleOwner{ClusterRole: &clusterRole},
			Client:  reconciliation.ClusterRoleModifier{Client: rbacClient.ClusterRoles()},
			Confirm: true,
		}
		err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
			result, err := opts.Run()
			if err != nil {
				return err
			}
			switch {
			case result.Protected && result.Operation != reconciliation.ReconcileNone:
				klog.Warningf("skipped reconcile-protected clusterrole.%s/%s with missing permissions: %v", rbac.GroupName, clusterRole.Name, result.MissingRules)
			case result.Operation == reconciliation.ReconcileUpdate:
				klog.V(2).Infof("updated clusterrole.%s/%s with additional permissions: %v", rbac.GroupName, clusterRole.Name, result.MissingRules)
			case result.Operation == reconciliation.ReconcileCreate:
				klog.V(2).Infof("created clusterrole.%s/%s", rbac.GroupName, clusterRole.Name)
			}
			return nil
		})
		if err != nil {
			// don't fail on failures, try to create as many as you can
			klog.Warningf("unable to reconcile clusterrole.%s/%s: %v", rbac.GroupName, clusterRole.Name, err)
			return err
		}
	}
	return nil
}

// syncClusterRoleBindings synchronize the clusterrolebindings to upstream cluster.
func syncClusterRoleBindings(coreClient typedcorev1.CoreV1Interface, rbacClient typedrbacv1.RbacV1Interface, tenantId string) error {
	if _, err := rbacClient.ClusterRoleBindings().List(context.TODO(), metav1.ListOptions{ResourceVersion: "0"}); err != nil {
		klog.Warningf("Failed to list the clusterrolebindings %s with error %v", tenantId, err)
		return err
	}

	clusterRoleBindings := []rbacv1.ClusterRoleBinding{
		rbacv1helpers.NewClusterBinding(tenantId + "-" + "cluster-admin").Users(tenantId + "-" + "admin").BindingOrDie(),
	}

	for _, clusterRoleBinding := range clusterRoleBindings {
		opts := reconciliation.ReconcileRoleBindingOptions{
			RoleBinding: reconciliation.ClusterRoleBindingAdapter{ClusterRoleBinding: &clusterRoleBinding},
			Client:      reconciliation.ClusterRoleBindingClientAdapter{Client: rbacClient.ClusterRoleBindings()},
			Confirm:     true,
		}
		err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
			result, err := opts.Run()
			if err != nil {
				return err
			}
			switch {
			case result.Protected && result.Operation != reconciliation.ReconcileNone:
				klog.Warningf("skipped reconcile-protected clusterrolebinding.%s/%s with missing subjects: %v", rbac.GroupName, clusterRoleBinding.Name, result.MissingSubjects)
			case result.Operation == reconciliation.ReconcileUpdate:
				klog.V(2).Infof("updated clusterrolebinding.%s/%s with additional subjects: %v", rbac.GroupName, clusterRoleBinding.Name, result.MissingSubjects)
			case result.Operation == reconciliation.ReconcileCreate:
				klog.V(2).Infof("created clusterrolebinding.%s/%s", rbac.GroupName, clusterRoleBinding.Name)
			case result.Operation == reconciliation.ReconcileRecreate:
				klog.V(2).Infof("recreated clusterrolebinding.%s/%s", rbac.GroupName, clusterRoleBinding.Name)
			}
			return nil
		})
		if err != nil {
			// don't fail on failures, try to create as many as you can
			klog.Warningf("unable to reconcile clusterrole.%s/%s: %v", rbac.GroupName, clusterRoleBinding.Name, err)
			return err
		}
	}
	return nil
}

// syncTenantStack syncs the tenant service stack, including coredns.
func syncTenantStack(
	ctx context.Context,
	coreClient typedcorev1.CoreV1Interface,
	appsClient typedappsv1.AppsV1Interface,
	tenant *tenantv1a1.Tenant,
	zooHost string,
	zooPort int,
) error {
	if err := syncCoredns(ctx, coreClient, appsClient, tenant, zooHost, zooPort); err != nil {
		return fmt.Errorf("sync coredns: %w", err)
	}

	return nil
}

// genCertAndKubeconfig signs the certificate/key and generates the kubeconfig for the tenant;
// the generated kubeconfig will be attached in the tenant's annotation.
func genCertAndKubeconfig(
	tenantCli tenantclient.TenantV1alpha1Interface,
	tenantId string,
	tenantlister tenantlister.TenantLister,
	clientCAFile, clientCAKeyFile string,
	kubeZooBindAddress string, kubeZooSecurePort int,
) error {
	tenant, err := tenantlister.Get(tenantId)
	if err != nil {
		return errors.Errorf("Error fetching object with key %s from store: %v", tenantId, err)
	}
	// 1. return if Kubeconfig is already generated
	if len(tenant.Annotations) != 0 && tenant.Annotations[util.AnnotationTenantKubeConfigBase64] != "" {
		return nil
	}

	// 2. Generate the certificate and the key
	cert, key, err := util.NewTenantCertAndKey(clientCAFile, clientCAKeyFile, tenantId)
	if err != nil {
		klog.Warningf("fail to generate the certificate for the tenant(%s): %v", tenantId, err)
		return err
	}

	// 3. Generate the kubeconfig
	caCertByts, err := ioutil.ReadFile(clientCAFile)
	if err != nil {
		klog.Warningf("fail to read CA from file(%s): %v", clientCAFile, err)
		return err
	}

	serverAddress := fmt.Sprintf("https://%s:%d", kubeZooBindAddress, kubeZooSecurePort)
	kbcfgByts, err := util.GenKubeconfig(serverAddress, tenantId, caCertByts, util.EncodePrivateKeyPEM(key), util.EncodeCertPEM(cert))
	if err != nil {
		klog.Warningf("fail to generate the kubeconfig for tenant(%s): %v", tenantId, err)
		return err
	}
	kbcfgB64Str := base64.StdEncoding.EncodeToString(kbcfgByts)

	// 4. Attach the kubeconfig to the annotation
	if tenant.Annotations == nil {
		tenant.Annotations = make(map[string]string)
	}
	tenant.Annotations[util.AnnotationTenantKubeConfigBase64] = kbcfgB64Str
	if _, err := tenantCli.Tenants().Update(context.TODO(), tenant, metav1.UpdateOptions{}); err != nil {
		klog.Warningf("fail to update the tenant with new annotation(%s): %v", util.AnnotationTenantKubeConfigBase64, err)
		return err
	}
	klog.V(4).Infof("kubeconfig of tenant(%s) is created", tenantId)
	return nil
}
