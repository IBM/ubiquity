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
	Type            string = "type"
	TypeFileset     string = "fileset"
	TypeLightweight string = "lightweight"

	FilesetID string = "fileset"
	Directory string = "directory"
	Quota     string = "quota"

	Filesystem string = "filesystem"

	IsPreexisting string = "isPreexisting"

	Cluster string = "clusterId"
)

func NewSpectrumLocalClient(logger *log.Logger, config resources.UbiquityServerConfig, database *gorm.DB) (resources.StorageClient, error) {
	if config.ConfigPath == "" {
		return nil, fmt.Errorf("spectrumLocalClient: init: missing required parameter 'spectrumConfigPath'")
	}
	if config.SpectrumScaleConfig.DefaultFilesystemName == "" {
		return nil, fmt.Errorf("spectrumLocalClient: init: missing required parameter 'spectrumDefaultFileSystem'")
	}
	return newSpectrumLocalClient(logger, config.SpectrumScaleConfig, database, resources.SpectrumScale)
}

func NewSpectrumLocalClientWithConnectors(logger *log.Logger, connector connectors.SpectrumScaleConnector, spectrumExecutor utils.Executor, config resources.SpectrumScaleConfig, datamodel SpectrumDataModel) (resources.StorageClient, error) {
	err := datamodel.CreateVolumeTable()
	if err != nil {
		return &spectrumLocalClient{}, err
	}
	return &spectrumLocalClient{logger: logger, connector: connector, dataModel: datamodel, executor: spectrumExecutor, config: config, activationLock: &sync.RWMutex{}}, nil
}

func newSpectrumLocalClient(logger *log.Logger, config resources.SpectrumScaleConfig, database *gorm.DB, backend string) (*spectrumLocalClient, error) {
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
	return &spectrumLocalClient{logger: logger, connector: client, dataModel: datamodel, config: config, executor: utils.NewExecutor(), activationLock: &sync.RWMutex{}}, nil
}

func (s *spectrumLocalClient) Activate(activateRequest resources.ActivateRequest) (err error) {
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

func (s *spectrumLocalClient) CreateVolume(createVolumeRequest resources.CreateVolumeRequest) (err error) {
	s.logger.Println("spectrumLocalClient: create start")
	defer s.logger.Println("spectrumLocalClient: create end")

	_, volExists, err := s.dataModel.GetVolume(createVolumeRequest.Name)

	if err != nil {
		s.logger.Println(err.Error())
		return err
	}

	if volExists {
		return fmt.Errorf("Volume already exists")
	}

	s.logger.Printf("Opts for create: %#v\n", createVolumeRequest.Opts)

	if len(createVolumeRequest.Opts) == 0 {
		//fileset
		return s.createFilesetVolume(s.config.DefaultFilesystemName, createVolumeRequest.Name, createVolumeRequest.Opts)
	}
	s.logger.Printf("Trying to determine type for request\n")
	userSpecifiedType, err := determineTypeFromRequest(s.logger, createVolumeRequest.Opts)
	if err != nil {
		s.logger.Printf("Error determining type: %s\n", err.Error())
		return err
	}
	s.logger.Printf("Volume type requested: %s", userSpecifiedType)
	isExistingVolume, filesystem, existingFileset, existingLightWeightDir, err := s.validateAndParseParams(s.logger, createVolumeRequest.Opts)
	if err != nil {
		s.logger.Printf("Error in validate params: %s\n", err.Error())
		return err
	}

	s.logger.Printf("Params for create: %s,%s,%s,%s\n", isExistingVolume, filesystem, existingFileset, existingLightWeightDir)

	if isExistingVolume && userSpecifiedType == TypeFileset {
		quota, quotaSpecified := createVolumeRequest.Opts[Quota]
		if quotaSpecified {
			return s.updateDBWithExistingFilesetQuota(filesystem, createVolumeRequest.Name, existingFileset, quota.(string), createVolumeRequest.Opts)
		}
		return s.updateDBWithExistingFileset(filesystem, createVolumeRequest.Name, existingFileset, createVolumeRequest.Opts)
	}

	if isExistingVolume && userSpecifiedType == TypeLightweight {
		return s.updateDBWithExistingDirectory(filesystem, createVolumeRequest.Name, existingFileset, existingLightWeightDir, createVolumeRequest.Opts)
	}

	if userSpecifiedType == TypeFileset {
		quota, quotaSpecified := createVolumeRequest.Opts[Quota]
		if quotaSpecified {
			return s.createFilesetQuotaVolume(filesystem, createVolumeRequest.Name, quota.(string), createVolumeRequest.Opts)
		}
		return s.createFilesetVolume(filesystem, createVolumeRequest.Name, createVolumeRequest.Opts)
	}
	if userSpecifiedType == TypeLightweight {
		return s.createLightweightVolume(filesystem, createVolumeRequest.Name, existingFileset, createVolumeRequest.Opts)
	}
	return fmt.Errorf("Internal error")
}

func (s *spectrumLocalClient) RemoveVolume(removeVolumeRequest resources.RemoveVolumeRequest) (err error) {
	s.logger.Println("spectrumLocalClient: remove start")
	defer s.logger.Println("spectrumLocalClient: remove end")

	existingVolume, volExists, err := s.dataModel.GetVolume(removeVolumeRequest.Name)

	if err != nil {
		s.logger.Println(err.Error())
		return err
	}

	if volExists == false {
		return fmt.Errorf("Volume not found")
	}

	if existingVolume.Type == Lightweight {
		err = s.dataModel.DeleteVolume(removeVolumeRequest.Name)
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
	err = s.dataModel.DeleteVolume(removeVolumeRequest.Name)

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

func (s *spectrumLocalClient) GetVolume(getVolumeRequest resources.GetVolumeRequest) (resources.Volume, error) {
	s.logger.Println("spectrumLocalClient: GetVolume start")
	defer s.logger.Println("spectrumLocalClient: GetVolume finish")

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
	s.logger.Println("spectrumLocalClient: GetVolumeConfig start")
	defer s.logger.Println("spectrumLocalClient: GetVolumeConfig finish")

	existingVolume, volExists, err := s.dataModel.GetVolume(getVolumeConfigRequest.Name)

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
		if existingVolume.Type == Lightweight {
			volumeConfigDetails[Directory] = existingVolume.Directory
		}

		return volumeConfigDetails, nil
	}
	return nil, fmt.Errorf("Volume not found")
}

func (s *spectrumLocalClient) Attach(attachRequest resources.AttachRequest) (volumeMountpoint string, err error) {
	s.logger.Println("spectrumLocalClient: attach start")
	defer s.logger.Println("spectrumLocalClient: attach end")

	existingVolume, volExists, err := s.dataModel.GetVolume(attachRequest.Name)

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

	existingVolume.Volume.Mountpoint = volumeMountpoint

	err = s.dataModel.UpdateVolumeMountpoint(attachRequest.Name, volumeMountpoint)

	if err != nil {
		s.logger.Println(err.Error())
		return "", err
	}

	return volumeMountpoint, nil
}

func (s *spectrumLocalClient) Detach(detachRequest resources.DetachRequest) (err error) {
	s.logger.Println("spectrumLocalClient: detach start")
	defer s.logger.Println("spectrumLocalClient: detach end")

	existingVolume, volExists, err := s.dataModel.GetVolume(detachRequest.Name)

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

	err = s.dataModel.UpdateVolumeMountpoint(detachRequest.Name, "")
	if err != nil {
		s.logger.Println(err.Error())
		return err
	}

	return nil
}

func (s *spectrumLocalClient) ListVolumes(listVolumesRequest resources.ListVolumesRequest) ([]resources.Volume, error) {
	s.logger.Println("spectrumLocalClient: list start")
	defer s.logger.Println("spectrumLocalClient: list end")
	var err error

	volumesInDb, err := s.dataModel.ListVolumes()

	if err != nil {
		s.logger.Printf("error retrieving volumes from db %#v\n", err)
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
	getVolumeConfigRequest := resources.GetVolumeConfigRequest{Name: name}
	volumeConfigDetails, err := s.GetVolumeConfig(getVolumeConfigRequest)

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
	userSpecifiedType, exists := opts[Type]
	if exists == false {
		_, exists := opts[Directory]
		if exists == true {
			return TypeLightweight, nil
		}
		return TypeFileset, nil
	}

	if userSpecifiedType.(string) != TypeFileset && userSpecifiedType.(string) != TypeLightweight {
		return "", fmt.Errorf("Unknown 'type' = %s specified", userSpecifiedType.(string))
	}

	return userSpecifiedType.(string), nil
}

func (s *spectrumLocalClient) validateAndParseParams(logger *log.Logger, opts map[string]interface{}) (bool, string, string, string, error) {
	logger.Println("validateAndParseParams start")
	defer logger.Println("validateAndParseParams end")
	existingFileset, existingFilesetSpecified := opts[TypeFileset]
	existingLightWeightDir, existingLightWeightDirSpecified := opts[Directory]
	filesystem, filesystemSpecified := opts[Filesystem]
	_, uidSpecified := opts[UserSpecifiedUID]
	_, gidSpecified := opts[UserSpecifiedGID]

	userSpecifiedType, err := determineTypeFromRequest(logger, opts)
	if err != nil {
		logger.Printf("%s", err.Error())
		return false, "", "", "", err
	}

	if uidSpecified && gidSpecified {
		if existingFilesetSpecified && userSpecifiedType != TypeLightweight {
			return true, "", "", "", fmt.Errorf("uid/gid cannot be specified along with existing fileset")
		}
		if existingLightWeightDirSpecified {
			return true, "", "", "", fmt.Errorf("uid/gid cannot be specified along with existing lightweight volume")
		}
	}

	if (userSpecifiedType == TypeFileset && existingFilesetSpecified) || (userSpecifiedType == TypeLightweight && existingLightWeightDirSpecified) {
		if filesystemSpecified == false {
			logger.Println("'filesystem' is a required opt for using existing volumes")
			return true, filesystem.(string), existingFileset.(string), existingLightWeightDir.(string), fmt.Errorf("'filesystem' is a required opt for using existing volumes")
		}
		if existingLightWeightDirSpecified && !existingFilesetSpecified {
			logger.Println("'fileset' is a required opt for using existing lightweight volumes")
			return true, filesystem.(string), existingFileset.(string), existingLightWeightDir.(string), fmt.Errorf("'fileset' is a required opt for using existing lightweight volumes")
		}
		if userSpecifiedType == TypeLightweight && existingLightWeightDir != nil {
			_, quotaSpecified := opts[Quota]
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

	} else if userSpecifiedType == TypeLightweight {
		//lightweight -- new
		if filesystemSpecified && existingFilesetSpecified {

			_, quotaSpecified := opts[Quota]
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

	if volume.Type == Lightweight {
		return path.Join(fsMountpoint, volume.Fileset, volume.Directory), nil
	}

	return path.Join(fsMountpoint, volume.Fileset), nil

}
func (s *spectrumLocalClient) updatePermissions(name string) error {
	s.logger.Println("spectrumLocalClient: updatePermissions-start")
	defer s.logger.Println("spectrumLocalClient: updatePermissions-end")
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
	args := []string{"chmod", "777", filesetPath}
	_, err = s.executor.Execute("sudo", args)
	if err != nil {
		s.logger.Printf("Failed to change permissions of filesetpath %s: %s", filesetPath, err.Error())
		return err
	}
	if volumeType == Lightweight {
		directory, exists := volumeConfig[Directory]
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
