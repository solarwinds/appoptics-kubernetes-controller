package appoptics

import (
	"github.com/solarwinds/appoptics-kubernetes-controller/pkg/apis/appoptics-kubernetes-controller/v1"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"net/http"
	"strconv"
	"strings"
	"testing"
)

const spaceError = "spaceError"

type SimpleSpace struct {
	Name string `json:"name","omitempty"`
}

func TestSpacesService_Create(t *testing.T) {
	space, err := NewSpacesService(client).Create("CPUs")
	if err != nil {
		t.Errorf("error running Create: %v", err)
	}

	assert.Equal(t, space.Name, "CPUs")
	assert.Equal(t, 1, space.ID)

}

func TestExistingSpacesSync(t *testing.T) {
	ts := v1.Status{ID: 1, LastUpdated: "Yesterday"}

	data := `
---
name: DevOps Alerts
charts:
`

	td := v1.TokenAndDataSpec{Namespace: "Default", Data: data, Secret: "blah"}

	ts1, err := aoc.Sync(td, &ts, Dashboard, nil)
	if err != nil {
		t.Errorf("error running TestSpacesSync: %v", err)
	}

	assert.Equal(t, ts1.ID, ts.ID)
}

func TestNewSpacesSync(t *testing.T) {

	ts := v1.Status{ID: 0, LastUpdated: "Yesterday"}

	data := `
---
name: DevOps Alerts
charts:
`

	td := v1.TokenAndDataSpec{Namespace: "Default", Data: data, Secret: "blah"}

	ts1, err := aoc.Sync(td, &ts, Dashboard, nil)
	if err != nil {
		t.Errorf("error running TestSpacesSync: %v", err)
	}

	assert.Equal(t, 1, ts1.ID)
}

func TestDeletedInAppopticsButNotInCRDSpacesSync(t *testing.T) {
	invalidID := testNotFoundId
	newID := 1
	ts := v1.Status{ID: invalidID, LastUpdated: "Yesterday"}

	data := `
---
name: DevOps Alerts
charts:
`

	td := v1.TokenAndDataSpec{Namespace: "Default", Data: data, Secret: "blah"}

	ts1, err := aoc.Sync(td, &ts, Dashboard, nil)
	if err != nil {
		t.Errorf("error running TestSpacesSync: %v", err)
	}

	// Because it doesn't exist it should change its ID to a new ID from AO
	assert.NotEqual(t, invalidID, ts1.ID)
	assert.Equal(t, newID, ts1.ID)
}

func TestNewSpaceCreateErrorInAppoptics(t *testing.T) {
	ts := v1.Status{ID: 0, LastUpdated: "Yesterday"}

	data := `
---
name: ` + spaceError + `
charts:
`

	td := v1.TokenAndDataSpec{Namespace: "Default", Data: data, Secret: "blah"}

	_, err := aoc.Sync(td, &ts, Dashboard, nil)
	assert.NotEqual(t, nil, err)
	assert.Equal(t, `{"errors":{"request":["Internal Server Error"]}}`, err.Error())
}

func TestOutOfSyncSpaceCreateErrorInAppoptics(t *testing.T) {
	ts := v1.Status{ID: testNotFoundId, LastUpdated: "Yesterday"}

	data := `
---
name: ` + spaceError + `
charts:
`
	td := v1.TokenAndDataSpec{Namespace: "Default", Data: data, Secret: "blah"}

	_, err := aoc.Sync(td, &ts, Dashboard, nil)
	assert.NotEqual(t, nil, err)
	assert.Equal(t, `{"errors":{"request":["Internal Server Error"]}}`, err.Error())
}

func TestOutOfSyncSpaceCreateErrorThenRetrieveErrorInAppoptics(t *testing.T) {
	ts := v1.Status{ID: testInternalServerErrorId, LastUpdated: "Yesterday"}

	data := `
---
name: ` + spaceError + `
charts:
`
	td := v1.TokenAndDataSpec{Namespace: "Default", Data: data, Secret: "blah"}

	_, err := aoc.Sync(td, &ts, Dashboard, nil)
	assert.NotEqual(t, nil, err)
	assert.Equal(t, `{"errors":{"request":["Internal Server Error"]}}`, err.Error())
}

func TestUpdateSpacesErrorsSync(t *testing.T) {

	ts := v1.Status{ID: 3, LastUpdated: "Yesterday"}

	data := `
---
name: ` + spaceError + `
charts:
`

	td := v1.TokenAndDataSpec{Namespace: "Default", Data: data, Secret: "blah"}

	_, err := aoc.Sync(td, &ts, Dashboard, nil)
	assert.NotEqual(t, nil, err)
	assert.Equal(t, `{"errors":{"request":["`+spaceError+`"]}}`, err.Error())
}

func TestExistingSpacesFailsChartSync(t *testing.T) {
	ts := v1.Status{ID: 1, LastUpdated: "Yesterday"}
	data := `
---
name: DevOps Alerts
charts:
- name: I am a test chart
  id: ` + strconv.Itoa(testInternalServerErrorId) + `
  type: line
  streams:
  - summary_function: average
    downsample_function: average
    tags:
    - name: "@source"
      dynamic: true
    composite: |
      s("rainy.days.are.bad", {})`

	td := v1.TokenAndDataSpec{Namespace: "Default", Data: data, Secret: "blah"}

	_, err := aoc.Sync(td, &ts, Dashboard, nil)

	assert.Equal(t, err.Error(), `{"errors":{"request":["Internal Server Error"]}}`)
}

func TestDeletingSpaceSuccessSync(t *testing.T) {
	err := aoc.Remove(1, Dashboard)
	if err != nil {
		t.Errorf("error running TestSpacesSync: %v", err)
	}

	assert.Equal(t, nil, err)
}

func TestDeletingSpaceErrorSync(t *testing.T) {
	err := aoc.Remove(testInternalServerErrorId, Dashboard)
	assert.Equal(t, err.Error(), `{"errors":{"request":["Internal Server Error"]}}`)
}

func CreateSpaceHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var simpleSpace SimpleSpace
		err := JsonValidateAndDecode(r.Body, &simpleSpace)

		if err != nil {
			http.Error(w, "Malformed Data", http.StatusInternalServerError)
		}

		if strings.Compare(simpleSpace.Name, spaceError) == 0 {
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

		if strings.Compare(simpleSpace.Name, spaceError) == 0 {
			http.Error(w, `{"errors":{"request":["`+spaceError+`"]}}`, http.StatusInternalServerError)
			return
		}
		responseBody := `{
  "name": "MEMORY"
}`

		w.Write([]byte(responseBody))
	}
}

func DeleteSpaceHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		ID, err := strconv.Atoi(vars["id"])
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
