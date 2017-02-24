package web_server

import (
	"fmt"
	"log"
	"net/http"

	"github.ibm.com/almaden-containers/ubiquity/resources"

	"github.com/gorilla/mux"
)

type StorageApiServer struct {
	storageApiHandler *StorageApiHandler
	logger            *log.Logger
}

func NewStorageApiServer(logger *log.Logger, backends map[resources.Backend]resources.StorageClient, config resources.UbiquityServerConfig) (*StorageApiServer, error) {
	return &StorageApiServer{storageApiHandler: NewStorageApiHandler(logger, backends), logger: logger}, nil
}

func (s *StorageApiServer) InitializeHandler() http.Handler {
	router := mux.NewRouter()
	router.HandleFunc("/ubiquity_storage/{backend}/activate", s.storageApiHandler.Activate()).Methods("POST")
	router.HandleFunc("/ubiquity_storage/{backend}/volumes", s.storageApiHandler.CreateVolume()).Methods("POST")
	router.HandleFunc("/ubiquity_storage/{backend}/volumes", s.storageApiHandler.ListVolumes()).Methods("GET")
	router.HandleFunc("/ubiquity_storage/{backend}/volumes/{volume}", s.storageApiHandler.RemoveVolume()).Methods("DELETE")
	router.HandleFunc("/ubiquity_storage/{backend}/volumes/{volume}/attach", s.storageApiHandler.AttachVolume()).Methods("PUT")
	router.HandleFunc("/ubiquity_storage/{backend}/volumes/{volume}/detach", s.storageApiHandler.DetachVolume()).Methods("PUT")
	router.HandleFunc("/ubiquity_storage/{backend}/volumes/{volume}", s.storageApiHandler.GetVolume()).Methods("GET")
	return router
}

func (s *StorageApiServer) Start(port int) error {
	router := s.InitializeHandler()
	http.Handle("/", router)

	fmt.Println(fmt.Sprintf("Starting Storage API server on port %d ....", port))
	fmt.Println("CTL-C to exit/stop Storage API server service")
	return http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
