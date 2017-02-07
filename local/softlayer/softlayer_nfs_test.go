package softlayer_test

import (
	"fmt"
	"log"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.ibm.com/almaden-containers/ubiquity/resources"

	"github.ibm.com/alchemy-containers/armada-slclient-lib/datatypes"
	"github.ibm.com/almaden-containers/ubiquity/fakes"
	"github.ibm.com/almaden-containers/ubiquity/local/softlayer"
)

var _ = Describe("local-client", func() {
	var (
		client                      resources.StorageClient
		logger                      *log.Logger
		fakeDbClient                *fakes.FakeDatabaseClient
		fakeSoftlayerDataModel      *fakes.FakeSoftlayerDataModel
		fakeLock                    *fakes.FakeFileLock
		fakeExec                    *fakes.FakeExecutor
		fakeConfig                  resources.SpectrumScaleConfig
		fakeSoftlayerStorageService *fakes.FakeSoftlayer_Storage_Service
		err                         error
	)
	BeforeEach(func() {
		logger = log.New(os.Stdout, "ubiquity: ", log.Lshortfile|log.LstdFlags)
		fakeDbClient = new(fakes.FakeDatabaseClient)
		fakeLock = new(fakes.FakeFileLock)
		fakeExec = new(fakes.FakeExecutor)
		fakeSoftlayerDataModel = new(fakes.FakeSoftlayerDataModel)
		fakeSoftlayerStorageService = new(fakes.FakeSoftlayer_Storage_Service)
		fakeConfig = resources.SpectrumScaleConfig{}
		client, err = softlayer.NewSoftlayerLocalClientWithDataModelAndSLService(logger, fakeSoftlayerDataModel, fakeLock, fakeSoftlayerStorageService)
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
		})

		It("should fail when fileLock failes to unlock", func() {
			fakeLock.LockReturns(nil)
			fakeLock.UnlockReturns(fmt.Errorf("error in unlock call"))
			err = client.Activate()
			Expect(err).To(HaveOccurred())
			Expect(fakeLock.LockCallCount()).To(Equal(1))
			Expect(fakeLock.UnlockCallCount()).To(Equal(1))
		})

		It("should fail when dbClient CreateVolumeTable errors", func() {
			fakeLock.LockReturns(nil)
			fakeLock.UnlockReturns(nil)

			fakeSoftlayerDataModel.CreateVolumeTableReturns(fmt.Errorf("error in creating volume table"))
			err = client.Activate()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("error in creating volume table"))
			Expect(fakeLock.LockCallCount()).To(Equal(1))
			Expect(fakeLock.UnlockCallCount()).To(Equal(1))
			Expect(fakeSoftlayerDataModel.CreateVolumeTableCallCount()).To(Equal(1))
		})

		It("should succeed when everything is fine with no mounting", func() {
			fakeLock.LockReturns(nil)
			fakeLock.UnlockReturns(nil)
			fakeSoftlayerDataModel.CreateVolumeTableReturns(nil)
			err = client.Activate()
			Expect(err).ToNot(HaveOccurred())
			Expect(fakeLock.LockCallCount()).To(Equal(1))
			Expect(fakeLock.UnlockCallCount()).To(Equal(1))
			Expect(fakeSoftlayerDataModel.CreateVolumeTableCallCount()).To(Equal(1))
		})

		It("should succeed when everything is fine with mounting", func() {
			fakeLock.LockReturns(nil)
			fakeLock.UnlockReturns(nil)
			fakeSoftlayerDataModel.CreateVolumeTableReturns(nil)
			err = client.Activate()
			Expect(err).ToNot(HaveOccurred())
			Expect(fakeLock.LockCallCount()).To(Equal(1))
			Expect(fakeLock.UnlockCallCount()).To(Equal(1))
			Expect(fakeSoftlayerDataModel.CreateVolumeTableCallCount()).To(Equal(1))
		})

	})
	Context(".CreateVolume", func() {
		var (
			opts map[string]interface{}
		)
		BeforeEach(func() {
			opts = make(map[string]interface{})
			opts[softlayer.SL_LOCATION] = "fake-location"
			client, err = softlayer.NewSoftlayerLocalClientWithDataModelAndSLService(logger, fakeSoftlayerDataModel, fakeLock, fakeSoftlayerStorageService)
			Expect(err).ToNot(HaveOccurred())
			fakeLock.LockReturns(nil)
			fakeLock.UnlockReturns(nil)
			fakeSoftlayerDataModel.CreateVolumeTableReturns(nil)
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
			Expect(fakeSoftlayerDataModel.GetVolumeCallCount()).To(Equal(0))

		})

		It("should fail when fileLock failes to release the lock", func() {
			fakeLock.LockReturns(nil)
			fakeLock.UnlockReturns(fmt.Errorf("error in unlock call"))

			err = client.CreateVolume("fake-volume", opts)
			Expect(err).To(HaveOccurred())
			Expect(fakeLock.LockCallCount()).To(Equal(2))
			Expect(fakeLock.UnlockCallCount()).To(Equal(2))
			Expect(fakeSoftlayerDataModel.GetVolumeCallCount()).To(Equal(1))

		})

		It("should fail when dbClient volumeExists errors", func() {
			fakeLock.LockReturns(nil)
			fakeLock.UnlockReturns(nil)
			fakeSoftlayerDataModel.GetVolumeReturns(softlayer.SoftlayerVolume{}, false, fmt.Errorf("error checking if volume exists"))
			err = client.CreateVolume("fake-volume", opts)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("error checking if volume exists"))
			Expect(fakeLock.LockCallCount()).To(Equal(2))
			Expect(fakeLock.UnlockCallCount()).To(Equal(2))
			Expect(fakeSoftlayerDataModel.GetVolumeCallCount()).To(Equal(1))
			Expect(fakeSoftlayerStorageService.CreateStorageCallCount()).To(Equal(0))
		})

		It("should fail when dbClient volumeExists returns true", func() {
			fakeLock.LockReturns(nil)
			fakeLock.UnlockReturns(nil)
			fakeSoftlayerDataModel.GetVolumeReturns(softlayer.SoftlayerVolume{}, true, nil)
			err = client.CreateVolume("fake-volume", opts)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Volume already exists"))
			Expect(fakeLock.LockCallCount()).To(Equal(2))
			Expect(fakeLock.UnlockCallCount()).To(Equal(2))
			Expect(fakeSoftlayerDataModel.GetVolumeCallCount()).To(Equal(1))
			Expect(fakeSoftlayerStorageService.CreateStorageCallCount()).To(Equal(0))
		})

		It("should fail when softlayer client fails to create fileshare", func() {
			fakeSoftlayerStorageService.CreateStorageReturns(datatypes.SoftLayer_Storage{}, fmt.Errorf("error creating fileset"))
			err = client.CreateVolume("fake-fileset", opts)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("error creating fileset"))
			Expect(fakeSoftlayerStorageService.CreateStorageCallCount()).To(Equal(1))
			Expect(fakeSoftlayerDataModel.InsertFileshareCallCount()).To(Equal(0))
		})

		It("should fail when dbClient fails to insert fileshare record", func() {
			fakeShare := datatypes.SoftLayer_Storage{Id: 1, Username: "fake-username", ServiceResourceBackendIpAddress: "fake-ip"}
			fakeSoftlayerStorageService.CreateStorageReturns(fakeShare, nil)
			fakeSoftlayerDataModel.InsertFileshareReturns(fmt.Errorf("error inserting fileshare"))

			err = client.CreateVolume("fake-fileshare", opts)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("error inserting fileshare"))
			Expect(fakeSoftlayerStorageService.CreateStorageCallCount()).To(Equal(1))
			Expect(fakeSoftlayerDataModel.InsertFileshareCallCount()).To(Equal(1))
		})

		It("should succeed to create fileshare", func() {
			fakeShare := datatypes.SoftLayer_Storage{Id: 1, Username: "fake-username", ServiceResourceBackendIpAddress: "fake-ip"}
			fakeSoftlayerStorageService.CreateStorageReturns(fakeShare, nil)
			fakeSoftlayerDataModel.InsertFileshareReturns(nil)

			err = client.CreateVolume("fake-fileshare", opts)
			Expect(err).ToNot(HaveOccurred())
			Expect(fakeSoftlayerStorageService.CreateStorageCallCount()).To(Equal(1))
			Expect(fakeSoftlayerDataModel.InsertFileshareCallCount()).To(Equal(1))
		})

	})
	Context(".RemoveVolume", func() {
		It("should fail when the fileLock fails to aquire the lock", func() {
			fakeLock.LockReturns(fmt.Errorf("failed to aquire lock"))
			err = client.RemoveVolume("fake-volume", false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("failed to aquire lock"))
			Expect(fakeSoftlayerDataModel.GetVolumeCallCount()).To(Equal(0))
		})

		It("should fail when the dbClient fails to check the volume", func() {
			fakeLock.LockReturns(nil)
			fakeSoftlayerDataModel.GetVolumeReturns(softlayer.SoftlayerVolume{}, false, fmt.Errorf("failed getting volume"))
			err = client.RemoveVolume("fake-volume", false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("failed getting volume"))
		})

		It("should fail when slClient fails to removeAccess to fileshare", func() {
			fakeLock.LockReturns(nil)
			volume := softlayer.SoftlayerVolume{Name: "fake-volume"}
			fakeSoftlayerDataModel.GetVolumeReturns(volume, true, nil)
			fakeSoftlayerStorageService.RemoveAccessFromAllSubnetsReturns(false, fmt.Errorf("error in removeAccess"))
			err = client.RemoveVolume("fake-volume", false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("error in removeAccess"))
			Expect(fakeSoftlayerDataModel.GetVolumeCallCount()).To(Equal(1))
			Expect(fakeSoftlayerDataModel.DeleteVolumeCallCount()).To(Equal(0))
			Expect(fakeSoftlayerStorageService.RemoveAccessFromAllSubnetsCallCount()).To(Equal(1))
		})

		It("should fail when dbClient fails to delete fileshare", func() {
			fakeLock.LockReturns(nil)
			volume := softlayer.SoftlayerVolume{Name: "fake-volume"}
			fakeSoftlayerDataModel.GetVolumeReturns(volume, true, nil)
			fakeSoftlayerStorageService.RemoveAccessFromAllSubnetsReturns(true, nil)
			fakeSoftlayerDataModel.DeleteVolumeReturns(fmt.Errorf("error deleting volume"))
			err = client.RemoveVolume("fake-volume", false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("error deleting volume"))
			Expect(fakeSoftlayerDataModel.GetVolumeCallCount()).To(Equal(1))
			Expect(fakeSoftlayerDataModel.DeleteVolumeCallCount()).To(Equal(1))
			Expect(fakeSoftlayerStorageService.DeleteStorageCallCount()).To(Equal(0))
		})

		It("should fail when forceDelete is true and slclient fails to delete fileshare", func() {
			fakeLock.LockReturns(nil)
			volume := softlayer.SoftlayerVolume{Name: "fake-volume"}
			fakeSoftlayerDataModel.GetVolumeReturns(volume, true, nil)
			fakeSoftlayerStorageService.RemoveAccessFromAllSubnetsReturns(true, nil)
			fakeSoftlayerDataModel.DeleteVolumeReturns(nil)
			fakeSoftlayerStorageService.DeleteStorageReturns(fmt.Errorf("error deleting fileshare"))
			err = client.RemoveVolume("fake-volume", true)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("error deleting fileshare"))
			Expect(fakeSoftlayerDataModel.GetVolumeCallCount()).To(Equal(1))
			Expect(fakeSoftlayerDataModel.DeleteVolumeCallCount()).To(Equal(0))
			Expect(fakeSoftlayerStorageService.DeleteStorageCallCount()).To(Equal(1))
		})

		It("should succeed when forceDelete is true", func() {
			fakeLock.LockReturns(nil)
			volume := softlayer.SoftlayerVolume{Name: "fake-volume"}
			fakeSoftlayerDataModel.GetVolumeReturns(volume, true, nil)
			fakeSoftlayerStorageService.RemoveAccessFromAllSubnetsReturns(true, nil)
			fakeSoftlayerDataModel.DeleteVolumeReturns(nil)
			fakeSoftlayerStorageService.DeleteStorageReturns(nil)
			err = client.RemoveVolume("fake-volume", true)
			Expect(err).ToNot(HaveOccurred())
			Expect(fakeSoftlayerDataModel.GetVolumeCallCount()).To(Equal(1))
			Expect(fakeSoftlayerDataModel.DeleteVolumeCallCount()).To(Equal(1))
			Expect(fakeSoftlayerStorageService.DeleteStorageCallCount()).To(Equal(1))
		})

		It("should succeed when type is fileshare and forceDelete is false", func() {
			fakeLock.LockReturns(nil)
			volume := softlayer.SoftlayerVolume{Name: "fake-volume"}
			fakeSoftlayerDataModel.GetVolumeReturns(volume, true, nil)
			fakeSoftlayerStorageService.RemoveAccessFromAllSubnetsReturns(true, nil)
			fakeSoftlayerDataModel.DeleteVolumeReturns(nil)
			fakeSoftlayerStorageService.DeleteStorageReturns(nil)
			err = client.RemoveVolume("fake-volume", false)
			Expect(err).ToNot(HaveOccurred())
			Expect(fakeSoftlayerDataModel.GetVolumeCallCount()).To(Equal(1))
			Expect(fakeSoftlayerDataModel.DeleteVolumeCallCount()).To(Equal(1))
			Expect(fakeSoftlayerStorageService.DeleteStorageCallCount()).To(Equal(0))
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
			Expect(fakeSoftlayerDataModel.ListVolumesCallCount()).To(Equal(0))
		})
		It("should fail when dbClient fails to list volumes", func() {
			fakeLock.LockReturns(nil)
			fakeSoftlayerDataModel.ListVolumesReturns(nil, fmt.Errorf("error listing volumes"))
			volumes, err := client.ListVolumes()
			Expect(err).To(HaveOccurred())
			Expect(len(volumes)).To(Equal(0))
			Expect(err.Error()).To(Equal("error listing volumes"))
			Expect(fakeLock.LockCallCount()).To(Equal(1))
			Expect(fakeSoftlayerDataModel.ListVolumesCallCount()).To(Equal(1))
		})
		It("should succeed to list volumes", func() {
			fakeLock.LockReturns(nil)

			volume1 := softlayer.SoftlayerVolume{Name: "fake-volume1", Mountpoint: "fake-mountpoint"}
			volume2 := softlayer.SoftlayerVolume{Name: "fake-volume2", Mountpoint: "fake-mountpoint"}
			volumesList := make([]softlayer.SoftlayerVolume, 2)
			volumesList[0] = volume1
			volumesList[1] = volume2
			fakeSoftlayerDataModel.ListVolumesReturns(volumesList, nil)
			volumes, err := client.ListVolumes()
			Expect(err).ToNot(HaveOccurred())
			Expect(len(volumes)).To(Equal(2))
			Expect(fakeLock.LockCallCount()).To(Equal(1))
			Expect(fakeSoftlayerDataModel.ListVolumesCallCount()).To(Equal(1))
		})

	})

	Context("GetVolume", func() {
		It("should fail when fileLock fails to aquire the lock", func() {
			fakeLock.LockReturns(fmt.Errorf("error aquiring the lock"))
			_, _, err = client.GetVolume("fake-volume")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("error aquiring the lock"))
			Expect(fakeLock.LockCallCount()).To(Equal(1))
			Expect(fakeSoftlayerDataModel.GetVolumeCallCount()).To(Equal(0))
		})

		It("should fail when dbClient fails to check if the volume exists", func() {
			fakeLock.LockReturns(nil)
			fakeSoftlayerDataModel.GetVolumeReturns(softlayer.SoftlayerVolume{}, false, fmt.Errorf("error checking volume"))
			_, _, err = client.GetVolume("fake-volume")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("error checking volume"))
			Expect(fakeLock.LockCallCount()).To(Equal(1))
			Expect(fakeSoftlayerDataModel.GetVolumeCallCount()).To(Equal(1))
		})

		It("should fail when volume exists and dbClient fails to getVolume", func() {
			fakeLock.LockReturns(nil)
			fakeSoftlayerDataModel.GetVolumeReturns(softlayer.SoftlayerVolume{}, true, fmt.Errorf("error getting volume"))
			_, _, err = client.GetVolume("fake-volume")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("error getting volume"))
			Expect(fakeLock.LockCallCount()).To(Equal(1))
			Expect(fakeSoftlayerDataModel.GetVolumeCallCount()).To(Equal(1))
		})

		It("should succeed  when volume exists", func() {
			fakeLock.LockReturns(nil)
			volume := softlayer.SoftlayerVolume{Name: "fake-volume", Mountpoint: "fake-mountpoint"}
			fakeSoftlayerDataModel.GetVolumeReturns(volume, true, nil)
			vol, _, err := client.GetVolume("fake-volume")
			Expect(err).ToNot(HaveOccurred())
			Expect(vol.Name).To(Equal("fake-volume"))
			Expect(vol.Mountpoint).To(Equal("fake-mountpoint"))
			Expect(fakeLock.LockCallCount()).To(Equal(1))
			Expect(fakeSoftlayerDataModel.GetVolumeCallCount()).To(Equal(1))
		})

	})

	Context(".Attach", func() {

		It("should fail when volume exists and dbClient fails to getVolume", func() {
			fakeSoftlayerDataModel.GetVolumeReturns(softlayer.SoftlayerVolume{}, false, fmt.Errorf("error getting volume"))
			mountpath, err := client.Attach("fake-volume")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("error getting volume"))
			Expect(mountpath).To(Equal(""))
			Expect(fakeSoftlayerDataModel.GetVolumeCallCount()).To(Equal(1))
		})

		It("should fail when slClient fails to attach it", func() {
			volume := softlayer.SoftlayerVolume{Name: "fake-volume", Id: -1}
			fakeSoftlayerDataModel.GetVolumeReturns(volume, true, nil)
			fakeSoftlayerStorageService.AllowAccessFromAllSubnetsReturns(false, fmt.Errorf("failed to grant access"))
			mountpath, err := client.Attach("fake-volume")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("failed to grant access"))
			Expect(mountpath).To(Equal(""))
			Expect(fakeSoftlayerDataModel.GetVolumeCallCount()).To(Equal(1))
			Expect(fakeSoftlayerStorageService.AllowAccessFromAllSubnetsCallCount()).To(Equal(1))
		})

		It("should succeed when volume is valid", func() {
			volume := softlayer.SoftlayerVolume{Name: "fake-volume", Id: -1, Mountpoint: "fake-mountpoint"}
			fakeSoftlayerDataModel.GetVolumeReturns(volume, true, nil)
			mountpath, err := client.Attach("fake-volume")
			Expect(err).ToNot(HaveOccurred())
			Expect(mountpath).To(Equal("fake-mountpoint"))
			Expect(fakeSoftlayerDataModel.GetVolumeCallCount()).To(Equal(1))
		})

	})

	Context(".Detach", func() {

		It("should fail when volume exists and dbClient fails to getVolume", func() {
			fakeSoftlayerDataModel.GetVolumeReturns(softlayer.SoftlayerVolume{}, false, fmt.Errorf("error getting volume"))
			err = client.Detach("fake-volume")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("error getting volume"))
			Expect(fakeSoftlayerDataModel.GetVolumeCallCount()).To(Equal(1))
		})

		It("should succeed when volume exists and is detachable", func() {
			volume := softlayer.SoftlayerVolume{Name: "fake-volume", Mountpoint: "fake-mountpoint"}
			fakeSoftlayerDataModel.GetVolumeReturns(volume, true, nil)
			err := client.Detach("fake-volume")
			Expect(err).ToNot(HaveOccurred())
			Expect(fakeSoftlayerDataModel.GetVolumeCallCount()).To(Equal(1))

		})

	})
})
