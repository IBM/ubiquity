package remote

import (
	"fmt"
	"log"
	"strings"

	"os"
	"os/exec"

	"net/http"

	"encoding/json"

	"bytes"

	"github.ibm.com/almaden-containers/ubiquity.git/model"
	"github.ibm.com/almaden-containers/ubiquity.git/utils"
)

type spectrumRemoteClient struct {
	logger        *log.Logger
	filesystem    string
	mountpoint    string
	isActivated   bool
	isMounted     bool
	httpClient    *http.Client
	storageApiURL string
}

func NewSpectrumRemoteClient(logger *log.Logger, filesystem, mountpoint string, storageApiURL string) model.StorageClient {
	return &spectrumRemoteClient{logger: logger, filesystem: filesystem, mountpoint: mountpoint, storageApiURL: storageApiURL, httpClient: &http.Client{}}
}
func (s *spectrumRemoteClient) Activate() (err error) {
	s.logger.Println("spectrumRemoteClient: Activate start")
	defer s.logger.Println("spectrumRemoteClient: Activate end")

	if s.isActivated {
		return nil
	}

	//check if filesystem is mounted
	mounted, err := s.isSpectrumScaleMounted()

	if err != nil {
		s.logger.Printf("Internal error on activate %#v", err)
		return err
	}

	if mounted == false {
		err = s.mount()

		if err != nil {
			s.logger.Printf("Internal error on activate %#v", err)
			return err
		}
	}

	// call remote activate
	activateURL := utils.FormatURL(s.storageApiURL, s.GetPluginName(), "activate")
	response, err := s.httpExecute("POST", activateURL, nil)
	if err != nil {
		s.logger.Printf("Error in activate remote call %#v", err)
		return fmt.Errorf("Error in activate remote call")
	}

	if response.StatusCode != http.StatusOK {
		s.logger.Printf("Error in activate remote call %#v\n", response)
		return fmt.Errorf("Error in activate remote call")
	}
	s.logger.Println("spectrumRemoteClient: Activate success")
	s.isActivated = true
	return nil
}

func (s *spectrumRemoteClient) GetPluginName() string {
	return "spectrum-scale"
}

func (s *spectrumRemoteClient) CreateVolume(name string, opts map[string]interface{}) (err error) {
	s.logger.Println("spectrumRemoteClient: create start")
	defer s.logger.Println("spectrumRemoteClient: create end")

	createRemoteURL := utils.FormatURL(s.storageApiURL, s.GetPluginName(), "volumes")
	createVolumeRequest := model.CreateRequest{Name: name, Opts: opts}

	response, err := s.httpExecute("POST", createRemoteURL, createVolumeRequest)
	if err != nil {
		s.logger.Printf("Error in create volume remote call %#v", err)
		return fmt.Errorf("Error in create volume remote call")
	}

	if response.StatusCode != http.StatusOK {
		s.logger.Printf("Error in create volume remote call %#v", response)
		return fmt.Errorf("Error in create volume remote call")
	}

	return nil
}

func (s *spectrumRemoteClient) httpExecute(requestType string, requestURL string, rawPayload interface{}) (*http.Response, error) {
	payload, err := json.MarshalIndent(rawPayload, "", " ")
	if err != nil {
		s.logger.Printf("Internal error marshalling params %#v", err)
		return nil, fmt.Errorf("Internal error marshalling params")
	}

	request, err := http.NewRequest(requestType, requestURL, bytes.NewBuffer(payload))
	if err != nil {
		s.logger.Printf("Error in creating request %#v", err)
		return nil, fmt.Errorf("Error in creating request")
	}

	return s.httpClient.Do(request)
}

func (s *spectrumRemoteClient) RemoveVolume(name string, forceDelete bool) (err error) {
	s.logger.Println("spectrumRemoteClient: remove start")
	defer s.logger.Println("spectrumRemoteClient: remove end")

	removeRemoteURL := utils.FormatURL(s.storageApiURL, s.GetPluginName(), "volumes", name)
	removeRequest := model.RemoveRequest{Name: name, ForceDelete: forceDelete}

	response, err := s.httpExecute("DELETE", removeRemoteURL, removeRequest)
	if err != nil {
		s.logger.Printf("Error in remove volume remote call %#v", err)
		return fmt.Errorf("Error in remove volume remote call")
	}

	if response.StatusCode != http.StatusOK {
		s.logger.Printf("Error in remove volume remote call %#v", response)
		return fmt.Errorf("Error in remove volume remote call")
	}

	return nil
}

//GetVolume(string) (*model.VolumeMetadata, *string, *map[string]interface {}, error)
func (s *spectrumRemoteClient) GetVolume(name string) (model.VolumeMetadata, model.SpectrumConfig, error) {
	s.logger.Println("spectrumRemoteClient: get start")
	defer s.logger.Println("spectrumRemoteClient: get finish")

	getRemoteURL := utils.FormatURL(s.storageApiURL, s.GetPluginName(), "volumes", name)
	response, err := s.httpExecute("GET", getRemoteURL, nil)
	if err != nil {
		s.logger.Printf("Error in get volume remote call %#v", err)
		return model.VolumeMetadata{}, model.SpectrumConfig{}, fmt.Errorf("Error in get volume remote call")
	}

	if response.StatusCode != http.StatusOK {
		s.logger.Printf("Error in get volume remote call %#v", response)
		return model.VolumeMetadata{}, model.SpectrumConfig{}, fmt.Errorf("Error in get volume remote call")
	}

	getResponse := model.GetResponse{}
	err = utils.UnmarshalResponse(response, &getResponse)
	if err != nil {
		s.logger.Printf("Error in unmarshalling response for get remote call %#v for response %#v", err, response)
		return model.VolumeMetadata{}, model.SpectrumConfig{}, fmt.Errorf("Error in unmarshalling response for get remote call")
	}

	return getResponse.Volume, getResponse.Config, nil
}

func (s *spectrumRemoteClient) Attach(name string) (string, error) {
	s.logger.Println("spectrumRemoteClient: attach start")
	defer s.logger.Println("spectrumRemoteClient: attach end")

	attachRemoteURL := utils.FormatURL(s.storageApiURL, s.GetPluginName(), "volumes", name, "attach")
	response, err := s.httpExecute("PUT", attachRemoteURL, nil)
	if err != nil {
		s.logger.Printf("Error in attach volume remote call %#v", err)
		return "", fmt.Errorf("Error in attach volume remote call")
	}

	if response.StatusCode != http.StatusOK {
		s.logger.Printf("Error in attach volume remote call %#v", response)
		return "", fmt.Errorf("Error in attach volume remote call")
	}

	mountResponse := model.MountResponse{}
	err = utils.UnmarshalResponse(response, &mountResponse)
	if err != nil {
		return "", fmt.Errorf("Error in unmarshalling response for attach remote call")
	}
	return mountResponse.Mountpoint, nil
}

func (s *spectrumRemoteClient) Detach(name string) error {
	s.logger.Println("spectrumRemoteClient: detach start")
	defer s.logger.Println("spectrumRemoteClient: detach end")

	detachRemoteURL := utils.FormatURL(s.storageApiURL, s.GetPluginName(), "volumes", name, "detach")
	response, err := s.httpExecute("PUT", detachRemoteURL, nil)
	if err != nil {
		s.logger.Printf("Error in detach volume remote call %#v", err)
		return fmt.Errorf("Error in detach volume remote call")
	}

	if response.StatusCode != http.StatusOK {
		s.logger.Printf("Error in detach volume remote call %#v", response)
		return fmt.Errorf("Error in detach volume remote call")
	}

	return nil

}

func (s *spectrumRemoteClient) ListVolumes() ([]model.VolumeMetadata, error) {
	s.logger.Println("spectrumRemoteClient: list start")
	defer s.logger.Println("spectrumRemoteClient: list end")

	listRemoteURL := utils.FormatURL(s.storageApiURL, s.GetPluginName(), "volumes")
	response, err := s.httpExecute("GET", listRemoteURL, nil)
	if err != nil {
		s.logger.Printf("Error in list volume remote call %#v", err)
		return nil, fmt.Errorf("Error in list volume remote call")
	}

	if response.StatusCode != http.StatusOK {
		s.logger.Printf("Error in list volume remote call %#v", err)
		return nil, fmt.Errorf("Error in list volume remote call")
	}

	listResponse := model.ListResponse{}
	err = utils.UnmarshalResponse(response, &listResponse)
	if err != nil {
		s.logger.Printf("Error in unmarshalling response for get remote call %#v for response %#v", err, response)
		return []model.VolumeMetadata{}, nil
	}

	return listResponse.Volumes, nil

}

func (s *spectrumRemoteClient) mount() error {
	s.logger.Println("spectrumRemoteClient: mount start")
	defer s.logger.Println("spectrumRemoteClient: mount end")

	if s.isMounted == true {
		return nil
	}

	spectrumCommand := "/usr/lpp/mmfs/bin/mmmount"
	args := []string{s.filesystem, s.mountpoint}
	cmd := exec.Command(spectrumCommand, args...)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("Failed to mount filesystem")
	}
	s.logger.Println(output)
	s.isMounted = true
	return nil
}

func (s *spectrumRemoteClient) isSpectrumScaleMounted() (bool, error) {
	s.logger.Println("spectrumRemoteClient: isMounted start")
	defer s.logger.Println("spectrumRemoteClient: isMounted end")

	if s.isMounted == true {
		return s.isMounted, nil
	}

	spectrumCommand := "/usr/lpp/mmfs/bin/mmlsmount"
	args := []string{s.filesystem, "-L", "-Y"}
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
		s.isMounted = false
		return s.isMounted, nil
	} else {
		// checkif mounted on current node -- compare node name
		currentNode, _ := os.Hostname()
		s.logger.Printf("spectrumRemoteClient: node name: %s\n", currentNode)
		for _, node := range mountedNodes {
			if node == currentNode {
				s.isMounted = true
				return s.isMounted, nil
			}
		}
	}
	s.isMounted = false
	return s.isMounted, nil
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
