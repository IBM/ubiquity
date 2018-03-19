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
	device string
	mountpoint string
}

func (e *DeviceAlreadyMountedToWrongMountpoint) Error() string {
	return fmt.Sprintf("Device is already mounted but to unexpected mountpoint. device=[%s], mountpoint=[%s]", e.device, e.mountpoint)
}
