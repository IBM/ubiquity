package utils_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/IBM/ubiquity/fakes"
	"github.com/IBM/ubiquity/utils"
	"github.com/IBM/ubiquity/utils/logs"
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

var fakeWwn = "6005076306FFD69d0000000000001004"

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
mpathc (36005076306ffd69d0000000000001004) dm-4 IBM     ,2145
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

var fakeMultipathOutputWithDifferentSpaces = `
mpathj (36005076306ffd69d0000000000001004) dm-27 IBM     ,2107900
size=1.0G features='0' hwhandler='0' wp=rw
'-+- policy='service-time 0' prio=1 status=active
  |- 33:0:12:1 sdcp 69:208 active ready running
  |- 33:0:8:1  sdcn 69:176 active ready running
  |- 33:0:9:1  sdco 69:192 active ready running
  |- 34:0:10:1 sdcr 69:240 active ready running
  |- 34:0:12:1 sdcs 70:0   active ready running
  '- 34:0:9:1  sdcq 69:224 active ready running
`

var fakeMultipathOutputWithWarnings = `
Apr 04 16:38:06 | sde: couldn't get target port group
Apr 04 16:38:06 | sdd: couldn't get target port group
mpathj (36005076306ffd69d0000000000001004) dm-17 IBM     ,2145
size=1.0G features='1 queue_if_no_path' hwhandler='0' wp=rw
|-+- policy='service-time 0' prio=0 status=enabled
| '- 39:0:0:1 sde 8:64 failed faulty running
'-+- policy='service-time 0' prio=0 status=enabled
  '- 40:0:0:1 sdd 8:48 failed faulty running
mpathi (3600507680c8701159800000000001af3) dm-14 IBM     ,2145
size=20G features='1 queue_if_no_path' hwhandler='0' wp=rw
|-+- policy='service-time 0' prio=50 status=active
| '- 40:0:0:0 sdb 8:16 active ready running
'-+- policy='service-time 0' prio=10 status=enabled
  '- 39:0:0:0 sdc 8:32 active ready running
`

var fakeMultipathOutputWithWarningsRemoved = `mpathj (36005076306ffd69d0000000000001004) dm-17 IBM     ,2145
size=1.0G features='1 queue_if_no_path' hwhandler='0' wp=rw
|-+- policy='service-time 0' prio=0 status=enabled
| '- 39:0:0:1 sde 8:64 failed faulty running
'-+- policy='service-time 0' prio=0 status=enabled
  '- 40:0:0:1 sdd 8:48 failed faulty running
mpathi (3600507680c8701159800000000001af3) dm-14 IBM     ,2145
size=20G features='1 queue_if_no_path' hwhandler='0' wp=rw
|-+- policy='service-time 0' prio=50 status=active
| '- 40:0:0:0 sdb 8:16 active ready running
'-+- policy='service-time 0' prio=10 status=enabled
  '- 39:0:0:0 sdc 8:32 active ready running`

var _ = FDescribe("scbe_mounter_test", func() {
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

		It("should get device names from multipath output with different spaces", func() {
			fakeExec.ExecuteWithTimeoutReturns([]byte(fakeMultipathOutputWithDifferentSpaces), nil)
			_, devMapper, devNames, err := utils.GetMultipathOutputAndDeviceMapperAndDevice(fakeWwn, fakeExec)
			Ω(err).ShouldNot(HaveOccurred())
			Expect(devMapper).To(Equal("mpathj"))
			Expect(devNames).To(Equal([]string{"sdcp", "sdcn", "sdco", "sdcr", "sdcs", "sdcq"}))
		})

		It("should get device names from multipath output with warning header", func() {
			fakeExec.ExecuteWithTimeoutReturns([]byte(fakeMultipathOutputWithWarnings), nil)
			_, devMapper, devNames, err := utils.GetMultipathOutputAndDeviceMapperAndDevice(fakeWwn, fakeExec)
			Ω(err).ShouldNot(HaveOccurred())
			Expect(devMapper).To(Equal("mpathj"))
			Expect(devNames).To(Equal([]string{"sde", "sdd"}))
		})
	})

	Context("ExcludeNoTargetPortGroupMessagesFromMultipathOutput", func() {

		It("should get device names from multipath output", func() {
			logger := logs.GetLogger()
			out := utils.ExcludeNoTargetPortGroupMessagesFromMultipathOutput(fakeMultipathOutputWithWarnings, logger)
			Expect(out).To(Equal(fakeMultipathOutputWithWarningsRemoved))
		})
	})
})
