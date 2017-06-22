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