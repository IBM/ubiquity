package connectors

import (
	"log"
	"github.com/IBM/ubiquity/resources"
)

//go:generate counterfeiter -o ../../../fakes/fake_spectrum.go . SpectrumScaleConnector
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
	ListFilesets(filesystemName string) ([]resources.VolumeMetadata, error)
	ListFileset(filesystemName string, filesetName string) (resources.VolumeMetadata, error)
	IsFilesetLinked(filesystemName string, filesetName string) (bool, error)
	//TODO modify quota from string to Capacity (see kubernetes)
	ListFilesetQuota(filesystemName string, filesetName string) (string, error)
	SetFilesetQuota(filesystemName string, filesetName string, quota string) error
}

const (
	USER_SPECIFIED_FILESET_TYPE string = "fileset-type"
	USER_SPECIFIED_INODE_LIMIT string = "inode-limit"
)

func GetSpectrumScaleConnector(logger *log.Logger, config resources.SpectrumScaleConfig) (SpectrumScaleConnector, error) {
	if config.RestConfig.Endpoint != "" {
		logger.Printf("Initializing SpectrumScale REST connector with restConfig: %+v\n", config.RestConfig)
		return NewSpectrumRest(logger, config.RestConfig)
	}
	if config.SshConfig.User != "" && config.SshConfig.Host != "" {
		if config.SshConfig.Port == "" || config.SshConfig.Port == "0" {
			config.SshConfig.Port = "22"
		}
		logger.Printf("Initializing SpectrumScale SSH connector with sshConfig: %+v\n", config.SshConfig)
		return NewSpectrumSSH(logger, config.SshConfig)
	}
	logger.Println("Initializing SpectrumScale MMCLI Connector")
	return NewSpectrumMMCLI(logger)
}
