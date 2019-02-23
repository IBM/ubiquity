package initiator

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/IBM/ubiquity/utils"
	"github.com/IBM/ubiquity/utils/logs"
)

type linuxSCSI struct {
	exec   utils.Executor
	logger logs.Logger
}

// RemoveSCSIDevice removes a scsi device based upon /dev/sdX name.
func (ls *linuxSCSI) RemoveSCSIDevice(device string) error {
	deviceName := strings.Replace(device, "/dev/", "", 1)
	path := fmt.Sprintf("/sys/block/%s/device/delete", deviceName)
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		ls.logger.Debug(fmt.Sprintf("Remove SCSI device %s with %s", device, path))
		if err := ioutil.WriteFile(path, []byte("1"), 0666); err != nil {
			cmd := fmt.Sprintf(`echo "1" > %s`, path)
			return ls.logger.ErrorRet(&utils.CommandExecuteError{cmd, err}, "failed")
		}
	}
	return nil
}
