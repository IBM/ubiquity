package backends

import (
	"log"
	"fmt"

	common "github.ibm.com/almaden-containers/spectrum-common.git/core"
	"github.ibm.com/almaden-containers/ibm-storage-broker.git/model"
	"strings"
)

type SpectrumBackend struct {
	logger *log.Logger
	mountpoint string
}

func NewSpectrumBackend(logger *log.Logger, mountpoint string) *SpectrumBackend {
	return &SpectrumBackend{logger: logger, mountpoint: mountpoint}
}

func (s *SpectrumBackend) GetServices() []model.Service {
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

	return []model.Service{service}
}

func (s *SpectrumBackend) CreateVolume(serviceInstance model.ServiceInstance, name string, opts map[string]interface{}) error {
	client := s.getSpectrumClient(serviceInstance)
	return client.Create(name, opts)
}

func (s *SpectrumBackend) RemoveVolume(serviceInstance model.ServiceInstance, name string) error {
	client := s.getSpectrumClient(serviceInstance)
	return client.Remove(name)
}

func (s *SpectrumBackend) ListVolumes(serviceInstance model.ServiceInstance) ([]model.VolumeMetadata, error){
	client := s.getSpectrumClient(serviceInstance)
	spectrumVolumeMetaData, err := client.List()

	volumeMetaData := make([]model.VolumeMetadata, len(spectrumVolumeMetaData))
	for i, e := range spectrumVolumeMetaData {
		volumeMetaData[i] = model.VolumeMetadata{
			Name: e.Name,
			Mountpoint: e.Mountpoint,
		}
	}

	return volumeMetaData, err
}

func (s *SpectrumBackend) GetVolume(serviceInstance model.ServiceInstance, name string) (volumeMetadata *model.VolumeMetadata, clientDriverName string, config *map[string]interface{}, err error) {
	client := s.getSpectrumClient(serviceInstance)
	spectrumVolumeMetaData, spectrumConfig, err := client.Get(name)

	volumeMetadata = &model.VolumeMetadata {
		Name: spectrumVolumeMetaData.Name,
		Mountpoint: spectrumVolumeMetaData.Mountpoint,
	}

	configMap := make(map[string]interface{})
	configMap["fileset"] = spectrumConfig.FilesetId
	configMap["filesystem"] = spectrumConfig.Filesystem
	clientDriverName = fmt.Sprintf("spectrum-scale-%s", spectrumConfig.Filesystem)

	return volumeMetadata, clientDriverName, &configMap, err
}

func (s *SpectrumBackend) getSpectrumClient(serviceInstance model.ServiceInstance) common.SpectrumClient {
	// TODO: clean up usage of planId for plan name
	planIdSplit := strings.Split(serviceInstance.PlanId, "-")
	planName := planIdSplit[len(planIdSplit)-1]
	return common.NewSpectrumClient(s.logger, planName, fmt.Sprintf("%s/%s", s.mountpoint, planName))
}
