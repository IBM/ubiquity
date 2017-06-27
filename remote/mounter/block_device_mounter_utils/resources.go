package block_device_mounter_utils

import (
	"github.com/IBM/ubiquity/utils/logs"
	"github.com/IBM/ubiquity/remote/mounter/block_device_utils"
)

//go:generate counterfeiter -o ../fakes/fake_block_device_mounter_utils.go . BlockDeviceMounterUtils
type BlockDeviceMounterUtils interface {
	RescanAll(withISCSI bool) error
	MountDeviceFlow(devicePath string, fsType string, mountPoint string) error
	Discover(volumeWwn string) (string, error)
	UnmountDeviceFlow(devicePath string) error
}

func NewBlockDeviceMounterUtilsWithBlockDeviceUtils(blockDeviceUtilsInst block_device_utils.BlockDeviceUtils) BlockDeviceMounterUtils {
	return newBlockDeviceMounterUtils(blockDeviceUtilsInst)
}

func NewBlockDeviceMounterUtils() BlockDeviceMounterUtils {
	return newBlockDeviceMounterUtils(block_device_utils.NewBlockDeviceUtils())
}

func newBlockDeviceMounterUtils(blockDeviceUtilsInst block_device_utils.BlockDeviceUtils) BlockDeviceMounterUtils {
	return &blockDeviceMounterUtils{logger: logs.GetLogger(), blockDeviceUtils: blockDeviceUtilsInst}
}
