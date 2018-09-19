package appoptics

import (
	"encoding/json"
	aoApi "github.com/appoptics/appoptics-api-go"
	"github.com/appoptics/appoptics-kubernetes-controller/pkg/apis/appoptics-kubernetes-controller/v1"
)

type AlertsService struct {
	aoApi.AlertsService
	client aoApi.Client
}

func NewAlertsService(c *aoApi.Client) *AlertsService {
	return &AlertsService{*aoApi.NewAlertsService(c), *c}
}

func (r *Synchronizer) syncAlert(alert aoApi.Alert, status *v1.Status, specChange bool) (*v1.Status, error) {
	alertsService := NewAlertsService(r.Client)
	// If we dont have an ID for it then we assume its new and create it
	if status.ID == 0 {
		return r.createAlert(alert, status, alertsService)
	} else {
		// Lets ensure that the ID we have exists in AppOptics
		aoAlert, err := alertsService.Retrieve(status.ID)
		if err != nil {
			// If its a not found error thats ok we can try to create it now
			if CheckIfErrorIsAppOpticsNotFoundError(err) {
				return r.createAlert(alert, status, alertsService)
			} else {
				return nil, err
			}
		} else {
			//Service exists in AppOptics now lets check that they are actually synced
			if status.UpdatedAt != *aoAlert.UpdatedAt || specChange {
				// Local vs Remote are different so update AO
				//SET THE ALERT ID FOR THE OBJECT ABOUT TO BE PUT

				alert.ID = &status.ID

				//Associate services
				for _, service := range alert.Services {
					err = alertsService.AssociateToService(*alert.ID, *service.ID)
					if err != nil {
						return nil, err
					}
				}

				err = alertsService.Update(&alert)
				if err != nil {
					return nil, err
				}

				aoAlert, err = alertsService.Retrieve(status.ID)
				if err != nil {
					return nil, err
				}
				hash, err := json.Marshal(aoAlert)
				if err != nil {
					return nil, err
				}
				status.Hashes.AppOptics = hash
				status.UpdatedAt = *aoAlert.UpdatedAt
			}
		}
	}
	return status, nil
}

func (r *Synchronizer) createAlert(alert aoApi.Alert, status *v1.Status, alertsService *AlertsService) (*v1.Status, error) {
	//Associate services
	services := alert.Services
	alert.Services = nil

	aoAlert, err := alertsService.Create(&alert)
	if err != nil {
		return nil, err
	}

	for _, service := range services {
		err = alertsService.AssociateToService(*aoAlert.ID, *service.ID)
		if err != nil {
			return nil, err
		}
	}
	status.ID = *aoAlert.ID
	status.UpdatedAt = *aoAlert.UpdatedAt
	hash, err := json.Marshal(aoAlert)
	if err != nil {
		return nil, err
	}
	status.Hashes.AppOptics = hash
	return status, nil
}

func (r *Synchronizer) RemoveAlert(ID int) error {
	alertsService := NewAlertsService(r.Client)
	// If we dont have an ID for it then we assume its new and create it
	if ID != 0 {
		// Lets ensure that the ID we have exists in AppOptics
		return alertsService.Delete(ID)
	}
	return nil
}
