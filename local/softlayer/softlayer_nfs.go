package softlayer

import (
	"log"

	"fmt"

	"strconv"

	"github.com/jinzhu/gorm"
	"github.ibm.com/alchemy-containers/armada-slclient-lib/interfaces"
	"github.ibm.com/alchemy-containers/armada-slclient-lib/services"
	"github.ibm.com/almaden-containers/ubiquity/resources"
	"github.ibm.com/almaden-containers/ubiquity/utils"
	"k8s.io/kubernetes/staging/src/k8s.io/client-go/_vendor/github.com/golang/glog"
)

type softlayerLocalClient struct {
	logger                  *log.Logger
	fileLock                utils.FileLock
	dataModel               SoftlayerDataModel
	isActivated             bool
	softlayerStorageService interfaces.Softlayer_Storage_Service
}

const (
	USER_SPECIFIED_SIZE = "size"
	DEFAULT_SIZE        = 1
	STORAGE_CLASS_IOPS  = "iops"
	DEFAULT_IOPS        = 1
	SL_LOCATION         = "sl_location"
)

func NewSoftlayerLocalClient(logger *log.Logger, config resources.SoftlayerConfig, db *gorm.DB, fileLock utils.FileLock) (resources.StorageClient, error) {

	//Get the service client
	// has to look for access details in config before preoceeding
	serviceClient := services.NewServiceClient()
	softlayerStorageService := serviceClient.GetSoftLayerStorageService()

	return &softlayerLocalClient{logger: logger, fileLock: fileLock, dataModel: NewSoftlayerDataModel(logger, db, resources.SOFTLAYER_NFS), softlayerStorageService: softlayerStorageService}, nil
}
func NewSoftlayerLocalClientWithDataModelAndSLService(logger *log.Logger, datamodel SoftlayerDataModel, fileLock utils.FileLock, softlayerStorageService interfaces.Softlayer_Storage_Service) (resources.StorageClient, error) {
	return &softlayerLocalClient{logger: logger, fileLock: fileLock, dataModel: datamodel, softlayerStorageService: softlayerStorageService}, nil
}

func (s *softlayerLocalClient) Activate() (err error) {
	s.logger.Println("softlayerLocalClient: Activate start")
	defer s.logger.Println("softlayerLocalClient: Activate end")
	err = s.fileLock.Lock()
	if err != nil {
		s.logger.Printf("Error aquiring lock %v", err)
		return err
	}
	defer func() {
		lockErr := s.fileLock.Unlock()
		if lockErr != nil && err == nil {
			err = lockErr
		}
	}()

	if s.isActivated {
		return nil
	}
	err = s.dataModel.CreateVolumeTable()

	if err != nil {
		s.logger.Println(err.Error())
		return err
	}

	s.isActivated = true
	return nil
}

func (s *softlayerLocalClient) CreateVolume(name string, opts map[string]interface{}) (err error) {
	s.logger.Println("softlayerLocalClient: create start")
	defer s.logger.Println("softlayerLocalClient: create end")

	err = s.fileLock.Lock()
	if err != nil {
		s.logger.Printf("error aquiring lock %v", err)
		return err
	}
	defer func() {
		lockErr := s.fileLock.Unlock()
		if lockErr != nil && err == nil {
			err = lockErr
		}
	}()

	_, volExists, err := s.dataModel.GetVolume(name)

	if err != nil {
		s.logger.Println(err.Error())
		return err
	}

	if volExists {
		return fmt.Errorf("Volume already exists")
	}

	s.logger.Printf("Opts for create: %#v\n", opts)

	return s.createFileShareVolume(name, opts)
}

func (s *softlayerLocalClient) createFileShareVolume(name string, opts map[string]interface{}) (err error) {
	s.logger.Println("softlayerLocalClient: createFileShareVolume start")
	defer s.logger.Println("softlayerLocalClient: createFileShareVolume end")

	iFSSize, iFSSizeSpecified := opts[USER_SPECIFIED_SIZE]
	if iFSSizeSpecified == false {
		iFSSize = DEFAULT_SIZE
	}
	fFSIops, fFSIopsSpecified := opts[STORAGE_CLASS_IOPS]
	if fFSIopsSpecified == false {
		fFSIops = DEFAULT_IOPS
	}
	sLocation, sLocationSpecified := opts[SL_LOCATION]
	if sLocationSpecified == false {
		err := fmt.Errorf("%s is a required parameter", SL_LOCATION)
		return err
	}
	storage, err := s.softlayerStorageService.CreateStorage(iFSSize.(int), fFSIops.(int), sLocation.(string))
	if err != nil {
		s.logger.Printf("Error creating volume (%s): %s", name, err.Error())
		return err
	}
	s.logger.Printf("Volume created: %#v\n", storage)

	opts[SERVER_BACKEND_IP] = storage.ServiceResourceBackendIpAddress
	opts[SERVER_USER_NAME] = storage.Username
	err = s.dataModel.InsertFileshare(storage.Id, name, fmt.Sprintf("%s/%s", storage.ServiceResourceBackendIpAddress, storage.Username), opts)
	if err != nil {
		s.logger.Printf("Error inserting fileshare %v", err)
		return err
	}

	return nil
}

func (s *softlayerLocalClient) RemoveVolume(name string, forceDelete bool) (err error) {
	s.logger.Println("softlayerLocalClient: remove start")
	defer s.logger.Println("softlayerLocalClient: remove end")
	err = s.fileLock.Lock()
	if err != nil {
		s.logger.Printf("failed to aquire lock %v", err)
		return err
	}
	defer func() {
		lockErr := s.fileLock.Unlock()
		if lockErr != nil && err == nil {
			err = lockErr
		}
	}()

	existingVolume, exists, err := s.dataModel.GetVolume(name)

	if err != nil {
		s.logger.Println(err.Error())
		return err
	}
	if exists == false {
		s.logger.Println(err.Error())
		return err
	}

	bUnauthorized, err := s.softlayerStorageService.RemoveAccessFromAllSubnets(existingVolume.FileshareID)
	if !bUnauthorized {
		glog.Infoln("\nUnable to un-authorize Storage for storageid: %d", existingVolume.FileshareID)
		return err
	}

	if forceDelete == true {
		err = s.softlayerStorageService.DeleteStorage(existingVolume.FileshareID)
		if err != nil {
			glog.Infoln("\nDelete failed for storageid: %d with Error: %s", existingVolume.FileshareID, err)
			return err
		} else {
			glog.Infoln("\nDelete successful for storageid: %d", existingVolume.FileshareID)
		}
	}
	err = s.dataModel.DeleteVolume(name)

	if err != nil {
		s.logger.Println(err.Error())
		return err
	}
	return nil
}

func (s *softlayerLocalClient) GetVolume(name string) (volumeMetadata resources.VolumeMetadata, volumeConfigDetails map[string]interface{}, err error) {
	s.logger.Println("softlayerLocalClient: get start")
	defer s.logger.Println("softlayerLocalClient: get finish")

	err = s.fileLock.Lock()
	if err != nil {
		s.logger.Printf("error aquiring lock", err)
		return resources.VolumeMetadata{}, nil, err
	}
	defer func() {
		lockErr := s.fileLock.Unlock()
		if lockErr != nil && err == nil {
			err = lockErr
		}
	}()

	existingVolume, volExists, err := s.dataModel.GetVolume(name)

	if err != nil {
		s.logger.Println(err.Error())
		return resources.VolumeMetadata{}, nil, err
	}

	if volExists {

		volumeMetadata = resources.VolumeMetadata{Name: existingVolume.Volume.Name, Mountpoint: existingVolume.Mountpoint}
		volumeConfigDetails = map[string]interface{}{"FileshareId": strconv.Itoa(existingVolume.FileshareID)}
		return volumeMetadata, volumeConfigDetails, nil
	}
	return resources.VolumeMetadata{}, nil, fmt.Errorf("Volume not found")

}

func (s *softlayerLocalClient) Attach(name string) (string, error) {
	s.logger.Println("softlayerLocalClient: attach start")
	defer s.logger.Println("softlayerLocalClient: attach end")

	existingVolume, exists, err := s.dataModel.GetVolume(name)

	if err != nil {
		s.logger.Println(err.Error())
		return "", err
	}
	if exists == false {
		s.logger.Println(err.Error())
		return "", err
	}

	bIsAuthorized, err11 := s.softlayerStorageService.AllowAccessFromAllSubnets(existingVolume.FileshareID)
	if err11 != nil {
		glog.Infoln("\nUnable to authorize Storage for storageid: %d with Error: %s", existingVolume.FileshareID, err11)
		return "", err11
	}
	fmt.Println(bIsAuthorized)
	glog.Infoln("\n Storage authorized with storageId: %s", existingVolume.FileshareID)

	return existingVolume.Mountpoint, nil
}
func (s *softlayerLocalClient) Detach(name string) (err error) {
	s.logger.Println("softlayerLocalClient: detach start")
	defer s.logger.Println("softlayerLocalClient: detach end")
	existingVolume, exists, err := s.dataModel.GetVolume(name)

	if err != nil {
		s.logger.Println(err.Error())
		return err
	}
	if exists == false {
		s.logger.Println(err.Error())
		return err
	}

	bUnauthorized, err := s.softlayerStorageService.RemoveAccessFromAllSubnets(existingVolume.FileshareID)
	if !bUnauthorized {
		glog.Infoln("\nUnable to un-authorize Storage for storageid: %d", existingVolume.FileshareID)
		return err
	}
	return nil
}

func (s *softlayerLocalClient) ListVolumes() ([]resources.VolumeMetadata, error) {
	s.logger.Println("softlayerLocalClient: list start")
	defer s.logger.Println("softlayerLocalClient: list end")
	var err error
	err = s.fileLock.Lock()
	if err != nil {
		s.logger.Printf("error aquiring lock", err)
		return nil, err
	}
	defer func() {
		lockErr := s.fileLock.Unlock()
		if lockErr != nil && err == nil {
			err = lockErr
		}
	}()

	volumesInDb, err := s.dataModel.ListVolumes()

	if err != nil {
		s.logger.Printf("error retrieving volumes from db %#v\n", err)
		return nil, err
	}
	s.logger.Printf("Volumes in db: %d\n", len(volumesInDb))
	var volumes []resources.VolumeMetadata
	for _, volume := range volumesInDb {
		s.logger.Printf("Volume from db: %#v\n", volume)
		volumes = append(volumes, resources.VolumeMetadata{Name: volume.Volume.Name, Mountpoint: volume.Mountpoint})
	}

	return volumes, nil
}
