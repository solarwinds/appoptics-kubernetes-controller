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
	status, err = r.syncSpace(dash, status)
	if err != nil {
		return nil, err
	}

	aoChartHash, err := r.getChartHash(status.ID)
	if err != nil {
		return nil, err
	}

	specHash, err := Hash(spec)
	if err != nil {
		return nil, err
	}
	// Sync Charts
	if bytes.Compare(status.Hashes.AppOptics, aoChartHash) != 0 || bytes.Compare(status.Hashes.Spec,specHash) != 0  {
		err = r.syncCharts(dash.Charts, status.ID)
		if err != nil {
			return nil, err
		}

		status.Hashes.AppOptics, err = r.getChartHash(status.ID)
		if err != nil {
			return nil, err
		}
		status.Hashes.Spec = specHash
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
	specHash, err := Hash([]byte(specString))
	if err != nil {
		return nil, err
	}
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

func Hash(s interface{}) ([]byte, error) {
	byteArr, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	h := sha1.New()
	io.WriteString(h, string(byteArr))
	return h.Sum(nil), nil
}