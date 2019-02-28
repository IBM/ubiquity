package initiator_test

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/IBM/ubiquity/fakes"
	"github.com/IBM/ubiquity/remote/mounter/initiator"
	"github.com/IBM/ubiquity/resources"
)

const tmp_prefix = "/tmp/ubiquity_test"

const FAKE_FC_HOST_SYSFS_PATH = tmp_prefix + "/sys/class/fc_host"
const FAKE_SCSI_HOST_SYSFS_PATH = tmp_prefix + "/sys/class/scsi_host"
const FAKE_SYS_BLOCK_PATH = tmp_prefix + "/sys/block"

var fakeSystoolOutput = `Class = "fc_host"

  Class Device = "host0"
  Class Device path = "/sys/devices/css0/0.0.0000/0.0.a200/host0/fc_host/host0"
    active_fc4s         = "0x00 0x00 0x01 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 "
    dev_loss_tmo        = "60"
    maxframe_size       = "2112 bytes"
    node_name           = "0x5005076400c1b2b8"
    permanent_port_name = "0xc05076e5118011d1"
    port_id             = "0x0ecfd3"
    port_name           = "0xc05076e511803cb0"
    port_state          = "Online"
    port_type           = "NPIV VPORT"
    serial_number       = "IBM0200000001B2B8"
    speed               = "8 Gbit"
    supported_classes   = "Class 2, Class 3"
    supported_fc4s      = "0x00 0x00 0x01 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 "
    supported_speeds    = "4 Gbit, 8 Gbit, 16 Gbit"
    symbolic_name       = "IBM     3906            0200000001B2B8  PCHID: 011D NPIV UlpId: 004C0308   DEVNO: 0.0.a200 NAME: stk8s008"
    tgtid_bind_type     = "wwpn (World Wide Port Name)"
    uevent              =

    Device = "host0"
    Device path = "/sys/devices/css0/0.0.0000/0.0.a200/host0"
      uevent              = "DEVTYPE=scsi_host"


  Class Device = "host1"
  Class Device path = "/sys/devices/css0/0.0.0001/0.0.a300/host1/fc_host/host1"
    active_fc4s         = "0x00 0x00 0x01 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 "
    dev_loss_tmo        = "60"
    maxframe_size       = "2112 bytes"
    node_name           = "0x5005076400c1b2b8"
    permanent_port_name = "0xc05076e511801811"
    port_id             = "0x0fc613"
    port_name           = "0xc05076e511803b48"
    port_state          = "Online"
    port_type           = "NPIV VPORT"
    serial_number       = "IBM0200000001B2B8"
    speed               = "8 Gbit"
    supported_classes   = "Class 2, Class 3"
    supported_fc4s      = "0x00 0x00 0x01 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 "
    supported_speeds    = "4 Gbit, 8 Gbit, 16 Gbit"
    symbolic_name       = "IBM     3906            0200000001B2B8  PCHID: 0181 NPIV UlpId: 00450308   DEVNO: 0.0.a300 NAME: stk8s008"
    tgtid_bind_type     = "wwpn (World Wide Port Name)"
    uevent              =

    Device = "host1"
    Device path = "/sys/devices/css0/0.0.0001/0.0.a300/host1"
      uevent              = "DEVTYPE=scsi_host"


`

var _ = Describe("Test FC Initiator", func() {
	var (
		fakeExec         *fakes.FakeExecutor
		fcInitiator      initiator.Initiator
		realFcSysPath    string
		realScsiSysPath  string
		realSysBlockPath string
		cmdErr           error = errors.New("command error")
	)
	volumeMountProperties := &resources.VolumeMountProperties{WWN: "wwn", LunNumber: float64(1)}

	BeforeEach(func() {
		err := os.MkdirAll(FAKE_FC_HOST_SYSFS_PATH, os.ModePerm)
		Ω(err).ShouldNot(HaveOccurred())

		err = os.MkdirAll(FAKE_SCSI_HOST_SYSFS_PATH, os.ModePerm)
		Ω(err).ShouldNot(HaveOccurred())

		err = os.MkdirAll(FAKE_SYS_BLOCK_PATH, os.ModePerm)
		Ω(err).ShouldNot(HaveOccurred())

		realFcSysPath = initiator.FC_HOST_SYSFS_PATH
		realScsiSysPath = initiator.SCSI_HOST_SYSFS_PATH
		realSysBlockPath = initiator.SYS_BLOCK_PATH
		initiator.FC_HOST_SYSFS_PATH = FAKE_FC_HOST_SYSFS_PATH
		initiator.SCSI_HOST_SYSFS_PATH = FAKE_SCSI_HOST_SYSFS_PATH
		initiator.SYS_BLOCK_PATH = FAKE_SYS_BLOCK_PATH

		fakeExec = new(fakes.FakeExecutor)
		fcInitiator = initiator.NewLinuxFibreChannelWithExecutor(fakeExec)
	})

	AfterEach(func() {
		initiator.FC_HOST_SYSFS_PATH = realFcSysPath
		initiator.SCSI_HOST_SYSFS_PATH = realScsiSysPath
		initiator.SYS_BLOCK_PATH = realSysBlockPath

		err := os.RemoveAll(tmp_prefix)
		Ω(err).ShouldNot(HaveOccurred())
	})

	Context("GetHBAs", func() {

		Context("get HBAs from systool", func() {
			BeforeEach(func() {
				fakeExec.ExecuteReturns([]byte(fakeSystoolOutput), nil)
			})

			It("should get from systool if it is installed", func() {
				hbas := fcInitiator.GetHBAs()
				Expect(hbas).To(Equal([]string{"host0", "host1"}))
			})
		})

		Context("get HBAs from sys fc path", func() {
			hbas := []string{"host0", "host1", "host2"}

			BeforeEach(func() {
				for _, hba := range hbas {
					fullPath := FAKE_FC_HOST_SYSFS_PATH + "/" + hba
					err := os.MkdirAll(fullPath, os.ModePerm)
					Ω(err).ShouldNot(HaveOccurred())
				}
			})

			It("should get from sys fc path if systool is not installed", func() {
				fakeExec.IsExecutableReturns(cmdErr)
				hbasRes := fcInitiator.GetHBAs()
				Expect(hbasRes).To(Equal(hbas))
			})

			It("should get from sys fc path if systool returns error", func() {
				fakeExec.ExecuteReturns([]byte{}, cmdErr)
				hbasRes := fcInitiator.GetHBAs()
				Expect(hbasRes).To(Equal(hbas))
			})
		})

	})

	Context("RescanHosts", func() {
		var hbas = []string{"host0"}
		var scanPath = FAKE_SCSI_HOST_SYSFS_PATH + "/" + hbas[0]
		var scanFile = scanPath + "/scan"

		BeforeEach(func() {
			err := os.MkdirAll(scanPath, os.ModePerm)
			Ω(err).ShouldNot(HaveOccurred())
			_, err = os.Create(scanFile)
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("should write '- - lunid' to the hba scan file", func() {
			err := fcInitiator.RescanHosts(hbas, volumeMountProperties)
			Ω(err).ShouldNot(HaveOccurred())
			data, err := ioutil.ReadFile(scanFile)
			Ω(err).ShouldNot(HaveOccurred())
			Expect(string(data)).To(Equal(fmt.Sprintf("- - %g", volumeMountProperties.LunNumber)))
		})
	})

	Context("RemoveSCSIDevice", func() {
		var devName = "sda"
		var dev = "/dev/" + devName
		var deletePath = FAKE_SYS_BLOCK_PATH + fmt.Sprintf("/%s/device", devName)
		var deleteFile = deletePath + "/delete"

		BeforeEach(func() {
			err := os.MkdirAll(deletePath, os.ModePerm)
			Ω(err).ShouldNot(HaveOccurred())
			_, err = os.Create(deleteFile)
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("should write 1 to the device delete file", func() {
			err := fcInitiator.RemoveSCSIDevice(dev)
			Ω(err).ShouldNot(HaveOccurred())
			data, err := ioutil.ReadFile(deleteFile)
			Ω(err).ShouldNot(HaveOccurred())
			Expect(string(data)).To(Equal("1"))
		})
	})
})
