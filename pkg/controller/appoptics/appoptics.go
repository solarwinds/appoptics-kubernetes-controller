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
	Sync(v1.TokenAndDataSpec, *v1.Status, string, listers.AppOpticsServiceNamespaceLister) (*v1.Status, error)
	Remove(int, string) error
}

type AOCommunicator struct {
	Client aoApi.Client
}

func NewAOCommunicator(token string) AOCommunicator {
	client := aoApi.NewClient(token)
	return AOCommunicator{*client}
}

func (aoc *AOCommunicator) Remove(ID int, kind string) error {
	switch strings.ToLower(kind) {
	case Dashboard:
		// delete all charts
		spacesService := NewSpacesService(&aoc.Client)
		err := spacesService.Delete(ID)
		if err != nil && !CheckIfErrorIsAppOpticsNotFoundError(err, kind, ID) {
			return err
		}
	case Service:
		servicesService := NewServicesService(&aoc.Client)
		err := servicesService.Delete(ID)
		if err != nil && !CheckIfErrorIsAppOpticsNotFoundError(err, kind, ID) {
			return err
		}
	case Alert:
		alertsService := NewAlertsService(&aoc.Client, nil)
		err := alertsService.Delete(ID)
		if err != nil && !CheckIfErrorIsAppOpticsNotFoundError(err, kind, ID) {
			return err
		}
	}
	return nil
}

func (aoc *AOCommunicator) Sync(spec v1.TokenAndDataSpec, status *v1.Status, kind string, lister listers.AppOpticsServiceNamespaceLister) (*v1.Status, error) {
	switch strings.ToLower(kind) {
	case Dashboard:
		spacesService := NewSpacesService(&aoc.Client)
		return spacesService.Sync(spec, status)
	case Service:
		servicesService := NewServicesService(&aoc.Client)
		return servicesService.Sync(spec, status)
	case Alert:
		alertService := NewAlertsService(&aoc.Client, lister)
		return alertService.Sync(spec, status)
	}
	return status, nil
}
