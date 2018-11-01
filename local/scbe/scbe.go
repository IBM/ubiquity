/**
 * Copyright 2017 IBM Corp.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package scbe

import (
	"encoding/json"
	"fmt"
	"github.com/IBM/ubiquity/database"
	"github.com/IBM/ubiquity/resources"
	"github.com/IBM/ubiquity/utils"
	"github.com/IBM/ubiquity/utils/logs"
	"strconv"
	"strings"
	"sync"
)

type scbeLocalClient struct {
	logger         logs.Logger
	dataModel      ScbeDataModelWrapper
	isActivated    bool
	config         resources.ScbeConfig
	activationLock *sync.RWMutex
	locker         utils.Locker
	restClients    *sync.Map
}

const (
	OptionNameForServiceName = "profile"
	OptionNameForVolumeSize  = "size"
	volumeNamePrefix         = "u_"
	AttachedToNothing        = "" // during provisioning the volume is not attached to any host
	EmptyHost                = ""
	ComposeVolumeName        = volumeNamePrefix + "%s_%s" // e.g u_instance1_volName
	MaxVolumeNameLength      = 63                         // IBM block storage max volume name cannot exceed this length

	GetVolumeConfigExtraParams = 2 // number of extra params added to the VolumeConfig beyond the scbe volume struct
)

var (
	SupportedFSTypes = []string{"ext4", "xfs"}
)

func NewScbeLocalClient(config resources.ScbeConfig) (resources.StorageClient, error) {
	datamodel := NewScbeDataModelWrapper()
	scbeRestClient, err := NewScbeRestClient(config.ConnectionInfo)
	if err != nil {
		return nil, logs.GetLogger().ErrorRet(err, "NewScbeRestClient failed")
	}
	return NewScbeLocalClientWithNewScbeRestClientAndDataModel(config, datamodel, scbeRestClient)
}

func NewScbeLocalClientWithNewScbeRestClientAndDataModel(config resources.ScbeConfig, dataModel ScbeDataModelWrapper, scbeRestClient ScbeRestClient) (resources.StorageClient, error) {
	if err := validateScbeConfig(&config); err != nil {
		return &scbeLocalClient{}, err
	}

	client := &scbeLocalClient{
		logger:         logs.GetLogger(),
		dataModel:      dataModel,
		config:         config,
		activationLock: &sync.RWMutex{},
		locker:         utils.NewLocker(),
		restClients:    new(sync.Map),
	}

	if err := client.basicScbeLocalClientStartupAndValidation(scbeRestClient); err != nil {
		return &scbeLocalClient{}, err
	}

	return client, nil
}

// basicScbeLocalClientStartup validate config params, login to SCBE and validate default exist
func (s *scbeLocalClient) basicScbeLocalClientStartupAndValidation(restClient ScbeRestClient) error {
	defer s.logger.Trace(logs.DEBUG)()

	// login
	if err := restClient.Login(); err != nil {
		return s.logger.ErrorRet(err, "scbeRestClient.Login() failed")
	}

	s.restClients.Store(s.config.ConnectionInfo.CredentialInfo, restClient)

	// service existence
	s.logger.Info("validate scbeRestClient.ServiceExist", logs.Args{{"DefaultService", s.config.DefaultService}})
	isExist, err := restClient.ServiceExist(s.config.DefaultService)
	if err != nil {
		return s.logger.ErrorRet(err, "scbeRestClient.ServiceExist failed")
	}
	if isExist == false {
		return s.logger.ErrorRet(&activateDefaultServiceError{s.config.DefaultService, s.config.ConnectionInfo.ManagementIP}, "failed")
	}

	// db volume
	volumes, err := restClient.GetVolumes("")
	if err != nil {
		return s.logger.ErrorRet(err, "scbeRestClient.GetVolumes failed")
	}
	for _, volInfo := range volumes {
		if database.IsDatabaseVolume(volInfo.Name) && s.isInstanceVolume(volInfo.Name) {
			volume := &ScbeVolume{
				Volume: resources.Volume{Name: database.VolumeNameSuffix, Backend: resources.SCBE},
				WWN:    volInfo.Wwn,
				FSType: s.config.DefaultFilesystemType,
			}
			s.logger.Info("update db volume", logs.Args{{"volume", volume}})
			s.dataModel.UpdateDatabaseVolume(volume)
			break
		}
	}

	return nil
}

func (s *scbeLocalClient) getAuthenticatedScbeRestClient(credential resources.CredentialInfo) (ScbeRestClient, error) {
	restClient, exists := s.restClients.Load(credential)
	if !exists {
		newClient, err := newScbeRestClientGen(resources.ConnectionInfo{
			CredentialInfo: credential, Port: s.config.ConnectionInfo.Port, ManagementIP: s.config.ConnectionInfo.ManagementIP})
		if err != nil {
			return nil, s.logger.ErrorRet(err, "newScbeRestClientGen failed")
		}
		if err := newClient.Login(); err != nil {
			return nil, s.logger.ErrorRet(err, "newClient.Login() failed")
		}
		restClient, _ = s.restClients.LoadOrStore(credential, newClient)
	}
	return restClient.(ScbeRestClient), nil
}

func (s *scbeLocalClient) isInstanceVolume(volName string) bool {
	defer s.logger.Trace(logs.DEBUG)()
	isInstanceVolume := strings.HasPrefix(volName, fmt.Sprintf(ComposeVolumeName, s.config.UbiquityInstanceName, ""))
	logs.GetLogger().Debug("", logs.Args{{volName, isInstanceVolume}})
	return isInstanceVolume
}

func validateScbeConfig(config *resources.ScbeConfig) error {
	logger := logs.GetLogger()

	if config.DefaultVolumeSize == "" {
		// means customer didn't configure the default
		config.DefaultVolumeSize = resources.DefaultForScbeConfigParamDefaultVolumeSize
		logger.Debug("No DefaultVolumeSize defined in conf file, so set the DefaultVolumeSize to value " + resources.DefaultForScbeConfigParamDefaultVolumeSize)
	}
	_, err := strconv.Atoi(config.DefaultVolumeSize)
	if err != nil {
		return logger.ErrorRet(&ConfigDefaultSizeNotNumError{}, "failed")
	}

	if config.DefaultFilesystemType == "" {
		// means customer didn't configure the default
		config.DefaultFilesystemType = resources.DefaultForScbeConfigParamDefaultFilesystem
		logger.Debug("No DefaultFileSystemType defined in conf file, so set the DefaultFileSystemType to value " + resources.DefaultForScbeConfigParamDefaultFilesystem)
	} else if !utils.StringInSlice(config.DefaultFilesystemType, SupportedFSTypes) {
		return logger.ErrorRet(
			&ConfigDefaultFilesystemTypeNotSupported{
				config.DefaultFilesystemType,
				strings.Join(SupportedFSTypes, ",")}, "failed")
	}

	if len(config.UbiquityInstanceName) > resources.UbiquityInstanceNameMaxSize {
		return logger.ErrorRet(&ConfigScbeUbiquityInstanceNameWrongSize{}, "failed")
	}
	// TODO add more verification on the config file.
	return nil
}

func (s *scbeLocalClient) Activate(activateRequest resources.ActivateRequest) error {
	defer s.logger.Trace(logs.DEBUG)()

	// authenticate
	_, err := s.getAuthenticatedScbeRestClient(activateRequest.CredentialInfo)
	if err != nil {
		return s.logger.ErrorRet(err, "getAuthenticatedScbeRestClient failed")
	}

	s.activationLock.RLock()
	if s.isActivated {
		s.activationLock.RUnlock()
		return nil
	}
	s.activationLock.RUnlock()

	s.activationLock.Lock() //get a write lock to prevent others from repeating these actions
	defer s.activationLock.Unlock()

	// Nothing special to activate SCBE
	s.isActivated = true
	return nil
}

// CreateVolume parse and validate the given options and trigger the volume creation
func (s *scbeLocalClient) CreateVolume(createVolumeRequest resources.CreateVolumeRequest) (err error) {
	defer s.logger.Trace(logs.DEBUG)()

	// authenticate
	scbeRestClient, err := s.getAuthenticatedScbeRestClient(createVolumeRequest.CredentialInfo)
	if err != nil {
		return s.logger.ErrorRet(err, "getAuthenticatedScbeRestClient failed")
	}

	// verify volume does not exist
	if _, err = s.dataModel.GetVolume(createVolumeRequest.Name, false); err != nil {
		return s.logger.ErrorRet(err, "dataModel.GetVolume failed", logs.Args{{"name", createVolumeRequest.Name}})
	}

	// validate size option given
	sizeStr, ok := createVolumeRequest.Opts[OptionNameForVolumeSize]
	if !ok {
		sizeStr = s.config.DefaultVolumeSize
		s.logger.Debug("No size given to create volume, so using the default_size",
			logs.Args{{"volume", createVolumeRequest.Name}, {"default_size", sizeStr}})
	}

	// validate size is a number
	size, err := strconv.Atoi(sizeStr.(string))
	if err != nil {
		return s.logger.ErrorRet(&provisionParamIsNotNumberError{createVolumeRequest.Name, OptionNameForVolumeSize}, "failed")
	}

	// validate fstype option given
	fstypeInt, ok := createVolumeRequest.Opts[resources.OptionNameForVolumeFsType]
	var fstype string
	if !ok {
		fstype = s.config.DefaultFilesystemType
		s.logger.Debug("No default file system type given to create a volume, so using the default_fstype",
			logs.Args{{"volume", createVolumeRequest.Name}, {"default_fstype", fstype}})
	} else {
		fstype = fstypeInt.(string)
	}
	if !utils.StringInSlice(fstype, SupportedFSTypes) {
		return s.logger.ErrorRet(
			&FsTypeNotSupportedError{createVolumeRequest.Name, fstype, strings.Join(SupportedFSTypes, ",")}, "failed")
	}

	// Get the profile option
	profile := s.config.DefaultService
	if createVolumeRequest.Opts[OptionNameForServiceName] != "" && createVolumeRequest.Opts[OptionNameForServiceName] != nil {
		profile = createVolumeRequest.Opts[OptionNameForServiceName].(string)
	}

	// Generate the designated volume name by template
	volNameToCreate := fmt.Sprintf(ComposeVolumeName, s.config.UbiquityInstanceName, createVolumeRequest.Name)

	// Validate volume length ok
	volNamePrefixForCheckLength := fmt.Sprintf(ComposeVolumeName, s.config.UbiquityInstanceName, "")
	volNamePrefixForCheckLengthLen := len(volNamePrefixForCheckLength)
	if len(volNameToCreate) > MaxVolumeNameLength {
		maxVolLength := MaxVolumeNameLength - volNamePrefixForCheckLengthLen // its dynamic because it depends on the UbiquityInstanceName len
		return s.logger.ErrorRet(&VolumeNameExceededMaxLengthError{createVolumeRequest.Name, maxVolLength}, "failed")
	}

	// Provision the volume on SCBE service
	volInfo := ScbeVolumeInfo{}
	volInfo, err = scbeRestClient.CreateVolume(volNameToCreate, profile, size)
	if err != nil {
		return s.logger.ErrorRet(err, "scbeRestClient.CreateVolume failed")
	}

	err = s.dataModel.InsertVolume(createVolumeRequest.Name, volInfo.Wwn, fstype)
	if err != nil {
		return s.logger.ErrorRet(err, "dataModel.InsertVolume failed")
	}

	s.logger.Info("succeeded", logs.Args{{"volume", createVolumeRequest.Name}, {"profile", profile}})
	return nil
}

func (s *scbeLocalClient) RemoveVolume(removeVolumeRequest resources.RemoveVolumeRequest) (err error) {
	defer s.logger.Trace(logs.DEBUG)()

	// authenticate
	scbeRestClient, err := s.getAuthenticatedScbeRestClient(removeVolumeRequest.CredentialInfo)
	if err != nil {
		return s.logger.ErrorRet(err, "getAuthenticatedScbeRestClient failed")
	}

	existingVolume, err := s.dataModel.GetVolume(removeVolumeRequest.Name, true)
	if err != nil {
		return s.logger.ErrorRet(err, "dataModel.GetVolume failed")
	}

	hostAttach, err := scbeRestClient.GetVolMapping(existingVolume.WWN)
	if err != nil {
		return s.logger.ErrorRet(err, "scbeRestClient.GetVolMapping failed")
	}

	if hostAttach != EmptyHost {
		return s.logger.ErrorRet(&CannotDeleteVolWhichAttachedToHostError{removeVolumeRequest.Name, hostAttach}, "failed")
	}

	if err = scbeRestClient.DeleteVolume(existingVolume.WWN); err != nil {
		return s.logger.ErrorRet(err, "scbeRestClient.DeleteVolume failed")
	}
	
	// REMOVE THIS!!!
	return s.logger.ErrorRet(fmt.Errorf("some error"), "######TRYING TO DO IDEMPOTENT ISSUE")

	if err = s.dataModel.DeleteVolume(removeVolumeRequest.Name); err != nil {
		return s.logger.ErrorRet(err, "dataModel.DeleteVolume failed")
	}

	return nil
}

func (s *scbeLocalClient) GetVolume(getVolumeRequest resources.GetVolumeRequest) (resources.Volume, error) {
	defer s.logger.Trace(logs.DEBUG)()

	// authenticate
	_, err := s.getAuthenticatedScbeRestClient(getVolumeRequest.CredentialInfo)
	if err != nil {
		return resources.Volume{}, s.logger.ErrorRet(err, "getAuthenticatedScbeRestClient failed")
	}

	existingVolume, err := s.dataModel.GetVolume(getVolumeRequest.Name, true)
	if err != nil {
		return resources.Volume{}, s.logger.ErrorRet(err, "dataModel.GetVolume failed")
	}

	return resources.Volume{
		Name:       existingVolume.Volume.Name,
		Backend:    existingVolume.Volume.Backend,
		Mountpoint: existingVolume.Volume.Mountpoint}, nil
}

func (s *scbeLocalClient) GetVolumeConfig(getVolumeConfigRequest resources.GetVolumeConfigRequest) (map[string]interface{}, error) {
	defer s.logger.Trace(logs.DEBUG)()

	// authenticate
	scbeRestClient, err := s.getAuthenticatedScbeRestClient(getVolumeConfigRequest.CredentialInfo)
	if err != nil {
		return nil, s.logger.ErrorRet(err, "getAuthenticatedScbeRestClient failed")
	}

	// get volume wwn from name
	scbeVolume, err := s.dataModel.GetVolume(getVolumeConfigRequest.Name, true)
	if err != nil {
		return nil, s.logger.ErrorRet(err, "dataModel.GetVolume failed")
	}

	// get volume full info from scbe
	volumeInfo, err := scbeRestClient.GetVolumes(scbeVolume.WWN)
	if err != nil {
		return nil, s.logger.ErrorRet(err, "scbeRestClient.GetVolumes failed")
	}

	// verify volume is found
	if len(volumeInfo) != 1 {
		return nil, s.logger.ErrorRet(&VolumeNotFoundOnArrayError{VolName: getVolumeConfigRequest.Name}, "failed", logs.Args{{"volumeInfo", volumeInfo}})
	}

	// serialize scbeVolumeInfo to json
	jsonData, err := json.Marshal(volumeInfo[0])
	if err != nil {
		return nil, s.logger.ErrorRet(err, "json.Marshal failed")
	}

	// convert json to map[string]interface{}
	var volConfig map[string]interface{}
	if err = json.Unmarshal(jsonData, &volConfig); err != nil {
		return nil, s.logger.ErrorRet(err, "json.Unmarshal failed")
	}

	// The ubiquity remote will use this extra info to determine the fstype needed to be created on this volume while attaching
	volConfig[resources.OptionNameForVolumeFsType] = scbeVolume.FSType

	// The ubiquity remote will use this extra info to determine is-attached
	hostAttach, err := scbeRestClient.GetVolMapping(scbeVolume.WWN)
	if err != nil {
		return nil, s.logger.ErrorRet(err, "scbeRestClient.GetVolMapping failed")
	}

	volConfig[resources.ScbeKeyVolAttachToHost] = hostAttach

	return volConfig, nil
}

func (s *scbeLocalClient) Attach(attachRequest resources.AttachRequest) (string, error) {
	defer s.logger.Trace(logs.DEBUG)()

	// authenticate
	scbeRestClient, err := s.getAuthenticatedScbeRestClient(attachRequest.CredentialInfo)
	if err != nil {
		return "", s.logger.ErrorRet(err, "getAuthenticatedScbeRestClient failed")
	}

	if attachRequest.Host == EmptyHost {
		return "", s.logger.ErrorRet(
			&InValidRequestError{"attachRequest", "Host", attachRequest.Host, "none empty string"}, "failed")
	}
	if attachRequest.Name == "" {
		return "", s.logger.ErrorRet(
			&InValidRequestError{"attachRequest", "Name", attachRequest.Name, "none empty string"}, "failed")
	}

	existingVolume, err := s.dataModel.GetVolume(attachRequest.Name, true)
	if err != nil {
		return "", s.logger.ErrorRet(err, "dataModel.GetVolume failed")
	}

	hostAttach, err := scbeRestClient.GetVolMapping(existingVolume.WWN)
	if err != nil {
		return "", s.logger.ErrorRet(err, "scbeRestClient.GetVolMapping failed")
	}

	if hostAttach == attachRequest.Host {
		// if already map to the given host then just ignore and succeed to attach
		s.logger.Info("Volume already attached, skip backend attach", logs.Args{{"volume", attachRequest.Name}, {"host", attachRequest.Host}})
		volumeMountpoint := fmt.Sprintf(resources.PathToMountUbiquityBlockDevices, existingVolume.WWN)
		return volumeMountpoint, nil
	} else if hostAttach != "" {
		return "", s.logger.ErrorRet(&volAlreadyAttachedError{attachRequest.Name, hostAttach}, "failed")
	}

	// Lock will ensure no other caller attach a volume from the same host concurrently, Prevent SCBE race condition on get next available lun ID
	s.locker.WriteLock(attachRequest.Host)
	s.logger.Debug("Attaching", logs.Args{{"volume", existingVolume}})
	if _, err = scbeRestClient.MapVolume(existingVolume.WWN, attachRequest.Host); err != nil {
		s.locker.WriteUnlock(attachRequest.Host)
		return "", s.logger.ErrorRet(err, "scbeRestClient.MapVolume failed")
	}
	s.locker.WriteUnlock(attachRequest.Host)

	volumeMountpoint := fmt.Sprintf(resources.PathToMountUbiquityBlockDevices, existingVolume.WWN)
	return volumeMountpoint, nil
}

func (s *scbeLocalClient) Detach(detachRequest resources.DetachRequest) (err error) {
	defer s.logger.Trace(logs.DEBUG)()
	host2detach := detachRequest.Host

	// authenticate
	scbeRestClient, err := s.getAuthenticatedScbeRestClient(detachRequest.CredentialInfo)
	if err != nil {
		return s.logger.ErrorRet(err, "getAuthenticatedScbeRestClient failed")
	}

	existingVolume, err := s.dataModel.GetVolume(detachRequest.Name, true)
	if err != nil {
		return s.logger.ErrorRet(err, "dataModel.GetVolume failed")
	}

	hostAttach, err := scbeRestClient.GetVolMapping(existingVolume.WWN)
	if err != nil {
		return s.logger.ErrorRet(err, "scbeRestClient.GetVolMapping failed")
	}

	if hostAttach == EmptyHost {
		s.logger.Warning("Volume is already detached from host.", logs.Args{{"volume", existingVolume.WWN}})
		return nil
	}

	// TODO idempotent, if volume attach to different host, then we should also return succeed.
	s.logger.Debug("Detaching", logs.Args{{"volume", existingVolume}})
	if err = scbeRestClient.UnmapVolume(existingVolume.WWN, host2detach); err != nil {
		return s.logger.ErrorRet(err, "scbeRestClient.UnmapVolume failed")
	}

	return nil
}

func (s *scbeLocalClient) ListVolumes(listVolumesRequest resources.ListVolumesRequest) ([]resources.Volume, error) {
	defer s.logger.Trace(logs.DEBUG)()
	var err error

	// authenticate
	_, err = s.getAuthenticatedScbeRestClient(listVolumesRequest.CredentialInfo)
	if err != nil {
		return nil, s.logger.ErrorRet(err, "getAuthenticatedScbeRestClient failed")
	}

	volumesInDb, err := s.dataModel.ListVolumes()
	if err != nil {
		return nil, s.logger.ErrorRet(err, "dataModel.ListVolumes failed")
	}

	s.logger.Debug("Volumes in db", logs.Args{{"num", len(volumesInDb)}})
	var volumes []resources.Volume
	for _, volume := range volumesInDb {
		s.logger.Debug("Volumes from db", logs.Args{{"volume", volume}})
		volumes = append(volumes, volume.Volume)
	}

	return volumes, nil
}

func (s *scbeLocalClient) getVolumeMountPoint(volume ScbeVolume) (string, error) {
	defer s.logger.Trace(logs.DEBUG)()

	//TODO return mountpoint
	return "some mount point", nil
}
