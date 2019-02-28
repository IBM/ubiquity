package connectors_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/IBM/ubiquity/fakes"
	"github.com/IBM/ubiquity/remote/mounter/initiator"
	"github.com/IBM/ubiquity/remote/mounter/initiator/connectors"
	fakeinitiator "github.com/IBM/ubiquity/remote/mounter/initiator/fakes"
	"github.com/IBM/ubiquity/resources"
)

var fakeMpathOutputForLUN255 = `
  |- 0:0:4:255 sda 8:0   active ready running
  |- 0:0:5:255 sdb 8:16  active ready running
  |- 0:0:6:255 sdc 8:32  active ready running
  |- 0:0:7:255 sdd 8:48  active ready running
  |- 1:0:4:255 sde 8:64  active ready running
  |- 1:0:5:255 sdf 8:80  active ready running
  |- 1:0:6:255 sdg 8:96  active ready running
  ` + "`- 1:0:7:255 sdh 8:112 active ready running\n"

var fakeMpathOutputForLUN0 = `
  |- 0:0:4:0 sda 8:0   active ready running
  |- 0:0:5:0 sdb 8:16  active ready running
  |- 0:0:6:0 sdc 8:32  active ready running
  |- 0:0:7:0 sdd 8:48  active ready running
  |- 1:0:4:0 sde 8:64  active ready running
  |- 1:0:5:0 sdf 8:80  active ready running
  |- 1:0:6:0 sdg 8:96  active ready running
  ` + "`- 1:0:7:0 sdh 8:112 active ready running\n"

var fakeMpathOutputForLUNZlinux = `
  |- 0:0:4:1076248592 sda 8:0   active ready running
  |- 0:0:5:1076248592 sdb 8:16  active ready running
  |- 0:0:6:1076248592 sdc 8:32  active ready running
  |- 0:0:7:1076248592 sdd 8:48  active ready running
  |- 1:0:4:1076248592 sde 8:64  active ready running
  |- 1:0:5:1076248592 sdf 8:80  active ready running
  |- 1:0:6:1076248592 sdg 8:96  active ready running
  ` + "`- 1:0:7:1076248592 sdh 8:112 active ready running\n"

var _ = Describe("Test Fibre Channel Connector", func() {
	var (
		fakeExec      *fakes.FakeExecutor
		fakeInitiator *fakeinitiator.FakeInitiator
		fcConnector   initiator.Connector
	)
	volumeMountProperties := &resources.VolumeMountProperties{WWN: "wwn", LunNumber: float64(1)}

	BeforeEach(func() {
		fakeExec = new(fakes.FakeExecutor)
		fakeInitiator = new(fakeinitiator.FakeInitiator)
		fcConnector = connectors.NewFibreChannelConnectorWithAllFields(fakeExec, fakeInitiator)
	})

	Context("ConnectVolume", func() {

		BeforeEach(func() {
			fakeInitiator.GetHBAsReturns([]string{"host0"})
		})

		It("should rescan all host HBAs", func() {
			err := fcConnector.ConnectVolume(volumeMountProperties)
			Ω(err).ShouldNot(HaveOccurred())
			Expect(fakeInitiator.RescanHostsCallCount()).To(Equal(1))
		})
	})

	Context("DisconnectVolume", func() {

		Context("remove lun 255", func() {

			BeforeEach(func() {
				fakeExec.ExecuteReturns([]byte(fakeMpathOutputForLUN255), nil)
			})

			It("should remove all the scsi devices", func() {
				err := fcConnector.DisconnectVolume(volumeMountProperties)
				Ω(err).ShouldNot(HaveOccurred())
				cmd, args := fakeExec.ExecuteArgsForCall(0)
				Expect(cmd).To(Equal("multipath"))
				Expect(args).To(Equal([]string{"-ll", "|", fmt.Sprintf(`egrep "[0-9]+:[0-9]+:[0-9]+:%g "`, volumeMountProperties.LunNumber)}))

				Expect(fakeInitiator.RemoveSCSIDeviceCallCount()).To(Equal(8))
				var a byte = 97
				for i := 0; i < 8; i++ {
					expectDev := "/dev/sd" + string(a+byte(i))
					dev := fakeInitiator.RemoveSCSIDeviceArgsForCall(i)
					Expect(dev).To(Equal(expectDev))
				}
			})
		})

		Context("remove lun 0", func() {

			BeforeEach(func() {
				fakeExec.ExecuteReturns([]byte(fakeMpathOutputForLUN0), nil)
			})

			It("should remove all the scsi devices", func() {
				err := fcConnector.DisconnectVolume(volumeMountProperties)
				Ω(err).ShouldNot(HaveOccurred())
				cmd, args := fakeExec.ExecuteArgsForCall(0)
				Expect(cmd).To(Equal("multipath"))
				Expect(args).To(Equal([]string{"-ll", "|", fmt.Sprintf(`egrep "[0-9]+:[0-9]+:[0-9]+:%g "`, volumeMountProperties.LunNumber)}))

				Expect(fakeInitiator.RemoveSCSIDeviceCallCount()).To(Equal(8))
				var a byte = 97
				for i := 0; i < 8; i++ {
					expectDev := "/dev/sd" + string(a+byte(i))
					dev := fakeInitiator.RemoveSCSIDeviceArgsForCall(i)
					Expect(dev).To(Equal(expectDev))
				}
			})
		})

		Context("remove zlinux lun", func() {

			BeforeEach(func() {
				fakeExec.ExecuteReturns([]byte(fakeMpathOutputForLUNZlinux), nil)
			})

			It("should remove all the scsi devices", func() {
				err := fcConnector.DisconnectVolume(volumeMountProperties)
				Ω(err).ShouldNot(HaveOccurred())
				cmd, args := fakeExec.ExecuteArgsForCall(0)
				Expect(cmd).To(Equal("multipath"))
				Expect(args).To(Equal([]string{"-ll", "|", fmt.Sprintf(`egrep "[0-9]+:[0-9]+:[0-9]+:%g "`, volumeMountProperties.LunNumber)}))

				Expect(fakeInitiator.RemoveSCSIDeviceCallCount()).To(Equal(8))
				var a byte = 97
				for i := 0; i < 8; i++ {
					expectDev := "/dev/sd" + string(a+byte(i))
					dev := fakeInitiator.RemoveSCSIDeviceArgsForCall(i)
					Expect(dev).To(Equal(expectDev))
				}
			})
		})
	})
})
