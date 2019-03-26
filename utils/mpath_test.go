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

var fakeMultipathOutputWithMultiplePathGroups = `
mpathc (6005076306ffd69d0000000000001004) dm-4 IBM     ,2145
size=1.0G features='0' hwhandler='0' wp=rw
|-+- policy='service-time 0' prio=50 status=active
| '- 43:0:0:3 sda 8:112 active ready running
-+- policy='service-time 0' prio=10 status=enabled
  '- 44:0:0:3 sdb 8:144 active ready running
mpathb (3600507680c87011598000000000013a7) dm-3 IBM     ,2145
size=1.0G features='1 queue_if_no_path' hwhandler='0' wp=rw
|-+- policy='service-time 0' prio=50 status=active
| '- 44:0:0:0 sdc 8:32  active ready running
'-+- policy='service-time 0' prio=10 status=enabled
  '- 43:0:0:0 sdb 8:16  active ready running
`

var _ = Describe("scbe_mounter_test", func() {
	var (
		fakeExec *fakes.FakeExecutor
	)

	BeforeEach(func() {
		fakeExec = new(fakes.FakeExecutor)
	})

	Context("GetMultipathOutputAndDeviceMapperAndDevice", func() {

		It("should get device names from multipath output", func() {
			fakeExec.ExecuteWithTimeoutReturns([]byte(fakeMultipathOutput), nil)
			_, devMapper, devNames, err := utils.GetMultipathOutputAndDeviceMapperAndDevice(fakeWwn, fakeExec)
			Ω(err).ShouldNot(HaveOccurred())
			Expect(devMapper).To(Equal("mpathg"))
			Expect(devNames).To(Equal([]string{"sde", "sdf", "sdg"}))
		})

		It("should get device names from multipath output with multiple path groups", func() {
			fakeExec.ExecuteWithTimeoutReturns([]byte(fakeMultipathOutputWithMultiplePathGroups), nil)
			_, devMapper, devNames, err := utils.GetMultipathOutputAndDeviceMapperAndDevice(fakeWwn, fakeExec)
			Ω(err).ShouldNot(HaveOccurred())
			Expect(devMapper).To(Equal("mpathc"))
			Expect(devNames).To(Equal([]string{"sda", "sdb"}))
		})
	})
})
