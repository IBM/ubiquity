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

package block_device_utils_test

import (
	"errors"
	"fmt"

	"github.com/IBM/ubiquity/fakes"
	"github.com/IBM/ubiquity/remote/mounter/block_device_utils"
	fakeinitiator "github.com/IBM/ubiquity/remote/mounter/initiator/fakes"
	"github.com/IBM/ubiquity/resources"
	"github.com/IBM/ubiquity/utils"

	"context"
	"io/ioutil"
	"os/exec"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("block_device_utils_test", func() {
	var (
		fakeExec        *fakes.FakeExecutor
		fakeFcConnector *fakeinitiator.FakeConnector
		bdUtils         block_device_utils.BlockDeviceUtils
		err             error
		cmdErr          error = errors.New("command error")
	)
	volumeMountProperties := &resources.VolumeMountProperties{WWN: "wwn", LunNumber: 1}

	BeforeEach(func() {
		fakeExec = new(fakes.FakeExecutor)
		fakeFcConnector = new(fakeinitiator.FakeConnector)
		bdUtils = block_device_utils.NewBlockDeviceUtilsWithExecutorAndConnector(fakeExec, fakeFcConnector)
	})

	Context(".Rescan", func() {
		It("Rescan ISCSI calls 'sudo iscsiadm -m session --rescan'", func() {
			err = bdUtils.Rescan(block_device_utils.ISCSI, volumeMountProperties)
			Expect(err).ToNot(HaveOccurred())
			Expect(fakeExec.ExecuteWithTimeoutCallCount()).To(Equal(1))
			_, cmd, args := fakeExec.ExecuteWithTimeoutArgsForCall(0)
			Expect(cmd).To(Equal("iscsiadm"))
			Expect(args).To(Equal([]string{"-m", "session", "--rescan"}))
		})
		It(`Rescan SCSI calls fcConnector.ConnectVolume`, func() {
			fakeExec.ExecuteWithTimeoutReturns([]byte(volumeMountProperties.WWN), nil)
			err = bdUtils.Rescan(block_device_utils.SCSI, volumeMountProperties)
			Expect(err).ToNot(HaveOccurred())
			Expect(fakeExec.ExecuteWithTimeoutCallCount()).To(Equal(1))
			_, cmd, args := fakeExec.ExecuteWithTimeoutArgsForCall(0)
			// check the existence of the new volume after rescan using multipath -ll
			Expect(cmd).To(Equal("multipath"))
			Expect(args).To(Equal([]string{"-ll"}))
			Expect(fakeFcConnector.ConnectVolumeCallCount()).To(Equal(1))
		})
		It("Rescan ISCSI does not fail if iscsiadm command missing", func() {
			fakeExec.IsExecutableReturns(cmdErr)
			err = bdUtils.Rescan(block_device_utils.ISCSI, volumeMountProperties)
			Expect(err).To(Not(HaveOccurred()))
			Expect(fakeExec.ExecuteCallCount()).To(Equal(0))
			Expect(fakeExec.IsExecutableCallCount()).To(Equal(1))
			Expect(fakeExec.IsExecutableArgsForCall(0)).To(Equal("iscsiadm"))
		})
		It("Rescan ISCSI fails if iscsiadm execution fails", func() {
			fakeExec.ExecuteWithTimeoutReturns([]byte{}, cmdErr)
			err = bdUtils.Rescan(block_device_utils.ISCSI, volumeMountProperties)
			Expect(err.Error()).To(MatchRegexp(cmdErr.Error()))
		})
		It("Rescan fails if unknown protocol", func() {
			err = bdUtils.Rescan(2, volumeMountProperties)
			Expect(err).To(HaveOccurred())
		})
	})
	Context(".CleanupDevices", func() {
		It("Cleanup ISCSI calls 'sudo iscsiadm -m session --rescan'", func() {
			err = bdUtils.CleanupDevices(block_device_utils.ISCSI, volumeMountProperties)
			Expect(err).ToNot(HaveOccurred())
			Expect(fakeExec.ExecuteWithTimeoutCallCount()).To(Equal(1))
			_, cmd, args := fakeExec.ExecuteWithTimeoutArgsForCall(0)
			Expect(cmd).To(Equal("iscsiadm"))
			Expect(args).To(Equal([]string{"-m", "session", "--rescan"}))
		})
		It(`Cleanup SCSI calls fcConnector.DisconnectVolume`, func() {
			err = bdUtils.CleanupDevices(block_device_utils.SCSI, volumeMountProperties)
			Expect(err).ToNot(HaveOccurred())
			Expect(fakeFcConnector.DisconnectVolumeCallCount()).To(Equal(1))
		})
	})
	Context(".ReloadMultipath", func() {
		It("ReloadMultipath calls multipath command", func() {
			err = bdUtils.ReloadMultipath()
			Expect(err).ToNot(HaveOccurred())
			Expect(fakeExec.ExecuteWithTimeoutCallCount()).To(Equal(2))
			tiemout, cmd, args := fakeExec.ExecuteWithTimeoutArgsForCall(0)
			Expect(cmd).To(Equal("multipath"))
			Expect(args).To(Equal([]string{}))
			Expect(tiemout).To(Equal(block_device_utils.MultipathTimeout))
			tiemout, cmd, args = fakeExec.ExecuteWithTimeoutArgsForCall(1)
			Expect(cmd).To(Equal("multipath"))
			Expect(args).To(Equal([]string{"-r"}))
			Expect(tiemout).To(Equal(block_device_utils.MultipathTimeout))
		})
		It("ReloadMultipath fails if multipath command is missing", func() {
			fakeExec.IsExecutableReturns(cmdErr)
			err = bdUtils.ReloadMultipath()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(MatchRegexp(cmdErr.Error()))
		})
		It("ReloadMultipath fails if multipath command fails", func() {
			fakeExec.ExecuteWithTimeoutReturns([]byte{}, cmdErr)
			err = bdUtils.ReloadMultipath()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(MatchRegexp(cmdErr.Error()))
		})
	})
	Context(".Discover", func() {
		It("Discover returns path for volume", func() {
			volumeId := "0x6001738cfc9035eb0000000000cea5f6"
			result := "mpath"
			inq_result := fmt.Sprintf(`VPD INQUIRY: Device Identification page
							Designation descriptor number 1, descriptor length: 20
							designator_type: NAA,  code_set: Binary
							associated with the addressed logical unit
							NAA 6, IEEE Company_id: 0x1738
							Vendor Specific Identifier: 0xcfc9035eb
							Vendor Specific Identifier Extension: 0xcea5f6
							[%s]`, volumeId)
			fakeExec.ExecuteWithTimeoutReturnsOnCall(0, []byte(fmt.Sprintf("%s (%s) dm-1", result, volumeId)),
				nil)
			fakeExec.ExecuteWithTimeoutReturnsOnCall(1, []byte(fmt.Sprintf("%s", inq_result)), nil) // for getWwnByScsiInq
			mpath, err := bdUtils.Discover(strings.TrimPrefix(volumeId, "0x"), true)
			Expect(err).ToNot(HaveOccurred())
			Expect(mpath).To(Equal("/dev/mapper/" + result))
			Expect(fakeExec.ExecuteWithTimeoutCallCount()).To(Equal(2))
			_, cmd, args := fakeExec.ExecuteWithTimeoutArgsForCall(0)
			Expect(cmd).To(Equal("multipath"))
			Expect(args).To(Equal([]string{"-ll"}))
			_, cmd, args = fakeExec.ExecuteWithTimeoutArgsForCall(1)
			Expect(cmd).To(Equal("sg_inq"))
			Expect(args).To(Equal([]string{"-p", "0x83", "/dev/mapper/mpath"}))
		})
		It("Discover fails if multipath command is missing", func() {
			volumeId := "volume-id"
			fakeExec.IsExecutableReturns(cmdErr)
			_, err := bdUtils.Discover(volumeId, true)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(MatchRegexp(cmdErr.Error()))
			_, err = bdUtils.Discover(volumeId, false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(MatchRegexp(cmdErr.Error()))
		})
		It("Discover fails if multipath -ll command fails", func() {
			volumeId := "volume-id"
			fakeExec.ExecuteWithTimeoutReturns([]byte{}, cmdErr)
			_, err := bdUtils.Discover(volumeId, true)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(MatchRegexp(cmdErr.Error()))
		})
		It("Discover fails if wrong wwn is found", func() {
			volumeId := "0x6001738cfc9035eb0000000000cea5f6"
			wrongVolumeId := "0x6001738cfc9035eb000000000079sfjs"
			result := "mpath"
			inq_result := fmt.Sprintf(`VPD INQUIRY: Device Identification page
							Designation descriptor number 1, descriptor length: 20
							designator_type: NAA,  code_set: Binary
							associated with the addressed logical unit
							NAA 6, IEEE Company_id: 0x1738
							Vendor Specific Identifier: 0xcfc9035eb
							Vendor Specific Identifier Extension: 0xcea5f6
							[%s]`, wrongVolumeId)
			fakeExec.ExecuteWithTimeoutReturnsOnCall(0, []byte(fmt.Sprintf("%s (%s) dm-1", result, volumeId)),
				nil)
			fakeExec.ExecuteWithTimeoutReturnsOnCall(1, []byte(fmt.Sprintf("%s", inq_result)), nil) // for getWwnByScsiInq
			_, err := bdUtils.Discover(strings.TrimPrefix(volumeId, "0x"), true)
			Expect(err).To(HaveOccurred())
			_, cmd, args := fakeExec.ExecuteWithTimeoutArgsForCall(0)
			Expect(cmd).To(Equal("multipath"))
			Expect(args).To(Equal([]string{"-ll"}))
			_, cmd, args = fakeExec.ExecuteWithTimeoutArgsForCall(1)
			Expect(cmd).To(Equal("sg_inq"))
			Expect(args).To(Equal([]string{"-p", "0x83", "/dev/mapper/mpath"}))
		})
		It("Discover fails if volume not found", func() {
			volumeId := "volume-id"
			fakeExec.ExecuteReturns([]byte(fmt.Sprintf(
				"mpath (other-volume-1) dm-1\nmpath (other-volume-2) dm-2")), nil)
			_, err := bdUtils.Discover(volumeId, true)
			Expect(err).To(HaveOccurred())
		})
		It("should return volume not found when sg_inq command fails", func() {
			volumeId := "0x6001738cfc9035eb0000000000cea5f6"
			mpathOutput := `mpathhe (36001738cfc9035eb0000000000cea5f6) dm-3 IBM     ,2810XIV
							size=19G features='1 queue_if_no_path' hwhandler='0' wp=rw
							-+- policy='service-time 0' prio=1 status=active
							|- 33:0:0:1 sdb 8:16 active ready running
							- 34:0:0:1 sdc 8:32 active ready running`
			fakeExec.ExecuteWithTimeoutReturnsOnCall(0, []byte(mpathOutput), nil)
			returnError := &exec.ExitError{}
			//this execute with timeout makes the GetWwnByScsiInq to return an error
			fakeExec.ExecuteWithTimeoutReturnsOnCall(1, []byte(""), returnError)
			_, err := bdUtils.Discover(volumeId, true)
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(&block_device_utils.VolumeNotFoundError{volumeId}))
		})
		It("should return faulty error when failing sg_inq on faulty device.", func() {
			volumeId := "6001738cfc9035ea0000000000796463"
			device := "mpathx"
			mpathOutput := fmt.Sprintf(`%s (3%s) dm-4 IBM     ,2810XIV         
							size=954M features='1 queue_if_no_path' hwhandler='0' wp=rw
							-+- policy='service-time 0' prio=0 status=active
							  |- 35:0:0:2 sdd 8:48 failed faulty running
							  |- 36:0:0:2 sdg 8:96 failed faulty running
							  - 37:0:0:2 sde 8:64 failed faulty running`, device, volumeId)
			fakeExec.ExecuteWithTimeoutReturnsOnCall(0, []byte(mpathOutput), nil)
			returnError := &exec.ExitError{}
			//this execute with timeout makes the GetWwnByScsiInq to return an error
			fakeExec.ExecuteWithTimeoutReturnsOnCall(2, []byte(""), returnError)
			fakeExec.ExecuteReturns([]byte(mpathOutput), nil)
			_, err := bdUtils.Discover(volumeId, true)
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(&block_device_utils.FaultyDeviceError{fmt.Sprintf("/dev/mapper/%s", device)}))
		})
		It("should return command execution error if device is not faulty.", func() {
			volumeId := "6001738cfc9035ea0000000000796463"
			device := "mpathx"
			mpathOutput := fmt.Sprintf(`%s (3%s) dm-4 IBM     ,2810XIV         
							size=954M features='1 queue_if_no_path' hwhandler='0' wp=rw
							-+- policy='service-time 0' prio=0 status=active
							  |- 35:0:0:2 sdd 8:48 active ready running
							  |- 36:0:0:2 sdg 8:96 active ready running
							  - 37:0:0:2 sde 8:64 active ready running`, device, volumeId)
			fakeExec.ExecuteWithTimeoutReturnsOnCall(0, []byte(mpathOutput), nil)
			returnError := &exec.ExitError{}
			//this execute with timeout makes the GetWwnByScsiInq to return an error
			fakeExec.ExecuteWithTimeoutReturnsOnCall(2, []byte(""), returnError)
			fakeExec.ExecuteReturns([]byte(mpathOutput), nil)
			_, err := bdUtils.Discover(volumeId, true)
			Expect(err).To(HaveOccurred())
			_, ok := err.(*block_device_utils.CommandExecuteError)
			Expect(ok).To(BeTrue())
		})
		It("should return faulty device error on volume with no vendor ", func() {
			volumeId := "6001738cfc9035eb0000000000cea5f6"
			mpathOutput := `mpatha (36001738cfc9035eb0000000000cea5f6) dm-2
size=224G features='1 queue_if_no_path' hwhandler='0' wp=rw`
			fakeExec.ExecuteWithTimeoutReturnsOnCall(0, []byte(mpathOutput), nil)
			returnError := &exec.ExitError{}
			//this execute with timeout makes the GetWwnByScsiInq to return an error
			fakeExec.ExecuteWithTimeoutReturnsOnCall(1, []byte(""), returnError)
			_, err := bdUtils.Discover(volumeId, true)
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(&block_device_utils.FaultyDeviceError{fmt.Sprintf("/dev/mapper/%s", "mpatha")}))
		})
	})
	Context(".DiscoverBySgInq", func() {
		It("should return mpathhe", func() {
			mpathOutput := `mpathhe (36001738cfc9035eb0000000000cea5f6) dm-3 IBM     ,2810XIV
							size=19G features='1 queue_if_no_path' hwhandler='0' wp=rw
							-+- policy='service-time 0' prio=1 status=active
							|- 33:0:0:1 sdb 8:16 active ready running
							- 34:0:0:1 sdc 8:32 active ready running`
			volWwn := "0x6001738cfc9035eb0000000000cea5f6"
			expectedWwn := strings.TrimPrefix(volWwn, "0x")
			inqResult := fmt.Sprintf(`VPD INQUIRY: Device Identification page
							Designation descriptor number 1, descriptor length: 20
							designator_type: NAA,  code_set: Binary
							associated with the addressed logical unit
							NAA 6, IEEE Company_id: 0x1738
							Vendor Specific Identifier: 0xcfc9035eb
							Vendor Specific Identifier Extension: 0xcea5f6
							[%s]`, volWwn)
			fakeExec.ExecuteWithTimeoutReturns([]byte(fmt.Sprintf("%s", inqResult)), nil)
			dev, err := bdUtils.DiscoverBySgInq(mpathOutput, expectedWwn)
			Expect(dev).To(Equal("mpathhe"))
			Expect(err).ToNot(HaveOccurred())
			Expect(fakeExec.ExecuteWithTimeoutCallCount()).To(Equal(1))
			_, cmd, _ := fakeExec.ExecuteWithTimeoutArgsForCall(0)
			Expect(cmd).To(Equal("sg_inq"))
		})

		It("should return the 3rd mpath device mpathhg", func() {
			volWwn := "0x6001738cfc9035eb0000000000cea5f6"
			expectedWwn := strings.TrimPrefix(volWwn, "0x")
			mpathOutput := fmt.Sprintf(`mpathhe (36001738cfc9035eb0000000000cerwr) dm-3 IBM     ,2810XIV
							size=19G features='1 queue_if_no_path' hwhandler='0' wp=rw
							-+- policy='service-time 0' prio=1 status=active
							|- 32:0:0:1 sdb 8:16 fault faulty running
							- 31:0:0:1 sdc 8:32 fault faulty running
mpathhf (36001738cfc9035eb0000000000crwrfw24) dm-3 IBM     ,2810XIV
							size=19G features='1 queue_if_no_path' hwhandler='0' wp=rw
							-+- policy='service-time 0' prio=1 status=active
							|- 33:0:0:1 sdb 8:16 fault faulty running
							- 34:0:0:1 sdc 8:32 fault faulty running
mpathhg (3%s) dm-3 IBM     ,2810XIV
							size=19G features='1 queue_if_no_path' hwhandler='0' wp=rw
							-+- policy='service-time 0' prio=1 status=active
							|- 33:0:0:1 sdb 8:16 active ready running
							- 34:0:0:1 sdc 8:32 active ready running`, expectedWwn)
			inqResult := fmt.Sprintf(`VPD INQUIRY: Device Identification page
							Designation descriptor number 1, descriptor length: 20
							designator_type: NAA,  code_set: Binary
							associated with the addressed logical unit
							NAA 6, IEEE Company_id: 0x1738
							Vendor Specific Identifier: 0xcfc9035eb
							Vendor Specific Identifier Extension: 0xcea5f6
							[%s]`, volWwn)

			fmt.Println(mpathOutput)
			// the first call to sg_inq will only be on a non faulty device!
			fakeExec.ExecuteWithTimeoutReturnsOnCall(0, []byte(fmt.Sprintf("%s", inqResult)), nil)
			dev, err := bdUtils.DiscoverBySgInq(mpathOutput, expectedWwn)
			Expect(dev).To(Equal("mpathhg"))
			Expect(err).ToNot(HaveOccurred())
			Expect(fakeExec.ExecuteWithTimeoutCallCount()).To(Equal(1))
			_, cmd, _ := fakeExec.ExecuteWithTimeoutArgsForCall(0)
			Expect(cmd).To(Equal("sg_inq"))
		})

		It("should not sg_inq none IBM devices like vendor AAA or faulty device ##", func() {
			volWwn := "6001738cfc9035eb0000000000cea5f6"
			deviceName := "mpathhe"
			mpathOutput := fmt.Sprintf(`
mpathha (36001738cfc9035eb0000000000ceaaaa) dm-3 AAA,BBB
mpathhb (36001738cfc9035eb0000000000cea###) dm-3 ##,##
%s (3%s) dm-3 IBM     ,2810XIV`, deviceName, volWwn)

			volWwnHexa := "0x" + volWwn
			inq_result := fmt.Sprintf(`VPD INQUIRY: Device Identification page
							Designation descriptor number 1, descriptor length: 20
							designator_type: NAA,  code_set: Binary
							associated with the addressed logical unit
							NAA 6, IEEE Company_id: 0x1738
							Vendor Specific Identifier: 0xcfc9035eb
							Vendor Specific Identifier Extension: 0xcea5f6
							[%s]`, volWwnHexa)
			fakeExec.ExecuteWithTimeoutReturns([]byte(fmt.Sprintf("%s", inq_result)), nil)
			dev, err := bdUtils.DiscoverBySgInq(mpathOutput, volWwn)
			Expect(dev).To(Equal(deviceName))
			Expect(err).ToNot(HaveOccurred())
			Expect(fakeExec.ExecuteWithTimeoutCallCount()).To(Equal(1)) // only 1 sg_inq on the right one, means it skip the AAA and ## devces as expected.
			_, cmd, _ := fakeExec.ExecuteWithTimeoutArgsForCall(0)
			Expect(cmd).To(Equal("sg_inq"))
		})

		It("should return wwn command fails", func() {
			mpathOutput := `mpathhe (36001738cfc9035eb0000000000cea5f6) dm-3 IBM     ,2810XIV
							size=19G features='1 queue_if_no_path' hwhandler='0' wp=rw
							-+- policy='service-time 0' prio=1 status=active
							|- 33:0:0:1 sdb 8:16 active ready running
							- 34:0:0:1 sdc 8:32 active ready running`
			wwn := "wwn"
			fakeExec.ExecuteReturns([]byte{}, cmdErr)
			_, err := bdUtils.DiscoverBySgInq(mpathOutput, wwn)
			Expect(err).To(HaveOccurred())
		})
		It("should return actual error when sg_inq command fails", func() {
			mpathOutput := `mpathhe (36001738cfc9035eb0000000000cea5f6) dm-3 IBM     ,2810XIV
							size=19G features='1 queue_if_no_path' hwhandler='0' wp=rw
							-+- policy='service-time 0' prio=1 status=active
							|- 33:0:0:1 sdb 8:16 active ready running
							- 34:0:0:1 sdc 8:32 active ready running`
			wwn := "wwn"
			returnError := &exec.ExitError{}
			//this execute with timeout makes the GetWwnByScsiInq to return an error
			fakeExec.ExecuteWithTimeoutReturns([]byte(""), returnError)
			_, err := bdUtils.DiscoverBySgInq(mpathOutput, wwn)
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(&block_device_utils.VolumeNotFoundError{wwn}))
		})
	})
	Context(".GetWwnByScsiInq", func() {
		It("GetWwnByScsiInq fails if sg_inq command fails", func() {
			dev := "dev"
			fakeExec.ExecuteWithTimeoutReturns([]byte{}, cmdErr)
			_, err := bdUtils.GetWwnByScsiInq("", dev)
			Expect(err).To(HaveOccurred())
		})
		It("should return wwn for mpath device", func() {
			dev := "dev"
			expecedWwn := "0x6001738cfc9035eb0000000000AAAAAA"
			result := fmt.Sprintf(`VPD INQUIRY: Device Identification page
							Designation descriptor number 1, descriptor length: 20
							designator_type: NAA,  code_set: Binary
							associated with the addressed logical unit
							NAA 6, IEEE Company_id: 0x1738
							Vendor Specific Identifier: 0xcfc9035eb
							Vendor Specific Identifier Extension: 0xcea5f6
							[%s]`, expecedWwn)
			fakeExec.ExecuteWithTimeoutReturns([]byte(fmt.Sprintf("%s", result)), nil)
			wwn, err := bdUtils.GetWwnByScsiInq("", dev)
			Expect(err).ToNot(HaveOccurred())
			Expect(wwn).To(Equal(strings.TrimPrefix(expecedWwn, "0x")))
			Expect(fakeExec.ExecuteWithTimeoutCallCount()).To(Equal(1))
			_, cmd, args := fakeExec.ExecuteWithTimeoutArgsForCall(0)
			Expect(cmd).To(Equal("sg_inq"))
			Expect(args).To(Equal([]string{"-p", "0x83", dev}))
		})
		It("should return wwn for mpath device on zLinux output", func() {
			dev := "dev"
			expecedWwn := "0x6001738cfc9035eb0000000000AAAAAA"
			result := fmt.Sprintf(`VPD INQUIRY: Device Identification page
                                                        Designation descriptor number 1, descriptor length: 20
                                                        designator_type: NAA,  code_set: Binary
                                                        associated with the addressed logical unit
                                                        NAA 6, IEEE Company_id: 0x1738
                                                        Vendor Specific Identifier: 0xcfc9035eb
                                                        Vendor Specific Extension Identifier: 0xcea5f6
                                                        [%s]`, expecedWwn)
			fakeExec.ExecuteWithTimeoutReturns([]byte(fmt.Sprintf("%s", result)), nil)
			wwn, err := bdUtils.GetWwnByScsiInq("", dev)
			Expect(err).ToNot(HaveOccurred())
			Expect(wwn).To(Equal(strings.TrimPrefix(expecedWwn, "0x")))
			Expect(fakeExec.ExecuteWithTimeoutCallCount()).To(Equal(1))
			_, cmd, args := fakeExec.ExecuteWithTimeoutArgsForCall(0)
			Expect(cmd).To(Equal("sg_inq"))
			Expect(args).To(Equal([]string{"-p", "0x83", dev}))
		})
		It("should not find wwn for device", func() {
			dev := "dev"
			expecedWwn := "6001738cfc9035eb0000000000AAAAAA"
			result := fmt.Sprintf(`VPD INQUIRY: Device Identification page
							Designation descriptor number 1, descriptor length: 20
							designator_type: NAA,  code_set: Binary
							associated with the addressed logical unit
							NAA 6, IEEE Company_id: 0x1738
							Vendor Specific Identifier: 0xcfc9035eb
							Vendor Specific Identifier Extension: 0xcea5f6
							[%s]`, expecedWwn)
			fakeExec.ExecuteWithTimeoutReturns([]byte(fmt.Sprintf("%s", result)), nil)
			_, err := bdUtils.GetWwnByScsiInq("", dev)
			Expect(err).To(HaveOccurred())
			Expect(fakeExec.ExecuteWithTimeoutCallCount()).To(Equal(1))
			_, cmd, args := fakeExec.ExecuteWithTimeoutArgsForCall(0)
			Expect(cmd).To(Equal("sg_inq"))
			Expect(args).To(Equal([]string{"-p", "0x83", dev}))
		})
		It("should return error if device is faulty", func() {
			dev := "mpathhe"
			mpath := `mpathhe (36001738cfc9035eb0000000000cea5f6) dm-3 IBM     ,2810XIV
							size=19G features='1 queue_if_no_path' hwhandler='0' wp=rw
							-+- policy='service-time 0' prio=1 status=active
							|- 33:0:0:1 sdb 8:16 fault faulty running
							- 34:0:0:1 sdc 8:32 fault faulty running`
			_, err := bdUtils.GetWwnByScsiInq(mpath, dev)
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(&block_device_utils.FaultyDeviceError{dev}))
			Expect(fakeExec.ExecuteWithTimeoutCallCount()).To(Equal(0))
		})
		It("should return an error if check for faulty device failed", func() {
			dev := "mpath"
			mpath := `mpathhe (36001738cfc9035eb0000000000cea5f6) dm-3 IBM     ,2810XIV
							size=19G features='1 queue_if_no_path' hwhandler='0' wp=rw
							-+- policy='service-time 0' prio=1 status=active
							|- 33:0:0:1 sdb 8:16 fault faulty running
							- 34:0:0:1 sdc 8:32 fault faulty running`
			_, err := bdUtils.GetWwnByScsiInq(mpath, dev)
			Expect(err).To(HaveOccurred())
			Expect(fakeExec.ExecuteWithTimeoutCallCount()).To(Equal(1))
		})
	})

	Context(".Cleanup", func() {
		It("Cleanup calls dmsetup and multipath", func() {
			mpath := "mpath"
			err = bdUtils.Cleanup(mpath)
			Expect(err).ToNot(HaveOccurred())
			Expect(fakeExec.ExecuteWithTimeoutCallCount()).To(Equal(1))
			_, cmd1, args1 := fakeExec.ExecuteWithTimeoutArgsForCall(0)
			Expect(cmd1).To(Equal("multipath"))
			Expect(args1).To(Equal([]string{"-f", mpath}))
		})
		It("should succeed to Cleanup mpath if the device not exist", func() {
			mpath := "mpath"
			fakeExec.StatReturns(nil, cmdErr)
			fakeExec.IsNotExistReturns(true)
			err = bdUtils.Cleanup(mpath)
			Expect(err).ToNot(HaveOccurred())
			Expect(fakeExec.ExecuteCallCount()).To(Equal(0))
		})
		It("should fail to Cleanup mpath if the device state error (rather then not exist)", func() {
			mpath := "mpath"
			fakeExec.StatReturns(nil, cmdErr)
			fakeExec.IsNotExistReturns(false)
			err = bdUtils.Cleanup(mpath)
			Expect(err).To(HaveOccurred())
			Expect(fakeExec.ExecuteCallCount()).To(Equal(0))
			Expect(fakeExec.IsExecutableCallCount()).To(Equal(0))
		})

		It("Cleanup fails if dmsetup command missing", func() {
			mpath := "mpath"
			fakeExec.IsExecutableReturns(cmdErr)
			err = bdUtils.Cleanup(mpath)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(MatchRegexp(cmdErr.Error()))
		})
		It("Cleanup fails if dmsetup command fails", func() {
			mpath := "mpath"
			fakeExec.ExecuteWithTimeoutReturns([]byte{}, cmdErr)
			err = bdUtils.Cleanup(mpath)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(MatchRegexp(cmdErr.Error()))
		})
		It("Cleanup fails if multipath command missing", func() {
			mpath := "/dev/mapper/mpath"
			fakeExec.IsExecutableReturnsOnCall(0, cmdErr)
			err = bdUtils.Cleanup(mpath)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(MatchRegexp(cmdErr.Error()))
			Expect(fakeExec.IsExecutableCallCount()).To(Equal(1))
		})
		It("Cleanup fails if multipath command timeout exceeds", func() {
			mpath := "mpath"
			fakeExec.ExecuteWithTimeoutReturnsOnCall(0, []byte{}, context.DeadlineExceeded)
			err = bdUtils.Cleanup(mpath)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(MatchRegexp(context.DeadlineExceeded.Error()))
		})
	})
	Context(".CheckFs", func() {
		It("CheckFs detects exiting filesystem on device", func() {
			mpath := "mpath"
			fakeExec.ExecuteReturns([]byte{}, nil)
			fs, err := bdUtils.CheckFs(mpath)
			Expect(err).ToNot(HaveOccurred())
			Expect(fs).To(Equal(false))
			Expect(fakeExec.ExecuteCallCount()).To(Equal(1))
			cmd, args := fakeExec.ExecuteArgsForCall(0)
			Expect(cmd).To(Equal("blkid"))
			Expect(args).To(Equal([]string{mpath}))
		})
		It("CheckFs detects empty device", func() {
			err = ioutil.WriteFile("/tmp/tst.sh", []byte("exit 2"), 0777)
			Expect(err).ToNot(HaveOccurred())
			executor := utils.NewExecutor()
			_, exitErr2 := executor.Execute("sh", []string{"/tmp/tst.sh"})
			Expect(exitErr2).To(HaveOccurred())
			mpath := "mpath"
			fakeExec.ExecuteReturns([]byte{}, exitErr2)
			fs, err := bdUtils.CheckFs(mpath)
			Expect(err).ToNot(HaveOccurred())
			Expect(fs).To(Equal(true))
			Expect(fakeExec.ExecuteCallCount()).To(Equal(1))
			cmd, args := fakeExec.ExecuteArgsForCall(0)
			Expect(cmd).To(Equal("blkid"))
			Expect(args).To(Equal([]string{mpath}))
		})
		It("CheckFs fails if blkid missing", func() {
			mpath := "mpath"
			fakeExec.IsExecutableReturns(cmdErr)
			_, err = bdUtils.CheckFs(mpath)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(MatchRegexp(cmdErr.Error()))
		})
		It("CheckFs fails if blkid fails", func() {
			mpath := "mpath"
			fakeExec.ExecuteReturns([]byte{}, cmdErr)
			_, err := bdUtils.CheckFs(mpath)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(MatchRegexp(cmdErr.Error()))
		})
	})
	Context(".MakeFs", func() {
		It("MakeFs creates fs by type", func() {
			mpath := "mpath"
			fstype := "fstype"
			err = bdUtils.MakeFs(mpath, fstype)
			Expect(err).To(Not(HaveOccurred()))
			Expect(fakeExec.ExecuteCallCount()).To(Equal(1))
			cmd, args := fakeExec.ExecuteArgsForCall(0)
			Expect(cmd).To(Equal("mkfs"))
			Expect(args).To(Equal([]string{"-t", fstype, mpath}))
		})
		It("MakeFs fails if mkfs missing", func() {
			mpath := "mpath"
			fakeExec.IsExecutableReturns(cmdErr)
			err = bdUtils.MakeFs(mpath, "")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(MatchRegexp(cmdErr.Error()))
		})
		It("MakeFs fails if mkfs command fails", func() {
			mpath := "mpath"
			fakeExec.ExecuteReturns([]byte{}, cmdErr)
			err = bdUtils.MakeFs(mpath, "")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(MatchRegexp(cmdErr.Error()))
		})
	})
	Context(".MountFs", func() {
		It("MountFs succeeds", func() {
			mpath := "mpath"
			mpoint := "mpoint"
			err = bdUtils.MountFs(mpath, mpoint)
			Expect(err).To(Not(HaveOccurred()))
			Expect(fakeExec.ExecuteWithTimeoutCallCount()).To(Equal(2))
			_, cmd, args := fakeExec.ExecuteWithTimeoutArgsForCall(0)
			Expect(cmd).To(Equal("mount"))
			Expect(args).To(Equal([]string{mpath, mpoint}))

			_, cmd, args = fakeExec.ExecuteWithTimeoutArgsForCall(1)
			Expect(cmd).To(Equal("chmod"))
			Expect(args).To(Equal([]string{"775", mpoint}))
		})
		It("MountFs fails if mount command missing", func() {
			mpath := "mpath"
			mpoint := "mpoint"
			fakeExec.IsExecutableReturns(cmdErr)
			err = bdUtils.MountFs(mpath, mpoint)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(MatchRegexp(cmdErr.Error()))
		})
		It("MountFs fails if mount command fails", func() {
			mpath := "mpath"
			mpoint := "mpoint"
			fakeExec.ExecuteWithTimeoutReturns([]byte{}, cmdErr)
			err = bdUtils.MountFs(mpath, mpoint)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(MatchRegexp(cmdErr.Error()))
		})
		It("MountFs fails if chmod command fails", func() {
			mpath := "mpath"
			mpoint := "mpoint"
			fakeExec.ExecuteWithTimeoutReturnsOnCall(0, []byte{}, nil)
			fakeExec.ExecuteWithTimeoutReturnsOnCall(1, []byte{}, cmdErr)
			err = bdUtils.MountFs(mpath, mpoint)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(MatchRegexp(cmdErr.Error()))
		})
	})
	Context(".IsDeviceMounted", func() {
		It("should fail if mount command missing", func() {
			mpoint := "mpoint"
			fakeExec.IsExecutableReturns(cmdErr)
			isMounted, mounts, err := bdUtils.IsDeviceMounted(mpoint)
			Expect(err).To(HaveOccurred())
			Expect(isMounted).To(Equal(false))
			Expect(len(mounts)).To(Equal(0))
		})
		It("should fail if mount command fail", func() {
			mpoint := "mpoint"
			fakeExec.ExecuteWithTimeoutReturns([]byte{}, cmdErr)
			isMounted, mounts, err := bdUtils.IsDeviceMounted(mpoint)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(MatchRegexp(cmdErr.Error()))
			Expect(isMounted).To(Equal(false))
			Expect(len(mounts)).To(Equal(0))
		})
		It("should return false if device not found in mount output", func() {
			mpoint := "mpoint"
			mountOutput := `
/mpoint on /ubiquity/mpoint type ext4 (rw,relatime,data=ordered)
/dev/mapper/mpoint on /ubiquity/mpoint type ext4 (rw,relatime,data=ordered)
`
			fakeExec.ExecuteWithTimeoutReturns([]byte(mountOutput), nil)
			isMounted, mounts, err := bdUtils.IsDeviceMounted(mpoint)
			Expect(err).ToNot(HaveOccurred())
			Expect(isMounted).To(Equal(false))
			Expect(len(mounts)).To(Equal(0))
		})
		It("should return false if format of mount output is wrong", func() {
			mpoint := "mpoint"
			mountOutput := `
/mpoint on /ubiquity/mpoint type ext4 (rw,relatime,data=ordered)
/dev/mapper/mpoint on /ubiquity/mpoint type ext4 (rw,relatime,data=ordered)
`
			fakeExec.ExecuteWithTimeoutReturns([]byte(mountOutput), nil)
			isMounted, mounts, err := bdUtils.IsDeviceMounted(mpoint)
			Expect(err).ToNot(HaveOccurred())
			Expect(isMounted).To(Equal(false))
			Expect(len(mounts)).To(Equal(0))
		})

		It("should return true if device found in mount output", func() {
			mpoint := "mpoint"
			mountOutput := `
mpoint on /ubiquity/mpoint type ext4 (rw,relatime,data=ordered)
/dev/mapper/mpoint on /ubiquity/mpoint type ext4 (rw,relatime,data=ordered)
`
			fakeExec.ExecuteWithTimeoutReturns([]byte(mountOutput), nil)
			isMounted, mounts, err := bdUtils.IsDeviceMounted(mpoint)
			Expect(err).ToNot(HaveOccurred())
			Expect(isMounted).To(Equal(true))
			Expect(len(mounts)).To(Equal(1))
			Expect(mounts[0]).To(Equal("/ubiquity/mpoint"))
		})
		It("should return true if device found in mount output (2 mountpoints)", func() {
			mpoint := "mpoint"
			mountOutput := `
mpoint on /ubiquity/mpoint type ext4 (rw,relatime,data=ordered)
/dev/mapper/mpoint on /ubiquity/mpoint type ext4 (rw,relatime,data=ordered)
mpoint on /ubiquity/mpointSecond type ext4 (rw,relatime,data=ordered)
`
			fakeExec.ExecuteWithTimeoutReturns([]byte(mountOutput), nil)
			isMounted, mounts, err := bdUtils.IsDeviceMounted(mpoint)
			Expect(err).ToNot(HaveOccurred())
			Expect(isMounted).To(Equal(true))
			Expect(len(mounts)).To(Equal(2))
			Expect(mounts[0]).To(Equal("/ubiquity/mpoint"))
			Expect(mounts[1]).To(Equal("/ubiquity/mpointSecond"))
		})
	})
	Context(".IsDirIsAMountPoint", func() {
		It("should return false if DIR not found in mount output", func() {
			mpoint := "/wrong/wwn" // DIR
			mountOutput := `
/mpoint on /ubiquity/wwn1 type ext4 (rw,relatime,data=ordered)
/dev/mapper/mpoint on /ubiquity/mpoint type ext4 (rw,relatime,data=ordered)
`
			fakeExec.ExecuteWithTimeoutReturns([]byte(mountOutput), nil)
			isMounted, mounts, err := bdUtils.IsDirAMountPoint(mpoint)
			Expect(err).ToNot(HaveOccurred())
			Expect(isMounted).To(Equal(false))
			Expect(len(mounts)).To(Equal(0))
		})
		It("should return false if format of mount output is wrong", func() {
			mpoint := "/ubiquity/wwn1"
			mountOutput := `
wrong format on /ubiquity/mpoint type ext4 (rw,relatime,data=ordered)
/dev/mapper/mpoint on /ubiquity/mpoint type ext4 (rw,relatime,data=ordered)
`
			fakeExec.ExecuteWithTimeoutReturns([]byte(mountOutput), nil)
			isMounted, mounts, err := bdUtils.IsDirAMountPoint(mpoint)
			Expect(err).ToNot(HaveOccurred())
			Expect(isMounted).To(Equal(false))
			Expect(len(mounts)).To(Equal(0))
		})

		It("should return true if DIR found in mount output", func() {
			mpoint := "/ubiquity/wwn1"
			mountOutput := `
/fakedevice1 on /ubiquity/wwn1 type ext4 (rw,relatime,data=ordered)
/dev/mapper/mpoint on /ubiquity/mpoint type ext4 (rw,relatime,data=ordered)
`
			fakeExec.ExecuteWithTimeoutReturns([]byte(mountOutput), nil)
			isMounted, mounts, err := bdUtils.IsDirAMountPoint(mpoint)
			Expect(err).ToNot(HaveOccurred())
			Expect(isMounted).To(Equal(true))
			Expect(len(mounts)).To(Equal(1))
			Expect(mounts[0]).To(Equal("/fakedevice1"))
		})
		It("should return true if DIR found in mount output (2 devices to the same mountpoint)", func() {
			mpoint := "/ubiquity/wwn1"
			mountOutput := `
/fakedevice1 on /ubiquity/wwn1 type ext4 (rw,relatime,data=ordered)
/dev/mapper/mpoint on /ubiquity/mpoint type ext4 (rw,relatime,data=ordered)
/fakedevice2 on /ubiquity/wwn1 type ext4 (rw,relatime,data=ordered)
`
			fakeExec.ExecuteWithTimeoutReturns([]byte(mountOutput), nil)
			isMounted, mounts, err := bdUtils.IsDirAMountPoint(mpoint)
			Expect(err).ToNot(HaveOccurred())
			Expect(isMounted).To(Equal(true))
			Expect(len(mounts)).To(Equal(2))
			Expect(mounts[0]).To(Equal("/fakedevice1"))
			Expect(mounts[1]).To(Equal("/fakedevice2"))
		})
	})

	Context(".UmountFs", func() {
		It("UmountFs succeeds", func() {
			mpoint := "/dev/mapper/mpoint"
			fakeExec.ExecuteReturnsOnCall(0, nil, nil) // the umount command
			err = bdUtils.UmountFs(mpoint, "6001738CFC9035EA0000000000795164")
			Expect(err).To(Not(HaveOccurred()))
			Expect(fakeExec.ExecuteWithTimeoutCallCount()).To(Equal(1))

			_, cmd, args := fakeExec.ExecuteWithTimeoutArgsForCall(0)
			Expect(cmd).To(Equal("umount"))
			Expect(args).To(Equal([]string{mpoint}))
		})
		It("should succeed to UmountFs if mpath is already unmounted", func() {
			mpoint := "/dev/mapper/mpoint"
			mountOutput := `
/XXX/mpoint on /ubiquity/mpoint type ext4 (rw,relatime,data=ordered)
/dev/mapper/yyy on /ubiquity/yyy type ext4 (rw,relatime,data=ordered)
`
			fakeExec.ExecuteWithTimeoutReturnsOnCall(0, nil, cmdErr)              // the umount command should fail
			fakeExec.ExecuteWithTimeoutReturnsOnCall(1, []byte(mountOutput), nil) // mount for isMounted
			err = bdUtils.UmountFs(mpoint, "6001738CFC9035EA0000000000795164")
			Expect(err).To(Not(HaveOccurred()))
			Expect(fakeExec.ExecuteWithTimeoutCallCount()).To(Equal(2))
			_, cmd, _ := fakeExec.ExecuteWithTimeoutArgsForCall(0)
			Expect(cmd).To(Equal("umount")) // first check is the umount
			_, cmd, _ = fakeExec.ExecuteWithTimeoutArgsForCall(1)
			Expect(cmd).To(Equal("mount")) // second check is the umount
		})
		It("UmountFs fails if umount command missing", func() {
			mpoint := "/dev/mapper/mpoint"
			fakeExec.IsExecutableReturns(cmdErr)
			err = bdUtils.UmountFs(mpoint, "6001738CFC9035EA0000000000795164")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(MatchRegexp(cmdErr.Error()))
			Expect(fakeExec.ExecuteCallCount()).To(Equal(0))
		})
		It("UmountFs fails if umount command fails", func() {
			mpoint := "mpoint"
			fakeExec.ExecuteReturns([]byte{}, cmdErr)
			fakeExec.ExecuteWithTimeoutReturns([]byte{}, cmdErr)
			err = bdUtils.UmountFs(mpoint, "6001738CFC9035EA0000000000795164")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(MatchRegexp(cmdErr.Error()))
		})
		It("UmountFs fails if umount command fails", func() {
			mpoint := "mpoint"
			//fakeExec.ExecuteReturns([]byte{}, cmdErr)
			fakeExec.ExecuteWithTimeoutReturns([]byte{}, context.DeadlineExceeded)
			err = bdUtils.UmountFs(mpoint, "6001738CFC9035EA0000000000795164")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(MatchRegexp(context.DeadlineExceeded.Error()))
		})
	})
})

func TestGetBlockDeviceUtils(t *testing.T) {
	RegisterFailHandler(Fail)
	defer utils.InitUbiquityServerTestLogger()()
	RunSpecs(t, "BlockDeviceUtils Test Suite")
}
