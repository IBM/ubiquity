package connectors

import (
	"fmt"
	"strings"

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

func NewFibreChannelConnectorWithExecutorAndLogger(executor utils.Executor, logger logs.Logger) initiator.Connector {
	return newFibreChannelConnectorWithExecutorAndLogger(executor, logger)
}

func newFibreChannelConnector() *fibreChannelConnector {
	logger := logs.GetLogger()
	executor := utils.NewExecutor()
	return newFibreChannelConnectorWithExecutorAndLogger(executor, logger)
}

func newFibreChannelConnectorWithExecutorAndLogger(executor utils.Executor, logger logs.Logger) *fibreChannelConnector {
	linuxfc := initiator.NewLinuxFibreChannelWithExecutorAndLogger(executor, logger)

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
	paths := c.findPathsFromMultipathOutpot(volumeMountProperties)
	for _, path := range paths {
		device := fmt.Sprintf("/dev/%s", path)
		devices = append(devices, device)
	}

	c.logger.Debug("Remove devices", logs.Args{{"names", devices}})
	return c.removeDevices(devices)
}

func (c *fibreChannelConnector) removeDevices(devices []string) error {
	// Do we need to flush io?
	var err error
	for _, device := range devices {
		err = c.linuxfc.RemoveSCSIDevice(device)
	}
	return err
}

// TODO: it is not a good idea to find device paths in this way, try to improve it.
func (c *fibreChannelConnector) findPathsFromMultipathOutpot(volumeMountProperties *resources.VolumeMountProperties) []string {
	multipath := "multipath"
	if err := c.exec.IsExecutable(multipath); err != nil {
		c.logger.Warning("No multipath installed.")
		return []string{}
	}

	lunNumber := volumeMountProperties.LunNumber
	out, err := c.exec.Execute(multipath, []string{"-ll", "|", fmt.Sprintf("grep %d", lunNumber)})
	if err != nil {
		c.logger.Warning(fmt.Sprintf("Executing multipath failed with error: %v", err))
		return []string{}
	}
	return generatePathsFromMultipathOutput(out)
}

/*
generatePathsFromMultipathOutput analysises the output of command "multipath -ll |grep lunNumber",
and generates a list of path.

A sample output is:
  |- 0:0:4:255 sda 8:0   active ready running
  |- 0:0:5:255 sdb 8:16  active ready running
  |- 0:0:6:255 sdc 8:32  active ready running
  |- 0:0:7:255 sdd 8:48  active ready running
  |- 1:0:4:255 sde 8:64  active ready running
  |- 1:0:5:255 sdf 8:80  active ready running
  |- 1:0:6:255 sdg 8:96  active ready running
  `- 1:0:7:255 sdh 8:112 active ready running
*/
func generatePathsFromMultipathOutput(out []byte) []string {
	lines := strings.Split(string(out), "\n")
	paths := []string{}
	for _, line := range lines {
		line = strings.TrimSpace(line)
		path := strings.Split(line, " ")[2]
		if strings.HasPrefix(path, "sd") {
			paths = append(paths)
		}
	}
	return paths
}
