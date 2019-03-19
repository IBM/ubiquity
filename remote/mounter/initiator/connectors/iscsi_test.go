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

var _ = Describe("Test ISCSI Connector", func() {
	var (
		fakeExec       *fakes.FakeExecutor
		fakeInitiator  *fakeinitiator.FakeInitiator
		iscsiConnector initiator.Connector
	)
	volumeMountProperties := &resources.VolumeMountProperties{WWN: fakeWwn, LunNumber: 1}

	BeforeEach(func() {
		fakeExec = new(fakes.FakeExecutor)
		fakeInitiator = new(fakeinitiator.FakeInitiator)
		iscsiConnector = connectors.NewISCSIConnectorWithAllFields(fakeExec, fakeInitiator)
	})

	Context("DisconnectVolume", func() {

		BeforeEach(func() {
			fakeExec.ExecuteWithTimeoutReturns([]byte(fakeMultipathOutput), nil)
		})

		It("should call multipath and remove all the scsi devices", func() {
			err := iscsiConnector.DisconnectVolume(volumeMountProperties)
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
	})
})
