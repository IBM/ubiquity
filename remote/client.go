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
	"net/http"
	"github.com/IBM/ubiquity/resources"
	"reflect"
	"github.com/IBM/ubiquity/utils"
	"github.com/IBM/ubiquity/utils/logs"
)

type remoteClient struct {
	logger            logs.Logger
	isActivated       bool
	httpClient        *http.Client
	storageApiURL     string
	config            resources.UbiquityPluginConfig
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
		return s.logger.ErrorRet(err, "utils.HttpExecute failed")
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return s.logger.ErrorRet(utils.ExtractErrorResponse(response), "failed", logs.Args{{"response", response}})
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
		return s.logger.ErrorRet(err, "utils.HttpExecute failed")
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return s.logger.ErrorRet(utils.ExtractErrorResponse(response), "failed", logs.Args{{"response", response}})
	}

	return nil
}

func (s *remoteClient) RemoveVolume(removeVolumeRequest resources.RemoveVolumeRequest) error {
	defer s.logger.Trace(logs.DEBUG)()

	removeRemoteURL := utils.FormatURL(s.storageApiURL, "volumes", removeVolumeRequest.Name)

	removeVolumeRequest.CredentialInfo = s.config.CredentialInfo
	response, err := utils.HttpExecute(s.httpClient, "DELETE", removeRemoteURL, removeVolumeRequest)
	if err != nil {
		return s.logger.ErrorRet(err, "utils.HttpExecute failed")
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return s.logger.ErrorRet(utils.ExtractErrorResponse(response), "failed", logs.Args{{"response", response}})
	}

	return nil
}

func (s *remoteClient) GetVolume(getVolumeRequest resources.GetVolumeRequest) (resources.Volume, error) {
	defer s.logger.Trace(logs.DEBUG)()

	getRemoteURL := utils.FormatURL(s.storageApiURL, "volumes", getVolumeRequest.Name)
	getVolumeRequest.CredentialInfo = s.config.CredentialInfo
	response, err := utils.HttpExecute(s.httpClient, "GET", getRemoteURL, getVolumeRequest)
	if err != nil {
		return resources.Volume{}, s.logger.ErrorRet(err, "failed")
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return resources.Volume{}, s.logger.ErrorRet(utils.ExtractErrorResponse(response), "failed", logs.Args{{"response", response}})
	}

	getResponse := resources.GetResponse{}
	err = utils.UnmarshalResponse(response, &getResponse)
	if err != nil {
		return resources.Volume{}, s.logger.ErrorRet(err, "utils.UnmarshalResponse failed", logs.Args{{"response", response}})
	}

	return getResponse.Volume, nil
}

func (s *remoteClient) GetVolumeConfig(getVolumeConfigRequest resources.GetVolumeConfigRequest) (map[string]interface{}, error) {
	defer s.logger.Trace(logs.DEBUG)()

	getRemoteURL := utils.FormatURL(s.storageApiURL, "volumes", getVolumeConfigRequest.Name, "config")
	getVolumeConfigRequest.CredentialInfo = s.config.CredentialInfo
	response, err := utils.HttpExecute(s.httpClient, "GET", getRemoteURL, getVolumeConfigRequest)
	if err != nil {
		return nil, s.logger.ErrorRet(err, "failed")
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, s.logger.ErrorRet(utils.ExtractErrorResponse(response), "failed", logs.Args{{"response", response}})
	}

	getResponse := resources.GetConfigResponse{}
	err = utils.UnmarshalResponse(response, &getResponse)
	if err != nil {
		return nil, s.logger.ErrorRet(err, "utils.UnmarshalResponse failed", logs.Args{{"response", response}})
	}

	return getResponse.VolumeConfig, nil
}

func (s *remoteClient) Attach(attachRequest resources.AttachRequest) (string, error) {
	defer s.logger.Trace(logs.DEBUG)()

	attachRemoteURL := utils.FormatURL(s.storageApiURL, "volumes", attachRequest.Name, "attach")
	attachRequest.CredentialInfo = s.config.CredentialInfo
	response, err := utils.HttpExecute(s.httpClient, "PUT", attachRemoteURL, attachRequest)
	if err != nil {
		return "", s.logger.ErrorRet(err, "utils.HttpExecute failed")
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return "", s.logger.ErrorRet(utils.ExtractErrorResponse(response), "failed", logs.Args{{"response", response}})
	}

	return "", nil
}

func (s *remoteClient) Detach(detachRequest resources.DetachRequest) error {
	defer s.logger.Trace(logs.DEBUG)()

	detachRemoteURL := utils.FormatURL(s.storageApiURL, "volumes", detachRequest.Name, "detach")
	detachRequest.CredentialInfo = s.config.CredentialInfo
	response, err := utils.HttpExecute(s.httpClient, "PUT", detachRemoteURL, detachRequest)
	if err != nil {
		return s.logger.ErrorRet(err, "utils.HttpExecute failed")
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return s.logger.ErrorRet(utils.ExtractErrorResponse(response), "failed", logs.Args{{"response", response}})
	}

	return nil
}

func (s *remoteClient) ListVolumes(listVolumesRequest resources.ListVolumesRequest) ([]resources.Volume, error) {
	defer s.logger.Trace(logs.DEBUG)()

	listRemoteURL := utils.FormatURL(s.storageApiURL, "volumes")
	listVolumesRequest.CredentialInfo = s.config.CredentialInfo
	response, err := utils.HttpExecute(s.httpClient, "GET", listRemoteURL, listVolumesRequest)
	if err != nil {
		return nil, s.logger.ErrorRet(err, "failed")
	}

	defer response.Body.Close()


	if response.StatusCode != http.StatusOK {
		return nil, s.logger.ErrorRet(utils.ExtractErrorResponse(response), "failed", logs.Args{{"response", response}})
	}

	listResponse := resources.ListResponse{}
	err = utils.UnmarshalResponse(response, &listResponse)
	if err != nil {
		return nil, s.logger.ErrorRet(err, "utils.UnmarshalResponse failed", logs.Args{{"response", response}})
	}

	return listResponse.Volumes, nil

}