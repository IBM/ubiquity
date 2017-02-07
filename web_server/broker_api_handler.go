package web_server

import (
	"encoding/json"
	"log"
	"net/http"

	"github.ibm.com/almaden-containers/ubiquity/service_broker"
	"github.ibm.com/almaden-containers/ubiquity/resources"
	"github.ibm.com/almaden-containers/ubiquity/utils"
)

type BrokerApiHandler struct {
	Controller service_broker.BrokerController
	log        *log.Logger
}

func NewBrokerApiHandler(logger *log.Logger, controller service_broker.BrokerController) *BrokerApiHandler {

	return &BrokerApiHandler{log: logger, Controller: controller}
}

func (h *BrokerApiHandler) Catalog() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		h.log.Println("start")

		catalog, err := h.Controller.GetCatalog(h.log)
		if err != nil {
			utils.WriteResponse(w, http.StatusOK, struct{}{})
			return
		}
		utils.WriteResponse(w, http.StatusOK, catalog)
	}
}

func (h *BrokerApiHandler) CreateServiceInstance() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		instanceId := utils.ExtractVarsFromRequest(req, "service_instance_guid")
		var instance resources.ServiceInstance
		err := utils.UnmarshalDataFromRequest(req, &instance)
		if err != nil {
			utils.WriteResponse(w, 409, struct{}{})
			return
		}
		serviceInstanceExists := h.Controller.ServiceInstanceExists(h.log, instanceId)
		if serviceInstanceExists {
			if h.Controller.ServiceInstancePropertiesMatch(h.log, instanceId, instance) == true {
				response := resources.CreateServiceInstanceResponse{
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
		createResponse, err := h.Controller.CreateServiceInstance(h.log, instanceId, instance)
		if err != nil {
			utils.WriteResponse(w, 409, struct{}{})
			return
		}
		json, _ := json.Marshal(createResponse)
		h.log.Printf("Create response %s\n", string(json))
		utils.WriteResponse(w, 201, createResponse)
	}
}

func (h *BrokerApiHandler) DeleteServiceInstance() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		instanceId := utils.ExtractVarsFromRequest(req, "service_instance_guid")
		serviceInstanceExists := h.Controller.ServiceInstanceExists(h.log, instanceId)
		if serviceInstanceExists == false {
			utils.WriteResponse(w, 410, struct{}{})
			return
		}
		err := h.Controller.DeleteServiceInstance(h.log, instanceId)
		if err != nil {
			utils.WriteResponse(w, 409, struct{}{})
			return
		}
		utils.WriteResponse(w, 200, struct{}{})
	}
}

func (h *BrokerApiHandler) BindServiceInstance() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		instanceId := utils.ExtractVarsFromRequest(req, "service_instance_guid")
		bindingId := utils.ExtractVarsFromRequest(req, "service_binding_guid")
		var binding resources.ServiceBinding
		err := utils.UnmarshalDataFromRequest(req, &binding)
		if err != nil {
			h.log.Println("Error unmarshaling data from request")
			utils.WriteResponse(w, 409, struct{}{})
			return
		}
		serviceBindingExists := h.Controller.ServiceBindingExists(h.log, instanceId, bindingId)
		h.log.Printf("Looking for serviceid: %s, bindingid: %s\n", instanceId, bindingId)
		if serviceBindingExists {
			h.log.Printf("Looking for serviceid: %s, bindingid: %s......found\n", instanceId, bindingId)
			if h.Controller.ServiceBindingPropertiesMatch(h.log, instanceId, bindingId, binding) == true {
				response, err := h.Controller.GetBinding(h.log, instanceId, bindingId)
				if err != nil {
					h.log.Println("Error getting existing binding")
					utils.WriteResponse(w, 409, struct{}{})
					return
				}
				utils.WriteResponse(w, 200, response)
				return
			} else {
				h.log.Println("Error binding properties different from before")

				utils.WriteResponse(w, 409, struct{}{})
				return
			}
		}
		bindResponse, err := h.Controller.BindServiceInstance(h.log, instanceId, bindingId, binding)
		if err != nil {
			h.log.Println("Error controller bindserviceInstance failed ")
			utils.WriteResponse(w, 409, struct{}{})
			return
		}
		json, _ := json.Marshal(bindResponse)
		h.log.Printf("Bind response %s\n", string(json))

		utils.WriteResponse(w, 201, bindResponse)
	}
}

func (h *BrokerApiHandler) UnbindServiceInstance() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		instanceId := utils.ExtractVarsFromRequest(req, "service_instance_guid")
		bindingId := utils.ExtractVarsFromRequest(req, "service_binding_guid")
		err := h.Controller.UnbindServiceInstance(h.log, instanceId, bindingId)
		if err != nil {
			utils.WriteResponse(w, 410, struct{}{})
			return
		}
		utils.WriteResponse(w, 200, struct{}{})
	}

}
