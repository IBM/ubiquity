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

package connectors

import (
	"log"

	"github.com/IBM/ubiquity/resources"
)

//go:generate counterfeiter -o ../../../fakes/fake_spectrum.go . SpectrumScaleConnector
type SpectrumScaleConnector interface {
	//Cluster operations
	GetClusterId() (string, error)
	//Filesystem operations
	IsFilesystemMounted(filesystemName string) (bool, error)
	MountFileSystem(filesystemName string) error
	ListFilesystems() ([]string, error)
	GetFilesystemMountpoint(filesystemName string) (string, error)
	//Fileset operations
	CreateFileset(filesystemName string, filesetName string, opts map[string]interface{}) error
	DeleteFileset(filesystemName string, filesetName string) error
	LinkFileset(filesystemName string, filesetName string) error
	UnlinkFileset(filesystemName string, filesetName string) error
	ListFilesets(filesystemName string) ([]resources.Volume, error)
	ListFileset(filesystemName string, filesetName string) (resources.Volume, error)
	IsFilesetLinked(filesystemName string, filesetName string) (bool, error)
	//TODO modify quota from string to Capacity (see kubernetes)
	ListFilesetQuota(filesystemName string, filesetName string) (string, error)
	SetFilesetQuota(filesystemName string, filesetName string, quota string) error
	ExportNfs(volumeMountpoint string, clientConfig string) error
	UnexportNfs(volumeMountpoint string) error
}

const (
	UserSpecifiedFilesetType string = "fileset-type"
	UserSpecifiedInodeLimit  string = "inode-limit"
)

func GetSpectrumScaleConnector(logger *log.Logger, config resources.SpectrumScaleConfig) (SpectrumScaleConnector, error) {
	if config.RestConfig.Endpoint != "" {
		logger.Printf("Initializing SpectrumScale REST connector\n")
		return NewSpectrumRestV2(logger, config.RestConfig)
	}
	if config.SshConfig.User != "" && config.SshConfig.Host != "" {
		if config.SshConfig.Port == "" || config.SshConfig.Port == "0" {
			config.SshConfig.Port = "22"
		}
		logger.Printf("Initializing SpectrumScale SSH connector with sshConfig: %+v\n", config.SshConfig)
		return NewSpectrumSSH(logger, config.SshConfig)
	}
	logger.Println("Initializing SpectrumScale MMCLI Connector")
	return NewSpectrumMMCLI(logger)
}
