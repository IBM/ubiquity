package mounter

import (
	"fmt"
	"log"

	"github.com/IBM/ubiquity/resources"
)

//go:generate counterfeiter -o ../../fakes/fake_mounter.go . Mounter

type Mounter interface {
	Mount(mountpoint string, volumeConfig map[string]interface{}) (string, error)
	Unmount(volumeConfig map[string]interface{}) error
}

func GetMounterForVolume(logger *log.Logger, volume resources.Volume) (Mounter, error) {
	if volume.Backend == resources.SPECTRUM_SCALE {
		return NewSpectrumScaleMounter(logger), nil
	} else if volume.Backend == resources.SOFTLAYER_NFS || volume.Backend == resources.SPECTRUM_SCALE_NFS {
		return NewNfsMounter(logger), nil
	}
	return nil, fmt.Errorf("Mounter not found for volume: %s", volume.Name)
}
