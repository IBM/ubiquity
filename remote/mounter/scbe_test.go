package mounter_test

import (
	"fmt"
	"testing"

	"github.com/IBM/ubiquity/fakes"
	"github.com/IBM/ubiquity/remote/mounter"
	"github.com/IBM/ubiquity/remote/mounter/block_device_utils"
	"github.com/IBM/ubiquity/resources"
	"github.com/IBM/ubiquity/utils"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("scbe_mounter_test", func() {
	var (
		fakeExec    *fakes.FakeExecutor
		fakeBdUtils *fakes.FakeBlockDeviceMounterUtils
		scbeMounter resources.Mounter
	)

	BeforeEach(func() {
		fakeExec = new(fakes.FakeExecutor)
		fakeBdUtils = new(fakes.FakeBlockDeviceMounterUtils)
		scbeMounter = mounter.NewScbeMounterWithExecuter(fakeBdUtils, fakeExec)
		fakeExec.IsDirEmptyReturns(true, nil)
	})

	Context(".Unmount", func() { 
		It("should continue flow if volume is not discovered", func() {
			returnedErr := &block_device_utils.VolumeNotFoundError{"volumewwn"}
			fakeBdUtils.DiscoverReturns("", returnedErr)
			volumeConfig := make(map[string]interface{})
			volumeConfig["Wwn"] = "volumewwn"
			err := scbeMounter.Unmount(resources.UnmountRequest{volumeConfig, resources.RequestContext{}})
			Expect(err).To(BeNil())
			Expect(fakeBdUtils.UnmountDeviceFlowCallCount()).To(Equal(0))
			Expect(fakeExec.StatCallCount()).To(Equal(1))
		})
		It("should fail if error from discovery is not voumeNotFound", func() {
			returnedErr := fmt.Errorf("An error has occured")
			fakeBdUtils.DiscoverReturns("", returnedErr)
			volumeConfig := make(map[string]interface{})
			volumeConfig["Wwn"] = "volumewwn"
			err := scbeMounter.Unmount(resources.UnmountRequest{volumeConfig, resources.RequestContext{}})
			Expect(err).To(Equal(returnedErr))
			Expect(fakeBdUtils.UnmountDeviceFlowCallCount()).To(Equal(0))
			Expect(fakeExec.StatCallCount()).To(Equal(0))
		})
		It("should call unmountDeviceFlow if discover succeeded", func() {
			volumeConfig := make(map[string]interface{})
			volumeConfig["Wwn"] = "volumewwn"
			err := scbeMounter.Unmount(resources.UnmountRequest{volumeConfig, resources.RequestContext{}})
			Expect(err).ToNot(HaveOccurred())
			Expect(fakeBdUtils.UnmountDeviceFlowCallCount()).To(Equal(1))
			Expect(fakeExec.StatCallCount()).To(Equal(1))
		})
		It("should fail if unmountDeviceFlow returns an error", func() {
			returnedErr := fmt.Errorf("An error has occured")
			fakeBdUtils.UnmountDeviceFlowReturns(returnedErr)
			volumeConfig := make(map[string]interface{})
			volumeConfig["Wwn"] = "volumewwn"
			err := scbeMounter.Unmount(resources.UnmountRequest{volumeConfig, resources.RequestContext{}})
			Expect(err).To(HaveOccurred())
			Expect(fakeBdUtils.UnmountDeviceFlowCallCount()).To(Equal(1))
			Expect(fakeExec.StatCallCount()).To(Equal(0))
		})
		It("should not fail if stat on file returns error", func() {
			returnedErr := fmt.Errorf("An error has occured")
			fakeExec.StatReturns(nil, returnedErr)
			volumeConfig := make(map[string]interface{})
			volumeConfig["Wwn"] = "volumewwn"
			err := scbeMounter.Unmount(resources.UnmountRequest{volumeConfig, resources.RequestContext{}})
			Expect(err).ToNot(HaveOccurred())
			Expect(fakeBdUtils.UnmountDeviceFlowCallCount()).To(Equal(1))
			Expect(fakeExec.StatCallCount()).To(Equal(1))
			Expect(fakeExec.RemoveAllCallCount()).To(Equal(0))
		})
		It("should  fail if Removeall returns error", func() {
			returnedErr := fmt.Errorf("An error has occured")
			fakeExec.RemoveAllReturns(returnedErr)
			volumeConfig := make(map[string]interface{})
			volumeConfig["Wwn"] = "volumewwn"
			err := scbeMounter.Unmount(resources.UnmountRequest{volumeConfig, resources.RequestContext{}})
			Expect(err).To(HaveOccurred())
			Expect(fakeBdUtils.UnmountDeviceFlowCallCount()).To(Equal(1))
			Expect(fakeExec.StatCallCount()).To(Equal(1))
			Expect(fakeExec.RemoveAllCallCount()).To(Equal(1))
		})
		It("should  fail if mountpoint dir is not empty", func() {
			fakeExec.IsDirEmptyReturns(false, nil)
			volumeConfig := make(map[string]interface{})
			volumeConfig["Wwn"] = "volumewwn"
			err := scbeMounter.Unmount(resources.UnmountRequest{volumeConfig, resources.RequestContext{}})
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(&mounter.DirecotryIsNotEmptyError{fmt.Sprintf("/ubiquity/%s", volumeConfig["Wwn"])}))
			Expect(fakeBdUtils.UnmountDeviceFlowCallCount()).To(Equal(1))
			Expect(fakeExec.StatCallCount()).To(Equal(1))
			Expect(fakeExec.RemoveAllCallCount()).To(Equal(0))
		})
		It("should  fail if mountpoint dir returns erorr", func() {
			returnedErr := fmt.Errorf("An error has occured")
			fakeExec.IsDirEmptyReturns(false, returnedErr)
			volumeConfig := make(map[string]interface{})
			volumeConfig["Wwn"] = "volumewwn"
			err := scbeMounter.Unmount(resources.UnmountRequest{volumeConfig, resources.RequestContext{}})
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(returnedErr))
			Expect(fakeBdUtils.UnmountDeviceFlowCallCount()).To(Equal(1))
			Expect(fakeExec.StatCallCount()).To(Equal(1))
			Expect(fakeExec.RemoveAllCallCount()).To(Equal(0))
		})
		It("should  continue if discover failed on faulty device", func() {
			returnedErr := &block_device_utils.FaultyDeviceError{"mapthx"}
			fakeBdUtils.DiscoverReturns("", returnedErr)
			volumeConfig := make(map[string]interface{})
			volumeConfig["Wwn"] = "volumewwn"
			_ = scbeMounter.Unmount(resources.UnmountRequest{volumeConfig, resources.RequestContext{}})
			Expect(fakeBdUtils.DiscoverCallCount()).To(Equal(1))
			Expect(fakeBdUtils.UnmountDeviceFlowCallCount()).To(Equal(1))
			Expect(fakeExec.StatCallCount()).To(Equal(1))
			Expect(fakeExec.RemoveAllCallCount()).To(Equal(1))
		})
	})
})

func TestSCBEMounter(t *testing.T) {
	RegisterFailHandler(Fail)
	defer utils.InitUbiquityServerTestLogger()()
	RunSpecs(t, "SCBEMounter Test Suite")
}
