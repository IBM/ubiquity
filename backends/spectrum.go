package backends

import (
	"log"
	"fmt"

	common "github.ibm.com/almaden-containers/spectrum-common.git/core"
	"github.ibm.com/almaden-containers/ibm-storage-broker.git/model"
)

type SpectrumBackend struct {
	logger *log.Logger
	client common.SpectrumClient
}

func NewSpectrumBackend(logger *log.Logger, servicePlan, mountpoint *string) *SpectrumBackend {
	return &SpectrumBackend{logger: logger, client: common.NewSpectrumClient(logger, *servicePlan, *mountpoint)}
}

func (s *SpectrumBackend) GetServices() []model.Service {
	plan1 := model.ServicePlan{
		Name:        "gold",
		Id:          "gold",
		Description: "Gold Tier Performance and Resiliency",
		Metadata:    nil,
		Free:        true,
	}

	plan2 := model.ServicePlan{
		Name:        "bronze",
		Id:          "bronze",
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

func (s *SpectrumBackend) CreateVolume(name string, opts map[string]interface{}) error {
	return s.client.Create(name, opts)
}

func (s *SpectrumBackend) RemoveVolume(name string) error {
	return s.client.Remove(name)
}

func (s *SpectrumBackend) ListVolumes() ([]model.VolumeMetadata, error){
	spectrumVolumeMetaData, err := s.client.List()

	volumeMetaData := make([]model.VolumeMetadata, len(spectrumVolumeMetaData))
	for i, e := range spectrumVolumeMetaData {
		volumeMetaData[i] = model.VolumeMetadata{
			Name: e.Name,
			Mountpoint: e.Mountpoint,
		}
	}

	return volumeMetaData, err
}

func (s *SpectrumBackend) GetVolume(name string) (volumeMetaData *model.VolumeMetadata, clientDriverName string, config *map[string]interface{}, err error) {
	spectrumVolumeMetaData, spectrumConfig, err := s.client.Get(name)

	volumeMetaData = &model.VolumeMetadata {
		Name: spectrumVolumeMetaData.Name,
		Mountpoint: spectrumVolumeMetaData.Mountpoint,
	}

	configMap := make(map[string]interface{})
	configMap["fileset"] = spectrumConfig.FilesetId
	configMap["filesystem"] = spectrumConfig.Filesystem
	clientDriverName = fmt.Sprintf("spectrum-scale-%s", spectrumConfig.Filesystem)

	return volumeMetaData, clientDriverName, &configMap, err
}
