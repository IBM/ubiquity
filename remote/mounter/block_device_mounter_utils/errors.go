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

import "fmt"

type DeviceAlreadyMountedToWrongMountpoint struct {
	device     string
	mountpoint string
}

func (e *DeviceAlreadyMountedToWrongMountpoint) Error() string {
	return fmt.Sprintf("Device is already mounted but to unexpected mountpoint. device=[%s], mountpoint=[%s]", e.device, e.mountpoint)
}

type DirPathAlreadyMountedToWrongDevice struct {
	mountPoint            string
	expectedDevice        string
	unexpectedDevicesRefs []string
}

func (e *DirPathAlreadyMountedToWrongDevice) Error() string {
	return fmt.Sprintf("[%s] directory is already a mountpoint but to unexpected devices=%v (expected mountpoint only on device=[%s])",
		e.mountPoint, e.unexpectedDevicesRefs, e.expectedDevice)
}

type PVIsAlreadyUsedByAnotherPod struct {
	mountpoint string
	slink      []string
}

func (e *PVIsAlreadyUsedByAnotherPod) Error() string {
	return fmt.Sprintf("PV is already in use by another pod and has an existing slink to mountpoint. mountpoint=[%s], slinks=[%s]", e.mountpoint, e.slink)
}

type WrongK8sDirectoryPathError struct {
	k8smountdir string
}

func (e *WrongK8sDirectoryPathError) Error() string {
	return fmt.Sprintf("Expected to find \"%s\" directory in k8s mount path. k8smountdir=[%s]", K8sPodsDirecotryName, e.k8smountdir)
}
