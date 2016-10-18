package local

import (
	"fmt"
	"log"
	"strings"

	"github.ibm.com/almaden-containers/ubiquity.git/model"
	"github.ibm.com/almaden-containers/ubiquity.git/utils"

	"os"
	"os/exec"
	"os/signal"
	"path"
	"strconv"
	"syscall"
	"time"
)

type spectrumLocalClient struct {
	logger                                   *log.Logger
	defaultFilesystem                        string
	mountpoint                               string
	dbClient                                 *utils.DatabaseClient
	filelock                                 *utils.FileLock
	filesetForLightWeightVolumes             string
	isActivated                              bool
	isFilesetForLightweightVolumeInitialized bool
	isMounted                                bool
	//spectrumClients map[string]common.SpectrumClient  // cached SpectrumClient instance (one per service plan)
}

const (
	USER_SPECIFIED_TYPE string = "type"

	USER_SPECIFIED_DIRECTORY string = "directory"
	USER_SPECIFIED_QUOTA     string = "quota"

	USER_SPECIFIED_FILESET    string = "fileset"
	USER_SPECIFIED_FILESYSTEM string = "filesystem"

	FILESET_TYPE  string = "fileset"
	LTWT_VOL_TYPE string = "lightweight"

	FILESET = iota
	LIGHTWEIGHT
	FILESET_QUOTA
)

func NewSpectrumLocalClient(logger *log.Logger, mountpoint, filesetForLightWeightVolumes, defaultFilesystem string) (model.StorageClient, error) {

	dbClient := utils.NewDatabaseClient(logger, mountpoint)
	err := dbClient.Init()
	if err != nil {
		logger.Fatalln(err.Error())
		return nil, err
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

	return &spectrumLocalClient{logger: logger, mountpoint: mountpoint, dbClient: dbClient,
		filelock: utils.NewFileLock(logger, mountpoint), filesetForLightWeightVolumes: filesetForLightWeightVolumes, defaultFilesystem: defaultFilesystem}, nil
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
	mounted, err := s.isSpectrumScaleMounted(s.defaultFilesystem)

	if err != nil {
		s.logger.Println(err.Error())
		return err
	}

	if mounted == false {
		err = s.mount(s.defaultFilesystem)

		if err != nil {
			s.logger.Println(err.Error())
			return err
		}
	}

	clusterId, err := getClusterId()

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

func (s *spectrumLocalClient) GetPluginName() string {
	return "spectrum"
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

	if len(opts) == 0 {
		//fileset
		return s.createFilesetVolume(s.defaultFilesystem, name)
	}
	userSpecifiedType, err := determineTypeFromRequest(opts)
	if err != nil {
		return err
	}

	isExistingVolume, filesystem, existingFileset, existingLightWeightDir, err := s.validateAndParseParams(opts)
	if err != nil {
		return err
	}

	if isExistingVolume && userSpecifiedType == FILESET_TYPE {
		quota, quotaSpecified := opts[USER_SPECIFIED_QUOTA]
		if quotaSpecified {
			return s.updateDBWithExistingFilesetQuota(filesystem, name, existingFileset, quota.(string))
		}
		return s.updateDBWithExistingFileset(filesystem, name, existingFileset)
	}

	if isExistingVolume && userSpecifiedType == LTWT_VOL_TYPE {
		return s.updateDBWithExistingDirectory(filesystem, name, existingFileset, existingLightWeightDir)
	}

	if userSpecifiedType == FILESET_TYPE {
		quota, quotaSpecified := opts[USER_SPECIFIED_QUOTA]
		if quotaSpecified {
			return s.createFilesetQuotaVolume(filesystem, name, quota)
		}
		return s.createFilesetVolume(filesystem, name)
	}
	if userSpecifiedType == LTWT_VOL_TYPE {
		return s.createLightweightVolume(filesystem, name, existingFileset)
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

	if volExists {

		existingVolume, err := s.dbClient.GetVolume(name)

		if err != nil {
			s.logger.Println(err.Error())
			return err
		}

		if existingVolume.VolumeType == FILESET ||
			existingVolume.VolumeType == FILESET_QUOTA {

			isFilesetLinked, err := s.isFilesetLinked(existingVolume.FileSystem, existingVolume.Fileset)

			if err != nil {
				s.logger.Println(err.Error())
				return err
			}

			if isFilesetLinked {
				err := s.unlinkFileset(existingVolume.FileSystem, existingVolume.Fileset)

				if err != nil {
					s.logger.Println(err.Error())
					return err
				}
			}
			//if forceDelete {
			//	err = s.deleteFileset(existingVolume.FileSystem, existingVolume.Fileset)
			//
			//	if err != nil {
			//		s.logger.Println(err.Error())
			//		return err
			//	}
			//}
		}
		//else if existingVolume.VolumeType == LIGHTWEIGHT && forceDelete {
		//
		//		lightweightVolumePath := path.Join(s.mountpoint, existingVolume.Fileset, existingVolume.Directory)
		//
		//		err := os.RemoveAll(lightweightVolumePath)
		//
		//		if err != nil {
		//			s.logger.Println(err.Error())
		//			return err
		//		}
		//	}

		err = s.dbClient.DeleteVolume(name)

		if err != nil {
			s.logger.Println(err.Error())
			return err
		}
		return nil
	}
	return fmt.Errorf("Volume not found")
}

//GetVolume(string) (*model.VolumeMetadata, *string, *map[string]interface {}, error)
func (s *spectrumLocalClient) GetVolume(name string) (volumeMetadata model.VolumeMetadata, volumeConfigDetails model.SpectrumConfig, err error) {
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
		return model.VolumeMetadata{}, model.SpectrumConfig{}, err
	}

	if volExists {

		existingVolume, err := s.dbClient.GetVolume(name)

		if err != nil {
			s.logger.Println(err.Error())
			return model.VolumeMetadata{}, model.SpectrumConfig{}, err
		}

		volumeMetadata = model.VolumeMetadata{Name: existingVolume.VolumeName, Mountpoint: existingVolume.Mountpoint}
		volumeConfigDetails = model.SpectrumConfig{FilesetId: existingVolume.Fileset, Filesystem: existingVolume.FileSystem}
		return volumeMetadata, volumeConfigDetails, nil
	}
	return model.VolumeMetadata{}, model.SpectrumConfig{}, fmt.Errorf("Volume not found")
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

	if existingVolume.VolumeType == LIGHTWEIGHT {
		mountPath = path.Join(s.mountpoint, existingVolume.Fileset, existingVolume.Directory)
	} else {

		isFilesetLinked, err := s.isFilesetLinked(existingVolume.FileSystem, existingVolume.Fileset)

		if err != nil {
			s.logger.Println(err.Error())
			return "", err
		}

		if !isFilesetLinked {

			err = s.linkFileset(existingVolume.FileSystem, existingVolume.Fileset)

			if err != nil {
				s.logger.Println(err.Error())
				return "", err
			}
		}

		mountPath = path.Join(s.mountpoint, existingVolume.Fileset)
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

//func (s *spectrumLocalClient) isLightweightVolumesInitialized() (bool, error) {
//	s.logger.Println("spectrumLocalClient: isLightweightVolumesInitialized start")
//	defer s.logger.Println("spectrumLocalClient: isLightweightVolumesInitialized end")
//
//	isDirFilesetLinked, err := s.isFilesetLinked(s.filesetForLightWeightVolumes)
//
//	if err != nil {
//		return false, fmt.Errorf("Lightweight volumes not initialized: %s", err)
//	}
//
//	if !isDirFilesetLinked {
//		return false, fmt.Errorf("Lightweight volumes not initialized: fileset %s not linked", s.filesetForLightWeightVolumes)
//	}
//	return true, nil
//}

func (s *spectrumLocalClient) isFilesetLinked(filesystem, filesetName string) (bool, error) {
	s.logger.Println("spectrumLocalClient: isFilesetLinked start")
	defer s.logger.Println("spectrumLocalClient: isFilesetLinked end")

	spectrumCommand := "/usr/lpp/mmfs/bin/mmlsfileset"
	args := []string{filesystem, filesetName, "-Y"}
	cmd := exec.Command(spectrumCommand, args...)
	outputBytes, err := cmd.Output()
	if err != nil {
		return false, err
	}

	spectrumOutput := string(outputBytes)

	lines := strings.Split(spectrumOutput, "\n")

	if len(lines) == 1 {
		return false, fmt.Errorf("Error listing fileset %s", filesetName)
	}

	tokens := strings.Split(lines[1], ":")
	if len(tokens) >= 11 {
		if tokens[10] == "Linked" {
			return true, nil
		} else {
			return false, nil
		}
	}

	return false, fmt.Errorf("Error listing fileset %s after parsing", filesetName)
}

func getClusterId() (string, error) {

	var clusterId string

	spectrumCommand := "/usr/lpp/mmfs/bin/mmlscluster"
	cmd := exec.Command(spectrumCommand)
	outputBytes, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("Error running command: %s", err.Error())
	}
	spectrumOutput := string(outputBytes)

	lines := strings.Split(spectrumOutput, "\n")
	tokens := strings.Split(lines[4], ":")

	if len(tokens) == 2 {
		if strings.TrimSpace(tokens[0]) == "GPFS cluster id" {
			clusterId = strings.TrimSpace(tokens[1])
		}
	}
	return clusterId, nil
}

func (s *spectrumLocalClient) mount(filesystem string) (err error) {
	s.logger.Println("spectrumLocalClient: mount start")
	defer s.logger.Println("spectrumLocalClient: mount end")

	if s.isMounted == true {
		return nil
	}

	spectrumCommand := "/usr/lpp/mmfs/bin/mmmount"
	args := []string{filesystem}
	cmd := exec.Command(spectrumCommand, args...)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("Failed to mount filesystem")
	}
	s.logger.Println(output)
	s.isMounted = true
	return nil
}

func (s *spectrumLocalClient) isSpectrumScaleMounted(filesystem string) (isMounted bool, err error) {
	s.logger.Println("spectrumLocalClient: isMounted start")
	defer s.logger.Println("spectrumLocalClient: isMounted end")

	if s.isMounted == true {
		isMounted = true
		return isMounted, nil
	}

	spectrumCommand := "/usr/lpp/mmfs/bin/mmlsmount"
	args := []string{filesystem, "-L", "-Y"}
	cmd := exec.Command(spectrumCommand, args...)
	outputBytes, err := cmd.Output()
	if err != nil {
		s.logger.Printf("Error running command\n")
		s.logger.Println(err)
		return false, err
	}
	mountedNodes := extractMountedNodes(string(outputBytes))
	if len(mountedNodes) == 0 {
		//not mounted anywhere
		isMounted = false
		return isMounted, nil
	} else {
		// checkif mounted on current node -- compare node name
		currentNode, _ := os.Hostname()
		s.logger.Printf("spectrumLocalClient: node name: %s\n", currentNode)
		for _, node := range mountedNodes {
			if node == currentNode {
				s.isMounted = true
				isMounted = true
				return isMounted, nil
			}
		}
	}
	isMounted = false
	return isMounted, nil
}

func extractMountedNodes(spectrumOutput string) []string {
	var nodes []string
	lines := strings.Split(spectrumOutput, "\n")
	if len(lines) == 1 {
		return nodes
	}
	for _, line := range lines[1:] {
		tokens := strings.Split(line, ":")
		if len(tokens) > 10 {
			if tokens[11] != "" {
				nodes = append(nodes, tokens[11])
			}
		}
	}
	return nodes
}

//func (s *spectrumLocalClient) create(name string, opts map[string]interface{}) error {
//	s.logger.Println("spectrumLocalClient: createNew start")
//	defer s.logger.Println("spectrumLocalClient: createNew end")
//
//	var err error
//	if len(opts) > 0 {
//		userSpecifiedType, typeExists := opts[TYPE_OPT]
//		userSpecifiedQuota, quotaExists := opts[QUOTA_OPT]
//		userSpecifiedFilesystem, userSpecifiedFilesystemExists := opts[FILESYSTEM_OPT]
//		if userSpecifiedFilesystemExists == false {
//			userSpecifiedFilesystem = s.defaultFilesystem
//		}
//
//		if typeExists {
//			if userSpecifiedType.(string) == FILESET_TYPE {
//				if quotaExists {
//					err = s.createFilesetQuotaVolume(userSpecifiedFilesystem, name, userSpecifiedQuota.(string))
//				} else {
//					err = s.createFilesetVolume(userSpecifiedFilesystem, name)
//				}
//			} else if userSpecifiedType.(string) == LTWT_VOL_TYPE {
//				err = s.createLightweightVolume(userSpecifiedFilesystem, name, opts)
//			} else {
//				return fmt.Errorf("Invalid type %s", userSpecifiedType.(string))
//			}
//		} else if quotaExists {
//			err = s.createFilesetQuotaVolume(userSpecifiedFilesystem, name, userSpecifiedQuota.(string))
//		}
//	} else {
//		err = s.createFilesetVolume(s.defaultFilesystem, name)
//	}
//
//	if err != nil {
//		s.logger.Println(err.Error())
//		return err
//	}
//
//	return nil
//}

func (s *spectrumLocalClient) createFilesetVolume(filesystem, name string) error {
	s.logger.Println("spectrumLocalClient: createFilesetVolume start")
	defer s.logger.Println("spectrumLocalClient: createFilesetVolume end")

	filesetName := generateFilesetName()

	err := s.createFileset(filesystem, filesetName)

	if err != nil {
		return err
	}

	err = s.dbClient.InsertFilesetVolume(filesetName, name, filesystem)

	if err != nil {
		return err
	}

	s.logger.Printf("Created fileset volume with fileset %s\n", filesetName)
	return nil
}

func (s *spectrumLocalClient) createFilesetQuotaVolume(filesystem, name, quota string) error {
	s.logger.Println("spectrumLocalClient: createFilesetQuotaVolume start")
	defer s.logger.Println("spectrumLocalClient: createFilesetQuotaVolume end")

	filesetName := generateFilesetName()

	err := s.createFileset(filesystem, filesetName)

	if err != nil {
		return err
	}

	err = s.setFilesetQuota(filesystem, filesetName, quota)

	if err != nil {
		return err
	}

	err = s.dbClient.InsertFilesetQuotaVolume(filesetName, quota, name, filesystem)

	if err != nil {
		return err
	}

	s.logger.Printf("Created fileset volume with fileset %s, quota %s\n", filesetName, quota)
	return nil
}

func (s *spectrumLocalClient) setFilesetQuota(filesystem, filesetName, quota string) error {
	s.logger.Println("spectrumLocalClient: setFilesetQuota start")
	defer s.logger.Println("spectrumLocalClient: setFilesetQuota end")

	s.logger.Printf("setting quota to %s for fileset %s\n", quota, filesetName)

	spectrumCommand := "/usr/lpp/mmfs/bin/mmsetquota"
	args := []string{filesystem + ":" + filesetName, "--block", quota + ":" + quota}
	cmd := exec.Command(spectrumCommand, args...)
	output, err := cmd.Output()

	if err != nil {
		return fmt.Errorf("Failed to set quota for fileset %s", filesetName)
	}

	s.logger.Printf("setFilesetQuota output: %s\n", string(output))
	return nil
}

func (s *spectrumLocalClient) createLightweightVolume(filesystem, name, fileset string) error {
	s.logger.Println("spectrumLocalClient: createLightweightVolume start")
	defer s.logger.Println("spectrumLocalClient: createLightweightVolume end")

	filesetLinked, err := s.isFilesetLinked(filesystem, fileset)

	if err != nil {
		s.logger.Println(err.Error())
		return err
	}

	if !filesetLinked {
		err = s.linkFileset(filesystem, fileset)

		if err != nil {
			s.logger.Println(err.Error())
			return err
		}
	}
	//lightweightVolumeFileset = userSpecifiedFileset.(string)
	//
	//else {
	//		if !s.isFilesetForLightweightVolumeInitialized {
	//			err := s.initLightweightVolumes()
	//
	//			if err != nil {
	//				s.logger.Println(err.Error())
	//				return err
	//			}
	//			s.isFilesetForLightweightVolumeInitialized = true
	//		}
	//		lightweightVolumeFileset = s.filesetForLightWeightVolumes
	//	}

	lightweightVolumeName := generateLightweightVolumeName()

	lightweightVolumePath := path.Join(s.mountpoint, fileset, lightweightVolumeName)

	err = os.Mkdir(lightweightVolumePath, 0755)

	if err != nil {
		return fmt.Errorf("Failed to create directory path %s : %s", lightweightVolumePath, err.Error())
	}

	err = s.dbClient.InsertLightweightVolume(fileset, lightweightVolumeName, name, filesystem)

	if err != nil {
		return err
	}

	s.logger.Printf("Created LightWeight volume at directory path: %s\n", lightweightVolumePath)
	return nil
}

func (s *spectrumLocalClient) linkFileset(filesystem, filesetName string) error {
	s.logger.Println("spectrumLocalClient: linkFileset start")
	defer s.logger.Println("spectrumLocalClient: linkFileset end")

	spectrumCommand := "/usr/lpp/mmfs/bin/mmlinkfileset"
	filesetPath := path.Join(s.mountpoint, filesetName)
	args := []string{filesystem, filesetName, "-J", filesetPath}
	cmd := exec.Command(spectrumCommand, args...)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("Failed to link fileset: %s", err.Error())
	}
	s.logger.Printf("spectrumLocalClient: Linkfileset output: %s\n", string(output))

	//hack for now
	args = []string{"-R", "777", filesetPath}
	cmd = exec.Command("chmod", args...)
	output, err = cmd.Output()
	if err != nil {
		return fmt.Errorf("Failed to set permissions for fileset: %s", err.Error())
	}
	return nil
}

func (s *spectrumLocalClient) unlinkFileset(filesystem, filesetName string) error {
	s.logger.Println("spectrumLocalClient: unlinkFileset start")
	defer s.logger.Println("spectrumLocalClient: unlinkFileset end")

	spectrumCommand := "/usr/lpp/mmfs/bin/mmunlinkfileset"
	args := []string{filesystem, filesetName}
	cmd := exec.Command(spectrumCommand, args...)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("Failed to unlink fileset %s: %s", filesetName, err.Error())
	}
	s.logger.Printf("spectrumLocalClient: unLinkfileset output: %s\n", string(output))
	return nil
}

func (s *spectrumLocalClient) createFileset(filesystem, filesetName string) error {
	s.logger.Println("spectrumLocalClient: createFileset start")
	defer s.logger.Println("spectrumLocalClient: createFileset end")

	s.logger.Printf("creating a new fileset: %s\n", filesetName)

	// create fileset
	spectrumCommand := "/usr/lpp/mmfs/bin/mmcrfileset"
	args := []string{filesystem, filesetName}
	cmd := exec.Command(spectrumCommand, args...)
	output, err := cmd.Output()

	if err != nil {
		return fmt.Errorf("Failed to create fileset %s", filesetName)
	}

	s.logger.Printf("Createfileset output: %s\n", string(output))
	return nil
}

func generateLightweightVolumeName() string {
	return "LtwtVol" + strconv.FormatInt(time.Now().UnixNano(), 10)
}

func (s *spectrumLocalClient) deleteFileset(filesystem, filesetName string) error {
	s.logger.Println("spectrumLocalClient: deleteFileset start")
	defer s.logger.Println("spectrumLocalClient: deleteFileset end")

	spectrumCommand := "/usr/lpp/mmfs/bin/mmdelfileset"
	args := []string{filesystem, filesetName, "-f"}
	cmd := exec.Command(spectrumCommand, args...)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("Failed to remove fileset %s: %s ", filesetName, err.Error())
	}
	s.logger.Printf("spectrumLocalClient: deleteFileset output: %s\n", string(output))
	return nil
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
		volumes = append(volumes, model.VolumeMetadata{Name: volume.VolumeName, Mountpoint: volume.Mountpoint})
	}

	return volumes, nil
}

func generateFilesetName() string {
	return strconv.FormatInt(time.Now().UnixNano(), 10)
}

//TODO move updates to DB file

func (s *spectrumLocalClient) updateDBWithExistingFileset(filesystem, name, userSpecifiedFileset string) error {
	s.logger.Println("spectrumLocalClient:  updateDBWithExistingFileset start")
	defer s.logger.Println("spectrumLocalClient: updateDBWithExistingFileset end")
	s.logger.Printf("User specified fileset: %s\n", userSpecifiedFileset)

	spectrumCommand := "/usr/lpp/mmfs/bin/mmlsfileset"
	args := []string{filesystem, userSpecifiedFileset, "-Y"}
	cmd := exec.Command(spectrumCommand, args...)
	_, err := cmd.Output()
	if err != nil {
		s.logger.Println(err)
		return err
	}

	err = s.dbClient.InsertFilesetVolume(userSpecifiedFileset, name, filesystem)

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

	if volumeConfigDetails.FilesetId != userSpecifiedFileset {
		return fmt.Errorf("volume %s with fileset %s not found", name, userSpecifiedFileset)
	}
	return nil
}

func (s *spectrumLocalClient) updateDBWithExistingFilesetQuota(filesystem, name, userSpecifiedFileset, quota string) error {
	s.logger.Println("spectrumLocalClient:  updateDBWithExistingFilesetQuota start")
	defer s.logger.Println("spectrumLocalClient: updateDBWithExistingFilesetQuota end")

	err := s.verifyFilesetQuota(filesystem, userSpecifiedFileset, quota)

	if err != nil {
		s.logger.Println(err.Error())
		return err
	}

	err = s.dbClient.InsertFilesetQuotaVolume(userSpecifiedFileset, quota, name)

	if err != nil {
		s.logger.Println(err.Error())
		return err
	}
	return nil
}

func (s *spectrumLocalClient) updateDBWithExistingDirectory(filesystem, name, userSpecifiedFileset, userSpecifiedDirectory string) error {
	s.logger.Println("spectrumLocalClient:  updateDBWithExistingDirectory start")
	defer s.logger.Println("spectrumLocalClient: updateDBWithExistingDirectory end")
	s.logger.Printf("User specified fileset: %s, User specified directory: %s\n", userSpecifiedFileset, userSpecifiedDirectory)

	linked, err := s.isFilesetLinked(filesystem, userSpecifiedFileset)
	if err != nil {
		return err
	}
	if linked == false {
		err := s.linkFileset(filesystem, userSpecifiedFileset)
		if err != nil {
			return err
		}
	}

	directoryPath := path.Join(s.mountpoint, userSpecifiedFileset, userSpecifiedDirectory)

	_, err = os.Stat(directoryPath)

	if err != nil {
		if os.IsNotExist(err) {
			s.logger.Printf("directory path %s doesn't exist", directoryPath)
			return err
		}

		s.logger.Printf("Error stating directoryPath %s: %s", directoryPath, err.Error())
		return err
	}

	err = s.dbClient.InsertLightweightVolume(userSpecifiedFileset, userSpecifiedDirectory, name, filesystem)

	if err != nil {
		s.logger.Println(err.Error())
		return err
	}
	return nil
}

func (s *spectrumLocalClient) verifyFilesetQuota(filesystem, filesetName, quota string) error {
	s.logger.Println("spectrumLocalClient: verifyFilesetQuota start")
	defer s.logger.Println("spectrumLocalClient: verifyFilesetQuota end")

	spectrumCommand := "/usr/lpp/mmfs/bin/mmlsquota"
	args := []string{"-j", filesetName, filesystem, "--block-size", "auto"}

	cmd := exec.Command(spectrumCommand, args...)
	outputBytes, err := cmd.Output()

	if err != nil {
		return fmt.Errorf("Failed to list quota for fileset %s: %s", filesetName, err.Error())
	}

	spectrumOutput := string(outputBytes)

	lines := strings.Split(spectrumOutput, "\n")

	if len(lines) > 2 {
		tokens := strings.Fields(lines[2])

		if len(tokens) > 3 {
			if tokens[3] == quota {
				return nil
			}
		} else {
			fmt.Errorf("Error parsing tokens while listing quota for fileset %s", filesetName)
		}
	}
	return fmt.Errorf("Mismatch between user-specified and listed quota for fileset %s", filesetName)
}

func determineTypeFromRequest(opts map[string]interface{}) (string, error) {
	userSpecifiedType, exists := opts[USER_SPECIFIED_TYPE]
	if exists == nil {
		_, exists := opts[USER_SPECIFIED_DIRECTORY]
		if exists != nil {
			return LTWT_VOL_TYPE, nil
		}
		return FILESET_TYPE, nil
	}

	if userSpecifiedType.(string) != FILESET_TYPE || userSpecifiedType.(string) != LTWT_VOL_TYPE {
		return "", fmt.Errorf("Unknown 'type' = %s specified", userSpecifiedType.(string))
	}

	return userSpecifiedType.(string), nil
}
func (s *spectrumLocalClient) validateAndParseParams(opts map[string]interface{}) (bool, string, string, string, error) {
	existingFileset, existingFilesetSpecified := opts[USER_SPECIFIED_FILESET]
	existingLightWeightDir, existingLightWeightDirSpecified := opts[USER_SPECIFIED_DIRECTORY]
	filesystem, filesystemSpecified := opts[USER_SPECIFIED_FILESYSTEM]

	if existingFilesetSpecified || existingLightWeightDirSpecified {
		if filesystemSpecified == false {
			return true, filesystem, existingFileset, existingLightWeightDir, fmt.Errorf("'filesystem' is a required opt for using existing volumes")
		}
		if existingLightWeightDirSpecified && !existingFilesetSpecified {
			return true, filesystem, existingFileset, existingLightWeightDir, fmt.Errorf("'fileset' is a required opt for using existing lightweight volumes")
		}
		return true, filesystem, existingFileset, existingLightWeightDir, nil
	} else if filesystemSpecified == false {
		return false, s.defaultFilesystem, "", "", nil

	}

	return false, filesystem, existingFileset, existingLightWeightDir, nil
}
