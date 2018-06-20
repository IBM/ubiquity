package mounter_test

import (
	. "github.com/IBM/ubiquity/remote/mounter"
	. "github.com/IBM/ubiquity/resources"
	"github.com/IBM/ubiquity/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/IBM/ubiquity/utils/logs"

	"errors"
)

var _ = Describe("Scbe", func() {
	defer logs.InitStdoutLogger(logs.DEBUG)()

	var (
		callErr  error = errors.New("error")
		scbeRemoteConfig = ScbeRemoteConfig{true}
		fakeBlockDeviceUtilsMounter = new(fakes.FakeBlockDeviceMounterUtils)
		sMounter Mounter = NewTestScbeMounter(scbeRemoteConfig, fakeBlockDeviceUtilsMounter)

		mountRequestForDS8kLun0 = MountRequest{Mountpoint:"test_mountpointDS8k", VolumeConfig:map[string]interface{}{"Name":"u_vol","PhysicalCapacity":2040,
		"Profile":"gold", "StorageType":"2107", "UsedCapacity":2040, "Wwn":"wwn", "attach-to":"xnode1",
		"LogicalCapacity":2040, "LunNumber":float64(0), "PoolName":"pool", "StorageName":"IBM.2107", "fstype":"ext4"}}

		mountRequestForSVCLun0 = MountRequest{Mountpoint:"test_mountpointSVC", VolumeConfig:map[string]interface{}{"Name":"u_vol", "PhysicalCapacity":2040,
		"Profile":"gold", "StorageType":"2706", "UsedCapacity":2040, "Wwn":"wwn", "attach-to":"node1",
		"LogicalCapacity":2040, "LunNumber":float64(0), "PoolName":"pool", "StorageName":"IBM.2706", "fstype":"ext4"}}

		mountRequestForDS8kLun1 = MountRequest{Mountpoint:"test_mountpointDS8k", VolumeConfig:map[string]interface{}{"Name":"u_vol","PhysicalCapacity":2040,
		"Profile":"gold", "StorageType":"2107", "UsedCapacity":2040, "Wwn":"wwn", "attach-to":"node1",
		"LogicalCapacity":2040, "LunNumber":float64(1), "PoolName":"pool", "StorageName":"IBM.2107", "fstype":"ext4"}}

		mountRequestForSVCLun1 = MountRequest{Mountpoint:"test_mountpointSVC", VolumeConfig:map[string]interface{}{"Name":"u_vol", "PhysicalCapacity":2040,
		"Profile":"gold", "StorageType":"2706", "UsedCapacity":2040, "Wwn":"wwn", "attach-to":"node1",
		"LogicalCapacity":2040, "LunNumber":float64(1), "PoolName":"pool", "StorageName":"IBM.2706", "fstype":"ext4"}}
	)

	Context("Mount", func() {
		It("should success to discover ds8k with lun0", func() {
			fakeBlockDeviceUtilsMounter.DiscoverReturnsOnCall(0, "", callErr)
			fakeBlockDeviceUtilsMounter.DiscoverReturnsOnCall(1, "wwn", nil )
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
