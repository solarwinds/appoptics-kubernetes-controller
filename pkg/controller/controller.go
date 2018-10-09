package controller

import (
	"fmt"
	"time"

	"github.com/golang/glog"
	clientset "github.com/solarwinds/appoptics-kubernetes-controller/pkg/client/clientset/versioned"
	"github.com/solarwinds/appoptics-kubernetes-controller/pkg/client/clientset/versioned/scheme"
	aoscheme "github.com/solarwinds/appoptics-kubernetes-controller/pkg/client/clientset/versioned/scheme"
	informers "github.com/solarwinds/appoptics-kubernetes-controller/pkg/client/informers/externalversions"
	listers "github.com/solarwinds/appoptics-kubernetes-controller/pkg/client/listers/appoptics-kubernetes-controller/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"strings"
)

const (
	AppopticsFinalizer = "appoptics.io"

	add    addFinalizer = true
	remove addFinalizer = false
)

type addFinalizer bool

type Controller struct {
	kubeclientset   kubernetes.Interface
	aoclientset     clientset.Interface
	cachesSynced    []cache.InformerSynced
	dashboardLister listers.AppOpticsDashboardLister
	serviceLister   listers.AppOpticsServiceLister
	alertLister     listers.AppOpticsAlertLister
	workqueue       workqueue.RateLimitingInterface
	recorder        record.EventRecorder
	resyncTime      int64
}

// NewController returns a new controller
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

	aoscheme.AddToScheme(scheme.Scheme)

	glog.V(4).Info("Creating event broadcaster")

	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(glog.Infof)
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

	glog.Info("Setting up event handlers")
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

func (c *Controller) Run(threadiness int, stopCh <-chan struct{}) error {
	defer runtime.HandleCrash()
	defer c.workqueue.ShutDown()

	glog.Info("Starting AppOptics controller")

	glog.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, c.cachesSynced...); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	glog.Info("Starting workers")
	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	glog.Info("Started workers")
	<-stopCh
	glog.Info("Shutting down workers")

	return nil
}

func (c *Controller) runWorker() {
	for c.processNextWorkItem() {
	}
}

func (c *Controller) processNextWorkItem() bool {
	obj, shutdown := c.workqueue.Get()

	if shutdown {
		return false
	}

	err := func(obj interface{}) error {
		defer c.workqueue.Done(obj)
		var key string
		var ok bool
		if key, ok = obj.(string); !ok {
			c.workqueue.Forget(obj)
			runtime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}
		if err := c.syncHandler(key); err != nil {
			return fmt.Errorf("error syncing '%s': %s", key, err.Error())
		}
		c.workqueue.Forget(obj)
		glog.Infof("Successfully synced '%s'", key)
		return nil
	}(obj)

	if err != nil {
		runtime.HandleError(err)
		return true
	}

	return true
}

func (c *Controller) enqueue(obj interface{}, kind string) {
	metaObj, err := meta.Accessor(obj)
	if err != nil {
		runtime.HandleError(err)
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
