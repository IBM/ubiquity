package service_broker_test

import (
	"fmt"
	"log"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.ibm.com/almaden-containers/ubiquity/fakes"
	"github.ibm.com/almaden-containers/ubiquity/resources"
	"github.ibm.com/almaden-containers/ubiquity/service_broker"
)

var _ = Describe("ibm-storage-broker Broker", func() {
	var (
		controller      service_broker.BrokerController
		localMountPoint string
		serviceGuid     string
		instanceMap     map[string]*resources.ServiceInstance
		bindingMap      map[string]*resources.ServiceBinding
		testLogger      log.Logger
		fakeBackend     *fakes.FakeStorageClient
		configPath      string
		storageBackends map[string]resources.StorageClient
	)
	BeforeEach(func() {
		serviceGuid = "some-service-guid"
		localMountPoint = "/tmp/share"
		configPath = "/tmp/ibm-storage-broker"
		instanceMap = make(map[string]*resources.ServiceInstance)
		bindingMap = make(map[string]*resources.ServiceBinding)
		fakeBackend = new(fakes.FakeStorageClient)

		storageBackends = make(map[string]resources.StorageClient)
		storageBackends["fake-backend"] = fakeBackend
		controller = service_broker.NewController(storageBackends, configPath)
		//controller = core.NewController(fakeBackend, configPath, instanceMap, bindingMap)

	})
	Context(".Catalog", func() {
		FIt("should produce a valid catalog", func() {
			catalog, err := controller.GetCatalog(testLogger)
			Expect(err).ToNot(HaveOccurred())
			Expect(catalog).ToNot(BeNil())
			Expect(catalog.Services).ToNot(BeNil())
			Expect(len(catalog.Services)).To(Equal(1))
			Expect(catalog.Services[0].Name).To(Equal("spectrum-scale"))
			Expect(catalog.Services[0].Requires).ToNot(BeNil())
			Expect(len(catalog.Services[0].Requires)).To(Equal(1))
			Expect(catalog.Services[0].Requires[0]).To(Equal("volume_mount"))

			Expect(catalog.Services[0].Plans).ToNot(BeNil())
			Expect(len(catalog.Services[0].Plans)).To(Equal(1))
			Expect(catalog.Services[0].Plans[0].Name).To(Equal("free"))

			Expect(catalog.Services[0].Bindable).To(Equal(true))
		})
		Context(".CreateServiceInstance", func() {
			var (
				instance resources.ServiceInstance
			)
			BeforeEach(func() {
				instance = resources.ServiceInstance{}
				instance.PlanId = "some-planId"
				instance.Parameters = map[string]interface{}{"some-property": "some-value"}

			})
			It("should create a valid service instance", func() {
				successfulServiceInstanceCreate(testLogger, fakeBackend, instance, controller, serviceGuid)
			})
			Context("should fail to create service instance", func() {
				It("when share creation errors", func() {
					properties := map[string]interface{}{"some-property": "some-value"}
					instance.Parameters = properties
					fakeBackend.CreateVolumeReturns(fmt.Errorf("Failed to create fileset"))
					_, err := controller.CreateServiceInstance(testLogger, "service-instance-guid", instance)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("Failed to create fileset"))
				})
				It("should error when updating internal bookkeeping fails", func() {
					controller = service_broker.NewController(storageBackends, "/non-existent-path")
					_, err := controller.CreateServiceInstance(testLogger, "service-instance-guid", instance)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal(fmt.Sprintf("open /non-existent-path/service_instances.json: no such file or directory")))
				})

			})
		})
		Context(".ServiceInstanceExists", func() {
			var (
				instance resources.ServiceInstance
			)
			BeforeEach(func() {
				instance = resources.ServiceInstance{}
				instance.PlanId = "some-planId"
				instance.Parameters = map[string]interface{}{"some-property": "some-value"}

			})
			It("should confirm existence of service instance", func() {
				successfulServiceInstanceCreate(testLogger, fakeBackend, instance, controller, serviceGuid)
				serviceExists := controller.ServiceInstanceExists(testLogger, serviceGuid)
				Expect(serviceExists).To(Equal(true))
			})
			It("should confirm non-existence of service instance", func() {
				serviceExists := controller.ServiceInstanceExists(testLogger, serviceGuid)
				Expect(serviceExists).To(Equal(false))
			})
		})
		Context(".ServiceInstancePropertiesMatch", func() {
			var (
				instance resources.ServiceInstance
			)
			BeforeEach(func() {
				instance = resources.ServiceInstance{}
				instance.PlanId = "some-planId"
				instance.Parameters = map[string]interface{}{"some-property": "some-value"}

			})
			It("should return true if properties match", func() {
				successfulServiceInstanceCreate(testLogger, fakeBackend, instance, controller, serviceGuid)
				anotherInstance := resources.ServiceInstance{}
				properties := map[string]interface{}{"some-property": "some-value"}
				anotherInstance.Parameters = properties
				anotherInstance.PlanId = "some-planId"
				propertiesMatch := controller.ServiceInstancePropertiesMatch(testLogger, serviceGuid, anotherInstance)
				Expect(propertiesMatch).To(Equal(true))
			})
			It("should return false if properties do not match", func() {
				successfulServiceInstanceCreate(testLogger, fakeBackend, instance, controller, serviceGuid)
				anotherInstance := resources.ServiceInstance{}
				properties := map[string]interface{}{"some-property": "some-value"}
				anotherInstance.Parameters = properties
				anotherInstance.PlanId = "some-other-planId"
				propertiesMatch := controller.ServiceInstancePropertiesMatch(testLogger, serviceGuid, anotherInstance)
				Expect(propertiesMatch).ToNot(Equal(true))
			})
		})

		Context(".ServiceInstanceDelete", func() {
			var (
				instance resources.ServiceInstance
			)
			BeforeEach(func() {
				instance = resources.ServiceInstance{}
				instance.PlanId = "some-planId"
				instance.Parameters = map[string]interface{}{"some-property": "some-value"}
			})
			It("should delete service instance", func() {
				successfulServiceInstanceCreate(testLogger, fakeBackend, instance, controller, serviceGuid)
				err := controller.DeleteServiceInstance(testLogger, serviceGuid)
				Expect(err).ToNot(HaveOccurred())

				serviceExists := controller.ServiceInstanceExists(testLogger, serviceGuid)
				Expect(serviceExists).To(Equal(false))
			})
			It("should error when trying to delete non-existence service instance", func() {
				fakeBackend.RemoveVolumeReturns(fmt.Errorf("Failed to delete fileset, fileset not found"))
				err := controller.DeleteServiceInstance(testLogger, serviceGuid)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("Failed to delete fileset, fileset not found"))
			})
			It("should error when updating internal bookkeeping fails", func() {
				controller = service_broker.NewController(storageBackends, "/non-existent-path")
				err := controller.DeleteServiceInstance(testLogger, serviceGuid)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal(fmt.Sprintf("open /non-existent-path/service_instances.json: no such file or directory")))
			})

		})
		Context(".BindServiceInstance", func() {
			var (
				instance    resources.ServiceInstance
				bindingInfo resources.ServiceBinding
			)
			BeforeEach(func() {
				instance = resources.ServiceInstance{}
				instance.PlanId = "some-planId"
				instance.Parameters = map[string]interface{}{"some-property": "some-value"}
				bindingInfo = resources.ServiceBinding{}
				successfulServiceInstanceCreate(testLogger, fakeBackend, instance, controller, serviceGuid)
			})
			It("should be able bind service instance", func() {
				config := resources.SpectrumScaleConfig{}
				//model.SpectrumConfig
				fakeBackend.GetVolumeReturns(&resources.VolumeMetadata{Mountpoint: "/gpfs/fileset1"}, &config, nil)
				bindingResponse, err := controller.BindServiceInstance(testLogger, serviceGuid, "some-binding-id", bindingInfo)
				Expect(err).ToNot(HaveOccurred())
				Expect(bindingResponse.VolumeMounts).ToNot(BeNil())
				Expect(len(bindingResponse.VolumeMounts)).To(Equal(1))
			})
			Context("should fail", func() {
				It("when unable to find the backing share", func() {
					config := resources.SpectrumScaleConfig{}
					fakeBackend.GetVolumeReturns(&resources.VolumeMetadata{}, &config, fmt.Errorf("Cannot find fileset, internal error"))
					_, err := controller.BindServiceInstance(testLogger, serviceGuid, "some-binding-id", bindingInfo)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("Cannot find fileset, internal error"))
				})
				It("when updating internal bookkeeping fails", func() {
					config := resources.SpectrumScaleConfig{}
					fakeBackend.GetVolumeReturns(&resources.VolumeMetadata{Mountpoint: "/gpfs/fileset1"}, &config, nil)
					controller = service_broker.NewController(storageBackends, "/non-existent-path")
					_, err := controller.BindServiceInstance(testLogger, serviceGuid, "some-binding-id", bindingInfo)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal(fmt.Sprintf("open /non-existent-path/service_bindings.json: no such file or directory")))
				})
			})
		})
		Context(".ServiceBindingExists", func() {
			var (
				instance  resources.ServiceInstance
				bindingId string
			)
			BeforeEach(func() {
				instance = resources.ServiceInstance{}
				instance.PlanId = "some-planId"
				instance.Parameters = map[string]interface{}{"some-property": "some-value"}
				bindingId = "some-binding-id"
			})
			It("should confirm existence of service instance", func() {
				config := resources.SpectrumScaleConfig{}
				fakeBackend.GetVolumeReturns(&resources.VolumeMetadata{Mountpoint: "/gpfs/fileset1"}, &config, nil)
				binding := resources.ServiceBinding{}
				successfulServiceInstanceCreate(testLogger, fakeBackend, instance, controller, serviceGuid)
				successfulServiceBindingCreate(testLogger, fakeBackend, binding, controller, serviceGuid, bindingId)
				bindingExists := controller.ServiceBindingExists(testLogger, serviceGuid, bindingId)
				Expect(bindingExists).To(Equal(true))
			})
			It("should confirm non-existence of service binding", func() {
				bindingExists := controller.ServiceBindingExists(testLogger, serviceGuid, bindingId)
				Expect(bindingExists).To(Equal(false))
			})
		})
		Context(".ServiceBindingPropertiesMatch", func() {
			var (
				instance  resources.ServiceInstance
				bindingId string
			)
			BeforeEach(func() {
				instance = resources.ServiceInstance{}
				instance.PlanId = "some-planId"
				instance.Parameters = map[string]interface{}{"some-property": "some-value"}
				bindingId = "some-binding-id"

			})
			It("should return true if properties match", func() {
				binding := resources.ServiceBinding{}
				config := resources.SpectrumScaleConfig{}
				fakeBackend.GetVolumeReturns(&resources.VolumeMetadata{Mountpoint: "/gpfs/fileset1"}, &config, nil)
				successfulServiceInstanceCreate(testLogger, fakeBackend, instance, controller, serviceGuid)
				successfulServiceBindingCreate(testLogger, fakeBackend, binding, controller, serviceGuid, bindingId)
				anotherBinding := resources.ServiceBinding{}
				propertiesMatch := controller.ServiceBindingPropertiesMatch(testLogger, serviceGuid, bindingId, anotherBinding)
				Expect(propertiesMatch).To(Equal(true))
			})
			It("should return false if properties do not match", func() {
				binding := resources.ServiceBinding{}
				config := resources.SpectrumScaleConfig{}
				fakeBackend.GetVolumeReturns(&resources.VolumeMetadata{Mountpoint: "/gpfs/fileset1"}, &config, nil)
				successfulServiceInstanceCreate(testLogger, fakeBackend, instance, controller, serviceGuid)
				successfulServiceBindingCreate(testLogger, fakeBackend, binding, controller, serviceGuid, bindingId)
				anotherBinding := resources.ServiceBinding{}
				anotherBinding.AppId = "some-other-appId"
				propertiesMatch := controller.ServiceBindingPropertiesMatch(testLogger, serviceGuid, bindingId, anotherBinding)
				Expect(propertiesMatch).ToNot(Equal(true))
			})
		})
	})
})

func successfulServiceInstanceCreate(testLogger log.Logger, fakeBackend *fakes.FakeStorageClient, instance resources.ServiceInstance, controller service_broker.BrokerController, serviceGuid string) {
	fakeBackend.CreateVolumeReturns(nil)
	createResponse, err := controller.CreateServiceInstance(testLogger, serviceGuid, instance)
	Expect(err).ToNot(HaveOccurred())
	Expect(createResponse.DashboardUrl).ToNot(Equal(""))
}

func successfulServiceBindingCreate(testLogger log.Logger, fakeBackend *fakes.FakeStorageClient, binding resources.ServiceBinding, controller service_broker.BrokerController, serviceGuid string, bindingId string) {
	bindResponse, err := controller.BindServiceInstance(testLogger, serviceGuid, bindingId, binding)
	Expect(err).ToNot(HaveOccurred())
	Expect(bindResponse.VolumeMounts).ToNot(BeNil())
	Expect(len(bindResponse.VolumeMounts)).To(Equal(1))
}
