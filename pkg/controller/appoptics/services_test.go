package appoptics

import (
	aoApi "github.com/appoptics/appoptics-api-go"
	"github.com/solarwinds/appoptics-kubernetes-controller/pkg/apis/appoptics-kubernetes-controller/v1"
	"github.com/stretchr/testify/assert"
	"net/http"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"strconv"
)

// This tests an Existing Service in AppOptics being successfully updates
func TestExistingServiceSyncSuccess(t *testing.T) {

	ts := v1.Status{ID: 1, LastUpdated: "Yesterday"}

	data := `
           {
                "type": "mail",
                "settings": {
                    "addresses": "SW-MSP-PE-Event-Bus-oncall@swmspdevops.opsgenie.net"
                },
                "title": "TEST"
           }
`

	td := v1.TokenAndDataSpec{Namespace: "Default", Data: data, Secret: "blah"}

	ts1, err := aoc.Sync(td, &ts, Service, nil)
	if err != nil {
		t.Errorf("error running TestExistingServiceSync: %v", err)
	}

	assert.Equal(t, ts1.ID, ts.ID)
}

// This tests an Existing Service failing to be updated
func TestUpdateExistingServicesSyncFailure(t *testing.T) {

	ts := v1.Status{ID: 3, LastUpdated: "Yesterday"}

	data := `{
  "id": 145,
  "type": "mail",
  "settings": {
    "addresses": "george@example.com,fred@example.com"
  },
  "title": "NewServiceError"
}`

	td := v1.TokenAndDataSpec{Namespace: "Default", Data: data, Secret: "blah"}

	_, err := aoc.Sync(td, &ts, Service, nil)
	assert.NotEqual(t, nil, err)
	assert.Equal(t, `{"errors":{"request":["Test Error"]}}`, err.Error())
}

// This tests a new service being creating in AppOptics
func TestNewServiceSyncSuccess(t *testing.T) {

	ts := v1.Status{ID: 0, LastUpdated: "Yesterday"}

	data := `
           {
                "type": "mail",
                "settings": {
                    "addresses": "SW-MSP-PE-Event-Bus-oncall@swmspdevops.opsgenie.net"
                },
                "title": "TEST"
           }
`

	td := v1.TokenAndDataSpec{Namespace: "Default", Data: data, Secret: "blah"}

	ts1, err := aoc.Sync(td, &ts, Service, nil)
	if err != nil {
		t.Errorf("error running TestExistingServiceSync: %v", err)
	}

	assert.NotEqual(t, 0, ts1.ID)
	assert.Equal(t, 145, ts1.ID)
}

// Test recreating a Service that has been deleting in AO but exists as a CRD still
func TestDeletedInAppopticsButNotInCRDServiceSyncSuccess(t *testing.T) {
	newID := 145

	ts := v1.Status{ID: testNotFoundId, LastUpdated: "Yesterday"}

	data := `
           {
                "type": "mail",
                "settings": {
                    "addresses": "SW-MSP-PE-Event-Bus-oncall@swmspdevops.opsgenie.net"
                },
                "title": "TEST"
           }`

	td := v1.TokenAndDataSpec{Namespace: "Default", Data: data, Secret: "blah"}

	ts1, err := aoc.Sync(td, &ts, Service, nil)
	if err != nil {
		t.Errorf("error running TestSpacesSync: %v", err)
	}

	// Because it doesn't exist it should change its ID to a new ID from AO
	assert.NotEqual(t, testNotFoundId, ts1.ID)
	assert.Equal(t, newID, ts1.ID)
}

// Test the case that a new Service fails to create in AO
func TestNewServiceCreateErrorInAppopticsFailure(t *testing.T) {

	ts := v1.Status{ID: 0, LastUpdated: "Yesterday"}

	data := `
           {
                "type": "mail",
                "settings": {
                    "addresses": "SW-MSP-PE-Event-Bus-oncall@swmspdevops.opsgenie.net"
                },
                "title": "NewServiceCreateError"
           }`

	td := v1.TokenAndDataSpec{Namespace: "Default", Data: data, Secret: "blah"}

	_, err := aoc.Sync(td, &ts, Service, nil)
	assert.NotEqual(t, nil, err)
	assert.Equal(t, `{"errors":{"request":["Test Error"]}}`, err.Error())
}

// Test recreating a Service that has been deleting in AO but exists as a CRD still but fails due to unknown AO Error
func TestMissingServiceCreateErrorInAppopticsFailure(t *testing.T) {

	ts := v1.Status{ID: testNotFoundId, LastUpdated: "Yesterday"}

	data := `
           {
                "type": "mail",
                "settings": {
                    "addresses": "SW-MSP-PE-Event-Bus-oncall@swmspdevops.opsgenie.net"
                },
                "title": "NewServiceCreateError"
           }`

	td := v1.TokenAndDataSpec{Namespace: "Default", Data: data, Secret: "blah"}

	_, err := aoc.Sync(td, &ts, Service, nil)
	assert.NotEqual(t, nil, err)
	assert.Equal(t, `{"errors":{"request":["Test Error"]}}`, err.Error())
}

// Test recreating a Service that has been deleting in AO but exists as a CRD still, then fails to create
func TestOutOfSyncServiceCreateErrorThenRetrieveErrorInAppoptics(t *testing.T) {

	ts := v1.Status{ID: testInternalServerErrorId, LastUpdated: "Yesterday"}

	data := `
           {
                "type": "mail",
                "settings": {
                    "addresses": "SW-MSP-PE-Event-Bus-oncall@swmspdevops.opsgenie.net"
                },
                "title": "NewServiceCreateError"
           }`
	td := v1.TokenAndDataSpec{Namespace: "Default", Data: data, Secret: "blah"}

	_, err := aoc.Sync(td, &ts, Service, nil)
	assert.NotEqual(t, nil, err)
	assert.Equal(t, `{"errors":{"request":["Test Error"]}}`, err.Error())
}

func TestDeletingServiceSuccessSync(t *testing.T) {
	err := aoc.Remove(1, Service)
	if err != nil {
		t.Errorf("error running TestSpacesSync: %v", err)
	}

	assert.Equal(t, nil, err)
}

func TestDeletingServiceErrorSync(t *testing.T) {
	err := aoc.Remove(testInternalServerErrorId, Service)

	assert.Equal(t, err.Error(), `{"errors":{"request":["Internal Server Error"]}}`)
}

func CreateServiceHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var service aoApi.Service
		err := JsonValidateAndDecode(r.Body, &service)

		if err != nil {
			http.Error(w, `{"errors":{"request":["Malformed Data"]}}`, http.StatusInternalServerError)
			return
		}

		if *service.Title == "NewServiceCreateError" {
			http.Error(w, `{"errors":{"request":["Test Error"]}}`, http.StatusInternalServerError)
			return
		}
		responseBody := `{
          "id": 145,
          "type": "campfire",
          "settings": {
            "room": "Ops",
            "token": "1234567890ABCDEF",
            "subdomain": "acme"
          },
            "title": "Notify Ops Room"
        }`
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(responseBody))
	}
}

func RetrieveServiceHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		ID := vars["serviceId"]

		if ID == strconv.Itoa(testNotFoundId) {
			http.Error(w, `{"errors":{"request":["Not Found"]}}`, http.StatusNotFound)
			return
		} else if strings.Compare(ID, strconv.Itoa(testInternalServerErrorId)) == 0 {
			http.Error(w, `{"errors":{"request":["Test Error"]}}`, http.StatusInternalServerError)
			return
		}
		responseBody := `{
  "id": 145,
  "type": "mail",
  "settings": {
    "addresses": "george@example.com,fred@example.com"
  },
  "title": "Notify Ops Room"
}`
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(responseBody))
	}
}

func UpdateServiceHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var service aoApi.Service
		err := JsonValidateAndDecode(r.Body, &service)

		if err != nil {
			http.Error(w, "Malformed Data", http.StatusInternalServerError)
		}

		if strings.Compare(*service.Title, "NewServiceError") == 0 {
			http.Error(w, `{"errors":{"request":["Test Error"]}}`, http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func DeleteServiceHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		ID, err := strconv.Atoi(vars["serviceId"])
		if err != nil {
			http.Error(w, `{"errors":{"request":["Internal Server Error"]}}`, http.StatusInternalServerError)
			return
		}
		if ID == testInternalServerErrorId {
			http.Error(w, `{"errors":{"request":["Internal Server Error"]}}`, http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		return
	}
}
