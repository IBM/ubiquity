package spectrumscale

import (
	"log"

	"github.ibm.com/almaden-containers/ubiquity/utils"

	"os"
	"path"

	"fmt"

	"github.com/jinzhu/gorm"
	"github.ibm.com/almaden-containers/ubiquity/local/spectrumscale/connectors"
	"github.ibm.com/almaden-containers/ubiquity/resources"
)

type spectrumLocalClient struct {
	logger      *log.Logger
	connector   connectors.SpectrumScaleConnector
	dataModel   SpectrumDataModel
	fileLock    utils.FileLock
	executor    utils.Executor
	isActivated bool
	isMounted   bool
	config      resources.SpectrumScaleConfig
}

const (
	USER_SPECIFIED_TYPE string = "type"

	USER_SPECIFIED_DIRECTORY string = "directory"
	USER_SPECIFIED_QUOTA     string = "quota"

	USER_SPECIFIED_FILESET    string = "fileset"
	USER_SPECIFIED_FILESYSTEM string = "filesystem"

	FILESET_TYPE  string = "fileset"
	LTWT_VOL_TYPE string = "lightweight"
)

func NewSpectrumLocalClient(logger *log.Logger, config resources.SpectrumScaleConfig, database *gorm.DB, fileLock utils.FileLock) (resources.StorageClient, error) {
	if config.ConfigPath == "" {
		return nil, fmt.Errorf("spectrumLocalClient: init: missing required parameter 'spectrumConfigPath'")
	}
	if config.DefaultFilesystem == "" {
		return nil, fmt.Errorf("spectrumLocalClient: init: missing required parameter 'spectrumDefaultFileSystem'")
	}
	return newSpectrumLocalClient(logger, config, database, fileLock, resources.SPECTRUM_SCALE)
}

func NewSpectrumLocalClientWithConnectors(logger *log.Logger, connector connectors.SpectrumScaleConnector, fileLock utils.FileLock, spectrumExecutor utils.Executor, config resources.SpectrumScaleConfig, datamodel SpectrumDataModel) (resources.StorageClient, error) {
	return &spectrumLocalClient{logger: logger, connector: connector, dataModel: datamodel, executor: spectrumExecutor, config: config, fileLock: fileLock}, nil
}

func newSpectrumLocalClient(logger *log.Logger, config resources.SpectrumScaleConfig, database *gorm.DB, fileLock utils.FileLock, backend resources.Backend) (*spectrumLocalClient, error) {
	logger.Println("spectrumLocalClient: init start")
	defer logger.Println("spectrumLocalClient: init end")

	client, err := connectors.GetSpectrumScaleConnector(logger, config)
	if err != nil {
		logger.Fatalln(err.Error())
		return &spectrumLocalClient{}, err
	}
	return &spectrumLocalClient{logger: logger, connector: client, dataModel: NewSpectrumDataModel(logger, database, backend), config: config, fileLock: fileLock, executor: utils.NewExecutor(logger)}, nil
}

func (s *spectrumLocalClient) Activate() (err error) {
	s.logger.Println("spectrumLocalClient: Activate start")
	defer s.logger.Println("spectrumLocalClient: Activate end")

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

	//check if filesystem is mounted

	mounted, err := s.connector.IsFilesystemMounted(s.config.DefaultFilesystem)

	if err != nil {
		s.logger.Println(err.Error())
		return err
	}

	if mounted == false {
		err = s.connector.MountFileSystem(s.config.DefaultFilesystem)

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

	err = s.dataModel.CreateVolumeTable()

	if err != nil {
		s.logger.Println(err.Error())
		return err
	}

	s.isActivated = true
	return nil
}

func (s *spectrumLocalClient) CreateVolume(name string, opts map[string]interface{}) (err error) {
	s.logger.Println("spectrumLocalClient: create start")
	defer s.logger.Println("spectrumLocalClient: create end")

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

	if len(opts) == 0 {
		//fileset
		return s.createFilesetVolume(s.config.DefaultFilesystem, name, opts)
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

	if isExistingVolume && userSpecifiedType == FILESET_TYPE {
		quota, quotaSpecified := opts[USER_SPECIFIED_QUOTA]
		if quotaSpecified {
			return s.updateDBWithExistingFilesetQuota(filesystem, name, existingFileset, quota.(string), opts)
		}
		return s.updateDBWithExistingFileset(filesystem, name, existingFileset, opts)
	}

	if isExistingVolume && userSpecifiedType == LTWT_VOL_TYPE {
		return s.updateDBWithExistingDirectory(filesystem, name, existingFileset, existingLightWeightDir, opts)
	}

	if userSpecifiedType == FILESET_TYPE {
		quota, quotaSpecified := opts[USER_SPECIFIED_QUOTA]
		if quotaSpecified {
			return s.createFilesetQuotaVolume(filesystem, name, quota.(string), opts)
		}
		return s.createFilesetVolume(filesystem, name, opts)
	}
	if userSpecifiedType == LTWT_VOL_TYPE {
		return s.createLightweightVolume(filesystem, name, existingFileset, opts)
	}
	return fmt.Errorf("Internal error")
}

func (s *spectrumLocalClient) RemoveVolume(name string, forceDelete bool) (err error) {
	s.logger.Println("spectrumLocalClient: remove start")
	defer s.logger.Println("spectrumLocalClient: remove end")

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
		if forceDelete == true {
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
	if forceDelete {
		err = s.connector.DeleteFileset(existingVolume.FileSystem, existingVolume.Fileset)

		if err != nil {
			s.logger.Println(err.Error())
			return err
		}
	}

	return nil
}

func (s *spectrumLocalClient) GetVolume(name string) (volumeMetadata resources.VolumeMetadata, volumeConfigDetails map[string]interface{}, err error) {
	s.logger.Println("spectrumLocalClient: get start")
	defer s.logger.Println("spectrumLocalClient: get finish")

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

		volumeMountpoint, isLinked, err := s.getVolumeMountPoint(existingVolume)
		if err != nil {
			s.logger.Println(err.Error())
			return resources.VolumeMetadata{}, nil, err
		}
		volumeMetadata = resources.VolumeMetadata{Name: existingVolume.Volume.Name}
		if isLinked {
			volumeMetadata.Mountpoint = volumeMountpoint
		}
		volumeConfigDetails = make(map[string]interface{})
		volumeConfigDetails["filesetId"] = existingVolume.Fileset
		volumeConfigDetails["filesystem"] = existingVolume.FileSystem
		volumeConfigDetails["clusterId"] = existingVolume.ClusterId
		if existingVolume.GID != "" {
			volumeConfigDetails["gid"] = existingVolume.GID
		}
		if existingVolume.UID != "" {
			volumeConfigDetails["uid"] = existingVolume.UID
		}
		volumeConfigDetails["isPreexisting"] = existingVolume.IsPreexisting

		return volumeMetadata, volumeConfigDetails, nil
	}
	return resources.VolumeMetadata{}, nil, fmt.Errorf("Volume not found")
}

func (s *spectrumLocalClient) Attach(name string) (volumeMountpoint string, err error) {
	s.logger.Println("spectrumLocalClient: attach start")
	defer s.logger.Println("spectrumLocalClient: attach end")

	err = s.fileLock.Lock()
	if err != nil {
		s.logger.Printf("failed to aquire lock %v", err)
		return "", err
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
		return "", err
	}

	if !volExists {
		return "", fmt.Errorf("Volume not found")
	}

	volumeMountpoint, linked, err := s.getVolumeMountPoint(existingVolume)

	if err != nil {
		s.logger.Println(err.Error())
		return "", err
	}
	if linked == false {
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

	err = s.fileLock.Lock()
	if err != nil {
		s.logger.Printf("error aquiring the lock %v", err)
		return err
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
		return err
	}

	if !volExists {
		return fmt.Errorf("Volume not found")
	}

	_, linked, err := s.getVolumeMountPoint(existingVolume)
	if err != nil {
		return err
	}
	if linked == false {
		return fmt.Errorf("volume not attached")
	}

	return nil
}
func (s *spectrumLocalClient) ListVolumes() ([]resources.VolumeMetadata, error) {
	s.logger.Println("spectrumLocalClient: list start")
	defer s.logger.Println("spectrumLocalClient: list end")
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

		volumeMountpoint, _, err := s.getVolumeMountPoint(volume)
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

//func (s *spectrumLocalClient) changePermissionsOfFileset(filesystem, filesetName, uid, gid string) error {
//	s.logger.Println("spectrumLocalClient: changeOwnerOfFileset start")
//	defer s.logger.Println("spectrumLocalClient: changeOwnerOfFileset end")
//
//	s.logger.Printf("Changing Owner of Fileset %s to uid %s , gid %s", filesetName, uid, gid)
//
//	mountpoint, err := s.connector.GetFilesystemMountpoint(filesystem)
//	if err != nil {
//		s.logger.Printf("Failed to change permissions of fileset %s : %s", filesetName, err.Error())
//		return err
//	}
//
//	filesetPath := path.Join(mountpoint, filesetName)
//	args := []string{"chown", "-R", fmt.Sprintf("%s:%s", uid, gid), filesetPath}
//	_, err = s.executor.Execute("sudo", args)
//	if err != nil {
//		s.logger.Printf("Failed to change permissions of fileset %s: %s", filesetName, err.Error())
//		return err
//	}
//	return nil
//}

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

	lightweightVolumePath := path.Join(mountpoint, fileset, lightweightVolumeName)

	//	err = s.executor.Mkdir(lightweightVolumePath, 0755)
	args := []string{"mkdir", "-p", lightweightVolumePath}
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

	_, volumeConfigDetails, err := s.GetVolume(name)

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
	userSpecifiedType, exists := opts[USER_SPECIFIED_TYPE]
	if exists == false {
		_, exists := opts[USER_SPECIFIED_DIRECTORY]
		if exists == true {
			return LTWT_VOL_TYPE, nil
		}
		return FILESET_TYPE, nil
	}

	if userSpecifiedType.(string) != FILESET_TYPE && userSpecifiedType.(string) != LTWT_VOL_TYPE {
		return "", fmt.Errorf("Unknown 'type' = %s specified", userSpecifiedType.(string))
	}

	return userSpecifiedType.(string), nil
}
func (s *spectrumLocalClient) validateAndParseParams(logger *log.Logger, opts map[string]interface{}) (bool, string, string, string, error) {
	logger.Println("validateAndParseParams start")
	defer logger.Println("validateAndParseParams end")
	existingFileset, existingFilesetSpecified := opts[USER_SPECIFIED_FILESET]
	existingLightWeightDir, existingLightWeightDirSpecified := opts[USER_SPECIFIED_DIRECTORY]
	filesystem, filesystemSpecified := opts[USER_SPECIFIED_FILESYSTEM]
	_, uidSpecified := opts[USER_SPECIFIED_UID]
	_, gidSpecified := opts[USER_SPECIFIED_GID]

	userSpecifiedType, err := determineTypeFromRequest(logger, opts)
	if err != nil {
		logger.Printf("%s", err.Error())
		return false, "", "", "", err
	}

	if uidSpecified && gidSpecified {
		if existingFilesetSpecified && userSpecifiedType != LTWT_VOL_TYPE {
			return true, "", "", "", fmt.Errorf("uid/gid cannot be specified along with existing fileset")
		}
		if existingLightWeightDirSpecified {
			return true, "", "", "", fmt.Errorf("uid/gid cannot be specified along with existing lightweight volume")
		}
	}

	if (userSpecifiedType == FILESET_TYPE && existingFilesetSpecified) || (userSpecifiedType == LTWT_VOL_TYPE && existingLightWeightDirSpecified) {
		if filesystemSpecified == false {
			logger.Println("'filesystem' is a required opt for using existing volumes")
			return true, filesystem.(string), existingFileset.(string), existingLightWeightDir.(string), fmt.Errorf("'filesystem' is a required opt for using existing volumes")
		}
		if existingLightWeightDirSpecified && !existingFilesetSpecified {
			logger.Println("'fileset' is a required opt for using existing lightweight volumes")
			return true, filesystem.(string), existingFileset.(string), existingLightWeightDir.(string), fmt.Errorf("'fileset' is a required opt for using existing lightweight volumes")
		}
		if userSpecifiedType == LTWT_VOL_TYPE && existingLightWeightDir != nil {
			_, quotaSpecified := opts[USER_SPECIFIED_QUOTA]
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

	} else if userSpecifiedType == LTWT_VOL_TYPE {
		//lightweight -- new
		if filesystemSpecified && existingFilesetSpecified {

			_, quotaSpecified := opts[USER_SPECIFIED_QUOTA]
			if quotaSpecified {
				logger.Println("'quota' is not supported for lightweight volumes")
				return false, "", "", "", fmt.Errorf("'quota' is not supported for lightweight volumes")
			}

			return false, filesystem.(string), existingFileset.(string), "", nil
		}
		return false, "", "", "", fmt.Errorf("'filesystem' and 'fileset' are required opts for using lightweight volumes")
	} else if filesystemSpecified == false {
		return false, s.config.DefaultFilesystem, "", "", nil

	} else {
		return false, filesystem.(string), "", "", nil
	}

}

func (s *spectrumLocalClient) getVolumeMountPoint(volume SpectrumScaleVolume) (string, bool, error) {
	s.logger.Println("getVolumeMountPoint start")
	defer s.logger.Println("getVolumeMountPoint end")

	fsMountpoint, err := s.connector.GetFilesystemMountpoint(volume.FileSystem)
	if err != nil {
		return "", false, err
	}

	isFilesetLinked, err := s.connector.IsFilesetLinked(volume.FileSystem, volume.Fileset)

	if err != nil {
		s.logger.Println(err.Error())
		return "", false, err
	}

	if volume.Type == LIGHTWEIGHT {
		return path.Join(fsMountpoint, volume.Fileset, volume.Directory), isFilesetLinked, nil
	}

	return path.Join(fsMountpoint, volume.Fileset), isFilesetLinked, nil

}
