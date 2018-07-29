package block_device_mounter_utils

import (
	"fmt"
	"testing"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/IBM/ubiquity/utils/logs"
	"github.com/IBM/ubiquity/utils/utils_fakes"
	"github.com/IBM/ubiquity/utils"
)

var _ = Describe("block_device_mounter_utils_private_tests", func() {
	Context(".getK8sBaseDir", func() {
		It("should succeed if path is correct", func() {
			res, err := getK8sPodsBaseDir("/var/lib/kubelet/pods/1f94f1d9-8f36-11e8-b227-005056a4d4cb/volumes/ibm~ubiquity-k8s-flex/pvc-123")
			Expect(err).NotTo(HaveOccurred())
			Expect(res).To(Equal("/var/lib/kubelet/pods"))
		})
		It("should succeed if path is correct and not default", func() {
			res, err := getK8sPodsBaseDir("/tmp/kubelet/pods/1f94f1d9-8f36-11e8-b227-005056a4d4cb/volumes/ibm~ubiquity-k8s-flex/pvc-123")
			Expect(err).NotTo(HaveOccurred())
			Expect(res).To(Equal("/tmp/kubelet/pods"))
		})
		It("should fail if path is not of correct structure", func() {
			k8smountpoint := "/tmp/kubelet/soemthing/1f94f1d9-8f36-11e8-b227-005056a4d4cb/volumes/ibm~ubiquity-k8s-flex/pvc-123"
			res, err := getK8sPodsBaseDir(k8smountpoint)
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(&WrongK8sDirectoryPathError{k8smountpoint}))
			Expect(res).To(Equal(""))
		})
		It("should fail if path is not of correct structure", func() {
			k8smountpoint := "/tmp/kubelet/soemthing/pods/1f94f1d9-8f36-11e8-b227-005056a4d4cb/volumes/ibm~ubiquity-k8s-flex"
			res, err := getK8sPodsBaseDir(k8smountpoint)
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(&WrongK8sDirectoryPathError{k8smountpoint}))
			Expect(res).To(Equal(""))
		})
	})
	Context(".checkSlinkAlreadyExistsOnMountPoint", func() {
		var (
			fakeExecutor *utils_fakes.FakeExecutor
			mountPoint string
		)
		BeforeEach(func() {
			fakeExecutor = new(utils_fakes.FakeExecutor)
			mountPoint = "/tmp/kubelet/pods/1f94f1d9-8f36-11e8-b227-005056a4d4cb/volumes/ibm~ubiquity-k8s-flex/pvc-123"
		})
		It("should return no error if it is the first volume", func() {
			fakeExecutor.GetGlobFilesReturns([]string{}, nil)
			err := checkSlinkAlreadyExistsOnMountPoint("mountPoint", mountPoint, logs.GetLogger(), fakeExecutor)
			Expect(err).To(BeNil())
		})
		It("should return no error if there are no other links", func() {
			fakeExecutor.GetGlobFilesReturns([]string{"/tmp/file1", "/tmp/file2"}, nil)
			fakeExecutor.IsSameFileReturns(false)
			err:= checkSlinkAlreadyExistsOnMountPoint("mountPoint",mountPoint, logs.GetLogger(), fakeExecutor)
			Expect(err).To(BeNil())
		})
		It("should return an error if this mountpoint already has links", func() {
			file := "/tmp/file1"
			fakeExecutor.GetGlobFilesReturns([]string{file}, nil)
			fakeExecutor.IsSameFileReturns(true)
			err:= checkSlinkAlreadyExistsOnMountPoint("mountPoint", mountPoint, logs.GetLogger(), fakeExecutor)
			Expect(err).ToNot(BeNil())
			Expect(err).To(Equal(&PVIsAlreadyUsedByAnotherPod{"mountPoint", []string{file}}))
		})
		It("should return no errors if this mountpoint has only one links and it is the current pvc", func() {
			file := mountPoint
			fakeExecutor.GetGlobFilesReturns([]string{file}, nil)
			fakeExecutor.IsSameFileReturns(true)
			err := checkSlinkAlreadyExistsOnMountPoint("mountPoint", file, logs.GetLogger(), fakeExecutor)
			Expect(err).To(BeNil())
		})
		It("should return error if getk8sbaseDir returns an error", func() {
			k8sMountPoint := "/tmp/kubelet/something"
			err := checkSlinkAlreadyExistsOnMountPoint("mountPoint", k8sMountPoint, logs.GetLogger(), fakeExecutor)
			Expect(err).To(Equal(&WrongK8sDirectoryPathError{k8sMountPoint}))
		})
		It("should return error if glob  returns an error", func() {
			errstrObj := fmt.Errorf("An error ooccured")
			fakeExecutor.GetGlobFilesReturns(nil, errstrObj)
			err := checkSlinkAlreadyExistsOnMountPoint("mountPoint", mountPoint, logs.GetLogger(), fakeExecutor)
			Expect(err).To(Equal(errstrObj))

		})
		It("should return error if stat function returns an error", func() {
			errstrObj := fmt.Errorf("An error ooccured")
			fakeExecutor.GetGlobFilesReturns([]string{"/tmp/file1"}, nil)
			fakeExecutor.StatReturns(nil, errstrObj)
			err:= checkSlinkAlreadyExistsOnMountPoint("mountPoint", mountPoint, logs.GetLogger(), fakeExecutor)
			Expect(err).To(Equal(errstrObj))
		})
		It("should return error if stat on mountpoint function returns an error", func() {
			errstrObj := fmt.Errorf("An error ooccured")
			fakeExecutor.GetGlobFilesReturns([]string{"/tmp/file1"}, nil)
			fakeExecutor.StatReturnsOnCall(1, nil, errstrObj)
			err := checkSlinkAlreadyExistsOnMountPoint("mountPoint", mountPoint, logs.GetLogger(), fakeExecutor)
			Expect(err).To(Equal(errstrObj))
		})
	})
})

func TestGetBlockDeviceUtilsInternal(t *testing.T) {
	RegisterFailHandler(Fail)
	defer utils.InitUbiquityServerTestLogger()()
	RunSpecs(t, "BlockDeviceMounterUtilsInternal Test Suite")
}

