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

package scbe

import "strconv"

func NewScbeVolumeInfo(volume *ScbeResponseVolume) (ScbeVolumeInfo) {
    return ScbeVolumeInfo{
        Name:             volume.Name,
        Wwn:              volume.ScsiIdentifier,
        Profile:          volume.ServiceName,
        StorageType:      volume.ArrayType,
        StorageName:      volume.ArrayName,
        PoolName:         volume.PoolName,
        LogicalCapacity:  strconv.Itoa(volume.LogicalCapacity),
        PhysicalCapacity: strconv.Itoa(volume.PhysicalCapacity),
        UsedCapacity:     strconv.Itoa(volume.UsedCapacity),
    }
}