package appoptics

import (
	"bytes"
	"encoding/json"
	aoApi "github.com/appoptics/appoptics-api-go"
	"github.com/appoptics/appoptics-kubernetes-controller/pkg/apis/appoptics-kubernetes-controller/v1"
	listers "github.com/appoptics/appoptics-kubernetes-controller/pkg/client/listers/appoptics-kubernetes-controller/v1"
	"github.com/ghodss/yaml"
)

type AlertsService struct {
	aoApi.AlertsService
	client aoApi.Client
	lister listers.ServiceNamespaceLister
}

func NewAlertsService(c *aoApi.Client, lister listers.ServiceNamespaceLister) *AlertsService {
	return &AlertsService{*aoApi.NewAlertsService(c), *c, lister}
}

func (as *AlertsService) Sync(spec v1.TokenAndDataSpec, status *v1.Status) (*v1.Status, error) {
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
			service, err := as.lister.Get(serviceStr)
			if err != err {
				return nil, err
			}
			if service != nil && service.Status.ID != 0 {
				notificationServices = append(notificationServices, &aoApi.Service{ID: &service.Status.ID})
			}
		}
	}

	customAlert.Services = notificationServices
	// If we dont have an ID for it then we assume its new and create it
	if status.ID == 0 {
		return as.createAlert(customAlert, status)
	} else {
		// Lets ensure that the ID we have exists in AppOptics
		aoAlert, err := as.Retrieve(status.ID)
		if err != nil {
			// If its a not found error thats ok we can try to create it now
			if CheckIfErrorIsAppOpticsNotFoundError(err) {
				return as.createAlert(customAlert, status)
			} else {
				return nil, err
			}
		} else {
			//Service exists in AppOptics now lets check that they are actually synced
			if status.UpdatedAt != *aoAlert.UpdatedAt || specChanged {
				// Local vs Remote are different so update AO
				//SET THE ALERT ID FOR THE OBJECT ABOUT TO BE PUT
				customAlert.ID = aoAlert.ID
				//Associate services
				for _, service := range aoAlert.Services {
					err = as.AssociateToService(*aoAlert.ID, *service.ID)
					if err != nil {
						return nil, err
					}
				}
				// Update the alert
				err = as.Update(&customAlert)
				if err != nil {
					return nil, err
				}

				// Retrieve the Updated alert
				aoAlert, err = as.Retrieve(status.ID)
				if err != nil {
					return nil, err
				}
				// Hash the current AO Object
				hash, err := json.Marshal(aoAlert)
				if err != nil {
					return nil, err
				}

				// Store the Hashes
				status.Hashes.AppOptics = hash
				status.UpdatedAt = *aoAlert.UpdatedAt
			}
		}
	}
	return status, nil
}

func (as *AlertsService) createAlert(alert aoApi.Alert, status *v1.Status) (*v1.Status, error) {
	//Associate services
	services := alert.Services
	// Nil out as the current Alert.Services struct is not an array of ints
	alert.Services = nil

	aoAlert, err := as.Create(&alert)
	if err != nil {
		return nil, err
	}

	// Associate Services to the Alert
	for _, service := range services {
		err = as.AssociateToService(*aoAlert.ID, *service.ID)
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
