package scbe

import (
	"fmt"
	"github.com/IBM/ubiquity/resources"
	"github.com/jinzhu/gorm"
	"log"
	"strconv"
	"sync"
)

type scbeLocalClient struct {
	logger         *log.Logger
	dataModel      ScbeDataModel
	scbeRestClient ScbeRestClient
	isActivated    bool
	config         resources.ScbeConfig
	activationLock *sync.RWMutex
}

const (
	OptionNameForServiceName        = "profile"
	OptionNameForVolumeSize         = "size"
	volumeNamePrefix                = "u_"
	DefaultUbiquityInstanceName     = "ubiquity_instance1" // TODO this should be part of the configuration
	AttachedToNothing               = ""                   // during provisioning the volume is not attached to any host
	PathToMountUbiquityBlockDevices = "/ubiquity/%s"       // %s is the WWN of the volume
	EmptyHost                       = ""

	VolConfigKeyWWN      = "wwn"
	VolConfigKeyProfile  = "profile"
	VolConfigKeyID       = "id"
	VolConfigKeyVolumeID = "volume_id"
)

var (
	ComposeVolumeName = volumeNamePrefix + "%s_%s"
)

// prefix_ubiquityIntanceName_northboundVolumeName

func NewScbeLocalClient(logger *log.Logger, config resources.ScbeConfig, database *gorm.DB) (resources.StorageClient, error) {
	if config.ConfigPath == "" {
		return nil, fmt.Errorf("scbeLocalClient: init: missing required parameter 'spectrumConfigPath'")
	}

	return newScbeLocalClient(logger, config, database, resources.SCBE)
}

func NewScbeLocalClientWithNewScbeRestClientAndDataModel(logger *log.Logger, config resources.ScbeConfig, dataModel ScbeDataModel, scbeRestClient ScbeRestClient) (resources.StorageClient, error) {
	if config.ConfigPath == "" {
		return nil, fmt.Errorf("scbeLocalClient: init: missing required parameter 'spectrumConfigPath'")
	}
	return &scbeLocalClient{
		logger:         logger,
		scbeRestClient: scbeRestClient, // TODO need to mock it in more advance way
		dataModel:      dataModel,
		config:         config,
		activationLock: &sync.RWMutex{},
	}, nil
}

func newScbeLocalClient(logger *log.Logger, config resources.ScbeConfig, database *gorm.DB, backend resources.Backend) (*scbeLocalClient, error) {
	logger.Println("scbeLocalClient: init start")
	defer logger.Println("scbeLocalClient: init end")

	datamodel := NewScbeDataModel(logger, database, backend)
	err := datamodel.CreateVolumeTable()
	if err != nil {
		return &scbeLocalClient{}, err
	}

	scbeRestClient := NewScbeRestClient(logger, config.ConnectionInfo)

	client := &scbeLocalClient{
		logger:         logger,
		scbeRestClient: scbeRestClient,
		dataModel:      datamodel,
		config:         config,
		activationLock: &sync.RWMutex{},
	}
	return client, nil
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
	err := s.scbeRestClient.Login()
	if err != nil {
		s.logger.Printf("Error in login remote call %#v", err)
		return fmt.Errorf("Error in login remote call")
	} else {
		s.logger.Printf("Succeeded to login to SCBE %s", s.config.ConnectionInfo.ManagementIP)
	}

	var isExist bool
	isExist, err = s.scbeRestClient.ServiceExist(s.config.DefaultService)
	if err != nil {
		msg := fmt.Sprintf("Error in activate SCBE backend while checking default service. (%#v)", err)
		s.logger.Printf(msg)
		return fmt.Errorf(msg)
	}

	if isExist == false {
		msg := fmt.Sprintf("Error in activate SCBE backend %#v. The default service %s does not exist in SCBE %s",
			err, s.config.DefaultService, s.config.ConnectionInfo.ManagementIP)
		s.logger.Printf(msg)
		return fmt.Errorf(msg)
	} else {
		s.logger.Printf("The default service [%s] exist in SCBE %s",
			s.config.DefaultService, s.config.ConnectionInfo.ManagementIP)
	}

	s.logger.Println("scbe remoteClient: Activate success")

	s.isActivated = true
	return nil
}

// CreateVolume parse and validate the given options and trigger the volume creation
func (s *scbeLocalClient) CreateVolume(name string, opts map[string]interface{}) (err error) {
	s.logger.Println("scbeLocalClient: create start")
	defer s.logger.Println("scbeLocalClient: create end")

	_, volExists, err := s.dataModel.GetVolume(name)
	if err != nil {
		s.logger.Println(err.Error())
		return err
	}

	// validate service exist
	if volExists {
		return fmt.Errorf(MsgVolumeAlreadyExistInDB, name)
	}

	// validate size option given
	sizeStr := opts[OptionNameForVolumeSize]
	if sizeStr == "" || sizeStr == nil {
		return fmt.Errorf(MsgOptionSizeIsMissing)
	}

	// validate size is a number
	size, err := strconv.Atoi(sizeStr.(string))
	if err != nil {
		return fmt.Errorf(MsgOptionMustBeNumber, sizeStr)
	}
	// Get the profile option
	profile := s.config.DefaultService
	if opts[OptionNameForServiceName] != "" && opts[OptionNameForServiceName] != nil {
		profile = opts[OptionNameForServiceName].(string)
	}

	// Provision the volume on SCBE service
	volNameToCreate := fmt.Sprintf(ComposeVolumeName, DefaultUbiquityInstanceName, name) // TODO need to get real instance name
	volInfo := ScbeVolumeInfo{}
	volInfo, err = s.scbeRestClient.CreateVolume(volNameToCreate, profile, size)
	if err != nil {
		return err
	}

	err = s.dataModel.InsertVolume(name, volInfo.Wwn, profile, AttachedToNothing)
	if err != nil {
		return err
	}
	msg := fmt.Sprintf("scbeLocalClient: Successfully create volume %s on profile %s", name, profile)
	defer s.logger.Println(msg)

	return nil
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
		return fmt.Errorf("Volume [%s] not found", name)
	}

	err = s.dataModel.DeleteVolume(name)

	if err != nil {
		s.logger.Println(err.Error())
		return err
	}
	err = s.scbeRestClient.DeleteVolume(existingVolume.WWN)
	if err != nil {
		return err
	}

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

	scbeVolume, volExists, err := s.dataModel.GetVolume(name)

	if err != nil {
		s.logger.Println(err.Error())
		return nil, err
	}

	if !volExists {
		return nil, fmt.Errorf("Volume not found")
	}

	volumeConfigDetails := make(map[string]interface{})
	volumeConfigDetails[VolConfigKeyWWN] = scbeVolume.WWN
	volumeConfigDetails[VolConfigKeyProfile] = scbeVolume.Profile
	volumeConfigDetails[VolConfigKeyID] = scbeVolume.ID
	volumeConfigDetails[VolConfigKeyVolumeID] = scbeVolume.VolumeID

	return volumeConfigDetails, nil

}
func (s *scbeLocalClient) Attach(name string) (string, error) {
	s.logger.Println("scbeLocalClient: attach start")
	defer s.logger.Println("scbeLocalClient: attach end")
	host2attach := s.config.HostnameTmp // TODO this is workaround for issue #23 (remove it when #23 will be fixed and use the host that will be given as argument to the interface)

	existingVolume, volExists, err := s.dataModel.GetVolume(name)
	if err != nil {
		s.logger.Println(err.Error())
		return "", err
	}

	if !volExists {
		return "", fmt.Errorf(fmt.Sprintf(MsgVolumeNotFoundInUbiquityDB, "detach", name))
	}

	if existingVolume.AttachTo == host2attach {
		// if already map to the given host then just ignore and succeed to attach
		s.logger.Printf(fmt.Sprintf(MsgVolumeAlreadyAttached, host2attach))
		volumeMountpoint := fmt.Sprintf(PathToMountUbiquityBlockDevices, existingVolume.WWN)
		return volumeMountpoint, nil
	} else if existingVolume.AttachTo != "" {
		// if it attached to other node , then exit with error
		msg := fmt.Sprintf(MsgCannotAttachVolThatAlreadyAttached, name, host2attach, existingVolume.AttachTo)
		s.logger.Printf(msg)
		return "", fmt.Errorf(msg)
	}

	s.logger.Printf("Attaching vol %#v", existingVolume)
	_, err = s.scbeRestClient.MapVolume(existingVolume.WWN, host2attach)
	if err != nil {
		return "", err
	}

	err = s.dataModel.UpdateVolumeAttachTo(name, existingVolume, host2attach)
	if err != nil {
		return "", err
	}

	volumeMountpoint := fmt.Sprintf(PathToMountUbiquityBlockDevices, existingVolume.WWN)
	return volumeMountpoint, nil
}

func (s *scbeLocalClient) Detach(name string) (err error) {
	s.logger.Println("scbeLocalClient: detach start")
	defer s.logger.Println("scbeLocalClient: detach end")
	host2detach := s.config.HostnameTmp // TODO this is workaround for issue #23 (remove it when #23 will be fixed and use the host that will be given as argument to the interface)

	existingVolume, volExists, err := s.dataModel.GetVolume(name)
	if err != nil {
		s.logger.Println(err.Error())
		return err
	}

	if !volExists {
		return fmt.Errorf(fmt.Sprintf(MsgVolumeNotFoundInUbiquityDB, "detach", name))
	}

	// Fail if vol already detach
	if existingVolume.AttachTo == EmptyHost {
		msg := fmt.Sprintf(MsgCannotDetachVolThatAlreadyDetached, name, host2detach)
		s.logger.Printf(msg)
		return fmt.Errorf(msg)
	}

	s.logger.Printf("Detach vol %#v", existingVolume)
	err = s.scbeRestClient.UnmapVolume(existingVolume.WWN, host2detach)
	if err != nil {
		return err
	}

	err = s.dataModel.UpdateVolumeAttachTo(name, existingVolume, EmptyHost)
	if err != nil {
		return err
	}

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

func (s *scbeLocalClient) createVolume(name string, wwn string, profile string) error {
	s.logger.Println("scbeLocalClient: createVolume start")
	defer s.logger.Println("scbeLocalClient: createVolume end")

	err := s.dataModel.InsertVolume(name, wwn, profile, "")

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
