package local

import (
	"fmt"
	"log"

	"github.ibm.com/almaden-containers/ubiquity/model"
	"github.ibm.com/almaden-containers/ubiquity/spectrum"
	"github.ibm.com/almaden-containers/ubiquity/utils"

	"os"
	"os/exec"
	"os/signal"
	"os/user"
	"path"
	"syscall"
)

type spectrumLocalClient struct {
	logger      *log.Logger
	client      spectrum.Spectrum
	dbClient    *utils.DatabaseClient
	filelock    *utils.FileLock
	isActivated bool
	isMounted   bool
	config      model.SpectrumConfig
}

const (
	USER_SPECIFIED_TYPE string = "type"

	USER_SPECIFIED_DIRECTORY string = "directory"
	USER_SPECIFIED_QUOTA     string = "quota"

	USER_SPECIFIED_FILESET    string = "fileset"
	USER_SPECIFIED_FILESYSTEM string = "filesystem"

	USER_SPECIFIED_UID string = "uid"
	USER_SPECIFIED_GID string = "gid"

	FILESET_TYPE  string = "fileset"
	LTWT_VOL_TYPE string = "lightweight"
)

func NewSpectrumLocalClient(logger *log.Logger, config model.SpectrumConfig) (model.StorageClient, error) {
	if config.ConfigPath == "" {
		return nil, fmt.Errorf("spectrumLocalClient: init: missing required parameter 'spectrumConfigPath'")
	}
	if config.DefaultFilesystem == "" {
		return nil, fmt.Errorf("spectrumLocalClient: init: missing required parameter 'spectrumDefaultFileSystem'")
	}
	return newSpectrumLocalClient(logger, config)
}

func newSpectrumLocalClient(logger *log.Logger, config model.SpectrumConfig) (*spectrumLocalClient, error) {
	logger.Println("spectrumLocalClient: init start")
	defer logger.Println("spectrumLocalClient: init end")

	ubiquityConfigPath, err := setupConfigDirectory(logger, config.ConfigPath)
	if err != nil {
		return &spectrumLocalClient{}, err
	}

	dbClient := utils.NewDatabaseClient(logger, ubiquityConfigPath)
	err = dbClient.Init()
	if err != nil {
		logger.Fatalln(err.Error())
		return &spectrumLocalClient{}, err
	}

	// Catch Ctrl-C / interrupts to perform DB connection cleanup
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	go func() {
		<-c
		dbClient.Close()
		os.Exit(1)
	}()

	client, err := spectrum.GetSpectrumClient(logger, config.Connector, map[string]interface{}{})
	return &spectrumLocalClient{logger: logger, client: client, dbClient: dbClient,
		filelock: utils.NewFileLock(logger, ubiquityConfigPath), config: config}, nil
}

func setupConfigDirectory(logger *log.Logger, configPath string) (string, error) {
	logger.Println("setupConfigPath start")
	defer logger.Println("setupConfigPath end")
	ubiquityConfigPath := path.Join(configPath, ".config")
	log.Printf("User specified config path: %s", configPath)

	if _, err := os.Stat(ubiquityConfigPath); os.IsNotExist(err) {
		args := []string{"mkdir", ubiquityConfigPath}
		cmd := exec.Command("sudo", args...)
		_, err := cmd.Output()
		if err != nil {
			logger.Printf("Error creating config directory %s", ubiquityConfigPath)
			return "", err
		}
	}
	currentUser, err := user.Current()
	if err != nil {
		logger.Printf("Error determining current user: %s", err.Error())
		return "", err
	}

	args := []string{"chown", "-R", fmt.Sprintf("%s:%s", currentUser.Uid, currentUser.Gid), ubiquityConfigPath}
	cmd := exec.Command("sudo", args...)
	_, err = cmd.Output()
	if err != nil {
		logger.Printf("Error setting permissions on config directory %s", ubiquityConfigPath)
		return "", err
	}

	return ubiquityConfigPath, nil
}

func (s *spectrumLocalClient) Activate() (err error) {
	s.logger.Println("spectrumLocalClient: Activate start")
	defer s.logger.Println("spectrumLocalClient: Activate end")

	s.filelock.Lock()
	defer func() {
		lockErr := s.filelock.Unlock()
		if lockErr != nil && err == nil {
			err = lockErr
		}
	}()

	if s.isActivated {
		return nil
	}

	//check if filesystem is mounted

	mounted, err := s.client.IsFilesystemMounted(s.config.DefaultFilesystem)

	if err != nil {
		s.logger.Println(err.Error())
		return err
	}

	if mounted == false {
		err = s.client.MountFileSystem(s.config.DefaultFilesystem)

		if err != nil {
			s.logger.Println(err.Error())
			return err
		}
	}

	clusterId, err := s.client.GetClusterId()

	if err != nil {
		s.logger.Println(err.Error())
		return err
	}

	if len(clusterId) == 0 {
		clusterIdErr := fmt.Errorf("Unable to retrieve clusterId: clusterId is empty")
		s.logger.Println(clusterIdErr.Error())
		return clusterIdErr
	}

	s.dbClient.ClusterId = clusterId

	err = s.dbClient.CreateVolumeTable()

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

	s.filelock.Lock()
	defer func() {
		lockErr := s.filelock.Unlock()
		if lockErr != nil && err == nil {
			err = lockErr
		}
	}()

	volExists, err := s.dbClient.VolumeExists(name)

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
		s.logger.Printf("Error invalidate params: %s\n", err.Error())
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

	s.filelock.Lock()
	defer func() {
		lockErr := s.filelock.Unlock()
		if lockErr != nil && err == nil {
			err = lockErr
		}
	}()

	volExists, err := s.dbClient.VolumeExists(name)

	if err != nil {
		s.logger.Println(err.Error())
		return err
	}

	if volExists == false {
		return fmt.Errorf("Volume not found")
	}

	existingVolume, err := s.dbClient.GetVolume(name)

	if err != nil {
		s.logger.Println(err.Error())
		return err
	}
	if existingVolume.Type == utils.LIGHTWEIGHT {
		err = s.dbClient.DeleteVolume(name)

		if err != nil {
			s.logger.Println(err.Error())
			return err
		}
		if forceDelete == true {
			mountpoint, err := s.client.GetFilesystemMountpoint(existingVolume.FileSystem)
			if err != nil {
				s.logger.Println(err.Error())
				return err
			}
			lightweightVolumePath := path.Join(mountpoint, existingVolume.Fileset, existingVolume.Directory)

			err = os.RemoveAll(lightweightVolumePath)

			if err != nil {
				s.logger.Println(err.Error())
				return err
			}
		}
		return nil
	}

	isFilesetLinked, err := s.client.IsFilesetLinked(existingVolume.FileSystem, existingVolume.Fileset)

	if err != nil {
		s.logger.Println(err.Error())
		return err
	}

	if isFilesetLinked {
		err := s.client.UnlinkFileset(existingVolume.FileSystem, existingVolume.Fileset)

		if err != nil {
			s.logger.Println(err.Error())
			return err
		}
	}
	err = s.dbClient.DeleteVolume(name)

	if err != nil {
		s.logger.Println(err.Error())
		return err
	}
	if forceDelete {
		err = s.client.DeleteFileset(existingVolume.FileSystem, existingVolume.Fileset)

		if err != nil {
			s.logger.Println(err.Error())
			return err
		}
	}

	return nil
}

//GetVolume(string) (*model.VolumeMetadata, *string, *map[string]interface {}, error)
func (s *spectrumLocalClient) GetVolume(name string) (volumeMetadata model.VolumeMetadata, volumeConfigDetails map[string]interface{}, err error) {
	s.logger.Println("spectrumLocalClient: get start")
	defer s.logger.Println("spectrumLocalClient: get finish")

	s.filelock.Lock()
	defer func() {
		lockErr := s.filelock.Unlock()
		if lockErr != nil && err == nil {
			err = lockErr
		}
	}()

	volExists, err := s.dbClient.VolumeExists(name)

	if err != nil {
		s.logger.Println(err.Error())
		return model.VolumeMetadata{}, nil, err
	}

	if volExists {

		existingVolume, err := s.dbClient.GetVolume(name)

		if err != nil {
			s.logger.Println(err.Error())
			return model.VolumeMetadata{}, nil, err
		}

		volumeMetadata = model.VolumeMetadata{Name: existingVolume.Name, Mountpoint: existingVolume.Mountpoint}
		volumeConfigDetails = map[string]interface{}{"FilesetId": existingVolume.Fileset, "Filesystem": existingVolume.FileSystem}
		return volumeMetadata, volumeConfigDetails, nil
	}
	return model.VolumeMetadata{}, nil, fmt.Errorf("Volume not found")
}

func (s *spectrumLocalClient) Attach(name string) (mountPath string, err error) {
	s.logger.Println("spectrumLocalClient: attach start")
	defer s.logger.Println("spectrumLocalClient: attach end")

	s.filelock.Lock()
	defer func() {
		lockErr := s.filelock.Unlock()
		if lockErr != nil && err == nil {
			err = lockErr
		}
	}()

	volExists, err := s.dbClient.VolumeExists(name)

	if err != nil {
		s.logger.Println(err.Error())
		return "", err
	}

	if !volExists {
		return "", fmt.Errorf("Volume not found")
	}

	existingVolume, err := s.dbClient.GetVolume(name)

	if err != nil {
		s.logger.Println(err.Error())
		return "", err
	}

	if existingVolume.Mountpoint != "" {
		return existingVolume.Mountpoint, nil
	}
	mountpoint, err := s.client.GetFilesystemMountpoint(existingVolume.FileSystem)
	if err != nil {
		s.logger.Println(err.Error())
		return "", err
	}
	if existingVolume.Type == utils.LIGHTWEIGHT {

		mountPath = path.Join(mountpoint, existingVolume.Fileset, existingVolume.Directory)
	} else {

		isFilesetLinked, err := s.client.IsFilesetLinked(existingVolume.FileSystem, existingVolume.Fileset)

		if err != nil {
			s.logger.Println(err.Error())
			return "", err
		}

		if !isFilesetLinked {

			err = s.client.LinkFileset(existingVolume.FileSystem, existingVolume.Fileset)

			if err != nil {
				s.logger.Println(err.Error())
				return "", err
			}
		}

		mountPath = path.Join(mountpoint, existingVolume.Fileset)
	}

	// change owner of linked fileset if User and Group specified.
	if len(existingVolume.AdditionalData) > 0 {

		uid, uidSpecified := existingVolume.AdditionalData[USER_SPECIFIED_UID]
		gid, gidSpecified := existingVolume.AdditionalData[USER_SPECIFIED_GID]

		if uidSpecified && gidSpecified {
			err := s.changePermissionsOfFileset(existingVolume.FileSystem, existingVolume.Fileset, uid, gid)

			if err != nil {
				s.logger.Println(err.Error())
				return "", err
			}
		}
	}

	err = s.dbClient.UpdateVolumeMountpoint(name, mountPath)

	if err != nil {
		s.logger.Println(err.Error())
		return "", fmt.Errorf("internal error updating database")
	}

	return mountPath, nil
}

func (s *spectrumLocalClient) Detach(name string) (err error) {
	s.logger.Println("spectrumLocalClient: detach start")
	defer s.logger.Println("spectrumLocalClient: detach end")

	s.filelock.Lock()
	defer func() {
		lockErr := s.filelock.Unlock()
		if lockErr != nil && err == nil {
			err = lockErr
		}
	}()

	volExists, err := s.dbClient.VolumeExists(name)

	if err != nil {
		s.logger.Println(err.Error())
		return err
	}

	if !volExists {
		return fmt.Errorf("Volume not found")
	}

	existingVolume, err := s.dbClient.GetVolume(name)

	if err != nil {
		s.logger.Println(err.Error())
		return err
	}

	if existingVolume.Mountpoint == "" {
		return fmt.Errorf("volume not attached")
	}

	err = s.dbClient.UpdateVolumeMountpoint(name, "")

	if err != nil {
		s.logger.Println(err.Error())
		return fmt.Errorf("internal error updating database")
	}
	return nil
}

func (s *spectrumLocalClient) createFilesetVolume(filesystem, name string, opts map[string]interface{}) error {
	s.logger.Println("spectrumLocalClient: createFilesetVolume start")
	defer s.logger.Println("spectrumLocalClient: createFilesetVolume end")

	filesetName := generateFilesetName(name)

	err := s.client.CreateFileset(filesystem, filesetName, opts)

	if err != nil {
		return err
	}

	err = s.dbClient.InsertFilesetVolume(filesetName, name, filesystem, opts)

	if err != nil {
		return err
	}

	s.logger.Printf("Created fileset volume with fileset %s\n", filesetName)
	return nil
}
func (s *spectrumLocalClient) changePermissionsOfFileset(filesystem, filesetName, uid, gid string) error {
	s.logger.Println("spectrumLocalClient: changeOwnerOfFileset start")
	defer s.logger.Println("spectrumLocalClient: changeOwnerOfFileset end")

	s.logger.Printf("Changing Owner of Fileset %s to uid %s , gid %s", filesetName, uid, gid)

	mountpoint, err := s.client.GetFilesystemMountpoint(filesystem)
	if err != nil {
		return fmt.Errorf("Failed to change permissions of fileset %s : %s", filesetName, err.Error())
	}

	filesetPath := path.Join(mountpoint, filesetName)
	args := []string{"chown", "-R", fmt.Sprintf("%s:%s", uid, gid), filesetPath}
	cmd := exec.Command("sudo", args...)
	_, err = cmd.Output()
	if err != nil {
		return fmt.Errorf("Failed to change permissions of fileset %s: %s", filesetName, err.Error())
	}
	return nil
}

func (s *spectrumLocalClient) createFilesetQuotaVolume(filesystem, name, quota string, opts map[string]interface{}) error {
	s.logger.Println("spectrumLocalClient: createFilesetQuotaVolume start")
	defer s.logger.Println("spectrumLocalClient: createFilesetQuotaVolume end")

	filesetName := generateFilesetName(name)

	err := s.client.CreateFileset(filesystem, name, opts)

	if err != nil {
		return err
	}

	err = s.client.SetFilesetQuota(filesystem, filesetName, quota)

	if err != nil {
		deleteErr := s.client.DeleteFileset(filesystem, filesetName)
		if deleteErr != nil {
			return fmt.Errorf("Error setting quota (rollback error on delete fileset %s - manual cleanup needed)", filesetName)
		}
		return err
	}

	err = s.dbClient.InsertFilesetQuotaVolume(filesetName, quota, name, filesystem, opts)

	if err != nil {
		return err
	}

	s.logger.Printf("Created fileset volume with fileset %s, quota %s\n", filesetName, quota)
	return nil
}

func (s *spectrumLocalClient) createLightweightVolume(filesystem, name, fileset string, opts map[string]interface{}) error {
	s.logger.Println("spectrumLocalClient: createLightweightVolume start")
	defer s.logger.Println("spectrumLocalClient: createLightweightVolume end")

	filesetLinked, err := s.client.IsFilesetLinked(filesystem, fileset)

	if err != nil {
		s.logger.Println(err.Error())
		return fmt.Errorf("Error finding fileset '%s' on filesystem '%s'", fileset, filesystem)
	}

	if !filesetLinked {
		err = s.client.LinkFileset(filesystem, fileset)

		if err != nil {
			s.logger.Println(err.Error())
			return err
		}
	}

	lightweightVolumeName := generateLightweightVolumeName(name)

	mountpoint, err := s.client.GetFilesystemMountpoint(filesystem)
	if err != nil {
		s.logger.Println(err.Error())
		return err
	}

	lightweightVolumePath := path.Join(mountpoint, fileset, lightweightVolumeName)

	err = os.Mkdir(lightweightVolumePath, 0755)

	if err != nil {
		return fmt.Errorf("Failed to create directory path %s : %s", lightweightVolumePath, err.Error())
	}

	err = s.dbClient.InsertLightweightVolume(fileset, lightweightVolumeName, name, filesystem, opts)

	if err != nil {
		return err
	}

	s.logger.Printf("Created LightWeight volume at directory path: %s\n", lightweightVolumePath)
	return nil
}

func generateLightweightVolumeName(name string) string {
	return name //TODO: check for convension/valid names
}

func (s *spectrumLocalClient) ListVolumes() ([]model.VolumeMetadata, error) {
	s.logger.Println("spectrumLocalClient: list start")
	defer s.logger.Println("spectrumLocalClient: list end")
	var err error
	s.filelock.Lock()
	defer func() {
		lockErr := s.filelock.Unlock()
		if lockErr != nil && err == nil {
			err = lockErr
		}
	}()

	volumesInDb, err := s.dbClient.ListVolumes()

	if err != nil {
		s.logger.Printf("error retrieving volumes from db %#v\n", err)
		return nil, err
	}
	s.logger.Printf("Volumes in db: %d\n", len(volumesInDb))
	var volumes []model.VolumeMetadata
	for _, volume := range volumesInDb {
		s.logger.Printf("Volume from db: %#v\n", volume)
		volumes = append(volumes, model.VolumeMetadata{Name: volume.Name, Mountpoint: volume.Mountpoint})
	}

	return volumes, nil
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

	_, err := s.client.ListFileset(filesystem, userSpecifiedFileset)
	if err != nil {
		s.logger.Printf("Fileset does not exist %v", err.Error())
		return err
	}

	err = s.dbClient.InsertFilesetVolume(userSpecifiedFileset, name, filesystem, opts)

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

	filesetQuota, err := s.client.ListFilesetQuota(filesystem, userSpecifiedFileset)

	if err != nil {
		s.logger.Println(err.Error())
		return err
	}
	if filesetQuota != quota {
		s.logger.Printf("Mismatch between user-specified and listed quota for fileset %s", userSpecifiedFileset)
		return fmt.Errorf("Mismatch between user-specified and listed quota for fileset %s", userSpecifiedFileset)

	}

	err = s.dbClient.InsertFilesetQuotaVolume(userSpecifiedFileset, quota, name, filesystem, opts)

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

	linked, err := s.client.IsFilesetLinked(filesystem, userSpecifiedFileset)
	if err != nil {
		return err
	}
	if linked == false {
		err := s.client.LinkFileset(filesystem, userSpecifiedFileset)
		if err != nil {
			return err
		}
	}

	mountpoint, err := s.client.GetFilesystemMountpoint(filesystem)
	if err != nil {
		s.logger.Println(err.Error())
		return err
	}

	directoryPath := path.Join(mountpoint, userSpecifiedFileset, userSpecifiedDirectory)

	_, err = os.Stat(directoryPath)

	if err != nil {
		if os.IsNotExist(err) {
			s.logger.Printf("directory path %s doesn't exist", directoryPath)
			return err
		}

		s.logger.Printf("Error stating directoryPath %s: %s", directoryPath, err.Error())
		return err
	}

	err = s.dbClient.InsertLightweightVolume(userSpecifiedFileset, userSpecifiedDirectory, name, filesystem, opts)

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
	userSpecifiedType, err := determineTypeFromRequest(logger, opts)
	if err != nil {
		logger.Printf("%s", err.Error())
		return false, "", "", "", err
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
