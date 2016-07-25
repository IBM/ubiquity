package backends

import (
	common "github.ibm.com/almaden-containers/spectrum-common.git/core"
	"github.ibm.com/almaden-containers/ibm-storage-broker.git/model"
	"log"
)

type SpectrumBackend struct {
	client common.SpectrumClient
}

func NewSpectrumBackend(logger *log.Logger, servicePlan, mountpoint *string) *SpectrumBackend {
	return &SpectrumBackend{common.NewSpectrumClient(logger, *servicePlan, *mountpoint)}
}

func (s *SpectrumBackend) Create(name string, opts map[string]interface{}) error {
	return s.client.Create(name, opts)
}

func (s *SpectrumBackend) Remove(name string) error {
	return s.client.Remove(name)
}

func (s *SpectrumBackend) Attach(name string) (string, error){
	return s.client.Attach(name)
}

func (s *SpectrumBackend) Detach(name string) error{
	return s.client.Detach(name)
}

func (s *SpectrumBackend) List() ([]model.VolumeMetadata, error){
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

func (s *SpectrumBackend) Get(name string) (volumeMetaData *model.VolumeMetadata, config *map[string]interface{}, err error) {
	spectrumVolumeMetaData, spectrumConfig, err := s.client.Get(name)

	volumeMetaData = &model.VolumeMetadata {
		Name: spectrumVolumeMetaData.Name,
		Mountpoint: spectrumVolumeMetaData.Mountpoint,
	}

	configMap := make(map[string]interface{})
	configMap["fileset"] = spectrumConfig.FilesetId
	configMap["filesystem"] = spectrumConfig.Filesystem

	return volumeMetaData, &configMap, err
}

func (s *SpectrumBackend) IsMounted() (bool, error){
	return s.client.IsMounted()
}

func (s *SpectrumBackend) Mount() error{
	return s.client.Mount()
}
