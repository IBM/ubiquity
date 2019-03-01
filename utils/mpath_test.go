package utils_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/IBM/ubiquity/fakes"
	"github.com/IBM/ubiquity/utils"
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

var fakeWwn = "6005076306ffd69d0000000000001004"

var fakeMultipathOutput = `
mpathg (36005076306ffd69d0000000000001004) dm-14 IBM     ,2107900
size=1.0G features='1 queue_if_no_path' hwhandler='0' wp=rw
` + "`-+- policy='service-time 0' prio=1 status=active" + `
  |- 29:0:1:1 sde 8:64 active ready running
  |- 29:0:6:1 sdf 8:80 active ready running
  ` + "`- 29:0:7:1 sdg 8:96 active ready running" + `
mpathf (36005076306ffd69d000000000000010a) dm-2 IBM     ,2107900
size=2.0G features='1 queue_if_no_path' hwhandler='0' wp=rw
` + "`-+- policy='service-time 0' prio=1 status=enabled" + `
  |- 29:0:1:0 sdb 8:16 active ready running
  |- 29:0:6:0 sdc 8:32 active ready running
  ` + "`- 29:0:7:0 sdd 8:48 active ready running\n"

var _ = Describe("scbe_mounter_test", func() {
	var (
		fakeExec *fakes.FakeExecutor
	)

	BeforeEach(func() {
		fakeExec = new(fakes.FakeExecutor)
	})

	Context("GetMultipathOutputAndDeviceMapperAndDevice", func() {

		BeforeEach(func() {
			fakeExec.ExecuteWithTimeoutReturns([]byte(fakeMultipathOutput), nil)
		})

		It("should call DisconnectAll ", func() {
			_, devMapper, devNames, err := utils.GetMultipathOutputAndDeviceMapperAndDevice(fakeWwn, fakeExec)
			Ω(err).ShouldNot(HaveOccurred())
			Expect(devMapper).To(Equal("mpathg"))
			Expect(devNames).To(Equal([]string{"sde", "sdf", "sdg"}))
		})
	})
})
