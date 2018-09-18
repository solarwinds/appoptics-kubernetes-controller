package appoptics

import (
	"encoding/json"
	aoApi "github.com/appoptics/appoptics-api-go"
	"github.com/appoptics/appoptics-kubernetes-controller/pkg/apis/appoptics-kubernetes-controller/v1"
	"reflect"
	"time"
)

func (r *Synchronizer) syncService(service aoApi.Service, status *v1.Status) (*v1.Status, error) {
	servicesService := aoApi.NewServiceService(r.Client)
	// If we dont have an ID for it then we assume its new and create it
	if status.ID == 0 {
		return r.createService(service, status, servicesService)
	} else {
		// Lets ensure that the ID we have exists in AppOptics
		aoService, err := servicesService.Retrieve(status.ID)
		if err != nil {
			if CheckIfErrorIsAppOpticsNotFoundError(err) {
				return r.createService(service, status, servicesService)
			} else {
				return nil, err
			}
		} else {
			//Service exists in AppOptics now lets check that they are actually synced
			service.ID = &status.ID
			if !reflect.DeepEqual(&service, aoService) {
				// Local vs Remote are different so update AO
				err = servicesService.Update(&service)
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

func (r *Synchronizer) createService(service aoApi.Service, status *v1.Status, servicesService *aoApi.ServicesService) (*v1.Status, error) {
	aoService, err := servicesService.Create(&service)
	if err != nil {
		return nil, err
	}
	status.ID = *aoService.ID
	status.UpdatedAt = int(time.Now().Unix())
	hash, err := json.Marshal(aoService)
	if err != nil {
		return nil, err
	}
	status.Hashes.AppOptics = hash
	return status, nil
}

func (r *Synchronizer) RemoveService(ID int) error {
	servicesService := aoApi.NewServiceService(r.Client)
	if ID != 0 {
		err := servicesService.Delete(ID)
		if err != nil {
			return err
		}
	}
	return nil
}
