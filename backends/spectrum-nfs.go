package backends

import (
	"github.ibm.com/almaden-containers/ibm-storage-broker.git/model"
	"log"
	"fmt"
)

type SpectrumNfsBackend struct {
	SpectrumBackend
	NfsServerAddr string
	NfsClientCIDR string
}

func NewSpectrumNfsBackend(logger *log.Logger, mountpoint, nfsServerAddr, nfsClientCIDR string) *SpectrumNfsBackend {
	return &SpectrumNfsBackend{NfsServerAddr: nfsServerAddr, NfsClientCIDR: nfsClientCIDR, SpectrumBackend: SpectrumBackend{logger, mountpoint}}
}

func (s *SpectrumNfsBackend) GetServices() []model.Service {
	plan1 := model.ServicePlan{
		Name:        "gold",
		Id:          "spectrum-scale-nfs-gold",
		Description: "Gold Tier Performance and Resiliency",
		Metadata:    nil,
		Free:        true,
	}

	plan2 := model.ServicePlan{
		Name:        "bronze",
		Id:          "spectrum-scale-nfs-bronze",
		Description: "Bronze Tier Performance and Resiliency",
		Metadata:    nil,
		Free:        true,
	}

	service := s.SpectrumBackend.GetServices()[0]

	service.Name = "spectrum-scale-nfs"
	service.Id = "spectrum-nfs-service-guid"
	service.Description = "Provides the Spectrum FS volume service using NFS transport"
	service.Tags = []string{"gpfs", "nfs"}
	service.Plans = []model.ServicePlan{plan1, plan2}

	return []model.Service{service}
}

func (s *SpectrumNfsBackend) CreateVolume(serviceInstance model.ServiceInstance, name string, opts map[string]interface{}) error {
	client := s.getSpectrumClient(serviceInstance)
	if err := s.SpectrumBackend.CreateVolume(serviceInstance, name, opts); err != nil {
		return err
	}
	if _, err := client.Attach(name); err != nil {
		return err
	}
	_, err := client.ExportNfs(name, s.NfsClientCIDR)
	return err
}

func (s *SpectrumNfsBackend) RemoveVolume(serviceInstance model.ServiceInstance, name string) error {
	client := s.getSpectrumClient(serviceInstance)
	if err := client.UnexportNfs(name); err != nil {
		s.logger.Printf("SpectrumNfsBackend: Could not unexport volume %s (error=%s)", name, err.Error())
	}
	if err := client.Detach(name); err != nil {
		s.logger.Printf("SpectrumNfsBackend: Could not detach volume %s (error=%s)", name, err.Error())
	}
	return s.SpectrumBackend.RemoveVolume(serviceInstance, name)
}

func (s *SpectrumNfsBackend) GetVolume(serviceInstance model.ServiceInstance, name string) (volumeMetadata *model.VolumeMetadata, clientDriverName string, config *map[string]interface{}, err error) {
	clientDriverName = "nfs-plugin"
	volumeMetadata, _, spectrumBackendConfig, err := s.SpectrumBackend.GetVolume(serviceInstance, name)
	nfsShare := fmt.Sprintf("%s:%s", s.NfsServerAddr, (*volumeMetadata).Mountpoint)
	(*spectrumBackendConfig)["nfs_share"] = nfsShare
	s.logger.Printf("Adding nfs_share %s to bind config\n", nfsShare)
	return volumeMetadata, clientDriverName, spectrumBackendConfig, err
}
