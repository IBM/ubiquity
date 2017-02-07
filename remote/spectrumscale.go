package remote

import (
	"fmt"
	"log"

	"net/http"

	"github.ibm.com/almaden-containers/ubiquity/resources"
	"github.ibm.com/almaden-containers/ubiquity/utils"
)

type spectrumRemoteClient struct {
	logger        *log.Logger
	isActivated   bool
	isMounted     bool
	httpClient    *http.Client
	storageApiURL string
	backendName   string
}

func NewSpectrumRemoteClient(logger *log.Logger, backendName, storageApiURL string) (resources.StorageClient, error) {
	return &spectrumRemoteClient{logger: logger, storageApiURL: storageApiURL, httpClient: &http.Client{}, backendName: backendName}, nil
}

func (s *spectrumRemoteClient) Activate() (err error) {
	s.logger.Println("spectrumRemoteClient: Activate start")
	defer s.logger.Println("spectrumRemoteClient: Activate end")

	if s.isActivated {
		return nil
	}

	// call remote activate
	activateURL := utils.FormatURL(s.storageApiURL, s.backendName, "activate")
	response, err := utils.HttpExecute(s.httpClient, s.logger, "POST", activateURL, nil)
	if err != nil {
		s.logger.Printf("Error in activate remote call %#v", err)
		return fmt.Errorf("Error in activate remote call")
	}

	if response.StatusCode != http.StatusOK {
		s.logger.Printf("Error in activate remote call %#v\n", response)
		return utils.ExtractErrorResponse(response)
	}
	s.logger.Println("spectrumRemoteClient: Activate success")
	s.isActivated = true
	return nil
}

func (s *spectrumRemoteClient) CreateVolume(name string, opts map[string]interface{}) (err error) {
	s.logger.Println("spectrumRemoteClient: create start")
	defer s.logger.Println("spectrumRemoteClient: create end")

	createRemoteURL := utils.FormatURL(s.storageApiURL, s.backendName, "volumes")
	createVolumeRequest := resources.CreateRequest{Name: name, Opts: opts}

	response, err := utils.HttpExecute(s.httpClient, s.logger, "POST", createRemoteURL, createVolumeRequest)
	if err != nil {
		s.logger.Printf("Error in create volume remote call %s", err.Error())
		return fmt.Errorf("Error in create volume remote call(http error)")
	}

	if response.StatusCode != http.StatusOK {
		s.logger.Printf("Error in create volume remote call %#v", response)
		return utils.ExtractErrorResponse(response)
	}

	return nil
}

func (s *spectrumRemoteClient) RemoveVolume(name string, forceDelete bool) (err error) {
	s.logger.Println("spectrumRemoteClient: remove start")
	defer s.logger.Println("spectrumRemoteClient: remove end")

	removeRemoteURL := utils.FormatURL(s.storageApiURL, s.backendName, "volumes", name)
	removeRequest := resources.RemoveRequest{Name: name, ForceDelete: forceDelete}

	response, err := utils.HttpExecute(s.httpClient, s.logger, "DELETE", removeRemoteURL, removeRequest)
	if err != nil {
		s.logger.Printf("Error in remove volume remote call %#v", err)
		return fmt.Errorf("Error in remove volume remote call")
	}

	if response.StatusCode != http.StatusOK {
		s.logger.Printf("Error in remove volume remote call %#v", response)
		return utils.ExtractErrorResponse(response)
	}

	return nil
}

func (s *spectrumRemoteClient) GetVolume(name string) (resources.VolumeMetadata, map[string]interface{}, error) {
	s.logger.Println("spectrumRemoteClient: get start")
	defer s.logger.Println("spectrumRemoteClient: get finish")

	getRemoteURL := utils.FormatURL(s.storageApiURL, s.backendName, "volumes", name)
	response, err := utils.HttpExecute(s.httpClient, s.logger, "GET", getRemoteURL, nil)
	if err != nil {
		s.logger.Printf("Error in get volume remote call %#v", err)
		return resources.VolumeMetadata{}, nil, fmt.Errorf("Error in get volume remote call")
	}

	if response.StatusCode != http.StatusOK {
		s.logger.Printf("Error in get volume remote call %#v", response)
		return resources.VolumeMetadata{}, nil, utils.ExtractErrorResponse(response)
	}

	getResponse := resources.GetResponse{}
	err = utils.UnmarshalResponse(response, &getResponse)
	if err != nil {
		s.logger.Printf("Error in unmarshalling response for get remote call %#v for response %#v", err, response)
		return resources.VolumeMetadata{}, nil, fmt.Errorf("Error in unmarshalling response for get remote call")
	}

	return getResponse.Volume, getResponse.Config, nil
}

func (s *spectrumRemoteClient) Attach(name string) (string, error) {
	s.logger.Println("spectrumRemoteClient: attach start")
	defer s.logger.Println("spectrumRemoteClient: attach end")

	attachRemoteURL := utils.FormatURL(s.storageApiURL, s.backendName, "volumes", name, "attach")
	response, err := utils.HttpExecute(s.httpClient, s.logger, "PUT", attachRemoteURL, nil)
	if err != nil {
		s.logger.Printf("Error in attach volume remote call %#v", err)
		return "", fmt.Errorf("Error in attach volume remote call")
	}

	if response.StatusCode != http.StatusOK {
		s.logger.Printf("Error in attach volume remote call %#v", response)

		return "", utils.ExtractErrorResponse(response)
	}

	mountResponse := resources.MountResponse{}
	err = utils.UnmarshalResponse(response, &mountResponse)
	if err != nil {
		return "", fmt.Errorf("Error in unmarshalling response for attach remote call")
	}
	return mountResponse.Mountpoint, nil
}

func (s *spectrumRemoteClient) Detach(name string) error {
	s.logger.Println("spectrumRemoteClient: detach start")
	defer s.logger.Println("spectrumRemoteClient: detach end")

	detachRemoteURL := utils.FormatURL(s.storageApiURL, s.backendName, "volumes", name, "detach")
	response, err := utils.HttpExecute(s.httpClient, s.logger, "PUT", detachRemoteURL, nil)
	if err != nil {
		s.logger.Printf("Error in detach volume remote call %#v", err)
		return fmt.Errorf("Error in detach volume remote call")
	}

	if response.StatusCode != http.StatusOK {
		s.logger.Printf("Error in detach volume remote call %#v", response)
		return utils.ExtractErrorResponse(response)
	}

	return nil

}

func (s *spectrumRemoteClient) ListVolumes() ([]resources.VolumeMetadata, error) {
	s.logger.Println("spectrumRemoteClient: list start")
	defer s.logger.Println("spectrumRemoteClient: list end")

	listRemoteURL := utils.FormatURL(s.storageApiURL, s.backendName, "volumes")
	response, err := utils.HttpExecute(s.httpClient, s.logger, "GET", listRemoteURL, nil)
	if err != nil {
		s.logger.Printf("Error in list volume remote call %#v", err)
		return nil, fmt.Errorf("Error in list volume remote call")
	}

	if response.StatusCode != http.StatusOK {
		s.logger.Printf("Error in list volume remote call %#v", err)
		return nil, utils.ExtractErrorResponse(response)
	}

	listResponse := resources.ListResponse{}
	err = utils.UnmarshalResponse(response, &listResponse)
	if err != nil {
		s.logger.Printf("Error in unmarshalling response for get remote call %#v for response %#v", err, response)
		return []resources.VolumeMetadata{}, nil
	}

	return listResponse.Volumes, nil

}
