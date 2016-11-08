package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"reflect"

	"github.ibm.com/almaden-containers/ubiquity/model"
	"github.ibm.com/almaden-containers/ubiquity/utils"
)

const (
	DEFAULT_POLLING_INTERVAL_SECONDS = 10
	DEFAULT_CONTAINER_PATH           = "/var/vcap/data/"
)

//go:generate counterfeiter -o ../fakes/fake_broker_controller.go . BrokerController

type BrokerController interface {
	GetCatalog(logger *log.Logger) (model.Catalog, error)
	CreateServiceInstance(logger *log.Logger, serverInstanceId string, instance model.ServiceInstance) (model.CreateServiceInstanceResponse, error)
	ServiceInstanceExists(logger *log.Logger, serviceInstanceId string) bool
	ServiceInstancePropertiesMatch(logger *log.Logger, serviceInstanceId string, instance model.ServiceInstance) bool
	DeleteServiceInstance(logger *log.Logger, serviceInstanceId string) error
	BindServiceInstance(logger *log.Logger, serverInstanceId string, bindingId string, bindingInfo model.ServiceBinding) (model.CreateServiceBindingResponse, error)
	ServiceBindingExists(logger *log.Logger, serviceInstanceId string, bindingId string) bool
	ServiceBindingPropertiesMatch(logger *log.Logger, serviceInstanceId string, bindingId string, binding model.ServiceBinding) bool
	GetBinding(logger *log.Logger, instanceId, bindingId string) (model.ServiceBinding, error)
	UnbindServiceInstance(logger *log.Logger, serverInstanceId string, bindingId string) error
}

func getServices(backend model.StorageClient) model.Service {
	//TOOD:  branch on backend
	plan1 := model.ServicePlan{
		Name:        "gold",
		Id:          "spectrum-scale-gold",
		Description: "Gold Tier Performance and Resiliency",
		Metadata:    nil,
		Free:        true,
	}

	plan2 := model.ServicePlan{
		Name:        "bronze",
		Id:          "spectrum-scale-bronze",
		Description: "Bronze Tier Performance and Resiliency",
		Metadata:    nil,
		Free:        true,
	}

	service := model.Service{
		Name:            "spectrum-scale",
		Id:              "spectrum-service-guid",
		Description:     "Provides the Spectrum FS volume service, including volume creation and volume mounts",
		Bindable:        true,
		PlanUpdateable:  false,
		Tags:            []string{"gpfs"},
		Requires:        []string{"volume_mount"},
		Metadata:        nil,
		Plans:           []model.ServicePlan{plan1, plan2},
		DashboardClient: nil,
	}

	return service
}

type controller struct {
	backends map[string]model.StorageClient
	//log         *log.Logger
	instanceMap map[string]*model.ServiceInstance
	bindingMap  map[string]*model.ServiceBinding
	configPath  string
}

func NewController(backends map[string]model.StorageClient, configPath string) BrokerController {

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

func (c *controller) GetCatalog(logger *log.Logger) (model.Catalog, error) {
	allServices := make([]model.Service, 0, len(c.backends))
	for _, backend := range c.backends {
		allServices = append(allServices, getServices(backend))
	}
	catalog := model.Catalog{Services: allServices}
	return catalog, nil
}

func (c *controller) CreateServiceInstance(logger *log.Logger, serviceInstanceId string, instance model.ServiceInstance) (model.CreateServiceInstanceResponse, error) {
	backend, exists := c.backends[instance.ServiceId]
	if !exists {
		logger.Printf("Error: backend does not exist")
		return model.CreateServiceInstanceResponse{}, fmt.Errorf("Error: backend does not exist")
	}

	instance.Id = serviceInstanceId
	instance.DashboardUrl = "http://dashboard_url"
	instance.LastOperation = &model.LastOperation{
		State:                    "in progress",
		Description:              "creating service instance...",
		AsyncPollIntervalSeconds: DEFAULT_POLLING_INTERVAL_SECONDS,
	}

	volumeName := getVolumeNameForServiceInstance(&instance)
	fmt.Printf("CreateServiceInstance: Creating service instance %s with volume %s: \n", serviceInstanceId, volumeName)

	if err := backend.CreateVolume(volumeName, nil); err != nil {
		logger.Printf("Error: %s", err.Error())
		return model.CreateServiceInstanceResponse{}, err
	}

	c.instanceMap[serviceInstanceId] = &instance

	if err := persistServiceInstances(c.configPath, c.instanceMap); err != nil {
		logger.Printf("Error: %s", err.Error())
		return model.CreateServiceInstanceResponse{}, err
	}

	response := model.CreateServiceInstanceResponse{
		DashboardUrl:  instance.DashboardUrl,
		LastOperation: instance.LastOperation,
	}

	return response, nil
}

func (c *controller) ServiceInstanceExists(logger *log.Logger, serviceInstanceId string) bool {
	_, exists := c.instanceMap[serviceInstanceId]
	return exists
}

func (c *controller) ServiceInstancePropertiesMatch(logger *log.Logger, serviceInstanceId string, instance model.ServiceInstance) bool {
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

func (c *controller) BindServiceInstance(logger *log.Logger, serviceInstanceId string, bindingId string, bindingInfo model.ServiceBinding) (model.CreateServiceBindingResponse, error) {
	serviceInstance := c.instanceMap[serviceInstanceId]
	backend, exists := c.backends[serviceInstance.ServiceId]
	if !exists {
		logger.Printf("Error: backend does not exist")
		return model.CreateServiceBindingResponse{}, fmt.Errorf("Error: backend does not exist")
	}

	c.bindingMap[bindingId] = &bindingInfo
	volumeName := getVolumeNameForServiceInstance(serviceInstance)
	//volumeMetadata *model.VolumeMetadata, volumeConfigDetails *model.SpectrumConfig, err error
	_, config, err := backend.GetVolume(volumeName)
	if err != nil {
		logger.Printf("Error: %s", err.Error())
		return model.CreateServiceBindingResponse{}, err
	}
	containerMountPath := determineContainerMountPath(bindingInfo.Parameters, serviceInstanceId)

	configJson, err := json.Marshal(config)
	if err != nil {
		logger.Printf("Error: %s", err.Error())
		return model.CreateServiceBindingResponse{}, err
	}

	privateDetails := model.VolumeMountPrivateDetails{Driver: "spctrum-scale", GroupId: volumeName, Config: string(configJson)} //TODO: fix hardcoding
	volumeMount := model.VolumeMount{ContainerPath: containerMountPath, Mode: "rw", Private: privateDetails}
	volumeMounts := []model.VolumeMount{volumeMount}

	if err = persistServiceBindings(c.configPath, c.bindingMap); err != nil {
		logger.Printf("Error: %s", err.Error())
		return model.CreateServiceBindingResponse{}, err
	}

	createBindingResponse := model.CreateServiceBindingResponse{VolumeMounts: volumeMounts}
	return createBindingResponse, nil
}

func (c *controller) ServiceBindingExists(logger *log.Logger, serviceInstanceId string, bindingId string) bool {
	_, exists := c.bindingMap[bindingId]
	return exists
}

func (c *controller) ServiceBindingPropertiesMatch(logger *log.Logger, serviceInstanceId string, bindingId string, binding model.ServiceBinding) bool {
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

func (c *controller) GetBinding(logger *log.Logger, instanceId, bindingId string) (model.ServiceBinding, error) {
	binding, exists := c.bindingMap[bindingId]
	if exists == true {
		return *binding, nil
	}
	return model.ServiceBinding{}, fmt.Errorf("binding not found")

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

func loadServiceInstances(configPath string) (map[string]*model.ServiceInstance, error) {
	var serviceInstancesMap map[string]*model.ServiceInstance

	err := utils.ReadAndUnmarshal(&serviceInstancesMap, configPath, "service_instances.json")
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("WARNING: service instance data file '%s' does not exist: \n", "service_instances.json")
			serviceInstancesMap = make(map[string]*model.ServiceInstance)
		} else {
			return nil, errors.New(fmt.Sprintf("Could not load the service instances, message: %s", err.Error()))
		}
	}

	return serviceInstancesMap, nil
}

func persistServiceInstances(configPath string, instanceMap map[string]*model.ServiceInstance) error {
	return utils.MarshalAndRecord(instanceMap, configPath, "service_instances.json")
}

func loadServiceBindings(configPath string) (map[string]*model.ServiceBinding, error) {
	var bindingMap map[string]*model.ServiceBinding
	err := utils.ReadAndUnmarshal(&bindingMap, configPath, "service_bindings.json")
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("WARNING: key map data file '%s' does not exist: \n", "service_bindings.json")
			bindingMap = make(map[string]*model.ServiceBinding)
		} else {
			return nil, errors.New(fmt.Sprintf("Could not load the service instances, message: %s", err.Error()))
		}
	}

	return bindingMap, nil
}

func persistServiceBindings(configPath string, bindingMap map[string]*model.ServiceBinding) error {
	return utils.MarshalAndRecord(bindingMap, configPath, "service_bindings.json")
}

func getVolumeNameForServiceInstance(serviceInstance *model.ServiceInstance) string {
	volumeName := serviceInstance.Id // default to Service Instance ID as volume name if not provided
	if serviceInstance.Parameters != nil {
		volumeNameParam, ok := serviceInstance.Parameters.(map[string]interface{})["volumeName"]
		if ok {
			volumeName = volumeNameParam.(string)
		}
	}
	return volumeName
}
