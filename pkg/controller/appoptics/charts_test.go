package appoptics

import (
	"encoding/base64"
	"net/http"
	"strconv"
	"testing"

	aoApi "github.com/appoptics/appoptics-api-go"
	"github.com/ghodss/yaml"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestDeletingChartsInAppOptics(t *testing.T) {

	chartsService := NewChartsService(client)

	data := `

                - name: I am a test chart
                  id: 1
                  type: line
                  streams:
                  - summary_function: average
                    downsample_function: average
                    tags:
                    - name: "@source"
                      dynamic: true
                    composite: |
                      s("lovely.sunny.days.are.nice", {})`

	var charts []*aoApi.Chart
	err := yaml.Unmarshal([]byte(data), &charts)
	if err != nil {
		t.Errorf("error running TestSpacesSync: %v", err)
	}

	err = chartsService.DeleteAll(charts, 0)
	if err != nil {
		t.Errorf("error running TestSpacesSync: %v", err)
	}

	assert.Equal(t, nil, err)
}

func TestFailDeletingChartsInAppOptics(t *testing.T) {

	chartsService := NewChartsService(client)

	data := `

                - name: I am a test chart
                  id: ` + strconv.Itoa(testNotFoundId) + `
                  type: line
                  streams:
                  - summary_function: average
                    downsample_function: average
                    tags:
                    - name: "@source"
                      dynamic: true
                    composite: |
                      s("rainy.days.are.bad", {})`

	var charts []*aoApi.Chart
	err := yaml.Unmarshal([]byte(data), &charts)
	if err != nil {
		t.Errorf("error running TestSpacesSync: %v", err)
	}

	err = chartsService.DeleteAll(charts, 0)
	assert.Equal(t, `{"errors":{"request":["Not Found"]}}`, err.Error())
}

func TestSyncingChartsWithAppOptics(t *testing.T) {
	data := `

                - name: I am a test chart
                  id: 1
                  type: line
                  streams:
                  - summary_function: average
                    downsample_function: average
                    tags:
                    - name: "@source"
                      dynamic: true
                    composite: |
                      s("rainy.days.are.bad", {})`
	var charts []*aoApi.Chart
	err := yaml.Unmarshal([]byte(data), &charts)
	if err != nil {
		t.Errorf("error running TestSpacesSync: %v", err)
	}
	chartService := NewChartsService(client)
	err = chartService.syncCharts(charts, 0)
	if err != nil {
		t.Errorf("error running TestSpacesSync: %v", err)
	}

	assert.Equal(t, nil, err)
}

func TestSyncingChartsWithAppOpticsListErrorResponse(t *testing.T) {
	data := `

                - name: I am a test chart
                  id: 1
                  type: line
                  streams:
                  - summary_function: average
                    downsample_function: average
                    tags:
                    - name: "@source"
                      dynamic: true
                    composite: |
                      s("rainy.days.are.bad", {})`
	var charts []*aoApi.Chart
	err := yaml.Unmarshal([]byte(data), &charts)
	if err != nil {
		t.Errorf("error running TestSpacesSync: %v", err)
	}
	chartService := NewChartsService(client)
	err = chartService.syncCharts(charts, testNotFoundId)
	assert.NotEqual(t, nil, err)
	assert.Equal(t, `{"errors":{"request":["Test Error"]}}`, err.Error())
}

func TestSyncingChartsHash(t *testing.T) {

	chartsService := NewChartsService(client)
	hash, err := chartsService.getChartHash(1)
	if err != nil {
		t.Errorf("error running TestSpacesSync: %v", err)
	}

	assert.Equal(t, "l5DfpfwQV8AoEflpgdpxRcl7WWY=", base64.StdEncoding.EncodeToString(hash))

}
func TestSyncingChartsWithAppOpticsDeletingOldChartsErrorResponse(t *testing.T) {
	data := `

                - name: I am a test chart
                  id: ` + strconv.Itoa(testNotFoundId) + `
                  type: line
                  streams:
                  - summary_function: average
                    downsample_function: average
                    tags:
                    - name: "@source"
                      dynamic: true
                    composite: |
                      s("rainy.days.are.bad", {})`
	var charts []*aoApi.Chart
	err := yaml.Unmarshal([]byte(data), &charts)
	if err != nil {
		t.Errorf("error running TestSpacesSync: %v", err)
	}
	chartService := NewChartsService(client)
	err = chartService.syncCharts(charts, testInternalServerErrorId)
	assert.NotEqual(t, nil, err.Error())
	assert.Equal(t, `{"errors":{"request":["Internal Server Error"]}}`, err.Error())

}

func TestFailCreatingChartsInAppOptics(t *testing.T) {

	chartsService := NewChartsService(client)

	data := `

                - name: I am a test chart
                  id: ` + strconv.Itoa(testNotFoundId) + `
                  type: line
                  streams:
                  - summary_function: average
                    downsample_function: average
                    tags:
                    - name: "@source"
                      dynamic: true
                    composite: |
                      s("rainy.days.are.bad", {})`

	var charts []*aoApi.Chart
	err := yaml.Unmarshal([]byte(data), &charts)
	if err != nil {
		t.Errorf("error running TestSpacesSync: %v", err)
	}

	err = chartsService.DeleteAll(charts, 0)
	assert.Equal(t, `{"errors":{"request":["Not Found"]}}`, err.Error())
}

func ListChartsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		ID := vars["spaceId"]
		if ID == strconv.Itoa(testNotFoundId) {
			http.Error(w, `{"errors":{"request":["Test Error"]}}`, http.StatusInternalServerError)
			return
		}
		responseBody := `[
 {
   "id": ` + ID + `,
   "name": "CPU Usage",
   "type": "line",
   "streams": [
     {
       "id": 27035309,
       "metric": "cpu.percent.idle",
       "type": "gauge",
       "tags": [
         {
           "name": "environment",
           "values": [
             "*"
           ]
         }
       ]
     },
     {
       "id": 27035310,
       "metric": "cpu.percent.user",
       "type": "gauge",
       "tags": [
         {
           "name": "environment",
           "values": [
             "prod"
           ]
         }
       ]
     }
   ],
   "thresholds": null
 }
]`
		w.Write([]byte(responseBody))
	}
}

func CreateChartHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var chart aoApi.Chart
		err := JsonValidateAndDecode(r.Body, &chart)

		if err != nil {
			http.Error(w, `{"errors":{"request":["Malformed Data"]}}`, http.StatusInternalServerError)
			return
		}

		if chart.ID == testInternalServerErrorId {
			http.Error(w, `{"errors":{"request":["Internal Server Error"]}}`, http.StatusInternalServerError)
			return
		}
		responseBody := `{
 "id": ` + strconv.Itoa(testNotFoundId) + `,
 "name": "CPU Usage",
 "type": "line",
 "streams": [
   {
     "id": ` + strconv.Itoa(testNotFoundId) + `,
     "metric": "cpu.percent.idle",
     "type": "gauge",
     "tags": [
       {
         "name": "environment",
         "values": [
           "*"
         ]
       }
     ]
   },
   {
     "id": 27032886,
     "metric": "cpu.percent.user",
     "type": "gauge",
     "tags": [
       {
         "name": "environment",
         "values": [
           "prod"
         ]
       }
     ]
   }
 ],
 "thresholds": null
}`
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(responseBody))
	}
}

func DeleteChartHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		ID, err := strconv.Atoi(vars["chartId"])
		if err != nil {
			http.Error(w, `{"errors":{"request":["Malformed ID"]}}`, http.StatusNotFound)
			return
		}
		if ID == testNotFoundId {
			http.Error(w, `{"errors":{"request":["Not Found"]}}`, http.StatusNotFound)
			return
		} else if ID == testInternalServerErrorId {
			http.Error(w, `{"errors":{"request":["Internal Server Error"]}}`, http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
