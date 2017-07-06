/**
 * Copyright 2017 IBM Corp.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package remote

import (
	"fmt"
	"log"

	"net/http"

	"github.com/IBM/ubiquity/resources"

	"reflect"

	"github.com/IBM/ubiquity/remote/mounter"
	"github.com/IBM/ubiquity/utils"
)

type remoteClient struct {
	logger            *log.Logger
	isActivated       bool
	isMounted         bool
	httpClient        *http.Client
	storageApiURL     string
	config            resources.UbiquityPluginConfig
	mounterPerBackend map[string]resources.Mounter
}

func NewRemoteClient(logger *log.Logger, storageApiURL string, config resources.UbiquityPluginConfig) (resources.StorageClient, error) {
	return &remoteClient{logger: logger, storageApiURL: storageApiURL, httpClient: &http.Client{}, config: config, mounterPerBackend: make(map[string]resources.Mounter)}, nil
}

func (s *remoteClient) Activate(activateRequest resources.ActivateRequest) error {
	s.logger.Println("remoteClient: Activate start")
	defer s.logger.Println("remoteClient: Activate end")

	if s.isActivated {
		return nil
	}

	// call remote activate
	activateURL := utils.FormatURL(s.storageApiURL, "activate")
	response, err := utils.HttpExecute(s.httpClient, s.logger, "POST", activateURL, activateRequest)
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

func (s *remoteClient) CreateVolume(createVolumeRequest resources.CreateVolumeRequest) error {
	s.logger.Println("remoteClient: create start")
	defer s.logger.Println("remoteClient: create end")

	createRemoteURL := utils.FormatURL(s.storageApiURL, "volumes")

	if reflect.DeepEqual(s.config.SpectrumNfsRemoteConfig, resources.SpectrumNfsRemoteConfig{}) == false {
		createVolumeRequest.Opts["nfsClientConfig"] = s.config.SpectrumNfsRemoteConfig.ClientConfig
	}

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

func (s *remoteClient) RemoveVolume(removeVolumeRequest resources.RemoveVolumeRequest) error {
	s.logger.Println("remoteClient: remove start")
	defer s.logger.Println("remoteClient: remove end")

	removeRemoteURL := utils.FormatURL(s.storageApiURL, "volumes", removeVolumeRequest.Name)

	response, err := utils.HttpExecute(s.httpClient, s.logger, "DELETE", removeRemoteURL, removeVolumeRequest)
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

func (s *remoteClient) GetVolume(getVolumeRequest resources.GetVolumeRequest) (resources.Volume, error) {
	s.logger.Println("remoteClient: get start")
	defer s.logger.Println("remoteClient: get finish")

	getRemoteURL := utils.FormatURL(s.storageApiURL, "volumes", getVolumeRequest.Name)
	response, err := utils.HttpExecute(s.httpClient, s.logger, "GET", getRemoteURL, getVolumeRequest)
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

func (s *remoteClient) GetVolumeConfig(getVolumeConfigRequest resources.GetVolumeConfigRequest) (map[string]interface{}, error) {
	s.logger.Println("remoteClient: GetVolumeConfig start")
	defer s.logger.Println("remoteClient: GetVolumeConfig finish")

	getRemoteURL := utils.FormatURL(s.storageApiURL, "volumes", getVolumeConfigRequest.Name, "config")
	response, err := utils.HttpExecute(s.httpClient, s.logger, "GET", getRemoteURL, getVolumeConfigRequest)
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

func (s *remoteClient) Attach(attachRequest resources.AttachRequest) (string, error) {
	s.logger.Println("remoteClient: attach start")
	defer s.logger.Println("remoteClient: attach end")

	attachRemoteURL := utils.FormatURL(s.storageApiURL, "volumes", attachRequest.Name, "attach")
	response, err := utils.HttpExecute(s.httpClient, s.logger, "PUT", attachRemoteURL, attachRequest)
	if err != nil {
		s.logger.Printf("Error in attach volume remote call %#v", err)
		return "", fmt.Errorf("Error in attach volume remote call")
	}

	if response.StatusCode != http.StatusOK {
		s.logger.Printf("Error in attach volume remote call %#v", response)

		return "", utils.ExtractErrorResponse(response)
	}

	attachResponse := resources.AttachResponse{}
	err = utils.UnmarshalResponse(response, &attachResponse)
	if err != nil {
		return "", fmt.Errorf("Error in unmarshalling response for attach remote call")
	}
	getVolumeConfigRequest := resources.GetVolumeConfigRequest{Name: attachRequest.Name}
	volumeConfig, err := s.GetVolumeConfig(getVolumeConfigRequest)
	if err != nil {
		return "", err
	}
	getVolumeRequest := resources.GetVolumeRequest{Name: attachRequest.Name}
	volume, err := s.GetVolume(getVolumeRequest)

	mounter, err := s.getMounterForBackend(volume.Backend)
	if err != nil {
		return "", fmt.Errorf("Error determining mounter for volume: %s", err.Error())
	}
	mountRequest := resources.MountRequest{Mountpoint: attachResponse.Mountpoint, VolumeConfig: volumeConfig}
	mountpoint, err := mounter.Mount(mountRequest)
	if err != nil {
		return "", err
	}

	return mountpoint, nil
}

func (s *remoteClient) Detach(detachRequest resources.DetachRequest) error {
	s.logger.Println("remoteClient: detach start")
	defer s.logger.Println("remoteClient: detach end")

	getVolumeRequest := resources.GetVolumeRequest{Name: detachRequest.Name}
	volume, err := s.GetVolume(getVolumeRequest)

	mounter, err := s.getMounterForBackend(volume.Backend)
	if err != nil {
		return fmt.Errorf("Volume not found")
	}

	getVolumeConfigRequest := resources.GetVolumeConfigRequest{Name: detachRequest.Name}
	volumeConfig, err := s.GetVolumeConfig(getVolumeConfigRequest)
	if err != nil {
		return err
	}
	unmountRequest := resources.UnmountRequest{VolumeConfig: volumeConfig}
	err = mounter.Unmount(unmountRequest)
	if err != nil {
		return err
	}

	detachRemoteURL := utils.FormatURL(s.storageApiURL, "volumes", detachRequest.Name, "detach")
	response, err := utils.HttpExecute(s.httpClient, s.logger, "PUT", detachRemoteURL, detachRequest)
	if err != nil {
		s.logger.Printf("Error in detach volume remote call %#v", err)
		return fmt.Errorf("Error in detach volume remote call")
	}

	if response.StatusCode != http.StatusOK {
		s.logger.Printf("Error in detach volume remote call %#v", response)
		return utils.ExtractErrorResponse(response)
	}

	afterDetachRequest := resources.AfterDetachRequest{VolumeConfig: volumeConfig}
	if err := mounter.ActionAfterDetach(afterDetachRequest); err != nil {
		s.logger.Printf(fmt.Sprintf("Error execute action after detaching the volume : %#v", err))
		return err
	}
	return nil

}

func (s *remoteClient) ListVolumes(listVolumesRequest resources.ListVolumesRequest) ([]resources.Volume, error) {
	s.logger.Println("remoteClient: list start")
	defer s.logger.Println("remoteClient: list end")

	listRemoteURL := utils.FormatURL(s.storageApiURL, "volumes")
	response, err := utils.HttpExecute(s.httpClient, s.logger, "GET", listRemoteURL, listVolumesRequest)
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
		return []resources.Volume{}, nil
	}

	return listResponse.Volumes, nil

}

// Return the mounter object. If mounter object already used(in the map mounterPerBackend) then just reuse it
func (s *remoteClient) getMounterForBackend(backend string) (resources.Mounter, error) {
	s.logger.Println("remoteClient: getMounterForVolume start")
	defer s.logger.Println("remoteClient: getMounterForVolume end")
	mounterInst, ok := s.mounterPerBackend[backend]
	if ok {
		s.logger.Printf("getMounterForVolume reuse existing mounter for backend " + backend)
		return mounterInst, nil
	} else if backend == resources.SpectrumScale {
		s.mounterPerBackend[backend] = mounter.NewSpectrumScaleMounter(s.logger)
	} else if backend == resources.SoftlayerNFS || backend == resources.SpectrumScaleNFS {
		s.mounterPerBackend[backend] = mounter.NewNfsMounter(s.logger)
	} else if backend == resources.SCBE {
		s.mounterPerBackend[backend] = mounter.NewScbeMounter()
	} else {
		return nil, fmt.Errorf("Mounter not found for backend: %s", backend)
	}
	return s.mounterPerBackend[backend], nil
}
