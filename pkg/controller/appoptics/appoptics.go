package appoptics

import (
	aoApi "github.com/appoptics/appoptics-api-go"
	"github.com/solarwinds/appoptics-kubernetes-controller/pkg/apis/appoptics-kubernetes-controller/v1"
	listers "github.com/solarwinds/appoptics-kubernetes-controller/pkg/client/listers/appoptics-kubernetes-controller/v1"
	"strings"
)

type AOResource interface {
	Sync(v1.TokenAndDataSpec, *v1.Status) (*v1.Status, error)
	Delete(int) error
}

type AOResourceCommunicator interface {
	Sync(v1.TokenAndDataSpec, *v1.Status, string, listers.ServiceNamespaceLister) (*v1.Status, error)
	Remove(int, string) error
}

type AOCommunicator struct {
	Token string
}

func (aoc *AOCommunicator) Remove(ID int, kind string) error {
	client := aoApi.NewClient(aoc.Token)
	switch strings.ToLower(kind) {
	case Dashboard:
		chartsService := NewChartsService(client)
		aoCharts, err := chartsService.List(ID)
		if err != nil && !CheckIfErrorIsAppOpticsNotFoundError(err){
			return err
		}

		// DELETE ALL AO CHARTS AND CREATE OURS
		if len(aoCharts) != 0 {
			err = chartsService.DeleteAll(aoCharts, ID)
			if err != nil && !CheckIfErrorIsAppOpticsNotFoundError(err){
				return err
			}
		}

		// delete all charts
		spacesService := NewSpacesService(client)
		err = spacesService.Delete(ID)
		if err != nil && !CheckIfErrorIsAppOpticsNotFoundError(err){
			return err
		}
	case Service:
		servicesService := NewServicesService(client)
		err := servicesService.Delete(ID)
		if err != nil && !CheckIfErrorIsAppOpticsNotFoundError(err){
			return err
		}
	case Alert:
		alertsService := NewAlertsService(client, nil)
		err := alertsService.Delete(ID)
		if err != nil && !CheckIfErrorIsAppOpticsNotFoundError(err){
			return err
		}
	}
	return nil
}

func (aoc *AOCommunicator) Sync(spec v1.TokenAndDataSpec, status *v1.Status, kind string, lister listers.ServiceNamespaceLister) (*v1.Status, error) {
	client := aoApi.NewClient(aoc.Token)
	switch strings.ToLower(kind) {
	case Dashboard:
		spacesService := NewSpacesService(client)
		return spacesService.Sync(spec, status)
	case Service:
		servicesService := NewServicesService(client)
		return servicesService.Sync(spec, status)
	case Alert:
		alertService := NewAlertsService(client, lister)
		return alertService.Sync(spec, status)
	}
	return status, nil
}
