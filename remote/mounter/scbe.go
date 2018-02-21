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

package mounter

import (
	"fmt"
	"github.com/IBM/ubiquity/remote/mounter/block_device_mounter_utils"
	"github.com/IBM/ubiquity/resources"
	"github.com/IBM/ubiquity/utils"
	"github.com/IBM/ubiquity/utils/logs"
	"time"
)

const lockUmountWindowMessage = " To prevent concurrent rescans during umountFlow until the Detach finished(ActionAfterDetach)."

type scbeMounter struct {
	logger                  logs.Logger
	blockDeviceMounterUtils block_device_mounter_utils.BlockDeviceMounterUtils
	exec                    utils.Executor
	config                  resources.ScbeRemoteConfig
}

func NewScbeMounter(scbeRemoteConfig resources.ScbeRemoteConfig) resources.Mounter {
	blockDeviceMounterUtils := block_device_mounter_utils.NewBlockDeviceMounterUtils()
	return &scbeMounter{
		logger:                  logs.GetLogger(),
		blockDeviceMounterUtils: blockDeviceMounterUtils,
		exec: utils.NewExecutor(),
		config: scbeRemoteConfig,
	}
}

func (s *scbeMounter) Mount(mountRequest resources.MountRequest) (string, error) {
	defer s.logger.Trace(logs.DEBUG)()
	volumeWWN := mountRequest.VolumeConfig["Wwn"].(string)

	// Rescan OS
	if err := s.blockDeviceMounterUtils.RescanAll(!s.config.SkipRescanISCSI, volumeWWN, false); err != nil {
		return "", s.logger.ErrorRet(err, "RescanAll failed")
	}

	// Discover device
	devicePath, err := s.blockDeviceMounterUtils.Discover(volumeWWN, true)
	if err != nil {
		return "", s.logger.ErrorRet(err, "Discover failed", logs.Args{{"volumeWWN", volumeWWN}})
	}

	// Create mount point if needed   // TODO consider to move it inside the util
	if _, err := s.exec.Stat(mountRequest.Mountpoint); err != nil {
		s.logger.Info("Create mountpoint directory " + mountRequest.Mountpoint)
		if err := s.exec.MkdirAll(mountRequest.Mountpoint, 0700); err != nil {
			return "", s.logger.ErrorRet(err, "MkdirAll failed", logs.Args{{"mountpoint", mountRequest.Mountpoint}})
		}
	}

	// Mount device and mkfs if needed
	var fstype string
	fstypeInterface, ok := mountRequest.VolumeConfig[resources.OptionNameForVolumeFsType]
	if !ok {
		// the backend should do this default, but this is just for safe
		fstype = resources.DefaultForScbeConfigParamDefaultFilesystem
	} else {
		fstype = fstypeInterface.(string)
	}

	if err := s.blockDeviceMounterUtils.MountDeviceFlow(devicePath, fstype, mountRequest.Mountpoint); err != nil {
		return "", s.logger.ErrorRet(err, "MountDeviceFlow failed", logs.Args{{"devicePath", devicePath}})
	}

	return mountRequest.Mountpoint, nil
}

func (s *scbeMounter) Unmount(unmountRequest resources.UnmountRequest) error {
	defer s.logger.Trace(logs.DEBUG)()

	volumeWWN := unmountRequest.VolumeConfig["Wwn"].(string)
	mountpoint := fmt.Sprintf(resources.PathToMountUbiquityBlockDevices, volumeWWN)
	devicePath, err := s.blockDeviceMounterUtils.Discover(volumeWWN, true)
	if err != nil {
		return s.logger.ErrorRet(err, "Discover failed", logs.Args{{"volumeWWN", volumeWWN}})
	}

	s.rescanTryLock(volumeWWN)

	if err := s.blockDeviceMounterUtils.UnmountDeviceFlow(devicePath); err != nil {
		return s.logger.ErrorRet(err, "UnmountDeviceFlow failed", logs.Args{{"devicePath", devicePath}})
	}

	s.logger.Info("Delete mountpoint directory if exist", logs.Args{{"mountpoint", mountpoint}})
	// TODO move this part to the util
	if _, err := s.exec.Stat(mountpoint); err == nil {
		// TODO consider to add the prefix of the wwn in the OS (multipath -ll output)
		if err := s.exec.RemoveAll(mountpoint); err != nil {
			return s.logger.ErrorRet(err, "RemoveAll failed", logs.Args{{"mountpoint", mountpoint}})
		}
	}

	return nil

}

func (s *scbeMounter) ActionAfterDetach(request resources.AfterDetachRequest) error {
	defer s.logger.Trace(logs.DEBUG)()
	volumeWWN := request.VolumeConfig["Wwn"].(string)

	s.rescanUnlock(volumeWWN)
	// Rescan OS
	if err := s.blockDeviceMounterUtils.RescanAll(!s.config.SkipRescanISCSI, volumeWWN, true); err != nil {
		return s.logger.ErrorRet(err, "RescanAll failed")
	}
	return nil
}

func (s *scbeMounter) rescanTryLock(volumeWWN string) {
	/* To prevent concurrent rescans during blockDeviceMounterUtils.UnmountDeviceFlow and just before ActionAfterDetach start.
	   During this window of time, there is no mpath device but the volume still mapped to the host.
	   Therefor a parallel rescan or multipath reload could return back the delete mpath device and lead to faulty multipath device.

	   TODO: We should refactor the resources.Mounter and add new interface that will use lock in the same context.
	         Currently the lock and unlock is not in the same context, the lock done in Unmount() interface and unlock from ActionAfterDetach interface.
	   */
	s.logger.Debug("Ask for rescanLock for volumeWWN." + lockUmountWindowMessage, logs.Args{{"volumeWWN", volumeWWN}})
	rescanFlock := s.blockDeviceMounterUtils.RescanFlock()
	for {
		err := rescanFlock.TryLock()
		if err == nil {
			s.logger.Debug("Got rescanLock for volumeWWN." + lockUmountWindowMessage, logs.Args{{"volumeWWN", volumeWWN}})
			break
		}
		s.logger.Debug("rescanLock.TryLock failed for volumeWWN." + lockUmountWindowMessage, logs.Args{{"volumeWWN", volumeWWN}, {"error", err}})
		time.Sleep(time.Duration(500*time.Millisecond))
	}
}

func (s *scbeMounter) rescanUnlock(volumeWWN string) {
	// Just release the rescan lock. More detail in rescanTryLock function.
	s.blockDeviceMounterUtils.RescanFlock().Unlock()
	s.logger.Debug("Released rescanLock for volumeWWN." + lockUmountWindowMessage, logs.Args{{"volumeWWN", volumeWWN}})
}
