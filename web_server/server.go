package web_server

import (
	"fmt"
	"log"
	"net/http"

	"github.ibm.com/almaden-containers/ibm-storage-broker.git/core"
	"github.ibm.com/almaden-containers/ibm-storage-broker.git/model"
	"github.ibm.com/almaden-containers/ibm-storage-broker.git/utils"

	"github.com/gorilla/mux"
	"encoding/json"
)

type Server struct {
	controller core.Controller
	log        log.Logger
}

func NewServer(controller core.Controller, log log.Logger) (*Server, error) {
	return &Server{controller: controller, log: log}, nil
}

func (s *Server) InitializeHandler() http.Handler {
	router := mux.NewRouter()

	router.HandleFunc("/v2/catalog", s.newCatalogHandler()).Methods("GET")
	//router.HandleFunc("/v2/service_instances/{service_instance_guid}", s.controller.GetServiceInstance).Methods("GET")
	router.HandleFunc("/v2/service_instances/{service_instance_guid}", s.newCreateServiceInstanceHandler()).Methods("PUT")
	router.HandleFunc("/v2/service_instances/{service_instance_guid}", s.newDeleteServiceInstanceHandler()).Methods("DELETE")
	router.HandleFunc("/v2/service_instances/{service_instance_guid}/service_bindings/{service_binding_guid}", s.newBindServiceInstanceHandler()).Methods("PUT")
	router.HandleFunc("/v2/service_instances/{service_instance_guid}/service_bindings/{service_binding_guid}", s.newUnbindServiceInstanceHandler()).Methods("DELETE")
	return router
}

func (s *Server) Start(port string) error {
	router := s.InitializeHandler()
	http.Handle("/", router)

	fmt.Println("Starting server on port " + port + "...")
	fmt.Println("CTL-C to break out of broker")
	return http.ListenAndServe(":"+port, nil)
}

func (s *Server) newCatalogHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		s.log.Println("start")

		catalog, err := s.controller.GetCatalog(s.log)
		if err != nil {
			utils.WriteResponse(w, http.StatusOK, struct{}{})
			return
		}
		utils.WriteResponse(w, http.StatusOK, catalog)
	}
}

func (s *Server) newCreateServiceInstanceHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		instanceId := utils.ExtractVarsFromRequest(req, "service_instance_guid")
		var instance model.ServiceInstance
		err := utils.UnmarshalDataFromRequest(req, &instance)
		if err != nil {
			utils.WriteResponse(w, 409, struct{}{})
			return
		}
		serviceInstanceExists := s.controller.ServiceInstanceExists(s.log, instanceId)
		if serviceInstanceExists {
			if s.controller.ServiceInstancePropertiesMatch(s.log, instanceId, instance) == true {
				response := model.CreateServiceInstanceResponse{
					DashboardUrl:  "http://dashboard_url",
					LastOperation: nil,
				}
				utils.WriteResponse(w, 200, response)
				return
			} else {
				utils.WriteResponse(w, 409, struct{}{})
				return
			}
		}
		createResponse, err := s.controller.CreateServiceInstance(s.log, instanceId, instance)
		if err != nil {
			utils.WriteResponse(w, 409, struct{}{})
			return
		}
		json,_ := json.Marshal(createResponse)
		s.log.Printf("Create response %s\n", string(json))
		utils.WriteResponse(w, 201, createResponse)
	}
}

func (s *Server) newDeleteServiceInstanceHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		instanceId := utils.ExtractVarsFromRequest(req, "service_instance_guid")
		serviceInstanceExists := s.controller.ServiceInstanceExists(s.log, instanceId)
		if serviceInstanceExists == false {
			utils.WriteResponse(w, 410, struct{}{})
			return
		}
		err := s.controller.DeleteServiceInstance(s.log, instanceId)
		if err != nil {
			utils.WriteResponse(w, 409, struct{}{})
			return
		}
		utils.WriteResponse(w, 200, struct{}{})
	}
}

func (s *Server) newBindServiceInstanceHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		instanceId := utils.ExtractVarsFromRequest(req, "service_instance_guid")
		bindingId := utils.ExtractVarsFromRequest(req, "service_binding_guid")
		var binding model.ServiceBinding
		err := utils.UnmarshalDataFromRequest(req, &binding)
		if err != nil {
			s.log.Println("Error unmarshaling data from request")
			utils.WriteResponse(w, 409, struct{}{})
			return
		}
		serviceBindingExists := s.controller.ServiceBindingExists(s.log, instanceId, bindingId)
		s.log.Printf("Looking for serviceid: %s, bindingid: %s\n", instanceId, bindingId)
		if serviceBindingExists {
			s.log.Printf("Looking for serviceid: %s, bindingid: %s......found\n", instanceId, bindingId)
			if s.controller.ServiceBindingPropertiesMatch(s.log, instanceId, bindingId, binding) == true {
				response, err := s.controller.GetBinding(s.log, instanceId, bindingId)
				if err != nil {
					s.log.Println("Error getting existing binding")
					utils.WriteResponse(w, 409, struct{}{})
					return
				}
				utils.WriteResponse(w, 200, response)
				return
			} else {
				s.log.Println("Error binding properties different from before")

				utils.WriteResponse(w, 409, struct{}{})
				return
			}
		}
		bindResponse, err := s.controller.BindServiceInstance(s.log, instanceId, bindingId, binding)
		if err != nil {
			s.log.Println("Error controller bindserviceInstance failed ")
			utils.WriteResponse(w, 409, struct{}{})
			return
		}
		json,_ := json.Marshal(bindResponse)
		s.log.Printf("Bind response %s\n", string(json))

		utils.WriteResponse(w, 201, bindResponse)
	}
}

func (s *Server) newUnbindServiceInstanceHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		instanceId := utils.ExtractVarsFromRequest(req, "service_instance_guid")
		bindingId := utils.ExtractVarsFromRequest(req, "service_binding_guid")
		err := s.controller.UnbindServiceInstance(s.log, instanceId, bindingId)
		if err != nil {
			utils.WriteResponse(w, 410, struct{}{})
			return
		}
		utils.WriteResponse(w, 200, struct{}{})
	}

}
