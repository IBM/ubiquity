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
    Rescan() (error)
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
    protocol Protocol
}


func GetBlockDeviceUtils(logger *log.Logger, protocol Protocol) BlockDeviceUtils {
    return &impBlockDeviceUtils{logger: logger, exec: utils.NewExecutor(logger), protocol: protocol}
}

func GetBlockDeviceUtilsWithExecutor(logger *log.Logger, protocol Protocol, executor utils.Executor) BlockDeviceUtils {
    return &impBlockDeviceUtils{logger: logger, exec: executor, protocol: protocol}
}