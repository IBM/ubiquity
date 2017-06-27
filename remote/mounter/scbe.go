package mounter

import (
	"fmt"
	"github.com/IBM/ubiquity/utils/logs"
	"github.com/IBM/ubiquity/remote/mounter/block_device_mounter_utils"
	"github.com/IBM/ubiquity/resources"
	"github.com/IBM/ubiquity/utils"
)

type scbeMounter struct {
	logger                  logs.Logger
	blockDeviceMounterUtils block_device_mounter_utils.BlockDeviceMounterUtils
	exec                    utils.Executor
}

func NewScbeMounter() resources.Mounter {
	blockDeviceMounterUtils := block_device_mounter_utils.NewBlockDeviceMounterUtils()
	return &scbeMounter{
		logger:                  logs.GetLogger(),
		blockDeviceMounterUtils: blockDeviceMounterUtils,
		exec: utils.NewExecutor(),
	}
}

func (s *scbeMounter) Mount(mountRequest resources.MountRequest) (string, error) {
	defer s.logger.Trace(logs.DEBUG)()

	// Rescan OS
	if err := s.blockDeviceMounterUtils.RescanAll(true); err != nil {
		return "", s.logger.ErrorRet(err, "RescanAll failed")
	}

	// Discover device
	volumeWWN := mountRequest.VolumeConfig["Wwn"].(string) // TODO use the const from local/scbe
	devicePath, err := s.blockDeviceMounterUtils.Discover(volumeWWN)
	if err != nil {
		return "", s.logger.ErrorRet(err, "Discover failed", logs.Args{{"volumeWWN", volumeWWN}})
	}

	// Create mount point if needed   // TODO consider to move it inside the util
	if _, err := s.exec.Stat(mountRequest.Mountpoint); err != nil {
		s.logger.Info("Create mountpoint directory " + mountRequest.Mountpoint)
		if err := s.exec.MkdirAll(mountRequest.Mountpoint, 0700); err != nil {
			return "", s.logger.ErrorRet(err, "MkdirAll failed", logs.Args{{"mountpoint", mountRequest.Mountpoint}})
		}
	}

	// Mount device and mkfs if needed
	fstype := resources.DefaultForScbeConfigParamDefaultFilesystem // TODO uses volumeConfig['fstype']
	if err := s.blockDeviceMounterUtils.MountDeviceFlow(devicePath, fstype, mountRequest.Mountpoint); err != nil {
		return "", s.logger.ErrorRet(err, "MountDeviceFlow failed", logs.Args{{"devicePath", devicePath}})
	}

	return mountRequest.Mountpoint, nil
}

func (s *scbeMounter) Unmount(unmountRequest resources.UnmountRequest) error {
	defer s.logger.Trace(logs.DEBUG)()

	volumeWWN := unmountRequest.VolumeConfig["Wwn"].(string)
	mountpoint := fmt.Sprintf(resources.PathToMountUbiquityBlockDevices, volumeWWN)
	devicePath, err := s.blockDeviceMounterUtils.Discover(volumeWWN)
	if err != nil {
		return s.logger.ErrorRet(err, "Discover failed", logs.Args{{"volumeWWN", volumeWWN}})
	}

	if err := s.blockDeviceMounterUtils.UnmountDeviceFlow(devicePath); err != nil {
		return s.logger.ErrorRet(err, "UnmountDeviceFlow failed", logs.Args{{"devicePath", devicePath}})
	}

	s.logger.Info("Delete mountpoint directory if exist", logs.Args{{"mountpoint", mountpoint}})
	// TODO move this part to the util
	if _, err := s.exec.Stat(mountpoint); err == nil {
		// TODO consider to add the prefix of the wwn in the OS (multipath -ll output)
		if err := s.exec.RemoveAll(mountpoint); err != nil {
			return s.logger.ErrorRet(err, "RemoveAll failed", logs.Args{{"mountpoint", mountpoint}})
		}
	}

	return nil

}

func (s *scbeMounter) ActionAfterDetach(request resources.AfterDetachRequest) error {
	defer s.logger.Trace(logs.DEBUG)()

	// Rescan OS
	if err := s.blockDeviceMounterUtils.RescanAll(true); err != nil {
		return s.logger.ErrorRet(err, "RescanAll failed")
	}
	return nil
}
