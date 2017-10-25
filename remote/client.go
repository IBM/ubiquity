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
	"io/ioutil"
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

func (s *remoteClient) Activate(activateRequest resources.ActivateRequest) error {
	s.logger.Println("remoteClient: Activate start")
	defer s.logger.Println("remoteClient: Activate end")

	if s.isActivated {
		return nil
	}

	// call remote activate
	activateURL := utils.FormatURL(s.storageApiURL, "activate")
	activateRequest.CredentialInfo = s.config.CredentialInfo
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

	createVolumeRequest.CredentialInfo = s.config.CredentialInfo
	response, err := utils.HttpExecute(s.httpClient, s.logger, "POST", createRemoteURL, createVolumeRequest)
	if err != nil {
		s.logger.Printf("Error in create volume remote call %s", err.Error())
		return fmt.Errorf("Error in create volume remote call(http error)")
	}
	_, err = ioutil.ReadAll(response.Body)
	defer response.Body.Close()

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

	removeVolumeRequest.CredentialInfo = s.config.CredentialInfo
	response, err := utils.HttpExecute(s.httpClient, s.logger, "DELETE", removeRemoteURL, removeVolumeRequest)
	if err != nil {
		s.logger.Printf("Error in remove volume remote call %#v", err)
		return fmt.Errorf("Error in remove volume remote call")
	}
	_, err = ioutil.ReadAll(response.Body)
	defer response.Body.Close()

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
	getVolumeRequest.CredentialInfo = s.config.CredentialInfo
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
	getVolumeConfigRequest.CredentialInfo = s.config.CredentialInfo
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
	attachRequest.CredentialInfo = s.config.CredentialInfo
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

	return "", nil
}

func (s *remoteClient) Detach(detachRequest resources.DetachRequest) error {
	s.logger.Println("remoteClient: detach start")
	defer s.logger.Println("remoteClient: detach end")

	detachRemoteURL := utils.FormatURL(s.storageApiURL, "volumes", detachRequest.Name, "detach")
	detachRequest.CredentialInfo = s.config.CredentialInfo
	response, err := utils.HttpExecute(s.httpClient, s.logger, "PUT", detachRemoteURL, detachRequest)
	if err != nil {
		s.logger.Printf("Error in detach volume remote call %#v", err)
		return fmt.Errorf("Error in detach volume remote call")
	}
	_, err = ioutil.ReadAll(response.Body)
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		s.logger.Printf("Error in detach volume remote call %#v", response)
		return utils.ExtractErrorResponse(response)
	}

	return nil
}

func (s *remoteClient) ListVolumes(listVolumesRequest resources.ListVolumesRequest) ([]resources.Volume, error) {
	s.logger.Println("remoteClient: list start")
	defer s.logger.Println("remoteClient: list end")

	listRemoteURL := utils.FormatURL(s.storageApiURL, "volumes")
	listVolumesRequest.CredentialInfo = s.config.CredentialInfo
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
		s.mounterPerBackend[backend] = mounter.NewScbeMounter(s.config.ScbeRemoteConfig)
	} else {
		return nil, fmt.Errorf("Mounter not found for backend: %s", backend)
	}
	return s.mounterPerBackend[backend], nil
}
