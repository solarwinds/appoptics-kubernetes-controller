package checks

import (
	aoApi "github.com/appoptics/appoptics-api-go"
	"strconv"
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

func (r *ResourcesToSync) syncSpace(dash CustomSpace, ID int) (int, error) {

	spacesService := NewSpacesService(r.Client)
	// If we dont have an ID for it then we assume its new and create it
	if ID == 0 {
		space, err := spacesService.create(dash.Name)
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
				space, err := spacesService.create(dash.Name)
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

func (s *SpacesService) create(name string) (*aoApi.Space, error) {
	newSpace := &SimpleSpace{Name: name}

	req, err := s.client.NewRequest("POST", "spaces", newSpace)
	if err != nil {
		return nil, err
	}

	res, err := s.client.Do(req, newSpace)
	if err != nil {
		return nil, err
	}

	idHeaderValue := res.Header.Get("Location")
	idStartIdx := strings.LastIndex(idHeaderValue, "/")
	idStr := idHeaderValue[idStartIdx+1:]
	id, err := strconv.Atoi(idStr)

	return &aoApi.Space{ID: id, Name: name}, nil
}
