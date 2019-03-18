package connectors

import (
	"fmt"

	"github.com/IBM/ubiquity/remote/mounter/initiator"
	"github.com/IBM/ubiquity/resources"
	"github.com/IBM/ubiquity/utils"
	"github.com/IBM/ubiquity/utils/logs"
)

type scsiConnector struct {
	exec   utils.Executor
	logger logs.Logger
	initi  initiator.Initiator
}

// DisconnectVolume will do following things:
// 1. flush multipath device: multipath -f /dev/mapper/mpathx
// 2. flush device io for all devices: blockdev --flushbufs /dev/sdx (not implemented yet)
// 3. remove all devices by path from host: echo 1 > /sys/block/sdx/device/delete
func (c *scsiConnector) DisconnectVolume(volumeMountProperties *resources.VolumeMountProperties) error {
	var err error

	devices := []string{}
	devMapper := volumeMountProperties.DeviceMapper
	devNames := volumeMountProperties.Devices
	if devMapper == "" {
		_, devMapper, devNames, err = utils.GetMultipathOutputAndDeviceMapperAndDevice(volumeMountProperties.WWN, c.exec)
		if err != nil || devMapper == "" {
			return c.logger.ErrorRet(err, "Failed to get multipath output before disconnecting volume")
		}
	}

	// flush multipath device
	c.logger.Info("Flush multipath device", logs.Args{{"name", devMapper}})
	c.initi.FlushMultipath(devMapper)

	for _, devName := range devNames {
		device := fmt.Sprintf("/dev/%s", devName)
		devices = append(devices, device)
	}

	c.logger.Info("Remove devices", logs.Args{{"names", devices}})
	err = c.removeDevices(devices)
	if err != nil {
		return c.logger.ErrorRet(err, "Failed to remove devices")
	}

	// If flushing the multipath failed before, try now after we have removed the devices.
	c.logger.Info("Flush multipath device again after removing the devices", logs.Args{{"name", devMapper}})
	c.initi.FlushMultipath(devMapper)
	return nil
}

func (c *scsiConnector) removeDevices(devices []string) error {
	var err error
	for _, device := range devices {
		err = c.initi.RemoveSCSIDevice(device)
	}
	return err
}
