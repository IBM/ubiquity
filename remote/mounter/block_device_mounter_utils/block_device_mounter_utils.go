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
    "github.com/nightlyone/lockfile"
	"path/filepath"
	"os"
	"time"
)

type blockDeviceMounterUtils struct {
	logger            logs.Logger
	blockDeviceUtils  block_device_utils.BlockDeviceUtils
	rescanFlock       lockfile.Lockfile
	mpathFlock        lockfile.Lockfile
}

func NewBlockDeviceMounterUtilsWithBlockDeviceUtils(blockDeviceUtils block_device_utils.BlockDeviceUtils) BlockDeviceMounterUtils {
	return newBlockDeviceMounterUtils(blockDeviceUtils)
}

func NewBlockDeviceMounterUtils() BlockDeviceMounterUtils {
	return newBlockDeviceMounterUtils(block_device_utils.NewBlockDeviceUtils())
}

func newBlockDeviceMounterUtils(blockDeviceUtils block_device_utils.BlockDeviceUtils) BlockDeviceMounterUtils {
	rescanLock, err := lockfile.New(filepath.Join(os.TempDir(), "ubiquity.rescan.lock"))
	if err != nil {
		panic(err)
	}
	mpathLock, err := lockfile.New(filepath.Join(os.TempDir(), "ubiquity.mpath.lock"))
	if err != nil {
		panic(err)
	}

	return &blockDeviceMounterUtils{logger: logs.GetLogger(),
		blockDeviceUtils:  blockDeviceUtils,
		rescanFlock:       rescanLock,
		mpathFlock:        mpathLock,
	}
}

// MountDeviceFlow create filesystem on the device (if needed) and then mount it on a given mountpoint
func (b *blockDeviceMounterUtils) MountDeviceFlow(devicePath string, fsType string, mountPoint string) error {
	defer b.logger.Trace(logs.INFO, logs.Args{{"devicePath", devicePath}, {"fsType", fsType}, {"mountPoint", mountPoint}})()

	needToCreateFS, err := b.blockDeviceUtils.CheckFs(devicePath)
	if err != nil {
		return b.logger.ErrorRet(err, "CheckFs failed")
	}
	if needToCreateFS {
		if err = b.blockDeviceUtils.MakeFs(devicePath, fsType); err != nil {
			return b.logger.ErrorRet(err, "MakeFs failed")
		}
	}
	if err = b.blockDeviceUtils.MountFs(devicePath, mountPoint); err != nil {
		return b.logger.ErrorRet(err, "MountFs failed")
	}

	return nil
}

// UnmountDeviceFlow umount device, clean device and remove mountpoint folder
func (b *blockDeviceMounterUtils) UnmountDeviceFlow(devicePath string) error {
	defer b.logger.Trace(logs.INFO, logs.Args{{"devicePath", devicePath}})

	err := b.blockDeviceUtils.UmountFs(devicePath)
	if err != nil {
		return b.logger.ErrorRet(err, "UmountFs failed")
	}

	// locking for concurrent md delete operation
	b.logger.Debug("Ask for mpathLock for device", logs.Args{{"device", devicePath}})
	for {
		err := b.mpathFlock.TryLock()
		if err == nil {
			break
		}
		b.logger.Debug("mpathFlock.TryLock failed", logs.Args{{"error", err}})
		time.Sleep(time.Duration(500*time.Millisecond))
	}
	b.logger.Debug("Got mpathLock for device", logs.Args{{"device", devicePath}})
	defer b.mpathFlock.Unlock()
	defer b.logger.Debug("Released mpathLock for device", logs.Args{{"device", devicePath}})

	if err := b.blockDeviceUtils.Cleanup(devicePath); err != nil {
		return b.logger.ErrorRet(err, "Cleanup failed")
	}

	// TODO delete the directory here
	return nil
}

// RescanAll triggers the following OS rescanning :
// 1. iSCSI rescan (if protocol given is iscsi)
// 2. SCSI rescan
// 3. multipathing rescan
// return error if one of the steps fail
func (b *blockDeviceMounterUtils) RescanAll(withISCSI bool, wwn string, rescanForCleanUp bool) error {
	defer b.logger.Trace(logs.INFO, logs.Args{{"withISCSI", withISCSI}})

	// locking for concurrent rescans and reduce rescans if no need
	b.logger.Debug("Ask for rescanLock for volumeWWN", logs.Args{{"volumeWWN", wwn}})
	for {
		err := b.rescanFlock.TryLock()
		if err == nil {
			break
		}
		b.logger.Debug("rescanLock.TryLock failed", logs.Args{{"error", err}})
		time.Sleep(time.Duration(500*time.Millisecond))
	}
	b.logger.Debug("Got rescanLock for volumeWWN", logs.Args{{"volumeWWN", wwn}})
	defer b.rescanFlock.Unlock()
	defer b.logger.Debug("Released rescanLock for volumeWWN", logs.Args{{"volumeWWN", wwn}})

	if !rescanForCleanUp {
		// Only when run rescan for new device, try to check if its already exist to reduce rescans
		device, _ := b.Discover(wwn)

		if (device != "") {
			// if need rescan for discover new device but the new device is already exist then skip the rescan
			b.logger.Debug(
				"Skip rescan, because there is already multiple device for volumeWWN",
				logs.Args{{"volumeWWN", wwn}, {"multiple", device}})
			return nil
		}
	}
	// TODO : if rescanForCleanUp we need to check if block device is not longer exist and if so skip the rescan!

	// Do the rescans operations
	if withISCSI {
		if err := b.blockDeviceUtils.Rescan(block_device_utils.ISCSI); err != nil {
			return b.logger.ErrorRet(err, "Rescan failed", logs.Args{{"protocol", block_device_utils.ISCSI}})
		}
	}
	if err := b.blockDeviceUtils.Rescan(block_device_utils.SCSI); err != nil {
		return b.logger.ErrorRet(err, "Rescan failed", logs.Args{{"protocol", block_device_utils.SCSI}})
	}
	if !rescanForCleanUp {
		if err := b.blockDeviceUtils.ReloadMultipath(); err != nil {
			return b.logger.ErrorRet(err, "ReloadMultipath failed")
		}
	}
	return nil
}

func (b *blockDeviceMounterUtils) Discover(volumeWwn string) (string, error) {
	return b.blockDeviceUtils.Discover(volumeWwn)
}
