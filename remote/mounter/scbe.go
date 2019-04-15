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
	"os"

	"github.com/IBM/ubiquity/remote/mounter/block_device_mounter_utils"
	"github.com/IBM/ubiquity/remote/mounter/block_device_utils"
	"github.com/IBM/ubiquity/resources"
	"github.com/IBM/ubiquity/utils"
	"github.com/IBM/ubiquity/utils/logs"
)

type scbeMounter struct {
	logger                  logs.Logger
	blockDeviceMounterUtils block_device_mounter_utils.BlockDeviceMounterUtils
	exec                    utils.Executor
}

func newScbMounter(blockDeviceMounterUtils block_device_mounter_utils.BlockDeviceMounterUtils, executer utils.Executor) resources.Mounter {
	return &scbeMounter{
		logger:                  logs.GetLogger(),
		blockDeviceMounterUtils: blockDeviceMounterUtils,
		exec:                    executer,
	}
}

func NewScbeMounter() resources.Mounter {
	blockDeviceMounterUtils := block_device_mounter_utils.NewBlockDeviceMounterUtils()
	return newScbMounter(blockDeviceMounterUtils, utils.NewExecutor())
}

func NewScbeMounterWithExecuter(blockDeviceMounterUtils block_device_mounter_utils.BlockDeviceMounterUtils, executer utils.Executor) resources.Mounter {
	return newScbMounter(blockDeviceMounterUtils, executer)
}

func (s *scbeMounter) prepareVolumeMountProperties(vcGetter resources.VolumeConfigGetter) *resources.VolumeMountProperties {
	volumeConfig := vcGetter.GetVolumeConfig()
	volumeWWN := volumeConfig["Wwn"].(string)
	volumeLunNumber := -1
	if volumeLunNumberInterface, exists := volumeConfig[resources.ScbeKeyVolAttachLunNumToHost]; exists {

		// LunNumber is int, but after json.Marshal and json.UNmarshal it will become float64.
		// see https://stackoverflow.com/questions/39152481/unmarshaling-a-json-integer-to-an-empty-interface-results-in-wrong-type-assertio
		// but LunNumber should be int, so convert it here.
		volumeLunNumber = int(volumeLunNumberInterface.(float64))
	}
	return &resources.VolumeMountProperties{WWN: volumeWWN, LunNumber: volumeLunNumber}
}

func (s *scbeMounter) Mount(mountRequest resources.MountRequest) (string, error) {
	defer s.logger.Trace(logs.DEBUG)()
	volumeMountProperties := s.prepareVolumeMountProperties(&mountRequest)

	// Rescan OS
	if err := s.blockDeviceMounterUtils.RescanAll(volumeMountProperties); err != nil {
		return "", s.logger.ErrorRet(err, "RescanAll failed")
	}

	// Discover device
	devicePath, err := s.blockDeviceMounterUtils.Discover(volumeMountProperties.WWN, true)
	if err != nil {
		return "", s.logger.ErrorRet(err, "Discover failed", logs.Args{{"volumeWWN", volumeMountProperties.WWN}})
	}

	// Create mount point if needed   // TODO consider to move it inside the util
	if _, err := s.exec.Stat(mountRequest.Mountpoint); err != nil {
		s.logger.Info("Create mountpoint directory " + mountRequest.Mountpoint)
		if err := s.exec.MkdirAll(mountRequest.Mountpoint, 0700); err != nil {
			return "", s.logger.ErrorRet(err, "MkdirAll failed", logs.Args{{"mountpoint", mountRequest.Mountpoint}})
		}
	} else {
		s.logger.Warning("Idempotent issue : mount point directory already exists", logs.Args{{"mountpoint", mountRequest.Mountpoint}})
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
	mountpoint := fmt.Sprintf(resources.PathToMountUbiquityBlockDevices, volumeWWN) // TODO instead of build the mountpoint it should come from unmountRequest.
	skipUnmountFlow := false
	devicePath, err := s.blockDeviceMounterUtils.Discover(volumeWWN, true)
	if err != nil {
		switch err.(type) {
		case *block_device_utils.VolumeNotFoundError:
			s.logger.Warning("Idempotent issue encountered: volume not found. skipping UnmountDeviceFlow ", logs.Args{{"volumeWWN", volumeWWN}})
			skipUnmountFlow = true
		case *block_device_utils.FaultyDeviceError:
			s.logger.Warning("Idempotent issue encountered: mpath device is faulty. continuing with UnmountDeviceFlow", logs.Args{{"device", devicePath}, {"volumeWWN", volumeWWN}})
		default:
			return s.logger.ErrorRet(err, "Discover failed", logs.Args{{"volumeWWN", volumeWWN}})
		}
	}

	if !skipUnmountFlow {
		if err := s.blockDeviceMounterUtils.UnmountDeviceFlow(devicePath, volumeWWN); err != nil {
			return s.logger.ErrorRet(err, "UnmountDeviceFlow failed", logs.Args{{"devicePath", devicePath}})
		}
	}

	s.logger.Info("Delete mountpoint directory if exist", logs.Args{{"mountpoint", mountpoint}})
	// TODO move this part to the util
	if _, err := s.exec.Stat(mountpoint); err == nil {
		s.logger.Debug("Checking if mountpoint is empty and can be deleted", logs.Args{{"mountpoint", mountpoint}})
		emptyDir, err := s.exec.IsDirEmpty(mountpoint)
		if err != nil {
			return s.logger.ErrorRet(err, "Getting number of files failed.", logs.Args{{"mountpoint", mountpoint}})
		}
		if emptyDir != true {
			return s.logger.ErrorRet(&DirecotryIsNotEmptyError{mountpoint}, "Directory is not empty and cannot be removed.", logs.Args{{"mountpoint", mountpoint}})
		}

		if err := s.exec.RemoveAll(mountpoint); err != nil { // TODO its enough to do Remove without All.
			return s.logger.ErrorRet(err, "RemoveAll failed", logs.Args{{"mountpoint", mountpoint}})
		}
	} else {
		if os.IsNotExist(err) {
			s.logger.Warning("Idempotent issue encountered: mountpoint directory does not exist.", logs.Args{{"mountpoint", mountpoint}})
		}
	}

	return nil

}

func (s *scbeMounter) ActionAfterDetach(request resources.AfterDetachRequest) error {
	// no action after detach
	return nil
}
