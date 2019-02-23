/**
 * Copyright 2017 IBM Corp.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package block_device_utils

import (
	"fmt"
	"time"

	"github.com/IBM/ubiquity/resources"
	"github.com/IBM/ubiquity/utils/logs"
)

const rescanIscsiTimeout = 1 * 60 * 1000
const rescanScsiTimeout = 2 * 60 * 1000

var FcHostDir = "/sys/class/fc_host/"
var ScsiHostDir = "/sys/class/scsi_host/"

func (b *blockDeviceUtils) Rescan(protocol Protocol, volumeMountProperties *resources.VolumeMountProperties) error {
	defer b.logger.Trace(logs.DEBUG)()

	switch protocol {
	case SCSI:
		return b.RescanSCSI(volumeMountProperties)
	case ISCSI:
		return b.RescanISCSI()
	default:
		return b.logger.ErrorRet(&unsupportedProtocolError{protocol}, "failed")
	}
}

func (b *blockDeviceUtils) Disconnect(protocol Protocol, volumeMountProperties *resources.VolumeMountProperties) error {
	defer b.logger.Trace(logs.DEBUG)()

	switch protocol {
	case SCSI:
		return b.DisconnectSCSI(volumeMountProperties)
	case ISCSI:
		return b.DisconnectISCSI()
	default:
		return b.logger.ErrorRet(&unsupportedProtocolError{protocol}, "failed")
	}
}

func (b *blockDeviceUtils) RescanISCSI() error {
	defer b.logger.Trace(logs.DEBUG)()
	rescanCmd := "iscsiadm"
	if err := b.exec.IsExecutable(rescanCmd); err != nil {
		b.logger.Debug("No iscisadm installed skipping ISCSI rescan")
		return nil
	}

	args := []string{"-m", "session", "--rescan"}

	if _, err := b.exec.ExecuteWithTimeout(rescanIscsiTimeout, rescanCmd, args); err != nil {
		if b.IsExitStatusCode(err, 21) {
			// error code 21 : ISCSI_ERR_NO_OBJS_FOUND - no records/targets/sessions/portals found to execute operation on.
			b.logger.Warning(NoIscsiadmCommnadWarningMessage, logs.Args{{"command", fmt.Sprintf("[%s %s]", rescanCmd, args)}})
			return nil

		} else {
			return b.logger.ErrorRet(&CommandExecuteError{rescanCmd, err}, "failed")
		}
	}
	return nil
}

func (b *blockDeviceUtils) RescanSCSI(volumeMountProperties *resources.VolumeMountProperties) error {
	var err error
	for i := 0; i < 6; i++ {
		if err = b.fcConnector.ConnectVolume(volumeMountProperties); err != nil {
			return b.logger.ErrorRet(err, "RescanSCSI failed", logs.Args{{"volumeWWN", volumeMountProperties.WWN}})
		}
		if _, _, err = b.getMultipathOutputAndDeviceUid(volumeMountProperties.WWN); err == nil {
			return nil
		}
		b.logger.Warning("Can't find the new volume in multipath output after rescan, sleep one second and try again.")
		time.Sleep(1 * time.Second)
	}
	return b.logger.ErrorRet(err, "RescanSCSI failed", logs.Args{{"volumeWWN", volumeMountProperties.WWN}})
}

// TODO: improve it to make it faster
func (b *blockDeviceUtils) DisconnectISCSI() error {
	return b.RescanISCSI()
}

func (b *blockDeviceUtils) DisconnectSCSI(volumeMountProperties *resources.VolumeMountProperties) error {
	return b.fcConnector.DisconnectVolume(volumeMountProperties)
}
