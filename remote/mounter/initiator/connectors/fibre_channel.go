package connectors

import (
	"github.com/IBM/ubiquity/remote/mounter/initiator"
	"github.com/IBM/ubiquity/resources"
	"github.com/IBM/ubiquity/utils"
	"github.com/IBM/ubiquity/utils/logs"
)

type fibreChannelConnector struct {
	*scsiConnector
}

func NewFibreChannelConnector() initiator.Connector {
	return newFibreChannelConnector()
}

func NewFibreChannelConnectorWithExecutor(executor utils.Executor) initiator.Connector {
	logger := logs.GetLogger()
	return newFibreChannelConnectorWithExecutorAndLogger(executor, logger)
}

func NewFibreChannelConnectorWithAllFields(executor utils.Executor, initi initiator.Initiator) initiator.Connector {
	logger := logs.GetLogger()
	return &fibreChannelConnector{&scsiConnector{logger: logger, exec: executor, initi: initi}}
}

func newFibreChannelConnector() *fibreChannelConnector {
	executor := utils.NewExecutor()
	logger := logs.GetLogger()
	return newFibreChannelConnectorWithExecutorAndLogger(executor, logger)
}

func newFibreChannelConnectorWithExecutorAndLogger(executor utils.Executor, logger logs.Logger) *fibreChannelConnector {
	initi := initiator.NewLinuxFibreChannelWithExecutor(executor)
	return &fibreChannelConnector{&scsiConnector{logger: logger, exec: executor, initi: initi}}
}

// ConnectVolume attach the volume to host by rescaning all the active FC HBAs.
func (c *fibreChannelConnector) ConnectVolume(volumeMountProperties *resources.VolumeMountProperties) error {
	hbas := c.initi.GetHBAs()

	if len(hbas) == 0 {
		c.logger.Warning("No FC HBA is found.")
		return nil
	}

	return c.initi.RescanHosts(hbas, volumeMountProperties)
}
