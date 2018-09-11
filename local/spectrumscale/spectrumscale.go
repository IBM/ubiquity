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

package spectrumscale

import (
    "github.com/IBM/ubiquity/utils/logs"
	"github.com/IBM/ubiquity/utils"
	"path"
	"fmt"
	"github.com/IBM/ubiquity/local/spectrumscale/connectors"
	"sync"
	"github.com/IBM/ubiquity/resources"
	"os"
)

type spectrumLocalClient struct {
	logger         logs.Logger
	connector      connectors.SpectrumScaleConnector
	dataModel      SpectrumDataModelWrapper
	executor       utils.Executor
	isActivated    bool
	isMounted      bool
	config         resources.SpectrumScaleConfig
	activationLock *sync.RWMutex
}

const (
	Type            string = "type"
	TypeFileset     string = "fileset"

	FilesetID string = "fileset"
	Directory string = "directory"
	Quota     string = "quota"

	Filesystem string = "filesystem"

	IsPreexisting string = "isPreexisting"

	Cluster string = "clusterId"
)

func NewSpectrumLocalClient(config resources.UbiquityServerConfig) (resources.StorageClient, error) {
	if config.SpectrumScaleConfig.DefaultFilesystemName == "" {
		return nil, fmt.Errorf("spectrumLocalClient: init: missing required parameter 'spectrumDefaultFileSystem'")
	}

	return newSpectrumLocalClient(config.SpectrumScaleConfig, resources.SpectrumScale)
}

func NewSpectrumLocalClientWithConnectors(logger logs.Logger, connector connectors.SpectrumScaleConnector, spectrumExecutor utils.Executor, config resources.SpectrumScaleConfig, datamodel SpectrumDataModelWrapper) (resources.StorageClient, error) {
	err := datamodel.CreateVolumeTable()
	if err != nil {
		return &spectrumLocalClient{}, err
	}
	return &spectrumLocalClient{logger: logger, connector: connector, dataModel: datamodel, executor: spectrumExecutor, config: config, activationLock: &sync.RWMutex{}}, nil
}

func newSpectrumLocalClient(config resources.SpectrumScaleConfig, backend string) (*spectrumLocalClient, error) {
    logger := logs.GetLogger()
    defer s.logger.Trace(logs.DEBUG)()

	client, err := connectors.GetSpectrumScaleConnector(logger, config)
	if err != nil {
        logger.Debug("", logs.Args{{"Error", err.Error()}})
		return &spectrumLocalClient{}, err
	}
	datamodel := NewSpectrumDataModelWrapper(backend)

	dbName := datamodel.GetDbName()

	logger.Debug("Going to check for DB volume fileset", logs.Args{{"Filesystem", config.DefaultFilesystemName}, {"Fileset", dbName}})

	volume, err := client.ListFileset(config.DefaultFilesystemName, dbName)

	if err == nil {
		logger.Debug("DB volume fileset present")
		fsMountpoint, err := client.GetFilesystemMountpoint(config.DefaultFilesystemName)
	        if err != nil {
			logger.Debug("Error while getting filesystem", logs.Args{{"Filesystem", config.DefaultFilesystemName}})
	                return &spectrumLocalClient{}, err
	        }
		mountLoc := path.Join(fsMountpoint, dbName)
		logger.Debug("Mount Location for DB", logs.Args{{"Mount", mountLoc}})

		scaleDbVol := &SpectrumScaleVolume{Volume: resources.Volume{Name: volume.Name, Backend: backend, Mountpoint: mountLoc}, Type: Fileset, FileSystem: config.DefaultFilesystemName, Fileset: dbName}
		datamodel.UpdateDatabaseVolume(scaleDbVol)
	} else {
		logger.Debug("DB Vol Fileset Not Found", logs.Args{{"Filesystem", config.DefaultFilesystemName}, {"Fileset", dbName}})

	}
	return &spectrumLocalClient{logger: logger, connector: client, dataModel: datamodel, config: config, executor: utils.NewExecutor(), activationLock: &sync.RWMutex{}}, nil
}

func (s *spectrumLocalClient) Activate(activateRequest resources.ActivateRequest) (err error) {
    defer s.logger.Trace(logs.DEBUG)()

	s.activationLock.RLock()
	if s.isActivated {
		s.activationLock.RUnlock()
		return nil
	}
	s.activationLock.RUnlock()

	s.activationLock.Lock() //get a write lock to prevent others from repeating these actions
	defer s.activationLock.Unlock()

	//check if filesystem is mounted
	mounted, err := s.connector.IsFilesystemMounted(s.config.DefaultFilesystemName)

	if err != nil {
		s.logger.Debug("", logs.Args{{"Error",err.Error()}})
		return err
	}

	if mounted == false {
		err = s.connector.MountFileSystem(s.config.DefaultFilesystemName)

		if err != nil {
			s.logger.Debug("", logs.Args{{"Error",err.Error()}})
			return err
		}
	}

	clusterId, err := s.connector.GetClusterId()

	if err != nil {
		s.logger.Debug("", logs.Args{{"Error", err.Error()}})
		return err
	}

	if len(clusterId) == 0 {
		clusterIdErr := fmt.Errorf("Unable to retrieve clusterId: clusterId is empty")
		s.logger.Debug("", logs.Args{{"Error", clusterIdErr.Error()}})
		return clusterIdErr
	}

	s.dataModel.SetClusterId(clusterId)

	s.isActivated = true
	return nil
}

func (s *spectrumLocalClient) CreateVolume(createVolumeRequest resources.CreateVolumeRequest) (err error) {
    defer s.logger.Trace(logs.DEBUG)()

	_, volExists, err := s.dataModel.GetVolume(createVolumeRequest.Name)

	if err != nil {
		s.logger.Debug("", logs.Args{{"Error", err.Error()}})
		return err
	}

	if volExists {
		return fmt.Errorf("Volume already exists")
	}

	s.logger.Debug("Opts for create:", logs.Args{{"Opts", createVolumeRequest.Opts}})

	if len(createVolumeRequest.Opts) == 0 {
		//fileset
		return s.createFilesetVolume(s.config.DefaultFilesystemName, createVolumeRequest.Name, createVolumeRequest.Opts)
	}
	s.logger.Debug("Trying to determine type for request\n")
	userSpecifiedType, err := determineTypeFromRequest(s.logger, createVolumeRequest.Opts)
	if err != nil {
		s.logger.Debug("Error determining type", logs.Args{{"Error", err.Error()}})
		return err
	}
	s.logger.Debug("Volume type requested", logs.Args{{"userSpecifiedType",userSpecifiedType}})
	isExistingVolume, filesystem, existingFileset, err := s.validateAndParseParams(s.logger, createVolumeRequest.Opts)
	if err != nil {
		s.logger.Debug("Error in validate params", logs.Args{{"Error", err.Error()}})
		return err
	}


	if isExistingVolume && userSpecifiedType == TypeFileset {
		quota, quotaSpecified := createVolumeRequest.Opts[Quota]
		if quotaSpecified {
			return s.updateDBWithExistingFilesetQuota(filesystem, createVolumeRequest.Name, existingFileset, quota.(string), createVolumeRequest.Opts)
		}
		return s.updateDBWithExistingFileset(filesystem, createVolumeRequest.Name, existingFileset, createVolumeRequest.Opts)
	}

	if userSpecifiedType == TypeFileset {
		quota, quotaSpecified := createVolumeRequest.Opts[Quota]
		if quotaSpecified {
			return s.createFilesetQuotaVolume(filesystem, createVolumeRequest.Name, quota.(string), createVolumeRequest.Opts)
		}
		return s.createFilesetVolume(filesystem, createVolumeRequest.Name, createVolumeRequest.Opts)
	}
	return fmt.Errorf("Internal error")
}

func (s *spectrumLocalClient) RemoveVolume(removeVolumeRequest resources.RemoveVolumeRequest) (err error) {
    defer s.logger.Trace(logs.DEBUG)()

	existingVolume, volExists, err := s.dataModel.GetVolume(removeVolumeRequest.Name)

	if err != nil {
		s.logger.Debug("",logs.Args{{"Error", err.Error()}})
		return err
	}

	if volExists == false {
		return fmt.Errorf("Volume not found")
	}

	isFilesetLinked, err := s.connector.IsFilesetLinked(existingVolume.FileSystem, existingVolume.Fileset)

	if err != nil {
		s.logger.Debug("", logs.Args{{"Error", err.Error()}})
		return err
	}

	if isFilesetLinked {
		err := s.connector.UnlinkFileset(existingVolume.FileSystem, existingVolume.Fileset)

		if err != nil {
			s.logger.Debug("", logs.Args{{"Error", err.Error()}})
			return err
		}
	}
	err = s.dataModel.DeleteVolume(removeVolumeRequest.Name)

	if err != nil {
		s.logger.Debug("", logs.Args{{"Error", err.Error()}})
		return err
	}
	if s.config.ForceDelete == true && existingVolume.IsPreexisting == false {
		err = s.connector.DeleteFileset(existingVolume.FileSystem, existingVolume.Fileset)

		if err != nil {
		    s.logger.Debug("", logs.Args{{"Error", err.Error()}})
			return err
		}
	}

	return nil
}

func (s *spectrumLocalClient) GetVolume(getVolumeRequest resources.GetVolumeRequest) (resources.Volume, error) {
    defer s.logger.Trace(logs.DEBUG)()

	existingVolume, volExists, err := s.dataModel.GetVolume(getVolumeRequest.Name)
	if err != nil {
		return resources.Volume{}, err
	}
	if volExists == false {
		return resources.Volume{}, fmt.Errorf("Volume not found")
	}

	return resources.Volume{Name: existingVolume.Volume.Name, Backend: existingVolume.Volume.Backend, Mountpoint: existingVolume.Volume.Mountpoint}, nil
}

func (s *spectrumLocalClient) GetVolumeConfig(getVolumeConfigRequest resources.GetVolumeConfigRequest) (volumeConfigDetails map[string]interface{}, err error) {
    defer s.logger.Trace(logs.DEBUG)()

	existingVolume, volExists, err := s.dataModel.GetVolume(getVolumeConfigRequest.Name)

	if err != nil {
		s.logger.Debug("", logs.Args{{"Error", err.Error()}})
		return nil, err
	}

	if volExists {
		volumeConfigDetails = make(map[string]interface{})
		volumeMountpoint, err := s.getVolumeMountPoint(existingVolume)
		if err != nil {
			s.logger.Debug("", logs.Args{{"Error", err.Error()}})
			return nil, err
		}

		isFilesetLinked, err := s.connector.IsFilesetLinked(existingVolume.FileSystem, existingVolume.Fileset)
		if err != nil {
			s.logger.Debug("", logs.Args{{"Error", err.Error()}})
			return nil, err
		}
		if isFilesetLinked {
			volumeConfigDetails["mountpoint"] = volumeMountpoint
		}

		volumeConfigDetails[FilesetID] = existingVolume.Fileset
		volumeConfigDetails[Filesystem] = existingVolume.FileSystem
		volumeConfigDetails[Cluster] = existingVolume.ClusterId
		if existingVolume.GID != "" {
			volumeConfigDetails[UserSpecifiedGID] = existingVolume.GID
		}
		if existingVolume.UID != "" {
			volumeConfigDetails[UserSpecifiedUID] = existingVolume.UID
		}
		volumeConfigDetails[IsPreexisting] = existingVolume.IsPreexisting
		volumeConfigDetails[Type] = existingVolume.Type

		return volumeConfigDetails, nil
	}
	return nil, fmt.Errorf("Volume not found")
}

func (s *spectrumLocalClient) Attach(attachRequest resources.AttachRequest) (volumeMountpoint string, err error) {
    defer s.logger.Trace(logs.DEBUG)()

	existingVolume, volExists, err := s.dataModel.GetVolume(attachRequest.Name)

	if err != nil {
		s.logger.Debug("", logs.Args{{"Error", err.Error()}})
		return "", err
	}

	if !volExists {
		return "", fmt.Errorf("Volume not found")
	}

	volumeMountpoint, err = s.getVolumeMountPoint(existingVolume)

	if err != nil {
		s.logger.Debug("", logs.Args{{"Error", err.Error()}})
		return "", err
	}

	isFilesetLinked, err := s.connector.IsFilesetLinked(existingVolume.FileSystem, existingVolume.Fileset)

	if err != nil {
		s.logger.Debug("", logs.Args{{"Error", err.Error()}})
		return "", err
	}

	if isFilesetLinked == false {
		err = s.connector.LinkFileset(existingVolume.FileSystem, existingVolume.Fileset)

		if err != nil {
			s.logger.Debug("", logs.Args{{"Error", err.Error()}})
			return "", err
		}
	}

	existingVolume.Volume.Mountpoint = volumeMountpoint

	err = s.dataModel.UpdateVolumeMountpoint(attachRequest.Name, volumeMountpoint)

	if err != nil {
		s.logger.Debug("", logs.Args{{"Error", err.Error()}})
		return "", err
	}

	return volumeMountpoint, nil
}

func (s *spectrumLocalClient) Detach(detachRequest resources.DetachRequest) (err error) {
    defer s.logger.Trace(logs.DEBUG)()

	existingVolume, volExists, err := s.dataModel.GetVolume(detachRequest.Name)

	if err != nil {
		s.logger.Debug("", logs.Args{{"Error", err.Error()}})
		return err
	}

	if !volExists {
		return fmt.Errorf("Volume not found")
	}

	_, err = s.getVolumeMountPoint(existingVolume)
	if err != nil {
		return err
	}
	isFilesetLinked, err := s.connector.IsFilesetLinked(existingVolume.FileSystem, existingVolume.Fileset)

	if err != nil {
		s.logger.Debug("", logs.Args{{"Error", err.Error()}})
		return err
	}
	if isFilesetLinked == false {
		return fmt.Errorf("volume not attached")
	}

	err = s.dataModel.UpdateVolumeMountpoint(detachRequest.Name, "")
	if err != nil {
		s.logger.Debug("", logs.Args{{"Error", err.Error()}})
		return err
	}

	return nil
}

func (s *spectrumLocalClient) ListVolumes(listVolumesRequest resources.ListVolumesRequest) ([]resources.Volume, error) {
    defer s.logger.Trace(logs.DEBUG)()

	var err error

	volumesInDb, err := s.dataModel.ListVolumes()

	if err != nil {
		s.logger.Debug("error retrieving volumes from db", logs.Args{{"Error",err}})
		return nil, err
	}
	//s.logger.Printf("Volumes in db: %d\n", len(volumesInDb))
	//var volumes []resources.Volume
	//for _, volume := range volumesInDb {
	//	s.logger.Printf("Volume from db: %#v\n", volume)
	//
	//	volumeMountpoint, err := s.getVolumeMountPoint(volume)
	//	if err != nil {
	//		s.logger.Println(err.Error())
	//		return nil, err
	//	}
	//
	//	volumes = append(volumes, resources.Volume{Name: volume.Name, Mountpoint: volumeMountpoint})
	//}

	return volumesInDb, nil
}

func (s *spectrumLocalClient) createFilesetVolume(filesystem, name string, opts map[string]interface{}) error {
    defer s.logger.Trace(logs.DEBUG)()

	filesetName := generateFilesetName(name)

	var dbVolerr error
	var err error
	if s.dataModel.IsDbVolume(name) {
		/* Check if fileset is present */
		s.logger.Debug("DB Volume, check if present",logs.Args{{"VolumeName",name},{"Filesystem",filesystem}})
		_, dbVolerr = s.connector.ListFileset(filesystem, name)
	}

	if !s.dataModel.IsDbVolume(name) || dbVolerr != nil {

		s.logger.Debug("Create fileset: DB volume is not present OR it is normal volume creation", logs.Args{{"VolumeName", filesetName}, {"Filesystem", filesystem}})
		err := s.connector.CreateFileset(filesystem, filesetName, opts)

		if  err != nil {
			s.logger.Debug("Error creating fileset", logs.Args{{"Error", err}})
			return err
		}
	}

	err = s.dataModel.InsertFilesetVolume(filesetName, name, filesystem, false, opts)

	if err != nil {
		s.logger.Debug("Error inserting fileset", logs.Args{{"Error", err}})
		return err
	}

	s.logger.Debug("Created fileset volume with fileset ", logs.Args{{"filesetName", filesetName}})
	return nil
}

func (s *spectrumLocalClient) createFilesetQuotaVolume(filesystem, name, quota string, opts map[string]interface{}) error {
    defer s.logger.Trace(logs.DEBUG)()

	filesetName := generateFilesetName(name)
        var dbVolerr error
	var err error
        if s.dataModel.IsDbVolume(name) {
                /* Check if fileset is present */
		s.logger.Debug("DB Volume, check if present",logs.Args{{"VolumeName",name},{"Filesystem",filesystem}})

                _, dbVolerr = s.connector.ListFileset(filesystem, name)
        }   
        
        if !s.dataModel.IsDbVolume(name) || dbVolerr != nil {

		s.logger.Debug("Create fileset: DB volume is not present OR it is normal volume creation", logs.Args{{"VolumeName", filesetName}, {"Filesystem", filesystem}})
                err := s.connector.CreateFileset(filesystem, filesetName, opts)
          
                if  err != nil {
                        s.logger.Debug("Error creating fileset", logs.Args{{"Error", err}})       
                        return err
                }
		err = s.connector.SetFilesetQuota(filesystem, filesetName, quota)

		if err != nil {
			deleteErr := s.connector.DeleteFileset(filesystem, filesetName)
			if deleteErr != nil {
				return fmt.Errorf("Error setting quota (rollback error on delete fileset %s - manual cleanup needed)", filesetName)
			}
			return err
		}
	}

	err = s.dataModel.InsertFilesetQuotaVolume(filesetName, quota, name, filesystem, false, opts)

	if err != nil {
		return err
	}

	s.logger.Debug("Created fileset volume with fileset quota ", logs.Args{{"filesetName", filesetName},{"quota", quota}})
	return nil
}

func generateFilesetName(name string) string {
	//TODO: placeholder for now
	return name
	//return strconv.FormatInt(time.Now().UnixNano(), 10)
}

//TODO move updates to DB file

func (s *spectrumLocalClient) updateDBWithExistingFileset(filesystem, name, userSpecifiedFileset string, opts map[string]interface{}) error {
    defer s.logger.Trace(logs.DEBUG)()

	s.logger.Debug("User specified fileset", logs.Args{{"userSpecifiedFileset", userSpecifiedFileset}})

	_, err := s.connector.ListFileset(filesystem, userSpecifiedFileset)
	if err != nil {
		s.logger.Debug("Fileset does not exist",logs.Args{{"Error", err.Error()}})
		return err
	}

	err = s.dataModel.InsertFilesetVolume(userSpecifiedFileset, name, filesystem, true, opts)

	if err != nil {
		s.logger.Debug("",logs.Args{{"Error", err.Error()}})
		return err
	}
	return nil
}

func (s *spectrumLocalClient) checkIfVolumeExistsInDB(name, userSpecifiedFileset string) error {
    defer s.logger.Trace(logs.DEBUG)()
	getVolumeConfigRequest := resources.GetVolumeConfigRequest{Name: name}
	volumeConfigDetails, err := s.GetVolumeConfig(getVolumeConfigRequest)

	if err != nil {
		s.logger.Debug("",logs.Args{{"Error", err.Error()}})
		return err
	}

	if volumeConfigDetails["FilesetId"] != userSpecifiedFileset {
		return fmt.Errorf("volume %s with fileset %s not found", name, userSpecifiedFileset)
	}
	return nil
}

func (s *spectrumLocalClient) updateDBWithExistingFilesetQuota(filesystem, name, userSpecifiedFileset, quota string, opts map[string]interface{}) error {
    defer s.logger.Trace(logs.DEBUG)()

	filesetQuota, err := s.connector.ListFilesetQuota(filesystem, userSpecifiedFileset)

	if err != nil {
		s.logger.Debug("",logs.Args{{"Error", err.Error()}})
		return err
	}

	if s.config.RestConfig.ManagementIP != "" {
		s.logger.Debug("For REST connector converting quotas to bytes\n")
		filesetQuotaBytes, err := utils.ConvertToBytes(s.logger, filesetQuota)
		if err != nil {
			s.logger.Debug("utils.ConvertToBytes failed", logs.Args{{"Error", err}})
			return err
		}

		quotasBytes, err := utils.ConvertToBytes(s.logger, quota)
		if err != nil {
            s.logger.Debug("utils.ConvertToBytes failed ", logs.Args{{"Error", err}})
			return err
		}

		if filesetQuotaBytes != quotasBytes {
			return fmt.Errorf("Mismatch between user-specified %v and listed quota %v for fileset %s", quotasBytes, filesetQuotaBytes, userSpecifiedFileset)
		}
	} else {
		if filesetQuota != quota {
			return fmt.Errorf("Mismatch between user-specified and listed quota for fileset %s", userSpecifiedFileset)

		}
	}

	err = s.dataModel.InsertFilesetQuotaVolume(userSpecifiedFileset, quota, name, filesystem, true, opts)

	if err != nil {
		s.logger.Debug("", logs.Args{{"Error", err.Error()}})
		return err
	}
	return nil
}

func determineTypeFromRequest(logger logs.Logger, opts map[string]interface{}) (string, error) {
    defer s.logger.Trace(logs.DEBUG)()

	userSpecifiedType, exists := opts[Type]
	if exists == false {
		return TypeFileset, nil
	}

	if userSpecifiedType.(string) != TypeFileset {
		return "", fmt.Errorf("Unknown 'type' = %s specified", userSpecifiedType.(string))
	}

	return userSpecifiedType.(string), nil
}

func (s *spectrumLocalClient) validateAndParseParams(logger logs.Logger, opts map[string]interface{}) (bool, string, string, error) {
    defer s.logger.Trace(logs.DEBUG)()

	existingFileset, existingFilesetSpecified := opts[TypeFileset]
	filesystem, filesystemSpecified := opts[Filesystem]
	_, uidSpecified := opts[UserSpecifiedUID]
	_, gidSpecified := opts[UserSpecifiedGID]

	userSpecifiedType, err := determineTypeFromRequest(logger, opts)
	if err != nil {
		logger.Debug("", logs.Args{{"Error", err.Error()}})
		return false, "", "", err
	}

	if uidSpecified && gidSpecified {
		if existingFilesetSpecified{
			return true, "", "", fmt.Errorf("uid/gid cannot be specified along with existing fileset")
		}
	}

	if (userSpecifiedType == TypeFileset && existingFilesetSpecified){
		if filesystemSpecified == false {
			logger.Debug("'filesystem' is a required opt for using existing volumes")
			return true, filesystem.(string), existingFileset.(string), fmt.Errorf("'filesystem' is a required opt for using existing volumes")
		}
		logger.Debug("Valid: existing FILESET")
		return true, filesystem.(string), existingFileset.(string), nil

	} else if filesystemSpecified == false {
		return false, s.config.DefaultFilesystemName, "", nil

		}else {
			return false, filesystem.(string), "", nil
		}
}

func (s *spectrumLocalClient) getVolumeMountPoint(volume SpectrumScaleVolume) (string, error) {
    defer s.logger.Trace(logs.DEBUG)()

	fsMountpoint, err := s.connector.GetFilesystemMountpoint(volume.FileSystem)
	if err != nil {
		return "", err
	}

	//isFilesetLinked, err := s.connector.IsFilesetLinked(volume.FileSystem, volume.Fileset)
	//
	//if err != nil {
	//	s.logger.Println(err.Error())
	//	return "", err
	//}

	return path.Join(fsMountpoint, volume.Fileset), nil

}
func (s *spectrumLocalClient) updatePermissions(name string) error {
    defer s.logger.Trace(logs.DEBUG)()

	getVolumeConfigRequest := resources.GetVolumeConfigRequest{Name: name}
	volumeConfig, err := s.GetVolumeConfig(getVolumeConfigRequest)
	if err != nil {
		return err
	}
	filesystem, exists := volumeConfig[Filesystem]
	if exists == false {
		return fmt.Errorf("Cannot determine filesystem for volume: %s", name)
	}
	fsMountpoint, err := s.connector.GetFilesystemMountpoint(filesystem.(string))
	if err != nil {
		return err
	}
	volumeType, exists := volumeConfig[Type]
	if exists == false {
		return fmt.Errorf("Cannot determine type for volume: %s", name)
	}
	fileset, exists := volumeConfig[FilesetID]
	if exists == false {
		return fmt.Errorf("Cannot determine filesetId for volume: %s", name)
	}
	// executor := utils.NewExecutor() // TODO check why its here ( #39: new logger in block_device_mounter_utils)
	filesetPath := path.Join(fsMountpoint, fileset.(string))
	//chmod 777 mountpoint
	args := []string{"777", filesetPath}
	_, err = s.executor.Execute("chmod", args)
	if err != nil {
		s.logger.Debug("Failed to change permissions of filesetpath", logs.Args{{"filesetPath", filesetPath}, {"Error", err.Error()}})
		return err
	}
	return nil
}
