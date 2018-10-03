package appoptics

import (
	"fmt"
	"net/http/httptest"
	"os"
	"testing"

	"encoding/json"
	"github.com/appoptics/appoptics-api-go"
	"github.com/gorilla/mux"
	"github.com/solarwinds/appoptics-kubernetes-controller/pkg/apis/appoptics-kubernetes-controller/v1"
	v12 "github.com/solarwinds/appoptics-kubernetes-controller/pkg/client/listers/appoptics-kubernetes-controller/v1"
	"io"
	"io/ioutil"
	v13 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
	"strings"
)

var (
	client *appoptics.Client
	server *httptest.Server
	aoc    *AOCommunicator
)

const testNotFoundId int = 9
const testInternalServerErrorId int = 8

func setup() {
	router := NewServerTestMux()
	server = httptest.NewServer(router)
	serverURLWithVersion := fmt.Sprintf("%s/v1/", server.URL)
	client = appoptics.NewClient("deadbeef", appoptics.BaseURLClientOption(serverURLWithVersion))
	aoc = &AOCommunicator{Client: *client}
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

	// Spaces
	router.Handle("/v1/spaces", CreateSpaceHandler()).Methods("POST")
	router.Handle("/v1/spaces/{id}", RetrieveSpaceHandler()).Methods("GET")
	router.Handle("/v1/spaces/{id}", UpdateSpaceHandler()).Methods("PUT")
	router.Handle("/v1/spaces/{id}", DeleteSpaceHandler()).Methods("DELETE")

	// Charts
	router.Handle("/v1/spaces/{spaceId}/charts", ListChartsHandler()).Methods("GET")
	router.Handle("/v1/spaces/{spaceId}/charts", CreateChartHandler()).Methods("POST")
	router.Handle("/v1/spaces/{spaceId}/charts/{chartId}", DeleteChartHandler()).Methods("DELETE")

	// Services
	router.Handle("/v1/services", CreateServiceHandler()).Methods("POST")
	router.Handle("/v1/services/{serviceId}", RetrieveServiceHandler()).Methods("GET")
	router.Handle("/v1/services/{serviceId}", UpdateServiceHandler()).Methods("PUT")
	router.Handle("/v1/services/{serviceId}", DeleteServiceHandler()).Methods("DELETE")

	// Alerts
	router.Handle("/v1/alerts", CreateAlertHandler()).Methods("POST")
	router.Handle("/v1/alerts/{alertId}", RetrieveAlertHandler()).Methods("GET")
	router.Handle("/v1/alerts/{alertId}", UpdateAlertHandler()).Methods("PUT")
	router.Handle("/v1/alerts/{alertId}", DeleteAlertHandler()).Methods("DELETE")
	router.Handle("/v1/alerts/{alertId}/services", AssociateAlertHandler()).Methods("POST")
	router.Handle("/v1/alerts/{alertId}/services/{serviceId}", DisassociateAlertHandler()).Methods("DELETE")

	return router
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

type mockAOCommunicator struct {
	Token string
}

func (maoc *mockAOCommunicator) Remove(ID int, kind string) error {
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

func (maoc *mockAOCommunicator) Sync(spec v1.TokenAndDataSpec, status *v1.Status, kind string, lister v12.ServiceNamespaceLister) (*v1.Status, error) {
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

func (msl *mockServiceLister) Get(name string) (*v1.Service, error) {
	tss := v1.Status{ID: 1}
	service := v1.Service{Status: tss}
	return &service, nil
}

func testIndexFunc(obj interface{}) ([]string, error) {
	pod := obj.(v13.Pod)
	return []string{pod.Labels["foo"]}, nil
}
func NewMockLister() *mockServiceLister {
	indexer := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{"testmodes": testIndexFunc})
	msl := mockServiceLister{namespace: "test", indexer: indexer}
	return &msl
}
