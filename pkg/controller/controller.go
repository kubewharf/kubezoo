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
	"net"
	"reflect"
	"strconv"
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
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
	rbacclient "k8s.io/client-go/kubernetes/typed/rbac/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/retry"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/component-helpers/auth/rbac/reconciliation"
	"k8s.io/klog"
	"k8s.io/kubernetes/pkg/apis/rbac"
	rbacv1helpers "k8s.io/kubernetes/pkg/apis/rbac/v1"

	quotav1alpha1 "github.com/kubewharf/kubezoo/pkg/apis/quota/v1alpha1"
	"github.com/kubewharf/kubezoo/pkg/common"
	"github.com/kubewharf/kubezoo/pkg/dynamic"
	quotaclient "github.com/kubewharf/kubezoo/pkg/generated/clientset/versioned/typed/quota/v1alpha1"
	tenantclient "github.com/kubewharf/kubezoo/pkg/generated/clientset/versioned/typed/tenant/v1alpha1"
	tenantlister "github.com/kubewharf/kubezoo/pkg/generated/listers/tenant/v1alpha1"
	"github.com/kubewharf/kubezoo/pkg/util"
)

type EventType int

const (
	Create = iota
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
	clusterquotaCli         quotaclient.QuotaV1alpha1Interface
	upstreamDiscoveryClient *discovery.DiscoveryClient
	upstreamDynamicClient   dynamic.Interface
	upstreamCRDClient       *apiextensions.Clientset
	upstreamCoreClient      v1.CoreV1Interface
	upstreamRbacClient      rbacclient.RbacV1Interface
	clientCAFile            string
	clientCAKeyFile         string
	kubeZooHostAddress      string
}

// newTenantController create a controller to handler the events of tenant.
func newTenantController(ti cache.SharedIndexInformer, tenantCli tenantclient.TenantV1alpha1Interface, coreCli v1.CoreV1Interface, rbacCli rbacclient.RbacV1Interface, quotaClient quotaclient.QuotaV1alpha1Interface, discoveryCli *discovery.DiscoveryClient, dynamicCli dynamic.Interface, crdClient *apiextensions.Clientset, clientCAFile, clientCAKeyFile, kubeZooBindAddress string, kubeZooSecurePort int) *TenantController {
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	var (
		newEvent Event
		err      error
	)
	ti.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			newEvent.tenantId, err = cache.MetaNamespaceKeyFunc(obj)
			if err == nil {
				newEvent.eventType = Create
				queue.Add(newEvent)
			}
		},
		UpdateFunc: func(_, new interface{}) {
			newEvent.tenantId, err = cache.MetaNamespaceKeyFunc(new)
			if err == nil {
				newEvent.eventType = Update
				queue.Add(newEvent)
			}
		},
		DeleteFunc: func(obj interface{}) {
			newEvent.tenantId, err = cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err == nil {
				newEvent.eventType = Delete
				queue.Add(newEvent)
			}
		},
	})

	return &TenantController{
		queue:                   queue,
		tenantInformer:          ti,
		tenantLister:            tenantlister.NewTenantLister(ti.GetIndexer()),
		tenantClient:            tenantCli,
		upstreamCoreClient:      coreCli,
		upstreamRbacClient:      rbacCli,
		clusterquotaCli:         quotaClient,
		upstreamDiscoveryClient: discoveryCli,
		upstreamDynamicClient:   dynamicCli,
		upstreamCRDClient:       crdClient,
		clientCAFile:            clientCAFile,
		clientCAKeyFile:         clientCAKeyFile,
		kubeZooHostAddress:      net.JoinHostPort(kubeZooBindAddress, strconv.Itoa(kubeZooSecurePort)),
	}
}

// Run starts the tenant controller
func Run(stopCh <-chan struct{}, ti cache.SharedIndexInformer, tenantCli tenantclient.TenantV1alpha1Interface, typedCli kubernetes.Interface, discoveryCli *discovery.DiscoveryClient, dynamicCli dynamic.Interface, crdClient *apiextensions.Clientset, quotaClient quotaclient.QuotaV1alpha1Interface, clientCAFile, clientCAKeyFile, kubeZooBindAddress string, kubeZooSecurePort int) {
	tc := newTenantController(ti, tenantCli, typedCli.CoreV1(), typedCli.RbacV1(), quotaClient, discoveryCli, dynamicCli, crdClient, clientCAFile, clientCAKeyFile, kubeZooBindAddress, kubeZooSecurePort)
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
	newEvent, quit := tc.queue.Get()

	if quit {
		return false
	}
	defer tc.queue.Done(newEvent)
	err := tc.processItem(newEvent.(Event))
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
func (tc *TenantController) processItem(e Event) error {
	// process events based on its type
	switch e.eventType {
	case Create:
		return tc.onTeanntCreate(e.tenantId)
	case Update:
		return tc.onTeanntUpdate(e.tenantId)
	case Delete:
		klog.Warningf("deleting tenant %v", e.tenantId)
		return nil
	}
	return nil
}

func (tc *TenantController) onTeanntCreate(tenantID string) error {
	if err := tc.onTenantAddOrUpdate(tenantID); err != nil {
		return err
	}

	if err := tc.syncResources(tenantID); err != nil {
		return err
	}

	if err := tc.syncClusterResourceQuota(tenantID); err != nil {
		return err
	}
	return nil
}

func (tc *TenantController) onTeanntUpdate(tenantID string) error {
	if err := tc.onTenantAddOrUpdate(tenantID); err != nil {
		return err
	}

	if err := tc.syncClusterResourceQuota(tenantID); err != nil {
		return err
	}
	return nil
}

// onTenantAddOrUpdate handles the Create or UPDATE event of a Tenant.
func (tc *TenantController) onTenantAddOrUpdate(tenantId string) error {
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

	return nil
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

	if err := tc.syncClusterResourceQuota(tenantId); err != nil {
		return errors.Errorf("fail to delete clusterResourceQuota for tenant %s: %v", tenantId, err)
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
func (tc *TenantController) syncResources(tenantId string) error {
	if tc.tenantClient == nil || tc.upstreamCoreClient == nil || tc.upstreamRbacClient == nil {
		return errors.New("Skip synchronize namespaces or RBAC resources since nil client.")
	}

	klog.V(4).Infof("Sync system resources for tenant %s", tenantId)
	if err := syncNamespaces(tc.upstreamCoreClient, tenantId); err != nil {
		return err
	}
	if err := syncClusterRoles(tc.upstreamCoreClient, tc.upstreamRbacClient, tenantId); err != nil {
		return err
	}
	if err := syncClusterRoleBindings(tc.upstreamCoreClient, tc.upstreamRbacClient, tenantId); err != nil {
		return err
	}
	if err := genCertAndKubeconfig(tc.tenantClient, tenantId, tc.tenantLister, tc.clientCAFile, tc.clientCAKeyFile, tc.kubeZooHostAddress); err != nil {
		return err
	}
	return nil
}

func (tc *TenantController) syncClusterResourceQuota(tenantID string) error {
	if tc.tenantClient == nil || tc.clusterquotaCli == nil {
		klog.Warning("Skip synchronize cluster resource quota since nil tenant or clusterResourceQuota client.")
		return nil
	}
	tenantQuotaName := fmt.Sprintf("%s-%s", common.TenantQuotaNamePrefix, tenantID)

	tenant, err := tc.tenantClient.Tenants().Get(context.TODO(), tenantID, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			nerr := tc.clusterquotaCli.ClusterResourceQuotas().Delete(context.TODO(), tenantQuotaName, metav1.DeleteOptions{})
			if apierrors.IsNotFound(nerr) {
				// ignore notFound
				nerr = nil
			}
			if nerr == nil {
				klog.Infof("delete cluster resource quota (%v) successfully", tenantQuotaName)
			}
			return nerr
		}
		return err
	}

	if tenant.DeletionTimestamp != nil {
		// skip sync, wait for cleanup
		return nil
	}

	clusterquota, err := tc.clusterquotaCli.ClusterResourceQuotas().Get(context.TODO(), tenantQuotaName, metav1.GetOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return err
	}
	expectedQuota := &quotav1alpha1.ClusterResourceQuota{
		ObjectMeta: metav1.ObjectMeta{
			Name: tenantQuotaName,
		},
		Spec: quotav1alpha1.ClusterResourceQuotaSpec{
			ResourceQuotaSpec: corev1.ResourceQuotaSpec{
				Hard: tenant.Spec.Quota.Hard,
			},
			NamepsaceSelector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					common.TenantNamespaceLabelKey: tenantID,
				},
			},
		},
	}
	if apierrors.IsNotFound(err) {
		// create
		_, err := tc.clusterquotaCli.ClusterResourceQuotas().Create(context.TODO(), expectedQuota, metav1.CreateOptions{})
		return err
	}

	// update
	if !reflect.DeepEqual(clusterquota.Spec.Hard, tenant.Spec.Quota.Hard) ||
		!reflect.DeepEqual(clusterquota.Spec.NamepsaceSelector, expectedQuota.Spec.NamepsaceSelector) {
		mutator := func(quota *quotav1alpha1.ClusterResourceQuota) error {
			quota.Spec.Hard = tenant.Spec.Quota.Hard
			quota.Spec.NamepsaceSelector = expectedQuota.Spec.NamepsaceSelector
			return nil
		}
		//TODO: retry
		return retry.RetryOnConflict(wait.Backoff{
			Steps:    20,
			Duration: 500 * time.Millisecond,
			Factor:   1.0,
			Jitter:   0.1},
			func() error {
				quota, err := tc.clusterquotaCli.ClusterResourceQuotas().Get(context.TODO(), tenantQuotaName, metav1.GetOptions{})
				if err != nil {
					return err
				}
				if err = mutator(quota); err != nil {
					return errors.Wrap(err, "unable to mutate ClusterResourceQuotas")
				}
				_, err = tc.clusterquotaCli.ClusterResourceQuotas().Update(context.TODO(), quota, metav1.UpdateOptions{})
				return err
			},
		)
	}

	// do nothing
	return nil
}

// syncNamespaces synchronize the system namespaces to upstream cluster.
func syncNamespaces(coreClient v1.CoreV1Interface, tenantId string) error {
	systemNamespaces := []string{metav1.NamespaceSystem, metav1.NamespacePublic, corev1.NamespaceNodeLease, corev1.NamespaceDefault}

	for _, systemNamespace := range systemNamespaces {
		expectNamespace := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name:      tenantId + "-" + systemNamespace,
				Namespace: "",
				Labels: map[string]string{
					common.TenantNamespaceLabelKey: tenantId,
				},
			},
		}
		ns, err := coreClient.Namespaces().Get(context.TODO(), expectNamespace.Name, metav1.GetOptions{})
		if err != nil {
			_, err = coreClient.Namespaces().Create(context.TODO(), expectNamespace, metav1.CreateOptions{})
			if err == nil {
				continue
			}
			if !apierrors.IsAlreadyExists(err) {
				klog.Warningf("Failed to create the tenant namespace %s with error %v", expectNamespace.Name, err)
				return err
			}
			// check namespaces's labels if it already exists
			ns, err = coreClient.Namespaces().Get(context.TODO(), expectNamespace.Name, metav1.GetOptions{})
			if err != nil {
				klog.Warningf("Failed to get the tenant namespace %s with error %v", expectNamespace.Name, err)
				return err
			}
		}

		if ns.Labels == nil {
			ns.Labels = make(map[string]string)
		}
		if ns.Labels[common.TenantNamespaceLabelKey] != tenantId {
			ns.Labels[common.TenantNamespaceLabelKey] = tenantId
			// TODO: retry on conflict
			_, err := coreClient.Namespaces().Update(context.TODO(), ns, metav1.UpdateOptions{})
			if err != nil {
				klog.Warningf("Failed to update the tenant namespace %s with error %v", expectNamespace.Name, err)
				return err
			}
		}
	}
	return nil
}

// syncClusterRoles synchronize the cluster roles to upstream cluster.
func syncClusterRoles(coreClient v1.CoreV1Interface, rbacClient rbacclient.RbacV1Interface, tenantId string) error {
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
func syncClusterRoleBindings(coreClient v1.CoreV1Interface, rbacClient rbacclient.RbacV1Interface, tenantId string) error {
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

// genCertAndKubeconfig signs the certificate/key and generates the kubeconfig for the tenant;
// the generated kubeconfig will be attached in the tenant's annotation.
func genCertAndKubeconfig(tenantCli tenantclient.TenantV1alpha1Interface, tenantId string, tenantlister tenantlister.TenantLister, clientCAFile, clientCAKeyFile, kubeZooHostAddress string) error {
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
	kbcfgByts, err := util.GenKubeconfig("https://"+kubeZooHostAddress, tenantId, caCertByts, util.EncodePrivateKeyPEM(key), util.EncodeCertPEM(cert))
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
