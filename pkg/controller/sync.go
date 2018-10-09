package controller

import (
	"fmt"
	"time"

	"github.com/golang/glog"
	v12 "github.com/solarwinds/appoptics-kubernetes-controller/pkg/apis/appoptics-kubernetes-controller/v1"
	"github.com/solarwinds/appoptics-kubernetes-controller/pkg/controller/appoptics"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/runtime"
)

const (
	// DateFormat is used to format dates, by default it will be RFC1123Z
	DateFormat = time.RFC1123Z

	// ErrUpdateStatus is used as part of the Event 'reason' when we fail to update the status
	ErrUpdateStatus = "ErrUpdateStatus"

	// SuccessUpdate is used as part of the Event 'reason' when we update the status successfully
	SuccessUpdate = "SuccessUpdate"

	// MessageResourceUpdated is the message used for Events when a resource is updated
	MessageResourceUpdated = "Updated resource %s"

	Dashboard = "Dashboard"
	Alert     = "Alert"
	Service   = "Service"
)

type CommonAOResource struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              v12.TokenAndDataSpec `json:"spec"`
	Status            v12.Status           `json:"status,omitempty"`
}

func (c *Controller) syncHandler(key string) error {

	namespace, kind, name, err := c.SplitMetaNamespaceKey(key)
	if err != nil {
		runtime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}
	currentTime := time.Now()

	switch kind {
	case Dashboard:
		dashboard, err := c.dashboardLister.AppOpticsDashboards(namespace).Get(name)
		if err != nil {
			if errors.IsNotFound(err) {
				runtime.HandleError(fmt.Errorf("%s '%s' in work queue no longer exists", kind, key))
				return nil
			}

			return err
		}

		if len(dashboard.Status.LastUpdated) > 0 {
			lastUpdated, err := time.Parse(DateFormat, dashboard.Status.LastUpdated)
			if err != nil {
				glog.Warningf("Error, date %s not in RFC1123Z format")
			} else {
				if currentTime.Unix()-lastUpdated.Unix() < c.resyncTime {
					return nil
				}
			}
		}

		// NEVER modify objects from the store. It's a read-only, local cache.
		updateStatus := dashboard.Status.DeepCopy()
		updateStatus.LastUpdated = currentTime.Format(DateFormat)

		aoResource := CommonAOResource(*dashboard)

		secret, err := c.kubeclientset.CoreV1().Secrets(namespace).Get(aoResource.Spec.Secret, metav1.GetOptions{})
		if err != nil {
			return err
		}

		aoc, err := c.GetCommunicator(secret)
		if err != nil {
			return err
		}

		if aoResource.DeletionTimestamp != nil {
			err = aoc.Remove(aoResource.Status.ID, kind)
			if err != nil {
				return err
			}
			c.finalizers(&aoResource, remove)

			uDashboard := v12.AppOpticsDashboard(aoResource)
			_, err := c.aoclientset.AppopticsV1().AppOpticsDashboards(namespace).Update(&uDashboard)
			if err != nil {
				return err
			}
			return nil
		}

		c.finalizers(&aoResource, add)
		updateStatus, err = aoc.Sync(aoResource.Spec, updateStatus, kind, nil)
		if err != nil {
			return err
		}

		udashboard := v12.AppOpticsDashboard(aoResource)
		dashboardCopy := udashboard.DeepCopy()

		dashboardCopy.Status = *updateStatus

		_, err = c.aoclientset.AppopticsV1().AppOpticsDashboards(dashboard.Namespace).Update(dashboardCopy)
		if err != nil {
			c.recorder.Event(dashboard, v1.EventTypeWarning, ErrUpdateStatus, err.Error())
		} else {
			c.recorder.Event(dashboardCopy, v1.EventTypeNormal, SuccessUpdate, MessageResourceUpdated)
		}
	case Service:
		service, err := c.serviceLister.AppOpticsServices(namespace).Get(name)
		if err != nil {
			if errors.IsNotFound(err) {
				runtime.HandleError(fmt.Errorf("%s '%s' in work queue no longer exists", kind, key))
				return nil
			}

			return err
		}

		if len(service.Status.LastUpdated) > 0 {
			lastUpdated, err := time.Parse(DateFormat, service.Status.LastUpdated)
			if err != nil {
				glog.Warningf("Error, date %s not in RFC1123Z format")
			} else {
				if currentTime.Unix()-lastUpdated.Unix() < c.resyncTime {
					return nil
				}
			}
		}
		// NEVER modify objects from the store. It's a read-only, local cache.
		updateStatus := service.Status.DeepCopy()
		updateStatus.LastUpdated = currentTime.Format(DateFormat)
		aoResource := CommonAOResource(*service)

		secret, err := c.kubeclientset.CoreV1().Secrets(namespace).Get(aoResource.Spec.Secret, metav1.GetOptions{})
		if err != nil {
			return err
		}

		aoc, err := c.GetCommunicator(secret)
		if err != nil {
			return err
		}

		if aoResource.DeletionTimestamp != nil {
			err = aoc.Remove(aoResource.Status.ID, kind)
			if err != nil {
				return err
			}
			c.finalizers(&aoResource, remove)

			uService := v12.AppOpticsService(aoResource)
			_, err := c.aoclientset.AppopticsV1().AppOpticsServices(namespace).Update(&uService)
			if err != nil {
				return err
			}
			return nil
		}

		c.finalizers(&aoResource, add)
		updateStatus, err = aoc.Sync(aoResource.Spec, updateStatus, kind, nil)
		if err != nil {
			return err
		}

		uService := v12.AppOpticsService(aoResource)
		serviceCopy := uService.DeepCopy()
		serviceCopy.Status = *updateStatus

		_, err = c.aoclientset.AppopticsV1().AppOpticsServices(service.Namespace).Update(serviceCopy)
		if err != nil {
			c.recorder.Event(service, v1.EventTypeWarning, ErrUpdateStatus, err.Error())
		} else {
			c.recorder.Event(serviceCopy, v1.EventTypeNormal, SuccessUpdate, MessageResourceUpdated)
		}
	case Alert:
		alert, err := c.alertLister.AppOpticsAlerts(namespace).Get(name)
		if err != nil {
			if errors.IsNotFound(err) {
				runtime.HandleError(fmt.Errorf("%s '%s' in work queue no longer exists", kind, key))
				return nil
			}

			return err
		}

		if len(alert.Status.LastUpdated) > 0 {
			lastUpdated, err := time.Parse(DateFormat, alert.Status.LastUpdated)
			if err != nil {
				glog.Warningf("Error, date %s not in RFC1123Z format")
			} else {
				if currentTime.Unix()-lastUpdated.Unix() < c.resyncTime {
					return nil
				}
			}
		}

		// NEVER modify objects from the store. It's a read-only, local cache.
		updateStatus := alert.Status.DeepCopy()
		updateStatus.LastUpdated = currentTime.Format(DateFormat)
		aoResource := CommonAOResource(*alert)

		secret, err := c.kubeclientset.CoreV1().Secrets(namespace).Get(aoResource.Spec.Secret, metav1.GetOptions{})
		if err != nil {
			return err
		}

		aoc, err := c.GetCommunicator(secret)
		if err != nil {
			return err
		}

		if aoResource.DeletionTimestamp != nil {
			err = aoc.Remove(aoResource.Status.ID, kind)
			if err != nil {
				return err
			}
			c.finalizers(&aoResource, remove)

			uAlerts := v12.AppOpticsAlert(aoResource)
			_, err := c.aoclientset.AppopticsV1().AppOpticsAlerts(namespace).Update(&uAlerts)
			if err != nil {
				return err
			}
			return nil
		}

		c.finalizers(&aoResource, add)
		updateStatus, err = aoc.Sync(aoResource.Spec, updateStatus, kind, c.serviceLister.AppOpticsServices(namespace))
		if err != nil {
			return err
		}

		uAlert := v12.AppOpticsAlert(aoResource)
		alertCopy := uAlert.DeepCopy()

		alertCopy.Status = *updateStatus

		_, err = c.aoclientset.AppopticsV1().AppOpticsAlerts(alert.Namespace).Update(alertCopy)
		if err != nil {
			c.recorder.Event(alert, v1.EventTypeWarning, ErrUpdateStatus, err.Error())
		} else {
			c.recorder.Event(alertCopy, v1.EventTypeNormal, SuccessUpdate, MessageResourceUpdated)
		}
	}

	return nil
}

func (c *Controller) GetCommunicator(secret *v1.Secret) (appoptics.AOCommunicator, error) {
	aoClientToken := ""
	if token, ok := secret.Data["token"]; ok {
		aoClientToken = string(token)
	} else {
		return appoptics.AOCommunicator{}, errors.NewNotFound(schema.GroupResource{}, "token")
	}
	return appoptics.NewAOCommunicator(aoClientToken), nil
}
