package appoptics

import (
	aoApi "github.com/appoptics/appoptics-api-go"
	"github.com/appoptics/appoptics-kubernetes-controller/pkg/apis/appoptics-kubernetes-controller/v1"
	listers "github.com/appoptics/appoptics-kubernetes-controller/pkg/client/listers/appoptics-kubernetes-controller/v1"
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
		spacesService := NewSpacesService(client)
		return spacesService.Delete(ID)
	case Service:
		servicesService := NewServicesService(client)
		return servicesService.Delete(ID)
	case Alert:
		alertsService := NewAlertsService(client, nil)
		return alertsService.Delete(ID)
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
