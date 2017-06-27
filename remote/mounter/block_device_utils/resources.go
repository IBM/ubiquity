package block_device_utils

import (
	"github.com/IBM/ubiquity/utils"
	"github.com/IBM/ubiquity/utils/logs"
)

type Protocol int

const (
	SCSI Protocol = iota
	ISCSI
)

//go:generate counterfeiter -o ../fakes/fake_block_device_utils.go . BlockDeviceUtils
type BlockDeviceUtils interface {
	Rescan(protocol Protocol) error
	ReloadMultipath() error
	Discover(volumeWwn string) (string, error)
	Cleanup(mpath string) error
	CheckFs(mpath string) (bool, error)
	MakeFs(mpath string, fsType string) error
	MountFs(mpath string, mpoint string) error
	UmountFs(mpoint string) error
}

type impBlockDeviceUtils struct {
	logger logs.Logger
	exec   utils.Executor
}

func NewBlockDeviceUtils() BlockDeviceUtils {
	return newBlockDeviceUtils(utils.NewExecutor())
}

func NewBlockDeviceUtilsWithExecutor(executor utils.Executor) BlockDeviceUtils {
	return newBlockDeviceUtils(executor)
}

func newBlockDeviceUtils(executor utils.Executor) BlockDeviceUtils {
	return &impBlockDeviceUtils{logger:logs.GetLogger(), exec: executor}
}