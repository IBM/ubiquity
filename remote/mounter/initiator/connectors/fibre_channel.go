package connectors

import (
	"fmt"

	"github.com/IBM/ubiquity/remote/mounter/initiator"
	"github.com/IBM/ubiquity/resources"
	"github.com/IBM/ubiquity/utils"
	"github.com/IBM/ubiquity/utils/logs"
)

type fibreChannelConnector struct {
	exec    utils.Executor
	logger  logs.Logger
	linuxfc initiator.Initiator
}

func NewFibreChannelConnector() initiator.Connector {
	return newFibreChannelConnector()
}

func NewFibreChannelConnectorWithExecutor(executor utils.Executor) initiator.Connector {
	return newFibreChannelConnectorWithExecutorAndLogger(executor)
}

func NewFibreChannelConnectorWithAllFields(executor utils.Executor, linuxfc initiator.Initiator) initiator.Connector {
	logger := logs.GetLogger()
	return &fibreChannelConnector{logger: logger, exec: executor, linuxfc: linuxfc}
}

func newFibreChannelConnector() *fibreChannelConnector {
	executor := utils.NewExecutor()
	return newFibreChannelConnectorWithExecutorAndLogger(executor)
}

func newFibreChannelConnectorWithExecutorAndLogger(executor utils.Executor) *fibreChannelConnector {
	logger := logs.GetLogger()
	linuxfc := initiator.NewLinuxFibreChannelWithExecutor(executor)

	return &fibreChannelConnector{logger: logger, exec: executor, linuxfc: linuxfc}
}

// ConnectVolume attach the volume to host by rescaning all the active FC HBAs.
func (c *fibreChannelConnector) ConnectVolume(volumeMountProperties *resources.VolumeMountProperties) error {
	hbas := c.linuxfc.GetHBAs()

	if len(hbas) == 0 {
		c.logger.Warning("No FC HBA is found.")
		return nil
	}

	return c.linuxfc.RescanHosts(hbas, volumeMountProperties)
}

// DisconnectVolume removes a volume from host by echo "1" to all scsi device's /delete
func (c *fibreChannelConnector) DisconnectVolume(volumeMountProperties *resources.VolumeMountProperties) error {
	devices := []string{}
	_, _, devNames, err := utils.GetMultipathOutputAndDeviceMapperAndDevice(volumeMountProperties.WWN, c.exec)
	if err != nil {
		return c.logger.ErrorRet(err, "Failed to get multipath output before disconnecting volume")
	}
	for _, devName := range devNames {
		device := fmt.Sprintf("/dev/%s", devName)
		devices = append(devices, device)
	}

	c.logger.Debug("Remove devices", logs.Args{{"names", devices}})
	err = c.removeDevices(devices)
	if err != nil {
		return c.logger.ErrorRet(err, "Failed to remove devices")
	}

	if _, devMapper, _, _ := utils.GetMultipathOutputAndDeviceMapperAndDevice(volumeMountProperties.WWN, c.exec); devMapper != "" {
		// flush multipath if device still exists after disconnection
		c.linuxfc.FlushMultipath(devMapper)
	}
	return nil
}

func (c *fibreChannelConnector) removeDevices(devices []string) error {
	// Do we need to flush io?
	var err error
	for _, device := range devices {
		err = c.linuxfc.RemoveSCSIDevice(device)
	}
	return err
}
