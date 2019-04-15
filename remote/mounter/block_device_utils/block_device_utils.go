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
	iscsiConnector      initiator.Connector
}

func NewBlockDeviceUtils() BlockDeviceUtils {
	return newBlockDeviceUtils(utils.NewExecutor())
}

func NewBlockDeviceUtilsWithExecutor(executor utils.Executor) BlockDeviceUtils {
	return newBlockDeviceUtils(executor)
}

func NewBlockDeviceUtilsWithExecutorAndConnector(executor utils.Executor, conns ...initiator.Connector) BlockDeviceUtils {
	return newBlockDeviceUtils(executor, conns...)
}

func newBlockDeviceUtils(executor utils.Executor, conns ...initiator.Connector) BlockDeviceUtils {
	logger := logs.GetLogger()

	// Prepare regex that going to be used in unmount interface
	pattern := "(?i)" + NotMountedErrorMessage
	regex, err := regexp.Compile(pattern)
	if err != nil {
		panic("failed prepare Already unmount regex")
	}

	var fcConnector, iscsiConnector initiator.Connector

	if len(conns) == 0 {
		fcConnector = connectors.NewFibreChannelConnectorWithExecutor(executor)
		iscsiConnector = connectors.NewISCSIConnectorWithExecutor(executor)
	} else {
		fcConnector = conns[0]
		iscsiConnector = conns[1]
	}

	return &blockDeviceUtils{logger: logger, exec: executor, regExAlreadyMounted: regex, fcConnector: fcConnector, iscsiConnector: iscsiConnector}
}
