package block_device_utils

import (
    "github.com/IBM/ubiquity/utils/logs"
    "github.com/IBM/ubiquity/utils"
)

type blockDeviceUtils struct {
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
    return &blockDeviceUtils{logger: logs.GetLogger(), exec: executor}
}

