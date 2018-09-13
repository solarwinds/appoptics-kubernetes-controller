package appoptics

import (
	"net/http"
	"testing"
	"github.com/appoptics/appoptics-kubernetes-controller/pkg/apis/appoptics-kubernetes-controller/v1"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
	v13 "k8s.io/api/core/v1"
	"github.com/gorilla/mux"
	"strconv"
	"strings"
)

// serviceNamespaceLister implements the ServiceNamespaceLister
// interface.
type mockServiceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all Services in the indexer for a given namespace.
func (s mockServiceLister) List(selector labels.Selector) (ret []*v1.Service, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1.Service))
	})
	return ret, err
}

func (msl *mockServiceLister) Get (name string) (*v1.Service, error){
	tss := v1.TimestampAndIdStatus{ID:1}
	service := v1.Service{Status:tss}
	return &service, nil
}

func testIndexFunc(obj interface{}) ([]string, error) {
	pod := obj.(v13.Pod)
	return []string{pod.Labels["foo"]}, nil
}
const errorName = "Error"

func TestExistingAlertSyncSuccess(t *testing.T){

	indexer := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{"testmodes": testIndexFunc})

	msl := mockServiceLister{namespace:"test", indexer:indexer}

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
	ts := v1.TimestampAndIdStatus{ID: 3, LastUpdated: "Yesterday"}
	td := v1.TokenAndDataSpec{Namespace: "Default", Data: data, Token: "blah"}

	ts1, err := syncronizer.SyncAlert(td, &ts, &msl)
	if err != nil {
		t.Errorf("error running TestExistingServiceSync: %v", err)
	}

	assert.Equal(t, ts1.ID, ts.ID)
}

func TestExistingAlertNotInAppOpticsSyncSuccess(t *testing.T){
	testName := "testName"
	ts1, err := syncronizer.syncAlert(AlertRequest{Name:&testName}, testNotFoundId)
	if err != nil {
		t.Errorf("error running TestExistingServiceSync: %v", err)
	}

	assert.NotEqual(t, ts1, testNotFoundId)
}

func TestExistingAlertNotInAppOpticsSyncFailure(t *testing.T){
	testErrorName := errorName
	_, err := syncronizer.syncAlert(AlertRequest{Name:&testErrorName}, testNotFoundId)
	assert.NotEqual(t, nil, err)
}


func TestNewAlertSyncSuccess(t *testing.T){
	testName := "newAlert"
	ID, err := syncronizer.syncAlert(AlertRequest{Name:&testName}, 0)
	assert.Equal(t, nil, err)
	assert.NotEqual(t, 0, ID)
}

func TestNewAlertSyncFailure(t *testing.T){
	testName := "Error"
	ID, err := syncronizer.syncAlert(AlertRequest{Name:&testName}, 0)
	assert.NotEqual(t, nil, err)
	assert.Equal(t, -1, ID)
}

func TestDeletingAlertSuccessSync(t *testing.T) {
	err := syncronizer.RemoveAlert(1)
	if err != nil {
		t.Errorf("error running TestSpacesSync: %v", err)
	}

	assert.Equal(t, nil, err)
}

func TestDeletingAlertErrorSync(t *testing.T) {
	err := syncronizer.RemoveAlert(testInternalServerErrorId)

	assert.Equal(t, err.Error(), `{"errors":{"request":["Internal Server Error"]}}`)
}

func CreateAlertHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var alertRequest AlertRequest
		err := JsonValidateAndDecode(r.Body, &alertRequest)

		if err != nil {
			http.Error(w, "Malformed Data", http.StatusInternalServerError)
		}

		if strings.Compare(*alertRequest.Name, "Error") == 0 {
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
         "id":17584,
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
		return;
	}
}