package block_device_utils


import (
    "log"
    "github.com/IBM/ubiquity/utils"
)


type Protocol int


const (
    SCSI Protocol = iota
    ISCSI
)


type BlockDeviceUtils interface {
    Rescan(protocol Protocol) (error)
    ReloadMultipath() (error)
    Discover(volumeWwn string) (string, error)
    Cleanup(mpath string) (error)
    CheckFs(mpath string) (bool, error)
    MakeFs(mpath string, fsType string) (error)
    MountFs(mpath string, mpoint string) (error)
    UmountFs(mpoint string) (error)
}


type impBlockDeviceUtils struct {
    logger *log.Logger
    exec utils.Executor
}


func NewBlockDeviceUtils(logger *log.Logger) BlockDeviceUtils {
    return &impBlockDeviceUtils{logger: logger, exec: utils.NewExecutor(logger)}
}

func NewBlockDeviceUtilsWithExecutor(logger *log.Logger, executor utils.Executor) BlockDeviceUtils {
    return &impBlockDeviceUtils{logger: logger, exec: executor}
}