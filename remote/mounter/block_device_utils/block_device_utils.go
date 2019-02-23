package block_device_utils

import (
	"regexp"

	"github.com/IBM/ubiquity/remote/mounter/initiator"
	"github.com/IBM/ubiquity/remote/mounter/initiator/connectors"
	"github.com/IBM/ubiquity/utils"
	"github.com/IBM/ubiquity/utils/logs"
)

type blockDeviceUtils struct {
	logger              logs.Logger
	exec                utils.Executor
	regExAlreadyMounted *regexp.Regexp
	fcConnector         initiator.Connector
}

func NewBlockDeviceUtils() BlockDeviceUtils {
	return newBlockDeviceUtils(utils.NewExecutor(), nil)
}

func NewBlockDeviceUtilsWithExecutor(executor utils.Executor) BlockDeviceUtils {
	return newBlockDeviceUtils(executor, nil)
}

func NewBlockDeviceUtilsWithExecutorAndConnector(executor utils.Executor, fcConnector initiator.Connector) BlockDeviceUtils {
	return newBlockDeviceUtils(executor, fcConnector)
}

func newBlockDeviceUtils(executor utils.Executor, fcConnector initiator.Connector) BlockDeviceUtils {
	logger := logs.GetLogger()

	// Prepare regex that going to be used in unmount interface
	pattern := "(?i)" + NotMountedErrorMessage
	regex, err := regexp.Compile(pattern)
	if err != nil {
		panic("failed prepare Already unmount regex")
	}

	if fcConnector == nil {
		fcConnector = connectors.NewFibreChannelConnectorWithExecutorAndLogger(executor, logger)
	}
	return &blockDeviceUtils{logger: logger, exec: executor, regExAlreadyMounted: regex, fcConnector: fcConnector}
}
