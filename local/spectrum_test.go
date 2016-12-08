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
	"github.ibm.com/almaden-containers/ubiquity/utils"
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
		client, err = local.NewSpectrumLocalClientWithClients(logger, fakeSpectrumClient, fakeDbClient, fakeLock, fakeExec, fakeConfig)
		Expect(err).ToNot(HaveOccurred())

	})

	Context(".Activate", func() {
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

		Context(".FilesetVolume", func() {
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

	Context(".RemoveVolume", func() {
		It("should fail when the fileLock fails to aquire the lock", func() {
			fakeLock.LockReturns(fmt.Errorf("failed to aquire lock"))
			err = client.RemoveVolume("fake-volume", false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("failed to aquire lock"))
			Expect(fakeDbClient.VolumeExistsCallCount()).To(Equal(0))
		})

		It("should fail when the dbClient fails to check the volume", func() {
			fakeLock.LockReturns(nil)
			fakeDbClient.VolumeExistsReturns(false, fmt.Errorf("failed checking volume"))
			err = client.RemoveVolume("fake-volume", false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("failed checking volume"))
			Expect(fakeDbClient.VolumeExistsCallCount()).To(Equal(1))
			Expect(fakeDbClient.GetVolumeCallCount()).To(Equal(0))
		})

		It("should fail when the dbClient does not find the volume", func() {
			fakeLock.LockReturns(nil)
			fakeDbClient.VolumeExistsReturns(false, nil)
			err = client.RemoveVolume("fake-volume", false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Volume not found"))
			Expect(fakeDbClient.VolumeExistsCallCount()).To(Equal(1))
			Expect(fakeDbClient.GetVolumeCallCount()).To(Equal(0))
		})

		It("should fail when the dbClient fails to get the volume", func() {
			fakeLock.LockReturns(nil)
			fakeDbClient.VolumeExistsReturns(true, nil)
			fakeDbClient.GetVolumeReturns(utils.Volume{}, fmt.Errorf("error getting volume"))
			err = client.RemoveVolume("fake-volume", false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("error getting volume"))
			Expect(fakeDbClient.VolumeExistsCallCount()).To(Equal(1))
			Expect(fakeDbClient.GetVolumeCallCount()).To(Equal(1))
			Expect(fakeDbClient.DeleteVolumeCallCount()).To(Equal(0))
		})

		It("should fail when type is lightweight and dbClient fails to delete the volume", func() {
			fakeLock.LockReturns(nil)
			fakeDbClient.VolumeExistsReturns(true, nil)
			volume := utils.Volume{Name: "fake-volume", Type: 1}
			fakeDbClient.GetVolumeReturns(volume, nil)
			fakeDbClient.DeleteVolumeReturns(fmt.Errorf("error deleting volume"))
			err = client.RemoveVolume("fake-volume", false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("error deleting volume"))
			Expect(fakeDbClient.VolumeExistsCallCount()).To(Equal(1))
			Expect(fakeDbClient.GetVolumeCallCount()).To(Equal(1))
			Expect(fakeDbClient.DeleteVolumeCallCount()).To(Equal(1))
			Expect(fakeSpectrumClient.GetFilesystemMountpointCallCount()).To(Equal(0))
		})

		It("should fail when type is lightweight and forcedelete is true and spectrumClient fails to get filesystem mountpoint", func() {
			fakeLock.LockReturns(nil)
			fakeDbClient.VolumeExistsReturns(true, nil)
			volume := utils.Volume{Name: "fake-volume", Type: 1}
			fakeDbClient.GetVolumeReturns(volume, nil)
			fakeDbClient.DeleteVolumeReturns(nil)
			fakeSpectrumClient.GetFilesystemMountpointReturns("", fmt.Errorf("error getting fs mountpoint"))
			err = client.RemoveVolume("fake-volume", true)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("error getting fs mountpoint"))
			Expect(fakeDbClient.VolumeExistsCallCount()).To(Equal(1))
			Expect(fakeDbClient.GetVolumeCallCount()).To(Equal(1))
			Expect(fakeDbClient.DeleteVolumeCallCount()).To(Equal(1))
			Expect(fakeSpectrumClient.GetFilesystemMountpointCallCount()).To(Equal(1))
			Expect(fakeExec.RemoveAllCallCount()).To(Equal(0))
		})

		It("should fail when type is lightweight and forcedelete is true and executor fails to remove volume folder", func() {
			fakeLock.LockReturns(nil)
			fakeDbClient.VolumeExistsReturns(true, nil)
			volume := utils.Volume{Name: "fake-volume", Type: 1}
			fakeDbClient.GetVolumeReturns(volume, nil)
			fakeDbClient.DeleteVolumeReturns(nil)
			fakeSpectrumClient.GetFilesystemMountpointReturns("fake-mountpoint", nil)
			fakeExec.RemoveAllReturns(fmt.Errorf("error removing path"))
			err = client.RemoveVolume("fake-volume", true)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("error removing path"))
			Expect(fakeDbClient.VolumeExistsCallCount()).To(Equal(1))
			Expect(fakeDbClient.GetVolumeCallCount()).To(Equal(1))
			Expect(fakeDbClient.DeleteVolumeCallCount()).To(Equal(1))
			Expect(fakeSpectrumClient.GetFilesystemMountpointCallCount()).To(Equal(1))
			Expect(fakeExec.RemoveAllCallCount()).To(Equal(1))
		})

		It("should succeed when type is lightweight and forcedelete is true", func() {
			fakeLock.LockReturns(nil)
			fakeDbClient.VolumeExistsReturns(true, nil)
			volume := utils.Volume{Name: "fake-volume", Type: 1}
			fakeDbClient.GetVolumeReturns(volume, nil)
			fakeDbClient.DeleteVolumeReturns(nil)
			fakeSpectrumClient.GetFilesystemMountpointReturns("fake-mountpoint", nil)
			fakeExec.RemoveAllReturns(nil)
			err = client.RemoveVolume("fake-volume", true)
			Expect(err).ToNot(HaveOccurred())
			Expect(fakeDbClient.VolumeExistsCallCount()).To(Equal(1))
			Expect(fakeDbClient.GetVolumeCallCount()).To(Equal(1))
			Expect(fakeDbClient.DeleteVolumeCallCount()).To(Equal(1))
			Expect(fakeSpectrumClient.GetFilesystemMountpointCallCount()).To(Equal(1))
			Expect(fakeExec.RemoveAllCallCount()).To(Equal(1))
		})

		It("should succeed when type is lightweight and forcedelete is false", func() {
			fakeLock.LockReturns(nil)
			fakeDbClient.VolumeExistsReturns(true, nil)
			volume := utils.Volume{Name: "fake-volume", Type: 1}
			fakeDbClient.GetVolumeReturns(volume, nil)
			fakeDbClient.DeleteVolumeReturns(nil)
			err = client.RemoveVolume("fake-volume", false)
			Expect(err).ToNot(HaveOccurred())
			Expect(fakeDbClient.VolumeExistsCallCount()).To(Equal(1))
			Expect(fakeDbClient.GetVolumeCallCount()).To(Equal(1))
			Expect(fakeDbClient.DeleteVolumeCallCount()).To(Equal(1))
			Expect(fakeSpectrumClient.GetFilesystemMountpointCallCount()).To(Equal(0))
			Expect(fakeExec.RemoveAllCallCount()).To(Equal(0))
		})

		It("should fail when type is fileset and spectrumClient fails to check filesetLinked", func() {
			fakeLock.LockReturns(nil)
			fakeDbClient.VolumeExistsReturns(true, nil)
			volume := utils.Volume{Name: "fake-volume", FileSystem: "fake-filesystem", Type: 0}
			fakeDbClient.GetVolumeReturns(volume, nil)
			fakeSpectrumClient.IsFilesetLinkedReturns(false, fmt.Errorf("error in IsFilesetLinked"))
			err = client.RemoveVolume("fake-volume", false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("error in IsFilesetLinked"))
			Expect(fakeDbClient.VolumeExistsCallCount()).To(Equal(1))
			Expect(fakeDbClient.GetVolumeCallCount()).To(Equal(1))
			Expect(fakeDbClient.DeleteVolumeCallCount()).To(Equal(0))
			Expect(fakeSpectrumClient.IsFilesetLinkedCallCount()).To(Equal(1))
			Expect(fakeSpectrumClient.UnlinkFilesetCallCount()).To(Equal(0))
		})

		It("should fail when type is fileset and fileset is linked and spectrumClient fails to unlink fileset", func() {
			fakeLock.LockReturns(nil)
			fakeDbClient.VolumeExistsReturns(true, nil)
			volume := utils.Volume{Name: "fake-volume", FileSystem: "fake-filesystem", Type: 0}
			fakeDbClient.GetVolumeReturns(volume, nil)
			fakeSpectrumClient.IsFilesetLinkedReturns(true, nil)
			fakeSpectrumClient.UnlinkFilesetReturns(fmt.Errorf("error in UnlinkFileset"))
			err = client.RemoveVolume("fake-volume", false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("error in UnlinkFileset"))
			Expect(fakeDbClient.VolumeExistsCallCount()).To(Equal(1))
			Expect(fakeDbClient.GetVolumeCallCount()).To(Equal(1))
			Expect(fakeDbClient.DeleteVolumeCallCount()).To(Equal(0))
			Expect(fakeSpectrumClient.IsFilesetLinkedCallCount()).To(Equal(1))
			Expect(fakeSpectrumClient.UnlinkFilesetCallCount()).To(Equal(1))
		})

		It("should fail when type is fileset and dbClient fails to delete fileset", func() {
			fakeLock.LockReturns(nil)
			fakeDbClient.VolumeExistsReturns(true, nil)
			volume := utils.Volume{Name: "fake-volume", FileSystem: "fake-filesystem", Type: 0}
			fakeDbClient.GetVolumeReturns(volume, nil)
			fakeSpectrumClient.IsFilesetLinkedReturns(false, nil)
			fakeDbClient.DeleteVolumeReturns(fmt.Errorf("error deleting volume"))
			err = client.RemoveVolume("fake-volume", false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("error deleting volume"))
			Expect(fakeDbClient.VolumeExistsCallCount()).To(Equal(1))
			Expect(fakeDbClient.GetVolumeCallCount()).To(Equal(1))
			Expect(fakeDbClient.DeleteVolumeCallCount()).To(Equal(1))
			Expect(fakeSpectrumClient.DeleteFilesetCallCount()).To(Equal(0))
		})

		It("should fail when type is fileset and forceDelete is true and spectrumClient fails to delete fileset", func() {
			fakeLock.LockReturns(nil)
			fakeDbClient.VolumeExistsReturns(true, nil)
			volume := utils.Volume{Name: "fake-volume", FileSystem: "fake-filesystem", Type: 0}
			fakeDbClient.GetVolumeReturns(volume, nil)
			fakeSpectrumClient.IsFilesetLinkedReturns(false, nil)
			fakeDbClient.DeleteVolumeReturns(nil)
			fakeSpectrumClient.DeleteFilesetReturns(fmt.Errorf("error deleting fileset"))
			err = client.RemoveVolume("fake-volume", true)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("error deleting fileset"))
			Expect(fakeDbClient.VolumeExistsCallCount()).To(Equal(1))
			Expect(fakeDbClient.GetVolumeCallCount()).To(Equal(1))
			Expect(fakeDbClient.DeleteVolumeCallCount()).To(Equal(1))
			Expect(fakeSpectrumClient.DeleteFilesetCallCount()).To(Equal(1))
		})

		It("should succeed when type is fileset and forceDelete is true", func() {
			fakeLock.LockReturns(nil)
			fakeDbClient.VolumeExistsReturns(true, nil)
			volume := utils.Volume{Name: "fake-volume", FileSystem: "fake-filesystem", Type: 0}
			fakeDbClient.GetVolumeReturns(volume, nil)
			fakeSpectrumClient.IsFilesetLinkedReturns(false, nil)
			fakeDbClient.DeleteVolumeReturns(nil)
			fakeSpectrumClient.DeleteFilesetReturns(nil)
			err = client.RemoveVolume("fake-volume", true)
			Expect(err).ToNot(HaveOccurred())
			Expect(fakeDbClient.VolumeExistsCallCount()).To(Equal(1))
			Expect(fakeDbClient.GetVolumeCallCount()).To(Equal(1))
			Expect(fakeDbClient.DeleteVolumeCallCount()).To(Equal(1))
			Expect(fakeSpectrumClient.DeleteFilesetCallCount()).To(Equal(1))
		})

		It("should succeed when type is fileset and forceDelete is false", func() {
			fakeLock.LockReturns(nil)
			fakeDbClient.VolumeExistsReturns(true, nil)
			volume := utils.Volume{Name: "fake-volume", FileSystem: "fake-filesystem", Type: 0}
			fakeDbClient.GetVolumeReturns(volume, nil)
			fakeSpectrumClient.IsFilesetLinkedReturns(false, nil)
			fakeDbClient.DeleteVolumeReturns(nil)
			fakeSpectrumClient.DeleteFilesetReturns(nil)
			err = client.RemoveVolume("fake-volume", false)
			Expect(err).ToNot(HaveOccurred())
			Expect(fakeDbClient.VolumeExistsCallCount()).To(Equal(1))
			Expect(fakeDbClient.GetVolumeCallCount()).To(Equal(1))
			Expect(fakeDbClient.DeleteVolumeCallCount()).To(Equal(1))
			Expect(fakeSpectrumClient.DeleteFilesetCallCount()).To(Equal(0))
		})

	})

	Context(".ListVolumes", func() {
		BeforeEach(func() {})
		It("should fail when fileLock fails to aquire the lock", func() {
			fakeLock.LockReturns(fmt.Errorf("error aquiring the lock"))
			volumes, err := client.ListVolumes()
			Expect(err).To(HaveOccurred())
			Expect(len(volumes)).To(Equal(0))
			Expect(err.Error()).To(Equal("error aquiring the lock"))
			Expect(fakeLock.LockCallCount()).To(Equal(1))
			Expect(fakeDbClient.ListVolumesCallCount()).To(Equal(0))
		})
		It("should fail when dbClient fails to list volumes", func() {
			fakeLock.LockReturns(nil)
			fakeDbClient.ListVolumesReturns(nil, fmt.Errorf("error listing volumes"))
			volumes, err := client.ListVolumes()
			Expect(err).To(HaveOccurred())
			Expect(len(volumes)).To(Equal(0))
			Expect(err.Error()).To(Equal("error listing volumes"))
			Expect(fakeLock.LockCallCount()).To(Equal(1))
			Expect(fakeDbClient.ListVolumesCallCount()).To(Equal(1))
		})
		It("should succeed to list volumes", func() {
			fakeLock.LockReturns(nil)

			volume1 := utils.Volume{Name: "fake-volume1", FileSystem: "fake-filesystem", Mountpoint: "fake-mountpoint"}
			volume2 := utils.Volume{Name: "fake-volume2", FileSystem: "fake-filesystem", Mountpoint: "fake-mountpoint"}
			volumesList := make([]utils.Volume, 2)
			volumesList[0] = volume1
			volumesList[1] = volume2
			fakeDbClient.ListVolumesReturns(volumesList, nil)
			volumes, err := client.ListVolumes()
			Expect(err).ToNot(HaveOccurred())
			Expect(len(volumes)).To(Equal(2))
			Expect(fakeLock.LockCallCount()).To(Equal(1))
			Expect(fakeDbClient.ListVolumesCallCount()).To(Equal(1))
		})

	})

	Context("GetVolume", func() {
		It("should fail when fileLock fails to aquire the lock", func() {
			fakeLock.LockReturns(fmt.Errorf("error aquiring the lock"))
			_, _, err = client.GetVolume("fake-volume")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("error aquiring the lock"))
			Expect(fakeLock.LockCallCount()).To(Equal(1))
			Expect(fakeDbClient.VolumeExistsCallCount()).To(Equal(0))
		})

		It("should fail when dbClient fails to check if the volume exists", func() {
			fakeLock.LockReturns(nil)
			fakeDbClient.VolumeExistsReturns(false, fmt.Errorf("error checking volume"))
			_, _, err = client.GetVolume("fake-volume")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("error checking volume"))
			Expect(fakeLock.LockCallCount()).To(Equal(1))
			Expect(fakeDbClient.VolumeExistsCallCount()).To(Equal(1))
		})

		It("should fail when volume exists and dbClient fails to getVolume", func() {
			fakeLock.LockReturns(nil)
			fakeDbClient.VolumeExistsReturns(true, nil)
			fakeDbClient.GetVolumeReturns(utils.Volume{}, fmt.Errorf("error getting volume"))
			_, _, err = client.GetVolume("fake-volume")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("error getting volume"))
			Expect(fakeLock.LockCallCount()).To(Equal(1))
			Expect(fakeDbClient.VolumeExistsCallCount()).To(Equal(1))
			Expect(fakeDbClient.GetVolumeCallCount()).To(Equal(1))
		})

		It("should fail when volume does not exist", func() {
			fakeLock.LockReturns(nil)
			fakeDbClient.VolumeExistsReturns(false, nil)
			_, _, err = client.GetVolume("fake-volume")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Volume not found"))
			Expect(fakeLock.LockCallCount()).To(Equal(1))
			Expect(fakeDbClient.VolumeExistsCallCount()).To(Equal(1))
		})

		It("should succeed  when volume exists", func() {
			fakeLock.LockReturns(nil)
			fakeDbClient.VolumeExistsReturns(true, nil)

			volume := utils.Volume{Name: "fake-volume", FileSystem: "fake-filesystem", Mountpoint: "fake-mountpoint"}
			fakeDbClient.GetVolumeReturns(volume, nil)
			vol, _, err := client.GetVolume("fake-volume")
			Expect(err).ToNot(HaveOccurred())
			Expect(vol.Name).To(Equal("fake-volume"))
			Expect(vol.Mountpoint).To(Equal("fake-mountpoint"))
			Expect(fakeLock.LockCallCount()).To(Equal(1))
			Expect(fakeDbClient.VolumeExistsCallCount()).To(Equal(1))
			Expect(fakeDbClient.GetVolumeCallCount()).To(Equal(1))
		})

	})

	Context(".Attach", func() {
		It("should fail when fileLock fails to aquire the lock", func() {
			fakeLock.LockReturns(fmt.Errorf("error aquiring the lock"))
			mountpath, err := client.Attach("fake-volume")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("error aquiring the lock"))
			Expect(mountpath).To(Equal(""))
			Expect(fakeLock.LockCallCount()).To(Equal(1))
			Expect(fakeDbClient.VolumeExistsCallCount()).To(Equal(0))
		})

		It("should fail when dbClient fails to check volumeExists", func() {
			fakeLock.LockReturns(nil)
			fakeDbClient.VolumeExistsReturns(false, fmt.Errorf("error in checking volume"))
			mountpath, err := client.Attach("fake-volume")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("error in checking volume"))
			Expect(mountpath).To(Equal(""))
			Expect(fakeLock.LockCallCount()).To(Equal(1))
			Expect(fakeDbClient.VolumeExistsCallCount()).To(Equal(1))
			Expect(fakeDbClient.GetVolumeCallCount()).To(Equal(0))
		})

		It("should fail when volume does not exist", func() {
			fakeLock.LockReturns(nil)
			fakeDbClient.VolumeExistsReturns(false, nil)
			mountpath, err := client.Attach("fake-volume")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Volume not found"))
			Expect(mountpath).To(Equal(""))
			Expect(fakeLock.LockCallCount()).To(Equal(1))
			Expect(fakeDbClient.VolumeExistsCallCount()).To(Equal(1))
			Expect(fakeDbClient.GetVolumeCallCount()).To(Equal(0))
		})

		It("should fail when volume exists and dbClient fails to getVolume", func() {
			fakeLock.LockReturns(nil)
			fakeDbClient.VolumeExistsReturns(true, nil)
			fakeDbClient.GetVolumeReturns(utils.Volume{}, fmt.Errorf("error getting volume"))
			mountpath, err := client.Attach("fake-volume")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("error getting volume"))
			Expect(mountpath).To(Equal(""))
			Expect(fakeLock.LockCallCount()).To(Equal(1))
			Expect(fakeDbClient.VolumeExistsCallCount()).To(Equal(1))
			Expect(fakeDbClient.GetVolumeCallCount()).To(Equal(1))
			Expect(fakeSpectrumClient.GetFilesystemMountpointCallCount()).To(Equal(0))
		})

		It("should fail when volume is not attached and dbClient fails to get filesystem mountpoint", func() {
			fakeLock.LockReturns(nil)
			fakeDbClient.VolumeExistsReturns(true, nil)
			volume := utils.Volume{Name: "fake-volume", FileSystem: "fake-filesystem"}
			fakeDbClient.GetVolumeReturns(volume, nil)
			fakeSpectrumClient.GetFilesystemMountpointReturns("", fmt.Errorf("error getting mountpoint"))
			mountpath, err := client.Attach("fake-volume")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("error getting mountpoint"))
			Expect(mountpath).To(Equal(""))
			Expect(fakeLock.LockCallCount()).To(Equal(1))
			Expect(fakeDbClient.VolumeExistsCallCount()).To(Equal(1))
			Expect(fakeDbClient.GetVolumeCallCount()).To(Equal(1))
			Expect(fakeSpectrumClient.GetFilesystemMountpointCallCount()).To(Equal(1))
			Expect(fakeSpectrumClient.IsFilesetLinkedCallCount()).To(Equal(0))
		})

		It("should fail when volume is fileset volume and spectrumClient fails to check fileset linked", func() {
			fakeLock.LockReturns(nil)
			fakeDbClient.VolumeExistsReturns(true, nil)
			volume := utils.Volume{Name: "fake-volume", FileSystem: "fake-filesystem", Type: utils.FILESET}
			fakeDbClient.GetVolumeReturns(volume, nil)
			fakeSpectrumClient.GetFilesystemMountpointReturns("fake-mountpoint", nil)
			fakeSpectrumClient.IsFilesetLinkedReturns(false, fmt.Errorf("error checking filesetlinked"))
			mountpath, err := client.Attach("fake-volume")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("error checking filesetlinked"))
			Expect(mountpath).To(Equal(""))
			Expect(fakeLock.LockCallCount()).To(Equal(1))
			Expect(fakeDbClient.VolumeExistsCallCount()).To(Equal(1))
			Expect(fakeDbClient.GetVolumeCallCount()).To(Equal(1))
			Expect(fakeSpectrumClient.GetFilesystemMountpointCallCount()).To(Equal(1))
			Expect(fakeSpectrumClient.IsFilesetLinkedCallCount()).To(Equal(1))
			Expect(fakeSpectrumClient.LinkFilesetCallCount()).To(Equal(0))
		})

		It("should fail when volume is fileset volume and spectrumClient fails to link it", func() {
			fakeLock.LockReturns(nil)
			fakeDbClient.VolumeExistsReturns(true, nil)
			volume := utils.Volume{Name: "fake-volume", FileSystem: "fake-filesystem", Type: utils.FILESET}
			fakeDbClient.GetVolumeReturns(volume, nil)
			fakeSpectrumClient.GetFilesystemMountpointReturns("fake-mountpoint", nil)
			fakeSpectrumClient.IsFilesetLinkedReturns(false, nil)
			fakeSpectrumClient.LinkFilesetReturns(fmt.Errorf("error linking fileset"))
			mountpath, err := client.Attach("fake-volume")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("error linking fileset"))
			Expect(mountpath).To(Equal(""))
			Expect(fakeLock.LockCallCount()).To(Equal(1))
			Expect(fakeDbClient.VolumeExistsCallCount()).To(Equal(1))
			Expect(fakeDbClient.GetVolumeCallCount()).To(Equal(1))
			Expect(fakeSpectrumClient.GetFilesystemMountpointCallCount()).To(Equal(1))
			Expect(fakeSpectrumClient.IsFilesetLinkedCallCount()).To(Equal(1))
			Expect(fakeSpectrumClient.LinkFilesetCallCount()).To(Equal(1))
		})

		It("should fail when volume is fileset volume with permissions and executor fails to execute permissions change", func() {
			fakeLock.LockReturns(nil)
			fakeDbClient.VolumeExistsReturns(true, nil)
			data := make(map[string]string)
			data["uid"] = "fake-uid"
			data["gid"] = "gid"
			volume := utils.Volume{Name: "fake-volume", FileSystem: "fake-filesystem", Type: utils.FILESET, AdditionalData: data}
			fakeDbClient.GetVolumeReturns(volume, nil)
			fakeSpectrumClient.GetFilesystemMountpointReturns("fake-mountpoint", nil)
			fakeSpectrumClient.IsFilesetLinkedReturns(false, nil)
			fakeSpectrumClient.LinkFilesetReturns(nil)
			fakeExec.ExecuteReturns(nil, fmt.Errorf("error executing command"))
			mountpath, err := client.Attach("fake-volume")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("error executing command"))
			Expect(mountpath).To(Equal(""))
			Expect(fakeLock.LockCallCount()).To(Equal(1))
			Expect(fakeDbClient.VolumeExistsCallCount()).To(Equal(1))
			Expect(fakeDbClient.GetVolumeCallCount()).To(Equal(1))
			Expect(fakeSpectrumClient.GetFilesystemMountpointCallCount()).To(Equal(2))
			Expect(fakeSpectrumClient.IsFilesetLinkedCallCount()).To(Equal(1))
			Expect(fakeSpectrumClient.LinkFilesetCallCount()).To(Equal(1))
			Expect(fakeExec.ExecuteCallCount()).To(Equal(1))
			Expect(fakeDbClient.UpdateVolumeMountpointCallCount()).To(Equal(0))
		})

		It("should fail when volume is lightweight volume with permissions and dbClient fails to update volume mountpoint", func() {
			fakeLock.LockReturns(nil)
			fakeDbClient.VolumeExistsReturns(true, nil)
			data := make(map[string]string)
			data["uid"] = "fake-uid"
			data["gid"] = "gid"
			volume := utils.Volume{Name: "fake-volume", FileSystem: "fake-filesystem", Type: utils.LIGHTWEIGHT, AdditionalData: data}
			fakeDbClient.GetVolumeReturns(volume, nil)
			fakeSpectrumClient.GetFilesystemMountpointReturns("fake-mountpoint", nil)
			fakeExec.ExecuteReturns(nil, nil)
			fakeDbClient.UpdateVolumeMountpointReturns(fmt.Errorf("error updating volume record"))
			mountpath, err := client.Attach("fake-volume")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("error updating volume record"))
			Expect(mountpath).To(Equal(""))
			Expect(fakeLock.LockCallCount()).To(Equal(1))
			Expect(fakeDbClient.VolumeExistsCallCount()).To(Equal(1))
			Expect(fakeDbClient.GetVolumeCallCount()).To(Equal(1))
			Expect(fakeSpectrumClient.GetFilesystemMountpointCallCount()).To(Equal(2))
			Expect(fakeSpectrumClient.IsFilesetLinkedCallCount()).To(Equal(0))
			Expect(fakeSpectrumClient.LinkFilesetCallCount()).To(Equal(0))
			Expect(fakeExec.ExecuteCallCount()).To(Equal(1))
			Expect(fakeDbClient.UpdateVolumeMountpointCallCount()).To(Equal(1))
		})

		It("should succeed when volume is lightweight volume with permissions", func() {
			fakeLock.LockReturns(nil)
			fakeDbClient.VolumeExistsReturns(true, nil)
			data := make(map[string]string)
			data["uid"] = "fake-uid"
			data["gid"] = "gid"
			volume := utils.Volume{Name: "fake-volume", FileSystem: "fake-filesystem", Type: utils.LIGHTWEIGHT, AdditionalData: data}
			fakeDbClient.GetVolumeReturns(volume, nil)
			fakeSpectrumClient.GetFilesystemMountpointReturns("fake-mountpoint", nil)
			fakeSpectrumClient.IsFilesetLinkedReturns(false, nil)
			fakeSpectrumClient.LinkFilesetReturns(nil)
			fakeExec.ExecuteReturns(nil, nil)
			mountpath, err := client.Attach("fake-volume")
			Expect(err).ToNot(HaveOccurred())
			Expect(mountpath).To(Equal("fake-mountpoint"))
			Expect(fakeLock.LockCallCount()).To(Equal(1))
			Expect(fakeDbClient.VolumeExistsCallCount()).To(Equal(1))
			Expect(fakeDbClient.GetVolumeCallCount()).To(Equal(1))
			Expect(fakeSpectrumClient.GetFilesystemMountpointCallCount()).To(Equal(2))
			Expect(fakeSpectrumClient.IsFilesetLinkedCallCount()).To(Equal(0))
			Expect(fakeSpectrumClient.LinkFilesetCallCount()).To(Equal(0))
			Expect(fakeExec.ExecuteCallCount()).To(Equal(1))
			Expect(fakeDbClient.UpdateVolumeMountpointCallCount()).To(Equal(1))
		})

		It("should succeed when volume is fileset volume with permissions", func() {
			fakeLock.LockReturns(nil)
			fakeDbClient.VolumeExistsReturns(true, nil)
			data := make(map[string]string)
			data["uid"] = "fake-uid"
			data["gid"] = "gid"
			volume := utils.Volume{Name: "fake-volume", FileSystem: "fake-filesystem", Type: utils.FILESET, AdditionalData: data}
			fakeDbClient.GetVolumeReturns(volume, nil)
			fakeSpectrumClient.GetFilesystemMountpointReturns("fake-mountpoint", nil)
			fakeSpectrumClient.IsFilesetLinkedReturns(false, nil)
			fakeSpectrumClient.LinkFilesetReturns(nil)
			fakeExec.ExecuteReturns(nil, nil)
			mountpath, err := client.Attach("fake-volume")
			Expect(err).ToNot(HaveOccurred())
			Expect(mountpath).To(Equal("fake-mountpoint"))
			Expect(fakeLock.LockCallCount()).To(Equal(1))
			Expect(fakeDbClient.VolumeExistsCallCount()).To(Equal(1))
			Expect(fakeDbClient.GetVolumeCallCount()).To(Equal(1))
			Expect(fakeSpectrumClient.GetFilesystemMountpointCallCount()).To(Equal(2))
			Expect(fakeSpectrumClient.IsFilesetLinkedCallCount()).To(Equal(1))
			Expect(fakeSpectrumClient.LinkFilesetCallCount()).To(Equal(1))
			Expect(fakeExec.ExecuteCallCount()).To(Equal(1))
			Expect(fakeDbClient.UpdateVolumeMountpointCallCount()).To(Equal(1))
		})

		It("should succeed when volume is already attached", func() {
			fakeLock.LockReturns(nil)
			fakeDbClient.VolumeExistsReturns(true, nil)
			volume := utils.Volume{Name: "fake-volume", FileSystem: "fake-filesystem", Mountpoint: "fake-mountpoint"}
			fakeDbClient.GetVolumeReturns(volume, nil)
			mountpath, err := client.Attach("fake-volume")
			Expect(err).ToNot(HaveOccurred())
			Expect(mountpath).To(Equal("fake-mountpoint"))
			Expect(fakeLock.LockCallCount()).To(Equal(1))
			Expect(fakeDbClient.VolumeExistsCallCount()).To(Equal(1))
			Expect(fakeDbClient.GetVolumeCallCount()).To(Equal(1))
		})

	})

	Context(".Detach", func() {
		It("should fail when fileLock fails to aquire the lock", func() {
			fakeLock.LockReturns(fmt.Errorf("error aquiring the lock"))
			err = client.Detach("fake-volume")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("error aquiring the lock"))
			Expect(fakeLock.LockCallCount()).To(Equal(1))
			Expect(fakeDbClient.VolumeExistsCallCount()).To(Equal(0))
		})

		It("should fail when dbClient fails to check volumeExists", func() {
			fakeLock.LockReturns(nil)
			fakeDbClient.VolumeExistsReturns(false, fmt.Errorf("error in checking volume"))
			err = client.Detach("fake-volume")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("error in checking volume"))
			Expect(fakeLock.LockCallCount()).To(Equal(1))
			Expect(fakeDbClient.VolumeExistsCallCount()).To(Equal(1))
			Expect(fakeDbClient.GetVolumeCallCount()).To(Equal(0))
		})

		It("should fail when volume does not exist", func() {
			fakeLock.LockReturns(nil)
			fakeDbClient.VolumeExistsReturns(false, nil)
			err = client.Detach("fake-volume")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Volume not found"))
			Expect(fakeLock.LockCallCount()).To(Equal(1))
			Expect(fakeDbClient.VolumeExistsCallCount()).To(Equal(1))
			Expect(fakeDbClient.GetVolumeCallCount()).To(Equal(0))
		})

		It("should fail when volume exists and dbClient fails to getVolume", func() {
			fakeLock.LockReturns(nil)
			fakeDbClient.VolumeExistsReturns(true, nil)
			fakeDbClient.GetVolumeReturns(utils.Volume{}, fmt.Errorf("error getting volume"))
			err = client.Detach("fake-volume")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("error getting volume"))
			Expect(fakeLock.LockCallCount()).To(Equal(1))
			Expect(fakeDbClient.VolumeExistsCallCount()).To(Equal(1))
			Expect(fakeDbClient.GetVolumeCallCount()).To(Equal(1))
		})

		It("should fail when volume exists but not attached", func() {
			fakeLock.LockReturns(nil)
			fakeDbClient.VolumeExistsReturns(true, nil)

			volume := utils.Volume{Name: "fake-volume", FileSystem: "fake-filesystem"}
			fakeDbClient.GetVolumeReturns(volume, nil)
			err := client.Detach("fake-volume")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("volume not attached"))
			Expect(fakeLock.LockCallCount()).To(Equal(1))
			Expect(fakeDbClient.VolumeExistsCallCount()).To(Equal(1))
			Expect(fakeDbClient.GetVolumeCallCount()).To(Equal(1))
			Expect(fakeDbClient.UpdateVolumeMountpointCallCount()).To(Equal(0))
		})

		It("should fail when dbClient fails to update volume record", func() {
			fakeLock.LockReturns(nil)
			fakeDbClient.VolumeExistsReturns(true, nil)
			volume := utils.Volume{Name: "fake-volume", FileSystem: "fake-filesystem", Mountpoint: "fake-mountpoint"}
			fakeDbClient.GetVolumeReturns(volume, nil)
			fakeDbClient.UpdateVolumeMountpointReturns(fmt.Errorf("error updating volume mountpoint"))
			err := client.Detach("fake-volume")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("error updating volume mountpoint"))
			Expect(fakeLock.LockCallCount()).To(Equal(1))
			Expect(fakeDbClient.VolumeExistsCallCount()).To(Equal(1))
			Expect(fakeDbClient.GetVolumeCallCount()).To(Equal(1))
			Expect(fakeDbClient.UpdateVolumeMountpointCallCount()).To(Equal(1))
		})

		It("should succeed when everything is all right", func() {
			fakeLock.LockReturns(nil)
			fakeDbClient.VolumeExistsReturns(true, nil)

			volume := utils.Volume{Name: "fake-volume", FileSystem: "fake-filesystem", Mountpoint: "fake-mountpoint"}
			fakeDbClient.GetVolumeReturns(volume, nil)
			err := client.Detach("fake-volume")
			Expect(err).ToNot(HaveOccurred())
			Expect(fakeLock.LockCallCount()).To(Equal(1))
			Expect(fakeDbClient.VolumeExistsCallCount()).To(Equal(1))
			Expect(fakeDbClient.GetVolumeCallCount()).To(Equal(1))
			Expect(fakeDbClient.UpdateVolumeMountpointCallCount()).To(Equal(1))
		})

	})

})
