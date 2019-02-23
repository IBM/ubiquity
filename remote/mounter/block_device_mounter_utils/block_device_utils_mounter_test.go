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

package block_device_mounter_utils_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/IBM/ubiquity/fakes"
	"github.com/IBM/ubiquity/remote/mounter/block_device_mounter_utils"
	"github.com/IBM/ubiquity/remote/mounter/block_device_utils"
	"github.com/IBM/ubiquity/resources"
	"github.com/IBM/ubiquity/utils"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("block_device_mounter_utils_test", func() {
	var (
		fakeBlockDeviceUtils    *fakes.FakeBlockDeviceUtils
		blockDeviceMounterUtils block_device_mounter_utils.BlockDeviceMounterUtils
		err                     error
		callErr                 error = errors.New("error")
	)
	volumeMountProperties := &resources.VolumeMountProperties{WWN: "wwn", LunNumber: 1}

	BeforeEach(func() {
		fakeBlockDeviceUtils = new(fakes.FakeBlockDeviceUtils)
		blockDeviceMounterUtils = block_device_mounter_utils.NewBlockDeviceMounterUtilsWithBlockDeviceUtils(fakeBlockDeviceUtils)
	})

	Context(".MountDeviceFlow", func() {
		It("should fail if checkfs failed", func() {
			fakeBlockDeviceUtils.CheckFsReturns(true, callErr)
			err = blockDeviceMounterUtils.MountDeviceFlow("fake_device", "fake_fstype", "fake_mountp")
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(callErr))
		})
		It("should fail if mkfs failed", func() {
			fakeBlockDeviceUtils.CheckFsReturns(true, nil)
			fakeBlockDeviceUtils.MakeFsReturns(callErr)
			err = blockDeviceMounterUtils.MountDeviceFlow("fake_device", "fake_fstype", "fake_mountp")
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(callErr))
			Expect(fakeBlockDeviceUtils.MakeFsCallCount()).To(Equal(1))

		})
		It("should fail if IsDeviceMounted failed", func() {
			fakeBlockDeviceUtils.CheckFsReturns(true, nil)
			fakeBlockDeviceUtils.MakeFsReturns(nil)
			fakeBlockDeviceUtils.IsDeviceMountedReturns(false, nil, callErr)

			err = blockDeviceMounterUtils.MountDeviceFlow("fake_device", "fake_fstype", "fake_mountp")

			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(callErr))
			Expect(fakeBlockDeviceUtils.CheckFsCallCount()).To(Equal(1))
			Expect(fakeBlockDeviceUtils.MakeFsCallCount()).To(Equal(1))
			Expect(fakeBlockDeviceUtils.IsDeviceMountedCallCount()).To(Equal(1))
			Expect(fakeBlockDeviceUtils.MountFsCallCount()).To(Equal(0))

		})
		It("should succeed mount even if already mounted (idempotent should skip mount)", func() {
			fakeBlockDeviceUtils.CheckFsReturns(true, nil)
			fakeBlockDeviceUtils.MakeFsReturns(nil)
			fakeBlockDeviceUtils.IsDeviceMountedReturns(true, []string{"fake_mountp"}, nil)

			err = blockDeviceMounterUtils.MountDeviceFlow("fake_device", "fake_fstype", "fake_mountp")

			Expect(err).NotTo(HaveOccurred())
			Expect(fakeBlockDeviceUtils.CheckFsCallCount()).To(Equal(1))
			Expect(fakeBlockDeviceUtils.MakeFsCallCount()).To(Equal(1))
			Expect(fakeBlockDeviceUtils.IsDeviceMountedCallCount()).To(Equal(1))
			Expect(fakeBlockDeviceUtils.MountFsCallCount()).To(Equal(0)) // skip due to idempotent

		})

		It("should fail to mount if the mountpoint already exist but on un expected mountpoint.", func() {
			fakeBlockDeviceUtils.CheckFsReturns(true, nil)
			fakeBlockDeviceUtils.MakeFsReturns(nil)
			fakeBlockDeviceUtils.IsDeviceMountedReturns(true, []string{"fake_mountpNOTEXPECTED"}, nil)

			err = blockDeviceMounterUtils.MountDeviceFlow("fake_device", "fake_fstype", "fake_mountp")

			Expect(err).To(HaveOccurred())
			_, ok := err.(*block_device_mounter_utils.DeviceAlreadyMountedToWrongMountpoint)
			Expect(ok).To(Equal(true))
			Expect(fakeBlockDeviceUtils.CheckFsCallCount()).To(Equal(1))
			Expect(fakeBlockDeviceUtils.MakeFsCallCount()).To(Equal(1))
			Expect(fakeBlockDeviceUtils.IsDeviceMountedCallCount()).To(Equal(1))
			Expect(fakeBlockDeviceUtils.MountFsCallCount()).To(Equal(0)) // skip due to idempotent
		})

		It("should fail to mount, while its not already mounted but the mount ops fails", func() {
			fakeBlockDeviceUtils.CheckFsReturns(true, nil)
			fakeBlockDeviceUtils.MakeFsReturns(nil)
			fakeBlockDeviceUtils.MountFsReturns(callErr)
			fakeBlockDeviceUtils.IsDeviceMountedReturns(false, nil, nil)

			err = blockDeviceMounterUtils.MountDeviceFlow("fake_device", "fake_fstype", "fake_mountp")

			Expect(err).To(HaveOccurred())
			Expect(fakeBlockDeviceUtils.CheckFsCallCount()).To(Equal(1))
			Expect(fakeBlockDeviceUtils.MakeFsCallCount()).To(Equal(1))
			Expect(fakeBlockDeviceUtils.IsDeviceMountedCallCount()).To(Equal(1))
			Expect(fakeBlockDeviceUtils.MountFsCallCount()).To(Equal(1)) // skip due to idempotent

		})

		It("should succeed to mount on regular flow (device not already mounted so mounting it with success)", func() {
			fakeBlockDeviceUtils.CheckFsReturns(true, nil)
			fakeBlockDeviceUtils.MakeFsReturns(nil)
			fakeBlockDeviceUtils.MountFsReturns(nil)
			fakeBlockDeviceUtils.IsDeviceMountedReturns(false, nil, nil)

			err = blockDeviceMounterUtils.MountDeviceFlow("fake_device", "fake_fstype", "fake_mountp")

			Expect(err).NotTo(HaveOccurred())
			Expect(fakeBlockDeviceUtils.CheckFsCallCount()).To(Equal(1))
			Expect(fakeBlockDeviceUtils.MakeFsCallCount()).To(Equal(1))
			Expect(fakeBlockDeviceUtils.IsDeviceMountedCallCount()).To(Equal(1))
			Expect(fakeBlockDeviceUtils.MountFsCallCount()).To(Equal(1)) // skip due to idempotent

		})

		It("should fail if IsDirAMountPoint failed", func() {
			fakeBlockDeviceUtils.CheckFsReturns(true, nil)
			fakeBlockDeviceUtils.IsDeviceMountedReturns(false, nil, nil)
			fakeBlockDeviceUtils.IsDirAMountPointReturns(false, nil, callErr)
			err = blockDeviceMounterUtils.MountDeviceFlow("fake_device", "fake_fstype", "fake_mountp")
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(callErr))
			Expect(fakeBlockDeviceUtils.IsDirAMountPointCallCount()).To(Equal(1))
		})

		It("should fail to mount if mountpoint is already mounted to wrong device", func() {
			fakeBlockDeviceUtils.CheckFsReturns(true, nil)
			fakeBlockDeviceUtils.IsDeviceMountedReturns(false, nil, nil)
			fakeBlockDeviceUtils.IsDirAMountPointReturns(true, []string{"/dev/mapper/mpathvfake1", "/dev/mapper/mpathvfake2"}, nil)
			err = blockDeviceMounterUtils.MountDeviceFlow("fake_device", "fake_fstype", "fake_mountp")
			Expect(err).To(HaveOccurred())
			_, ok := err.(*block_device_mounter_utils.DirPathAlreadyMountedToWrongDevice)
			Expect(ok).To(Equal(true))
			Expect(fakeBlockDeviceUtils.IsDirAMountPointCallCount()).To(Equal(1))
			Expect(fakeBlockDeviceUtils.MountFsCallCount()).To(Equal(0))

		})

		It("should fail if mountfs failed", func() {
			fakeBlockDeviceUtils.CheckFsReturns(true, nil)
			fakeBlockDeviceUtils.MakeFsReturns(nil)
			fakeBlockDeviceUtils.MountFsReturns(callErr)
			fakeBlockDeviceUtils.IsDirAMountPointReturns(false, nil, nil)

			err = blockDeviceMounterUtils.MountDeviceFlow("fake_device", "fake_fstype", "fake_mountp")
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(callErr))
			Expect(fakeBlockDeviceUtils.MountFsCallCount()).To(Equal(1))
		})
		It("should succeed (with create fs) if all if cool", func() {
			fakeBlockDeviceUtils.CheckFsReturns(true, nil)
			fakeBlockDeviceUtils.MakeFsReturns(nil)
			fakeBlockDeviceUtils.MountFsReturns(nil)
			fakeBlockDeviceUtils.IsDirAMountPointReturns(false, nil, nil)

			err = blockDeviceMounterUtils.MountDeviceFlow("fake_device", "fake_fstype", "fake_mountp")
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeBlockDeviceUtils.MakeFsCallCount()).To(Equal(1))
			Expect(fakeBlockDeviceUtils.MountFsCallCount()).To(Equal(1))
		})
		It("should succeed (without create fs) if all if cool", func() {
			fakeBlockDeviceUtils.CheckFsReturns(false, nil)
			fakeBlockDeviceUtils.MountFsReturns(nil)
			err = blockDeviceMounterUtils.MountDeviceFlow("fake_device", "fake_fstype", "fake_mountp")
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeBlockDeviceUtils.MakeFsCallCount()).To(Equal(0))
			Expect(fakeBlockDeviceUtils.MountFsCallCount()).To(Equal(1))
		})
	})
	Context(".RescanAll", func() {
		It("should succeed to skip rescan we try to rescan(for discover) a wwn that is already descovered", func() {
			fakeBlockDeviceUtils.DiscoverReturns("wwn", nil)
			err = blockDeviceMounterUtils.RescanAll(volumeMountProperties)
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeBlockDeviceUtils.RescanCallCount()).To(Equal(0))
		})
		It("should fail if iscsi rescan fail", func() {
			fakeBlockDeviceUtils.RescanReturnsOnCall(0, callErr)
			fakeBlockDeviceUtils.DiscoverReturns("", fmt.Errorf("device not exist yet"))
			err = blockDeviceMounterUtils.RescanAll(volumeMountProperties)
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(callErr))
			Expect(fakeBlockDeviceUtils.RescanCallCount()).To(Equal(1))
			protocol, _ := fakeBlockDeviceUtils.RescanArgsForCall(0)
			Expect(protocol).To(Equal(block_device_utils.ISCSI))
		})
		It("should fail if scsi rescan fail", func() {
			fakeBlockDeviceUtils.RescanReturnsOnCall(0, nil)
			fakeBlockDeviceUtils.RescanReturnsOnCall(1, callErr)
			fakeBlockDeviceUtils.DiscoverReturns("", fmt.Errorf("device not exist yet"))
			err = blockDeviceMounterUtils.RescanAll(volumeMountProperties)
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(callErr))
			Expect(fakeBlockDeviceUtils.RescanCallCount()).To(Equal(2))
			protocol, _ := fakeBlockDeviceUtils.RescanArgsForCall(0)
			Expect(protocol).To(Equal(block_device_utils.ISCSI))
			protocol, _ = fakeBlockDeviceUtils.RescanArgsForCall(1)
			Expect(protocol).To(Equal(block_device_utils.SCSI))
		})
		It("should fail if ReloadMultipath fail", func() {
			fakeBlockDeviceUtils.RescanReturnsOnCall(0, nil)
			fakeBlockDeviceUtils.RescanReturnsOnCall(1, nil)
			fakeBlockDeviceUtils.ReloadMultipathReturns(callErr)
			fakeBlockDeviceUtils.DiscoverReturns("", fmt.Errorf("device not exist yet"))
			err = blockDeviceMounterUtils.RescanAll(volumeMountProperties)
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(callErr))
			Expect(fakeBlockDeviceUtils.RescanCallCount()).To(Equal(2))
			protocol, _ := fakeBlockDeviceUtils.RescanArgsForCall(0)
			Expect(protocol).To(Equal(block_device_utils.ISCSI))
			protocol, _ = fakeBlockDeviceUtils.RescanArgsForCall(1)
			Expect(protocol).To(Equal(block_device_utils.SCSI))
			Expect(fakeBlockDeviceUtils.ReloadMultipathCallCount()).To(Equal(1))

		})
		It("should succeed to rescall all", func() {
			fakeBlockDeviceUtils.RescanReturnsOnCall(0, nil)
			fakeBlockDeviceUtils.RescanReturnsOnCall(1, nil)
			fakeBlockDeviceUtils.ReloadMultipathReturns(nil)
			fakeBlockDeviceUtils.DiscoverReturns("", fmt.Errorf("device not exist yet"))
			err = blockDeviceMounterUtils.RescanAll(volumeMountProperties)
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeBlockDeviceUtils.RescanCallCount()).To(Equal(2))
			protocol, _ := fakeBlockDeviceUtils.RescanArgsForCall(0)
			Expect(protocol).To(Equal(block_device_utils.ISCSI))
			protocol, _ = fakeBlockDeviceUtils.RescanArgsForCall(1)
			Expect(protocol).To(Equal(block_device_utils.SCSI))
			Expect(fakeBlockDeviceUtils.ReloadMultipathCallCount()).To(Equal(1))
		})
	})
	Context(".UnmountDeviceFlow", func() {
		It("should fail if unmount failed", func() {
			fakeBlockDeviceUtils.UmountFsReturns(callErr)
			err = blockDeviceMounterUtils.UnmountDeviceFlow("fake_device", "6001738CFC9035EA0000000000795164")
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(callErr))
		})
		It("should fail if dmsetup failed", func() {
			fakeBlockDeviceUtils.SetDmsetupReturns(callErr)
			fakeBlockDeviceUtils.UmountFsReturns(nil)
			err = blockDeviceMounterUtils.UnmountDeviceFlow("fake_device", "6001738CFC9035EA0000000000795164")
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(callErr))
			Expect(fakeBlockDeviceUtils.SetDmsetupCallCount()).To(Equal(1))
			Expect(fakeBlockDeviceUtils.UmountFsCallCount()).To(Equal(0))
		})
		It("should fail if Cleanup failed", func() {
			fakeBlockDeviceUtils.UmountFsReturns(nil)
			fakeBlockDeviceUtils.CleanupReturns(callErr)
			err = blockDeviceMounterUtils.UnmountDeviceFlow("fake_device", "6001738CFC9035EA0000000000795164")
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(callErr))
			Expect(fakeBlockDeviceUtils.CleanupCallCount()).To(Equal(1))
		})
		It("should succees if all is cool", func() {
			fakeBlockDeviceUtils.UmountFsReturns(nil)
			fakeBlockDeviceUtils.CleanupReturns(nil)
			err = blockDeviceMounterUtils.UnmountDeviceFlow("fake_device", "6001738CFC9035EA0000000000795164")
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeBlockDeviceUtils.CleanupCallCount()).To(Equal(1))
		})
	})
})

func TestGetBlockDeviceMounterUtils(t *testing.T) {
	RegisterFailHandler(Fail)
	defer utils.InitUbiquityServerTestLogger()()
	RunSpecs(t, "BlockDeviceMounterUtils Test Suite")
}
