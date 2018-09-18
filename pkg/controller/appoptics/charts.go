package appoptics

import (
	aoApi "github.com/appoptics/appoptics-api-go"
)

type ChartsService struct {
	aoApi.ChartsCommunicator
	client *aoApi.Client
}

func NewChartsService(c *aoApi.Client) *ChartsService {
	return &ChartsService{aoApi.NewChartsService(c), c}
}

func (c *ChartsService) DeleteAll(charts []*aoApi.Chart, spaceID int) error {
	for _, chart := range charts {
		err := c.Delete(*chart.ID, spaceID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *Synchronizer) syncCharts(dashCharts []*aoApi.Chart, spaceID int) error {
	chartsService := NewChartsService(r.Client)
	aoCharts, err := chartsService.List(spaceID)
	if err != nil {
		return err
	}

	// DELETE ALL AO CHARTS AND CREATE OURS
	if len(aoCharts) != 0 {
		err = chartsService.DeleteAll(aoCharts, spaceID)
		if err != nil {
			return err
		}
	}
	for _, chart := range dashCharts {
		_, err = chartsService.Create(chart, spaceID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *Synchronizer) getChartHash(spaceID int) ([]byte, error) {
	chartsService := NewChartsService(r.Client)

	if spaceID == 0 {
		return []byte(""), nil
	}
	aoCharts, err := chartsService.List(spaceID)
	if err != nil {
		return nil, err
	}

	// DELETE ALL AO CHARTS AND CREATE OURS
	if aoCharts != nil && len(aoCharts) != 0 {
		return Hash(aoCharts)
	}

	return []byte(""), nil
}
