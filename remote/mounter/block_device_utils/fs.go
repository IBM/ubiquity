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
	"github.com/IBM/ubiquity/utils/logs"
	"os/exec"
	"syscall"
	"bufio"
	"strings"
	"regexp"
	"fmt"
)

const (
	NotMountedErrorMessage = "not mounted" // Error while umount device that is already unmounted
)

func (b *blockDeviceUtils) CheckFs(mpath string) (bool, error) {
	defer b.logger.Trace(logs.DEBUG)()
	// TODO check first if mpath exist
	needFs := false
	blkidCmd := "blkid"
	if err := b.exec.IsExecutable(blkidCmd); err != nil {
		return false, b.logger.ErrorRet(&commandNotFoundError{blkidCmd, err}, "failed")
	}
	args := []string{mpath}
	outputBytes, err := b.exec.Execute(blkidCmd, args)
	if err != nil {
		if b.IsExitStatusCode(err, 2) {
			// TODO we can improve it by double check the fs type of this device and maybe log warning if its not the same fstype we expacted
			needFs = true
		} else {
			return false, b.logger.ErrorRet(&commandExecuteError{blkidCmd, err}, "failed")
		}
	}
	b.logger.Info("checked", logs.Args{{"needFs", needFs}, {"mpath", mpath}, {blkidCmd, outputBytes}})
	return needFs, nil
}

func (b *blockDeviceUtils) MakeFs(mpath string, fsType string) error {
	defer b.logger.Trace(logs.DEBUG)()
	mkfsCmd := "mkfs"
	if err := b.exec.IsExecutable(mkfsCmd); err != nil {
		return b.logger.ErrorRet(&commandNotFoundError{mkfsCmd, err}, "failed")
	}
	args := []string{"-t", fsType, mpath}
	if _, err := b.exec.Execute(mkfsCmd, args); err != nil {
		return b.logger.ErrorRet(&commandExecuteError{mkfsCmd, err}, "failed")
	}
	b.logger.Info("created", logs.Args{{"fsType", fsType}, {"mpath", mpath}})
	return nil
}

func (b *blockDeviceUtils) MountFs(mpath string, mpoint string) error {
	defer b.logger.Trace(logs.DEBUG)()
	mountCmd := "mount"
	if err := b.exec.IsExecutable(mountCmd); err != nil {
		return b.logger.ErrorRet(&commandNotFoundError{mountCmd, err}, "failed")
	}
	args := []string{mpath, mpoint}
	if _, err := b.exec.Execute(mountCmd, args); err != nil {
		return b.logger.ErrorRet(&commandExecuteError{mountCmd, err}, "failed")
	}
	b.logger.Info("mounted", logs.Args{{"mpoint", mpoint}})
	return nil
}

func (b *blockDeviceUtils) UmountFs(mpoint string) error {
	// Execute unmount operation (skip, if mpoint its already unmounted)

	defer b.logger.Trace(logs.DEBUG)()
	umountCmd := "umount"
	if err := b.exec.IsExecutable(umountCmd); err != nil {
		return b.logger.ErrorRet(&commandNotFoundError{umountCmd, err}, "failed")
	}

	args := []string{mpoint}
	if _, err := b.exec.Execute(umountCmd, args); err != nil {
		isMounted, _err := b.IsDeviceMounted(mpoint)
		if _err != nil {
			return _err
		}
		if ! isMounted{
			b.logger.Info("Device already unmounted.", logs.Args{{"mpoint", mpoint}})
			return nil
		}
		return b.logger.ErrorRet(&commandExecuteError{umountCmd, err}, "failed")
	}
	b.logger.Info("umounted", logs.Args{{"mpoint", mpoint}})
	return nil
}

func (b *blockDeviceUtils) IsDeviceMounted(devPath string) (bool, error) {
	/*
	   true, nil  : If device is mounted (check via pars the mount output)
	   false, nil : if device is not mounted
	   false, err : if failed to discover if device is mounted
	*/

	defer b.logger.Trace(logs.DEBUG)()
	mountCmd := "mount"
	if err := b.exec.IsExecutable(mountCmd); err != nil {
		return false, b.logger.ErrorRet(&commandNotFoundError{mountCmd, err}, "failed")
	}

	outputBytes, err := b.exec.Execute(mountCmd, nil)
	if err != nil {
		return false, b.logger.ErrorRet(&commandExecuteError{mountCmd, err}, "failed")
	}
	scanner := bufio.NewScanner(strings.NewReader(string(outputBytes[:])))
	pattern := fmt.Sprint("^" + devPath + `\s+`)
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return false, b.logger.ErrorRet(err, "failed")
	}
	for scanner.Scan() {
		line := scanner.Text()
		if regex.MatchString(line) {
			b.logger.Debug("Found mpath device as mounted device", logs.Args{{"mpath", devPath}, {"mountLine", line}})
			return true, nil
		}
	}
	b.logger.Debug("Not found mpath device as mounted device", logs.Args{{"mpath", devPath}})
	return false, nil
}


func (b *blockDeviceUtils) IsExitStatusCode(err error, code int) bool {
	defer b.logger.Trace(logs.DEBUG)()
	isExitStatusCode := false
	if status, ok := err.(*exec.ExitError); ok {
		if waitStatus, ok := status.ProcessState.Sys().(syscall.WaitStatus); ok {
			isExitStatusCode = waitStatus.ExitStatus() == code
		}
	}
	b.logger.Info("verified", logs.Args{{"isExitStatusCode", isExitStatusCode}, {"code", code}, {"error", err}})
	return isExitStatusCode
}
