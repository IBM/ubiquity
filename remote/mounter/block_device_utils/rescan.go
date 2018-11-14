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

	"github.com/IBM/ubiquity/utils/logs"

)

const rescanIscsiTimeout = 1 * 60 * 1000
const rescanScsiTimeout = 2 * 60 * 1000

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
		if b.IsExitStatusCode(err, 21){ 
			// error code 21 : ISCSI_ERR_NO_OBJS_FOUND - no records/targets/sessions/portals found to execute operation on. 
			b.logger.Warning("No active iscsi session exists. if iscsi connection was expected please check the connectivity of this node.")
			return nil
			
		}  else{
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
