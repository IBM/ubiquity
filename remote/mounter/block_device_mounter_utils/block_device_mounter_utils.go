package block_device_mounter_utils

import (
	"github.com/IBM/ubiquity/remote/mounter/block_device_utils"
	"github.com/IBM/ubiquity/utils/logs"
)

type blockDeviceMounterUtils struct {
	logger               logs.Logger
	blockDeviceUtils     block_device_utils.BlockDeviceUtils
}

// MountDeviceFlow create filesystem on the device (if needed) and then mount it on a given mountpoint
func (s *blockDeviceMounterUtils) MountDeviceFlow(devicePath string, fsType string, mountPoint string) error {
	defer s.logger.Trace(logs.INFO, logs.Args{{"devicePath", devicePath}, {"fsType", fsType}, {"mountPoint", mountPoint}})()

	needToCreateFS, err := s.blockDeviceUtils.CheckFs(devicePath)
	if err != nil {
		return s.logger.ErrorRet(err, "CheckFs failed")
	}
	if needToCreateFS {
		if err = s.blockDeviceUtils.MakeFs(devicePath, fsType); err != nil {
			return s.logger.ErrorRet(err, "MakeFs failed")
		}
	}
	if err = s.blockDeviceUtils.MountFs(devicePath, mountPoint); err != nil {
		return s.logger.ErrorRet(err, "MountFs failed")
	}

	return nil
}

// UnmountDeviceFlow umount device, clean device and remove mountpoint folder
func (s *blockDeviceMounterUtils) UnmountDeviceFlow(devicePath string) error {
	defer s.logger.Trace(logs.INFO, logs.Args{{"devicePath", devicePath}})

	err := s.blockDeviceUtils.UmountFs(devicePath)
	if err != nil {
		return s.logger.ErrorRet(err, "UmountFs failed")
	}

	if err := s.blockDeviceUtils.Cleanup(devicePath); err != nil {
		return s.logger.ErrorRet(err, "Cleanup failed")
	}

	// TODO delete the directory here
	return nil
}

// RescanAll triggers the following OS rescanning :
// 1. iSCSI rescan (if protocol given is iscsi)
// 2. SCSI rescan
// 3. multipathing rescan
// return error if one of the steps fail
func (s *blockDeviceMounterUtils) RescanAll(withISCSI bool) error {
	defer s.logger.Trace(logs.INFO, logs.Args{{"withISCSI", withISCSI}})

	if withISCSI {
		if err := s.blockDeviceUtils.Rescan(block_device_utils.ISCSI); err != nil {
			return s.logger.ErrorRet(err, "Rescan failed", logs.Args{{"protocol", block_device_utils.ISCSI}})
		}
	}
	if err := s.blockDeviceUtils.Rescan(block_device_utils.SCSI); err != nil {
		return s.logger.ErrorRet(err, "Rescan failed", logs.Args{{"protocol", block_device_utils.SCSI}})
	}

	if err := s.blockDeviceUtils.ReloadMultipath(); err != nil {
		return s.logger.ErrorRet(err, "ReloadMultipath failed")
	}
	return nil
}

func (s *blockDeviceMounterUtils) Discover(volumeWwn string) (string, error) {
	return s.blockDeviceUtils.Discover(volumeWwn)
}
