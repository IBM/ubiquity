package mounter_test

import (
	"fmt"
	"testing"

	"github.com/IBM/ubiquity/fakes"
	"github.com/IBM/ubiquity/remote/mounter"
	"github.com/IBM/ubiquity/remote/mounter/block_device_utils"
	"github.com/IBM/ubiquity/resources"
	"github.com/IBM/ubiquity/utils"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	///"errors"
)

const (
	fakePhysicalCapacity = 2040
	fakeLogicalCapacity  = 2040
	fakeUsedCapacity     = 2040
	fakeDS8kStoragetType = "2107"
	fakeV7kStorageType   = "2076"
	fakeA9kStorageType   = "2810"
	fakeProfile          = "gold"
)

var _ = Describe("scbe_mounter_test", func() {
	var (
		fakeExec    *fakes.FakeExecutor
		fakeBdUtils *fakes.FakeBlockDeviceMounterUtils
		scbeMounter resources.Mounter

		callErr error = &block_device_utils.VolumeNotFoundError{"wwn"}

		mountRequestForDS8kLun0 = resources.MountRequest{Mountpoint: "test_mountpointDS8k", VolumeConfig: map[string]interface{}{"Name": "u_vol", "PhysicalCapacity": fakePhysicalCapacity,
			"Profile": fakeProfile, "StorageType": fakeDS8kStoragetType, "UsedCapacity": fakeUsedCapacity, "Wwn": "wwn", "attach-to": "xnode1",
			"LogicalCapacity": fakeLogicalCapacity, "LunNumber": float64(0), "PoolName": "pool", "StorageName": "IBM.2107", "fstype": "ext4"}}
		mountRequestForSVCLun0 = resources.MountRequest{Mountpoint: "test_mountpointSVC", VolumeConfig: map[string]interface{}{"Name": "u_vol", "PhysicalCapacity": fakePhysicalCapacity,
			"Profile": fakeProfile, "StorageType": fakeV7kStorageType, "UsedCapacity": fakeUsedCapacity, "Wwn": "wwn", "attach-to": "node1",
			"LogicalCapacity": fakeLogicalCapacity, "LunNumber": float64(0), "PoolName": "pool", "StorageName": "IBM.2706", "fstype": "ext4"}}
		mountRequestForDS8kLun1 = resources.MountRequest{Mountpoint: "test_mountpointDS8k", VolumeConfig: map[string]interface{}{"Name": "u_vol", "PhysicalCapacity": fakePhysicalCapacity,
			"Profile": fakeProfile, "StorageType": fakeDS8kStoragetType, "UsedCapacity": fakeUsedCapacity, "Wwn": "wwn", "attach-to": "node1",
			"LogicalCapacity": fakeLogicalCapacity, "LunNumber": float64(1), "PoolName": "pool", "StorageName": "IBM.2107", "fstype": "ext4"}}
		mountRequestForSVCLun1 = resources.MountRequest{Mountpoint: "test_mountpointSVC", VolumeConfig: map[string]interface{}{"Name": "u_vol", "PhysicalCapacity": fakePhysicalCapacity,
			"Profile": fakeProfile, "StorageType": fakeV7kStorageType, "UsedCapacity": fakeUsedCapacity, "Wwn": "wwn", "attach-to": "node1",
			"LogicalCapacity": fakeLogicalCapacity, "LunNumber": float64(1), "PoolName": "pool", "StorageName": "IBM.2706", "fstype": "ext4"}}
		mountRequestForDS8kLun2 = resources.MountRequest{Mountpoint: "test_mountpointDS8k", VolumeConfig: map[string]interface{}{"Name": "u_vol", "PhysicalCapacity": fakePhysicalCapacity,
			"Profile": fakeProfile, "UsedCapacity": fakeUsedCapacity, "Wwn": "wwn", "attach-to": "node1",
			"LogicalCapacity": fakeLogicalCapacity, "LunNumber": float64(1), "PoolName": "pool", "StorageName": "IBM.2107", "fstype": "ext4"}}
		mountRequestForDS8kLun3 = resources.MountRequest{Mountpoint: "test_mountpointDS8k", VolumeConfig: map[string]interface{}{"Name": "u_vol", "PhysicalCapacity": fakePhysicalCapacity,
			"Profile": fakeProfile, "StorageType": fakeDS8kStoragetType, "UsedCapacity": fakeUsedCapacity, "Wwn": "wwn", "attach-to": "node1",
			"LogicalCapacity": fakeLogicalCapacity, "PoolName": "pool", "StorageName": "IBM.2107", "fstype": "ext4"}}
		mountRequestForA9kLun3 = resources.MountRequest{Mountpoint: "test_mountpointA9k", VolumeConfig: map[string]interface{}{"Name": "u_vol", "PhysicalCapacity": fakePhysicalCapacity,
			"Profile": fakeProfile, "StorageType": fakeA9kStorageType, "UsedCapacity": fakeUsedCapacity, "Wwn": "wwn", "attach-to": "node1",
			"LogicalCapacity": fakeLogicalCapacity, "PoolName": "pool", "StorageName": "IBM.2810", "fstype": "ext4"}}
	)

	BeforeEach(func() {
		fakeExec = new(fakes.FakeExecutor)
		fakeBdUtils = new(fakes.FakeBlockDeviceMounterUtils)
		scbeMounter = mounter.NewScbeMounterWithExecuter(fakeBdUtils, fakeExec)
		fakeExec.IsDirEmptyReturns(true, nil)
	})

	Context(".Unmount", func() {
		It("should continue flow if volume is not discovered", func() {
			returnedErr := &block_device_utils.VolumeNotFoundError{"volumewwn"}
			fakeBdUtils.DiscoverReturns("", returnedErr)
			volumeConfig := make(map[string]interface{})
			volumeConfig["Wwn"] = "volumewwn"
			err := scbeMounter.Unmount(resources.UnmountRequest{volumeConfig, resources.RequestContext{}})
			Expect(err).To(BeNil())
			Expect(fakeBdUtils.UnmountDeviceFlowCallCount()).To(Equal(0))
			Expect(fakeExec.StatCallCount()).To(Equal(1))
		})
		It("should fail if error from discovery is not voumeNotFound", func() {
			returnedErr := fmt.Errorf("An error has occured")
			fakeBdUtils.DiscoverReturns("", returnedErr)
			volumeConfig := make(map[string]interface{})
			volumeConfig["Wwn"] = "volumewwn"
			err := scbeMounter.Unmount(resources.UnmountRequest{volumeConfig, resources.RequestContext{}})
			Expect(err).To(Equal(returnedErr))
			Expect(fakeBdUtils.UnmountDeviceFlowCallCount()).To(Equal(0))
			Expect(fakeExec.StatCallCount()).To(Equal(0))
		})
		It("should call unmountDeviceFlow if discover succeeded", func() {
			volumeConfig := make(map[string]interface{})
			volumeConfig["Wwn"] = "volumewwn"
			err := scbeMounter.Unmount(resources.UnmountRequest{volumeConfig, resources.RequestContext{}})
			Expect(err).ToNot(HaveOccurred())
			Expect(fakeBdUtils.UnmountDeviceFlowCallCount()).To(Equal(1))
			Expect(fakeExec.StatCallCount()).To(Equal(1))
		})
		It("should fail if unmountDeviceFlow returns an error", func() {
			returnedErr := fmt.Errorf("An error has occured")
			fakeBdUtils.UnmountDeviceFlowReturns(returnedErr)
			volumeConfig := make(map[string]interface{})
			volumeConfig["Wwn"] = "volumewwn"
			err := scbeMounter.Unmount(resources.UnmountRequest{volumeConfig, resources.RequestContext{}})
			Expect(err).To(HaveOccurred())
			Expect(fakeBdUtils.UnmountDeviceFlowCallCount()).To(Equal(1))
			Expect(fakeExec.StatCallCount()).To(Equal(0))
		})
		It("should not fail if stat on file returns error", func() {
			returnedErr := fmt.Errorf("An error has occured")
			fakeExec.StatReturns(nil, returnedErr)
			volumeConfig := make(map[string]interface{})
			volumeConfig["Wwn"] = "volumewwn"
			err := scbeMounter.Unmount(resources.UnmountRequest{volumeConfig, resources.RequestContext{}})
			Expect(err).ToNot(HaveOccurred())
			Expect(fakeBdUtils.UnmountDeviceFlowCallCount()).To(Equal(1))
			Expect(fakeExec.StatCallCount()).To(Equal(1))
			Expect(fakeExec.RemoveAllCallCount()).To(Equal(0))
		})
		It("should  fail if Removeall returns error", func() {
			returnedErr := fmt.Errorf("An error has occured")
			fakeExec.RemoveAllReturns(returnedErr)
			volumeConfig := make(map[string]interface{})
			volumeConfig["Wwn"] = "volumewwn"
			err := scbeMounter.Unmount(resources.UnmountRequest{volumeConfig, resources.RequestContext{}})
			Expect(err).To(HaveOccurred())
			Expect(fakeBdUtils.UnmountDeviceFlowCallCount()).To(Equal(1))
			Expect(fakeExec.StatCallCount()).To(Equal(1))
			Expect(fakeExec.RemoveAllCallCount()).To(Equal(1))
		})
		It("should  fail if mountpoint dir is not empty", func() {
			fakeExec.IsDirEmptyReturns(false, nil)
			volumeConfig := make(map[string]interface{})
			volumeConfig["Wwn"] = "volumewwn"
			err := scbeMounter.Unmount(resources.UnmountRequest{volumeConfig, resources.RequestContext{}})
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(&mounter.DirecotryIsNotEmptyError{fmt.Sprintf("/ubiquity/%s", volumeConfig["Wwn"])}))
			Expect(fakeBdUtils.UnmountDeviceFlowCallCount()).To(Equal(1))
			Expect(fakeExec.StatCallCount()).To(Equal(1))
			Expect(fakeExec.RemoveAllCallCount()).To(Equal(0))
		})
		It("should  fail if mountpoint dir returns erorr", func() {
			returnedErr := fmt.Errorf("An error has occured")
			fakeExec.IsDirEmptyReturns(false, returnedErr)
			volumeConfig := make(map[string]interface{})
			volumeConfig["Wwn"] = "volumewwn"
			err := scbeMounter.Unmount(resources.UnmountRequest{volumeConfig, resources.RequestContext{}})
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(returnedErr))
			Expect(fakeBdUtils.UnmountDeviceFlowCallCount()).To(Equal(1))
			Expect(fakeExec.StatCallCount()).To(Equal(1))
			Expect(fakeExec.RemoveAllCallCount()).To(Equal(0))
		})
		It("should  continue if discover failed on faulty device", func() {
			returnedErr := &block_device_utils.FaultyDeviceError{"mapthx"}
			fakeBdUtils.DiscoverReturns("", returnedErr)
			volumeConfig := make(map[string]interface{})
			volumeConfig["Wwn"] = "volumewwn"
			_ = scbeMounter.Unmount(resources.UnmountRequest{volumeConfig, resources.RequestContext{}})
			Expect(fakeBdUtils.DiscoverCallCount()).To(Equal(1))
			Expect(fakeBdUtils.UnmountDeviceFlowCallCount()).To(Equal(1))
			Expect(fakeExec.StatCallCount()).To(Equal(1))
			Expect(fakeExec.RemoveAllCallCount()).To(Equal(1))
		})
	})
	Context("Mount", func() {
		It("should success to discover ds8k with lun0", func() {
			fakeBdUtils.DiscoverReturnsOnCall(0, "", callErr)
			fakeBdUtils.DiscoverReturnsOnCall(1, "wwn", nil)
			fakeBdUtils.RescanAllReturnsOnCall(0, nil)
			fakeBdUtils.RescanAllReturnsOnCall(1, nil)
			_, err := scbeMounter.Mount(mountRequestForDS8kLun0)
			Expect(err).ToNot(HaveOccurred())
		})
		It("should success to discover svc with lun0", func() {
			fakeBdUtils.DiscoverReturnsOnCall(0, "wwn", nil)
			fakeBdUtils.DiscoverReturnsOnCall(1, "wwn", nil)
			fakeBdUtils.RescanAllReturnsOnCall(0, nil)
			fakeBdUtils.RescanAllReturnsOnCall(1, nil)
			_, err := scbeMounter.Mount(mountRequestForSVCLun0)
			Expect(err).ToNot(HaveOccurred())
		})
		It("should fail to discover ds8k with lun1 if failed to discover with '-r' ", func() {
			fakeBdUtils.DiscoverReturns("", callErr)
			fakeBdUtils.RescanAllReturns(nil)
			_, err := scbeMounter.Mount(mountRequestForDS8kLun1)
			Expect(err).To(HaveOccurred())
		})
		It("should fail to discover svc with lun1 if failed to discover with '-r' ", func() {
			fakeBdUtils.DiscoverReturns("", callErr)
			fakeBdUtils.RescanAllReturns(callErr)
			_, err := scbeMounter.Mount(mountRequestForSVCLun1)
			Expect(err).To(HaveOccurred())
		})
		It("should fail to discover ds8k with lun2 if no storageType", func() {
			fakeBdUtils.DiscoverReturns("", callErr)
			fakeBdUtils.RescanAllReturnsOnCall(0, nil)
			_, err := scbeMounter.Mount(mountRequestForDS8kLun2)
			Expect(err).To(HaveOccurred())
		})
		It("should fail to discover ds8k with lun3 if no lunNumber", func() {
			fakeBdUtils.DiscoverReturns("", callErr)
			fakeBdUtils.RescanAllReturnsOnCall(0, nil)
			_, err := scbeMounter.Mount(mountRequestForDS8kLun3)
			Expect(err).To(HaveOccurred())
		})
		It("should fail to discover A9k with lun3 if no lunNumber", func() {
			fakeBdUtils.DiscoverReturns("", callErr)
			fakeBdUtils.RescanAllReturnsOnCall(0, nil)
			_, err := scbeMounter.Mount(mountRequestForA9kLun3)
			Expect(err).To(HaveOccurred())
		})
	})
})

func TestSCBEMounter(t *testing.T) {
	RegisterFailHandler(Fail)
	defer utils.InitUbiquityServerTestLogger()()
	RunSpecs(t, "SCBEMounter Test Suite")
}
