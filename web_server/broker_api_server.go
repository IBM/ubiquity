package web_server

import (
	"fmt"
	"log"
	"net/http"

	"github.ibm.com/almaden-containers/ubiquity/resources"
	"github.ibm.com/almaden-containers/ubiquity/service_broker"

	"github.com/gorilla/mux"
)

type BrokerApiServer struct {
	brokerApiHandler *BrokerApiHandler
	logger           *log.Logger
}

func NewBrokerApiServer(logger *log.Logger, backends map[resources.Backend]resources.StorageClient, config resources.UbiquityServerConfig) (*BrokerApiServer, error) {
	brokerController := service_broker.NewController(nil, config.BrokerConfig.ConfigPath)
	return &BrokerApiServer{brokerApiHandler: NewBrokerApiHandler(logger, brokerController), logger: logger}, nil
}

func (s *BrokerApiServer) InitializeHandler() http.Handler {
	router := mux.NewRouter()
	router.HandleFunc("/v2/catalog", s.brokerApiHandler.Catalog()).Methods("GET")
	router.HandleFunc("/v2/service_instances/{service_instance_guid}", s.brokerApiHandler.CreateServiceInstance()).Methods("PUT")
	router.HandleFunc("/v2/service_instances/{service_instance_guid}", s.brokerApiHandler.DeleteServiceInstance()).Methods("DELETE")
	router.HandleFunc("/v2/service_instances/{service_instance_guid}/service_bindings/{service_binding_guid}", s.brokerApiHandler.BindServiceInstance()).Methods("PUT")
	router.HandleFunc("/v2/service_instances/{service_instance_guid}/service_bindings/{service_binding_guid}", s.brokerApiHandler.UnbindServiceInstance()).Methods("DELETE")
	return router
}

func (s *BrokerApiServer) Start(port int) error {
	router := s.InitializeHandler()
	http.Handle("/", router)

	fmt.Println(fmt.Sprintf("Starting Broker API server on port %d ....", port))
	fmt.Println("CTL-C to exit/stop Broker API service")
	return http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
