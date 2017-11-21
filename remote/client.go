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
	"github.com/IBM/ubiquity/utils/logs"
	"io/ioutil"
)

type remoteClient struct {
	logger            logs.Logger
	legacylogger      *log.Logger
	isActivated       bool
	isMounted         bool
	httpClient        *http.Client
	storageApiURL     string
	config            resources.UbiquityPluginConfig
	mounterPerBackend map[string]resources.Mounter
}

func (s *remoteClient) Activate(activateRequest resources.ActivateRequest) error {
	defer s.logger.Trace(logs.DEBUG)()

	if s.isActivated {
		return nil
	}

	// call remote activate
	activateURL := utils.FormatURL(s.storageApiURL, "activate")
	activateRequest.CredentialInfo = s.config.CredentialInfo
	response, err := utils.HttpExecute(s.httpClient,"POST", activateURL, activateRequest)
	if err != nil {
		err = fmt.Errorf("Error in activate remote call %#v", err)
		return s.logger.ErrorRet(err, "failed")
	}

	if response.StatusCode != http.StatusOK {
		s.logger.Error(fmt.Sprintf("Error in activate remote call %#v", response))
		return s.logger.ErrorRet(utils.ExtractErrorResponse(response), "failed")
	}
	s.isActivated = true
	return nil
}

func (s *remoteClient) CreateVolume(createVolumeRequest resources.CreateVolumeRequest) error {
	defer s.logger.Trace(logs.DEBUG)()

	createRemoteURL := utils.FormatURL(s.storageApiURL, "volumes")

	if reflect.DeepEqual(s.config.SpectrumNfsRemoteConfig, resources.SpectrumNfsRemoteConfig{}) == false {
		createVolumeRequest.Opts["nfsClientConfig"] = s.config.SpectrumNfsRemoteConfig.ClientConfig
	}

	createVolumeRequest.CredentialInfo = s.config.CredentialInfo
	response, err := utils.HttpExecute(s.httpClient, "POST", createRemoteURL, createVolumeRequest)
	if err != nil {
        err = fmt.Errorf("Error in create volume remote call %s", err.Error())
        return s.logger.ErrorRet(err, "failed")
	}
	_, err = ioutil.ReadAll(response.Body)
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
        s.logger.Error(fmt.Sprintf("Error in create volume remote call %#v", response))
        return s.logger.ErrorRet(utils.ExtractErrorResponse(response), "failed")
	}

	return nil
}

func (s *remoteClient) RemoveVolume(removeVolumeRequest resources.RemoveVolumeRequest) error {
	defer s.logger.Trace(logs.DEBUG)()

	removeRemoteURL := utils.FormatURL(s.storageApiURL, "volumes", removeVolumeRequest.Name)

	removeVolumeRequest.CredentialInfo = s.config.CredentialInfo
	response, err := utils.HttpExecute(s.httpClient, "DELETE", removeRemoteURL, removeVolumeRequest)
	if err != nil {
        err = fmt.Errorf("Error in remove volume remote call %#v", err)
        return s.logger.ErrorRet(err, "failed")
	}
	_, err = ioutil.ReadAll(response.Body)
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
        s.logger.Error(fmt.Sprintf("Error in remove volume remote call %#v", response))
        return s.logger.ErrorRet(utils.ExtractErrorResponse(response), "failed")
	}

	return nil
}

func (s *remoteClient) GetVolume(getVolumeRequest resources.GetVolumeRequest) (resources.Volume, error) {
	defer s.logger.Trace(logs.DEBUG)()

	getRemoteURL := utils.FormatURL(s.storageApiURL, "volumes", getVolumeRequest.Name)
	getVolumeRequest.CredentialInfo = s.config.CredentialInfo
	response, err := utils.HttpExecute(s.httpClient, "GET", getRemoteURL, getVolumeRequest)
	if err != nil {
        err = fmt.Errorf("Error in get volume remote call %#v", err)
        return resources.Volume{}, s.logger.ErrorRet(err, "failed")
	}

	if response.StatusCode != http.StatusOK {
        s.logger.Error(fmt.Sprintf("Error in get volume remote call %#v", response))
        return resources.Volume{}, s.logger.ErrorRet(utils.ExtractErrorResponse(response), "failed")
	}

	getResponse := resources.GetResponse{}
	err = utils.UnmarshalResponse(response, &getResponse)
	if err != nil {
        err = fmt.Errorf("Error in unmarshalling response for get remote call %#v for response %#v", err, response)
		return resources.Volume{}, s.logger.ErrorRet(err, "failed")
	}

	return getResponse.Volume, nil
}

func (s *remoteClient) GetVolumeConfig(getVolumeConfigRequest resources.GetVolumeConfigRequest) (map[string]interface{}, error) {
	defer s.logger.Trace(logs.DEBUG)()

	getRemoteURL := utils.FormatURL(s.storageApiURL, "volumes", getVolumeConfigRequest.Name, "config")
	getVolumeConfigRequest.CredentialInfo = s.config.CredentialInfo
	response, err := utils.HttpExecute(s.httpClient, "GET", getRemoteURL, getVolumeConfigRequest)
	if err != nil {
        err = fmt.Errorf("Error in get volume remote call %#v", err)
        return nil, s.logger.ErrorRet(err, "failed")
	}

	if response.StatusCode != http.StatusOK {
        s.logger.Error(fmt.Sprintf("Error in get volume remote call %#v", response))
        return nil, s.logger.ErrorRet(utils.ExtractErrorResponse(response), "failed")
	}

	getResponse := resources.GetConfigResponse{}
	err = utils.UnmarshalResponse(response, &getResponse)
	if err != nil {
        err = fmt.Errorf("Error in unmarshalling response for get remote call %#v for response %#v", err, response)
        return nil, s.logger.ErrorRet(err, "failed")
	}

	return getResponse.VolumeConfig, nil
}

func (s *remoteClient) Attach(attachRequest resources.AttachRequest) (string, error) {
	defer s.logger.Trace(logs.DEBUG)()

	attachRemoteURL := utils.FormatURL(s.storageApiURL, "volumes", attachRequest.Name, "attach")
	attachRequest.CredentialInfo = s.config.CredentialInfo
	response, err := utils.HttpExecute(s.httpClient, "PUT", attachRemoteURL, attachRequest)
	if err != nil {
        err = fmt.Errorf("Error in attach volume remote call %#v", err)
        return "", s.logger.ErrorRet(err, "failed")
	}

	if response.StatusCode != http.StatusOK {
        s.logger.Error(fmt.Sprintf("Error in attach volume remote call %#v", response))
        return "", s.logger.ErrorRet(utils.ExtractErrorResponse(response), "failed")
	}

	attachResponse := resources.AttachResponse{}
	err = utils.UnmarshalResponse(response, &attachResponse)
	if err != nil {
        err = fmt.Errorf("Error in unmarshalling response for attach remote call %#v for response %#v", err, response)
        return "", s.logger.ErrorRet(err, "failed")
	}

	return "", nil
}

func (s *remoteClient) Detach(detachRequest resources.DetachRequest) error {
	defer s.logger.Trace(logs.DEBUG)()

	detachRemoteURL := utils.FormatURL(s.storageApiURL, "volumes", detachRequest.Name, "detach")
	detachRequest.CredentialInfo = s.config.CredentialInfo
	response, err := utils.HttpExecute(s.httpClient, "PUT", detachRemoteURL, detachRequest)
	if err != nil {
        err = fmt.Errorf("Error in detach volume remote call %#v", err)
        return s.logger.ErrorRet(err, "failed")
	}
	_, err = ioutil.ReadAll(response.Body)
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
        s.logger.Error(fmt.Sprintf("Error in detach volume remote call %#v", response))
        return s.logger.ErrorRet(utils.ExtractErrorResponse(response), "failed")
	}

	return nil
}

func (s *remoteClient) ListVolumes(listVolumesRequest resources.ListVolumesRequest) ([]resources.Volume, error) {
	defer s.logger.Trace(logs.DEBUG)()

	listRemoteURL := utils.FormatURL(s.storageApiURL, "volumes")
	listVolumesRequest.CredentialInfo = s.config.CredentialInfo
	response, err := utils.HttpExecute(s.httpClient, "GET", listRemoteURL, listVolumesRequest)
	if err != nil {
        err = fmt.Errorf("Error in list volume remote call %#v", err)
        return nil, s.logger.ErrorRet(err, "failed")
	}

	if response.StatusCode != http.StatusOK {
        s.logger.Error(fmt.Sprintf("Error in list volume remote call %#v", err))
        return nil, s.logger.ErrorRet(utils.ExtractErrorResponse(response), "failed")
	}

	listResponse := resources.ListResponse{}
	err = utils.UnmarshalResponse(response, &listResponse)
	if err != nil {
        err = fmt.Errorf("Error in unmarshalling response for get remote call %#v for response %#v", err, response)
        return []resources.Volume{}, s.logger.ErrorRet(err, "failed")
	}

	return listResponse.Volumes, nil

}

// Return the mounter object. If mounter object already used(in the map mounterPerBackend) then just reuse it
func (s *remoteClient) getMounterForBackend(backend string) (resources.Mounter, error) {
	defer s.logger.Trace(logs.DEBUG)()

	mounterInst, ok := s.mounterPerBackend[backend]
	if ok {
		s.logger.Debug("getMounterForVolume reuse existing mounter", logs.Args{{"backend", backend}})
		return mounterInst, nil
	} else if backend == resources.SpectrumScale {
		s.mounterPerBackend[backend] = mounter.NewSpectrumScaleMounter(s.legacylogger)
	} else if backend == resources.SoftlayerNFS || backend == resources.SpectrumScaleNFS {
		s.mounterPerBackend[backend] = mounter.NewNfsMounter(s.legacylogger)
	} else if backend == resources.SCBE {
		s.mounterPerBackend[backend] = mounter.NewScbeMounter(s.config.ScbeRemoteConfig)
	} else {
		return nil, s.logger.ErrorRet(fmt.Errorf("Mounter not found for backend: %s", backend), "failed")
	}
	return s.mounterPerBackend[backend], nil
}
