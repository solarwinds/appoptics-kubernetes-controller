package checks

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/appoptics/appoptics-kubernetes-controller/pkg/apis/appoptics-kubernetes-controller/v1"
	"github.com/stretchr/testify/assert"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"strconv"
)

const newSpaceError = "NewSpaceError"
const updateSpaceError = "UpdateSpaceError"

func TestSpacesService_Create(t *testing.T) {
	space, err := NewSpacesService(client).create("CPUs")
	if err != nil {
		t.Errorf("error running Create: %v", err)
	}

	assert.Equal(t, space.Name, "CPUs")
	assert.Equal(t, 1, space.ID)

}

func TestExistingSpacesSync(t *testing.T) {
	rts := ResourcesToSync{client}
	ts := v1.TimestampAndIdStatus{ID: 1, LastUpdated: "Yesterday"}

	data := `
---
name: DevOps Alerts
charts:
`

	td := v1.TokenAndDataSpec{Namespace: "Default", Data: data, Token: "blah"}

	ts1, err := rts.SyncDashboard(td, &ts)
	if err != nil {
		t.Errorf("error running TestSpacesSync: %v", err)
	}

	assert.Equal(t, ts1.ID, ts.ID)
}

func TestNewSpacesSync(t *testing.T) {

	rts := ResourcesToSync{client}
	ts := v1.TimestampAndIdStatus{ID: 0, LastUpdated: "Yesterday"}

	data := `
---
name: DevOps Alerts
charts:
`

	td := v1.TokenAndDataSpec{Namespace: "Default", Data: data, Token: "blah"}

	ts1, err := rts.SyncDashboard(td, &ts)
	if err != nil {
		t.Errorf("error running TestSpacesSync: %v", err)
	}

	assert.Equal(t, 1, ts1.ID)
}

func TestDeletedInAppopticsButNotInCRDSpacesSync(t *testing.T) {
	invalidID := testNotFoundId
	newID := 1
	rts := ResourcesToSync{client}
	ts := v1.TimestampAndIdStatus{ID: invalidID, LastUpdated: "Yesterday"}

	data := `
---
name: DevOps Alerts
charts:
`

	td := v1.TokenAndDataSpec{Namespace: "Default", Data: data, Token: "blah"}

	ts1, err := rts.SyncDashboard(td, &ts)
	if err != nil {
		t.Errorf("error running TestSpacesSync: %v", err)
	}

	// Because it doesn't exist it should change its ID to a new ID from AO
	assert.NotEqual(t, invalidID, ts1.ID)
	assert.Equal(t, newID, ts1.ID)
}

func TestNewSpaceCreateErrorInAppoptics(t *testing.T) {
	rts := ResourcesToSync{client}
	ts := v1.TimestampAndIdStatus{ID: 0, LastUpdated: "Yesterday"}

	data := `
---
name: `+newSpaceError+`
charts:
`

	td := v1.TokenAndDataSpec{Namespace: "Default", Data: data, Token: "blah"}

	_, err := rts.SyncDashboard(td, &ts)
	assert.NotEqual(t, nil, err)
	assert.Equal(t, `{"errors":{"request":["Internal Server Error"]}}`, err.Error())
}

func TestOutOfSyncSpaceCreateErrorInAppoptics(t *testing.T) {
	rts := ResourcesToSync{client}
	ts := v1.TimestampAndIdStatus{ID: testNotFoundId, LastUpdated: "Yesterday"}

	data := `
---
name: `+newSpaceError+`
charts:
`
	td := v1.TokenAndDataSpec{Namespace: "Default", Data: data, Token: "blah"}

	_, err := rts.SyncDashboard(td, &ts)
	assert.NotEqual(t, nil, err)
	assert.Equal(t, `{"errors":{"request":["Internal Server Error"]}}`, err.Error())
}

func TestOutOfSyncSpaceCreateErrorThenRetrieveErrorInAppoptics(t *testing.T) {
	rts := ResourcesToSync{client}
	ts := v1.TimestampAndIdStatus{ID: testInternalServerErrorId, LastUpdated: "Yesterday"}

	data := `
---
name: `+newSpaceError+`
charts:
`
	td := v1.TokenAndDataSpec{Namespace: "Default", Data: data, Token: "blah"}

	_, err := rts.SyncDashboard(td, &ts)
	assert.NotEqual(t, nil, err)
	assert.Equal(t, `{"errors":{"request":["Internal Server Error"]}}`, err.Error())
}

func TestUpdateSpacesErrorsSync(t *testing.T) {

	rts := ResourcesToSync{client}
	ts := v1.TimestampAndIdStatus{ID: 3, LastUpdated: "Yesterday"}

	data := `
---
name: `+updateSpaceError+`
charts:
`

	td := v1.TokenAndDataSpec{Namespace: "Default", Data: data, Token: "blah"}

	_, err := rts.SyncDashboard(td, &ts)
	assert.NotEqual(t, nil, err)
	assert.Equal(t, `{"errors":{"request":["`+updateSpaceError+`"]}}`, err.Error())
}

func TestExistingSpacesFailsChartSync(t *testing.T) {
	rts := ResourcesToSync{client}
	ts := v1.TimestampAndIdStatus{ID: 1, LastUpdated: "Yesterday"}

	data := `
---
name: DevOps Alerts
charts:
- name: I am a test chart
  id: `+strconv.Itoa(testInternalServerErrorId)+`
  type: line
  streams:
  - summary_function: average
    downsample_function: average
    tags:
    - name: "@source"
      dynamic: true
    composite: |
      s("rainy.days.are.bad", {})`


	td := v1.TokenAndDataSpec{Namespace: "Default", Data: data, Token: "blah"}

	_, err := rts.SyncDashboard(td, &ts)

	assert.Equal(t, err.Error(), `{"errors":{"request":["Internal Server Error"]}}`)
}

func CreateSpaceHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var simpleSpace SimpleSpace
		err := JsonValidateAndDecode(r.Body, &simpleSpace)

		if err != nil {
			http.Error(w, "Malformed Data", http.StatusInternalServerError)
		}

		if strings.Compare(simpleSpace.Name, newSpaceError) == 0 {
			http.Error(w, `{"errors":{"request":["Internal Server Error"]}}`, http.StatusInternalServerError)
			return
		}

		responseBody := `{
		  "id": 1,
		  "name": "CPUs"
			}`
		w.Header().Add("Location", "/a/b/1")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(responseBody))
	}
}

func RetrieveSpaceHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		ID, err := strconv.Atoi(vars["id"])
		if err != nil {
			http.Error(w, `{"errors":{"request":["ID Conversion Error"]}}`, http.StatusInternalServerError)
			return
		}
		if ID == testNotFoundId {
			http.Error(w, `{"errors":{"request":["Not Found"]}}`, http.StatusNotFound)
			return
		} else if ID == testInternalServerErrorId {
			http.Error(w, `{"errors":{"request":["Internal Server Error"]}}`, http.StatusInternalServerError)
			return
		}
			responseBody := `{
  "name": "CPUs",
  "id": ` + vars["id"] + `,
  "charts": [
    {
      "id": 915
    },
    {
      "id": 1321
    },
    {
      "id": 47842
    },
    {
      "id": 922
    }
  ]
}`
			w.Write([]byte(responseBody))


	}
}

func UpdateSpaceHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var simpleSpace SimpleSpace
		err := JsonValidateAndDecode(r.Body, &simpleSpace)

		if err != nil {
			http.Error(w, `{"errors":{"request":["Internal Server Error"]}}`, http.StatusInternalServerError)
			return
		}

		if strings.Compare(simpleSpace.Name, updateSpaceError) == 0 {
			http.Error(w, `{"errors":{"request":["`+updateSpaceError+`"]}}`, http.StatusInternalServerError)
			return
		}
		responseBody := `{
  "name": "MEMORY"
}`

		w.Write([]byte(responseBody))
	}
}

func JsonValidateAndDecode(body io.ReadCloser, v interface{}) error {
	//Read the body into a buffer to see if it is empty
	if body == nil {
		//Null or empty body will error on decode but it is not invalid
		return nil
	}

	buff, err := ioutil.ReadAll(body)
	if err != nil {
		return err
	}

	//If the payload is an empty string it will error on decode but is not invalid
	if len(buff) == 0 {
		return nil
	}

	//body is empty so create a new reader from the buffer
	decoder := json.NewDecoder(strings.NewReader(string(buff)))

	err = decoder.Decode(&v)
	if err != nil {
		return err
	}
	defer body.Close()

	//No errors occurred
	return nil
}
