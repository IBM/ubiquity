package local_test

import (
	"fmt"
	"log"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.ibm.com/almaden-containers/ubiquity/fakes"
	"github.ibm.com/almaden-containers/ubiquity/local"
	"github.ibm.com/almaden-containers/ubiquity/model"
)

var _ = Describe("local-client", func() {
	var (
		client             model.StorageClient
		logger             *log.Logger
		fakeSpectrumClient *fakes.FakeSpectrum
		fakeDbClient       *fakes.FakeDatabaseClient
		fakeLock           *fakes.FakeFileLock
		fakeExec           *fakes.FakeExecutor
		fakeConfig         model.SpectrumConfig
		err                error
	)
	BeforeEach(func() {
		logger = log.New(os.Stdout, "ubiquity: ", log.Lshortfile|log.LstdFlags)
		fakeSpectrumClient = new(fakes.FakeSpectrum)
		fakeDbClient = new(fakes.FakeDatabaseClient)
		fakeLock = new(fakes.FakeFileLock)
		fakeExec = new(fakes.FakeExecutor)
		fakeConfig = model.SpectrumConfig{}
	})

	Context(".Activate", func() {
		BeforeEach(func() {
			client, err = local.NewSpectrumLocalClientWithClients(logger, fakeSpectrumClient, fakeDbClient, fakeLock, fakeExec, fakeConfig)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should fail when fileLock failes to get the lock", func() {
			fakeLock.LockReturns(fmt.Errorf("error in lock call"))
			fakeLock.UnlockReturns(nil)
			err = client.Activate()
			Expect(err).To(HaveOccurred())
			Expect(fakeLock.LockCallCount()).To(Equal(1))
			Expect(fakeLock.UnlockCallCount()).To(Equal(0))
			Expect(fakeSpectrumClient.IsFilesystemMountedCallCount()).To(Equal(0))
			Expect(fakeSpectrumClient.MountFileSystemCallCount()).To(Equal(0))
		})

		It("should fail when fileLock failes to unlock", func() {
			fakeLock.LockReturns(nil)
			fakeLock.UnlockReturns(fmt.Errorf("error in unlock call"))
			err = client.Activate()
			Expect(err).To(HaveOccurred())
			Expect(fakeLock.LockCallCount()).To(Equal(1))
			Expect(fakeLock.UnlockCallCount()).To(Equal(1))
		})

		It("should fail when spectrum client IsFilesystemMounted errors", func() {
			fakeLock.LockReturns(nil)
			fakeLock.UnlockReturns(nil)
			fakeSpectrumClient.IsFilesystemMountedReturns(false, fmt.Errorf("error in isFilesystemMounted"))
			err = client.Activate()
			Expect(err).To(HaveOccurred())
			Expect(fakeLock.LockCallCount()).To(Equal(1))
			Expect(fakeLock.UnlockCallCount()).To(Equal(1))
			Expect(fakeSpectrumClient.IsFilesystemMountedCallCount()).To(Equal(1))
			Expect(fakeSpectrumClient.MountFileSystemCallCount()).To(Equal(0))
		})

		It("should fail when spectrum client MountFileSystem errors", func() {
			fakeLock.LockReturns(nil)
			fakeLock.UnlockReturns(nil)
			fakeSpectrumClient.IsFilesystemMountedReturns(false, nil)
			fakeSpectrumClient.MountFileSystemReturns(fmt.Errorf("error in mount filesystem"))
			err = client.Activate()
			Expect(err).To(HaveOccurred())
			Expect(fakeLock.LockCallCount()).To(Equal(1))
			Expect(fakeLock.UnlockCallCount()).To(Equal(1))
			Expect(fakeSpectrumClient.IsFilesystemMountedCallCount()).To(Equal(1))
			Expect(fakeSpectrumClient.MountFileSystemCallCount()).To(Equal(1))
			Expect(fakeSpectrumClient.GetClusterIdCallCount()).To(Equal(0))
		})

		It("should fail when spectrum client GetClusterID errors", func() {
			fakeLock.LockReturns(nil)
			fakeLock.UnlockReturns(nil)
			fakeSpectrumClient.IsFilesystemMountedReturns(true, nil)
			fakeSpectrumClient.GetClusterIdReturns("", fmt.Errorf("error getting the cluster ID"))
			err = client.Activate()
			Expect(err).To(HaveOccurred())
			Expect(fakeLock.LockCallCount()).To(Equal(1))
			Expect(fakeLock.UnlockCallCount()).To(Equal(1))
			Expect(fakeSpectrumClient.IsFilesystemMountedCallCount()).To(Equal(1))
			Expect(fakeSpectrumClient.MountFileSystemCallCount()).To(Equal(0))
			Expect(fakeSpectrumClient.GetClusterIdCallCount()).To(Equal(1))
			Expect(fakeDbClient.SetClusterIdCallCount()).To(Equal(0))
		})

		It("should fail when spectrum client GetClusterID return empty ID", func() {
			fakeLock.LockReturns(nil)
			fakeLock.UnlockReturns(nil)
			fakeSpectrumClient.IsFilesystemMountedReturns(true, nil)
			fakeSpectrumClient.GetClusterIdReturns("", nil)
			err = client.Activate()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Unable to retrieve clusterId: clusterId is empty"))
			Expect(fakeLock.LockCallCount()).To(Equal(1))
			Expect(fakeLock.UnlockCallCount()).To(Equal(1))
			Expect(fakeSpectrumClient.IsFilesystemMountedCallCount()).To(Equal(1))
			Expect(fakeSpectrumClient.MountFileSystemCallCount()).To(Equal(0))
			Expect(fakeSpectrumClient.GetClusterIdCallCount()).To(Equal(1))
			Expect(fakeDbClient.SetClusterIdCallCount()).To(Equal(0))
		})

		It("should fail when dbClient CreateVolumeTable errors", func() {
			fakeLock.LockReturns(nil)
			fakeLock.UnlockReturns(nil)
			fakeSpectrumClient.IsFilesystemMountedReturns(true, nil)
			fakeSpectrumClient.GetClusterIdReturns("fake-cluster", nil)
			fakeDbClient.CreateVolumeTableReturns(fmt.Errorf("error in creating volume table"))
			err = client.Activate()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("error in creating volume table"))
			Expect(fakeLock.LockCallCount()).To(Equal(1))
			Expect(fakeLock.UnlockCallCount()).To(Equal(1))
			Expect(fakeSpectrumClient.IsFilesystemMountedCallCount()).To(Equal(1))
			Expect(fakeSpectrumClient.MountFileSystemCallCount()).To(Equal(0))
			Expect(fakeSpectrumClient.GetClusterIdCallCount()).To(Equal(1))
			Expect(fakeDbClient.SetClusterIdCallCount()).To(Equal(1))
			Expect(fakeDbClient.CreateVolumeTableCallCount()).To(Equal(1))
		})

		It("should succeed when everything is fine with no mounting", func() {
			fakeLock.LockReturns(nil)
			fakeLock.UnlockReturns(nil)
			fakeSpectrumClient.IsFilesystemMountedReturns(true, nil)
			fakeSpectrumClient.GetClusterIdReturns("fake-cluster", nil)
			fakeDbClient.CreateVolumeTableReturns(nil)
			err = client.Activate()
			Expect(err).ToNot(HaveOccurred())
			Expect(fakeLock.LockCallCount()).To(Equal(1))
			Expect(fakeLock.UnlockCallCount()).To(Equal(1))
			Expect(fakeSpectrumClient.IsFilesystemMountedCallCount()).To(Equal(1))
			Expect(fakeSpectrumClient.MountFileSystemCallCount()).To(Equal(0))
			Expect(fakeSpectrumClient.GetClusterIdCallCount()).To(Equal(1))
			Expect(fakeDbClient.SetClusterIdCallCount()).To(Equal(1))
			Expect(fakeDbClient.CreateVolumeTableCallCount()).To(Equal(1))
		})

		It("should succeed when everything is fine with mounting", func() {
			fakeLock.LockReturns(nil)
			fakeLock.UnlockReturns(nil)
			fakeSpectrumClient.IsFilesystemMountedReturns(false, nil)
			fakeSpectrumClient.MountFileSystemReturns(nil)
			fakeSpectrumClient.GetClusterIdReturns("fake-cluster", nil)
			fakeDbClient.CreateVolumeTableReturns(nil)
			err = client.Activate()
			Expect(err).ToNot(HaveOccurred())
			Expect(fakeLock.LockCallCount()).To(Equal(1))
			Expect(fakeLock.UnlockCallCount()).To(Equal(1))
			Expect(fakeSpectrumClient.IsFilesystemMountedCallCount()).To(Equal(1))
			Expect(fakeSpectrumClient.MountFileSystemCallCount()).To(Equal(1))
			Expect(fakeSpectrumClient.GetClusterIdCallCount()).To(Equal(1))
			Expect(fakeDbClient.SetClusterIdCallCount()).To(Equal(1))
			Expect(fakeDbClient.CreateVolumeTableCallCount()).To(Equal(1))
		})

	})

	Context(".CreateVolume", func() {
		var (
			opts map[string]interface{}
		)
		BeforeEach(func() {
			client, err = local.NewSpectrumLocalClientWithClients(logger, fakeSpectrumClient, fakeDbClient, fakeLock, fakeExec, fakeConfig)
			Expect(err).ToNot(HaveOccurred())
			fakeLock.LockReturns(nil)
			fakeLock.UnlockReturns(nil)
			fakeSpectrumClient.IsFilesystemMountedReturns(false, nil)
			fakeSpectrumClient.MountFileSystemReturns(nil)
			fakeSpectrumClient.GetClusterIdReturns("fake-cluster", nil)
			fakeDbClient.CreateVolumeTableReturns(nil)
			err = client.Activate()
			Expect(err).ToNot(HaveOccurred())

		})

		It("should fail when fileLock failes to get the lock", func() {
			fakeLock.LockReturns(fmt.Errorf("error in lock call"))
			fakeLock.UnlockReturns(nil)

			err = client.CreateVolume("fake-volume", opts)
			Expect(err).To(HaveOccurred())
			Expect(fakeLock.LockCallCount()).To(Equal(2))
			Expect(fakeLock.UnlockCallCount()).To(Equal(1))
			Expect(fakeDbClient.VolumeExistsCallCount()).To(Equal(0))

		})

		It("should fail when fileLock failes to release the lock", func() {
			fakeLock.LockReturns(nil)
			fakeLock.UnlockReturns(fmt.Errorf("error in unlock call"))

			err = client.CreateVolume("fake-volume", opts)
			Expect(err).To(HaveOccurred())
			Expect(fakeLock.LockCallCount()).To(Equal(2))
			Expect(fakeLock.UnlockCallCount()).To(Equal(2))
			Expect(fakeDbClient.VolumeExistsCallCount()).To(Equal(1))

		})

		It("should fail when dbClient volumeExists errors", func() {
			fakeLock.LockReturns(nil)
			fakeLock.UnlockReturns(nil)
			fakeDbClient.VolumeExistsReturns(false, fmt.Errorf("error checking if volume exists"))
			err = client.CreateVolume("fake-volume", opts)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("error checking if volume exists"))
			Expect(fakeLock.LockCallCount()).To(Equal(2))
			Expect(fakeLock.UnlockCallCount()).To(Equal(2))
			Expect(fakeDbClient.VolumeExistsCallCount()).To(Equal(1))
			Expect(fakeSpectrumClient.CreateFilesetCallCount()).To(Equal(0))
		})

		It("should fail when dbClient volumeExists returns true", func() {
			fakeLock.LockReturns(nil)
			fakeLock.UnlockReturns(nil)
			fakeDbClient.VolumeExistsReturns(true, nil)
			err = client.CreateVolume("fake-volume", opts)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Volume already exists"))
			Expect(fakeLock.LockCallCount()).To(Equal(2))
			Expect(fakeLock.UnlockCallCount()).To(Equal(2))
			Expect(fakeDbClient.VolumeExistsCallCount()).To(Equal(1))
			Expect(fakeSpectrumClient.CreateFilesetCallCount()).To(Equal(0))
		})

		Context(".FilesetVolume", func() {
			BeforeEach(func() {
				opts = make(map[string]interface{})
				opts[""] = ""
			})
			It("", func() {
				Expect(true).To(Equal(true))
			})
		})

		Context(".FilesetQuotaVolume", func() {
			BeforeEach(func() {
				opts = make(map[string]interface{})
				opts[""] = ""
			})
			It("", func() {
				Expect(true).To(Equal(true))
			})
		})

		Context(".LightWeightVolume", func() {
			BeforeEach(func() {
				opts = make(map[string]interface{})
				opts[""] = ""
			})
			It("", func() {
				Expect(true).To(Equal(true))
			})
		})

	})
})
