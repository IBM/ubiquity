package block_device_utils

import (
    "github.com/IBM/ubiquity/utils/logs"
    "github.com/IBM/ubiquity/utils"
    "regexp"
)

type blockDeviceUtils struct {
    logger logs.Logger
    exec   utils.Executor
    regExAlreadyMounted *regexp.Regexp
}

func NewBlockDeviceUtils() BlockDeviceUtils {
    return newBlockDeviceUtils(utils.NewExecutor())
}

func NewBlockDeviceUtilsWithExecutor(executor utils.Executor) BlockDeviceUtils {
    return newBlockDeviceUtils(executor)
}

func newBlockDeviceUtils(executor utils.Executor) BlockDeviceUtils {
    logger := logs.GetLogger()

    // Prepare regex that going to be used in unmount interface
    pattern := "(?i)" + NotMountedErrorMessage
    regex, err := regexp.Compile(pattern)
    if err != nil {
        panic("failed prepare Already unmount regex")
    }

    return &blockDeviceUtils{logger: logger, exec: executor, regExAlreadyMounted: regex}
}

