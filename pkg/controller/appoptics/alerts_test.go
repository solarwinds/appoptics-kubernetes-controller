package appoptics

import (
	aoApi "github.com/appoptics/appoptics-api-go"
	"github.com/gorilla/mux"
	"github.com/solarwinds/appoptics-kubernetes-controller/pkg/apis/appoptics-kubernetes-controller/v1"
	"github.com/stretchr/testify/assert"
	"net/http"
	"strconv"
	"strings"
	"testing"
)

const errorName = "Error"

func TestExistingAlertSyncSuccess(t *testing.T) {

	data := `
               {
        "name": "TEST",
        "description": "ActiveControllerCount",
        "conditions": [
            {
                "type": "below",
                "metric_name": "kafka.controller.KafkaController.ActiveControllerCount",
                "source": null,
                "threshold": 1,
                "duration": 60,
                "summary_function": "count"
            }
        ],
        "services": [],
        "attributes": {"services": ["example"]},
        "active": true,
        "rearm_seconds": 120
    }
`
	ts := v1.Status{ID: 3, LastUpdated: "Yesterday"}
	td := v1.TokenAndDataSpec{Namespace: "Default", Data: data, Secret: "blah"}

	ts1, err := aoc.Sync(td, &ts, Alert, NewMockLister())
	if err != nil {
		t.Errorf("error running TestExistingServiceSync: %v", err)
	}

	assert.Equal(t, ts1.ID, ts.ID)
}

func TestExistingAlertSyncWithNewServiceSuccess(t *testing.T) {

	data := `
               {
        "name": "TEST",
        "description": "ActiveControllerCount",
        "conditions": [
            {
                "type": "below",
                "metric_name": "kafka.controller.KafkaController.ActiveControllerCount",
                "source": null,
                "threshold": 1,
                "duration": 60,
                "summary_function": "count"
            }
        ],
        "services": [],
        "attributes": {"services": ["example"]},
        "active": true,
        "rearm_seconds": 120
    }
`
	ts := v1.Status{ID: 3, LastUpdated: "Yesterday"}
	td := v1.TokenAndDataSpec{Namespace: "Default", Data: data, Secret: "blah"}

	ts1, err := aoc.Sync(td, &ts, Alert, NewMockLister())
	if err != nil {
		t.Errorf("error running TestExistingServiceSync: %v", err)
	}

	assert.Equal(t, ts1.ID, ts.ID)
}

func TestExistingAlertNotInAppOpticsSyncSuccess(t *testing.T) {
	data := `
    {
     "name": "testName"
	}`
	alertSpec := v1.TokenAndDataSpec{Namespace: "Default", Data: data, Secret: ""}
	ts1, err := aoc.Sync(alertSpec, &v1.Status{ID: testNotFoundId}, Alert, NewMockLister())
	if err != nil {
		t.Errorf("error running TestExistingServiceSync: %v", err)
	}

	assert.NotEqual(t, ts1, testNotFoundId)
}

func TestExistingAlertNotInAppOpticsSyncFailure(t *testing.T) {
	data := `
    {
     "name": ` + errorName + `
	}`
	alertSpec := v1.TokenAndDataSpec{Namespace: "Default", Data: data, Secret: ""}
	_, err := aoc.Sync(alertSpec, &v1.Status{ID: testNotFoundId}, Alert, NewMockLister())
	assert.NotEqual(t, nil, err)
}

func TestNewAlertSyncSuccess(t *testing.T) {

	data := `
    {
     "name": "newAlert"
	}`
	alertSpec := v1.TokenAndDataSpec{Namespace: "Default", Data: data, Secret: ""}
	ID, err := aoc.Sync(alertSpec, &v1.Status{ID: 0}, Alert, NewMockLister())
	assert.Equal(t, nil, err)
	assert.NotEqual(t, 0, ID)
}

func TestNewAlertSyncFailure(t *testing.T) {
	data := `
    {
     "name": "Error"
	}`
	alertSpec := v1.TokenAndDataSpec{Namespace: "Default", Data: data, Secret: ""}
	_, err := aoc.Sync(alertSpec, &v1.Status{ID: 0}, Alert, NewMockLister())
	assert.NotEqual(t, nil, err)
}

func TestDeletingAlertSuccessSync(t *testing.T) {

	err := aoc.Remove(0, Alert)
	if err != nil {
		t.Errorf("error running TestSpacesSync: %v", err)
	}

	assert.Equal(t, nil, err)
}

func TestDeletingAlertErrorSync(t *testing.T) {
	err := aoc.Remove(testInternalServerErrorId, Alert)
	assert.Equal(t, err.Error(), `{"errors":{"request":["Internal Server Error"]}}`)
}

func CreateAlertHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var alert aoApi.Alert
		err := JsonValidateAndDecode(r.Body, &alert)

		if err != nil {
			http.Error(w, "Malformed Data", http.StatusInternalServerError)
		}

		if strings.Compare(*alert.Name, "Error") == 0 {
			http.Error(w, `{"errors":{"request":["Internal Server Error"]}}`, http.StatusInternalServerError)
			return
		}
		responseBody := `{
   "id":1234567,
   "name":"production.web.frontend.response_time",
   "description":"Web Response Time",
   "conditions":[
      {
         "id":19376030,
         "type":"above",
         "metric_name":"web.nginx.response_time",
         "threshold":200.0,
         "summary_function":"max",
         "tags":[
            {
               "name":"tag_name",
               "grouped":false,
               "values":[
                  "tag_value"
               ]
            }
         ]
      }
   ],
   "services":[
      {
         "id":17584,
         "type":"slack",
         "settings":{
            "url":"https://hooks.slack.com/services/ABCDEFG/A1B2C3/asdfg1234"
         },
         "title":"librato-services"
      }
   ],
   "attributes":{
      "runbook_url":"http://myco.com/runbooks/response_time"
   },
   "active":true,
   "created_at":1484594787,
   "updated_at":1484594787,
   "version":2,
   "rearm_seconds":600,
   "rearm_per_signal":false,
   "md":true
}`
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(responseBody))
	}
}

func RetrieveAlertHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		ID := vars["alertId"]
		if ID == strconv.Itoa(testNotFoundId) {
			http.Error(w, `{"errors":{"request":["Not Found"]}}`, http.StatusNotFound)
			return
		}
		responseBody := `{
  "id": 123,
  "name": "production.web.frontend.response_time",
  "description":"Web Response Time",
  "conditions":[
      {
         "id":19375969,
         "type":"above",
         "metric_name":"web.nginx.response_time",
         "source":null,
         "threshold":200.0,
         "summary_function":"average",
         "tags":[
            {
               "name":"environment",
               "grouped":false,
               "values":[
                  "production"
               ]
            }
         ]
      }
   ],
  "services":[
      {
         "id":1,
         "type":"slack",
         "settings":{
            "url":"https://hooks.slack.com/services/XYZABC/a1b2c3/asdf"
         },
         "title":"appoptics-services"
      }
      
   ],
  "attributes": {
    "runbook_url": "http://myco.com/runbooks/response_time"
  },
  "active":true,
  "created_at":1484588756,
  "updated_at":1484588756,
  "version":2,
  "rearm_seconds":600,
  "rearm_per_signal":false,
  "md":true
}`
		w.Write([]byte(responseBody))
	}
}

func UpdateAlertHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		ID := vars["alertId"]
		if ID == strconv.Itoa(testInternalServerErrorId) {
			http.Error(w, `{"errors":{"request":["Not Found"]}}`, http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func DeleteAlertHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		ID, err := strconv.Atoi(vars["alertId"])
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

func DisassociateAlertHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}
}
func AssociateAlertHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}
}
