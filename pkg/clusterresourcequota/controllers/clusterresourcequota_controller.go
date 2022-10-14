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

package controllers

import (
	"context"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	quotautil "k8s.io/apiserver/pkg/quota/v1"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	quotav1alpha1 "github.com/kubewharf/kubezoo/pkg/apis/quota/v1alpha1"
)

var (
	ClusterResourceQuotaKind = "ClusterResourceQuota"

	LabelClusterResourceQuotaAutoUpdate = "clusterresourcequota.quota.kubezoo.io/autoupdate"
)

// ClusterResourceQuotaReconciler reconciles a ClusterResourceQuota object
type ClusterResourceQuotaReconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	Cache     cache.Cache
	APIReader client.Reader
	Logger    logr.Logger
}

// SetupWithManager sets up the controller with the Manager.
func (r *ClusterResourceQuotaReconciler) SetupWithManager(mgr ctrl.Manager) error {
	err := ctrl.NewControllerManagedBy(mgr).
		For(&quotav1alpha1.ClusterResourceQuota{}).
		Owns(&corev1.ResourceQuota{}).
		Watches(&source.Kind{Type: &corev1.Namespace{}}, &namesapceForClusterResourceQuotaHandler{
			cache:  r.Cache,
			logger: r.Logger,
		}).
		WithOptions(controller.Options{MaxConcurrentReconciles: 10}).
		Complete(r)
	if err != nil {
		return err
	}

	server := mgr.GetWebhookServer()
	a := NewAdmission(context.TODO(), r.Client)
	server.Register("/admission/validating/clusterresourcequotas", &admission.Webhook{Handler: a})
	return nil
}

//+kubebuilder:rbac:groups=quota.kubezoo.io,resources=clusterresourcequota,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=quota.kubezoo.io,resources=clusterresourcequota/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=quota.kubezoo.io,resources=clusterresourcequota/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ClusterResourceQuota object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *ClusterResourceQuotaReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var clusterquota quotav1alpha1.ClusterResourceQuota
	err := r.Client.Get(ctx, req.NamespacedName, &clusterquota)
	if err != nil {
		return reconcile.Result{}, client.IgnoreNotFound(err)
	}

	namespaces, err := GetNamespacesForClusterQuota(ctx, r.Client, &clusterquota)
	if err != nil {
		return reconcile.Result{}, err
	}

	quotaNamespaces, quotas, err := GetResourceQuotaForClusterQuota(ctx, r.Client, &clusterquota)
	if err != nil {
		return reconcile.Result{}, err
	}

	expectedQuotas := []*corev1.ResourceQuota{}

	for _, ns := range namespaces.List() {
		quota, err := r.ensureResourceQuotaInNamespace(ctx, &clusterquota, ns, quotas[ns])
		if err != nil {
			return reconcile.Result{}, err
		}
		expectedQuotas = append(expectedQuotas, quota)
		// delete synced namespace from quota namesapces
		quotaNamespaces.Delete(ns)
	}
	for _, ns := range quotaNamespaces.List() {
		// delete resource quotas in unmatched namespaces
		for _, quota := range quotas[ns] {
			r.Logger.Info("delete resource quota from unmatched namespace", "resourceQuota", client.ObjectKeyFromObject(quota).String(), "clusterResourceQuota", clusterquota.Name)
			r.Client.Delete(ctx, quota) //nolint
		}
	}

	if err := r.syncStatus(ctx, &clusterquota, expectedQuotas); err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (r *ClusterResourceQuotaReconciler) syncStatus(ctx context.Context, clusterquota *quotav1alpha1.ClusterResourceQuota, quotas []*corev1.ResourceQuota) error {
	statusLimitsDirty := !apiequality.Semantic.DeepEqual(clusterquota.Spec.Hard, clusterquota.Status.Hard)
	dirty := statusLimitsDirty || clusterquota.Status.Hard == nil || clusterquota.Status.Used == nil

	hardLimits := quotautil.Add(corev1.ResourceList{}, clusterquota.Spec.Hard)

	newUsage := corev1.ResourceList{}
	for _, quota := range quotas {
		newUsage = quotautil.Add(newUsage, quota.Status.Used)
	}

	usage := quotav1alpha1.ClusterResourceQuotaStatus{
		ResourceQuotaStatus: corev1.ResourceQuotaStatus{
			Hard: hardLimits,
			Used: newUsage,
		},
	}

	dirty = dirty || !quotautil.Equals(clusterquota.Status.Used, usage.Used)

	if dirty {
		r.Logger.Info("sync status of cluster resource quota", "clusterResourceQuota", clusterquota.Name, "from", clusterquota.Status, "to", usage)
		_, err := UpdateOnConflict(ctx, DefaultRetry, r.APIReader, r.Client.Status(), clusterquota, func() error {
			clusterquota.Status = usage
			return nil
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *ClusterResourceQuotaReconciler) ensureResourceQuotaInNamespace(ctx context.Context, clusterquota *quotav1alpha1.ClusterResourceQuota, namespace string, quotas []*corev1.ResourceQuota) (*corev1.ResourceQuota, error) {
	var matchedquota *corev1.ResourceQuota

	ownerDirty := false

	for i := range quotas {
		quota := quotas[i]
		rqKey := client.ObjectKeyFromObject(quota).String()
		owner := metav1.GetControllerOf(quota)
		if owner != nil && owner.Kind != ClusterResourceQuotaKind {
			// the quota is not controlled by cluster resource quota
			// remove label from resource quota
			delete(quota.Labels, quotav1alpha1.ClusterResourceQuotaCreatedby)
			r.Logger.Info("resource quota is not controlled by ClusterResourceQuota, delete createdBy label from it", "resourceQuota", rqKey)
			r.Client.Update(ctx, quota) //nolint
			continue
		}
		if owner == nil {
			// need to add owner
			ownerDirty = true
		}
		if matchedquota == nil {
			matchedquota = quota
			continue
		}
		r.Logger.Info("delete duplicated resource quota in namespace", "resourceQuota", rqKey, "clusterResourceQuota", clusterquota.Name)
		// delete duplicated quota
		r.Client.Delete(ctx, quota) //nolint
	}

	if matchedquota == nil {
		// create
		owner := metav1.NewControllerRef(clusterquota, quotav1alpha1.SchemeGroupVersion.WithKind(ClusterResourceQuotaKind))
		quota := &corev1.ResourceQuota{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    namespace,
				GenerateName: clusterquota.Name + "-",
				Labels: map[string]string{
					quotav1alpha1.ClusterResourceQuotaCreatedby: clusterquota.Name,
					LabelClusterResourceQuotaAutoUpdate:         "true",
				},
				OwnerReferences: []metav1.OwnerReference{*owner},
			},
			Spec: clusterquota.Spec.ResourceQuotaSpec,
		}
		r.Logger.Info("create resource quota in namespace", "namesapce", namespace, "generateName", quota.GenerateName, "clusterResourceQuota", clusterquota.Name)
		return quota, r.Client.Create(ctx, quota)
	}

	specDirty := !apiequality.Semantic.DeepEqual(clusterquota.Spec.ResourceQuotaSpec, matchedquota.Spec)

	// update owner or spec
	if ownerDirty || specDirty {
		r.Logger.Info("sync resource quota in namespace", "resourceQuota", client.ObjectKeyFromObject(matchedquota).String(), "clusterResourceQuota", clusterquota.Name)
		_, err := UpdateOnConflict(ctx, DefaultRetry, r.APIReader, r.Client, matchedquota, func() error {
			if ownerDirty {
				owner := metav1.NewControllerRef(clusterquota, quotav1alpha1.SchemeGroupVersion.WithKind(ClusterResourceQuotaKind))
				matchedquota.OwnerReferences = append(matchedquota.OwnerReferences, *owner)
			}
			if specDirty {
				matchedquota.Spec = clusterquota.Spec.ResourceQuotaSpec
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	return matchedquota, nil
}

func GetResourceQuotaForClusterQuota(ctx context.Context, c client.Client, clusterquota *quotav1alpha1.ClusterResourceQuota) (sets.String, map[string][]*corev1.ResourceQuota, error) {
	selector := labels.Set(map[string]string{
		quotav1alpha1.ClusterResourceQuotaCreatedby: clusterquota.Name,
	})

	// find already exists quota
	var quotaList corev1.ResourceQuotaList
	err := c.List(ctx, &quotaList, &client.ListOptions{LabelSelector: selector.AsSelector()})
	if err != nil {
		return nil, nil, err
	}

	namespaces := sets.NewString()
	namesapceToQuota := make(map[string][]*corev1.ResourceQuota)

	for i := range quotaList.Items {
		quota := &quotaList.Items[i]
		namesapceToQuota[quota.Namespace] = append(namesapceToQuota[quota.Namespace], quota)
		namespaces.Insert(quota.Namespace)
	}

	return namespaces, namesapceToQuota, err
}

func GetNamespacesForClusterQuota(ctx context.Context, c client.Client, clusterquota *quotav1alpha1.ClusterResourceQuota) (sets.String, error) {
	namespaces := sets.NewString()
	if clusterquota.Spec.NamepsaceSelector != nil {
		var nsList corev1.NamespaceList
		selector, err := metav1.LabelSelectorAsSelector(clusterquota.Spec.NamepsaceSelector)
		if err != nil {
			return nil, err
		}
		err = c.List(ctx, &nsList, &client.ListOptions{LabelSelector: selector})
		if err != nil {
			return nil, err
		}

		for _, ns := range nsList.Items {
			namespaces.Insert(ns.Name)
		}
	}

	for _, name := range clusterquota.Spec.Namespaces {
		var ns corev1.Namespace
		err := c.Get(ctx, types.NamespacedName{Name: name}, &ns)
		if err != nil {
			if errors.IsNotFound(err) {
				continue
			}
			return nil, err
		}
		namespaces.Insert(name)
	}
	return namespaces, nil
}

var _ handler.EventHandler = &namesapceForClusterResourceQuotaHandler{}

type namesapceForClusterResourceQuotaHandler struct {
	cache  cache.Cache
	logger logr.Logger
}

func (h *namesapceForClusterResourceQuotaHandler) Create(e event.CreateEvent, q workqueue.RateLimitingInterface) {
	ns, ok := e.Object.(*corev1.Namespace)
	if !ok {
		return
	}

	var quotaList quotav1alpha1.ClusterResourceQuotaList
	err := h.cache.List(context.TODO(), &quotaList)
	if err != nil {
		h.logger.Error(err, "failed to list cluster resource quota")
		return
	}

	for i := range quotaList.Items {
		quota := quotaList.Items[i]

		if matches(&quota, ns) {
			// enqueue quota
			q.Add(reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name: quota.Name,
				},
			})
		}
	}
}

func (h *namesapceForClusterResourceQuotaHandler) Update(event.UpdateEvent, workqueue.RateLimitingInterface) {
}

func (h *namesapceForClusterResourceQuotaHandler) Delete(event.DeleteEvent, workqueue.RateLimitingInterface) {
}

func (h *namesapceForClusterResourceQuotaHandler) Generic(event.GenericEvent, workqueue.RateLimitingInterface) {
}

func matches(quota *quotav1alpha1.ClusterResourceQuota, ns *corev1.Namespace) bool {
	selector, err := metav1.LabelSelectorAsSelector(quota.Spec.NamepsaceSelector)
	if err != nil {
		return false
	}
	set := sets.NewString(quota.Spec.Namespaces...)
	return selector.Matches(labels.Set(ns.Labels)) || set.Has(ns.Name)
}
