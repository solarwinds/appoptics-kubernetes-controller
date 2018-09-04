package checks

import (
	"fmt"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/appoptics/appoptics-api-go"
	"github.com/gorilla/mux"
)

var (
	client *appoptics.Client
	server *httptest.Server
)

const testNotFoundId int = 9
const testInternalServerErrorId int = 8

func setup() {
	router := NewServerTestMux()
	server = httptest.NewServer(router)
	serverURLWithVersion := fmt.Sprintf("%s/v1/", server.URL)
	client = appoptics.NewClient("deadbeef", appoptics.BaseURLClientOption(serverURLWithVersion))
}

func teardown() {
	server.Close()
}

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	teardown()
	os.Exit(code)
}

func NewServerTestMux() *mux.Router {
	router := mux.NewRouter()

	// Measurements

	// Metrics

	// Spaces
	router.Handle("/v1/spaces", CreateSpaceHandler()).Methods("POST")
	router.Handle("/v1/spaces/{id}", RetrieveSpaceHandler()).Methods("GET")
	router.Handle("/v1/spaces/{id}", UpdateSpaceHandler()).Methods("PUT")

	// Charts
	router.Handle("/v1/spaces/{spaceId}/charts", ListChartsHandler()).Methods("GET")
	router.Handle("/v1/spaces/{spaceId}/charts", CreateChartHandler()).Methods("POST")
	router.Handle("/v1/spaces/{spaceId}/charts/{chartId}", DeleteChartHandler()).Methods("DELETE")
	//
	//// Services
	//router.Handle("/v1/services", ListServicesHandler()).Methods("GET")
	router.Handle("/v1/services", CreateServiceHandler()).Methods("POST")
	router.Handle("/v1/services/{serviceId}", RetrieveServiceHandler()).Methods("GET")
	router.Handle("/v1/services/{serviceId}", UpdateServiceHandler()).Methods("PUT")
	//router.Handle("/v1/services/{serviceId}", DeleteServiceHandler()).Methods("DELETE")
	//
	//// Annotations
	//
	//// Alerts
	//router.Handle("/v1/alerts", ListAlertsHandler()).Methods("GET")
	router.Handle("/v1/alerts", CreateAlertHandler()).Methods("POST")
	router.Handle("/v1/alerts/{alertId}", RetrieveAlertHandler()).Methods("GET")
	router.Handle("/v1/alerts/{alertId}", UpdateAlertHandler()).Methods("PUT")
	//router.Handle("/v1/alerts/{alertId}", DeleteAlertHandler()).Methods("DELETE")
	//router.Handle("/v1/alerts/{alertId}/status", StatusAlertHandler()).Methods("GET")
	//
	//// API Tokens
	//router.Handle("/v1/api_tokens", ListApiTokensHandler()).Methods("GET")
	//router.Handle("/v1/api_tokens", CreateApiTokenHandler()).Methods("POST")
	//router.Handle("/v1/api_tokens/{tokenName}", RetrieveApiTokenHandler()).Methods("GET")
	//router.Handle("/v1/api_tokens/{tokenId}", UpdateApiTokenHandler()).Methods("PUT")
	//router.Handle("/v1/api_tokens/{tokenId}", DeleteApiTokenHandler()).Methods("DELETE")

	// Jobs

	// Snapshots

	return router
}
