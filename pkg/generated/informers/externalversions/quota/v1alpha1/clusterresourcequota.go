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

// Code generated by informer-gen. DO NOT EDIT.

package v1alpha1

import (
	"context"
	time "time"

	quotav1alpha1 "github.com/kubewharf/kubezoo/pkg/apis/quota/v1alpha1"
	versioned "github.com/kubewharf/kubezoo/pkg/generated/clientset/versioned"
	internalinterfaces "github.com/kubewharf/kubezoo/pkg/generated/informers/externalversions/internalinterfaces"
	v1alpha1 "github.com/kubewharf/kubezoo/pkg/generated/listers/quota/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// ClusterResourceQuotaInformer provides access to a shared informer and lister for
// ClusterResourceQuotas.
type ClusterResourceQuotaInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1alpha1.ClusterResourceQuotaLister
}

type clusterResourceQuotaInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
}

// NewClusterResourceQuotaInformer constructs a new informer for ClusterResourceQuota type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewClusterResourceQuotaInformer(client versioned.Interface, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredClusterResourceQuotaInformer(client, resyncPeriod, indexers, nil)
}

// NewFilteredClusterResourceQuotaInformer constructs a new informer for ClusterResourceQuota type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredClusterResourceQuotaInformer(client versioned.Interface, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.QuotaV1alpha1().ClusterResourceQuotas().List(context.TODO(), options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.QuotaV1alpha1().ClusterResourceQuotas().Watch(context.TODO(), options)
			},
		},
		&quotav1alpha1.ClusterResourceQuota{},
		resyncPeriod,
		indexers,
	)
}

func (f *clusterResourceQuotaInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredClusterResourceQuotaInformer(client, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *clusterResourceQuotaInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&quotav1alpha1.ClusterResourceQuota{}, f.defaultInformer)
}

func (f *clusterResourceQuotaInformer) Lister() v1alpha1.ClusterResourceQuotaLister {
	return v1alpha1.NewClusterResourceQuotaLister(f.Informer().GetIndexer())
}
