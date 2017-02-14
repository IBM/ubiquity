package web_server_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"

	"github.ibm.com/almaden-containers/ubiquity/resources"
	"github.ibm.com/almaden-containers/ubiquity/web_server"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.ibm.com/almaden-containers/ubiquity/fakes"
)

var _ = Describe("ibm-storage-broker Broker Handlers", func() {

	Context("when generating handlers", func() {
		var (
			handler        http.Handler
			fakeController *fakes.FakeBrokerController
			logger         *log.Logger
		)
		BeforeEach(func() {
			var buf bytes.Buffer
			logger = log.New(&buf, "logger: ", log.Lshortfile)
			fakeController = new(fakes.FakeBrokerController)
			//logger *log.Logger, backends map[string]model.StorageClient, config model.UbiquityServerConfig
			backends := make(map[string]resources.StorageClient)
			backends["dummy"] = fakeController
			config := &resources.UbiquityServerConfig{}
			server, err := web_server.NewServer(*logger, backends, config)
			Expect(err).ToNot(HaveOccurred())
			handler = server.InitializeHandler()
		})
		Context(".Catalog", func() {
			It("should produce valid catalog response", func() {
				fakeServices := []resources.Service{resources.Service{Id: "some-service-id"}}
				fakeCatalog := resources.Catalog{
					Services: fakeServices,
				}
				fakeController.GetCatalogReturns(fakeCatalog, nil)
				w := httptest.NewRecorder()
				r, _ := http.NewRequest("GET", "http://0.0.0.0/v2/catalog", nil)
				handler.ServeHTTP(w, r)
				Expect(w.Code).Should(Equal(200))
				catalog := resources.Catalog{}
				body, err := ioutil.ReadAll(w.Body)
				Expect(err).ToNot(HaveOccurred())
				err = json.Unmarshal(body, &catalog)
				Expect(err).ToNot(HaveOccurred())
				Expect(len(catalog.Services)).To(Equal(1))
			})
			It("should error on catalog generation error", func() {
				fakeCatalog := resources.Catalog{}
				fakeController.GetCatalogReturns(fakeCatalog, fmt.Errorf("Error building catalog"))
				w := httptest.NewRecorder()
				r, _ := http.NewRequest("GET", "http://0.0.0.0/v2/catalog", nil)
				handler.ServeHTTP(w, r)
				Expect(w.Code).Should(Equal(200))
				catalog := resources.Catalog{}
				body, err := ioutil.ReadAll(w.Body)
				Expect(err).ToNot(HaveOccurred())
				err = json.Unmarshal(body, &catalog)
				Expect(err).ToNot(HaveOccurred())
				Expect(len(catalog.Services)).To(Equal(0))
			})

		})

		Context(".ServiceInstanceCreate", func() {
			It("should produce valid create service instance response", func() {
				successfulCreateService(handler, fakeController)
			})
			It("should return 409 if service instance already exists with different properties", func() {
				successfulCreateService(handler, fakeController)
				fakeController.ServiceInstanceExistsReturns(true)
				fakeController.ServiceInstancePropertiesMatchReturns(false)
				fakeCreateResponse := resources.CreateServiceInstanceResponse{}
				fakeController.CreateServiceInstanceReturns(fakeCreateResponse, nil)
				serviceInstance := resources.ServiceInstance{
					Id:               "ibm-storage-broker-service-guid",
					DashboardUrl:     "http://dashboard_url",
					InternalId:       "ibm-storage-broker-service-guid",
					ServiceId:        "ibm-storage-broker-service-guid",
					PlanId:           "free-plan-guid",
					OrganizationGuid: "organization-guid",
					SpaceGuid:        "space-guid",
					LastOperation:    nil,
					Parameters:       "parameters",
				}
				w := httptest.NewRecorder()
				payload, err := json.Marshal(serviceInstance)
				Expect(err).ToNot(HaveOccurred())
				reader := bytes.NewReader(payload)
				r, _ := http.NewRequest("PUT", "http://0.0.0.0/v2/service_instances/cephfs-service-guid", reader)
				handler.ServeHTTP(w, r)
				Expect(w.Code).Should(Equal(409))
			})
			It("should return 409 if service details not valid json", func() {
				w := httptest.NewRecorder()
				reader := bytes.NewReader([]byte(""))
				r, _ := http.NewRequest("PUT", "http://0.0.0.0/v2/service_instances/cephfs-service-guid", reader)
				handler.ServeHTTP(w, r)
				Expect(w.Code).Should(Equal(409))
			})
			It("should return 409 if service creation fails", func() {
				serviceInstance := resources.ServiceInstance{
					Id:               "ibm-storage-broker-service-guid",
					DashboardUrl:     "http://dashboard_url",
					InternalId:       "ibm-storage-broker-service-guid",
					ServiceId:        "ibm-storage-broker-service-guid",
					PlanId:           "free-plan-guid",
					OrganizationGuid: "organization-guid",
					SpaceGuid:        "space-guid",
					LastOperation:    nil,
					Parameters:       "parameters",
				}
				payload, err := json.Marshal(serviceInstance)
				Expect(err).ToNot(HaveOccurred())
				reader := bytes.NewReader(payload)
				fakeController.ServiceInstanceExistsReturns(false)
				fakeCreateResponse := resources.CreateServiceInstanceResponse{}
				fakeController.CreateServiceInstanceReturns(fakeCreateResponse, fmt.Errorf("Error creating service instance"))
				w := httptest.NewRecorder()
				r, _ := http.NewRequest("PUT", "http://0.0.0.0/v2/service_instances/cephfs-service-guid", reader)
				handler.ServeHTTP(w, r)
				Expect(w.Code).Should(Equal(409))
			})
			It("should return 200 if service instance already exists with same properties", func() {
				successfulCreateService(handler, fakeController)
				fakeController.ServiceInstanceExistsReturns(true)
				fakeController.ServiceInstancePropertiesMatchReturns(true)
				fakeCreateResponse := resources.CreateServiceInstanceResponse{}
				fakeController.CreateServiceInstanceReturns(fakeCreateResponse, nil)
				serviceInstance := resources.ServiceInstance{
					Id:               "ibm-storage-broker-service-guid",
					DashboardUrl:     "http://dashboard_url",
					InternalId:       "ibm-storage-broker-service-guid",
					ServiceId:        "ibm-storage-broker-service-guid",
					PlanId:           "free-plan-guid",
					OrganizationGuid: "organization-guid",
					SpaceGuid:        "space-guid",
					LastOperation:    nil,
					Parameters:       "parameters",
				}
				w := httptest.NewRecorder()
				payload, err := json.Marshal(serviceInstance)
				Expect(err).ToNot(HaveOccurred())
				reader := bytes.NewReader(payload)
				r, _ := http.NewRequest("PUT", "http://0.0.0.0/v2/service_instances/cephfs-service-guid", reader)
				handler.ServeHTTP(w, r)
				Expect(w.Code).Should(Equal(200))
			})
		})

		Context(".ServiceInstanceDelete", func() {
			It("should produce valid delete service instance response", func() {
				successfulCreateService(handler, fakeController)
				successfulDeleteService(handler, fakeController)
			})
			It("should return 410 if service instance does not exist", func() {
				serviceInstance := resources.ServiceInstance{
					Id:               "ibm-storage-broker-service-guid",
					DashboardUrl:     "http://dashboard_url",
					InternalId:       "ibm-storage-broker-service-guid",
					ServiceId:        "ibm-storage-broker-service-guid",
					PlanId:           "free-plan-guid",
					OrganizationGuid: "organization-guid",
					SpaceGuid:        "space-guid",
					LastOperation:    nil,
					Parameters:       "parameters",
				}
				w := httptest.NewRecorder()
				payload, err := json.Marshal(serviceInstance)
				Expect(err).ToNot(HaveOccurred())
				reader := bytes.NewReader(payload)
				r, _ := http.NewRequest("DELETE", "http://0.0.0.0/v2/service_instances/cephfs-service-guid", reader)
				handler.ServeHTTP(w, r)
				Expect(w.Code).Should(Equal(410))
			})
			It("should return 409 if service instance deletion fails", func() {
				fakeController.ServiceInstanceExistsReturns(true)
				fakeController.DeleteServiceInstanceReturns(fmt.Errorf("error deleting service instance"))
				serviceInstance := resources.ServiceInstance{
					Id:               "ibm-storage-broker-service-guid",
					DashboardUrl:     "http://dashboard_url",
					InternalId:       "ibm-storage-broker-service-guid",
					ServiceId:        "ibm-storage-broker-service-guid",
					PlanId:           "free-plan-guid",
					OrganizationGuid: "organization-guid",
					SpaceGuid:        "space-guid",
					LastOperation:    nil,
					Parameters:       "parameters",
				}
				w := httptest.NewRecorder()
				payload, err := json.Marshal(serviceInstance)
				Expect(err).ToNot(HaveOccurred())
				reader := bytes.NewReader(payload)
				r, _ := http.NewRequest("DELETE", "http://0.0.0.0/v2/service_instances/cephfs-service-guid", reader)
				handler.ServeHTTP(w, r)
				Expect(w.Code).Should(Equal(409))
			})
		})

		Context(".ServiceInstanceBind", func() {
			It("should produce valid bind service instance response", func() {
				successfulCreateService(handler, fakeController)
				successfulBindService(handler, fakeController)
			})
			It("should return 409 if binding already exists with different properties", func() {
				successfulCreateService(handler, fakeController)
				successfulBindService(handler, fakeController)
				fakeController.ServiceBindingExistsReturns(true)
				fakeController.ServiceBindingPropertiesMatchReturns(false)
				fakeBindResponse := resources.CreateServiceBindingResponse{}
				fakeController.BindServiceInstanceReturns(fakeBindResponse, nil)
				binding := resources.ServiceBinding{
					Id: "ibm-storage-broker-service-guid",
				}
				w := httptest.NewRecorder()
				payload, err := json.Marshal(binding)
				Expect(err).ToNot(HaveOccurred())
				reader := bytes.NewReader(payload)
				r, _ := http.NewRequest("PUT", "http://0.0.0.0/v2/service_instances/cephfs-service-guid/service_bindings/cephfs-service-binding-guid", reader)
				handler.ServeHTTP(w, r)
				Expect(w.Code).Should(Equal(409))
			})
			It("should return 409 if service details not valid json", func() {
				w := httptest.NewRecorder()
				reader := bytes.NewReader([]byte(""))
				r, _ := http.NewRequest("PUT", "http://0.0.0.0/v2/service_instances/cephfs-service-guid/service_bindings/cephfs-service-binding-guid", reader)
				handler.ServeHTTP(w, r)
				Expect(w.Code).Should(Equal(409))
			})
			It("should return 409 if service binding fails", func() {
				binding := resources.ServiceBinding{
					Id: "ibm-storage-broker-service-guid",
				}
				payload, err := json.Marshal(binding)
				Expect(err).ToNot(HaveOccurred())
				reader := bytes.NewReader(payload)
				fakeController.ServiceBindingExistsReturns(false)
				fakeBindingResponse := resources.CreateServiceBindingResponse{}
				fakeController.BindServiceInstanceReturns(fakeBindingResponse, fmt.Errorf("Error binding service instance"))
				w := httptest.NewRecorder()
				r, _ := http.NewRequest("PUT", "http://0.0.0.0/v2/service_instances/cephfs-service-guid/service_bindings/cephfs-service-binding-guid", reader)
				handler.ServeHTTP(w, r)
				Expect(w.Code).Should(Equal(409))
			})
			It("should return 200 if service instance already exists with same properties", func() {
				successfulCreateService(handler, fakeController)
				successfulBindService(handler, fakeController)
				fakeController.ServiceBindingExistsReturns(true)
				fakeController.ServiceBindingPropertiesMatchReturns(true)
				fakeBindingResponse := resources.CreateServiceBindingResponse{}
				fakeController.BindServiceInstanceReturns(fakeBindingResponse, nil)
				binding := resources.ServiceBinding{
					Id: "ibm-storage-broker-service-guid",
				}
				w := httptest.NewRecorder()
				payload, err := json.Marshal(binding)
				Expect(err).ToNot(HaveOccurred())
				reader := bytes.NewReader(payload)
				r, _ := http.NewRequest("PUT", "http://0.0.0.0/v2/service_instances/cephfs-service-guid/service_bindings/cephfs-service-binding-guid", reader)
				handler.ServeHTTP(w, r)
				Expect(w.Code).Should(Equal(200))
			})
		})
	})
})

func successfulCreateService(handler http.Handler, fakeController *fakes.FakeBrokerController) {
	fakeCreateResponse := resources.CreateServiceInstanceResponse{}
	fakeController.CreateServiceInstanceReturns(fakeCreateResponse, nil)
	serviceInstance := resources.ServiceInstance{
		Id:               "ibm-storage-broker-service-guid",
		DashboardUrl:     "http://dashboard_url",
		InternalId:       "ibm-storage-broker-service-guid",
		ServiceId:        "ibm-storage-broker-service-guid",
		PlanId:           "free-plan-guid",
		OrganizationGuid: "organization-guid",
		SpaceGuid:        "space-guid",
		LastOperation:    nil,
		Parameters:       "parameters",
	}
	w := httptest.NewRecorder()
	payload, err := json.Marshal(serviceInstance)
	Expect(err).ToNot(HaveOccurred())
	reader := bytes.NewReader(payload)
	r, _ := http.NewRequest("PUT", "http://0.0.0.0/v2/service_instances/cephfs-service-guid", reader)
	handler.ServeHTTP(w, r)
	Expect(w.Code).Should(Equal(201))
	body, err := ioutil.ReadAll(w.Body)
	Expect(err).ToNot(HaveOccurred())
	createServiceResponse := resources.CreateServiceInstanceResponse{}
	err = json.Unmarshal(body, &createServiceResponse)
	Expect(err).ToNot(HaveOccurred())
}

func successfulDeleteService(handler http.Handler, fakeController *fakes.FakeBrokerController) {
	serviceInstance := resources.ServiceInstance{
		Id:               "ibm-storage-broker-service-guid",
		DashboardUrl:     "http://dashboard_url",
		InternalId:       "ibm-storage-broker-service-guid",
		ServiceId:        "ibm-storage-broker-service-guid",
		PlanId:           "free-plan-guid",
		OrganizationGuid: "organization-guid",
		SpaceGuid:        "space-guid",
		LastOperation:    nil,
		Parameters:       "parameters",
	}
	fakeController.ServiceInstanceExistsReturns(true)
	w := httptest.NewRecorder()
	payload, err := json.Marshal(serviceInstance)
	Expect(err).ToNot(HaveOccurred())
	reader := bytes.NewReader(payload)
	r, _ := http.NewRequest("DELETE", "http://0.0.0.0/v2/service_instances/cephfs-service-guid", reader)
	handler.ServeHTTP(w, r)
	Expect(w.Code).Should(Equal(200))
}

func successfulBindService(handler http.Handler, fakeController *fakes.FakeBrokerController) {
	fakeBindResponse := resources.CreateServiceBindingResponse{}
	fakeController.BindServiceInstanceReturns(fakeBindResponse, nil)
	binding := resources.ServiceBinding{
		Id: "ibm-storage-broker-service-guid",
	}
	w := httptest.NewRecorder()
	payload, err := json.Marshal(binding)
	Expect(err).ToNot(HaveOccurred())
	reader := bytes.NewReader(payload)
	r, _ := http.NewRequest("PUT", "http://0.0.0.0/v2/service_instances/cephfs-service-guid/service_bindings/cephfs-service-binding-guid", reader)
	handler.ServeHTTP(w, r)
	Expect(w.Code).Should(Equal(201))
	body, err := ioutil.ReadAll(w.Body)
	Expect(err).ToNot(HaveOccurred())
	bindingResponse := resources.CreateServiceBindingResponse{}
	err = json.Unmarshal(body, &bindingResponse)
	Expect(err).ToNot(HaveOccurred())
}
