package controller

import (
	"fmt"
	"time"

	"github.com/appoptics/appoptics-kubernetes-controller/pkg/controller/appoptics"
	"github.com/golang/glog"
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

func (c *Controller) syncHandler(key string) error {

	namespace, kind, name, err := c.SplitMetaNamespaceKey(key)
	if err != nil {
		runtime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}
	currentTime := time.Now()

	switch kind {
	case Dashboard:
		dashboard, err := c.dashboardLister.Dashboards(namespace).Get(name)
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

		secret, err := c.kubeclientset.CoreV1().Secrets(namespace).Get(dashboard.Spec.Secret, metav1.GetOptions{})
		if err != nil {
			return err
		}

		synchronizer := appoptics.Synchronizer{}
		if token, ok := secret.Data["token"]; ok {
			synchronizer = appoptics.NewSyncronizer(string(token))
		} else {
			return errors.NewNotFound(schema.GroupResource{}, "token")
		}

		if dashboard.DeletionTimestamp != nil {
			updateObj := dashboard.DeepCopy()
			if len(updateObj.Finalizers) > 0 {
				for i, v := range updateObj.Finalizers {
					if v == "appoptics.io" {
						err = synchronizer.RemoveSpace(updateObj.Status.ID)
						if err != nil {
							return err
						}
						updateObj.Finalizers = append(updateObj.Finalizers[:i], updateObj.Finalizers[i+1:]...)
						break
					}
				}
			}
			updateObj, err := c.aoclientset.AppopticsV1().Dashboards(namespace).Update(updateObj)
			if err != nil {
				return err
			}
			return nil
		}
		updateStatus, err = synchronizer.SyncSpace(dashboard.Spec, updateStatus)
		if err != nil {
			return err
		}
		dashboardCopy := dashboard.DeepCopy()

		dashboardCopy.Status = *updateStatus

		_, err = c.aoclientset.AppopticsV1().Dashboards(dashboard.Namespace).Update(dashboardCopy)
		if err != nil {
			c.recorder.Event(dashboard, v1.EventTypeWarning, ErrUpdateStatus, err.Error())
		} else {
			c.recorder.Event(dashboardCopy, v1.EventTypeNormal, SuccessUpdate, MessageResourceUpdated)
		}
	case Service:
		service, err := c.serviceLister.Services(namespace).Get(name)
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

		secret, err := c.kubeclientset.CoreV1().Secrets(namespace).Get(service.Spec.Secret, metav1.GetOptions{})
		if err != nil {
			return err
		}

		synchronizer := appoptics.Synchronizer{}
		if token, ok := secret.Data["token"]; ok {
			synchronizer = appoptics.NewSyncronizer(string(token))
		} else {
			return errors.NewNotFound(schema.GroupResource{}, "token")
		}

		if service.DeletionTimestamp != nil {
			updateObj := service.DeepCopy()
			if len(updateObj.Finalizers) > 0 {
				for i, v := range updateObj.Finalizers {
					if v == "appoptics.io" {
						err = synchronizer.RemoveService(updateObj.Status.ID)
						if err != nil {
							return err
						}
						updateObj.Finalizers = append(updateObj.Finalizers[:i], updateObj.Finalizers[i+1:]...)
						break
					}
				}
			}
			updateObj, err := c.aoclientset.AppopticsV1().Services(namespace).Update(updateObj)
			if err != nil {
				return err
			}
			return nil
		}
		updateStatus, err = synchronizer.SyncService(service.Spec, updateStatus)
		if err != nil {
			return err
		}

		serviceCopy := service.DeepCopy()

		serviceCopy.Status = *updateStatus

		_, err = c.aoclientset.AppopticsV1().Services(service.Namespace).Update(serviceCopy)
		if err != nil {
			c.recorder.Event(service, v1.EventTypeWarning, ErrUpdateStatus, err.Error())
		} else {
			c.recorder.Event(serviceCopy, v1.EventTypeNormal, SuccessUpdate, MessageResourceUpdated)
		}
	case Alert:
		alert, err := c.alertLister.Alerts(namespace).Get(name)
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
		secret, err := c.kubeclientset.CoreV1().Secrets(namespace).Get(alert.Spec.Secret, metav1.GetOptions{})
		if err != nil {
			return err
		}

		synchronizer := appoptics.Synchronizer{}
		if token, ok := secret.Data["token"]; ok {
			synchronizer = appoptics.NewSyncronizer(string(token))
		} else {
			return errors.NewNotFound(schema.GroupResource{}, "token")
		}
		// NEVER modify objects from the store. It's a read-only, local cache.
		updateStatus := alert.Status.DeepCopy()
		updateStatus.LastUpdated = currentTime.Format(DateFormat)

		if alert.DeletionTimestamp != nil {
			updateObj := alert.DeepCopy()
			if len(updateObj.Finalizers) > 0 {
				for i, v := range updateObj.Finalizers {
					if v == "appoptics.io" {
						err = synchronizer.RemoveAlert(updateObj.Status.ID)
						if err != nil {
							return err
						}
						updateObj.Finalizers = append(updateObj.Finalizers[:i], updateObj.Finalizers[i+1:]...)
						break
					}
				}
			}
			updateObj, err := c.aoclientset.AppopticsV1().Alerts(namespace).Update(updateObj)
			if err != nil {
				return err
			}
			return nil
		}
		updateStatus, err = synchronizer.SyncAlert(alert.Spec, updateStatus, c.serviceLister.Services(namespace))
		if err != nil {
			return err
		}
		alertCopy := alert.DeepCopy()

		alertCopy.Status = *updateStatus

		_, err = c.aoclientset.AppopticsV1().Alerts(alert.Namespace).Update(alertCopy)
		if err != nil {
			c.recorder.Event(alert, v1.EventTypeWarning, ErrUpdateStatus, err.Error())
		} else {
			c.recorder.Event(alertCopy, v1.EventTypeNormal, SuccessUpdate, MessageResourceUpdated)
		}
	}

	return nil
}
