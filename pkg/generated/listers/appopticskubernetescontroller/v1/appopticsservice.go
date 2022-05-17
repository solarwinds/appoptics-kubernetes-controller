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

// Code generated by lister-gen. DO NOT EDIT.

package v1

import (
	v1 "github.com/solarwinds/appoptics-kubernetes-controller/pkg/apis/appopticskubernetescontroller/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// AppOpticsServiceLister helps list AppOpticsServices.
// All objects returned here must be treated as read-only.
type AppOpticsServiceLister interface {
	// List lists all AppOpticsServices in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1.AppOpticsService, err error)
	// AppOpticsServices returns an object that can list and get AppOpticsServices.
	AppOpticsServices(namespace string) AppOpticsServiceNamespaceLister
	AppOpticsServiceListerExpansion
}

// appOpticsServiceLister implements the AppOpticsServiceLister interface.
type appOpticsServiceLister struct {
	indexer cache.Indexer
}

// NewAppOpticsServiceLister returns a new AppOpticsServiceLister.
func NewAppOpticsServiceLister(indexer cache.Indexer) AppOpticsServiceLister {
	return &appOpticsServiceLister{indexer: indexer}
}

// List lists all AppOpticsServices in the indexer.
func (s *appOpticsServiceLister) List(selector labels.Selector) (ret []*v1.AppOpticsService, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1.AppOpticsService))
	})
	return ret, err
}

// AppOpticsServices returns an object that can list and get AppOpticsServices.
func (s *appOpticsServiceLister) AppOpticsServices(namespace string) AppOpticsServiceNamespaceLister {
	return appOpticsServiceNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// AppOpticsServiceNamespaceLister helps list and get AppOpticsServices.
// All objects returned here must be treated as read-only.
type AppOpticsServiceNamespaceLister interface {
	// List lists all AppOpticsServices in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1.AppOpticsService, err error)
	// Get retrieves the AppOpticsService from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1.AppOpticsService, error)
	AppOpticsServiceNamespaceListerExpansion
}

// appOpticsServiceNamespaceLister implements the AppOpticsServiceNamespaceLister
// interface.
type appOpticsServiceNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all AppOpticsServices in the indexer for a given namespace.
func (s appOpticsServiceNamespaceLister) List(selector labels.Selector) (ret []*v1.AppOpticsService, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1.AppOpticsService))
	})
	return ret, err
}

// Get retrieves the AppOpticsService from the indexer for a given namespace and name.
func (s appOpticsServiceNamespaceLister) Get(name string) (*v1.AppOpticsService, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1.Resource("appopticsservice"), name)
	}
	return obj.(*v1.AppOpticsService), nil
}