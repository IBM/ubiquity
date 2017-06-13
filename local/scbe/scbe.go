package scbe

import (
	"crypto/tls"
	"fmt"
	"github.com/IBM/ubiquity/resources"
	"github.com/jinzhu/gorm"
	"strconv"
	"sync"
	"github.com/IBM/ubiquity/logutil"
	"errors"
)

type scbeLocalClient struct {
	logger         logutil.Logger
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

func NewScbeLocalClient(config resources.ScbeConfig, database *gorm.DB) (resources.StorageClient, error) {
	if config.ConfigPath == "" {
		return nil, fmt.Errorf("scbeLocalClient: init: missing required parameter 'spectrumConfigPath'")
	}

	return newScbeLocalClient(config, database, resources.SCBE)
}

func NewScbeLocalClientWithNewScbeRestClientAndDataModel(config resources.ScbeConfig, dataModel ScbeDataModel, scbeRestClient ScbeRestClient) (resources.StorageClient, error) {
	if config.ConfigPath == "" {
		return nil, fmt.Errorf("scbeLocalClient: init: missing required parameter 'spectrumConfigPath'")
	}
	return &scbeLocalClient{
		logger:         logutil.GetLogger(),
		scbeRestClient: scbeRestClient, // TODO need to mock it in more advance way
		dataModel:      dataModel,
		config:         config,
		activationLock: &sync.RWMutex{},
	}, nil
}

func newScbeLocalClient(config resources.ScbeConfig, database *gorm.DB, backend resources.Backend) (*scbeLocalClient, error) {
	defer logutil.GetLogger().Trace(logutil.DEBUG)()

	datamodel := NewScbeDataModel(database, backend)
	err := datamodel.CreateVolumeTable()
	if err != nil {
		return &scbeLocalClient{}, err
	}

	scbeRestClient := NewScbeRestClient(config.ConnectionInfo)

	client := &scbeLocalClient{
		logger:         logutil.GetLogger(),
		scbeRestClient: scbeRestClient,
		dataModel:      datamodel,
		config:         config,
		activationLock: &sync.RWMutex{},
	}
	return client, nil
}

func (s *scbeLocalClient) Activate() error {
	defer s.logger.Trace(logutil.DEBUG)()

	s.activationLock.RLock()
	if s.isActivated {
		s.activationLock.RUnlock()
		return nil
	}
	s.activationLock.RUnlock()

	s.activationLock.Lock() //get a write lock to prevent others from repeating these actions
	defer s.activationLock.Unlock()

	//do any needed configuration
	if err := s.scbeRestClient.Login(); err != nil {
		s.logger.Error("scbeRestClient.Login() failed", logutil.Args{{"error", err}})
		return err
	}
	s.logger.Info("scbeRestClient.Login() succeeded", logutil.Args{{"SCBE",s.config.ConnectionInfo.ManagementIP}})

	isExist, err := s.scbeRestClient.ServiceExist(s.config.DefaultService)
	if err != nil {
		s.logger.Error("scbeRestClient.ServiceExist failed", logutil.Args{{"error", err}})
		return err
	}

	if isExist == false {
		msg := fmt.Sprintf("Error in activate SCBE backend %#v. The default service %s does not exist in SCBE %s",
			err, s.config.DefaultService, s.config.ConnectionInfo.ManagementIP)
		s.logger.Error(msg)
		return fmt.Errorf(msg)
	}
	s.logger.Info("The default service exist in SCBE", logutil.Args{	{s.config.ConnectionInfo.ManagementIP, s.config.DefaultService}})

	s.isActivated = true
	return nil
}

// CreateVolume parse and validate the given options and trigger the volume creation
func (s *scbeLocalClient) CreateVolume(name string, opts map[string]interface{}) (err error) {
	defer s.logger.Trace(logutil.DEBUG)()

	_, volExists, err := s.dataModel.GetVolume(name)
	if err != nil {
		s.logger.Error("dataModel.GetVolume failed", logutil.Args{{"name", name}, {"error", err}})
		return err
	}

	// validate volume doesn't exist
	if volExists {
		err = fmt.Errorf(MsgVolumeAlreadyExistInDB, name)
		s.logger.Error("failed", logutil.Args{{"error", err}})
		return err
	}

	// validate size option given
	sizeStr := opts[OptionNameForVolumeSize]
	if sizeStr == "" || sizeStr == nil {
		err = errors.New(MsgOptionSizeIsMissing)
		s.logger.Error("failed", logutil.Args{{"error", err}})
		return err
	}

	// validate size is a number
	size, err := strconv.Atoi(sizeStr.(string))
	if err != nil {
		err = fmt.Errorf(MsgOptionMustBeNumber, sizeStr)
		s.logger.Error("failed", logutil.Args{{"error", err}})
		return err
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
		s.logger.Error("scbeRestClient.CreateVolume failed", logutil.Args{{"error", err}})
		return err
	}

	err = s.dataModel.InsertVolume(name, volInfo.Wwn, profile, AttachedToNothing)
	if err != nil {
		s.logger.Error("dataModel.InsertVolume failed", logutil.Args{{"error", err}})
		return err
	}

	s.logger.Info("succeeded", logutil.Args{{"volume", name}, {"profile", profile}})
	return nil
}

func (s *scbeLocalClient) RemoveVolume(name string) (err error) {
	defer s.logger.Trace(logutil.DEBUG)()

	existingVolume, volExists, err := s.dataModel.GetVolume(name)
	if err != nil {
		s.logger.Error("dataModel.GetVolume failed", logutil.Args{{"error", err}})
		return err
	}

	if volExists == false {
		err = fmt.Errorf("Volume [%s] not found", name)
		s.logger.Error("failed", logutil.Args{{"error", err}})
		return err
	}

	if err = s.dataModel.DeleteVolume(name); err != nil {
		s.logger.Error("failed", logutil.Args{{"error", err}})
		return err
	}

	if err = s.scbeRestClient.DeleteVolume(existingVolume.WWN); err != nil {
		s.logger.Error("failed", logutil.Args{{"error", err}})
		return err
	}

	return nil
}

func (s *scbeLocalClient) GetVolume(name string) (resources.Volume, error) {
	defer s.logger.Trace(logutil.DEBUG)()

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
	defer s.logger.Trace(logutil.DEBUG)()

	scbeVolume, volExists, err := s.dataModel.GetVolume(name)

	if err != nil {
		s.logger.Error("dataModel.GetVolume failed", logutil.Args{{"error", err}})
		return nil, err
	}

	if !volExists {
		err = fmt.Errorf("Volume not found")
		s.logger.Error("failed", logutil.Args{{"error", err}})
		return nil, err
	}

	volumeConfigDetails := make(map[string]interface{})
	volumeConfigDetails[VolConfigKeyWWN] = scbeVolume.WWN
	volumeConfigDetails[VolConfigKeyProfile] = scbeVolume.Profile
	volumeConfigDetails[VolConfigKeyID] = scbeVolume.ID
	volumeConfigDetails[VolConfigKeyVolumeID] = scbeVolume.VolumeID

	return volumeConfigDetails, nil

}
func (s *scbeLocalClient) Attach(name string) (string, error) {
	defer s.logger.Trace(logutil.DEBUG)()
	host2attach := s.config.HostnameTmp // TODO this is workaround for issue #23 (remove it when #23 will be fixed and use the host that will be given as argument to the interface)

	existingVolume, volExists, err := s.dataModel.GetVolume(name)
	if err != nil {
		s.logger.Error("dataModel.GetVolume failed", logutil.Args{{"error", err}})
		return "", err
	}

	if !volExists {
		err = fmt.Errorf(fmt.Sprintf(MsgVolumeNotFoundInUbiquityDB, "detach", name))
		s.logger.Error("failed", logutil.Args{{"error", err}})
		return "", err
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

	s.logger.Debug("Attaching", logutil.Args{{"volume", existingVolume}})
	if _, err = s.scbeRestClient.MapVolume(existingVolume.WWN, host2attach); err != nil {
		s.logger.Error("scbeRestClient.MapVolume failed", logutil.Args{{"error", err}})
		return "", err
	}

	if err = s.dataModel.UpdateVolumeAttachTo(name, existingVolume, host2attach); err != nil {
		s.logger.Error("dataModel.UpdateVolumeAttachTo failed", logutil.Args{{"error", err}})
		return "", err
	}

	volumeMountpoint := fmt.Sprintf(PathToMountUbiquityBlockDevices, existingVolume.WWN)
	return volumeMountpoint, nil
}

func (s *scbeLocalClient) Detach(name string) (err error) {
	defer s.logger.Trace(logutil.DEBUG)()
	host2detach := s.config.HostnameTmp // TODO this is workaround for issue #23 (remove it when #23 will be fixed and use the host that will be given as argument to the interface)

	existingVolume, volExists, err := s.dataModel.GetVolume(name)
	if err != nil {
		s.logger.Error("dataModel.GetVolume failed", logutil.Args{{"error", err}})
		return err
	}

	if !volExists {
		err = fmt.Errorf(fmt.Sprintf(MsgVolumeNotFoundInUbiquityDB, "detach", name))
		s.logger.Error("failed", logutil.Args{{"error", err}})
		return err
	}

	// Fail if vol already detach
	if existingVolume.AttachTo == EmptyHost {
		err = fmt.Errorf(fmt.Sprintf(MsgCannotDetachVolThatAlreadyDetached, name, host2detach))
		s.logger.Error("failed", logutil.Args{{"error", err}})
		return err
	}

	s.logger.Debug("Detaching", logutil.Args{{"volume", existingVolume}})
	if err = s.scbeRestClient.UnmapVolume(existingVolume.WWN, host2detach); err != nil {
		s.logger.Error("scbeRestClient.UnmapVolume failed", logutil.Args{{"error", err}})
		return err
	}

	if err = s.dataModel.UpdateVolumeAttachTo(name, existingVolume, EmptyHost); err != nil {
		s.logger.Error("dataModel.UpdateVolumeAttachTo failed", logutil.Args{{"error", err}})
		return err
	}

	return nil
}

func (s *scbeLocalClient) ListVolumes() ([]resources.VolumeMetadata, error) {
	defer s.logger.Trace(logutil.DEBUG)()
	var err error

	volumesInDb, err := s.dataModel.ListVolumes()
	if err != nil {
		s.logger.Error("dataModel.ListVolumes failed", logutil.Args{{"error", err}})
		return nil, err
	}

	s.logger.Debug("Volumes in db", logutil.Args{{"num", len(volumesInDb)}})
	var volumes []resources.VolumeMetadata
	for _, volume := range volumesInDb {
		s.logger.Debug("Volumes from db", logutil.Args{{"volume", volume}})
		volumeMountpoint, err := s.getVolumeMountPoint(volume)
		if err != nil {
			s.logger.Error("getVolumeMountPoint failed", logutil.Args{{"error", err}})
			return nil, err
		}

		volumes = append(volumes, resources.VolumeMetadata{Name: volume.Volume.Name, Mountpoint: volumeMountpoint})
	}

	return volumes, nil
}

func (s *scbeLocalClient) createVolume(name string, wwn string, profile string) error {
	defer s.logger.Trace(logutil.DEBUG)()

	if err := s.dataModel.InsertVolume(name, wwn, profile, ""); err != nil {
		s.logger.Error("dataModel.InsertVolume failed", logutil.Args{{"error", err}})
		return err
	}

	return nil
}
func (s *scbeLocalClient) getVolumeMountPoint(volume ScbeVolume) (string, error) {
	defer s.logger.Trace(logutil.DEBUG)()

	//TODO return mountpoint
	return "some mount point", nil
}
