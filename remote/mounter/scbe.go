package mounter

import (
	"github.com/IBM/ubiquity/remote/mounter/block_device_mounter_utils"
	"github.com/IBM/ubiquity/utils"
	"github.com/IBM/ubiquity/logutil"
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
		s.logger.Error("RescanAll failed", logutil.Args{{"error", err}})
		return "", err
	}

	// Discover device
	volumeWWN := volumeConfig["wwn"].(string) // TODO use the const from local/scbe
	devicePath, err := s.blockDeviceMounterUtils.Discover(volumeWWN)
	if err != nil {
		s.logger.Error("Discover failed", logutil.Args{{"volumeWWN", volumeWWN}, {"error", err}})
		return "", err
	}

	// Create mount point if needed   // TODO consider to move it inside the util
	exec := utils.NewExecutor()
	if _, err := exec.Stat(mountpoint); err != nil {
		s.logger.Info("Create mountpoint directory " + mountpoint)
		if err := exec.MkdirAll(mountpoint, 0700); err != nil {
			s.logger.Error("MkdirAll failed", logutil.Args{{"mountpoint", mountpoint}, {"error", err}})
			return "", err
		}
	}

	// Mount device and mkfs if needed
	fstype := "ext4" // TODO uses volumeConfig['fstype']
	if err := s.blockDeviceMounterUtils.MountDeviceFlow(devicePath, fstype, mountpoint); err != nil {
		s.logger.Error("MountDeviceFlow failed", logutil.Args{{"devicePath", devicePath}, {"error", err}})
		return "", err
	}

	return mountpoint, nil
}

func (s *scbeMounter) Unmount(volumeConfig map[string]interface{}) error {
	defer s.logger.Trace(logutil.DEBUG)()

	volumeWWN := volumeConfig["wwn"].(string) // TODO use the const from local/scbe
	mountpoint := "/ubiquity/" + volumeWWN    // TODO get the ubiquity prefix from const
	devicePath, err := s.blockDeviceMounterUtils.Discover(volumeWWN)
	if err != nil {
		s.logger.Error("Discover failed", logutil.Args{{"volumeWWN", volumeWWN}, {"error", err}})
		return err
	}

	if err := s.blockDeviceMounterUtils.UnmountDeviceFlow(devicePath); err != nil {
		s.logger.Error("UnmountDeviceFlow failed", logutil.Args{{"devicePath", devicePath}, {"error", err}})
		return err
	}


	s.logger.Info("Delete mountpoint directory if exist", logutil.Args{{"mountpoint", mountpoint}})
	// TODO move this part to the util
	exec := utils.NewExecutor()
	if _, err := exec.Stat(mountpoint); err == nil {
		// TODO consider to add the prefix of the wwn in the OS (multipath -ll output)
		if err := exec.RemoveAll(mountpoint); err != nil {
			s.logger.Error("RemoveAll failed", logutil.Args{{"mountpoint", mountpoint}, {"error", err}})
			return err
		}
	}

	return nil

}
func (s *scbeMounter) ActionAfterDetach(volumeConfig map[string]interface{}) error {
	defer s.logger.Trace(logutil.DEBUG)()

	// Rescan OS
	if err := s.blockDeviceMounterUtils.RescanAll(true); err != nil {
		s.logger.Error("RescanAll failed", logutil.Args{{"error", err}})
		return err
	}
	return nil
}
