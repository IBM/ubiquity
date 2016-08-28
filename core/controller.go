package core

import (
	"fmt"
	"log"
	"path"
	"reflect"
	"errors"
	"github.ibm.com/almaden-containers/ibm-storage-broker.git/model"
	"github.ibm.com/almaden-containers/ibm-storage-broker.git/utils"
	"encoding/json"
	"os"
)

const (
	DEFAULT_POLLING_INTERVAL_SECONDS = 10
	DEFAULT_CONTAINER_PATH           = "/var/vcap/data/"
)

//go:generate counterfeiter -o ./fakes/fake_controller.go . Controller

type Controller interface {
	GetCatalog(logger log.Logger) (model.Catalog, error)
	CreateServiceInstance(logger log.Logger, serverInstanceId string, instance model.ServiceInstance) (model.CreateServiceInstanceResponse, error)
	ServiceInstanceExists(logger log.Logger, serviceInstanceId string) bool
	ServiceInstancePropertiesMatch(logger log.Logger, serviceInstanceId string, instance model.ServiceInstance) bool
	DeleteServiceInstance(logger log.Logger, serviceInstanceId string) error
	BindServiceInstance(logger log.Logger, serverInstanceId string, bindingId string, bindingInfo model.ServiceBinding) (model.CreateServiceBindingResponse, error)
	ServiceBindingExists(logger log.Logger, serviceInstanceId string, bindingId string) bool
	ServiceBindingPropertiesMatch(logger log.Logger, serviceInstanceId string, bindingId string, binding model.ServiceBinding) bool
	GetBinding(logger log.Logger, instanceId, bindingId string) (model.ServiceBinding, error)
	UnbindServiceInstance(logger log.Logger, serverInstanceId string, bindingId string) error
}

type StorageBackend interface {
	GetServices() []model.Service
	CreateVolume(serviceInstance model.ServiceInstance, name string, opts map[string]interface{}) error
	RemoveVolume(serviceInstance model.ServiceInstance, name string) error
	ListVolumes(serviceInstance model.ServiceInstance) ([]model.VolumeMetadata, error)
	GetVolume(serviceInstance model.ServiceInstance, name string) (volumeMetadata *model.VolumeMetadata, clientDriverName string, config *map[string]interface{}, err error)
}

type controller struct {
	backends    map[*model.Service]StorageBackend
	log         *log.Logger
	instanceMap map[string]*model.ServiceInstance
	bindingMap  map[string]*model.ServiceBinding
	configPath  string
}

func NewController(backends map[*model.Service]StorageBackend, configPath string) Controller {

	existingServiceInstances, err := loadServiceInstances(configPath)
	if err != nil {
		log.Fatal(fmt.Sprintf("error reading existing service instances: %s", err.Error()))
	}
	for _, existingServiceInstance := range existingServiceInstances {
		_, err := getServiceById(backends, existingServiceInstance.ServiceId)
		if err != nil {
			log.Fatal(fmt.Sprintf("error reading existing service instances: service instance refers to non-existing or disabled service (ServiceId: %s)", existingServiceInstance.ServiceId))
		}
	}

	existingServiceBindings, err := loadServiceBindings(configPath)
	if err != nil {
		log.Fatal(fmt.Sprintf("error reading existing service bindings: %s", err.Error()))
	}
	for _, existingServiceBinding := range existingServiceBindings {
		_, err = getServiceById(backends, existingServiceBinding.ServiceId)
		if err != nil {
			log.Fatal(fmt.Sprintf("error reading existing service bindings: service binding refers to non-existing or disabled service (ServiceId: %s)", existingServiceBinding.ServiceId))
		}
	}

	return &controller{backends: backends, configPath: configPath, instanceMap: existingServiceInstances, bindingMap: existingServiceBindings}
}

func (c *controller) GetCatalog(logger log.Logger) (model.Catalog, error) {
	allServices := make([]model.Service, 0, len(c.backends))
	for service := range c.backends {
		allServices = append(allServices, *service)
	}
	catalog := model.Catalog{Services: allServices}
	return catalog, nil
}

func (c *controller) CreateServiceInstance(logger log.Logger, serviceInstanceId string, instance model.ServiceInstance) (model.CreateServiceInstanceResponse, error) {
	service, err := getServiceById(c.backends, instance.ServiceId)
	if err != nil {
		return model.CreateServiceInstanceResponse{}, err
	}
	if err := c.backends[service].CreateVolume(instance, serviceInstanceId, nil); err != nil {
		return model.CreateServiceInstanceResponse{}, err
	}

	instance.DashboardUrl = "http://dashboard_url"
	instance.Id = serviceInstanceId
	instance.LastOperation = &model.LastOperation{
		State:                    "in progress",
		Description:              "creating service instance...",
		AsyncPollIntervalSeconds: DEFAULT_POLLING_INTERVAL_SECONDS,
	}

	c.instanceMap[serviceInstanceId] = &instance

	if err := persistServiceInstances(c.configPath, c.instanceMap); err != nil {
		return model.CreateServiceInstanceResponse{}, err
	}

	response := model.CreateServiceInstanceResponse{
		DashboardUrl:  instance.DashboardUrl,
		LastOperation: instance.LastOperation,
	}

	return response, nil
}

func (c *controller) ServiceInstanceExists(logger log.Logger, serviceInstanceId string) bool {
	_, exists := c.instanceMap[serviceInstanceId]
	return exists
}

func (c *controller) ServiceInstancePropertiesMatch(logger log.Logger, serviceInstanceId string, instance model.ServiceInstance) bool {
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

func (c *controller) DeleteServiceInstance(logger log.Logger, serviceInstanceId string) error {
	serviceInstance := c.instanceMap[serviceInstanceId]
	service, err := getServiceById(c.backends, (*serviceInstance).ServiceId)
	if err != nil {
		return err
	}
	if err := c.backends[service].RemoveVolume(*serviceInstance, serviceInstanceId); err != nil {
		return err
	}

	delete(c.instanceMap, serviceInstanceId)

	if err := persistServiceInstances(c.configPath, c.instanceMap); err != nil {
		return err
	}
	return nil
}

func (c *controller) BindServiceInstance(logger log.Logger, serviceInstanceId string, bindingId string, bindingInfo model.ServiceBinding) (model.CreateServiceBindingResponse, error) {
	serviceInstance := c.instanceMap[serviceInstanceId]
	service, err := getServiceById(c.backends, (*serviceInstance).ServiceId)
	if err != nil {
		return model.CreateServiceBindingResponse{}, err
	}

	c.bindingMap[bindingId] = &bindingInfo
	_, clientDriverName, config, err := c.backends[service].GetVolume(*serviceInstance, serviceInstanceId)
	if err != nil {
		return model.CreateServiceBindingResponse{}, err
	}
	containerMountPath := determineContainerMountPath(bindingInfo.Parameters, serviceInstanceId)

	configJson, err := json.Marshal(*config)
	if err != nil{
		return model.CreateServiceBindingResponse{}, err
	}

	privateDetails := model.VolumeMountPrivateDetails{Driver: clientDriverName, GroupId: serviceInstanceId, Config: string(configJson)}
	volumeMount := model.VolumeMount{ContainerPath: containerMountPath, Mode: "rw", Private: privateDetails}
	volumeMounts := []model.VolumeMount{volumeMount}

	if err = persistServiceBindings(c.configPath, c.bindingMap); err != nil {
		return model.CreateServiceBindingResponse{}, err
	}

	createBindingResponse := model.CreateServiceBindingResponse{VolumeMounts: volumeMounts}
	return createBindingResponse, nil
}

func (c *controller) ServiceBindingExists(logger log.Logger, serviceInstanceId string, bindingId string) bool {
	_, exists := c.bindingMap[bindingId]
	return exists
}

func (c *controller) ServiceBindingPropertiesMatch(logger log.Logger, serviceInstanceId string, bindingId string, binding model.ServiceBinding) bool {
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

func (c *controller) GetBinding(logger log.Logger, instanceId, bindingId string) (model.ServiceBinding, error) {
	binding, exists := c.bindingMap[bindingId]
	if exists == true {
		return *binding, nil
	}
	return model.ServiceBinding{}, fmt.Errorf("binding not found")

}

func (c *controller) UnbindServiceInstance(logger log.Logger, serverInstanceId string, bindingId string) error {
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

func getServiceById(backendsMap map[*model.Service]StorageBackend, serviceId string) (*model.Service, error) {
	for service := range backendsMap {
		if (*service).Id == serviceId {
			return service, nil
		}
	}
	return nil, errors.New(fmt.Sprintf("Could not locate service for serviceId %s", serviceId))
}

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