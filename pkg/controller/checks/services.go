package checks

import (
	aoApi "github.com/appoptics/appoptics-api-go"
	"reflect"
)

func (r *ResourcesToSync) syncService(service aoApi.Service, ID int) (int, error) {
	servicesService := aoApi.NewServiceService(r.Client)
	// If we dont have an ID for it then we assume its new and create it
	if ID == 0 {
		service, err := servicesService.Create(&service)
		if err != nil {
			return -1, err
		}
		ID = *service.ID
	} else {
		// Lets ensure that the ID we have exists in AppOptics
		aoService, err := servicesService.Retrieve(ID)
		if err != nil {
			if CheckIfErrorIsAppOpticsNotFoundError(err) {
				aoService, err := servicesService.Create(&service)
				if err != nil {
					return -1, err
				}
				ID = *aoService.ID
			} else {
				return -1, err
			}
		} else {
			//Service exists in AppOptics now lets check that they are actually synced
			service.ID = aoService.ID
			if !reflect.DeepEqual(&service, aoService) {
				// Local vs Remote are different so update AO
				err = servicesService.Update(&service)
				if err != nil {
					return -1, err
				}
			}
		}
	}

	return ID, nil

}
