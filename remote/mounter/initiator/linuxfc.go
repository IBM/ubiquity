package initiator

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/IBM/ubiquity/resources"
	"github.com/IBM/ubiquity/utils"
	"github.com/IBM/ubiquity/utils/logs"
)

const SYSTOOL = "systool"
const SYSTOOL_TIMEOUT = 5 * 1000

var FC_HOST_SYSFS_PATH = "/sys/class/fc_host"
var SCSI_HOST_SYSFS_PATH = "/sys/class/scsi_host"

type linuxFibreChannel struct {
	*linuxSCSI
}

func NewLinuxFibreChannel() Initiator {
	return newLinuxFibreChannel()
}

func NewLinuxFibreChannelWithExecutor(executor utils.Executor) Initiator {
	return newLinuxFibreChannelWithExecutor(executor)
}

func newLinuxFibreChannel() *linuxFibreChannel {
	executor := utils.NewExecutor()
	return newLinuxFibreChannelWithExecutor(executor)
}

func newLinuxFibreChannelWithExecutor(executor utils.Executor) *linuxFibreChannel {
	logger := logs.GetLogger()
	return &linuxFibreChannel{&linuxSCSI{logger: logger, exec: executor}}
}

func (lfc *linuxFibreChannel) hasFCSupport() bool {
	fileInfo, err := os.Stat(FC_HOST_SYSFS_PATH)
	if err != nil {
		return false
	}
	return fileInfo.IsDir()
}

// Try to get the HBA channel and SCSI target for an HBA.
// return a string with "channel, target, lun" entries, the channel and target
// may be '-' wildcards if unable to determine them.
func (lfc *linuxFibreChannel) getHBAChannelScsiTarget(volumeMountProperties *resources.VolumeMountProperties) string {
	//TODO: get channel and target
	if volumeMountProperties.LunNumber == float64(-1) {
		return "- - -"
	}
	// use %g to print a float64 to int
	return fmt.Sprintf("- - %g", volumeMountProperties.LunNumber)
}

// GetHBAs return all the FC HBAs in the system
func (lfc *linuxFibreChannel) GetHBAs() []string {
	if !lfc.hasFCSupport() {
		return []string{}
	}

	if err := lfc.exec.IsExecutable(SYSTOOL); err != nil {
		lfc.logger.Warning(fmt.Sprintf("No systool installed, get from path %s instead.", FC_HOST_SYSFS_PATH))
		return lfc.getFcHBAsByPath()
	}

	out, err := lfc.exec.ExecuteWithTimeout(SYSTOOL_TIMEOUT, SYSTOOL, []string{"-c", "fc_host", "-v"})
	if err != nil {
		lfc.logger.Warning(fmt.Sprintf("Executing systool failed with error: %v. Get from path %s instead.", err, FC_HOST_SYSFS_PATH))
		return lfc.getFcHBAsByPath()
	}

	// No FC HBAs were found
	if len(out) == 0 {
		return []string{}
	}
	hbas := generateHBAsInfoFromSystoolOutput(out)
	lfc.logger.Debug(fmt.Sprintf("Find %d HBAs from systool output, getting the name of the online one(s)", len(hbas)))
	hbaNames := []string{}
	for _, hba := range hbas {
		if hba["port_state"] == "Online" {
			hbaNames = append(hbaNames, hba["ClassDevice"])
		}
	}
	return hbaNames
}

// getFcHBAsByPath returns the FC HBA names under path /sys/class/fc_host
func (lfc *linuxFibreChannel) getFcHBAsByPath() []string {
	hbas := []string{}
	hostInfos, err := ioutil.ReadDir(FC_HOST_SYSFS_PATH)
	if err != nil {
		return []string{}
	}

	for _, host := range hostInfos {
		hbas = append(hbas, host.Name())
	}
	return hbas
}

// RescanHosts rescan all the host HBAs for a certain lun if LunNumber is specified,
// if not, means LunNumber is -1, rescan all the luns.
func (lfc *linuxFibreChannel) RescanHosts(hbas []string, volumeMountProperties *resources.VolumeMountProperties) error {
	defer lfc.logger.Trace(logs.DEBUG)()

	ctl := lfc.getHBAChannelScsiTarget(volumeMountProperties)

	for _, hba := range hbas {
		hbaPath := SCSI_HOST_SYSFS_PATH + "/" + hba + "/scan"
		lfc.logger.Debug(fmt.Sprintf(`Scanning HBA with command: echo "%s" > %s`, ctl, hbaPath))
		if err := ioutil.WriteFile(hbaPath, []byte(ctl), 0666); err != nil {
			lfc.logger.Debug("Failed to scan HBA", logs.Args{{"name", hba}, {"err", err}})
			continue
		}
	}
	return nil
}

/*
generateHBAsInfoFromSystoolOutput analysises the output of command "systool -c fc_host -v",
and generates a list of HBA info.

A sample output is:
Class = "fc_host"

  Class Device = "host0"
  Class Device path = "/sys/devices/css0/0.0.0000/0.0.a200/host0/fc_host/host0"
    active_fc4s         = "0x00 0x00 0x01 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 "
    dev_loss_tmo        = "60"
    maxframe_size       = "2112 bytes"
    node_name           = "0x5005076400c1b2b8"
    permanent_port_name = "0xc05076e5118011d1"
    port_id             = "0x0ecfd3"
    port_name           = "0xc05076e511803cb0"
    port_state          = "Online"
    port_type           = "NPIV VPORT"
    serial_number       = "IBM0200000001B2B8"
    speed               = "8 Gbit"
    supported_classes   = "Class 2, Class 3"
    supported_fc4s      = "0x00 0x00 0x01 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 "
    supported_speeds    = "4 Gbit, 8 Gbit, 16 Gbit"
    symbolic_name       = "IBM     3906            0200000001B2B8  PCHID: 011D NPIV UlpId: 004C0308   DEVNO: 0.0.a200 NAME: stk8s008"
    tgtid_bind_type     = "wwpn (World Wide Port Name)"
    uevent              =

    Device = "host0"
    Device path = "/sys/devices/css0/0.0.0000/0.0.a200/host0"
      uevent              = "DEVTYPE=scsi_host"


  Class Device = "host1"
  Class Device path = "/sys/devices/css0/0.0.0001/0.0.a300/host1/fc_host/host1"
    active_fc4s         = "0x00 0x00 0x01 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 "
    dev_loss_tmo        = "60"
    maxframe_size       = "2112 bytes"
    node_name           = "0x5005076400c1b2b8"
    permanent_port_name = "0xc05076e511801811"
    port_id             = "0x0fc613"
    port_name           = "0xc05076e511803b48"
    port_state          = "Online"
    port_type           = "NPIV VPORT"
    serial_number       = "IBM0200000001B2B8"
    speed               = "8 Gbit"
    supported_classes   = "Class 2, Class 3"
    supported_fc4s      = "0x00 0x00 0x01 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 "
    supported_speeds    = "4 Gbit, 8 Gbit, 16 Gbit"
    symbolic_name       = "IBM     3906            0200000001B2B8  PCHID: 0181 NPIV UlpId: 00450308   DEVNO: 0.0.a300 NAME: stk8s008"
    tgtid_bind_type     = "wwpn (World Wide Port Name)"
    uevent              =

    Device = "host1"
    Device path = "/sys/devices/css0/0.0.0001/0.0.a300/host1"
      uevent              = "DEVTYPE=scsi_host"


*/
func generateHBAsInfoFromSystoolOutput(out []byte) []map[string]string {
	lines := strings.Split(string(out), "\n")
	// ignore the first 2 lines
	lines = lines[2:]
	hbas := []map[string]string{}
	hba := map[string]string{}
	lastline := ""
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// 2 newlines denotes a new hba port
		if line == "" && lastline == "" {
			if len(hba) > 0 {
				hbas = append(hbas, hba)
				hba = map[string]string{}
			}
		} else {
			val := strings.Split(line, "=")
			if len(val) == 2 {
				key := strings.Replace(strings.TrimSpace(val[0]), " ", "", -1)
				value := strings.Replace(strings.TrimSpace(val[1]), `"`, "", -1)
				hba[key] = value
			}
		}
		lastline = line
	}
	return hbas
}
