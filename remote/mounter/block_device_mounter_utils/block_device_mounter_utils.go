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

package block_device_mounter_utils

import (
	"github.com/IBM/ubiquity/remote/mounter/block_device_utils"
	"github.com/IBM/ubiquity/utils/logs"
	"sync"
)

type blockDeviceMounterUtils struct {
	logger            logs.Logger
	blockDeviceUtils  block_device_utils.BlockDeviceUtils
	rescanLock        *sync.RWMutex
	cleanMPDeviceLock *sync.RWMutex
}

// MountDeviceFlow create filesystem on the device (if needed) and then mount it on a given mountpoint
func (s *blockDeviceMounterUtils) MountDeviceFlow(devicePath string, fsType string, mountPoint string) error {
	defer s.logger.Trace(logs.INFO, logs.Args{{"devicePath", devicePath}, {"fsType", fsType}, {"mountPoint", mountPoint}})()

	needToCreateFS, err := s.blockDeviceUtils.CheckFs(devicePath)
	if err != nil {
		return s.logger.ErrorRet(err, "CheckFs failed")
	}
	if needToCreateFS {
		if err = s.blockDeviceUtils.MakeFs(devicePath, fsType); err != nil {
			return s.logger.ErrorRet(err, "MakeFs failed")
		}
	}
	if err = s.blockDeviceUtils.MountFs(devicePath, mountPoint); err != nil {
		return s.logger.ErrorRet(err, "MountFs failed")
	}

	return nil
}

// UnmountDeviceFlow umount device, clean device and remove mountpoint folder
func (s *blockDeviceMounterUtils) UnmountDeviceFlow(devicePath string) error {
	defer s.logger.Trace(logs.INFO, logs.Args{{"devicePath", devicePath}})

	err := s.blockDeviceUtils.UmountFs(devicePath)
	if err != nil {
		return s.logger.ErrorRet(err, "UmountFs failed")
	}

	// locking for concurrent md delete operation
	s.logger.Debug("Ask for cleanMPDeviceLock for device", logs.Args{{"device", devicePath}})
	s.cleanMPDeviceLock.Lock()
	s.logger.Debug("Recived cleanMPDeviceLock for device", logs.Args{{"device", devicePath}})
	defer s.cleanMPDeviceLock.Unlock()
	defer s.logger.Debug("Released cleanMPDeviceLock for device", logs.Args{{"device", devicePath}})

	if err := s.blockDeviceUtils.Cleanup(devicePath); err != nil {
		return s.logger.ErrorRet(err, "Cleanup failed")
	}

	// TODO delete the directory here
	return nil
}

// RescanAll triggers the following OS rescanning :
// 1. iSCSI rescan (if protocol given is iscsi)
// 2. SCSI rescan
// 3. multipathing rescan
// return error if one of the steps fail
func (s *blockDeviceMounterUtils) RescanAll(withISCSI bool, wwn string, rescanForCleanUp bool) error {
	defer s.logger.Trace(logs.INFO, logs.Args{{"withISCSI", withISCSI}})

	// locking for concurrent rescans and reduce rescans if no need
	s.logger.Debug("Ask for rescanLock for volumeWWN", logs.Args{{"volumeWWN", wwn}})
	s.rescanLock.Lock() // Prevent rescan in parallel
	s.logger.Debug("Recived rescanLock for volumeWWN", logs.Args{{"volumeWWN", wwn}})
	defer s.rescanLock.Unlock()
	defer s.logger.Debug("Released rescanLock for volumeWWN", logs.Args{{"volumeWWN", wwn}})

	device, _ := s.Discover(wwn)
	if !rescanForCleanUp && (device != "") {
		// if need rescan for discover new device but the new device is already exist then skip the rescan
		s.logger.Debug(
			"Skip rescan, because there is already multiple device for volumeWWN",
			logs.Args{{"volumeWWN", wwn}, {"multiple", device}})
		return nil
	}
	// TODO : if rescanForCleanUp we need to check if block device is not longer exist and if so skip the rescan!

	// Do the rescans operations
	if withISCSI {
		if err := s.blockDeviceUtils.Rescan(block_device_utils.ISCSI); err != nil {
			return s.logger.ErrorRet(err, "Rescan failed", logs.Args{{"protocol", block_device_utils.ISCSI}})
		}
	}
	if err := s.blockDeviceUtils.Rescan(block_device_utils.SCSI); err != nil {
		return s.logger.ErrorRet(err, "Rescan failed", logs.Args{{"protocol", block_device_utils.SCSI}})
	}
	if !rescanForCleanUp {
		if err := s.blockDeviceUtils.ReloadMultipath(); err != nil {
			return s.logger.ErrorRet(err, "ReloadMultipath failed")
		}
	}
	return nil
}

func (s *blockDeviceMounterUtils) Discover(volumeWwn string) (string, error) {
	return s.blockDeviceUtils.Discover(volumeWwn)
}
