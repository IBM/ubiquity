package initiator

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/IBM/ubiquity/utils"
	"github.com/IBM/ubiquity/utils/logs"
)

const multipathCmd = "multipath"
const FlushTimeout = 10 * 1000
const FlushRetries = 3

var SYS_BLOCK_PATH = "/sys/block"

type linuxSCSI struct {
	exec   utils.Executor
	logger logs.Logger
}

// FlushMultipath flushes the device, retry 3 times if it is failed.
func (ls *linuxSCSI) FlushMultipath(deviceMapName string) {
	if err := ls.exec.IsExecutable(multipathCmd); err != nil {
		return
	}

	for i := 0; i < FlushRetries; i++ {
		args := []string{"-f", deviceMapName}
		ls.logger.Info(fmt.Sprintf("Flush multipath by running: multipath -f %s", deviceMapName))
		_, err := ls.exec.ExecuteWithTimeout(FlushTimeout, multipathCmd, args)
		if err == nil {
			return
		}
	}
}

// RemoveSCSIDevice removes a scsi device based upon /dev/sdX name.
func (ls *linuxSCSI) RemoveSCSIDevice(device string) error {
	deviceName := strings.Replace(device, "/dev/", "", 1)
	path := SYS_BLOCK_PATH + fmt.Sprintf("/%s/device/delete", deviceName)
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		ls.logger.Debug(fmt.Sprintf("Remove SCSI device %s with %s", device, path))
		if err := ioutil.WriteFile(path, []byte("1"), 0666); err != nil {
			cmd := fmt.Sprintf(`echo "1" > %s`, path)
			return ls.logger.ErrorRet(&utils.CommandExecuteError{cmd, err}, "failed")
		}
	}
	return nil
}
