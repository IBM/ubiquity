package connectors

import (
	"github.com/IBM/ubiquity/remote/mounter/initiator"
	"github.com/IBM/ubiquity/resources"
	"github.com/IBM/ubiquity/utils"
	"github.com/IBM/ubiquity/utils/logs"
)

type iscsiConnector struct {
	*scsiConnector
}

func NewISCSIConnector() initiator.Connector {
	return newISCSIConnector()
}

func NewISCSIConnectorWithExecutor(executor utils.Executor) initiator.Connector {
	logger := logs.GetLogger()
	return newISCSIConnectorWithExecutorAndLogger(executor, logger)
}

func NewISCSIConnectorWithAllFields(executor utils.Executor, initi initiator.Initiator) initiator.Connector {
	logger := logs.GetLogger()
	return &iscsiConnector{&scsiConnector{logger: logger, exec: executor, initi: initi}}
}

func newISCSIConnector() *iscsiConnector {
	executor := utils.NewExecutor()
	logger := logs.GetLogger()
	return newISCSIConnectorWithExecutorAndLogger(executor, logger)
}

func newISCSIConnectorWithExecutorAndLogger(executor utils.Executor, logger logs.Logger) *iscsiConnector {
	initi := initiator.NewLinuxISCSIWithExecutor(executor)
	return &iscsiConnector{&scsiConnector{logger: logger, exec: executor, initi: initi}}
}

func (c *iscsiConnector) ConnectVolume(volumeMountProperties *resources.VolumeMountProperties) error {
	return nil
}
