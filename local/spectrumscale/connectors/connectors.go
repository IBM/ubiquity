package connectors

import (
	"fmt"
	"log"

	"github.ibm.com/almaden-containers/ubiquity/model"
)

//go:generate counterfeiter -o ../fakes/fake_spectrum.go . SpectrumScaleConnector
type SpectrumScaleConnector interface {
	//Cluster operations
	GetClusterId() (string, error)
	//Filesystem operations
	IsFilesystemMounted(filesystemName string) (bool, error)
	MountFileSystem(filesystemName string) error
	ListFilesystems() ([]string, error)
	GetFilesystemMountpoint(filesystemName string) (string, error)
	//Fileset operations
	CreateFileset(filesystemName string, filesetName string, opts map[string]interface{}) error
	DeleteFileset(filesystemName string, filesetName string) error
	LinkFileset(filesystemName string, filesetName string) error
	UnlinkFileset(filesystemName string, filesetName string) error
	ListFilesets(filesystemName string) ([]model.VolumeMetadata, error)
	ListFileset(filesystemName string, filesetName string) (model.VolumeMetadata, error)
	IsFilesetLinked(filesystemName string, filesetName string) (bool, error)
	//TODO modify quota from string to Capacity (see kubernetes)
	ListFilesetQuota(filesystemName string, filesetName string) (string, error)
	SetFilesetQuota(filesystemName string, filesetName string, quota string) error
}

func GetSpectrumScaleConnector(logger *log.Logger, connector string, opts map[string]interface{}) (SpectrumScaleConnector, error) {
	if connector == "mmcli" {
		return NewSpectrumMMCLI(logger, opts)
	}
	if connector == "rest" {
		return NewSpectrumRest(logger, opts)
	}
	if connector == "mmcli" {
		return NewSpectrumSSH(logger, opts)
	} else {
		return nil, fmt.Errorf("This protocol is not recognized")
	}
}
