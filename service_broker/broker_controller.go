package service_broker

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"reflect"

	"github.ibm.com/almaden-containers/ubiquity/resources"
	"github.ibm.com/almaden-containers/ubiquity/utils"
)

const (
	DEFAULT_POLLING_INTERVAL_SECONDS = 10
	DEFAULT_CONTAINER_PATH           = "/var/vcap/data/"
)

//go:generate counterfeiter -o ../fakes/fake_broker_controller.go . BrokerController

type BrokerController interface {
	GetCatalog(logger *log.Logger) (resources.Catalog, error)
	CreateServiceInstance(logger *log.Logger, serverInstanceId string, instance resources.ServiceInstance) (resources.CreateServiceInstanceResponse, error)
	ServiceInstanceExists(logger *log.Logger, serviceInstanceId string) bool
	ServiceInstancePropertiesMatch(logger *log.Logger, serviceInstanceId string, instance resources.ServiceInstance) bool
	DeleteServiceInstance(logger *log.Logger, serviceInstanceId string) error
	BindServiceInstance(logger *log.Logger, serverInstanceId string, bindingId string, bindingInfo resources.ServiceBinding) (resources.CreateServiceBindingResponse, error)
	ServiceBindingExists(logger *log.Logger, serviceInstanceId string, bindingId string) bool
	ServiceBindingPropertiesMatch(logger *log.Logger, serviceInstanceId string, bindingId string, binding resources.ServiceBinding) bool
	GetBinding(logger *log.Logger, instanceId, bindingId string) (resources.ServiceBinding, error)
	UnbindServiceInstance(logger *log.Logger, serverInstanceId string, bindingId string) error
}

func getServices(backend resources.StorageClient) resources.Service {
	//TOOD:  branch on backend
	plan1 := resources.ServicePlan{
		Name:        "gold",
		Id:          "spectrum-scale-gold",
		Description: "Gold Tier Performance and Resiliency",
		Metadata:    nil,
		Free:        true,
	}

	plan2 := resources.ServicePlan{
		Name:        "bronze",
		Id:          "spectrum-scale-bronze",
		Description: "Bronze Tier Performance and Resiliency",
		Metadata:    nil,
		Free:        true,
	}

	service := resources.Service{
		Name:            "spectrum-scale",
		Id:              "spectrum-service-guid",
		Description:     "Provides the Spectrum FS volume service, including volume creation and volume mounts",
		Bindable:        true,
		PlanUpdateable:  false,
		Tags:            []string{"gpfs"},
		Requires:        []string{"volume_mount"},
		Metadata:        nil,
		Plans:           []resources.ServicePlan{plan1, plan2},
		DashboardClient: nil,
	}

	return service
}

type controller struct {
	backends map[string]resources.StorageClient
	//log         *log.Logger
	instanceMap map[string]*resources.ServiceInstance
	bindingMap  map[string]*resources.ServiceBinding
	configPath  string
}

func NewController(backends map[string]resources.StorageClient, configPath string) BrokerController {

	existingServiceInstances, err := loadServiceInstances(configPath)
	if err != nil {
		log.Fatal(fmt.Sprintf("error reading existing service instances: %s", err.Error()))
	}
	for _, existingServiceInstance := range existingServiceInstances {
		_, exists := backends[existingServiceInstance.ServiceId]
		if !exists {
			log.Fatal(fmt.Sprintf("error reading existing service instances: service instance refers to non-existing or disabled service (ServiceId: %s)", existingServiceInstance.ServiceId))
		}
	}

	existingServiceBindings, err := loadServiceBindings(configPath)
	if err != nil {
		log.Fatal(fmt.Sprintf("error reading existing service bindings: %s", err.Error()))
	}
	for _, existingServiceBinding := range existingServiceBindings {
		_, exists := backends[existingServiceBinding.ServiceId]
		if !exists {
			log.Fatal(fmt.Sprintf("error reading existing service bindings: service binding refers to non-existing or disabled service (ServiceId: %s)", existingServiceBinding.ServiceId))
		}
	}

	return &controller{backends: backends, configPath: configPath, instanceMap: existingServiceInstances, bindingMap: existingServiceBindings}
}

func (c *controller) GetCatalog(logger *log.Logger) (resources.Catalog, error) {
	allServices := make([]resources.Service, 0, len(c.backends))
	for _, backend := range c.backends {
		allServices = append(allServices, getServices(backend))
	}
	catalog := resources.Catalog{Services: allServices}
	return catalog, nil
}

func (c *controller) CreateServiceInstance(logger *log.Logger, serviceInstanceId string, instance resources.ServiceInstance) (resources.CreateServiceInstanceResponse, error) {
	backend, exists := c.backends[instance.ServiceId]
	if !exists {
		logger.Printf("Error: backend does not exist")
		return resources.CreateServiceInstanceResponse{}, fmt.Errorf("Error: backend does not exist")
	}

	instance.Id = serviceInstanceId
	instance.DashboardUrl = "http://dashboard_url"
	instance.LastOperation = &resources.LastOperation{
		State:                    "in progress",
		Description:              "creating service instance...",
		AsyncPollIntervalSeconds: DEFAULT_POLLING_INTERVAL_SECONDS,
	}

	volumeName := getVolumeNameForServiceInstance(&instance)
	fmt.Printf("CreateServiceInstance: Creating service instance %s with volume %s: \n", serviceInstanceId, volumeName)

	if err := backend.CreateVolume(volumeName, nil); err != nil {
		logger.Printf("Error: %s", err.Error())
		return resources.CreateServiceInstanceResponse{}, err
	}

	c.instanceMap[serviceInstanceId] = &instance

	if err := persistServiceInstances(c.configPath, c.instanceMap); err != nil {
		logger.Printf("Error: %s", err.Error())
		return resources.CreateServiceInstanceResponse{}, err
	}

	response := resources.CreateServiceInstanceResponse{
		DashboardUrl:  instance.DashboardUrl,
		LastOperation: instance.LastOperation,
	}

	return response, nil
}

func (c *controller) ServiceInstanceExists(logger *log.Logger, serviceInstanceId string) bool {
	_, exists := c.instanceMap[serviceInstanceId]
	return exists
}

func (c *controller) ServiceInstancePropertiesMatch(logger *log.Logger, serviceInstanceId string, instance resources.ServiceInstance) bool {
	existingServiceInstance, exists := c.instanceMap[serviceInstanceId]
	if exists == false {
		return false
	}
	if existingServiceInstance.PlanId != instance.PlanId {
		return false
	}
	if existingServiceInstance.SpaceGuid != instance.SpaceGuid {
		return false
	}
	if existingServiceInstance.OrganizationGuid != instance.OrganizationGuid {
		return false
	}
	areParamsEqual := reflect.DeepEqual(existingServiceInstance.Parameters, instance.Parameters)
	return areParamsEqual
}

func (c *controller) DeleteServiceInstance(logger *log.Logger, serviceInstanceId string) error {
	serviceInstance := c.instanceMap[serviceInstanceId]
	backend, exists := c.backends[serviceInstance.ServiceId]
	if !exists {
		logger.Printf("Error: backend does not exist")
		return fmt.Errorf("Error: backend does not exist")
	}
	if err := backend.RemoveVolume(getVolumeNameForServiceInstance(serviceInstance), true); err != nil {
		logger.Printf("Error: %s", err.Error())
		return err
	}

	delete(c.instanceMap, serviceInstanceId)

	if err := persistServiceInstances(c.configPath, c.instanceMap); err != nil {
		logger.Printf("Error: %s", err.Error())
		return err
	}
	return nil
}

func (c *controller) BindServiceInstance(logger *log.Logger, serviceInstanceId string, bindingId string, bindingInfo resources.ServiceBinding) (resources.CreateServiceBindingResponse, error) {
	serviceInstance := c.instanceMap[serviceInstanceId]
	backend, exists := c.backends[serviceInstance.ServiceId]
	if !exists {
		logger.Printf("Error: backend does not exist")
		return resources.CreateServiceBindingResponse{}, fmt.Errorf("Error: backend does not exist")
	}

	c.bindingMap[bindingId] = &bindingInfo
	volumeName := getVolumeNameForServiceInstance(serviceInstance)
	//volumeMetadata *model.VolumeMetadata, volumeConfigDetails *model.SpectrumConfig, err error
	config, err := backend.GetVolumeConfig(volumeName)
	if err != nil {
		logger.Printf("Error: %s", err.Error())
		return resources.CreateServiceBindingResponse{}, err
	}
	containerMountPath := determineContainerMountPath(bindingInfo.Parameters, serviceInstanceId)

	configJson, err := json.Marshal(config)
	if err != nil {
		logger.Printf("Error: %s", err.Error())
		return resources.CreateServiceBindingResponse{}, err
	}

	privateDetails := resources.VolumeMountPrivateDetails{Driver: "spctrum-scale", GroupId: volumeName, Config: string(configJson)} //TODO: fix hardcoding
	volumeMount := resources.VolumeMount{ContainerPath: containerMountPath, Mode: "rw", Private: privateDetails}
	volumeMounts := []resources.VolumeMount{volumeMount}

	if err = persistServiceBindings(c.configPath, c.bindingMap); err != nil {
		logger.Printf("Error: %s", err.Error())
		return resources.CreateServiceBindingResponse{}, err
	}

	createBindingResponse := resources.CreateServiceBindingResponse{VolumeMounts: volumeMounts}
	return createBindingResponse, nil
}

func (c *controller) ServiceBindingExists(logger *log.Logger, serviceInstanceId string, bindingId string) bool {
	_, exists := c.bindingMap[bindingId]
	return exists
}

func (c *controller) ServiceBindingPropertiesMatch(logger *log.Logger, serviceInstanceId string, bindingId string, binding resources.ServiceBinding) bool {
	existingBinding, exists := c.bindingMap[bindingId]
	if exists == false {
		return false
	}
	if existingBinding.AppId != binding.AppId {
		return false
	}
	if existingBinding.ServicePlanId != binding.ServicePlanId {
		return false
	}
	if existingBinding.ServiceId != binding.ServiceId {
		return false
	}
	if existingBinding.ServiceInstanceId != binding.ServiceInstanceId {
		return false
	}
	if existingBinding.Id != binding.Id {
		return false
	}
	return true
}

func (c *controller) GetBinding(logger *log.Logger, instanceId, bindingId string) (resources.ServiceBinding, error) {
	binding, exists := c.bindingMap[bindingId]
	if exists == true {
		return *binding, nil
	}
	return resources.ServiceBinding{}, fmt.Errorf("binding not found")

}

func (c *controller) UnbindServiceInstance(logger *log.Logger, serverInstanceId string, bindingId string) error {
	delete(c.bindingMap, bindingId)
	err := utils.MarshalAndRecord(c.bindingMap, c.configPath, "service_bindings.json")
	if err != nil {
		return err
	}
	return nil
}

func determineContainerMountPath(parameters map[string]interface{}, volId string) string {
	if containerPath, ok := parameters["container_path"]; ok {
		return containerPath.(string)
	}
	if containerPath, ok := parameters["path"]; ok {
		return containerPath.(string)
	}
	return path.Join(DEFAULT_CONTAINER_PATH, volId)
}

//func getServiceById(backendsMap map[*model.Service]backends.StorageBackend, serviceId string) (*model.Service, error) {
//	for service := range backendsMap {
//		if (*service).Id == serviceId {
//			return service, nil
//		}
//	}
//	return nil, errors.New(fmt.Sprintf("Could not locate service for serviceId %s", serviceId))
//}

func loadServiceInstances(configPath string) (map[string]*resources.ServiceInstance, error) {
	var serviceInstancesMap map[string]*resources.ServiceInstance

	err := utils.ReadAndUnmarshal(&serviceInstancesMap, configPath, "service_instances.json")
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("WARNING: service instance data file '%s' does not exist: \n", "service_instances.json")
			serviceInstancesMap = make(map[string]*resources.ServiceInstance)
		} else {
			return nil, errors.New(fmt.Sprintf("Could not load the service instances, message: %s", err.Error()))
		}
	}

	return serviceInstancesMap, nil
}

func persistServiceInstances(configPath string, instanceMap map[string]*resources.ServiceInstance) error {
	return utils.MarshalAndRecord(instanceMap, configPath, "service_instances.json")
}

func loadServiceBindings(configPath string) (map[string]*resources.ServiceBinding, error) {
	var bindingMap map[string]*resources.ServiceBinding
	err := utils.ReadAndUnmarshal(&bindingMap, configPath, "service_bindings.json")
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("WARNING: key map data file '%s' does not exist: \n", "service_bindings.json")
			bindingMap = make(map[string]*resources.ServiceBinding)
		} else {
			return nil, errors.New(fmt.Sprintf("Could not load the service instances, message: %s", err.Error()))
		}
	}

	return bindingMap, nil
}

func persistServiceBindings(configPath string, bindingMap map[string]*resources.ServiceBinding) error {
	return utils.MarshalAndRecord(bindingMap, configPath, "service_bindings.json")
}

func getVolumeNameForServiceInstance(serviceInstance *resources.ServiceInstance) string {
	volumeName := serviceInstance.Id // default to Service Instance ID as volume name if not provided
	if serviceInstance.Parameters != nil {
		volumeNameParam, ok := serviceInstance.Parameters.(map[string]interface{})["volumeName"]
		if ok {
			volumeName = volumeNameParam.(string)
		}
	}
	return volumeName
}
