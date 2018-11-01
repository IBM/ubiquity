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

	"io/ioutil"
)

const rescanIscsiTimeout = 1 * 60 * 1000
const rescanScsiTimeout = 2 * 60 * 1000
const fcHostDirectory = "/sys/class/fc_host/"

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
		return b.logger.ErrorRet(&commandNotFoundError{rescanCmd, err}, "failed")
	}
	args := []string{"-m", "session", "--rescan"}
	if _, err := b.exec.ExecuteWithTimeout(rescanIscsiTimeout, rescanCmd, args); err != nil {
		return b.logger.ErrorRet(&CommandExecuteError{rescanCmd, err}, "failed")
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
	hostInfos, err := ioutil.ReadDir(fcHostDirectory)
	if err != nil {
		return b.logger.ErrorRet(err, "Getting fc_host failed.", logs.Args{{"fcHostDirectory", fcHostDirectory}})
	}
	if len(hostInfos) == 0 {
		return b.logger.ErrorRet(err, "There is no fc_host found.", logs.Args{{"fcHostDirectory", fcHostDirectory}})
	}

	for _, host := range hostInfos {
		rescanCmd := "echo"
		if err := b.exec.IsExecutable(rescanCmd); err != nil {
			return b.logger.ErrorRet(&commandNotFoundError{rescanCmd, err}, "failed")
		}
		rescanPara := []string{`"1"`, ">/sys/class/fc_host/" + host.Name() + "/issue_lip"}
		if _, err := b.exec.ExecuteWithTimeout(rescanIscsiTimeout, rescanCmd, rescanPara); err != nil {
			continue
		}
		rescanArgs := []string{`"---"`, ">/sys/class/scsi_host/" + host.Name() + "/scan"}
		if _, err := b.exec.ExecuteWithTimeout(rescanIscsiTimeout, rescanCmd, rescanArgs); err != nil {
			continue
		}
	}
	return nil
}
