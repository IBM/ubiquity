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
    "github.com/IBM/ubiquity/remote/mounter/block_device_utils"
    "github.com/IBM/ubiquity/fakes"
    "github.com/IBM/ubiquity/utils"
    "github.com/IBM/ubiquity/utils/logs"
    . "github.com/onsi/gomega"
    . "github.com/onsi/ginkgo"
    "testing"
    "errors"
    "fmt"
    "io/ioutil"
)

var _ = Describe("block_device_utils_test", func() {
    var (
        fakeExec      *fakes.FakeExecutor
        bdUtils       block_device_utils.BlockDeviceUtils
        err           error
        cmdErr        error = errors.New("command error")
    )

    BeforeEach(func() {
        fakeExec = new(fakes.FakeExecutor)
        bdUtils = block_device_utils.NewBlockDeviceUtilsWithExecutor(fakeExec)
    })

    Context(".Rescan", func() {
        It("Rescan ISCSI calls 'sudo iscsiadm -m session --rescan'", func() {
            err = bdUtils.Rescan(block_device_utils.ISCSI)
            Expect(err).ToNot(HaveOccurred())
            Expect(fakeExec.ExecuteCallCount()).To(Equal(1))
            cmd, args := fakeExec.ExecuteArgsForCall(0)
            Expect(cmd).To(Equal("sudo"))
            Expect(args).To(Equal([]string{"iscsiadm", "-m", "session", "--rescan"}))
        })
        It("Rescan SCSI calls 'sudo rescan-scsi-bus -r'", func() {
            err = bdUtils.Rescan(block_device_utils.SCSI)
            Expect(err).ToNot(HaveOccurred())
            Expect(fakeExec.ExecuteCallCount()).To(Equal(1))
            cmd, args := fakeExec.ExecuteArgsForCall(0)
            Expect(cmd).To(Equal("sudo"))
            Expect(args).To(Equal([]string{"rescan-scsi-bus", "-r"}))
        })
        /*
        It("Rescan ISCSI fails if iscsiadm command missing", func() {
            fakeExec.IsExecutableReturns(cmdErr)
            err = bdUtils.Rescan(block_device_utils.ISCSI)
            Expect(err).To(HaveOccurred())
            Expect(err.Error()).To(MatchRegexp(cmdErr.Error()))
            Expect(fakeExec.ExecuteCallCount()).To(Equal(0))
            Expect(fakeExec.IsExecutableCallCount()).To(Equal(1))
            Expect(fakeExec.IsExecutableArgsForCall(0)).To(Equal("iscsiadm"))
        })
        */
        It("Rescan SCSI fails if rescan-scsi-bus command missing", func() {
            fakeExec.IsExecutableReturns(cmdErr)
            err = bdUtils.Rescan(block_device_utils.SCSI)
            Expect(err).To(HaveOccurred())
            Expect(fakeExec.ExecuteCallCount()).To(Equal(0))
            Expect(fakeExec.IsExecutableCallCount()).To(Equal(2))
            Expect(fakeExec.IsExecutableArgsForCall(0)).To(Equal("rescan-scsi-bus"))
            Expect(fakeExec.IsExecutableArgsForCall(1)).To(Equal("rescan-scsi-bus.sh"))
        })
        It("Rescan ISCSI fails if iscsiadm execution fails", func() {
            fakeExec.ExecuteReturns([]byte{}, cmdErr)
            err = bdUtils.Rescan(block_device_utils.ISCSI)
            Expect(err.Error()).To(MatchRegexp(cmdErr.Error()))
        })
        It("Rescan SCSI fails if rescan-scsi-bus execution fails", func() {
            fakeExec.ExecuteReturns([]byte{}, cmdErr)
            err = bdUtils.Rescan(block_device_utils.SCSI)
            Expect(err.Error()).To(MatchRegexp(cmdErr.Error()))
        })
    })
    Context(".ReloadMultipath", func() {
        It("ReloadMultipath calls multipath command", func() {
            err = bdUtils.ReloadMultipath()
            Expect(err).ToNot(HaveOccurred())
            Expect(fakeExec.ExecuteCallCount()).To(Equal(1))
            cmd, args := fakeExec.ExecuteArgsForCall(0)
            Expect(cmd).To(Equal("sudo"))
            Expect(args).To(Equal([]string{"multipath", "-r"}))
        })
        It("ReloadMultipath fails if multipath command is missing", func() {
            fakeExec.IsExecutableReturns(cmdErr)
            err = bdUtils.ReloadMultipath()
            Expect(err).To(HaveOccurred())
            Expect(err.Error()).To(MatchRegexp(cmdErr.Error()))
        })
        It("ReloadMultipath fails if multipath command fails", func() {
            fakeExec.ExecuteReturns([]byte{}, cmdErr)
            err = bdUtils.ReloadMultipath()
            Expect(err).To(HaveOccurred())
            Expect(err.Error()).To(MatchRegexp(cmdErr.Error()))
        })
    })
    Context(".Discover", func() {
        It("Discover returns path for volume", func() {
            volumeId := "volume-id"
            result := "mpath"
            fakeExec.ExecuteReturns([]byte(fmt.Sprintf("%s (%s) dm-1", result, volumeId)), nil)
            mpath, err := bdUtils.Discover(volumeId)
            Expect(err).ToNot(HaveOccurred())
            Expect(mpath).To(Equal("/dev/mapper/" + result))
            Expect(fakeExec.ExecuteCallCount()).To(Equal(1))
            cmd, args := fakeExec.ExecuteArgsForCall(0)
            Expect(cmd).To(Equal("sudo"))
            Expect(args).To(Equal([]string{"multipath", "-ll"}))
        })
        It("Discover fails if multipath command is missing", func() {
            volumeId := "volume-id"
            fakeExec.IsExecutableReturns(cmdErr)
            _, err := bdUtils.Discover(volumeId)
            Expect(err).To(HaveOccurred())
            Expect(err.Error()).To(MatchRegexp(cmdErr.Error()))
        })
        It("Discover fails if multipath -ll command fails", func() {
            volumeId := "volume-id"
            fakeExec.ExecuteReturns([]byte{}, cmdErr)
            _, err := bdUtils.Discover(volumeId)
            Expect(err).To(HaveOccurred())
            Expect(err.Error()).To(MatchRegexp(cmdErr.Error()))
        })
        It("Discover fails if volume not found", func() {
            volumeId := "volume-id"
            fakeExec.ExecuteReturns([]byte(fmt.Sprintf(
                "mpath (other-volume-1) dm-1\nmpath (other-volume-2) dm-2")), nil)
            _, err := bdUtils.Discover(volumeId)
            Expect(err).To(HaveOccurred())
        })
    })
    Context(".Cleanup", func() {
        It("Cleanup calls dmsetup and multipath", func() {
            mpath := "mpath"
            err = bdUtils.Cleanup(mpath)
            Expect(err).ToNot(HaveOccurred())
            Expect(fakeExec.ExecuteCallCount()).To(Equal(2))
            cmd1, args1 := fakeExec.ExecuteArgsForCall(0)
            Expect(cmd1).To(Equal("sudo"))
            Expect(args1).To(Equal([]string{"dmsetup", "message", mpath, "0", "fail_if_no_path"}))
            cmd2, args2 := fakeExec.ExecuteArgsForCall(1)
            Expect(cmd2).To(Equal("sudo"))
            Expect(args2).To(Equal([]string{"multipath", "-f", mpath}))
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
            fakeExec.ExecuteReturns([]byte{}, cmdErr)
            err = bdUtils.Cleanup(mpath)
            Expect(err).To(HaveOccurred())
            Expect(err.Error()).To(MatchRegexp(cmdErr.Error()))
        })
        It("Cleanup fails if multipath command missing", func() {
            mpath := "/dev/mapper/mpath"
            fakeExec.IsExecutableReturnsOnCall(1, cmdErr)
            err = bdUtils.Cleanup(mpath)
            Expect(err).To(HaveOccurred())
            Expect(err.Error()).To(MatchRegexp(cmdErr.Error()))
            Expect(fakeExec.IsExecutableCallCount()).To(Equal(2))
        })
        It("Cleanup fails if multipath command fails", func() {
            mpath := "mpath"
            fakeExec.ExecuteReturnsOnCall(1, []byte{}, cmdErr)
            err = bdUtils.Cleanup(mpath)
            Expect(err).To(HaveOccurred())
            Expect(err.Error()).To(MatchRegexp(cmdErr.Error()))
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
            Expect(cmd).To(Equal("sudo"))
            Expect(args).To(Equal([]string{"blkid", mpath}))
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
            Expect(cmd).To(Equal("sudo"))
            Expect(args).To(Equal([]string{"blkid", mpath}))
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
            Expect(cmd).To(Equal("sudo"))
            Expect(args).To(Equal([]string{"mkfs", "-t", fstype, mpath}))
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
            Expect(fakeExec.ExecuteCallCount()).To(Equal(1))
            cmd, args := fakeExec.ExecuteArgsForCall(0)
            Expect(cmd).To(Equal("sudo"))
            Expect(args).To(Equal([]string{"mount", mpath, mpoint}))
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
            fakeExec.ExecuteReturns([]byte{}, cmdErr)
            err = bdUtils.MountFs(mpath, mpoint)
            Expect(err).To(HaveOccurred())
            Expect(err.Error()).To(MatchRegexp(cmdErr.Error()))
        })

    })
    Context(".UmountFs", func() {
        It("UmountFs succeeds", func() {
            mpoint := "mpoint"
            err = bdUtils.UmountFs(mpoint)
            Expect(err).To(Not(HaveOccurred()))
            Expect(fakeExec.ExecuteCallCount()).To(Equal(1))
            cmd, args := fakeExec.ExecuteArgsForCall(0)
            Expect(cmd).To(Equal("sudo"))
            Expect(args).To(Equal([]string{"umount", mpoint}))
        })
        It("UmountFs fails if umount command missing", func() {
            mpoint := "mpoint"
            fakeExec.IsExecutableReturns(cmdErr)
            err = bdUtils.UmountFs(mpoint)
            Expect(err).To(HaveOccurred())
            Expect(err.Error()).To(MatchRegexp(cmdErr.Error()))
        })
        It("UmountFs fails if umount command fails", func() {
            mpoint := "mpoint"
            fakeExec.ExecuteReturns([]byte{}, cmdErr)
            err = bdUtils.UmountFs(mpoint)
            Expect(err).To(HaveOccurred())
            Expect(err.Error()).To(MatchRegexp(cmdErr.Error()))
        })
    })
})


func TestGetBlockDeviceUtils(t *testing.T) {
    RegisterFailHandler(Fail)
    defer logs.InitStdoutLogger(logs.DEBUG)()
    RunSpecs(t, "BlockDeviceUtils Test Suite")
}
