package appoptics

import (
	aoApi "github.com/appoptics/appoptics-api-go"
	"github.com/ghodss/yaml"
	"github.com/appoptics/appoptics-kubernetes-controller/pkg/apis/appoptics-kubernetes-controller/v1"
	listers "github.com/appoptics/appoptics-kubernetes-controller/pkg/client/listers/appoptics-kubernetes-controller/v1"
	"strings"
)

type Synchronizer struct {
	Client *aoApi.Client
}

type AO interface {
	SyncSpace(v1.TokenAndDataSpec, *v1.TimestampAndIdStatus) (*v1.TimestampAndIdStatus, error)
	syncSpace(CustomSpace, int) (int, error)
	SyncService(v1.TokenAndDataSpec, *v1.TimestampAndIdStatus) (*v1.TimestampAndIdStatus, error)
	syncService(aoApi.Service, int) (int, error)
	SyncAlert(v1.TokenAndDataSpec, *v1.TimestampAndIdStatus, listers.ServiceNamespaceLister) (*v1.TimestampAndIdStatus, error)
	syncAlert(AlertRequest, int) (int, error)
	RemoveAlert(int) error
	RemoveService(int) error
	RemoveSpace(int) error
}

func NewSyncronizer(token string) Synchronizer{
	client := aoApi.NewClient(token)
	return Synchronizer{client}

}


func (r *Synchronizer) SyncSpace(spec v1.TokenAndDataSpec, status *v1.TimestampAndIdStatus) (*v1.TimestampAndIdStatus, error) {
	var dash CustomSpace
	err := yaml.Unmarshal([]byte(spec.Data), &dash)
	if err != nil {
		return nil, err
	}

	// Sync Space aka Dashboard at a high level
	ID, err := r.syncSpace(dash, status.ID)
	if err != nil {
		return nil, err
	}

	if ID != status.ID {
		status.ID = ID
	}

	// Sync Charts
	if dash.Charts != nil && len(dash.Charts) > 0 {
		err = r.syncCharts(dash.Charts, status.ID)
		if err != nil {
			return nil, err
		}
	}

	return status, nil
}

func (r *Synchronizer) SyncService(spec v1.TokenAndDataSpec, status *v1.TimestampAndIdStatus) (*v1.TimestampAndIdStatus, error) {
	var service aoApi.Service
	err := yaml.Unmarshal([]byte(spec.Data), &service)
	if err != nil {
		return nil, err
	}

	// Sync Service
	ID, err := r.syncService(service, status.ID)
	if err != nil {
		return nil, err
	}

	if ID != status.ID {
		status.ID = ID
	}
	return status, nil
}

func (r *Synchronizer) SyncAlert(spec v1.TokenAndDataSpec, status *v1.TimestampAndIdStatus, serviceNamespaceLister listers.ServiceNamespaceLister) (*v1.TimestampAndIdStatus, error) {
	var customAlert AlertRequest
	err := yaml.Unmarshal([]byte(spec.Data), &customAlert)
	if err != nil {
		return nil, err
	}
	var notificationServices []*int
	if services, ok := customAlert.Attributes["services"]; ok {
		for _, serviceObj := range services {
			service, err := serviceNamespaceLister.Get(serviceObj)
			if err != err {
				return nil, err
			}
			if service != nil && service.Status.ID != 0 {
				notificationServices = append(notificationServices, &service.Status.ID)
			}
		}
	}

	customAlert.Services = notificationServices

	// Sync Alert
	ID, err := r.syncAlert(customAlert, status.ID)
	if err != nil {
		return nil, err
	}

	if ID != status.ID {
		status.ID = ID
	}
	return status, nil
}

func CheckIfErrorIsAppOpticsNotFoundError(err error) (bool) {
	if errorResponse, ok := err.(*aoApi.ErrorResponse); ok {
		errorObj := errorResponse.Errors.(map[string]interface{})
		if requestErr, ok := errorObj["request"]; ok {
			for _, errorType := range requestErr.([]interface{}) {
				// The ID does not exist in AppOptics so create a new space
				if strings.Compare(errorType.(string), "Not Found") == 0 {
					return true
				}
			}
		}
	}
	return false
}