package mounter

import (
	"fmt"
	"github.com/IBM/ubiquity/logutil"
	"github.com/IBM/ubiquity/remote/mounter/block_device_mounter_utils"
	"github.com/IBM/ubiquity/resources"
	"github.com/IBM/ubiquity/utils"
)

type scbeMounter struct {
	logger                  logutil.Logger
	blockDeviceMounterUtils block_device_mounter_utils.BlockDeviceMounterUtils
}

func NewScbeMounter() Mounter {
	blockDeviceMounterUtils := block_device_mounter_utils.NewBlockDeviceMounterUtils()
	return &scbeMounter{logger: logutil.GetLogger(), blockDeviceMounterUtils: blockDeviceMounterUtils}
}

func (s *scbeMounter) Mount(mountpoint string, volumeConfig map[string]interface{}) (string, error) {
	defer s.logger.Trace(logutil.DEBUG)()

	// Rescan OS
	if err := s.blockDeviceMounterUtils.RescanAll(true); err != nil {
		return "", s.logger.ErrorRet(err, "RescanAll failed")
	}

	// Discover device
	volumeWWN := volumeConfig["wwn"].(string) // TODO use the const from local/scbe
	devicePath, err := s.blockDeviceMounterUtils.Discover(volumeWWN)
	if err != nil {
		return "", s.logger.ErrorRet(err, "Discover failed", logutil.Args{{"volumeWWN", volumeWWN}})
	}

	// Create mount point if needed   // TODO consider to move it inside the util
	exec := utils.NewExecutor()
	if _, err := exec.Stat(mountpoint); err != nil {
		s.logger.Info("Create mountpoint directory " + mountpoint)
		if err := exec.MkdirAll(mountpoint, 0700); err != nil {
			return "", s.logger.ErrorRet(err, "MkdirAll failed", logutil.Args{{"mountpoint", mountpoint}})
		}
	}

	// Mount device and mkfs if needed
	fstype := resources.DefaultForScbeConfigParamDefaultFilesystem // TODO uses volumeConfig['fstype']
	if err := s.blockDeviceMounterUtils.MountDeviceFlow(devicePath, fstype, mountpoint); err != nil {
		return "", s.logger.ErrorRet(err, "MountDeviceFlow failed", logutil.Args{{"devicePath", devicePath}})
	}

	return mountpoint, nil
}

func (s *scbeMounter) Unmount(volumeConfig map[string]interface{}) error {
	defer s.logger.Trace(logutil.DEBUG)()

	volumeWWN := volumeConfig["wwn"].(string)
	mountpoint := fmt.Sprintf(resources.PathToMountUbiquityBlockDevices, volumeWWN)
	devicePath, err := s.blockDeviceMounterUtils.Discover(volumeWWN)
	if err != nil {
		return s.logger.ErrorRet(err, "Discover failed", logutil.Args{{"volumeWWN", volumeWWN}})
	}

	if err := s.blockDeviceMounterUtils.UnmountDeviceFlow(devicePath); err != nil {
		return s.logger.ErrorRet(err, "UnmountDeviceFlow failed", logutil.Args{{"devicePath", devicePath}})
	}

	s.logger.Info("Delete mountpoint directory if exist", logutil.Args{{"mountpoint", mountpoint}})
	// TODO move this part to the util
	exec := utils.NewExecutor()
	if _, err := exec.Stat(mountpoint); err == nil {
		// TODO consider to add the prefix of the wwn in the OS (multipath -ll output)
		if err := exec.RemoveAll(mountpoint); err != nil {
			return s.logger.ErrorRet(err, "RemoveAll failed", logutil.Args{{"mountpoint", mountpoint}})
		}
	}

	return nil

}
func (s *scbeMounter) ActionAfterDetach(volumeConfig map[string]interface{}) error {
	defer s.logger.Trace(logutil.DEBUG)()

	// Rescan OS
	if err := s.blockDeviceMounterUtils.RescanAll(true); err != nil {
		return s.logger.ErrorRet(err, "RescanAll failed")
	}
	return nil
}
