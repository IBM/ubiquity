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

//go:generate counterfeiter -o ../fakes/fake_block_device_mounter_utils.go . BlockDeviceMounterUtils
type BlockDeviceMounterUtils interface {
	RescanAll(withISCSI bool, wwn string, rescanForCleanUp bool) error
	MountDeviceFlow(devicePath string, fsType string, mountPoint string) error
	Discover(volumeWwn string) (string, error)
	UnmountDeviceFlow(devicePath string) error
}

func NewBlockDeviceMounterUtilsWithBlockDeviceUtils(blockDeviceUtilsInst block_device_utils.BlockDeviceUtils) BlockDeviceMounterUtils {
	return newBlockDeviceMounterUtils(blockDeviceUtilsInst)
}

func NewBlockDeviceMounterUtils() BlockDeviceMounterUtils {
	return newBlockDeviceMounterUtils(block_device_utils.NewBlockDeviceUtils())
}

func newBlockDeviceMounterUtils(blockDeviceUtilsInst block_device_utils.BlockDeviceUtils) BlockDeviceMounterUtils {
	return &blockDeviceMounterUtils{logger: logs.GetLogger(),
		blockDeviceUtils:  blockDeviceUtilsInst,
		rescanLock:        &sync.RWMutex{},
		cleanMPDeviceLock: &sync.RWMutex{},
	}
}
