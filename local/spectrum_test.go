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

			It("should fail when spectrum client fails to create fileset", func() {
				fakeSpectrumClient.CreateFilesetReturns(fmt.Errorf("error creating fileset"))
				err = client.CreateVolume("fake-fileset", opts)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("error creating fileset"))
				Expect(fakeSpectrumClient.CreateFilesetCallCount()).To(Equal(1))
				Expect(fakeDbClient.InsertFilesetVolumeCallCount()).To(Equal(0))
			})

			It("should fail when dbClient fails to insert fileset record", func() {
				fakeSpectrumClient.CreateFilesetReturns(nil)
				fakeDbClient.InsertFilesetVolumeReturns(fmt.Errorf("error inserting fileset"))

				err = client.CreateVolume("fake-fileset", opts)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("error inserting fileset"))
				Expect(fakeSpectrumClient.CreateFilesetCallCount()).To(Equal(1))
				Expect(fakeDbClient.InsertFilesetVolumeCallCount()).To(Equal(1))
			})

			It("should succeed to create fileset", func() {
				fakeSpectrumClient.CreateFilesetReturns(nil)
				fakeDbClient.InsertFilesetVolumeReturns(nil)

				err = client.CreateVolume("fake-fileset", opts)
				Expect(err).ToNot(HaveOccurred())
				Expect(fakeSpectrumClient.CreateFilesetCallCount()).To(Equal(1))
				Expect(fakeDbClient.InsertFilesetVolumeCallCount()).To(Equal(1))
			})

		})

		Context(".FilesetQuotaVolume", func() {
			BeforeEach(func() {
				opts = make(map[string]interface{})
				opts["fileset"] = "fake-fileset"
				opts["type"] = "fileset"
				opts["filesystem"] = "fake-filesystem"
			})
			Context(".WithQuota", func() {
				BeforeEach(func() {
					opts["quota"] = "1Gi"
				})
				It("should fail when spectrum client fails to list fileset quota", func() {
					fakeSpectrumClient.ListFilesetQuotaReturns("", fmt.Errorf("error in list quota"))
					err = client.CreateVolume("fake-fileset", opts)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("error in list quota"))
					Expect(fakeDbClient.InsertFilesetQuotaVolumeCallCount()).To(Equal(0))
				})
				It("should fail when spectrum client returns a missmatching fileset quota", func() {
					fakeSpectrumClient.ListFilesetQuotaReturns("2Gi", nil)
					err = client.CreateVolume("fake-fileset", opts)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("Mismatch between user-specified and listed quota for fileset fake-fileset"))
					Expect(fakeDbClient.InsertFilesetQuotaVolumeCallCount()).To(Equal(0))
				})
				It("should fail when dbClient fails to insert Fileset quota volume", func() {
					fakeSpectrumClient.ListFilesetQuotaReturns("1Gi", nil)
					fakeDbClient.InsertFilesetQuotaVolumeReturns(fmt.Errorf("error inserting filesetquotavolume"))
					err = client.CreateVolume("fake-fileset", opts)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("error inserting filesetquotavolume"))
					Expect(fakeDbClient.InsertFilesetQuotaVolumeCallCount()).To(Equal(1))
				})
				It("should succeed when the options are well specified", func() {
					fakeSpectrumClient.ListFilesetQuotaReturns("1Gi", nil)
					fakeDbClient.InsertFilesetQuotaVolumeReturns(nil)
					err = client.CreateVolume("fake-fileset", opts)
					Expect(err).ToNot(HaveOccurred())
					Expect(fakeSpectrumClient.ListFilesetQuotaCallCount()).To(Equal(1))
					Expect(fakeDbClient.InsertFilesetQuotaVolumeCallCount()).To(Equal(1))
				})

			})
			Context(".WithNoQuota", func() {

				It("should fail when spectrum client fails to list fileset quota", func() {
					fakeSpectrumClient.ListFilesetReturns(model.VolumeMetadata{}, fmt.Errorf("error in list fileset"))
					err = client.CreateVolume("fake-fileset", opts)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("error in list fileset"))
					Expect(fakeDbClient.InsertFilesetVolumeCallCount()).To(Equal(0))
				})
				It("should fail when dbClient fails to insert Fileset quota volume", func() {
					fakeVolume := model.VolumeMetadata{Name: "fake-fileset", Mountpoint: "fake-mountpoint"}
					fakeSpectrumClient.ListFilesetReturns(fakeVolume, nil)
					fakeDbClient.InsertFilesetVolumeReturns(fmt.Errorf("error inserting filesetvolume"))
					err = client.CreateVolume("fake-fileset", opts)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("error inserting filesetvolume"))
					Expect(fakeDbClient.InsertFilesetVolumeCallCount()).To(Equal(1))
				})
				It("should succeed when parameters are well specified", func() {
					fakeVolume := model.VolumeMetadata{Name: "fake-fileset", Mountpoint: "fake-mountpoint"}
					fakeSpectrumClient.ListFilesetReturns(fakeVolume, nil)
					fakeDbClient.InsertFilesetVolumeReturns(nil)
					err = client.CreateVolume("fake-fileset", opts)
					Expect(err).ToNot(HaveOccurred())
					Expect(fakeDbClient.InsertFilesetVolumeCallCount()).To(Equal(1))
				})

			})
		})

		Context(".LightWeightVolume", func() {
			BeforeEach(func() {
				opts = make(map[string]interface{})
				opts["fileset"] = "fake-fileset"
				opts["filesystem"] = "fake-filesystem"
				opts["type"] = "lightweight"
			})
			It("should fail when spectrum client IsfilesetLinked errors", func() {
				fakeSpectrumClient.IsFilesetLinkedReturns(false, fmt.Errorf("error in checking fileset linked"))
				err = client.CreateVolume("fake-fileset", opts)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("error in checking fileset linked"))
				Expect(fakeSpectrumClient.IsFilesetLinkedCallCount()).To(Equal(1))
				Expect(fakeSpectrumClient.LinkFilesetCallCount()).To(Equal(0))
			})
			It("should fail when spectrum client LinkFileset errors", func() {
				fakeSpectrumClient.IsFilesetLinkedReturns(false, nil)
				fakeSpectrumClient.LinkFilesetReturns(fmt.Errorf("error linking fileset"))
				err = client.CreateVolume("fake-fileset", opts)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("error linking fileset"))
				Expect(fakeSpectrumClient.IsFilesetLinkedCallCount()).To(Equal(1))
				Expect(fakeSpectrumClient.LinkFilesetCallCount()).To(Equal(1))
			})

			It("should fail when spectrum client GetFilesystemMountpoint errors", func() {
				fakeSpectrumClient.IsFilesetLinkedReturns(true, nil)
				fakeSpectrumClient.GetFilesystemMountpointReturns("", fmt.Errorf("error getting mountpoint"))
				err = client.CreateVolume("fake-fileset", opts)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("error getting mountpoint"))
				Expect(fakeSpectrumClient.IsFilesetLinkedCallCount()).To(Equal(1))
				Expect(fakeSpectrumClient.LinkFilesetCallCount()).To(Equal(0))
				Expect(fakeSpectrumClient.GetFilesystemMountpointCallCount()).To(Equal(1))
			})

			It("should fail when spectrum client GetFilesystemMountpoint errors", func() {
				fakeSpectrumClient.IsFilesetLinkedReturns(true, nil)
				fakeSpectrumClient.GetFilesystemMountpointReturns("fake-mountpoint", nil)
				fakeExec.StatReturns(nil, fmt.Errorf("error in os.Stat"))
				fakeExec.MkdirReturns(fmt.Errorf("error in mkdir"))
				err = client.CreateVolume("fake-fileset", opts)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("error in mkdir"))
				Expect(fakeSpectrumClient.IsFilesetLinkedCallCount()).To(Equal(1))
				Expect(fakeSpectrumClient.LinkFilesetCallCount()).To(Equal(0))
				Expect(fakeSpectrumClient.GetFilesystemMountpointCallCount()).To(Equal(1))
				// Expect(fakeExec.StatCallCount()).To(Equal(1))
			})

		})

	})
})
