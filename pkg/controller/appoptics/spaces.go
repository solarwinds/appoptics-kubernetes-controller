package appoptics

import (
	"bytes"
	aoApi "github.com/appoptics/appoptics-api-go"
	"github.com/ghodss/yaml"
	"github.com/solarwinds/appoptics-kubernetes-controller/pkg/apis/appoptics-kubernetes-controller/v1"
	"strings"
)

type CustomSpace struct {
	aoApi.Space
	Charts []*aoApi.Chart `json:"charts,omitempty"`
}

type SpacesService struct {
	aoApi.SpacesCommunicator
	client *aoApi.Client
}

func NewSpacesService(c *aoApi.Client) *SpacesService {
	return &SpacesService{c.SpacesService(), c}
}

func (s *SpacesService) Sync(spec v1.TokenAndDataSpec, status *v1.Status) (*v1.Status, error) {
	var dash CustomSpace
	err := yaml.Unmarshal([]byte(spec.Data), &dash)
	if err != nil {
		return nil, err
	}

	// Sync Space aka Dashboard at a high level
	status, err = s.sync(dash, status)
	if err != nil {
		return nil, err
	}
	chartService := NewChartsService(s.client)
	aoChartHash, err := chartService.getChartHash(status.ID)
	if err != nil {
		return nil, err
	}

	specHash, err := Hash(spec)
	if err != nil {
		return nil, err
	}
	// Sync Charts
	if bytes.Compare(status.Hashes.AppOptics, aoChartHash) != 0 || bytes.Compare(status.Hashes.Spec, specHash) != 0 {
		err = chartService.syncCharts(dash.Charts, status.ID)
		if err != nil {
			return nil, err
		}

		status.Hashes.AppOptics, err = chartService.getChartHash(status.ID)
		if err != nil {
			return nil, err
		}
		status.Hashes.Spec = specHash
	}

	return status, nil
}

func (s *SpacesService) sync(dash CustomSpace, status *v1.Status) (*v1.Status, error) {
	// If we dont have an ID for it then we assume its new and create it
	if status.ID == 0 {
		space, err := s.Create(dash.Name)
		if err != nil {
			return nil, err
		}
		status.ID = space.ID
	} else {
		// Lets ensure that the ID we have exists in AppOptics
		aoSpace, err := s.Retrieve(status.ID)
		if err != nil {
			// If its a not found error thats ok we can try to create it now
			if CheckIfErrorIsAppOpticsNotFoundError(err, Dashboard, status.ID) {
				space, err := s.Create(dash.Name)
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
				_, err = s.Update(status.ID, dash.Name)
				if err != nil {
					return nil, err
				}
			}
		}
	}
	return status, nil

}
