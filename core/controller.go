package core

import (
	"fmt"
	"log"
	"path"
	"reflect"

	"github.ibm.com/almaden-containers/ibm-storage-broker.git/model"
	"github.ibm.com/almaden-containers/ibm-storage-broker.git/utils"
	"encoding/json"
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
	Create(name string, opts map[string]interface{}) error
	Remove(name string) error
	Attach(name string) (string, error)
	Detach(name string) error
	List() ([]model.VolumeMetadata, error)
	Get(name string) (volumeMetadata *model.VolumeMetadata, config *map[string]interface{}, err error)
	IsMounted() (bool, error)
	Mount() error
}

type controller struct {
	client      StorageBackend
	log         *log.Logger
	instanceMap map[string]*model.ServiceInstance
	bindingMap  map[string]*model.ServiceBinding
	configPath  string
}

func NewController(storageBackend StorageBackend, configPath string, instanceMap map[string]*model.ServiceInstance, bindingMap map[string]*model.ServiceBinding) Controller {
	return &controller{client: storageBackend, configPath: configPath, instanceMap: instanceMap, bindingMap: bindingMap}
}

func (c *controller) GetCatalog(logger log.Logger) (model.Catalog, error) {
	plan := model.ServicePlan{
		Name:        "free",
		Id:          "free-plan-guid",
		Description: "free ibm-storage-broker filesystem",
		Metadata:    nil,
		Free:        true,
	}

	service := model.Service{
		Name:            "ibm-storage-broker",
		Id:              "ibm-storage-broker-guid",
		Description:     "Provides the ibm-storage-broker volume service, including volume creation and volume mounts",
		Bindable:        true,
		PlanUpdateable:  false,
		Tags:            []string{},
		Requires:        []string{"volume_mount"},
		Metadata:        nil,
		Plans:           []model.ServicePlan{plan},
		DashboardClient: nil,
	}

	catalog := model.Catalog{
		Services: []model.Service{service},
	}

	return catalog, nil
}

func (c *controller) CreateServiceInstance(logger log.Logger, serviceInstanceId string, instance model.ServiceInstance) (model.CreateServiceInstanceResponse, error) {
	err := c.client.Create(serviceInstanceId, nil)
	if err != nil {
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
	err = utils.MarshalAndRecord(c.instanceMap, c.configPath, "service_instances.json")
	if err != nil {
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
	err := c.client.Remove(serviceInstanceId)
	if err != nil {
		return err
	}
	delete(c.instanceMap, serviceInstanceId)
	err = utils.MarshalAndRecord(c.instanceMap, c.configPath, "service_instances.json")
	if err != nil {
		return err
	}
	return nil
}
func (c *controller) BindServiceInstance(logger log.Logger, serviceInstanceId string, bindingId string, bindingInfo model.ServiceBinding) (model.CreateServiceBindingResponse, error) {
	c.bindingMap[bindingId] = &bindingInfo
	_, config, err := c.client.Get(serviceInstanceId)
	if err != nil {
		return model.CreateServiceBindingResponse{}, err
	}
	containerMountPath := determineContainerMountPath(bindingInfo.Parameters, serviceInstanceId)

	configJson,err := json.Marshal(*config)
	if err != nil{
		return model.CreateServiceBindingResponse{}, err
	}

	backendName := reflect.TypeOf(c.client).Name()

	privateDetails := model.VolumeMountPrivateDetails{Driver: backendName, GroupId: serviceInstanceId, Config: string(configJson)}
	volumeMount := model.VolumeMount{ContainerPath: containerMountPath, Mode: "rw", Private: privateDetails}
	volumeMounts := []model.VolumeMount{volumeMount}

	createBindingResponse := model.CreateServiceBindingResponse{VolumeMounts: volumeMounts}
	err = utils.MarshalAndRecord(c.bindingMap, c.configPath, "service_bindings.json")
	if err != nil {
		return model.CreateServiceBindingResponse{}, err
	}
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
