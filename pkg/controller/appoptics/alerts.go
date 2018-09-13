package appoptics

import (
	"fmt"
	aoApi "github.com/appoptics/appoptics-api-go"
	"reflect"
)

type AlertRequest struct {
	ID           *int                    `json:"id,omitempty"`
	Name         *string                 `json:"name,omitempty"`
	Description  *string                 `json:"description,omitempty"`
	Active       *bool                   `json:"active,omitempty"`
	Attributes   map[string][]string     `json:"attributes","omitempty"`
	RearmSeconds *int                    `json:"rearm_seconds,omitempty"`
	Conditions   []*aoApi.AlertCondition `json:"conditions,omitempty"`
	Services     []*int                  `json:"services,omitempty"` // correspond to IDs of Service objects
	CreatedAt    *int                    `json:"created_at,omitempty"`
	UpdatedAt    *int                    `json:"updated_at,omitempty"`
}

type AlertsService struct {
	aoApi.AlertsService
	client aoApi.Client
}

func NewAlertsService(c *aoApi.Client) *AlertsService {
	return &AlertsService{*aoApi.NewAlertsService(c), *c}
}

func (r *Synchronizer) syncAlert(alert AlertRequest, ID int) (int, error) {
	alertsService := NewAlertsService(r.Client)
	// If we dont have an ID for it then we assume its new and create it
	if ID == 0 {
		alert, err := alertsService.CreateCustom(&alert)
		if err != nil {
			return -1, err
		}
		ID = *alert.ID
	} else {
		// Lets ensure that the ID we have exists in AppOptics
		aoAlert, err := alertsService.Retrieve(ID)
		if err != nil {
			// If its a not found error thats ok we can try to create it now
			if CheckIfErrorIsAppOpticsNotFoundError(err) {
				alertResp, err := alertsService.CreateCustom(&alert)
				if err != nil {
					return -1, err
				}
				ID = *alertResp.ID
			} else {
				return  -1, err
			}
		} else {
			//Service exists in AppOptics now lets check that they are actually synced
			alert.ID = &ID
			if !reflect.DeepEqual(&alert, aoAlert) {
				// Local vs Remote are different so update AO
				err = alertsService.UpdateCustom(&alert)
				if err != nil {
					return -1, err
				}
			}
		}
	}
	return ID, nil
}

func (r *Synchronizer) RemoveAlert(ID int) error {
	alertsService := NewAlertsService(r.Client)
	// If we dont have an ID for it then we assume its new and create it
	if ID != 0 {
		// Lets ensure that the ID we have exists in AppOptics
		err := alertsService.DeleteCustom(ID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (as *AlertsService) CreateCustom(a *AlertRequest) (*aoApi.Alert, error) {
	req, err := as.client.NewRequest("POST", "alerts", a)
	if err != nil {
		return nil, err
	}

	createdAlert := &aoApi.Alert{}

	_, err = as.client.Do(req, createdAlert)
	if err != nil {
		return nil, err
	}

	return createdAlert, nil
}

func (as *AlertsService) UpdateCustom(a *AlertRequest) error {
	path := fmt.Sprintf("alerts/%d", *a.ID)
	req, err := as.client.NewRequest("PUT", path, a)
	_, err = as.client.Do(req, nil)
	if err != nil {
		return err
	}
	return nil
}

func (as *AlertsService) DeleteCustom(ID int) error {
	path := fmt.Sprintf("alerts/%d", ID)
	req, err := as.client.NewRequest("DELETE", path, nil)
	_, err = as.client.Do(req, nil)
	if err != nil {
		return err
	}
	return nil
}
