package core

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"

	"github.ibm.com/almaden-containers/spectrum-common.git/models"
)

//go:generate counterfeiter -o ../fakes/fake_spectrum_client.go . SpectrumClient

type SpectrumClient interface {
	Create(name string, opts map[string]interface{}) error
	Remove(name string) error
	Attach(name string) (string, error)
	Detach(name string) error
	List() ([]models.VolumeMetadata, error)
	Get(name string) (*models.VolumeMetadata, *models.SpectrumConfig, error)
	IsMounted() (bool, error)
	Mount() error
}

type Fileset struct {
	Name             string
	Mountpoint       string
	DockerVolumeName string
}

type MappingConfig struct {
	Mappings map[string]Fileset
}

func NewSpectrumClient(logger *log.Logger, filesystem, mountpoint string) SpectrumClient {
	return &MMCliFilesetClient{log: logger, Filesystem: filesystem, Mountpoint: mountpoint}
}

type MMCliFilesetClient struct {
	Filesystem string
	Mountpoint string
	log        *log.Logger
}

func (m *MMCliFilesetClient) Create(name string, opts map[string]interface{}) error {
	m.log.Println("MMCliFilesetClient: create start")
	defer m.log.Println("MMCliFilesetClient: create end")
	mappingConfig, err := m.retrieveMappingConfig()
	if err != nil {
		return err
	}
	_, ok := mappingConfig.Mappings[name]
	if ok == true {
		return fmt.Errorf("Volume already exists")
	}
	userSpecifiedFileset, exists := opts["fileset"]
	if exists == true {
		return m.updateMappingWithExistingFileset(name, userSpecifiedFileset.(string), mappingConfig)
	} else {
		return m.create(name, opts, mappingConfig)
	}

}
func (m *MMCliFilesetClient) updateMappingWithExistingFileset(name, userSpecifiedFileset string, mappingConfig MappingConfig) error {
	m.log.Println("MMCliFilesetClient:  updateMappingWithExistingFileset start")
	defer m.log.Println("MMCliFilesetClient: updateMappingWithExistingFileset end")
	m.log.Printf("User specified fileset: %s\n", userSpecifiedFileset)

	spectrumCommand := "/usr/lpp/mmfs/bin/mmlsfileset"
	args := []string{m.Filesystem, userSpecifiedFileset, "-Y"}
	cmd := exec.Command(spectrumCommand, args...)
	_, err := cmd.Output()
	if err != nil {
		m.log.Println(err)
		return err
	}
	mappingConfig.Mappings[name] = Fileset{Name: userSpecifiedFileset, DockerVolumeName: name}
	// persist mapping config
	err = m.persistMappingConfig(mappingConfig)
	if err != nil {
		return err
	}
	return nil
}

func (m *MMCliFilesetClient) create(name string, opts map[string]interface{}, mappingConfig MappingConfig) error {
	m.log.Println("MMCliFilesetClient: createNew start")
	defer m.log.Println("MMCliFilesetClient: createNew end")

	filesetName := generateFilesetName()
	m.log.Printf("creating a new fileset: %s\n", filesetName)
	// create fileset
	spectrumCommand := "/usr/lpp/mmfs/bin/mmcrfileset"
	args := []string{m.Filesystem, filesetName}
	cmd := exec.Command(spectrumCommand, args...)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("Failed to create fileset")
	}
	m.log.Printf("Createfileset output: %s\n", string(output))
	// add it to mapping config
	mappingConfig.Mappings[name] = Fileset{Name: filesetName, DockerVolumeName: name}
	// persist mapping config
	err = m.persistMappingConfig(mappingConfig)
	if err != nil {
		return err
	}
	return nil
}

func (m *MMCliFilesetClient) Remove(name string) error {
	m.log.Println("MMCliFilesetClient: remove start")
	defer m.log.Println("MMCliFilesetClient: remove end")
	mappingConfig, err := m.retrieveMappingConfig()
	if err != nil {
		return err
	}
	existingMapping, ok := mappingConfig.Mappings[name]
	if ok == true {
		spectrumCommand := "/usr/lpp/mmfs/bin/mmdelfileset"
		args := []string{m.Filesystem, existingMapping.Name, "-f"}
		cmd := exec.Command(spectrumCommand, args...)
		output, err := cmd.Output()
		if err != nil {
			return fmt.Errorf("Failed to remove fileset")
		}
		m.log.Printf("MMCliFilesetClient: Deletefileset output: %s\n", string(output))
		delete(mappingConfig.Mappings, name)
		err = m.persistMappingConfig(mappingConfig)
		if err != nil {
			return err
		}
	}
	return nil
}
func (m *MMCliFilesetClient) Attach(name string) (string, error) {
	m.log.Println("MMCliFilesetClient: attach start")
	defer m.log.Println("MMCliFilesetClient: attach end")
	mappingConfig, err := m.retrieveMappingConfig()
	if err != nil {
		return "", err
	}
	mapping, ok := mappingConfig.Mappings[name]
	if ok == false {
		return "", fmt.Errorf("fileset couldn't be located")
	}
	if mapping.Mountpoint != "" {
		//return "", fmt.Errorf("fileset already linked")
		return mapping.Mountpoint, nil
	}
	spectrumCommand := "/usr/lpp/mmfs/bin/mmlinkfileset"
	filesetPath := path.Join(m.Mountpoint, mapping.Name)
	args := []string{m.Filesystem, mapping.Name, "-J", filesetPath}
	cmd := exec.Command(spectrumCommand, args...)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("Failed to link fileset")
	}
	m.log.Printf("MMCliFilesetClient: Linkfileset output: %s\n", string(output))

	//hack for now
	args = []string{"-R", "777", filesetPath}
	cmd = exec.Command("chmod", args...)
	output, err = cmd.Output()
	if err != nil {
		return "", fmt.Errorf("Failed to set permissions for fileset")
	}

	mapping.Mountpoint = filesetPath
	mappingConfig.Mappings[name] = mapping
	err = m.persistMappingConfig(mappingConfig)
	if err != nil {
		return "", fmt.Errorf("internal error updating mapping")
	}

	return filesetPath, nil
}

func (m *MMCliFilesetClient) Detach(name string) error {
	m.log.Println("MMCliFilesetClient: detach start")
	defer m.log.Println("MMCliFilesetClient: detach end")
	mappingConfig, err := m.retrieveMappingConfig()
	if err != nil {
		return err
	}
	mapping, ok := mappingConfig.Mappings[name]
	if ok == false {
		return fmt.Errorf("fileset couldn't be located")
	}
	if mapping.Mountpoint == "" {
		return fmt.Errorf("fileset not linked")
	}
	spectrumCommand := "/usr/lpp/mmfs/bin/mmunlinkfileset"
	args := []string{m.Filesystem, mapping.Name}
	cmd := exec.Command(spectrumCommand, args...)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("Failed to unlink fileset")
	}
	m.log.Printf("MMCliFilesetClient: unLinkfileset output: %s\n", string(output))

	mapping.Mountpoint = ""
	mappingConfig.Mappings[name] = mapping
	err = m.persistMappingConfig(mappingConfig)
	if err != nil {
		return fmt.Errorf("internal error updating mapping")
	}
	return nil
}

func (m *MMCliFilesetClient) List() ([]models.VolumeMetadata, error) {
	m.log.Println("MMCliFilesetClient: list start")
	defer m.log.Println("MMCliFilesetClient: list end")
	mappingConfig, err := m.retrieveMappingConfig()
	if err != nil {
		return nil, err
	}
	var volumes []models.VolumeMetadata
	for _, fileset := range mappingConfig.Mappings {
		volumes = append(volumes, models.VolumeMetadata{Name: fileset.DockerVolumeName, Mountpoint: fileset.Mountpoint})
	}
	return volumes, nil
}

func (m *MMCliFilesetClient) Get(name string) (*models.VolumeMetadata, *models.SpectrumConfig, error) {
	m.log.Println("MMCliFilesetClient: get start")
	defer m.log.Println("MMCliFilesetClient: get finish")
	mappingConfig, err := m.retrieveMappingConfig()
	if err != nil {
		return nil, nil, err
	}
	fileset, ok := mappingConfig.Mappings[name]
	if ok == true {
		volumeMetadata := &models.VolumeMetadata{Name: fileset.DockerVolumeName, Mountpoint: fileset.Mountpoint}
		volumeConfigDetails := &models.SpectrumConfig{FilesetId: fileset.Name, Filesystem: m.Filesystem}
		return volumeMetadata, volumeConfigDetails, nil
	}
	return nil, nil, fmt.Errorf("Cannot find info")
}
func (m *MMCliFilesetClient) IsMounted() (bool, error) {
	m.log.Println("MMCliFilesetClient: isMounted start")
	defer m.log.Println("MMCliFilesetClient: isMounted end")
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
		return false, nil
	} else {
		// checkif mounted on current node -- compare node name
		currentNode, _ := os.Hostname()
		m.log.Printf("MMCliFilesetClient: node name: %s\n", currentNode)
		for _, node := range mountedNodes {
			if node == currentNode {
				return true, nil
			}
		}
	}
	return false, nil
}
func (m *MMCliFilesetClient) Mount() error {
	m.log.Println("MMCliFilesetClient: mount start")
	defer m.log.Println("MMCliFilesetClient: mount end")
	spectrumCommand := "/usr/lpp/mmfs/bin/mmmount"
	args := []string{m.Filesystem, m.Mountpoint}
	cmd := exec.Command(spectrumCommand, args...)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("Failed to mount filesystem")
	}
	m.log.Println(output)
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

func (m *MMCliFilesetClient) retrieveMappingConfig() (MappingConfig, error) {
	//m.log.Println("MMCliFilesetClient: retrieveMappingConfig start")
	//defer m.log.Println("MMCliFilesetClient: retrieveMappingConfig end")
	mappingFile, err := os.Open(path.Join(m.Mountpoint, ".docker.json"))
	if err != nil {
		m.log.Println(err.Error())
		if os.IsNotExist(err) == true {
			m.log.Println("file does not exist")
			mappingConfig := MappingConfig{Mappings: map[string]Fileset{}}
			err = m.persistMappingConfig(mappingConfig)
			if err != nil {
				return MappingConfig{}, fmt.Errorf("error initializing config file (%s)", err.Error())
			}
			return mappingConfig, nil
		} else {
			return MappingConfig{}, fmt.Errorf("error opening config file (%s)", err.Error())
		}
	}
	jsonParser := json.NewDecoder(mappingFile)
	var mappingConfig MappingConfig
	if err = jsonParser.Decode(&mappingConfig); err != nil {
		return MappingConfig{}, fmt.Errorf("error parsing config file (%s)", err.Error())
	}
	return mappingConfig, nil
}
func (m *MMCliFilesetClient) persistMappingConfig(mappingConfig MappingConfig) error {
	//m.log.Println("MMCliFilesetClient: persisteMappingConfig start")
	//defer m.log.Println("MMCliFilesetClient: persisteMappingConfig end")
	data, err := json.Marshal(&mappingConfig)
	if err != nil {
		return fmt.Errorf("Error marshalling mapping config to file: %s", err.Error())
	}
	err = ioutil.WriteFile(path.Join(m.Mountpoint, ".docker.json"), data, 0644)
	if err != nil {
		return fmt.Errorf("Error writing json spec: %s", err.Error())
	}
	return nil
}
func generateFilesetName() string {
	return strconv.FormatInt(time.Now().UnixNano(), 10)
}
