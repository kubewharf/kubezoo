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
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
	rbacclient "k8s.io/client-go/kubernetes/typed/rbac/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/retry"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"
	"k8s.io/kubernetes/pkg/apis/rbac"
	rbacv1helpers "k8s.io/kubernetes/pkg/apis/rbac/v1"
	"k8s.io/kubernetes/pkg/registry/rbac/reconciliation"

	tenantclient "github.com/kubewharf/kubezoo/pkg/generated/clientset/versioned/typed/tenant/v1alpha1"
	tenantlister "github.com/kubewharf/kubezoo/pkg/generated/listers/tenant/v1alpha1"
	"github.com/kubewharf/kubezoo/pkg/util"
)

type EventType int

const (
	Create = iota
	Delete
)

const maxRetries = 10

// Event indicate the informerEvent
type Event struct {
	tenantId  string
	eventType EventType
}

// TenantController take responsibility for tenant management including
// the tenant object, tenant certificate and the tenant's k8s resources.
type TenantController struct {
	queue              workqueue.RateLimitingInterface
	tenantInformer     cache.SharedIndexInformer
	tenantLister       tenantlister.TenantLister
	tenantClient       tenantclient.TenantV1alpha1Interface
	upstreamCoreClient v1.CoreV1Interface
	upstreamRbacClient rbacclient.RbacV1Interface
	clientCAFile       string
	clientCAKeyFile    string
	kubeZooHostAddress string
}

// newTenantController create a controller to handler the events of tenant.
func newTenantController(ti cache.SharedIndexInformer, tenantCli tenantclient.TenantV1alpha1Interface, coreCli v1.CoreV1Interface, rbacCli rbacclient.RbacV1Interface, clientCAFile, clientCAKeyFile, kubeZooBindAddress string, kubeZooSecurePort int) *TenantController {
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
		DeleteFunc: func(obj interface{}) {
			newEvent.tenantId, err = cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err == nil {
				newEvent.eventType = Delete
				queue.Add(newEvent)
			}
		},
	})

	return &TenantController{
		queue:              queue,
		tenantInformer:     ti,
		tenantLister:       tenantlister.NewTenantLister(ti.GetIndexer()),
		tenantClient:       tenantCli,
		upstreamCoreClient: coreCli,
		upstreamRbacClient: rbacCli,
		clientCAFile:       clientCAFile,
		clientCAKeyFile:    clientCAKeyFile,
		kubeZooHostAddress: net.JoinHostPort(kubeZooBindAddress, strconv.Itoa(kubeZooSecurePort)),
	}
}

// Run starts the tenant controller
func Run(stopCh <-chan struct{}, ti cache.SharedIndexInformer, tenantCli tenantclient.TenantV1alpha1Interface, typedCli kubernetes.Interface, clientCAFile, clientCAKeyFile, kubeZooBindAddress string, kubeZooSecurePort int) {
	tc := newTenantController(ti, tenantCli, typedCli.CoreV1(), typedCli.RbacV1(), clientCAFile, clientCAKeyFile, kubeZooBindAddress, kubeZooSecurePort)
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
		return tc.syncResources(e.tenantId)
	case Delete:
		klog.Warningf("deleting tenant %v", e.tenantId)
		return tc.deleteResources(e.tenantId)
	}
	return nil
}

// deleteResources deletes resources belonging to the tenant from the upstream cluster.
// TODO delete all cluster-scoped resources belonging to the tenant.
func (tc *TenantController) deleteResources(tenantId string) error {
	if tc.upstreamCoreClient == nil || tc.upstreamRbacClient == nil {
		return errors.New("Can't delete namespaces or RBAC resources with nil client.")
	}
	klog.V(4).Infof("delete resources for tenant %s", tenantId)
	if err := deleteNamespaces(tc.upstreamCoreClient, tenantId); err != nil {
		return errors.Errorf("fail to delete namespaces for tenant %s: %v", tenantId, err)
	}

	if err := deleteClusterRoles(tc.upstreamRbacClient, tenantId); err != nil {
		return errors.Errorf("fail to delete clusterroles for tenant %s: %v", tenantId, err)
	}

	if err := deleteClusterRoleBindings(tc.upstreamRbacClient, tenantId); err != nil {
		return errors.Errorf("fail to delete clusterrolebindings for tenant %s: %v", tenantId, err)
	}
	klog.Infof("deleted resources for tenant %s", tenantId)
	return nil
}

// deleteNamespaces deletes namespaces belonging to the tenant.
func deleteNamespaces(coreClient v1.CoreV1Interface, tenantId string) error {
	// get all namespaces belonging to the tenant
	nsList, err := coreClient.Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return errors.Wrap(err, "fail to list namespaces")
	}

	tenantPrefix := tenantId + "-"
	for _, ns := range nsList.Items {
		if strings.HasPrefix(ns.Name, tenantPrefix) {
			if err := coreClient.Namespaces().Delete(context.TODO(), ns.Name, metav1.DeleteOptions{}); err != nil {
				return errors.Wrap(err, "fail to delete the namespace")
			}
		}
	}

	return nil
}

// deleteClusterRoles deletes clusterroles belonging to the tenant.
func deleteClusterRoles(rbacClient rbacclient.RbacV1Interface, tenantId string) error {
	// get all clusterroles belonging to the tenant
	crList, err := rbacClient.ClusterRoles().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return errors.Wrap(err, "fail to list clusterroles")
	}

	tenantPrefix := tenantId + "-"
	for _, cr := range crList.Items {
		if strings.HasPrefix(cr.Name, tenantPrefix) {
			if err := rbacClient.ClusterRoles().Delete(context.TODO(), cr.Name, metav1.DeleteOptions{}); err != nil {
				return errors.Wrap(err, "fail to delete the clusterrole")
			}
		}
	}

	return nil
}

// deleteClusterRoleBindings deletes clusterrolebindings belonging to the tenant.
func deleteClusterRoleBindings(rbacClient rbacclient.RbacV1Interface, tenantId string) error {
	// get all clusterrolebindings belonging to the tenant
	crbList, err := rbacClient.ClusterRoleBindings().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return errors.Wrap(err, "fail to list clusterrolebindings")
	}

	tenantPrefix := tenantId + "-"
	for _, crb := range crbList.Items {
		if strings.HasPrefix(crb.Name, tenantPrefix) {
			if err := rbacClient.ClusterRoleBindings().Delete(context.TODO(), crb.Name, metav1.DeleteOptions{}); err != nil {
				return errors.Wrap(err, "fail to delete the clusterrolebinding")
			}
		}
	}

	return nil
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

// syncNamespaces synchronize the system namespaces to upstream cluster.
func syncNamespaces(coreClient v1.CoreV1Interface, tenantId string) error {
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
