package spectrumscale

import (
	"log"

	"github.com/IBM/ubiquity/utils"

	"os"
	"path"

	"fmt"

	"github.com/IBM/ubiquity/local/spectrumscale/connectors"
	"github.com/jinzhu/gorm"

	"sync"

	"github.com/IBM/ubiquity/resources"
)

type spectrumLocalClient struct {
	logger         *log.Logger
	connector      connectors.SpectrumScaleConnector
	dataModel      SpectrumDataModel
	executor       utils.Executor
	isActivated    bool
	isMounted      bool
	config         resources.SpectrumScaleConfig
	activationLock *sync.RWMutex
}

const (
	TYPE             string = "type"
	TYPE_FILESET     string = "fileset"
	TYPE_LIGHTWEIGHT string = "lightweight"

	FILESETID string = "fileset"
	DIRECTORY string = "directory"
	QUOTA     string = "quota"

	FILESYSTEM string = "filesystem"

	IS_PREEXISTING string = "isPreexisting"

	CLUSTER string = "clusterId"
)

func NewSpectrumLocalClient(logger *log.Logger, config resources.SpectrumScaleConfig, database *gorm.DB) (resources.StorageClient, error) {
	if config.ConfigPath == "" {
		return nil, fmt.Errorf("spectrumLocalClient: init: missing required parameter 'spectrumConfigPath'")
	}
	if config.DefaultFilesystemName == "" {
		return nil, fmt.Errorf("spectrumLocalClient: init: missing required parameter 'spectrumDefaultFileSystem'")
	}
	return newSpectrumLocalClient(logger, config, database, resources.SPECTRUM_SCALE)
}

func NewSpectrumLocalClientWithConnectors(logger *log.Logger, connector connectors.SpectrumScaleConnector, spectrumExecutor utils.Executor, config resources.SpectrumScaleConfig, datamodel SpectrumDataModel) (resources.StorageClient, error) {
	err := datamodel.CreateVolumeTable()
	if err != nil {
		return &spectrumLocalClient{}, err
	}
	return &spectrumLocalClient{logger: logger, connector: connector, dataModel: datamodel, executor: spectrumExecutor, config: config, activationLock: &sync.RWMutex{}}, nil
}

func newSpectrumLocalClient(logger *log.Logger, config resources.SpectrumScaleConfig, database *gorm.DB, backend resources.Backend) (*spectrumLocalClient, error) {
	logger.Println("spectrumLocalClient: init start")
	defer logger.Println("spectrumLocalClient: init end")
	client, err := connectors.GetSpectrumScaleConnector(logger, config)
	if err != nil {
		logger.Fatalln(err.Error())
		return &spectrumLocalClient{}, err
	}
	datamodel := NewSpectrumDataModel(logger, database, backend)
	err = datamodel.CreateVolumeTable()
	if err != nil {
		return &spectrumLocalClient{}, err
	}
	return &spectrumLocalClient{logger: logger, connector: client, dataModel: datamodel, config: config, executor: utils.NewExecutor(logger), activationLock: &sync.RWMutex{}}, nil
}

func (s *spectrumLocalClient) Activate() (err error) {
	s.logger.Println("spectrumLocalClient: Activate start")
	defer s.logger.Println("spectrumLocalClient: Activate end")

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
		s.logger.Println(err.Error())
		return err
	}

	if mounted == false {
		err = s.connector.MountFileSystem(s.config.DefaultFilesystemName)

		if err != nil {
			s.logger.Println(err.Error())
			return err
		}
	}

	clusterId, err := s.connector.GetClusterId()

	if err != nil {
		s.logger.Println(err.Error())
		return err
	}

	if len(clusterId) == 0 {
		clusterIdErr := fmt.Errorf("Unable to retrieve clusterId: clusterId is empty")
		s.logger.Println(clusterIdErr.Error())
		return clusterIdErr
	}

	s.dataModel.SetClusterId(clusterId)

	s.isActivated = true
	return nil
}

func (s *spectrumLocalClient) CreateVolume(name string, opts map[string]interface{}) (err error) {
	s.logger.Println("spectrumLocalClient: create start")
	defer s.logger.Println("spectrumLocalClient: create end")

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
		//fileset
		return s.createFilesetVolume(s.config.DefaultFilesystemName, name, opts)
	}
	s.logger.Printf("Trying to determine type for request\n")
	userSpecifiedType, err := determineTypeFromRequest(s.logger, opts)
	if err != nil {
		s.logger.Printf("Error determining type: %s\n", err.Error())
		return err
	}
	s.logger.Printf("Volume type requested: %s", userSpecifiedType)
	isExistingVolume, filesystem, existingFileset, existingLightWeightDir, err := s.validateAndParseParams(s.logger, opts)
	if err != nil {
		s.logger.Printf("Error in validate params: %s\n", err.Error())
		return err
	}

	s.logger.Printf("Params for create: %s,%s,%s,%s\n", isExistingVolume, filesystem, existingFileset, existingLightWeightDir)

	if isExistingVolume && userSpecifiedType == TYPE_FILESET {
		quota, quotaSpecified := opts[QUOTA]
		if quotaSpecified {
			return s.updateDBWithExistingFilesetQuota(filesystem, name, existingFileset, quota.(string), opts)
		}
		return s.updateDBWithExistingFileset(filesystem, name, existingFileset, opts)
	}

	if isExistingVolume && userSpecifiedType == TYPE_LIGHTWEIGHT {
		return s.updateDBWithExistingDirectory(filesystem, name, existingFileset, existingLightWeightDir, opts)
	}

	if userSpecifiedType == TYPE_FILESET {
		quota, quotaSpecified := opts[QUOTA]
		if quotaSpecified {
			return s.createFilesetQuotaVolume(filesystem, name, quota.(string), opts)
		}
		return s.createFilesetVolume(filesystem, name, opts)
	}
	if userSpecifiedType == TYPE_LIGHTWEIGHT {
		return s.createLightweightVolume(filesystem, name, existingFileset, opts)
	}
	return fmt.Errorf("Internal error")
}

func (s *spectrumLocalClient) RemoveVolume(name string) (err error) {
	s.logger.Println("spectrumLocalClient: remove start")
	defer s.logger.Println("spectrumLocalClient: remove end")

	existingVolume, volExists, err := s.dataModel.GetVolume(name)

	if err != nil {
		s.logger.Println(err.Error())
		return err
	}

	if volExists == false {
		return fmt.Errorf("Volume not found")
	}

	if existingVolume.Type == LIGHTWEIGHT {
		err = s.dataModel.DeleteVolume(name)
		if err != nil {
			s.logger.Println(err.Error())
			return err
		}

		if s.config.ForceDelete == true && existingVolume.IsPreexisting == false {
			mountpoint, err := s.connector.GetFilesystemMountpoint(existingVolume.FileSystem)
			if err != nil {
				s.logger.Println(err.Error())
				return err
			}
			lightweightVolumePath := path.Join(mountpoint, existingVolume.Fileset, existingVolume.Directory)

			err = s.executor.RemoveAll(lightweightVolumePath)

			if err != nil {
				s.logger.Println(err.Error())
				return err
			}
		}
		return nil
	}

	isFilesetLinked, err := s.connector.IsFilesetLinked(existingVolume.FileSystem, existingVolume.Fileset)

	if err != nil {
		s.logger.Println(err.Error())
		return err
	}

	if isFilesetLinked {
		err := s.connector.UnlinkFileset(existingVolume.FileSystem, existingVolume.Fileset)

		if err != nil {
			s.logger.Println(err.Error())
			return err
		}
	}
	err = s.dataModel.DeleteVolume(name)

	if err != nil {
		s.logger.Println(err.Error())
		return err
	}
	if s.config.ForceDelete == true && existingVolume.IsPreexisting == false {
		err = s.connector.DeleteFileset(existingVolume.FileSystem, existingVolume.Fileset)

		if err != nil {
			s.logger.Println(err.Error())
			return err
		}
	}

	return nil
}

func (s *spectrumLocalClient) GetVolume(name string) (resources.Volume, error) {
	s.logger.Println("spectrumLocalClient: GetVolume start")
	defer s.logger.Println("spectrumLocalClient: GetVolume finish")

	existingVolume, volExists, err := s.dataModel.GetVolume(name)
	if err != nil {
		return resources.Volume{}, err
	}
	if volExists == false {
		return resources.Volume{}, fmt.Errorf("Volume not found")
	}

	return resources.Volume{Name: existingVolume.Volume.Name, Backend: resources.Backend(existingVolume.Volume.Backend)}, nil
}

func (s *spectrumLocalClient) GetVolumeConfig(name string) (volumeConfigDetails map[string]interface{}, err error) {
	s.logger.Println("spectrumLocalClient: GetVolumeConfig start")
	defer s.logger.Println("spectrumLocalClient: GetVolumeConfig finish")

	existingVolume, volExists, err := s.dataModel.GetVolume(name)

	if err != nil {
		s.logger.Println(err.Error())
		return nil, err
	}

	if volExists {
		volumeConfigDetails = make(map[string]interface{})
		volumeMountpoint, err := s.getVolumeMountPoint(existingVolume)
		if err != nil {
			s.logger.Println(err.Error())
			return nil, err
		}

		isFilesetLinked, err := s.connector.IsFilesetLinked(existingVolume.FileSystem, existingVolume.Fileset)
		if err != nil {
			s.logger.Println(err.Error())
			return nil, err
		}
		if isFilesetLinked {
			volumeConfigDetails["mountpoint"] = volumeMountpoint
		}

		volumeConfigDetails[FILESETID] = existingVolume.Fileset
		volumeConfigDetails[FILESYSTEM] = existingVolume.FileSystem
		volumeConfigDetails[CLUSTER] = existingVolume.ClusterId
		if existingVolume.GID != "" {
			volumeConfigDetails[USER_SPECIFIED_GID] = existingVolume.GID
		}
		if existingVolume.UID != "" {
			volumeConfigDetails[USER_SPECIFIED_UID] = existingVolume.UID
		}
		volumeConfigDetails[IS_PREEXISTING] = existingVolume.IsPreexisting
		volumeConfigDetails[TYPE] = existingVolume.Type
		if existingVolume.Type == LIGHTWEIGHT {
			volumeConfigDetails[DIRECTORY] = existingVolume.Directory
		}

		return volumeConfigDetails, nil
	}
	return nil, fmt.Errorf("Volume not found")
}

func (s *spectrumLocalClient) Attach(name string) (volumeMountpoint string, err error) {
	s.logger.Println("spectrumLocalClient: attach start")
	defer s.logger.Println("spectrumLocalClient: attach end")

	existingVolume, volExists, err := s.dataModel.GetVolume(name)

	if err != nil {
		s.logger.Println(err.Error())
		return "", err
	}

	if !volExists {
		return "", fmt.Errorf("Volume not found")
	}

	volumeMountpoint, err = s.getVolumeMountPoint(existingVolume)

	if err != nil {
		s.logger.Println(err.Error())
		return "", err
	}

	isFilesetLinked, err := s.connector.IsFilesetLinked(existingVolume.FileSystem, existingVolume.Fileset)

	if err != nil {
		s.logger.Println(err.Error())
		return "", err
	}

	if isFilesetLinked == false {
		err = s.connector.LinkFileset(existingVolume.FileSystem, existingVolume.Fileset)

		if err != nil {
			s.logger.Println(err.Error())
			return "", err
		}
	}

	return volumeMountpoint, nil
}

func (s *spectrumLocalClient) Detach(name string) (err error) {
	s.logger.Println("spectrumLocalClient: detach start")
	defer s.logger.Println("spectrumLocalClient: detach end")

	existingVolume, volExists, err := s.dataModel.GetVolume(name)

	if err != nil {
		s.logger.Println(err.Error())
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
		s.logger.Println(err.Error())
		return err
	}
	if isFilesetLinked == false {
		return fmt.Errorf("volume not attached")
	}

	return nil
}
func (s *spectrumLocalClient) ListVolumes() ([]resources.VolumeMetadata, error) {
	s.logger.Println("spectrumLocalClient: list start")
	defer s.logger.Println("spectrumLocalClient: list end")
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

func (s *spectrumLocalClient) createFilesetVolume(filesystem, name string, opts map[string]interface{}) error {
	s.logger.Println("spectrumLocalClient: createFilesetVolume start")
	defer s.logger.Println("spectrumLocalClient: createFilesetVolume end")

	filesetName := generateFilesetName(name)

	err := s.connector.CreateFileset(filesystem, filesetName, opts)

	if err != nil {
		s.logger.Printf("Error creating fileset %v", err)
		return err
	}

	err = s.dataModel.InsertFilesetVolume(filesetName, name, filesystem, false, opts)

	if err != nil {
		s.logger.Printf("Error inserting fileset %v", err)
		return err
	}

	s.logger.Printf("Created fileset volume with fileset %s\n", filesetName)
	return nil
}

func (s *spectrumLocalClient) createFilesetQuotaVolume(filesystem, name, quota string, opts map[string]interface{}) error {
	s.logger.Println("spectrumLocalClient: createFilesetQuotaVolume start")
	defer s.logger.Println("spectrumLocalClient: createFilesetQuotaVolume end")

	filesetName := generateFilesetName(name)

	err := s.connector.CreateFileset(filesystem, filesetName, opts)

	if err != nil {
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

	err = s.dataModel.InsertFilesetQuotaVolume(filesetName, quota, name, filesystem, false, opts)

	if err != nil {
		return err
	}

	s.logger.Printf("Created fileset volume with fileset %s, quota %s\n", filesetName, quota)
	return nil
}

func (s *spectrumLocalClient) createLightweightVolume(filesystem, name, fileset string, opts map[string]interface{}) error {
	s.logger.Println("spectrumLocalClient: createLightweightVolume start")
	defer s.logger.Println("spectrumLocalClient: createLightweightVolume end")

	filesetLinked, err := s.connector.IsFilesetLinked(filesystem, fileset)

	if err != nil {
		s.logger.Println("error finding fileset in the filesystem %s", err.Error())
		return err
	}

	if !filesetLinked {
		err = s.connector.LinkFileset(filesystem, fileset)

		if err != nil {
			s.logger.Println(err.Error())
			return err
		}
	}

	lightweightVolumeName := generateLightweightVolumeName(name)

	mountpoint, err := s.connector.GetFilesystemMountpoint(filesystem)
	if err != nil {
		s.logger.Println(err.Error())
		return err
	}
	//open permissions on enclosing fileset
	args := []string{"chmod", "777", path.Join(mountpoint, fileset)}
	_, err = s.executor.Execute("sudo", args)

	if err != nil {
		s.logger.Printf("Failed update permissions of fileset %s containing LTW volumes with error: %s", fileset, err.Error())
		return err
	}

	lightweightVolumePath := path.Join(mountpoint, fileset, lightweightVolumeName)
	args = []string{"mkdir", "-p", lightweightVolumePath}
	_, err = s.executor.Execute("sudo", args)

	if err != nil {
		s.logger.Printf("Failed to create directory path %s : %s", lightweightVolumePath, err.Error())
		return err
	}
	s.logger.Printf("creating directory for lwv: %s\n", lightweightVolumePath)

	err = s.dataModel.InsertLightweightVolume(fileset, lightweightVolumeName, name, filesystem, false, opts)

	if err != nil {
		return err
	}

	s.logger.Printf("Created LightWeight volume at directory path: %s\n", lightweightVolumePath)
	return nil
}

func generateLightweightVolumeName(name string) string {
	return name //TODO: check for convension/valid names
}

func generateFilesetName(name string) string {
	//TODO: placeholder for now
	return name
	//return strconv.FormatInt(time.Now().UnixNano(), 10)
}

//TODO move updates to DB file

func (s *spectrumLocalClient) updateDBWithExistingFileset(filesystem, name, userSpecifiedFileset string, opts map[string]interface{}) error {
	s.logger.Println("spectrumLocalClient:  updateDBWithExistingFileset start")
	defer s.logger.Println("spectrumLocalClient: updateDBWithExistingFileset end")
	s.logger.Printf("User specified fileset: %s\n", userSpecifiedFileset)

	_, err := s.connector.ListFileset(filesystem, userSpecifiedFileset)
	if err != nil {
		s.logger.Printf("Fileset does not exist %v", err.Error())
		return err
	}

	err = s.dataModel.InsertFilesetVolume(userSpecifiedFileset, name, filesystem, true, opts)

	if err != nil {
		s.logger.Println(err.Error())
		return err
	}
	return nil
}

func (s *spectrumLocalClient) checkIfVolumeExistsInDB(name, userSpecifiedFileset string) error {
	s.logger.Println("spectrumLocalClient:  checkIfVolumeExistsIbDB start")
	defer s.logger.Println("spectrumLocalClient: checkIfVolumeExistsIbDB end")

	volumeConfigDetails, err := s.GetVolumeConfig(name)

	if err != nil {
		s.logger.Println(err.Error())
		return err
	}

	if volumeConfigDetails["FilesetId"] != userSpecifiedFileset {
		return fmt.Errorf("volume %s with fileset %s not found", name, userSpecifiedFileset)
	}
	return nil
}

func (s *spectrumLocalClient) updateDBWithExistingFilesetQuota(filesystem, name, userSpecifiedFileset, quota string, opts map[string]interface{}) error {
	s.logger.Println("spectrumLocalClient:  updateDBWithExistingFilesetQuota start")
	defer s.logger.Println("spectrumLocalClient: updateDBWithExistingFilesetQuota end")

	filesetQuota, err := s.connector.ListFilesetQuota(filesystem, userSpecifiedFileset)

	if err != nil {
		s.logger.Println(err.Error())
		return err
	}
	if filesetQuota != quota {
		s.logger.Printf("Mismatch between user-specified and listed quota for fileset %s", userSpecifiedFileset)
		return fmt.Errorf("Mismatch between user-specified and listed quota for fileset %s", userSpecifiedFileset)

	}

	err = s.dataModel.InsertFilesetQuotaVolume(userSpecifiedFileset, quota, name, filesystem, true, opts)

	if err != nil {
		s.logger.Println(err.Error())
		return err
	}
	return nil
}

func (s *spectrumLocalClient) updateDBWithExistingDirectory(filesystem, name, userSpecifiedFileset, userSpecifiedDirectory string, opts map[string]interface{}) error {
	s.logger.Println("spectrumLocalClient:  updateDBWithExistingDirectory start")
	defer s.logger.Println("spectrumLocalClient: updateDBWithExistingDirectory end")
	s.logger.Printf("User specified fileset: %s, User specified directory: %s\n", userSpecifiedFileset, userSpecifiedDirectory)

	linked, err := s.connector.IsFilesetLinked(filesystem, userSpecifiedFileset)
	if err != nil {
		return err
	}
	if linked == false {
		err := s.connector.LinkFileset(filesystem, userSpecifiedFileset)
		if err != nil {
			return err
		}
	}

	mountpoint, err := s.connector.GetFilesystemMountpoint(filesystem)
	if err != nil {
		s.logger.Println(err.Error())
		return err
	}

	directoryPath := path.Join(mountpoint, userSpecifiedFileset, userSpecifiedDirectory)

	_, err = s.executor.Stat(directoryPath)

	if err != nil {
		if os.IsNotExist(err) {
			s.logger.Printf("directory path %s doesn't exist", directoryPath)
			return err
		}

		s.logger.Printf("Error stating directoryPath %s: %s", directoryPath, err.Error())
		return err
	}

	err = s.dataModel.InsertLightweightVolume(userSpecifiedFileset, userSpecifiedDirectory, name, filesystem, true, opts)

	if err != nil {
		s.logger.Println(err.Error())
		return err
	}
	return nil
}

func determineTypeFromRequest(logger *log.Logger, opts map[string]interface{}) (string, error) {
	logger.Print("determineTypeFromRequest start\n")
	defer logger.Printf("determineTypeFromRequest end\n")
	userSpecifiedType, exists := opts[TYPE]
	if exists == false {
		_, exists := opts[DIRECTORY]
		if exists == true {
			return TYPE_LIGHTWEIGHT, nil
		}
		return TYPE_FILESET, nil
	}

	if userSpecifiedType.(string) != TYPE_FILESET && userSpecifiedType.(string) != TYPE_LIGHTWEIGHT {
		return "", fmt.Errorf("Unknown 'type' = %s specified", userSpecifiedType.(string))
	}

	return userSpecifiedType.(string), nil
}
func (s *spectrumLocalClient) validateAndParseParams(logger *log.Logger, opts map[string]interface{}) (bool, string, string, string, error) {
	logger.Println("validateAndParseParams start")
	defer logger.Println("validateAndParseParams end")
	existingFileset, existingFilesetSpecified := opts[TYPE_FILESET]
	existingLightWeightDir, existingLightWeightDirSpecified := opts[DIRECTORY]
	filesystem, filesystemSpecified := opts[FILESYSTEM]
	_, uidSpecified := opts[USER_SPECIFIED_UID]
	_, gidSpecified := opts[USER_SPECIFIED_GID]

	userSpecifiedType, err := determineTypeFromRequest(logger, opts)
	if err != nil {
		logger.Printf("%s", err.Error())
		return false, "", "", "", err
	}

	if uidSpecified && gidSpecified {
		if existingFilesetSpecified && userSpecifiedType != TYPE_LIGHTWEIGHT {
			return true, "", "", "", fmt.Errorf("uid/gid cannot be specified along with existing fileset")
		}
		if existingLightWeightDirSpecified {
			return true, "", "", "", fmt.Errorf("uid/gid cannot be specified along with existing lightweight volume")
		}
	}

	if (userSpecifiedType == TYPE_FILESET && existingFilesetSpecified) || (userSpecifiedType == TYPE_LIGHTWEIGHT && existingLightWeightDirSpecified) {
		if filesystemSpecified == false {
			logger.Println("'filesystem' is a required opt for using existing volumes")
			return true, filesystem.(string), existingFileset.(string), existingLightWeightDir.(string), fmt.Errorf("'filesystem' is a required opt for using existing volumes")
		}
		if existingLightWeightDirSpecified && !existingFilesetSpecified {
			logger.Println("'fileset' is a required opt for using existing lightweight volumes")
			return true, filesystem.(string), existingFileset.(string), existingLightWeightDir.(string), fmt.Errorf("'fileset' is a required opt for using existing lightweight volumes")
		}
		if userSpecifiedType == TYPE_LIGHTWEIGHT && existingLightWeightDir != nil {
			_, quotaSpecified := opts[QUOTA]
			if quotaSpecified {
				logger.Println("'quota' is not supported for lightweight volumes")
				return true, "", "", "", fmt.Errorf("'quota' is not supported for lightweight volumes")
			}
			logger.Println("Valid: existing LTWT")
			return true, filesystem.(string), existingFileset.(string), existingLightWeightDir.(string), nil
		} else {
			logger.Println("Valid: existing FILESET")
			return true, filesystem.(string), existingFileset.(string), "", nil
		}

	} else if userSpecifiedType == TYPE_LIGHTWEIGHT {
		//lightweight -- new
		if filesystemSpecified && existingFilesetSpecified {

			_, quotaSpecified := opts[QUOTA]
			if quotaSpecified {
				logger.Println("'quota' is not supported for lightweight volumes")
				return false, "", "", "", fmt.Errorf("'quota' is not supported for lightweight volumes")
			}

			return false, filesystem.(string), existingFileset.(string), "", nil
		}
		return false, "", "", "", fmt.Errorf("'filesystem' and 'fileset' are required opts for using lightweight volumes")
	} else if filesystemSpecified == false {
		return false, s.config.DefaultFilesystemName, "", "", nil

	} else {
		return false, filesystem.(string), "", "", nil
	}

}

func (s *spectrumLocalClient) getVolumeMountPoint(volume SpectrumScaleVolume) (string, error) {
	s.logger.Println("getVolumeMountPoint start")
	defer s.logger.Println("getVolumeMountPoint end")

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

	if volume.Type == LIGHTWEIGHT {
		return path.Join(fsMountpoint, volume.Fileset, volume.Directory), nil
	}

	return path.Join(fsMountpoint, volume.Fileset), nil

}
func (s *spectrumLocalClient) updatePermissions(name string) error {
	s.logger.Println("spectrumLocalClient: updatePermissions-start")
	defer s.logger.Println("spectrumLocalClient: updatePermissions-end")
	volumeConfig, err := s.GetVolumeConfig(name)
	if err != nil {
		return err
	}
	filesystem, exists := volumeConfig[FILESYSTEM]
	if exists == false {
		return fmt.Errorf("Cannot determine filesystem for volume: %s", name)
	}
	fsMountpoint, err := s.connector.GetFilesystemMountpoint(filesystem.(string))
	if err != nil {
		return err
	}
	volumeType, exists := volumeConfig[TYPE]
	if exists == false {
		return fmt.Errorf("Cannot determine type for volume: %s", name)
	}
	fileset, exists := volumeConfig[FILESETID]
	if exists == false {
		return fmt.Errorf("Cannot determine filesetId for volume: %s", name)
	}

	filesetPath := path.Join(fsMountpoint, fileset.(string))
	//chmod 777 mountpoint
	args := []string{"chmod", "777", filesetPath}
	_, err = s.executor.Execute("sudo", args)
	if err != nil {
		s.logger.Printf("Failed to change permissions of filesetpath %s: %s", filesetPath, err.Error())
		return err
	}
	if volumeType == LIGHTWEIGHT {
		directory, exists := volumeConfig[DIRECTORY]
		if exists == false {
			return fmt.Errorf("Cannot determine directory for volume: %s", name)
		}
		directoryPath := path.Join(filesetPath, directory.(string))
		args := []string{"chmod", "777", directoryPath}
		_, err = s.executor.Execute("sudo", args)
		if err != nil {
			s.logger.Printf("Failed to change permissions of directorypath %s: %s", directoryPath, err.Error())
			return err
		}
	}
	return nil
}
