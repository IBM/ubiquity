package core

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"

	"errors"

	"github.ibm.com/almaden-containers/spectrum-common.git/models"
)

//go:generate counterfeiter -o ../fakes/fake_spectrum_client.go . SpectrumClient

type SpectrumClient interface {
	Activate() error
	Create(name string, opts map[string]interface{}) error
	CreateWithoutProvisioning(name string, opts map[string]interface{}) error
	Remove(name string) error
	Attach(name string) (string, error)
	Detach(name string) error
	ExportNfs(name string, clientCIDR string) (string, error)
	UnexportNfs(name string) error
	List() ([]models.VolumeMetadata, error)
	Get(name string) (*models.VolumeMetadata, *models.SpectrumConfig, error)
	RemoveWithoutDeletingVolume(string) error
	GetFileSetForMountPoint(mountPoint string) (string, error)
}

type Fileset struct {
	Name             string
	Mountpoint       string
	DockerVolumeName string
}

type MappingConfig struct {
	Mappings map[string]Fileset
}

const (
	LIGHTWEIGHT_VOLUME_FILESET string = "LightweightVolumes"

	TYPE_OPT string = "type"
	DIR_OPT string = "directory"
	QUOTA_OPT string = "quota"
	FILESET_OPT string = "fileset"

	FILESET_TYPE string = "fileset"
	LTWT_VOL_TYPE string = "lightweight"

)

func NewSpectrumClient(logger *log.Logger, filesystem, mountpoint string, dbclient *DatabaseClient) SpectrumClient {
	return &MMCliFilesetClient{log: logger, Filesystem: filesystem, Mountpoint: mountpoint, DbClient: dbclient,
		filelock: NewFileLock(logger, filesystem, mountpoint), LightweightVolumeFileset: LIGHTWEIGHT_VOLUME_FILESET}
}

type MMCliFilesetClient struct {
	Filesystem                     string
	Mountpoint                     string
	log                            *log.Logger
	DbClient                       *DatabaseClient
	isMounted                      bool
	isActivated                    bool
	filelock                       *FileLock
	LightweightVolumeFileset       string
	isLightweightVolumeInitialized bool
}

func (m *MMCliFilesetClient) Activate() (err error) {
	m.log.Println("MMCliFilesetClient: Activate start")
	defer m.log.Println("MMCliFilesetClient: Activate end")

	m.filelock.Lock()
	defer func() {
		lockErr := m.filelock.Unlock()
		if lockErr != nil && err == nil {
			err = lockErr
		}
	}()

	if m.isActivated {
		return nil
	}

	//check if filesystem is mounted
	mounted, err := m.isSpectrumScaleMounted()

	if err != nil {
		m.log.Println(err.Error())
		return err
	}

	if mounted == false {
		err = m.mount()

		if err != nil {
			m.log.Println(err.Error())
			return err
		}
	}

	clusterId, err := getClusterId()

	if err != nil {
		m.log.Println(err.Error())
		return err
	}

	if len(clusterId) == 0 {
		clusterIdErr := errors.New("Unable to retrieve clusterId: clusterId is empty")
		m.log.Println(clusterIdErr.Error())
		return clusterIdErr
	}

	m.DbClient.ClusterId = clusterId

	err = m.DbClient.CreateVolumeTable()

	if err != nil {
		m.log.Println(err.Error())
		return err
	}

	m.isLightweightVolumeInitialized, _ = m.isLightweightVolumesInitialized()

	m.isActivated = true
	return nil
}

func (m *MMCliFilesetClient) Create(name string, opts map[string]interface{}) (err error) {
	m.log.Println("MMCliFilesetClient: create start")
	defer m.log.Println("MMCliFilesetClient: create end")

	m.filelock.Lock()
	defer func() {
		lockErr := m.filelock.Unlock()
		if lockErr != nil && err == nil {
			err = lockErr
		}
	}()

	volExists, err := m.DbClient.VolumeExists(name)

	if err != nil {
		m.log.Println(err.Error())
		return err
	}

	if volExists {
		return fmt.Errorf("Volume already exists")
	}

	if len(opts) > 0 {

		userSpecifiedType, typeExists := opts[TYPE_OPT]
		userSpecifiedFileset, filesetExists := opts[FILESET_OPT]
		userSpecifiedDirectory, dirExists := opts[DIR_OPT]
		userSpecifiedQuota, quotaExists := opts[QUOTA_OPT]

		if len(opts) == 1 {
			if typeExists || quotaExists {
				return m.create(name, opts)
			} else if filesetExists {
				return m.updateDBWithExistingFileset(name, userSpecifiedFileset.(string))
			} else if dirExists {
				return m.updateDBWithExistingDirectory(name, m.LightweightVolumeFileset, userSpecifiedDirectory.(string))
			}
			return errors.New("Invalid arguments")
		} else if len(opts) == 2 {
			if typeExists {
				if userSpecifiedType.(string) == FILESET_TYPE {
					if filesetExists {
						return m.updateDBWithExistingFileset(name, userSpecifiedFileset.(string))
					} else if quotaExists {
						return m.create(name, opts)
					}
					return errors.New("Invalid arguments")
				} else if userSpecifiedType.(string) == LTWT_VOL_TYPE {
					if filesetExists {
						return m.create(name, opts)
					} else if dirExists {
						return m.updateDBWithExistingDirectory(name, m.LightweightVolumeFileset, userSpecifiedDirectory.(string))
					}
				}
				return errors.New("Invalid arguments")
			} else if filesetExists {
				if dirExists {
					return m.updateDBWithExistingDirectory(name, userSpecifiedFileset.(string), userSpecifiedDirectory.(string))
				} else if quotaExists {
					return m.updateDBWithExistingFilesetQuota(name, userSpecifiedFileset.(string), userSpecifiedQuota.(string))
				}
			}
			return errors.New("Invalid arguments")
		} else if len(opts) == 3 {
			if typeExists {
				if userSpecifiedType.(string) == FILESET_TYPE && filesetExists && quotaExists {
					return m.updateDBWithExistingFilesetQuota(name, userSpecifiedFileset.(string), userSpecifiedQuota.(string))
				} else if userSpecifiedType.(string) == LTWT_VOL_TYPE && filesetExists && dirExists {
					return m.updateDBWithExistingDirectory(name, userSpecifiedFileset.(string), userSpecifiedDirectory.(string))
				}
			}
		}
		return errors.New("Invalid number of arguments")
	}

	return m.create(name, opts)
}

func (m *MMCliFilesetClient) CreateWithoutProvisioning(name string, opts map[string]interface{}) (err error) {
	m.log.Println("MMCliFilesetClient: CreateWithoutProvisioning start")
	defer m.log.Println("MMCliFilesetClient: createWithoutProvisioning end")

	m.filelock.Lock()
	defer func() {
		lockErr := m.filelock.Unlock()
		if lockErr != nil && err == nil {
			err = lockErr
		}
	}()

	volExists, err := m.DbClient.VolumeExists(name)

	if err != nil {
		m.log.Println(err.Error())
		return err
	}

	if volExists {
		return fmt.Errorf("Volume already exists")
	}
	userSpecifiedFileset, exists := opts[FILESET_OPT]
	if exists == true {
		return m.updateDBWithExistingFileset(name, userSpecifiedFileset.(string))
	} else {

		err := m.filesetExists(name)
		if err != nil {
			m.log.Printf("Fileset not found %#v", err)
			return err
		}

		err = m.DbClient.InsertFilesetVolume(userSpecifiedFileset.(string), name)

		if err != nil {
			m.log.Printf("Error persisting mapping %#v", err)
			return err
		}

	}
	return nil
}

func (m *MMCliFilesetClient) filesetExists(name string) error {
	m.log.Println("MMCliFilesetClient:  fileset exists start")
	defer m.log.Println("MMCliFilesetClient: fileset exists end")
	m.log.Printf("filesetExists: %s\n", name)

	spectrumCommand := "/usr/lpp/mmfs/bin/mmlsfileset"
	args := []string{m.Filesystem, name}
	cmd := exec.Command(spectrumCommand, args...)

	out, err := cmd.Output()
	if err != nil {
		m.log.Printf("error checking fileset %#v", err)
		return err
	}
	var line string
	reader := bytes.NewBufferString(string(out))
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line = (scanner.Text())
		lineSlice := strings.Split(line, " ")
		if lineSlice[0] == name {
			return nil
		}
	}
	m.log.Println("fileset not found")
	return fmt.Errorf("volume not found in the filesystem")
}

func (m *MMCliFilesetClient) updateDBWithExistingFileset(name, userSpecifiedFileset string) error {
	m.log.Println("MMCliFilesetClient:  updateDBWithExistingFileset start")
	defer m.log.Println("MMCliFilesetClient: updateDBWithExistingFileset end")
	m.log.Printf("User specified fileset: %s\n", userSpecifiedFileset)

	spectrumCommand := "/usr/lpp/mmfs/bin/mmlsfileset"
	args := []string{m.Filesystem, userSpecifiedFileset, "-Y"}
	cmd := exec.Command(spectrumCommand, args...)
	_, err := cmd.Output()
	if err != nil {
		m.log.Println(err)
		return err
	}

	err = m.DbClient.InsertFilesetVolume(userSpecifiedFileset, name)

	if err != nil {
		m.log.Println(err.Error())
		return err
	}
	return nil
}

func (m *MMCliFilesetClient) updateDBWithExistingFilesetQuota(name, userSpecifiedFileset, quota string) error {
	m.log.Println("MMCliFilesetClient:  updateDBWithExistingFilesetQuota start")
	defer m.log.Println("MMCliFilesetClient: updateDBWithExistingFilesetQuota end")

	err := m.verifyFilesetQuota(userSpecifiedFileset, quota)

	if err != nil {
		m.log.Println(err.Error())
		return err
	}

	err = m.DbClient.InsertFilesetQuotaVolume(userSpecifiedFileset, quota, name)

	if err != nil {
		m.log.Println(err.Error())
		return err
	}
	return nil
}

func (m *MMCliFilesetClient) updateDBWithExistingDirectory(name, userSpecifiedFileset, userSpecifiedDirectory string) error {
	m.log.Println("MMCliFilesetClient:  updateDBWithExistingDirectory start")
	defer m.log.Println("MMCliFilesetClient: updateDBWithExistingDirectory end")
	m.log.Printf("User specified fileset: %s, User specified directory: %s\n", userSpecifiedFileset, userSpecifiedDirectory)

	if userSpecifiedFileset != m.LightweightVolumeFileset {

		filesetLinked,err := m.isFilesetLinked(userSpecifiedFileset)

		if err != nil {
			m.log.Println(err.Error())
			return err
		}

		if !filesetLinked {
			err = fmt.Errorf("fileset %s not linked", userSpecifiedFileset)
			m.log.Println(err.Error())
			return err
		}
	} else {
		if !m.isLightweightVolumeInitialized {
			err := m.initLightweightVolumes()

			if err != nil {
				m.log.Println(err.Error())
				return err
			}
			m.isLightweightVolumeInitialized = true
		}
	}

	directoryPath := path.Join(m.Mountpoint, userSpecifiedFileset, userSpecifiedDirectory)

	_, err := os.Stat(directoryPath)

	if err != nil {
		if os.IsNotExist(err) {
			m.log.Printf("directory path %s doesn't exist", directoryPath)
			return err
		}

		m.log.Printf("Error stating directoryPath %s: %s", directoryPath, err.Error())
		return err
	}

	err = m.DbClient.InsertLightweightVolume(userSpecifiedFileset, userSpecifiedDirectory, name)

	if err != nil {
		m.log.Println(err.Error())
		return err
	}
	return nil
}

func (m *MMCliFilesetClient) create(name string, opts map[string]interface{}) error {
	m.log.Println("MMCliFilesetClient: createNew start")
	defer m.log.Println("MMCliFilesetClient: createNew end")

	var err error
	if len(opts) > 0 {
		userSpecifiedType, typeExists := opts[TYPE_OPT]
		userSpecifiedQuota, quotaExists := opts[QUOTA_OPT]

		if typeExists {
			if userSpecifiedType.(string) == FILESET_TYPE {
				if quotaExists {
					err = m.createFilesetQuotaVolume(name, userSpecifiedQuota.(string))
				} else {
					err = m.createFilesetVolume(name)
				}
			} else if userSpecifiedType.(string) == LTWT_VOL_TYPE {
				err = m.createLightweightVolume(name, opts)
			} else {
				return fmt.Errorf("Invalid type %s", userSpecifiedType.(string))
			}
		} else if quotaExists {
			err = m.createFilesetQuotaVolume(name, userSpecifiedQuota.(string))
		}
	} else {
		err = m.createFilesetVolume(name)
	}

	if err != nil {
		m.log.Println(err.Error())
		return err
	}

	return nil
}

func (m *MMCliFilesetClient) createFilesetVolume(name string) error {
	m.log.Println("MMCliFilesetClient: createFilesetVolume start")
	defer m.log.Println("MMCliFilesetClient: createFilesetVolume end")

	filesetName := generateFilesetName()

	err := m.createFileset(filesetName)

	if err != nil {
		return err
	}

	err =  m.DbClient.InsertFilesetVolume(filesetName, name)

	if err != nil {
		return err
	}

	m.log.Printf("Created fileset volume with fileset %s\n", filesetName)
	return nil
}

func (m *MMCliFilesetClient) createFilesetQuotaVolume(name, quota string) error {
	m.log.Println("MMCliFilesetClient: createFilesetQuotaVolume start")
	defer m.log.Println("MMCliFilesetClient: createFilesetQuotaVolume end")

	filesetName := generateFilesetName()

	err := m.createFileset(filesetName)

	if err != nil {
		return err
	}

	err = m.setFilesetQuota(filesetName, quota)

	if err != nil {
		return err
	}

	err = m.DbClient.InsertFilesetQuotaVolume(filesetName, quota, name)

	if err != nil {
		return err
	}

	m.log.Printf("Created fileset volume with fileset %s, quota %s\n", filesetName, quota)
	return nil
}

func (m *MMCliFilesetClient) createLightweightVolume(name string, opts map[string]interface{}) error {
	m.log.Println("MMCliFilesetClient: createLightweightVolume start")
	defer m.log.Println("MMCliFilesetClient: createLightweightVolume end")

	var lightweightVolumeFileset string
	userSpecifiedType, typeExists := opts[TYPE_OPT]
	userSpecifiedFileset, filesetExists := opts[FILESET_OPT]

	if len(opts) == 2 && typeExists && userSpecifiedType.(string) == LTWT_VOL_TYPE && filesetExists {

		filesetLinked,err := m.isFilesetLinked(userSpecifiedFileset.(string))

		if err != nil {
			m.log.Println(err.Error())
			return err
		}

		if !filesetLinked {
			err = m.linkFileset(userSpecifiedFileset.(string))

			if err != nil {
				m.log.Println(err.Error())
				return err
			}
		}
		lightweightVolumeFileset = userSpecifiedFileset.(string)
	} else {
		if !m.isLightweightVolumeInitialized {
			err := m.initLightweightVolumes()

			if err != nil {
				m.log.Println(err.Error())
				return err
			}
			m.isLightweightVolumeInitialized = true
		}
		lightweightVolumeFileset = m.LightweightVolumeFileset
	}

	lightweightVolumeName := generateLightweightVolumeName()

	lightweightVolumePath := path.Join(m.Mountpoint, lightweightVolumeFileset, lightweightVolumeName)

	err := os.Mkdir(lightweightVolumePath, 0755)

	if err != nil {
		return fmt.Errorf("Failed to create directory path %s : %s", lightweightVolumePath, err.Error())
	}

	err = m.DbClient.InsertLightweightVolume(lightweightVolumeFileset, lightweightVolumeName, name)

	if err != nil {
		return err
	}

	m.log.Printf("Created LightWeight volume at directory path: %s\n",lightweightVolumePath)
	return nil
}

func (m *MMCliFilesetClient) createFileset(filesetName string) error {
	m.log.Println("MMCliFilesetClient: createFileset start")
	defer m.log.Println("MMCliFilesetClient: createFileset end")

	m.log.Printf("creating a new fileset: %s\n", filesetName)

	// create fileset
	spectrumCommand := "/usr/lpp/mmfs/bin/mmcrfileset"
	args := []string{m.Filesystem, filesetName}
	cmd := exec.Command(spectrumCommand, args...)
	output, err := cmd.Output()

	if err != nil {
		return fmt.Errorf("Failed to create fileset %s", filesetName)
	}

	m.log.Printf("Createfileset output: %s\n", string(output))
	return nil
}

func (m *MMCliFilesetClient) isFilesetLinked( filesetName string) (bool,error) {
	m.log.Println("MMCliFilesetClient: isFilesetLinked start")
	defer m.log.Println("MMCliFilesetClient: isFilesetLinked end")

	spectrumCommand := "/usr/lpp/mmfs/bin/mmlsfileset"
	args := []string{m.Filesystem, filesetName, "-Y"}
	cmd := exec.Command(spectrumCommand, args...)
	outputBytes, err := cmd.Output()
	if err != nil {
		return false, err
	}

	spectrumOutput := string(outputBytes)

	lines := strings.Split(spectrumOutput, "\n")

	if len(lines) == 1 {
		return false,fmt.Errorf("Error listing fileset %s", filesetName)
	}

	tokens := strings.Split(lines[1], ":")
	if len(tokens) >= 11 {
		if tokens[10] == "Linked" {
			return true,nil
		} else {
			return false,nil
		}
	}

	return false, fmt.Errorf("Error listing fileset %s after parsing", filesetName)
}

func (m *MMCliFilesetClient) linkFileset(filesetName string) error {
	m.log.Println("MMCliFilesetClient: linkFileset start")
	defer m.log.Println("MMCliFilesetClient: linkFileset end")

	spectrumCommand := "/usr/lpp/mmfs/bin/mmlinkfileset"
	filesetPath := path.Join(m.Mountpoint, filesetName)
	args := []string{m.Filesystem, filesetName, "-J", filesetPath}
	cmd := exec.Command(spectrumCommand, args...)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("Failed to link fileset: %s", err.Error())
	}
	m.log.Printf("MMCliFilesetClient: Linkfileset output: %s\n", string(output))

	//hack for now
	args = []string{"-R", "777", filesetPath}
	cmd = exec.Command("chmod", args...)
	output, err = cmd.Output()
	if err != nil {
		return fmt.Errorf("Failed to set permissions for fileset: %s", err.Error())
	}
	return nil
}

func (m *MMCliFilesetClient) unlinkFileset(filesetName string) error {
	m.log.Println("MMCliFilesetClient: unlinkFileset start")
	defer m.log.Println("MMCliFilesetClient: unlinkFileset end")

	spectrumCommand := "/usr/lpp/mmfs/bin/mmunlinkfileset"
	args := []string{m.Filesystem, filesetName}
	cmd := exec.Command(spectrumCommand, args...)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("Failed to unlink fileset %s: %s", filesetName, err.Error())
	}
	m.log.Printf("MMCliFilesetClient: unLinkfileset output: %s\n", string(output))
	return nil
}

func (m *MMCliFilesetClient) deleteFileset(filesetName string) error {
	m.log.Println("MMCliFilesetClient: deleteFileset start")
	defer m.log.Println("MMCliFilesetClient: deleteFileset end")

	spectrumCommand := "/usr/lpp/mmfs/bin/mmdelfileset"
	args := []string{m.Filesystem, filesetName, "-f"}
	cmd := exec.Command(spectrumCommand, args...)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("Failed to remove fileset %s: %s ", filesetName, err.Error())
	}
	m.log.Printf("MMCliFilesetClient: deleteFileset output: %s\n", string(output))
	return nil
}

func (m *MMCliFilesetClient) verifyFilesetQuota(filesetName, quota string) error {
	m.log.Println("MMCliFilesetClient: verifyFilesetQuota start")
	defer m.log.Println("MMCliFilesetClient: verifyFilesetQuota end")

	spectrumCommand := "/usr/lpp/mmfs/bin/mmlsquota"
	args := []string{"-j", filesetName, m.Filesystem, "--block-size", "auto"}

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

func (m *MMCliFilesetClient) setFilesetQuota(filesetName, quota string) error {
	m.log.Println("MMCliFilesetClient: setFilesetQuota start")
	defer m.log.Println("MMCliFilesetClient: setFilesetQuota end")

	m.log.Printf("setting quota to %s for fileset %s\n", quota, filesetName)

	spectrumCommand := "/usr/lpp/mmfs/bin/mmsetquota"
	args := []string{m.Filesystem + ":" + filesetName, "--block", quota + ":" + quota}
	cmd := exec.Command(spectrumCommand, args...)
	output, err := cmd.Output()

	if err != nil {
		return fmt.Errorf("Failed to set quota for fileset %s", filesetName)
	}

	m.log.Printf("setFilesetQuota output: %s\n", string(output))
	return nil
}

func (m *MMCliFilesetClient) initLightweightVolumes() error {
	m.log.Println("MMCliFilesetClient: InitLightweightVolumes start")
	defer m.log.Println("MMCliFilesetClient: InitLightweightVolumes end")

	isDirFilesetLinked, err := m.isFilesetLinked(m.LightweightVolumeFileset)

	if err != nil {
		if err.Error() == "exit status 2" {

			err := m.createFileset(m.LightweightVolumeFileset)

			if err != nil {
				return fmt.Errorf("Error Initializing Lightweight Volumes : %s", err.Error())
			}
		} else {
			return fmt.Errorf("Error Initializing Lightweight Volumes : %s", err.Error())
		}
	}

	if !isDirFilesetLinked {
		err = m.linkFileset(m.LightweightVolumeFileset)

		if err != nil {
			return fmt.Errorf("Error Initializing Lightweight Volumes : %s", err.Error())
		}
	}

	return nil
}

func (m *MMCliFilesetClient) isLightweightVolumesInitialized() (bool, error) {
	m.log.Println("MMCliFilesetClient: isLightweightVolumesInitialized start")
	defer m.log.Println("MMCliFilesetClient: isLightweightVolumesInitialized end")

	isDirFilesetLinked, err := m.isFilesetLinked(m.LightweightVolumeFileset)

	if err != nil {
		return false, fmt.Errorf("Lightweight volumes not initialized: %s", err.Error())
	}

	if !isDirFilesetLinked {
		return false, fmt.Errorf("Lightweight volumes not initialized: fileset %s not linked", m.LightweightVolumeFileset)
	}
	return true, nil
}

func (m *MMCliFilesetClient) Remove(name string) (err error) {
	m.log.Println("MMCliFilesetClient: remove start")
	defer m.log.Println("MMCliFilesetClient: remove end")

	m.filelock.Lock()
	defer func() {
		lockErr := m.filelock.Unlock()
		if lockErr != nil && err == nil {
			err = lockErr
		}
	}()

	volExists, err := m.DbClient.VolumeExists(name)

	if err != nil {
		m.log.Println(err.Error())
		return err
	}

	if volExists {

		existingVolume, err := m.DbClient.GetVolume(name)

		if err != nil {
			m.log.Println(err.Error())
			return err
		}

		if existingVolume.VolumeType == FILESET ||
			existingVolume.VolumeType == FILESET_QUOTA {

			isFilesetLinked,err := m.isFilesetLinked(existingVolume.Fileset)

			if err != nil {
				m.log.Println(err.Error())
				return err
			}

			if isFilesetLinked {
				err := m.unlinkFileset(existingVolume.Fileset)

				if err != nil {
					m.log.Println(err.Error())
					return err
				}
			}

			err = m.deleteFileset(existingVolume.Fileset)

			if err != nil {
				m.log.Println(err.Error())
				return err
			}

		} else if existingVolume.VolumeType == LIGHTWEIGHT {

			lightweightVolumePath := path.Join(m.Mountpoint, existingVolume.Fileset, existingVolume.Directory)

			err := os.RemoveAll(lightweightVolumePath)

			if err != nil {
				m.log.Println(err.Error())
				return err
			}
		}

		err = m.DbClient.DeleteVolume(name)

		if err != nil {
			m.log.Println(err.Error())
			return err
		}
		return nil
	}
	return errors.New("Volume not found")
}

func (m *MMCliFilesetClient) RemoveWithoutDeletingVolume(name string) error {
	m.log.Println("MMCliFilesetClient: RemoveWithoutDeletingVolume start")
	defer m.log.Println("MMCliFilesetClient: RemoveWithoutDeletingVolume end")
	err := m.DbClient.DeleteVolume(name)
	if err != nil {
		m.log.Printf("error retrieving mapping %#v", err)
		return err
	}
	return nil
}

func (m *MMCliFilesetClient) Attach(name string) (Mountpoint string, err error) {
	m.log.Println("MMCliFilesetClient: attach start")
	defer m.log.Println("MMCliFilesetClient: attach end")

	m.filelock.Lock()
	defer func() {
		lockErr := m.filelock.Unlock()
		if lockErr != nil && err == nil {
			err = lockErr
		}
	}()

	volExists, err := m.DbClient.VolumeExists(name)

	if err != nil {
		m.log.Println(err.Error())
		return "", err
	}

	if !volExists {
		return "", errors.New("Volume not found")
	}

	existingVolume, err := m.DbClient.GetVolume(name)

	if err != nil {
		m.log.Println(err.Error())
		return "", err
	}

	if existingVolume.Mountpoint != "" {
		Mountpoint = existingVolume.Mountpoint
		return Mountpoint, nil
	}

	var mountPath string
	if existingVolume.VolumeType == FILESET ||
		existingVolume.VolumeType == FILESET_QUOTA {

		isFilesetLinked, err := m.isFilesetLinked(existingVolume.Fileset)

		if err != nil {
			m.log.Println(err.Error())
			return "",err
		}

		if !isFilesetLinked {

			err = m.linkFileset(existingVolume.Fileset)

			if err != nil {
				m.log.Println(err.Error())
				return "", err
			}
		}

		mountPath = path.Join(m.Mountpoint, existingVolume.Fileset)
	} else if existingVolume.VolumeType == LIGHTWEIGHT {
		mountPath = path.Join(m.Mountpoint, existingVolume.Fileset, existingVolume.Directory)
	}

	err = m.DbClient.UpdateVolumeMountpoint(name, mountPath)

	if err != nil {
		m.log.Println(err.Error())
		return "", fmt.Errorf("internal error updating database")
	}

	Mountpoint = mountPath
	return Mountpoint, nil
}

func (m *MMCliFilesetClient) Detach(name string) (err error) {
	m.log.Println("MMCliFilesetClient: detach start")
	defer m.log.Println("MMCliFilesetClient: detach end")

	m.filelock.Lock()
	defer func() {
		lockErr := m.filelock.Unlock()
		if lockErr != nil && err == nil {
			err = lockErr
		}
	}()

	volExists, err := m.DbClient.VolumeExists(name)

	if err != nil {
		m.log.Println(err.Error())
		return err
	}

	if !volExists {
		return errors.New("Volume not found")
	}

	existingVolume, err := m.DbClient.GetVolume(name)

	if err != nil {
		m.log.Println(err.Error())
		return err
	}

	if existingVolume.Mountpoint == "" {
		return fmt.Errorf("volume not attached")
	}

	err = m.DbClient.UpdateVolumeMountpoint(name, "")

	if err != nil {
		m.log.Println(err.Error())
		return fmt.Errorf("internal error updating database")
	}
	return nil
}

func (m *MMCliFilesetClient) ExportNfs(name string, clientCIDR string) (string, error) {
	m.log.Println("MMCliFilesetClient: ExportNfs start")
	defer m.log.Println("MMCliFilesetClient: ExportNfs end")

	volExists, err := m.DbClient.VolumeExists(name)

	if err != nil {
		m.log.Println(err.Error())
		return "", err
	}

	if !volExists {
		m.log.Println("MMCliFilesetClient ExportNfs: fileset not found")
		return "", fmt.Errorf("fileset couldn't be located")
	}

	existingVolume, err := m.DbClient.GetVolume(name)

	if err != nil {
		m.log.Println(err.Error())
		return "", err
	}

	if existingVolume.Mountpoint == "" {
		m.log.Println("MMCliFilesetClient ExportNfs: fileset not linked")
		return "", fmt.Errorf("fileset not linked")
	}

	spectrumCommand := "/usr/lpp/mmfs/bin/mmnfs"
	filesetPath := path.Join(m.Mountpoint, existingVolume.Fileset)
	args := []string{"export", "add", filesetPath, "--client", fmt.Sprintf("%s(Access_Type=RW,Protocols=3:4,Squash=no_root_squash)", clientCIDR)}
	cmd := exec.Command(spectrumCommand, args...)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("Failed to export fileset via NFS: %s", err.Error())
	}
	m.log.Printf("MMCliFilesetClient: ExportNfs output: %s\n", string(output))
	return filesetPath, nil
}

func (m *MMCliFilesetClient) UnexportNfs(name string) error {
	m.log.Println("MMCliFilesetClient: UnexportNfs start")
	defer m.log.Println("MMCliFilesetClient: UnexportNfs end")

	volExists, err := m.DbClient.VolumeExists(name)

	if err != nil {
		m.log.Println(err.Error())
		return err
	}

	if !volExists {
		m.log.Println("MMCliFilesetClient UnexportNfs: fileset not found")
		return fmt.Errorf("fileset couldn't be located")
	}

	existingVolume, err := m.DbClient.GetVolume(name)

	if err != nil {
		m.log.Println(err.Error())
		return err
	}

	spectrumCommand := "/usr/lpp/mmfs/bin/mmnfs"
	filesetPath := path.Join(m.Mountpoint, existingVolume.Fileset)
	args := []string{"export", "remove", filesetPath, "--force"}
	cmd := exec.Command(spectrumCommand, args...)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("Failed to unexport fileset via NFS: %s", err.Error())
	}
	m.log.Printf("MMCliFilesetClient: UnexportNfs output: %s\n", string(output))

	if err := m.DbClient.UpdateVolumeMountpoint(name, ""); err != nil {
		m.log.Printf("MMCliFilesetClient UnexportNfs: Could not update volume mountpoint: %s", err)
		return err
	}
	return nil
}

func (m *MMCliFilesetClient) List() (volumeList []models.VolumeMetadata, err error) {
	m.log.Println("MMCliFilesetClient: list start")
	defer m.log.Println("MMCliFilesetClient: list end")

	m.filelock.Lock()
	defer func() {
		lockErr := m.filelock.Unlock()
		if lockErr != nil && err == nil {
			err = lockErr
		}
	}()

	volumesInDb, err := m.DbClient.ListVolumes()

	if err != nil {
		m.log.Println(err.Error())
		return nil, err
	}

	var volumes []models.VolumeMetadata
	for _, volume := range volumesInDb {
		volumes = append(volumes, models.VolumeMetadata{Name: volume.VolumeName, Mountpoint: volume.Mountpoint})
	}
	volumeList = volumes
	return volumeList, nil
}

func (m *MMCliFilesetClient) Get(name string) (volumeMetadata *models.VolumeMetadata, volumeConfigDetails *models.SpectrumConfig, err error) {
	m.log.Println("MMCliFilesetClient: get start")
	defer m.log.Println("MMCliFilesetClient: get finish")

	m.filelock.Lock()
	defer func() {
		lockErr := m.filelock.Unlock()
		if lockErr != nil && err == nil {
			err = lockErr
		}
	}()

	volExists, err := m.DbClient.VolumeExists(name)

	if err != nil {
		m.log.Println(err.Error())
		return nil, nil, err
	}

	if volExists {

		existingVolume, err := m.DbClient.GetVolume(name)

		if err != nil {
			m.log.Println(err.Error())
			return nil, nil, err
		}

		volumeMetadata = &models.VolumeMetadata{Name: existingVolume.VolumeName, Mountpoint: existingVolume.Mountpoint}
		volumeConfigDetails = &models.SpectrumConfig{FilesetId: existingVolume.Fileset, Filesystem: m.Filesystem}
		return volumeMetadata, volumeConfigDetails, nil
	}
	return nil, nil, errors.New("Volume not found")
}

func (m *MMCliFilesetClient) isSpectrumScaleMounted() (isMounted bool, err error) {
	m.log.Println("MMCliFilesetClient: isMounted start")
	defer m.log.Println("MMCliFilesetClient: isMounted end")

	if m.isMounted == true {
		isMounted = true
		return isMounted, nil
	}

	spectrumCommand := "/usr/lpp/mmfs/bin/mmlsmount"
	args := []string{m.Filesystem, "-L", "-Y"}
	cmd := exec.Command(spectrumCommand, args...)
	outputBytes, err := cmd.Output()
	if err != nil {
		m.log.Printf("Error running command\n")
		m.log.Println(err)
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
		m.log.Printf("MMCliFilesetClient: node name: %s\n", currentNode)
		for _, node := range mountedNodes {
			if node == currentNode {
				m.isMounted = true
				isMounted = true
				return isMounted, nil
			}
		}
	}
	isMounted = false
	return isMounted, nil
}

func (m *MMCliFilesetClient) mount() (err error) {
	m.log.Println("MMCliFilesetClient: mount start")
	defer m.log.Println("MMCliFilesetClient: mount end")

	if m.isMounted == true {
		return nil
	}

	spectrumCommand := "/usr/lpp/mmfs/bin/mmmount"
	args := []string{m.Filesystem, m.Mountpoint}
	cmd := exec.Command(spectrumCommand, args...)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("Failed to mount filesystem")
	}
	m.log.Println(output)
	m.isMounted = true
	return nil
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

func (m *MMCliFilesetClient) GetFileSetForMountPoint(mountPoint string) (string, error) {

	volume, err := m.DbClient.GetVolumeForMountPoint(mountPoint)

	if err != nil {
		m.log.Println(err.Error())
		return "", err
	}
	return volume, nil
}

func generateFilesetName() string {
	return strconv.FormatInt(time.Now().UnixNano(), 10)
}

func generateLightweightVolumeName() string {
	return "LtwtVol" + strconv.FormatInt(time.Now().UnixNano(), 10)
}
