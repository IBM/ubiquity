package block_device_mounter_utils

import (
	"github.com/IBM/ubiquity/remote/mounter/block_device_utils"
	"github.com/IBM/ubiquity/logutil"
)

//go:generate counterfeiter -o ../fakes/fake_block_device_mounter_utils.go . BlockDeviceMounterUtils
type BlockDeviceMounterUtils interface {
	RescanAll(withISCSI bool) error
	MountDeviceFlow(devicePath string, fsType string, mountPoint string) error
	Discover(volumeWwn string) (string, error)
	UnmountDeviceFlow(devicePath string) error
}

type blockDeviceMounterUtils struct {
	logger               logutil.Logger
	blockDeviceUtils     block_device_utils.BlockDeviceUtils
}

func NewBlockDeviceMounterUtilsWithBlockDeviceUtils(blockDeviceUtilsInst block_device_utils.BlockDeviceUtils) BlockDeviceMounterUtils {
	mounterUtils := blockDeviceMounterUtils{blockDeviceUtils: blockDeviceUtilsInst}
	mounterUtils.logger = logutil.GetLogger()
	return &mounterUtils
}

func NewBlockDeviceMounterUtils() BlockDeviceMounterUtils {
	mounterUtils := blockDeviceMounterUtils{blockDeviceUtils: block_device_utils.NewBlockDeviceUtils()}
	mounterUtils.logger = logutil.GetLogger()
	return &mounterUtils
}
