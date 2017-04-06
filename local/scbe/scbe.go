package scbe

import (
	"log"
	"net/http"

	"fmt"

	"github.com/jinzhu/gorm"

	"sync"

	"github.com/IBM/ubiquity/resources"
	"github.com/IBM/ubiquity/utils"
)

type scbeLocalClient struct {
	logger         *log.Logger
	dataModel      ScbeDataModel
	httpClient     *http.Client
	isActivated    bool
	config         resources.ScbeConfig
	activationLock *sync.RWMutex
}

const ()

func NewScbeLocalClient(logger *log.Logger, config resources.ScbeConfig, database *gorm.DB) (resources.StorageClient, error) {
	if config.ConfigPath == "" {
		return nil, fmt.Errorf("scbeLocalClient: init: missing required parameter 'spectrumConfigPath'")
	}
	return newScbeLocalClient(logger, config, database, resources.SCBE)
}

func NewScbeLocalClientWithHTTPClientAndDataModel(logger *log.Logger, config resources.ScbeConfig, dataModel ScbeDataModel, httpClient *http.Client) (resources.StorageClient, error) {
	if config.ConfigPath == "" {
		return nil, fmt.Errorf("scbeLocalClient: init: missing required parameter 'spectrumConfigPath'")
	}
	return &scbeLocalClient{logger: logger, httpClient: httpClient, dataModel: dataModel, config: config, activationLock: &sync.RWMutex{}}, nil
}

func newScbeLocalClient(logger *log.Logger, config resources.ScbeConfig, database *gorm.DB, backend resources.Backend) (*scbeLocalClient, error) {
	logger.Println("scbeLocalClient: init start")
	defer logger.Println("scbeLocalClient: init end")

	datamodel := NewScbeDataModel(logger, database, backend)
	err := datamodel.CreateVolumeTable()
	if err != nil {
		return &scbeLocalClient{}, err
	}
	return &scbeLocalClient{logger: logger, httpClient: &http.Client{}, dataModel: datamodel, config: config, activationLock: &sync.RWMutex{}}, nil
}

func (s *scbeLocalClient) Activate() error {
	s.logger.Println("scbeLocalClient: Activate start")
	defer s.logger.Println("scbeLocalClient: Activate end")

	s.activationLock.RLock()
	if s.isActivated {
		s.activationLock.RUnlock()
		return nil
	}
	s.activationLock.RUnlock()

	s.activationLock.Lock() //get a write lock to prevent others from repeating these actions
	defer s.activationLock.Unlock()

	//do any needed configuration
	activateURL := utils.FormatURL(s.config.ScbeURL, "activate")
	response, err := utils.HttpExecute(s.httpClient, s.logger, "POST", activateURL, nil)
	if err != nil {
		s.logger.Printf("Error in activate remote call %#v", err)
		return fmt.Errorf("Error in activate remote call")
	}

	if response.StatusCode != http.StatusAccepted {
		s.logger.Printf("Error in activate scbe remote call %#v\n", response)
		return utils.ExtractErrorResponse(response)
	}
	s.logger.Println("scbe remoteClient: Activate success")

	s.isActivated = true
	return nil
}

func (s *scbeLocalClient) CreateVolume(name string, opts map[string]interface{}) (err error) {
	s.logger.Println("scbeLocalClient: create start")
	defer s.logger.Println("scbeLocalClient: create end")

	_, volExists, err := s.dataModel.GetVolume(name)

	if err != nil {
		s.logger.Println(err.Error())
		return err
	}

	if volExists {
		return fmt.Errorf("Volume already exists")
	}

	s.logger.Printf("Opts for create: %#v\n", opts)

	if len(opts) == 0 {

		return s.createVolume(name, opts)
	}
	return fmt.Errorf("Internal error")
}

func (s *scbeLocalClient) RemoveVolume(name string) (err error) {
	s.logger.Println("scbeLocalClient: remove start")
	defer s.logger.Println("scbeLocalClient: remove end")

	existingVolume, volExists, err := s.dataModel.GetVolume(name)

	if err != nil {
		s.logger.Println(err.Error())
		return err
	}

	if volExists == false {
		return fmt.Errorf("Volume not found")
	}

	err = s.dataModel.DeleteVolume(name)

	if err != nil {
		s.logger.Println(err.Error())
		return err
	}
	//TODO real delete
	fmt.Printf("Delete me %#v", existingVolume)

	return nil
}

func (s *scbeLocalClient) GetVolume(name string) (resources.Volume, error) {
	s.logger.Println("scbeLocalClient: GetVolume start")
	defer s.logger.Println("scbeLocalClient: GetVolume finish")

	existingVolume, volExists, err := s.dataModel.GetVolume(name)
	if err != nil {
		return resources.Volume{}, err
	}
	if volExists == false {
		return resources.Volume{}, fmt.Errorf("Volume not found")
	}

	return resources.Volume{Name: existingVolume.Volume.Name, Backend: resources.Backend(existingVolume.Volume.Backend)}, nil
}

func (s *scbeLocalClient) GetVolumeConfig(name string) (map[string]interface{}, error) {
	s.logger.Println("scbeLocalClient: GetVolumeConfig start")
	defer s.logger.Println("scbeLocalClient: GetVolumeConfig finish")

	existingVolume, volExists, err := s.dataModel.GetVolume(name)

	if err != nil {
		s.logger.Println(err.Error())
		return nil, err
	}

	if !volExists {
		return nil, fmt.Errorf("Volume not found")
	}

	volumeConfigDetails := make(map[string]interface{})
	//TODO fill the configDetails
	fmt.Printf("get my config %#v", existingVolume)
	return volumeConfigDetails, nil

}
func (s *scbeLocalClient) Attach(name string) (string, error) {
	s.logger.Println("scbeLocalClient: attach start")
	defer s.logger.Println("scbeLocalClient: attach end")

	existingVolume, volExists, err := s.dataModel.GetVolume(name)

	if err != nil {
		s.logger.Println(err.Error())
		return "", err
	}

	if !volExists {
		return "", fmt.Errorf("Volume not found")
	}

	//TODO do the attach and get the mountpoint
	fmt.Printf("attach me %#v", existingVolume)
	volumeMountpoint := ""
	return volumeMountpoint, nil
}

func (s *scbeLocalClient) Detach(name string) (err error) {
	s.logger.Println("scbeLocalClient: detach start")
	defer s.logger.Println("scbeLocalClient: detach end")

	existingVolume, volExists, err := s.dataModel.GetVolume(name)

	if err != nil {
		s.logger.Println(err.Error())
		return err
	}

	if !volExists {
		return fmt.Errorf("Volume not found")
	}

	//TODO detach
	fmt.Printf("Detach me %#v", existingVolume)
	return nil
}
func (s *scbeLocalClient) ListVolumes() ([]resources.VolumeMetadata, error) {
	s.logger.Println("scbeLocalClient: list start")
	defer s.logger.Println("scbeLocalClient: list end")
	var err error

	volumesInDb, err := s.dataModel.ListVolumes()

	if err != nil {
		s.logger.Printf("error retrieving volumes from db %#v\n", err)
		return nil, err
	}
	s.logger.Printf("Volumes in db: %d\n", len(volumesInDb))
	var volumes []resources.VolumeMetadata
	for _, volume := range volumesInDb {
		s.logger.Printf("Volume from db: %#v\n", volume)

		volumeMountpoint, err := s.getVolumeMountPoint(volume)
		if err != nil {
			s.logger.Println(err.Error())
			return nil, err
		}

		volumes = append(volumes, resources.VolumeMetadata{Name: volume.Volume.Name, Mountpoint: volumeMountpoint})
	}

	return volumes, nil
}

func (s *scbeLocalClient) createVolume(name string, opts map[string]interface{}) error {
	s.logger.Println("scbeLocalClient: createVolume start")
	defer s.logger.Println("scbeLocalClient: createVolume end")

	err := s.dataModel.InsertVolume(name, opts)

	if err != nil {
		s.logger.Printf("Error inserting volume %v", err)
		return err
	}

	return nil
}
func (s *scbeLocalClient) getVolumeMountPoint(volume ScbeVolume) (string, error) {
	s.logger.Println("scbeLocalClient getVolumeMountPoint start")
	defer s.logger.Println("scbeLocalClient getVolumeMountPoint end")

	//TODO return mountpoint
	return "some mount point", nil
}
