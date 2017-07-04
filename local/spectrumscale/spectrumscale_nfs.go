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

package spectrumscale

import (
	"fmt"
	"log"

	"github.com/IBM/ubiquity/resources"
	"github.com/IBM/ubiquity/utils"
	"github.com/jinzhu/gorm"
)

type spectrumNfsLocalClient struct {
	spectrumClient *spectrumLocalClient
	config         resources.SpectrumScaleConfig
	executor       utils.Executor
}

func NewSpectrumNfsLocalClient(logger *log.Logger, config resources.UbiquityServerConfig, db *gorm.DB) (resources.StorageClient, error) {
	logger.Println("spectrumNfsLocalClient: init start")
	defer logger.Println("spectrumNfsLocalClient: init end")

	if config.ConfigPath == "" {
		return nil, fmt.Errorf("spectrumNfsLocalClient: init: missing required parameter 'spectrumConfigPath'")
	}

	if config.SpectrumScaleConfig.DefaultFilesystemName == "" {
		return nil, fmt.Errorf("spectrumNfsLocalClient: init: missing required parameter 'spectrumDefaultFileSystem'")
	}

	if config.SpectrumScaleConfig.NfsServerAddr == "" {
		return nil, fmt.Errorf("spectrumNfsLocalClient: init: missing required parameter 'spectrumNfsServerAddr'")
	}

	spectrumClient, err := newSpectrumLocalClient(logger, config.SpectrumScaleConfig, db, resources.SpectrumScaleNFS)
	if err != nil {
		return nil, err
	}
	return &spectrumNfsLocalClient{config: config.SpectrumScaleConfig, spectrumClient: spectrumClient, executor: utils.NewExecutor()}, nil
}

func (s *spectrumNfsLocalClient) Activate(activateRequest resources.ActivateRequest) error {
	s.spectrumClient.logger.Println("spectrumNfsLocalClient: Activate-start")
	defer s.spectrumClient.logger.Println("spectrumNfsLocalClient: Activate-end")

	return s.spectrumClient.Activate(activateRequest)
}

func (s *spectrumNfsLocalClient) ListVolumes(listVolumesRequest resources.ListVolumesRequest) ([]resources.Volume, error) {
	s.spectrumClient.logger.Println("spectrumNfsLocalClient: List-volumes-start")
	defer s.spectrumClient.logger.Println("spectrumNfsLocalClient: List-volumes-end")

	return s.spectrumClient.ListVolumes(listVolumesRequest)

}
func (s *spectrumNfsLocalClient) Attach(attachRequest resources.AttachRequest) (string, error) {
	s.spectrumClient.logger.Println("spectrumNfsLocalClient: Attach-start")
	defer s.spectrumClient.logger.Println("spectrumNfsLocalClient: Attach-end")

	getVolumeConfigRequest := resources.GetVolumeConfigRequest{Name: attachRequest.Name}
	volumeConfig, err := s.GetVolumeConfig(getVolumeConfigRequest)
	if err != nil {
		return "", err
	}
	nfsShare, ok := volumeConfig["nfs_share"].(string)
	if !ok {
		err = fmt.Errorf("error getting NFS share info from volume config for volume %s", attachRequest.Name)
		s.spectrumClient.logger.Println("spectrumNfsLocalClient: error: %v", err.Error())
		return "", err
	}
	return nfsShare, nil
}

func (s *spectrumNfsLocalClient) Detach(detachRequest resources.DetachRequest) error {
	s.spectrumClient.logger.Println("spectrumNfsLocalClient: Detach-start")
	defer s.spectrumClient.logger.Println("spectrumNfsLocalClient: Detach-end")

	getVolumeConfigRequest := resources.GetVolumeConfigRequest{Name: detachRequest.Name}
	_, err := s.spectrumClient.GetVolumeConfig(getVolumeConfigRequest)

	if err != nil {
		s.spectrumClient.logger.Printf("spectrumNfsLocalClient: error in no-op detach for volume %s: %#v\n", detachRequest.Name, err)
		return err
	}

	return nil
}

func (s *spectrumNfsLocalClient) CreateVolume(createVolumeRequest resources.CreateVolumeRequest) error {

	s.spectrumClient.logger.Printf("spectrumNfsLocalClient: Create-start")
	defer s.spectrumClient.logger.Println("spectrumNfsLocalClient: Create-end")

	nfsClientConfig, ok := createVolumeRequest.Opts["nfsClientConfig"].(string)
	if !ok {
		errorMsg := "Cannot create volume (opts missing required parameter 'nfsClientConfig')"
		s.spectrumClient.logger.Printf("spectrumNfsLocalClient: Create: Error: %s", errorMsg)
		return fmt.Errorf(errorMsg)
	}

	spectrumOpts := make(map[string]interface{})
	for k, v := range createVolumeRequest.Opts {
		if k != "nfsClientConfig" {
			spectrumOpts[k] = v
		}
	}

	if err := s.spectrumClient.CreateVolume(createVolumeRequest); err != nil {
		s.spectrumClient.logger.Printf("spectrumNfsLocalClient: error creating volume %#v\n", err)
		return err
	}
	attachRequest := resources.AttachRequest{Name: createVolumeRequest.Name}
	if _, err := s.spectrumClient.Attach(attachRequest); err != nil {
		s.spectrumClient.logger.Printf("spectrumNfsLocalClient: error attaching volume %#v\n; deleting volume", err)

		removeVolumeRequest := resources.RemoveVolumeRequest{Name: createVolumeRequest.Name}
		s.spectrumClient.RemoveVolume(removeVolumeRequest)
		return err
	}

	if err := s.spectrumClient.updatePermissions(createVolumeRequest.Name); err != nil {
		s.spectrumClient.logger.Printf("spectrumNfsLocalClient: error updating permissions of volume %#v\n; deleting volume", err)
		removeVolumeRequest := resources.RemoveVolumeRequest{Name: createVolumeRequest.Name}
		s.spectrumClient.RemoveVolume(removeVolumeRequest)
		return err
	}

	if err := s.exportNfs(createVolumeRequest.Name, nfsClientConfig); err != nil {
		s.spectrumClient.logger.Printf("spectrumNfsLocalClient: error exporting volume %#v\n; deleting volume", err)
		removeVolumeRequest := resources.RemoveVolumeRequest{Name: createVolumeRequest.Name}
		s.spectrumClient.RemoveVolume(removeVolumeRequest)
		return err
	}
	return nil
}

func (s *spectrumNfsLocalClient) RemoveVolume(removeVolumeRequest resources.RemoveVolumeRequest) error {
	s.spectrumClient.logger.Println("spectrumNfsLocalClient: Remove-start")
	defer s.spectrumClient.logger.Println("spectrumNfsLocalClient: Remove-end")

	if err := s.unexportNfs(removeVolumeRequest.Name); err != nil {
		s.spectrumClient.logger.Printf("spectrumNfsLocalClient: Could not unexport volume %s (error=%s)", removeVolumeRequest.Name, err.Error())
	}
	detachRequest := resources.DetachRequest{Name: removeVolumeRequest.Name}
	if err := s.spectrumClient.Detach(detachRequest); err != nil {
		s.spectrumClient.logger.Printf("spectrumNfsLocalClient: Could not detach volume %s (error=%s)", removeVolumeRequest.Name, err.Error())
	}

	return s.spectrumClient.RemoveVolume(removeVolumeRequest)
}

func (s *spectrumNfsLocalClient) GetVolumeConfig(getVolumeConfigRequest resources.GetVolumeConfigRequest) (map[string]interface{}, error) {
	s.spectrumClient.logger.Println("spectrumNfsLocalClient: GetVolumeConfig-start")
	defer s.spectrumClient.logger.Println("spectrumNfsLocalClient: GetVolumeConfig-end")

	volumeConfig, err := s.spectrumClient.GetVolumeConfig(getVolumeConfigRequest)
	if err != nil {
		return volumeConfig, err
	}
	mountpoint, exists := volumeConfig["mountpoint"]
	if exists == false {
		return nil, fmt.Errorf("Volume :%s not found", getVolumeConfigRequest.Name)
	}
	nfsShare := fmt.Sprintf("%s:%s", s.config.NfsServerAddr, mountpoint)
	volumeConfig["nfs_share"] = nfsShare
	s.spectrumClient.logger.Printf("spectrumNfsLocalClient: GetVolume: Adding nfs_share %s to volume config for volume %s\n", nfsShare, getVolumeConfigRequest.Name)
	return volumeConfig, nil
}
func (s *spectrumNfsLocalClient) GetVolume(getVolumeRequest resources.GetVolumeRequest) (resources.Volume, error) {
	s.spectrumClient.logger.Println("spectrumNfsLocalClient: GetVolume start")
	defer s.spectrumClient.logger.Println("spectrumNfsLocalClient: GetVolume finish")
	return s.spectrumClient.GetVolume(getVolumeRequest)
}

func (s *spectrumNfsLocalClient) exportNfs(name, clientConfig string) error {
	s.spectrumClient.logger.Printf("spectrumNfsLocalClient: ExportNfs start with name=%#v and clientConfig=%#v\n", name, clientConfig)
	defer s.spectrumClient.logger.Printf("spectrumNfsLocalClient: ExportNfs end")

	existingVolume, exists, err := s.spectrumClient.dataModel.GetVolume(name)

	if err != nil {
		s.spectrumClient.logger.Printf("spectrumNfsLocalClient: DbClient.GetVolume returned error %#v\n", err.Error())
		return err
	}
	if exists == false {
		s.spectrumClient.logger.Printf("spectrumNfsLocalClient: volume %#s not found\n", err)
		return err
	}

	volumeMountpoint, err := s.spectrumClient.getVolumeMountPoint(existingVolume)
	if err != nil {
		return err
	}

	return s.spectrumClient.connector.ExportNfs(volumeMountpoint, clientConfig)
}

func (s *spectrumNfsLocalClient) unexportNfs(name string) error {
	s.spectrumClient.logger.Printf("spectrumNfsLocalClient: UnexportNfs start with name=%s\n", name)
	defer s.spectrumClient.logger.Printf("spectrumNfsLocalClient: ExportNfs end")

	existingVolume, exists, err := s.spectrumClient.dataModel.GetVolume(name)

	if err != nil {
		s.spectrumClient.logger.Printf("spectrumNfsLocalClient: error getting volume %#s \n", err)
		return err
	}
	if exists == false {
		s.spectrumClient.logger.Printf("spectrumNfsLocalClient: volume %#s not found\n", err)
		return err
	}

	volumeMountpoint, err := s.spectrumClient.getVolumeMountPoint(existingVolume)
	if err != nil {
		return err
	}

	return s.spectrumClient.connector.UnexportNfs(volumeMountpoint)
}
