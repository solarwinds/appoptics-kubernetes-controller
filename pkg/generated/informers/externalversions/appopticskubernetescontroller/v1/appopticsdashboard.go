/*
Copyright The Kubernetes Authors.

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

package v1

import (
	"context"
	time "time"

	appopticskubernetescontrollerv1 "github.com/solarwinds/appoptics-kubernetes-controller/pkg/apis/appopticskubernetescontroller/v1"
	versioned "github.com/solarwinds/appoptics-kubernetes-controller/pkg/generated/clientset/versioned"
	internalinterfaces "github.com/solarwinds/appoptics-kubernetes-controller/pkg/generated/informers/externalversions/internalinterfaces"
	v1 "github.com/solarwinds/appoptics-kubernetes-controller/pkg/generated/listers/appopticskubernetescontroller/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// AppOpticsDashboardInformer provides access to a shared informer and lister for
// AppOpticsDashboards.
type AppOpticsDashboardInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1.AppOpticsDashboardLister
}

type appOpticsDashboardInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	namespace        string
}

// NewAppOpticsDashboardInformer constructs a new informer for AppOpticsDashboard type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewAppOpticsDashboardInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredAppOpticsDashboardInformer(client, namespace, resyncPeriod, indexers, nil)
}

// NewFilteredAppOpticsDashboardInformer constructs a new informer for AppOpticsDashboard type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredAppOpticsDashboardInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.AppopticsV1().AppOpticsDashboards(namespace).List(context.TODO(), options)
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.AppopticsV1().AppOpticsDashboards(namespace).Watch(context.TODO(), options)
			},
		},
		&appopticskubernetescontrollerv1.AppOpticsDashboard{},
		resyncPeriod,
		indexers,
	)
}

func (f *appOpticsDashboardInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredAppOpticsDashboardInformer(client, f.namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *appOpticsDashboardInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&appopticskubernetescontrollerv1.AppOpticsDashboard{}, f.defaultInformer)
}

func (f *appOpticsDashboardInformer) Lister() v1.AppOpticsDashboardLister {
	return v1.NewAppOpticsDashboardLister(f.Informer().GetIndexer())
}