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
	"fmt"
	"github.com/IBM/ubiquity/fakes"
	"github.com/IBM/ubiquity/remote/mounter/block_device_mounter_utils"
	"github.com/IBM/ubiquity/remote/mounter/block_device_utils"
	"github.com/IBM/ubiquity/utils/logs"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"testing"
)

var _ = Describe("block_device_mounter_utils_test", func() {
	var (
		fakeBlockDeviceUtils    *fakes.FakeBlockDeviceUtils
		blockDeviceMounterUtils block_device_mounter_utils.BlockDeviceMounterUtils
		err                     error
	)

	BeforeEach(func() {
		fakeBlockDeviceUtils = new(fakes.FakeBlockDeviceUtils)
		blockDeviceMounterUtils = block_device_mounter_utils.NewBlockDeviceMounterUtilsWithBlockDeviceUtils(fakeBlockDeviceUtils)
	})

	Context(".MountDeviceFlow", func() {
		It("should fail if checkfs failed", func() {
			fakeBlockDeviceUtils.CheckFsReturns(true, fmt.Errorf("error"))
			err = blockDeviceMounterUtils.MountDeviceFlow("fake_device", "fake_fstype", "fake_mountp")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("error"))
		})
		It("should fail if mkfs failed", func() {
			fakeBlockDeviceUtils.CheckFsReturns(true, nil)
			fakeBlockDeviceUtils.MakeFsReturns(fmt.Errorf("error"))
			err = blockDeviceMounterUtils.MountDeviceFlow("fake_device", "fake_fstype", "fake_mountp")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("error"))
			Expect(fakeBlockDeviceUtils.MakeFsCallCount()).To(Equal(1))

		})
		It("should fail if mkfs failed", func() {
			fakeBlockDeviceUtils.CheckFsReturns(true, nil)
			fakeBlockDeviceUtils.MakeFsReturns(fmt.Errorf("error"))
			err = blockDeviceMounterUtils.MountDeviceFlow("fake_device", "fake_fstype", "fake_mountp")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("error"))
			Expect(fakeBlockDeviceUtils.MakeFsCallCount()).To(Equal(1))

		})
		It("should fail if mountfs failed", func() {
			fakeBlockDeviceUtils.CheckFsReturns(true, nil)
			fakeBlockDeviceUtils.MakeFsReturns(nil)
			fakeBlockDeviceUtils.MountFsReturns(fmt.Errorf("error"))
			err = blockDeviceMounterUtils.MountDeviceFlow("fake_device", "fake_fstype", "fake_mountp")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("error"))
			Expect(fakeBlockDeviceUtils.MountFsCallCount()).To(Equal(1))
		})
		It("should succeed (with create fs) if all if cool", func() {
			fakeBlockDeviceUtils.CheckFsReturns(true, nil)
			fakeBlockDeviceUtils.MakeFsReturns(nil)
			fakeBlockDeviceUtils.MountFsReturns(nil)
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
			err = blockDeviceMounterUtils.RescanAll(true, "wwn", false)
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeBlockDeviceUtils.RescanCallCount()).To(Equal(0))
		})
		It("should fail if scsi rescan fail", func() {
			fakeBlockDeviceUtils.RescanReturnsOnCall(0, nil)
			fakeBlockDeviceUtils.RescanReturnsOnCall(1, fmt.Errorf("error"))
			fakeBlockDeviceUtils.DiscoverReturns("", fmt.Errorf("device not exist yet"))
			err = blockDeviceMounterUtils.RescanAll(true, "wwn", false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("error"))
			Expect(fakeBlockDeviceUtils.RescanCallCount()).To(Equal(2))
			protocol := fakeBlockDeviceUtils.RescanArgsForCall(0)
			Expect(protocol).To(Equal(block_device_utils.ISCSI))
			protocol = fakeBlockDeviceUtils.RescanArgsForCall(1)
			Expect(protocol).To(Equal(block_device_utils.SCSI))
		})
		It("should fail if scsi rescan fail even if for clean up", func() {
			fakeBlockDeviceUtils.RescanReturnsOnCall(0, nil)
			fakeBlockDeviceUtils.RescanReturnsOnCall(1, fmt.Errorf("error"))
			fakeBlockDeviceUtils.DiscoverReturns("", fmt.Errorf("device not exist yet"))
			err = blockDeviceMounterUtils.RescanAll(true, "wwn", true)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("error"))
			Expect(fakeBlockDeviceUtils.RescanCallCount()).To(Equal(2))
			protocol := fakeBlockDeviceUtils.RescanArgsForCall(0)
			Expect(protocol).To(Equal(block_device_utils.ISCSI))
			protocol = fakeBlockDeviceUtils.RescanArgsForCall(1)
			Expect(protocol).To(Equal(block_device_utils.SCSI))
		})
		It("should fail if ReloadMultipath fail", func() {
			fakeBlockDeviceUtils.RescanReturnsOnCall(0, nil)
			fakeBlockDeviceUtils.RescanReturnsOnCall(1, nil)
			fakeBlockDeviceUtils.ReloadMultipathReturns(fmt.Errorf("error"))
			fakeBlockDeviceUtils.DiscoverReturns("", fmt.Errorf("device not exist yet"))
			err = blockDeviceMounterUtils.RescanAll(true, "wwn", false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("error"))
			Expect(fakeBlockDeviceUtils.RescanCallCount()).To(Equal(2))
			protocol := fakeBlockDeviceUtils.RescanArgsForCall(0)
			Expect(protocol).To(Equal(block_device_utils.ISCSI))
			protocol = fakeBlockDeviceUtils.RescanArgsForCall(1)
			Expect(protocol).To(Equal(block_device_utils.SCSI))
			Expect(fakeBlockDeviceUtils.ReloadMultipathCallCount()).To(Equal(1))

		})
		It("should succeed to rescall all", func() {
			fakeBlockDeviceUtils.RescanReturnsOnCall(0, nil)
			fakeBlockDeviceUtils.RescanReturnsOnCall(1, nil)
			fakeBlockDeviceUtils.ReloadMultipathReturns(nil)
			fakeBlockDeviceUtils.DiscoverReturns("", fmt.Errorf("device not exist yet"))
			err = blockDeviceMounterUtils.RescanAll(true, "wwn", false)
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeBlockDeviceUtils.RescanCallCount()).To(Equal(2))
			protocol := fakeBlockDeviceUtils.RescanArgsForCall(0)
			Expect(protocol).To(Equal(block_device_utils.ISCSI))
			protocol = fakeBlockDeviceUtils.RescanArgsForCall(1)
			Expect(protocol).To(Equal(block_device_utils.SCSI))
			Expect(fakeBlockDeviceUtils.ReloadMultipathCallCount()).To(Equal(1))
		})
		It("should succeed to rescall all (no iscsi call)", func() {
			fakeBlockDeviceUtils.RescanReturnsOnCall(0, nil)
			fakeBlockDeviceUtils.ReloadMultipathReturns(nil)
			fakeBlockDeviceUtils.DiscoverReturns("", fmt.Errorf("device not exist yet"))
			err = blockDeviceMounterUtils.RescanAll(false, "wwn", false)
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeBlockDeviceUtils.RescanCallCount()).To(Equal(1))
			protocol := fakeBlockDeviceUtils.RescanArgsForCall(0)
			Expect(protocol).To(Equal(block_device_utils.SCSI))
			Expect(fakeBlockDeviceUtils.ReloadMultipathCallCount()).To(Equal(1))
		})
	})
	Context(".UnmountDeviceFlow", func() {
		It("should fail if unmount failed", func() {
			fakeBlockDeviceUtils.UmountFsReturns(fmt.Errorf("error"))
			err = blockDeviceMounterUtils.UnmountDeviceFlow("fake_device")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("error"))
		})
		It("should fail if Cleanup failed", func() {
			fakeBlockDeviceUtils.UmountFsReturns(nil)
			fakeBlockDeviceUtils.CleanupReturns(fmt.Errorf("error"))
			err = blockDeviceMounterUtils.UnmountDeviceFlow("fake_device")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("error"))
			Expect(fakeBlockDeviceUtils.CleanupCallCount()).To(Equal(1))
		})
		It("should succees if all is cool", func() {
			fakeBlockDeviceUtils.UmountFsReturns(nil)
			fakeBlockDeviceUtils.CleanupReturns(nil)
			err = blockDeviceMounterUtils.UnmountDeviceFlow("fake_device")
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeBlockDeviceUtils.CleanupCallCount()).To(Equal(1))
		})

	})

})

func TestGetBlockDeviceUtils(t *testing.T) {
	RegisterFailHandler(Fail)
	defer logs.InitStdoutLogger(logs.DEBUG)()
	RunSpecs(t, "BlockDeviceUtils Test Suite")
}
