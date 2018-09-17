package appoptics

import (
	aoApi "github.com/appoptics/appoptics-api-go"
	"github.com/ghodss/yaml"
	"github.com/appoptics/appoptics-kubernetes-controller/pkg/apis/appoptics-kubernetes-controller/v1"
	listers "github.com/appoptics/appoptics-kubernetes-controller/pkg/client/listers/appoptics-kubernetes-controller/v1"
	"strings"
	"crypto/sha1"
	"io"
	"bytes"
	"encoding/json"
)

type Synchronizer struct {
	Client *aoApi.Client
}


type AO interface {
	SyncSpace(v1.TokenAndDataSpec, *v1.Status) (*v1.Status, error)
	syncSpace(CustomSpace, int) (int, error)
	SyncService(v1.TokenAndDataSpec, *v1.Status) (*v1.Status, error)
	syncService(aoApi.Alert, *v1.Status) (*v1.Status, error)
	SyncAlert(v1.TokenAndDataSpec, *v1.Status, listers.ServiceNamespaceLister) (*v1.Status, error)
	syncAlert(aoApi.Alert, *v1.Status, bool) (*v1.Status, error)
	RemoveAlert(int) error
	RemoveService(int) error
	RemoveSpace(int) error
}

func NewSyncronizer(token string) Synchronizer{
	client := aoApi.NewClient(token)
	return Synchronizer{client}

}


func (r *Synchronizer) SyncSpace(spec v1.TokenAndDataSpec, status *v1.Status) (*v1.Status, error) {
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
		err = r.syncCharts(dash.Charts, ID)
		if err != nil {
			return nil, err
		}
	}

	return status, nil
}

func (r *Synchronizer) SyncService(spec v1.TokenAndDataSpec, status *v1.Status) (*v1.Status, error) {
	var service aoApi.Service
	err := yaml.Unmarshal([]byte(spec.Data), &service)
	if err != nil {
		return nil, err
	}

	// Sync Service
	return r.syncService(service, status)
}

func (r *Synchronizer) SyncAlert(spec v1.TokenAndDataSpec, status *v1.Status, serviceNamespaceLister listers.ServiceNamespaceLister) (*v1.Status, error) {
	specString, err := json.Marshal(spec)
	if err != nil {
		return nil, err
	}
	specHash := Hash([]byte(specString))
	specChanged := bytes.Compare(specHash, status.Hashes.Spec) != 0
	if specChanged {
		status.Hashes.Spec = specHash
	}
	var customAlert aoApi.Alert
	err = yaml.Unmarshal([]byte(spec.Data), &customAlert)
	if err != nil {
		return nil, err
	}

	var notificationServices []*aoApi.Service
	if services, ok := customAlert.Attributes["services"]; ok {
		for _, serviceObj := range services.([]interface{}) {
			serviceStr := serviceObj.(string)
			service, err := serviceNamespaceLister.Get(serviceStr)
			if err != err {
				return nil, err
			}
			if service != nil && service.Status.ID != 0 {
				notificationServices = append(notificationServices, &aoApi.Service{ID:&service.Status.ID})
			}
		}
	}

	customAlert.Services = notificationServices

	// Sync Alert
	status, err = r.syncAlert(customAlert, status, specChanged)
	if err != nil {
		return nil, err
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

func Hash(s []byte) []byte {
	h := sha1.New()
	io.WriteString(h, string(s))
	return h.Sum(nil)
}