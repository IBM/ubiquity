package block_device_mounter_utils

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/IBM/ubiquity/utils/logs"
	"github.com/IBM/ubiquity/utils/utils_fakes"
)

var _ = Describe("block_device_mounter_utils_private_tests", func() {
	Context(".getK8sBaseDir", func() {
		It("should succeed if path is correct", func() {
			res, err := getK8sBaseDir("/var/lib/kubelet/pods/1f94f1d9-8f36-11e8-b227-005056a4d4cb/volumes/ibm~ubiquity-k8s-flex/pvc-123")
			Expect(err).NotTo(HaveOccurred())
			Expect(res).To(Equal("/var/lib/kubelet/pods"))
		})
		It("should succeed if path is correct and not default", func() {
			res, err := getK8sBaseDir("/tmp/kubelet/pods/1f94f1d9-8f36-11e8-b227-005056a4d4cb/volumes/ibm~ubiquity-k8s-flex/pvc-123")
			Expect(err).NotTo(HaveOccurred())
			Expect(res).To(Equal("/tmp/kubelet/pods"))
		})
		It("should fail if path is not of correct structure", func() {
			k8smountpoint := "/tmp/kubelet/soemthing/1f94f1d9-8f36-11e8-b227-005056a4d4cb/volumes/ibm~ubiquity-k8s-flex/pvc-123"
			res, err := getK8sBaseDir(k8smountpoint)
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(&WrongK8sDirectoryPathError{k8smountpoint}))
			Expect(res).To(Equal(""))
		})
	})
	Context(".checkSlinkAlreadyEistsOnMountPoint", func() {
		var (
			fakeExecutor *utils_fakes.FakeExecutor
		)
		BeforeEach(func() {
			fakeExecutor = new(utils_fakes.FakeExecutor)
		})
		It("should return false if it is the first volume", func() {
			fakeExecutor.GetGlobFilesReturns([]string{}, nil)
			res, err, links := checkSlinkAlreadyExistsOnMountPoint("mountPoint", "/tmp/kubelet/pods/adad/", logs.GetLogger(), fakeExecutor)
			Expect(err).To(BeNil())
			Expect(len(links)).To(Equal(0))
			Expect(res).To(Equal(false))
		})
		It("should return false if there are no other links", func() {
			fakeExecutor.GetGlobFilesReturns([]string{"/tmp/file1", "/tmp/file2"}, nil)
			fakeExecutor.IsSameFileReturns(false)
			res, err, links := checkSlinkAlreadyExistsOnMountPoint("mountPoint", "/tmp/kubelet/pods/adad/", logs.GetLogger(), fakeExecutor)
			Expect(err).To(BeNil())
			Expect(len(links)).To(Equal(0))
			Expect(res).To(Equal(false))
		})
		It("should return true if this mountpoint already has links", func() {
			file := "/tmp/file1"
			fakeExecutor.GetGlobFilesReturns([]string{file}, nil)
			fakeExecutor.IsSameFileReturns(true)
			res, err, links := checkSlinkAlreadyExistsOnMountPoint("mountPoint", "/tmp/kubelet/pods/adad/", logs.GetLogger(), fakeExecutor)
			Expect(err).To(BeNil())
			Expect(len(links)).To(Equal(1))
			Expect(links).To(Equal([]string{file}))
			Expect(res).To(Equal(true))
		})
		It("should return false if this mountpoint has only opne links and it is the current pvc", func() {
			file := "/tmp/kubelet/pods/adad/pvc-123"
			fakeExecutor.GetGlobFilesReturns([]string{file}, nil)
			fakeExecutor.IsSameFileReturns(true)
			res, err, links := checkSlinkAlreadyExistsOnMountPoint("mountPoint", "/tmp/kubelet/pods/adad/pvc-123", logs.GetLogger(), fakeExecutor)
			Expect(err).To(BeNil())
			Expect(len(links)).To(Equal(0))
			Expect(res).To(Equal(false))
		})
		It("should return error if getk8sbaseDir returns an error", func() {
			k8sMountPoint := "/tmp/kubelet/something"
			res, err, links := checkSlinkAlreadyExistsOnMountPoint("mountPoint", k8sMountPoint, logs.GetLogger(), fakeExecutor)
			Expect(err).To(Equal(&WrongK8sDirectoryPathError{k8sMountPoint}))
			Expect(len(links)).To(Equal(0))
			Expect(res).To(Equal(false))
		})
		It("should return error if glob  returns an error", func() {
			errstrObj := fmt.Errorf("An error ooccured")
			fakeExecutor.GetGlobFilesReturns(nil, errstrObj)
			res, err, links := checkSlinkAlreadyExistsOnMountPoint("mountPoint", "/tmp/kubelet/pods/adad/pvc-123", logs.GetLogger(), fakeExecutor)
			Expect(err).To(Equal(errstrObj))
			Expect(len(links)).To(Equal(0))
			Expect(res).To(Equal(false))
		})
		It("should return error if stat function returns an error", func() {
			errstrObj := fmt.Errorf("An error ooccured")
			fakeExecutor.GetGlobFilesReturns([]string{"/tmp/file1"}, nil)
			fakeExecutor.StatReturns(nil, errstrObj)
			res, err, links := checkSlinkAlreadyExistsOnMountPoint("mountPoint", "/tmp/kubelet/pods/adad/pvc-123", logs.GetLogger(), fakeExecutor)
			Expect(err).To(Equal(errstrObj))
			Expect(len(links)).To(Equal(0))
			Expect(res).To(Equal(false))
		})
		It("should return error if stat on mountpoint function returns an error", func() {
			errstrObj := fmt.Errorf("An error ooccured")
			fakeExecutor.GetGlobFilesReturns([]string{"/tmp/file1"}, nil)
			fakeExecutor.StatReturnsOnCall(1, nil, errstrObj)
			res, err, links := checkSlinkAlreadyExistsOnMountPoint("mountPoint", "/tmp/kubelet/pods/adad/pvc-123", logs.GetLogger(), fakeExecutor)
			Expect(err).To(Equal(errstrObj))
			Expect(len(links)).To(Equal(0))
			Expect(res).To(Equal(false))
		})
	})
})
