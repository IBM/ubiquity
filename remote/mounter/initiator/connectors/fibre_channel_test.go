package connectors_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/IBM/ubiquity/fakes"
	"github.com/IBM/ubiquity/remote/mounter/initiator"
	"github.com/IBM/ubiquity/remote/mounter/initiator/connectors"
	fakeinitiator "github.com/IBM/ubiquity/remote/mounter/initiator/fakes"
	"github.com/IBM/ubiquity/resources"
)

var fakeWwn = "6005076306ffd69d0000000000001004"

var fakeMultipathOutput = `
mpathg (36005076306ffd69d0000000000001004) dm-14 IBM     ,2107900
size=1.0G features='1 queue_if_no_path' hwhandler='0' wp=rw
` + "`-+- policy='service-time 0' prio=1 status=active" + `
  |- 29:0:1:1 sda 8:64 active ready running
  |- 29:0:6:1 sdb 8:80 active ready running
  ` + "`- 29:0:7:1 sdc 8:96 active ready running" + `
mpathf (36005076306ffd69d000000000000010a) dm-2 IBM     ,2107900
size=2.0G features='1 queue_if_no_path' hwhandler='0' wp=rw
` + "`-+- policy='service-time 0' prio=1 status=enabled" + `
  |- 29:0:1:0 sdb 8:16 active ready running
  |- 29:0:6:0 sdc 8:32 active ready running
  ` + "`- 29:0:7:0 sdd 8:48 active ready running\n"

var _ = Describe("Test Fibre Channel Connector", func() {
	var (
		fakeExec      *fakes.FakeExecutor
		fakeInitiator *fakeinitiator.FakeInitiator
		fcConnector   initiator.Connector
	)
	volumeMountProperties := &resources.VolumeMountProperties{WWN: fakeWwn, LunNumber: float64(1)}

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

		BeforeEach(func() {
			fakeExec.ExecuteWithTimeoutReturns([]byte(fakeMultipathOutput), nil)
		})

		It("should call multipath and remove all the scsi devices", func() {
			err := fcConnector.DisconnectVolume(volumeMountProperties)
			Ω(err).ShouldNot(HaveOccurred())
			Expect(fakeExec.ExecuteWithTimeoutCallCount()).To(Equal(1))
			Expect(fakeInitiator.RemoveSCSIDeviceCallCount()).To(Equal(3))
			var a byte = 97
			for i := 0; i < 3; i++ {
				expectDev := "/dev/sd" + string(a+byte(i))
				dev := fakeInitiator.RemoveSCSIDeviceArgsForCall(i)
				Expect(dev).To(Equal(expectDev))
			}
		})

		It("should not call multipath and will remove all the scsi devices", func() {
			devNames := []string{"sda", "sdb"}

			volumeMountProperties.DeviceMapper = "test"
			volumeMountProperties.Devices = devNames
			err := fcConnector.DisconnectVolume(volumeMountProperties)
			Ω(err).ShouldNot(HaveOccurred())
			Expect(fakeExec.ExecuteWithTimeoutCallCount()).To(Equal(0))
			Expect(fakeInitiator.RemoveSCSIDeviceCallCount()).To(Equal(2))
			var a byte = 97
			for i := 0; i < 2; i++ {
				expectDev := "/dev/sd" + string(a+byte(i))
				dev := fakeInitiator.RemoveSCSIDeviceArgsForCall(i)
				Expect(dev).To(Equal(expectDev))
			}
		})
	})
})
