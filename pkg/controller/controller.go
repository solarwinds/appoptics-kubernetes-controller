/*
Copyright 2017 The Kubernetes Authors.

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
	"fmt"
	"strings"
	"time"

	clientset "github.com/solarwinds/appoptics-kubernetes-controller/pkg/generated/clientset/versioned"
	aoscheme "github.com/solarwinds/appoptics-kubernetes-controller/pkg/generated/clientset/versioned/scheme"
	informers "github.com/solarwinds/appoptics-kubernetes-controller/pkg/generated/informers/externalversions"
	listers "github.com/solarwinds/appoptics-kubernetes-controller/pkg/generated/listers/appopticskubernetescontroller/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
)

const (
	controllerAgentName = "appoptics"
	AppopticsFinalizer  = "appoptics.io"

	add    addFinalizer = true
	remove addFinalizer = false
)

type addFinalizer bool

const (
	// SuccessSynced is used as part of the Event 'reason' when a Foo is synced
	SuccessSynced = "Synced"
	// ErrResourceExists is used as part of the Event 'reason' when a Foo fails
	// to sync due to a Deployment of the same name already existing.
	ErrResourceExists = "ErrResourceExists"

	// MessageResourceExists is the message used for Events when a resource
	// fails to sync due to a Deployment already existing
	MessageResourceExists = "Resource %q already exists and is not managed by Foo"
	// MessageResourceSynced is the message used for an Event fired when a Foo
	// is synced successfully
	MessageResourceSynced = "Foo synced successfully"
)

// Controller is the controller implementation for Foo resources
type Controller struct {
	// kubeclientset is a standard kubernetes clientset
	kubeclientset kubernetes.Interface
	// aoclientset is a clientset for our own API group
	aoclientset clientset.Interface

	cachesSynced    []cache.InformerSynced
	dashboardLister listers.AppOpticsDashboardLister
	serviceLister   listers.AppOpticsServiceLister
	alertLister     listers.AppOpticsAlertLister
	// workqueue is a rate limited work queue. This is used to queue work to be
	// processed instead of performing it as soon as a change happens. This
	// means we can ensure we only process a fixed amount of resources at a
	// time, and makes it easy to ensure we are never processing the same item
	// simultaneously in two different workers.
	workqueue workqueue.RateLimitingInterface
	// recorder is an event recorder for recording Event resources to the
	// Kubernetes API.
	recorder   record.EventRecorder
	resyncTime int64
}

// NewController returns a new sample controller
func NewController(
	kubeclientset kubernetes.Interface,
	aoclientset clientset.Interface,
	aoInformerFactory informers.SharedInformerFactory,
	controllerAgentName string,
	resyncTime int64) *Controller {

	var cachesSynced []cache.InformerSynced
	dashboardInformer := aoInformerFactory.Appoptics().V1().AppOpticsDashboards()
	cachesSynced = append(cachesSynced, dashboardInformer.Informer().HasSynced)

	serviceInformer := aoInformerFactory.Appoptics().V1().AppOpticsServices()
	cachesSynced = append(cachesSynced, serviceInformer.Informer().HasSynced)

	alertInformer := aoInformerFactory.Appoptics().V1().AppOpticsAlerts()
	cachesSynced = append(cachesSynced, alertInformer.Informer().HasSynced)

	// Create event broadcaster
	// Add sample-controller types to the default Kubernetes Scheme so Events can be
	// logged for sample-controller types.
	utilruntime.Must(aoscheme.AddToScheme(scheme.Scheme))
	klog.V(4).Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartStructuredLogging(0)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeclientset.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})

	controller := &Controller{
		kubeclientset:   kubeclientset,
		aoclientset:     aoclientset,
		cachesSynced:    cachesSynced,
		dashboardLister: dashboardInformer.Lister(),
		serviceLister:   serviceInformer.Lister(),
		alertLister:     alertInformer.Lister(),
		workqueue:       workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "AppOptics"),
		recorder:        recorder,
		resyncTime:      resyncTime,
	}

	klog.Info("Setting up event handlers")
	// we add handlers only for the Dashboards/Services/Alerts! we don't want to control pods and things like that
	// just our resource
	dashboardInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(new interface{}) {
			controller.enqueue(new, Dashboard)
		}, UpdateFunc: func(old, new interface{}) {
			controller.enqueue(new, Dashboard)
		},
	})
	//
	serviceInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(new interface{}) {
			controller.enqueue(new, Service)
		}, UpdateFunc: func(old, new interface{}) {
			controller.enqueue(new, Service)
		},
	})

	alertInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(new interface{}) {
			controller.enqueue(new, Alert)
		},
		UpdateFunc: func(old, new interface{}) {
			controller.enqueue(new, Alert)
		},
	})

	return controller
}

// Run will set up the event handlers for types we are interested in, as well
// as syncing informer caches and starting workers. It will block until stopCh
// is closed, at which point it will shutdown the workqueue and wait for
// workers to finish processing their current work items.
func (c *Controller) Run(workers int, stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer c.workqueue.ShutDown()

	// Start the informer factories to begin populating the informer caches
	klog.Info("Starting AppOptics controller")

	// Wait for the caches to be synced before starting workers
	klog.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, c.cachesSynced...); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	klog.Info("Starting workers")
	// Launch workers
	for i := 0; i < workers; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	klog.Info("Started workers")
	<-stopCh
	klog.Info("Shutting down workers")

	return nil
}

// runWorker is a long-running function that will continually call the
// processNextWorkItem function in order to read and process a message on the
// workqueue.
func (c *Controller) runWorker() {
	for c.processNextWorkItem() {
	}
}

// processNextWorkItem will read a single work item off the workqueue and
// attempt to process it, by calling the syncHandler.
func (c *Controller) processNextWorkItem() bool {
	obj, shutdown := c.workqueue.Get()

	if shutdown {
		return false
	}

	// We wrap this block in a func so we can defer c.workqueue.Done.
	err := func(obj interface{}) error {
		// We call Done here so the workqueue knows we have finished
		// processing this item. We also must remember to call Forget if we
		// do not want this work item being re-queued. For example, we do
		// not call Forget if a transient error occurs, instead the item is
		// put back on the workqueue and attempted again after a back-off
		// period.
		defer c.workqueue.Done(obj)
		var key string
		var ok bool
		// We expect strings to come off the workqueue. These are of the
		// form namespace/name. We do this as the delayed nature of the
		// workqueue means the items in the informer cache may actually be
		// more up to date that when the item was initially put onto the
		// workqueue.
		if key, ok = obj.(string); !ok {
			// As the item in the workqueue is actually invalid, we call
			// Forget here else we'd go into a loop of attempting to
			// process a work item that is invalid.
			c.workqueue.Forget(obj)
			utilruntime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}
		// Run the syncHandler, passing it the namespace/name string of the
		// Foo resource to be synced.
		if err := c.syncHandler(key); err != nil {
			// Put the item back on the workqueue to handle any transient errors.
			c.workqueue.AddRateLimited(key)
			return fmt.Errorf("error syncing '%s': %s, requeuing", key, err.Error())
		}
		// Finally, if no error occurs we Forget this item so it does not
		// get queued again until another change happens.
		c.workqueue.Forget(obj)
		klog.Infof("Successfully synced '%s'", key)
		return nil
	}(obj)

	if err != nil {
		utilruntime.HandleError(err)
		return true
	}

	return true
}

func (c *Controller) enqueue(obj interface{}, kind string) {
	metaObj, err := meta.Accessor(obj)
	if err != nil {
		utilruntime.HandleError(err)
		return
	}

	key := fmt.Sprintf("%s/%s/%s", metaObj.GetNamespace(), kind, metaObj.GetName())

	c.workqueue.AddRateLimited(key)
}

func (c *Controller) SplitMetaNamespaceKey(key string) (namespace, kind string, name string, err error) {
	parts := strings.Split(key, "/")
	switch len(parts) {
	case 1:
		// name only, no namespace
		return "", "", parts[0], nil
	case 3:
		// namespace and name
		return parts[0], parts[1], parts[2], nil
	}

	return "", "", "", fmt.Errorf("unexpected key format: %q", key)
}

func (c *Controller) finalizers(spec *CommonAOResource, isAdd addFinalizer) {
	if isAdd {
		if len(spec.Finalizers) != 1 || spec.Finalizers[0] != AppopticsFinalizer {
			spec.Finalizers = []string{
				AppopticsFinalizer,
			}
		}
	} else {
		if len(spec.Finalizers) != 0 {
			spec.Finalizers = []string{}
		}
	}
}
