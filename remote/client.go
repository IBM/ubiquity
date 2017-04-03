package remote

import (
	"fmt"
	"log"

	"net/http"

	"github.com/ibm/ubiquity/resources"

	"reflect"

	"github.com/ibm/ubiquity/remote/mounter"
	"github.com/ibm/ubiquity/utils"
)

type remoteClient struct {
	logger        *log.Logger
	isActivated   bool
	isMounted     bool
	httpClient    *http.Client
	storageApiURL string
	config        resources.UbiquityPluginConfig
}

func NewRemoteClient(logger *log.Logger, storageApiURL string, config resources.UbiquityPluginConfig) (resources.StorageClient, error) {
	return &remoteClient{logger: logger, storageApiURL: storageApiURL, httpClient: &http.Client{}, config: config}, nil
}

func (s *remoteClient) Activate() (err error) {
	s.logger.Println("remoteClient: Activate start")
	defer s.logger.Println("remoteClient: Activate end")

	if s.isActivated {
		return nil
	}

	// call remote activate
	activateURL := utils.FormatURL(s.storageApiURL, "activate")
	response, err := utils.HttpExecute(s.httpClient, s.logger, "POST", activateURL, nil)
	if err != nil {
		s.logger.Printf("Error in activate remote call %#v", err)
		return fmt.Errorf("Error in activate remote call")
	}

	if response.StatusCode != http.StatusOK {
		s.logger.Printf("Error in activate remote call %#v\n", response)
		return utils.ExtractErrorResponse(response)
	}
	s.logger.Println("remoteClient: Activate success")
	s.isActivated = true
	return nil
}

func (s *remoteClient) CreateVolume(name string, opts map[string]interface{}) (err error) {
	s.logger.Println("remoteClient: create start")
	defer s.logger.Println("remoteClient: create end")

	createRemoteURL := utils.FormatURL(s.storageApiURL, "volumes")

	if reflect.DeepEqual(s.config.SpectrumNfsRemoteConfig, resources.SpectrumNfsRemoteConfig{}) == false {
		opts["nfsClientConfig"] = s.config.SpectrumNfsRemoteConfig.ClientConfig
	}
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

func (s *remoteClient) RemoveVolume(name string) (err error) {
	s.logger.Println("remoteClient: remove start")
	defer s.logger.Println("remoteClient: remove end")

	removeRemoteURL := utils.FormatURL(s.storageApiURL, "volumes", name)
	removeRequest := resources.RemoveRequest{Name: name}

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

func (s *remoteClient) GetVolume(name string) (resources.Volume, error) {
	s.logger.Println("remoteClient: get start")
	defer s.logger.Println("remoteClient: get finish")

	getRemoteURL := utils.FormatURL(s.storageApiURL, "volumes", name)
	response, err := utils.HttpExecute(s.httpClient, s.logger, "GET", getRemoteURL, nil)
	if err != nil {
		s.logger.Printf("Error in get volume remote call %#v", err)
		return resources.Volume{}, fmt.Errorf("Error in get volume remote call")
	}

	if response.StatusCode != http.StatusOK {
		s.logger.Printf("Error in get volume remote call %#v", response)
		return resources.Volume{}, utils.ExtractErrorResponse(response)
	}

	getResponse := resources.GetResponse{}
	err = utils.UnmarshalResponse(response, &getResponse)
	if err != nil {
		s.logger.Printf("Error in unmarshalling response for get remote call %#v for response %#v", err, response)
		return resources.Volume{}, fmt.Errorf("Error in unmarshalling response for get remote call")
	}

	return getResponse.Volume, nil
}

func (s *remoteClient) GetVolumeConfig(name string) (map[string]interface{}, error) {
	s.logger.Println("remoteClient: GetVolumeConfig start")
	defer s.logger.Println("remoteClient: GetVolumeConfig finish")

	getRemoteURL := utils.FormatURL(s.storageApiURL, "volumes", name, "config")
	response, err := utils.HttpExecute(s.httpClient, s.logger, "GET", getRemoteURL, nil)
	if err != nil {
		s.logger.Printf("Error in get volume remote call %#v", err)
		return nil, fmt.Errorf("Error in get volume remote call")
	}

	if response.StatusCode != http.StatusOK {
		s.logger.Printf("Error in get volume remote call %#v", response)
		return nil, utils.ExtractErrorResponse(response)
	}

	getResponse := resources.GetConfigResponse{}
	err = utils.UnmarshalResponse(response, &getResponse)
	if err != nil {
		s.logger.Printf("Error in unmarshalling response for get remote call %#v for response %#v", err, response)
		return nil, fmt.Errorf("Error in unmarshalling response for get remote call")
	}

	return getResponse.VolumeConfig, nil
}

func (s *remoteClient) Attach(name string) (string, error) {
	s.logger.Println("remoteClient: attach start")
	defer s.logger.Println("remoteClient: attach end")

	attachRemoteURL := utils.FormatURL(s.storageApiURL, "volumes", name, "attach")
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

	volumeConfig, err := s.GetVolumeConfig(name)
	if err != nil {
		return "", err
	}

	mounter, err := s.getMounterForVolume(name)
	if err != nil {
		return "", fmt.Errorf("Error determining mounter for volume: %s", err.Error())
	}
	err = mounter.Mount(mountResponse.Mountpoint, volumeConfig)
	if err != nil {
		return "", err
	}
	return mountResponse.Mountpoint, nil
}

func (s *remoteClient) Detach(name string) error {
	s.logger.Println("remoteClient: detach start")
	defer s.logger.Println("remoteClient: detach end")

	mounter, err := s.getMounterForVolume(name)
	if err != nil {
		return fmt.Errorf("Volume not found")
	}

	volumeConfig, err := s.GetVolumeConfig(name)
	if err != nil {
		return err
	}

	err = mounter.Unmount(volumeConfig)
	if err != nil {
		return err
	}

	detachRemoteURL := utils.FormatURL(s.storageApiURL, "volumes", name, "detach")
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

func (s *remoteClient) ListVolumes() ([]resources.VolumeMetadata, error) {
	s.logger.Println("remoteClient: list start")
	defer s.logger.Println("remoteClient: list end")

	listRemoteURL := utils.FormatURL(s.storageApiURL, "volumes")
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

func (s *remoteClient) getMounterForVolume(name string) (mounter.Mounter, error) {
	s.logger.Println("remoteClient: getMounterForVolume start")
	defer s.logger.Println("remoteClient: getMounterForVolume end")
	volume, err := s.GetVolume(name)
	if err != nil {
		return nil, err
	}
	return mounter.GetMounterForVolume(s.logger, volume)
}
