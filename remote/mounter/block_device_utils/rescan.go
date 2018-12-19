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
	"errors"

	"fmt"
	"io/ioutil"

	"github.com/IBM/ubiquity/utils/logs"
)

const rescanIscsiTimeout = 1 * 60 * 1000
const rescanScsiTimeout = 2 * 60 * 1000

var FcHostDir = "/sys/class/fc_host/"
var ScsiHostDir = "/sys/class/scsi_host/"

func (b *blockDeviceUtils) Rescan(protocol Protocol) error {
	defer b.logger.Trace(logs.DEBUG)()

	switch protocol {
	case SCSI:
		return b.RescanSCSI()
	case ISCSI:
		return b.RescanISCSI()
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

func (b *blockDeviceUtils) RescanSCSI() error {
	defer b.logger.Trace(logs.DEBUG)()
	commands := []string{"rescan-scsi-bus", "rescan-scsi-bus.sh"}
	rescanCmd := ""
	for _, cmd := range commands {
		if err := b.exec.IsExecutable(cmd); err == nil {
			rescanCmd = cmd
			break
		}
	}
	if rescanCmd == "" {
		return b.logger.ErrorRet(&commandNotFoundError{commands[0], errors.New("")}, "failed")
	}
	args := []string{"-r"} // TODO should use -r only in clean up
	if _, err := b.exec.ExecuteWithTimeout(rescanScsiTimeout, rescanCmd, args); err != nil {
		return b.logger.ErrorRet(&CommandExecuteError{rescanCmd, err}, "failed")
	}
	return nil
}

func (b *blockDeviceUtils) RescanSCSILun0() error {
	defer b.logger.Trace(logs.DEBUG)()
	hostInfos, err := ioutil.ReadDir(FcHostDir)
	if err != nil {
		return b.logger.ErrorRet(err, "Getting fc_host failed.", logs.Args{{"FcHostDir", FcHostDir}})
	}
	if len(hostInfos) == 0 {
		err := fmt.Errorf("There is no fc_host found, please check the fc host.")
		return b.logger.ErrorRet(err, "There is no fc_host found.", logs.Args{{"FcHostDir", FcHostDir}})
	}

	for _, host := range hostInfos {
		b.logger.Debug("scan the host", logs.Args{{"name: ", host.Name()}})
		fcHostFile := FcHostDir + host.Name() + "/issue_lip"
		if err := ioutil.WriteFile(fcHostFile, []byte("1"), 0200); err != nil {
			b.logger.Debug("Write issue_lip failed", logs.Args{{"err", err}})
		}
		filename := ScsiHostDir + host.Name() + "/scan"
		b.logger.Debug("ScsiHostDir", logs.Args{{"ScsiHostDir", ScsiHostDir}})
		if err := ioutil.WriteFile(filename, []byte("- - -"), 0200); err != nil {
			b.logger.Debug("Write file scan failed", logs.Args{{"err", err}})
			continue
		}
	}
	return nil
}
