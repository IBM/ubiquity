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
	"os"
	"path/filepath"
	"time"
	"fmt"
	"strings"
)

const (
	WarningMessageIdempotentDeviceAlreadyMounted = "Device is already mounted, so skip mounting (Idempotent)."
	TimeoutMilisecondFindCommand = 30 * 1000     // max to wait for mount command
	K8sPodsDirecotryName = "pods"
	
)

type blockDeviceMounterUtils struct {
	logger           logs.Logger
	blockDeviceUtils block_device_utils.BlockDeviceUtils
	rescanFlock      lockfile.Lockfile
	mpathFlock       lockfile.Lockfile
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
		blockDeviceUtils: blockDeviceUtils,
		rescanFlock:      rescanLock,
		mpathFlock:       mpathLock,
	}
}

func (b *blockDeviceMounterUtils) getK8sBaseDir(k8sMountPoint string) (string, error ){
	/*
		we want to get from this kind of dir:
			/var/lib/kubelet/pods/1f94f1d9-8f36-11e8-b227-005056a4d4cb/volumes/ibm~ubiquity-k8s-flex/...
		to /var/lib/kubelet/pods/
		note: /var/lib can be changed in configuration so we need to get the base dir at runtime
	*/ 
	 out := strings.Split(k8sMountPoint, K8sPodsDirecotryName)
	 if len(out) ==1 {
	 	return "", &WrongK8sDirectoryPathError{k8sMountPoint}
	 }
	 return filepath.Join(out[0], K8sPodsDirecotryName) , nil
}

func (b *blockDeviceMounterUtils) checkSlinkAlreadyExistsOnMountPoint(mountPoint string, k8sMountPoint string) (bool, error, []string){
	/*
		returned params:
			1. bool: is this pvc already used,
			2. error : did an error occure in the check process
			3. slinks : the list of slinks found so that an appropriate error could be returned,
	*/
	
	defer b.logger.Trace(logs.INFO, logs.Args{{"mountPoint", mountPoint}, {"k8sMountPoint", k8sMountPoint}})()
	
	k8sBaseDir, err := b.getK8sBaseDir(k8sMountPoint)	
	if err != nil{
		return false, err, nil
	}
		
	file_pattern := filepath.Join(k8sBaseDir, "*","volumes", "ibm~ubiquity-k8s-flex","*")
	files, _ := filepath.Glob(file_pattern)
	
	slinks := []string{}
	
	// we will go over the files and check if any of them is the same as our destinated mountpoint
	for _, file := range files {
		fileStat, _ := os.Stat(file)
		mountStat, _ := os.Stat(mountPoint)
		isSameFile := os.SameFile(fileStat, mountStat)
		if isSameFile{
			slinks = append(slinks, file)
		}
	}
		
	if len(slinks) == 0 {
		// no slinks pointing to the mountpoint so we can continue mount process
		return false, nil, nil
	}
	
	if len(slinks) > 1 {
		// if there are more then 1 link already pointing to the mountpoint it is obviously used.
		return true, nil, slinks
	}
	
	slink := slinks[0]
	if slink != k8sMountPoint{
		// if the current link is not our volume then this pvc is in use
		return true, nil, slinks
	}
	
	return false, nil, nil
}

// MountDeviceFlow create filesystem on the device (if needed) and then mount it on a given mountpoint
func (b *blockDeviceMounterUtils) MountDeviceFlow(devicePath string, fsType string, mountPoint string, k8sMountPoint string) error {
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

	// Check if need to mount the device, if its already mounted then skip mounting
	isMounted, mountpointRefs, err := b.blockDeviceUtils.IsDeviceMounted(devicePath)
	if err != nil {
		return b.logger.ErrorRet(err, "fail to identify if device is mounted")
	}

	if isMounted {
		for _, mountpointi := range mountpointRefs {
			if mountpointi == mountPoint {
				// we need to check that there is no one else that is already mounted to this device before we continue.
				b.logger.Debug(fmt.Sprintf("k8s mount point : %s", k8sMountPoint))
				doesSlinkExists, err, slinkList := b.checkSlinkAlreadyExistsOnMountPoint(mountPoint, k8sMountPoint)
				if err!= nil {
					b.logger.ErrorRet(err, "fail")
				}
				if doesSlinkExists{
					// there is someone else that is mounted to this mount point
					return b.logger.ErrorRet(&PVCIsAlreadyUsedByAnotherPod{mountPoint, slinkList}, "fail")
				}
				
				b.logger.Warning(WarningMessageIdempotentDeviceAlreadyMounted, logs.Args{{"Device", devicePath}, {"mountpoint", mountPoint}})
				return nil // Indicate idempotent issue
			}
		}
		// In case device mounted but to different mountpoint as expected we fail with error. # TODO we may support it in the future after allow the umount flow to umount by mountpoint and not by device path.
		return b.logger.ErrorRet(&DeviceAlreadyMountedToWrongMountpoint{devicePath, mountPoint}, "fail")
	} else {
		// Check if mountpoint directory is not already mounted to un expected device. If so raise error to prevent double mounting.
		isMounted, devicesRefs, err := b.blockDeviceUtils.IsDirAMountPoint(mountPoint)
		if err != nil {
			return b.logger.ErrorRet(err, "fail to identify if mountpoint dir is actually mounted")
		}

		if isMounted {
			return b.logger.ErrorRet(&DirPathAlreadyMountedToWrongDevice{
				mountPoint: mountPoint, expectedDevice: devicePath, unexpectedDevicesRefs: devicesRefs},
				"fail")
		}
		if err = b.blockDeviceUtils.MountFs(devicePath, mountPoint); err != nil {
			return b.logger.ErrorRet(err, "MountFs failed")
		}
	}

	return nil
}

// UnmountDeviceFlow umount device, clean device and remove mountpoint folder
func (b *blockDeviceMounterUtils) UnmountDeviceFlow(devicePath string) error {
	// TODO consider also to receive the mountpoint(not only the devicepath), so the umount will use the mountpoint instead of the devicepath for future support double mounting of the same device.
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
		time.Sleep(time.Duration(500 * time.Millisecond))
	}
	b.logger.Debug("Got mpathLock for device", logs.Args{{"device", devicePath}})
	defer b.mpathFlock.Unlock()
	defer b.logger.Debug("Released mpathLock for device", logs.Args{{"device", devicePath}})

	if err := b.blockDeviceUtils.Cleanup(devicePath); err != nil {
		return b.logger.ErrorRet(err, "Cleanup failed")
	}

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
		time.Sleep(time.Duration(500 * time.Millisecond))
	}
	b.logger.Debug("Got rescanLock for volumeWWN", logs.Args{{"volumeWWN", wwn}})
	defer b.rescanFlock.Unlock()
	defer b.logger.Debug("Released rescanLock for volumeWWN", logs.Args{{"volumeWWN", wwn}})

	if !rescanForCleanUp {
		// Only when run rescan for new device, try to check if its already exist to reduce rescans
		device, _ := b.Discover(wwn, false) // no deep discovery

		if device != "" {
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

func (b *blockDeviceMounterUtils) Discover(volumeWwn string, deepDiscovery bool) (string, error) {
	return b.blockDeviceUtils.Discover(volumeWwn, deepDiscovery)
}
