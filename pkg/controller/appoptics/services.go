package appoptics

import (
	"encoding/json"
	"reflect"
	"time"

	aoApi "github.com/appoptics/appoptics-api-go"
	"github.com/ghodss/yaml"
	v1 "github.com/solarwinds/appoptics-kubernetes-controller/pkg/apis/appopticskubernetescontroller/v1"
)

type ServicesService struct {
	aoApi.ServicesCommunicator
	client *aoApi.Client
}

func NewServicesService(c *aoApi.Client) *ServicesService {
	return &ServicesService{c.ServicesService(), c}
}

func (ss *ServicesService) Sync(spec v1.TokenAndDataSpec, status *v1.Status) (*v1.Status, error) {
	var service aoApi.Service
	err := yaml.Unmarshal([]byte(spec.Data), &service)
	if err != nil {
		return nil, err
	}

	// If we dont have an ID for it then we assume its new and create it
	if status.ID == 0 {
		return ss.createService(service, status)
	} else {
		// Lets ensure that the ID we have exists in AppOptics
		aoService, err := ss.Retrieve(status.ID)
		if err != nil {
			if CheckIfErrorIsAppOpticsNotFoundError(err, Service, status.ID) {
				return ss.createService(service, status)
			} else {
				return nil, err
			}
		} else {
			//Service exists in AppOptics now lets check that they are actually synced
			service.ID = status.ID
			if !reflect.DeepEqual(&service, aoService) {
				// Local vs Remote are different so update AO
				err = ss.Update(&service)
				if err != nil {
					return nil, err
				}
				status.UpdatedAt = int(time.Now().Unix())
				serviceJson, err := json.Marshal(aoService)
				if err != nil {
					return nil, err
				}
				status.Hashes.AppOptics, err = Hash(serviceJson)
				if err != nil {
					return nil, err
				}
			}
		}
	}

	return status, nil

}

func (ss *ServicesService) createService(service aoApi.Service, status *v1.Status) (*v1.Status, error) {
	aoService, err := ss.Create(&service)
	if err != nil {
		return nil, err
	}
	status.ID = aoService.ID
	status.UpdatedAt = int(time.Now().Unix())
	hash, err := json.Marshal(aoService)
	if err != nil {
		return nil, err
	}
	status.Hashes.AppOptics = hash
	return status, nil
}
