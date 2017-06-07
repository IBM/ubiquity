package block_device_mounter_utils

import (
	"github.com/IBM/ubiquity/remote/mounter/block_device_utils"
	"log"
)

//go:generate counterfeiter -o ../fakes/fake_block_device_mounter_utils.go . BlockDeviceMounterUtils
type BlockDeviceMounterUtils interface {
	RescanAll(withISCSI bool) error
	MountDeviceFlow(devicePath string, fsType string, mountPoint string) error
	Discover(volumeWwn string) (string, error)
	UnmountDeviceFlow(devicePath string) error
}

type blockDeviceMounterUtils struct {
	logger               *log.Logger
	BlockDeviceUtilsInst block_device_utils.BlockDeviceUtils
}

func NewBlockDeviceMounterUtils(logger *log.Logger) BlockDeviceMounterUtils {
	return &blockDeviceMounterUtils{
		logger:               logger,
		BlockDeviceUtilsInst: block_device_utils.NewBlockDeviceUtils(logger),
	}
}

func NewBlockDeviceMounterUtilsWithExecutor(logger *log.Logger, blockDeviceUtils block_device_utils.BlockDeviceUtils) BlockDeviceMounterUtils {
	return &blockDeviceMounterUtils{
		logger:               logger,
		BlockDeviceUtilsInst: blockDeviceUtils,
	}
}
