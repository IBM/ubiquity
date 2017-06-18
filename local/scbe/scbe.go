package scbe

import (
	"errors"
	"fmt"
	"github.com/IBM/ubiquity/logutil"
	"github.com/IBM/ubiquity/resources"
	"github.com/jinzhu/gorm"
	"strconv"
	"sync"
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
	OptionNameForServiceName = "profile"
	OptionNameForVolumeSize  = "size"
	volumeNamePrefix         = "u_"
	AttachedToNothing        = "" // during provisioning the volume is not attached to any host
	EmptyHost                = ""

	VolConfigKeyWWN      = "wwn"
	VolConfigKeyProfile  = "profile"
	VolConfigKeyID       = "id"
	VolConfigKeyVolumeID = "volume_id"
	ComposeVolumeName    = volumeNamePrefix + "%s_%s" // e.g u_instance1_volName
)

func NewScbeLocalClient(config resources.ScbeConfig, database *gorm.DB) (resources.StorageClient, error) {
	logger := logutil.GetLogger()
	validateScbeConfig(config)
	datamodel := NewScbeDataModel(database, resources.SCBE)
	err := datamodel.CreateVolumeTable()
	if err != nil {
		return &scbeLocalClient{}, logger.ErrorRet(err, "failed")
	}
	scbeRestClient := NewScbeRestClient(config.ConnectionInfo)
	return NewScbeLocalClientWithNewScbeRestClientAndDataModel(config, datamodel, scbeRestClient)
}

func NewScbeLocalClientWithNewScbeRestClientAndDataModel(config resources.ScbeConfig, dataModel ScbeDataModel, scbeRestClient ScbeRestClient) (resources.StorageClient, error) {
	if err := validateScbeConfig(config); err != nil {
		return &scbeLocalClient{}, err
	}
	return &scbeLocalClient{
		logger:         logutil.GetLogger(),
		scbeRestClient: scbeRestClient, // TODO need to mock it in more advance way
		dataModel:      dataModel,
		config:         config,
		activationLock: &sync.RWMutex{},
	}, nil
}

func validateScbeConfig(config resources.ScbeConfig) error {
	logger := logutil.GetLogger()

	if config.DefaultVolumeSize == "" {
		// means customer didn't configure the default
		config.DefaultVolumeSize = resources.DefaultForScbeConfigParamDefaultVolumeSize
		logger.Debug("No DefaultVolumeSize defined in conf file, so set the DefaultVolumeSize to value " + resources.DefaultForScbeConfigParamDefaultVolumeSize)
	}
	_, err := strconv.Atoi(config.DefaultVolumeSize)
	if err != nil {
		return logger.ErrorRet(&ConfigDefaultSizeNotNumError{}, "failed")
	}

	if len(config.UbiquityInstanceName) > resources.UbiquityInstanceNameMaxSize {
		return logger.ErrorRet(&ConfigScbeUbiquityInstanceNameWrongSize{}, "failed")
	}

	// TODO add more verification on the config file.
	return nil
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
		return s.logger.ErrorRet(err, "scbeRestClient.Login() failed")
	}
	s.logger.Info("scbeRestClient.Login() succeeded", logutil.Args{{"SCBE", s.config.ConnectionInfo.ManagementIP}})

	isExist, err := s.scbeRestClient.ServiceExist(s.config.DefaultService)
	if err != nil {
		return s.logger.ErrorRet(err, "scbeRestClient.ServiceExist failed")
	}

	if isExist == false {
		return s.logger.ErrorRet(&activateDefaultServiceError{s.config.DefaultService, s.config.ConnectionInfo.ManagementIP}, "failed")
	}
	s.logger.Info("The default service exist in SCBE", logutil.Args{{s.config.ConnectionInfo.ManagementIP, s.config.DefaultService}})

	s.isActivated = true
	return nil
}

// CreateVolume parse and validate the given options and trigger the volume creation
func (s *scbeLocalClient) CreateVolume(name string, opts map[string]interface{}) (err error) {
	defer s.logger.Trace(logutil.DEBUG)()

	_, volExists, err := s.dataModel.GetVolume(name)
	if err != nil {
		return s.logger.ErrorRet(err, "dataModel.GetVolume failed", logutil.Args{{"name", name}})
	}

	// validate volume doesn't exist
	if volExists {
		return s.logger.ErrorRet(&volAlreadyExistsError{name}, "failed")
	}

	// validate size option given
	sizeStr, ok := opts[OptionNameForVolumeSize]
	if !ok {
		sizeStr = s.config.DefaultVolumeSize
		s.logger.Debug("No size given to create volume, so using the default_size",
			logutil.Args{{"volume", name}, {"default_size", sizeStr}})
	}

	// validate size is a number
	size, err := strconv.Atoi(sizeStr.(string))
	if err != nil {
		return s.logger.ErrorRet(&provisionParamIsNotNumberError{name, OptionNameForVolumeSize}, "failed")
	}
	// Get the profile option
	profile := s.config.DefaultService
	if opts[OptionNameForServiceName] != "" && opts[OptionNameForServiceName] != nil {
		profile = opts[OptionNameForServiceName].(string)
	}

	// Provision the volume on SCBE service
	volNameToCreate := fmt.Sprintf(ComposeVolumeName, s.config.UbiquityInstanceName, name)
	volInfo := ScbeVolumeInfo{}
	volInfo, err = s.scbeRestClient.CreateVolume(volNameToCreate, profile, size)
	if err != nil {
		return s.logger.ErrorRet(err, "scbeRestClient.CreateVolume failed")
	}

	err = s.dataModel.InsertVolume(name, volInfo.Wwn, AttachedToNothing)
	if err != nil {
		return s.logger.ErrorRet(err, "dataModel.InsertVolume failed")
	}

	s.logger.Info("succeeded", logutil.Args{{"volume", name}, {"profile", profile}})
	return nil
}

func (s *scbeLocalClient) RemoveVolume(name string) (err error) {
	defer s.logger.Trace(logutil.DEBUG)()

	existingVolume, volExists, err := s.dataModel.GetVolume(name)
	if err != nil {
		return s.logger.ErrorRet(err, "dataModel.GetVolume failed")
	}

	if volExists == false {
		return s.logger.ErrorRet(fmt.Errorf("Volume [%s] not found", name), "failed")
	}

	if err = s.dataModel.DeleteVolume(name); err != nil {
		return s.logger.ErrorRet(err, "dataModel.DeleteVolume failed")
	}

	if err = s.scbeRestClient.DeleteVolume(existingVolume.WWN); err != nil {
		return s.logger.ErrorRet(err, "scbeRestClient.DeleteVolume failed")
	}

	return nil
}

func (s *scbeLocalClient) GetVolume(name string) (resources.Volume, error) {
	defer s.logger.Trace(logutil.DEBUG)()

	existingVolume, volExists, err := s.dataModel.GetVolume(name)
	if err != nil {
		return resources.Volume{}, s.logger.ErrorRet(err, "dataModel.GetVolume failed")
	}
	if volExists == false {
		return resources.Volume{}, s.logger.ErrorRet(errors.New("Volume not found"), "failed")
	}

	return resources.Volume{Name: existingVolume.Volume.Name, Backend: resources.Backend(existingVolume.Volume.Backend)}, nil
}

func (s *scbeLocalClient) GetVolumeConfig(name string) (map[string]interface{}, error) {
	defer s.logger.Trace(logutil.DEBUG)()

	scbeVolume, volExists, err := s.dataModel.GetVolume(name)
	if err != nil {
		return nil, s.logger.ErrorRet(err, "dataModel.GetVolume failed")
	}

	if !volExists {
		return nil, s.logger.ErrorRet(errors.New("Volume not found"), "failed")
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
		return "", s.logger.ErrorRet(err, "dataModel.GetVolume failed")
	}

	if !volExists {
		return "", s.logger.ErrorRet(&volumeNotFoundError{name}, "failed")
	}

	if existingVolume.AttachTo == host2attach {
		// if already map to the given host then just ignore and succeed to attach
		s.logger.Info("Volume already attached, skip backend attach", logutil.Args{{"volume", name}, {"host", host2attach}})
		volumeMountpoint := fmt.Sprintf(resources.PathToMountUbiquityBlockDevices, existingVolume.WWN)
		return volumeMountpoint, nil
	} else if existingVolume.AttachTo != "" {
		return "", s.logger.ErrorRet(&volAlreadyAttachedError{name, existingVolume.AttachTo}, "failed")
	}

	s.logger.Debug("Attaching", logutil.Args{{"volume", existingVolume}})
	if _, err = s.scbeRestClient.MapVolume(existingVolume.WWN, host2attach); err != nil {
		return "", s.logger.ErrorRet(err, "scbeRestClient.MapVolume failed")
	}

	if err = s.dataModel.UpdateVolumeAttachTo(name, existingVolume, host2attach); err != nil {
		return "", s.logger.ErrorRet(err, "dataModel.UpdateVolumeAttachTo failed")
	}

	volumeMountpoint := fmt.Sprintf(resources.PathToMountUbiquityBlockDevices, existingVolume.WWN)
	return volumeMountpoint, nil
}

func (s *scbeLocalClient) Detach(name string) (err error) {
	defer s.logger.Trace(logutil.DEBUG)()
	host2detach := s.config.HostnameTmp // TODO this is workaround for issue #23 (remove it when #23 will be fixed and use the host that will be given as argument to the interface)

	existingVolume, volExists, err := s.dataModel.GetVolume(name)
	if err != nil {
		return s.logger.ErrorRet(err, "dataModel.GetVolume failed")
	}

	if !volExists {
		return s.logger.ErrorRet(&volumeNotFoundError{name}, "failed")
	}

	// Fail if vol already detach
	if existingVolume.AttachTo == EmptyHost {
		return s.logger.ErrorRet(&volNotAttachedError{name}, "failed")
	}

	s.logger.Debug("Detaching", logutil.Args{{"volume", existingVolume}})
	if err = s.scbeRestClient.UnmapVolume(existingVolume.WWN, host2detach); err != nil {
		return s.logger.ErrorRet(err, "scbeRestClient.UnmapVolume failed")
	}

	if err = s.dataModel.UpdateVolumeAttachTo(name, existingVolume, EmptyHost); err != nil {
		return s.logger.ErrorRet(err, "dataModel.UpdateVolumeAttachTo failed")
	}

	return nil
}

func (s *scbeLocalClient) ListVolumes() ([]resources.VolumeMetadata, error) {
	defer s.logger.Trace(logutil.DEBUG)()
	var err error

	volumesInDb, err := s.dataModel.ListVolumes()
	if err != nil {
		return nil, s.logger.ErrorRet(err, "dataModel.ListVolumes failed")
	}

	s.logger.Debug("Volumes in db", logutil.Args{{"num", len(volumesInDb)}})
	var volumes []resources.VolumeMetadata
	for _, volume := range volumesInDb {
		s.logger.Debug("Volumes from db", logutil.Args{{"volume", volume}})
		volumeMountpoint, err := s.getVolumeMountPoint(volume)
		if err != nil {
			return nil, s.logger.ErrorRet(err, "getVolumeMountPoint failed")
		}

		volumes = append(volumes, resources.VolumeMetadata{Name: volume.Volume.Name, Mountpoint: volumeMountpoint})
	}

	return volumes, nil
}

func (s *scbeLocalClient) createVolume(name string, wwn string, profile string) error {
	defer s.logger.Trace(logutil.DEBUG)()

	if err := s.dataModel.InsertVolume(name, wwn, ""); err != nil {
		return s.logger.ErrorRet(err, "dataModel.InsertVolume failed")
	}

	return nil
}
func (s *scbeLocalClient) getVolumeMountPoint(volume ScbeVolume) (string, error) {
	defer s.logger.Trace(logutil.DEBUG)()

	//TODO return mountpoint
	return "some mount point", nil
}
