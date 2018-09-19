package appoptics

import (
	aoApi "github.com/appoptics/appoptics-api-go"
	"github.com/appoptics/appoptics-kubernetes-controller/pkg/apis/appoptics-kubernetes-controller/v1"
	"strings"
)

type CustomSpace struct {
	aoApi.Space
	Charts []*aoApi.Chart `json:"charts","omitempty"`
}

type SimpleSpace struct {
	Name string `json:"name","omitempty"`
}

type SpacesService struct {
	aoApi.SpacesCommunicator
	client *aoApi.Client
}

func NewSpacesService(c *aoApi.Client) *SpacesService {
	return &SpacesService{c.SpacesService(), c}
}

func (r *Synchronizer) syncSpace(dash CustomSpace, status *v1.Status) (*v1.Status, error) {

	spacesService := NewSpacesService(r.Client)
	// If we dont have an ID for it then we assume its new and create it
	if status.ID == 0 {
		space, err := spacesService.Create(dash.Name)
		if err != nil {
			return nil, err
		}
		status.ID = space.ID
	} else {
		// Lets ensure that the ID we have exists in AppOptics
		aoSpace, err := spacesService.Retrieve(status.ID)
		if err != nil {
			// If its a not found error thats ok we can try to create it now
			if CheckIfErrorIsAppOpticsNotFoundError(err) {
				space, err := spacesService.Create(dash.Name)
				if err != nil {
					return nil, err
				}
				status.ID = space.ID
			} else {
				return nil, err
			}
		} else {
			//Service exists in AppOptics now lets check that they are actually synced
			if strings.Compare(aoSpace.Name, dash.Name) != 0 {
				_, err = spacesService.Update(status.ID, dash.Name)
				if err != nil {
					return nil, err
				}
			}
		}
	}

	return status, nil

}

func (r *Synchronizer) RemoveSpace(ID int) error {

	spacesService := NewSpacesService(r.Client)
	// If we dont have an ID for it then we assume its new and create it
	if ID != 0 {
		return spacesService.Delete(ID)
	}
	return nil
}
