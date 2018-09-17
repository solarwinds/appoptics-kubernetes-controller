package appoptics

import (
	aoApi "github.com/appoptics/appoptics-api-go"
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

func (r *Synchronizer) syncSpace(dash CustomSpace, ID int) (int, error) {

	spacesService := NewSpacesService(r.Client)
	// If we dont have an ID for it then we assume its new and create it
	if ID == 0 {
		space, err := spacesService.Create(dash.Name)
		if err != nil {
			return -1, err
		}
		ID = space.ID
	} else {
		// Lets ensure that the ID we have exists in AppOptics
		aoSpace, err := spacesService.Retrieve(ID)
		if err != nil {
			// If its a not found error thats ok we can try to create it now
			if CheckIfErrorIsAppOpticsNotFoundError(err) {
				space, err := spacesService.Create(dash.Name)
				if err != nil {
					return -1, err
				}
				ID = space.ID
			} else {
				return  -1, err
			}
		} else {
			//Service exists in AppOptics now lets check that they are actually synced
			if strings.Compare(aoSpace.Name, dash.Name) != 0 {
				_, err = spacesService.Update(ID, dash.Name)
				if err != nil {
					return -1, err
				}
			}
		}
	}

	return ID, nil

}

func (r *Synchronizer) RemoveSpace(ID int) error {

	spacesService := NewSpacesService(r.Client)
	// If we dont have an ID for it then we assume its new and create it
	if ID != 0 {
		err := spacesService.Delete(ID)
		if err != nil {
			return err
		}
	}
	return nil

}