package web_server

import (
	"fmt"
	"log"
	"net/http"

	"github.ibm.com/almaden-containers/ubiquity/service_broker"
	"github.ibm.com/almaden-containers/ubiquity/resources"

	"github.com/gorilla/mux"
)

type Server struct {
	brokerApiHandler  *BrokerApiHandler
	storageApiHandler *StorageApiHandler
	logger            *log.Logger
}

func NewServer(logger *log.Logger, backends map[resources.Backend]resources.StorageClient, config resources.UbiquityServerConfig) (*Server, error) {
	brokerController := service_broker.NewController(nil, config.BrokerConfig.ConfigPath)
	return &Server{brokerApiHandler: NewBrokerApiHandler(logger, brokerController), storageApiHandler: NewStorageApiHandler(logger, backends), logger: logger}, nil
}

func (s *Server) InitializeHandler() http.Handler {
	router := mux.NewRouter()

	router.HandleFunc("/v2/catalog", s.brokerApiHandler.Catalog()).Methods("GET")
	router.HandleFunc("/v2/service_instances/{service_instance_guid}", s.brokerApiHandler.CreateServiceInstance()).Methods("PUT")
	router.HandleFunc("/v2/service_instances/{service_instance_guid}", s.brokerApiHandler.DeleteServiceInstance()).Methods("DELETE")
	router.HandleFunc("/v2/service_instances/{service_instance_guid}/service_bindings/{service_binding_guid}", s.brokerApiHandler.BindServiceInstance()).Methods("PUT")
	router.HandleFunc("/v2/service_instances/{service_instance_guid}/service_bindings/{service_binding_guid}", s.brokerApiHandler.UnbindServiceInstance()).Methods("DELETE")

	router.HandleFunc("/ubiquity_storage/{backend}/activate", s.storageApiHandler.Activate()).Methods("POST")
	router.HandleFunc("/ubiquity_storage/{backend}/volumes", s.storageApiHandler.CreateVolume()).Methods("POST")
	router.HandleFunc("/ubiquity_storage/{backend}/volumes", s.storageApiHandler.ListVolumes()).Methods("GET")
	router.HandleFunc("/ubiquity_storage/{backend}/volumes/{volume}", s.storageApiHandler.RemoveVolume()).Methods("DELETE")
	router.HandleFunc("/ubiquity_storage/{backend}/volumes/{volume}/attach", s.storageApiHandler.AttachVolume()).Methods("PUT")
	router.HandleFunc("/ubiquity_storage/{backend}/volumes/{volume}/detach", s.storageApiHandler.DetachVolume()).Methods("PUT")
	router.HandleFunc("/ubiquity_storage/{backend}/volumes/{volume}", s.storageApiHandler.GetVolume()).Methods("GET")

	return router
}

func (s *Server) Start(port int) error {
	router := s.InitializeHandler()
	http.Handle("/", router)

	fmt.Println(fmt.Sprintf("Starting server on port %d ....", port))
	fmt.Println("CTL-C to exit/stop ubiquity service")
	return http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
