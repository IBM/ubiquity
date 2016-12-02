package spectrum

import (
	"log"

	"github.ibm.com/almaden-containers/ubiquity/model"
)

type spectrum_ssh struct {
	logger *log.Logger
}

func NewSpectrumSSH(logger *log.Logger, opts map[string]interface{}) Spectrum {
	return &spectrum_ssh{logger: logger}
}

func (s *spectrum_ssh) GetClusterId() (string, error) {
	return "", nil
}
func (s *spectrum_ssh) IsFilesystemMounted(filesystemName string) (bool, error) {
	return true, nil
}
func (s *spectrum_ssh) MountFileSystem(filesystemName string) error {
	return nil
}
func (s *spectrum_ssh) ListFilesystems() ([]string, error) {
	return nil, nil
}
func (s *spectrum_ssh) GetFilesystemMountpoint(filesystemName string) (string, error) {
	return "", nil
}
func (s *spectrum_ssh) CreateFileset(filesystemName string, fileSetName string, opts map[string]interface{}) error {
	return nil
}
func (s *spectrum_ssh) DeleteFileset(filesystemName string, filesetName string) error { return nil }
func (s *spectrum_ssh) LinkFileset(filesystemName string, filesetName string) error   { return nil }
func (s *spectrum_ssh) UnlinkFileset(filesystemName string, filesetName string) error { return nil }
func (s *spectrum_ssh) ListFilesets(filesystemName string) ([]model.VolumeMetadata, error) {
	return nil, nil
}
func (s *spectrum_ssh) ListFileset(filesystemName string, filesetName string) (model.VolumeMetadata, error) {
	return model.VolumeMetadata{}, nil
}

func (s *spectrum_ssh) IsFilesetLinked(filesystemName string, filesetName string) (bool, error) {
	return true, nil
}

//TODO modify quota from string to Capacity (see kubernetes)
func (s *spectrum_ssh) ListFilesetQuota(filesystemName string, filesetName string) (string, error) {
	return "", nil
}
func (s *spectrum_ssh) SetFilesetQuota(filesystemName string, filesetName string, quota string) error {
	return nil
}
