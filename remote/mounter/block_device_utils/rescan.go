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
	"github.com/IBM/ubiquity/utils"
	"github.com/IBM/ubiquity/utils/logs"
)

const rescanIscsiTimeout = 1 * 60 * 1000
const rescanScsiTimeout = 2 * 60 * 1000

// store the volume mount info to a local cache.
// It is just a workaround, We can not get the multipath devices from multipath -ll
// output in the cleanup stage. because we run multipath -f before it.
var volumeCache = make(map[string]*resources.VolumeMountProperties)

var FcHostDir = "/sys/class/fc_host/"
var ScsiHostDir = "/sys/class/scsi_host/"

func getVolumeFromCache(volumeMountProperties *resources.VolumeMountProperties) *resources.VolumeMountProperties {
	if volume, exists := volumeCache[volumeMountProperties.WWN]; exists {
		return volume
	}
	return volumeMountProperties
}

func storeVolumeToCache(volumeMountProperties *resources.VolumeMountProperties) {
	volume := new(resources.VolumeMountProperties)
	*volume = *volumeMountProperties
	volumeCache[volume.WWN] = volume
}

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

func (b *blockDeviceUtils) CleanupDevices(protocol Protocol, volumeMountProperties *resources.VolumeMountProperties) error {
	defer b.logger.Trace(logs.DEBUG)()

	switch protocol {
	case SCSI:
		return b.CleanupSCSIDevices(volumeMountProperties)
	case ISCSI:
		return b.CleanupISCSIDevices()
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
	defer b.logger.Trace(logs.DEBUG)()

	var err error
	var devMapper string
	var devNames []string
	for i := 0; i < 6; i++ {
		if err = b.fcConnector.ConnectVolume(volumeMountProperties); err != nil {
			return b.logger.ErrorRet(err, "RescanSCSI failed", logs.Args{{"volumeWWN", volumeMountProperties.WWN}})
		}
		if _, devMapper, devNames, err = utils.GetMultipathOutputAndDeviceMapperAndDevice(volumeMountProperties.WWN, b.exec); err == nil {
			volumeMountProperties.DeviceMapper = devMapper
			volumeMountProperties.Devices = devNames
			// store the volumeMountProperties to a local cache, it will be used in cleanup stage.
			storeVolumeToCache(volumeMountProperties)
			return nil
		}
		b.logger.Warning("Can't find the new volume in multipath output after rescan, sleep one second and try again.")
		time.Sleep(1 * time.Second)
	}
	return b.logger.ErrorRet(err, "RescanSCSI failed", logs.Args{{"volumeWWN", volumeMountProperties.WWN}})
}

// TODO: improve it to make it faster, for more details, see os_brick project.
func (b *blockDeviceUtils) CleanupISCSIDevices() error {
	return b.RescanISCSI()
}

func (b *blockDeviceUtils) CleanupSCSIDevices(volumeMountProperties *resources.VolumeMountProperties) error {
	defer b.logger.Trace(logs.DEBUG)()

	return b.fcConnector.DisconnectVolume(getVolumeFromCache(volumeMountProperties))
}
