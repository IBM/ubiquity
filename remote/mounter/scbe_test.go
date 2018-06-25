package mounter_test

import (
	"github.com/IBM/ubiquity/fakes"
	. "github.com/IBM/ubiquity/remote/mounter"
	. "github.com/IBM/ubiquity/resources"
	"github.com/IBM/ubiquity/utils/logs"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"errors"
)

const (
	fakePhysicalCapacity = 2040
	fakeLogicalCapacity  = 2040
	fakeUsedCapacity     = 2040
	fakeDS8kStoragetType = "2107"
	fakeV7kStorageType   = "2076"
	fakeProfile          = "gold"
)

var _ = Describe("Scbe", func() {
	defer logs.InitStdoutLogger(logs.DEBUG)()

	var (
		callErr                     error   = errors.New("error")
		scbeRemoteConfig                    = ScbeRemoteConfig{true}
		fakeBlockDeviceUtilsMounter         = new(fakes.FakeBlockDeviceMounterUtils)
		sMounter                    Mounter = NewTestScbeMounter(scbeRemoteConfig, fakeBlockDeviceUtilsMounter)

		mountRequestForDS8kLun0 = MountRequest{Mountpoint: "test_mountpointDS8k", VolumeConfig: map[string]interface{}{"Name": "u_vol", "PhysicalCapacity": fakePhysicalCapacity,
			"Profile": fakeProfile, "StorageType": fakeDS8kStoragetType, "UsedCapacity": fakeUsedCapacity, "Wwn": "wwn", "attach-to": "xnode1",
			"LogicalCapacity": fakeLogicalCapacity, "LunNumber": float64(0), "PoolName": "pool", "StorageName": "IBM.2107", "fstype": "ext4"}}

		mountRequestForSVCLun0 = MountRequest{Mountpoint: "test_mountpointSVC", VolumeConfig: map[string]interface{}{"Name": "u_vol", "PhysicalCapacity": fakePhysicalCapacity,
			"Profile": fakeProfile, "StorageType": fakeV7kStorageType, "UsedCapacity": fakeUsedCapacity, "Wwn": "wwn", "attach-to": "node1",
			"LogicalCapacity": fakeLogicalCapacity, "LunNumber": float64(0), "PoolName": "pool", "StorageName": "IBM.2706", "fstype": "ext4"}}

		mountRequestForDS8kLun1 = MountRequest{Mountpoint: "test_mountpointDS8k", VolumeConfig: map[string]interface{}{"Name": "u_vol", "PhysicalCapacity": fakePhysicalCapacity,
			"Profile": fakeProfile, "StorageType": fakeDS8kStoragetType, "UsedCapacity": fakeUsedCapacity, "Wwn": "wwn", "attach-to": "node1",
			"LogicalCapacity": fakeLogicalCapacity, "LunNumber": float64(1), "PoolName": "pool", "StorageName": "IBM.2107", "fstype": "ext4"}}

		mountRequestForSVCLun1 = MountRequest{Mountpoint: "test_mountpointSVC", VolumeConfig: map[string]interface{}{"Name": "u_vol", "PhysicalCapacity": fakePhysicalCapacity,
			"Profile": fakeProfile, "StorageType": fakeV7kStorageType, "UsedCapacity": fakeUsedCapacity, "Wwn": "wwn", "attach-to": "node1",
			"LogicalCapacity": fakeLogicalCapacity, "LunNumber": float64(1), "PoolName": "pool", "StorageName": "IBM.2706", "fstype": "ext4"}}
	)

	Context("Mount", func() {
		It("should success to discover ds8k with lun0", func() {
			fakeBlockDeviceUtilsMounter.DiscoverReturnsOnCall(0, "", callErr)
			fakeBlockDeviceUtilsMounter.DiscoverReturnsOnCall(1, "wwn", nil)
			fakeBlockDeviceUtilsMounter.RescanAllReturnsOnCall(0, nil)
			fakeBlockDeviceUtilsMounter.RescanAllTargetsReturnsOnCall(0, nil)
			_, err := sMounter.Mount(mountRequestForDS8kLun0)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should success to discover svc with lun0", func() {
			fakeBlockDeviceUtilsMounter.DiscoverReturnsOnCall(0, "wwn", nil)
			fakeBlockDeviceUtilsMounter.RescanAllReturnsOnCall(0, nil)
			_, err := sMounter.Mount(mountRequestForSVCLun0)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should fail to discover ds8k with lun1 if failed to discover with '-r' ", func() {
			fakeBlockDeviceUtilsMounter.DiscoverReturns("", callErr)
			fakeBlockDeviceUtilsMounter.RescanAllReturns(nil)
			_, err := sMounter.Mount(mountRequestForDS8kLun1)
			Expect(err).To(HaveOccurred())
		})

		It("should fail to discover svc with lun1 if failed to discover with '-r' ", func() {
			fakeBlockDeviceUtilsMounter.DiscoverReturns("", callErr)
			fakeBlockDeviceUtilsMounter.RescanAllReturns(callErr)
			_, err := sMounter.Mount(mountRequestForSVCLun1)
			Expect(err).To(HaveOccurred())
		})
	})
})
