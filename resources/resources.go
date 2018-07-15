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

package resources

import (
	"fmt"
	"github.com/jinzhu/gorm"
)

const (
	SpectrumScale     string = "spectrum-scale"
	SpectrumScaleNFS  string = "spectrum-scale-nfs"
	SoftlayerNFS      string = "softlayer-nfs"
	SCBE              string = "scbe"
	ScbeInterfaceName string = "Enabler for Containers"
)

type UbiquityServerConfig struct {
	Port                int
	LogPath             string
	ConfigPath          string
	SpectrumScaleConfig SpectrumScaleConfig
	ScbeConfig          ScbeConfig
	BrokerConfig        BrokerConfig
	DefaultBackend      string
	LogLevel            string
}

// TODO we should consider to move dedicated backend structs to the backend resource file instead of this one.
type SpectrumScaleConfig struct {
	DefaultFilesystemName string
	NfsServerAddr         string
	SshConfig             SshConfig
	RestConfig            RestConfig
	ForceDelete           bool
}

type CredentialInfo struct {
	UserName string `json:"username"`
	Password string `json:"password"`
	Group    string `json:"group"`
}

type ConnectionInfo struct {
	CredentialInfo CredentialInfo
	Port           int
	ManagementIP   string
}

type ScbeConfig struct {
	ConfigPath           string // TODO consider to remove later
	ConnectionInfo       ConnectionInfo
	DefaultService       string // SCBE storage service to be used by default if not mentioned by plugin
	DefaultVolumeSize    string // The default volume size in case not specified by user
	UbiquityInstanceName string // Prefix for the volume name in the storage side (max length 15 char)

	DefaultFilesystemType string // The default filesystem type to create on new provisioned volume during attachment to the host
}

const UbiquityInstanceNameMaxSize = 15
const DefaultForScbeConfigParamDefaultVolumeSize = "1"    // if customer don't mention size, then the default is 1gb
const DefaultForScbeConfigParamDefaultFilesystem = "ext4" // if customer don't mention fstype, then the default is ext4
const PathToMountUbiquityBlockDevices = "/ubiquity/%s"    // %s is the WWN of the volume # TODO this should be moved to docker plugin side
const OptionNameForVolumeFsType = "fstype"                // the option name of the fstype and also the key in the volumeConfig
const ScbeKeyVolAttachToHost = "attach-to"                // the key in map for volume to host attachments
const ScbeDefaultPort = 8440                              // the default port for SCBE management
const SslModeRequire = "require"
const SslModeVerifyFull = "verify-full"
const KeySslMode = "UBIQUITY_PLUGIN_SSL_MODE"
const KeyScbeSslMode = "SCBE_SSL_MODE"
const DefaultDbSslMode = SslModeVerifyFull
const DefaultScbeSslMode = SslModeVerifyFull
const DefaultPluginsSslMode = SslModeVerifyFull

type SshConfig struct {
	User string
	Host string
	Port string
}

type RestConfig struct {
	Endpoint string
	User     string
	Password string
	Hostname string
}

type SpectrumNfsRemoteConfig struct {
	ClientConfig string
}

type BrokerConfig struct {
	ConfigPath string
	Port       int //for CF Service broker
}

type UbiquityPluginConfig struct {
	DockerPlugin            UbiquityDockerPluginConfig
	LogPath                 string
	LogRotateMaxSize        int
	UbiquityServer          UbiquityServerConnectionInfo
	SpectrumNfsRemoteConfig SpectrumNfsRemoteConfig
	ScbeRemoteConfig        ScbeRemoteConfig
	Backends                []string
	LogLevel                string
	CredentialInfo          CredentialInfo
	SslConfig               UbiquityPluginSslConfig
}

type UbiquityDockerPluginConfig struct {
	//Address          string
	Port             int
	PluginsDirectory string
}

type UbiquityServerConnectionInfo struct {
	Address string
	Port    int
}

type ScbeRemoteConfig struct {
	SkipRescanISCSI bool
}

type UbiquityPluginSslConfig struct {
	UseSsl   bool
	SslMode  string
	VerifyCa string
}

//go:generate counterfeiter -o ../fakes/fake_storage_client.go . StorageClient

type StorageClient interface {
	Activate(activateRequest ActivateRequest) error
	CreateVolume(createVolumeRequest CreateVolumeRequest) error
	RemoveVolume(removeVolumeRequest RemoveVolumeRequest) error
	ListVolumes(listVolumeRequest ListVolumesRequest) ([]Volume, error)
	GetVolume(getVolumeRequest GetVolumeRequest) (Volume, error)
	GetVolumeConfig(getVolumeConfigRequest GetVolumeConfigRequest) (map[string]interface{}, error)
	Attach(attachRequest AttachRequest) (string, error)
	Detach(detachRequest DetachRequest) error
}

// volumeNotFoundError error for Attach, Detach, GetVolume, GetVolumeConfig, RemoveVolume interfaces if volume not found in Ubiquity DB
const VolumeNotFoundErrorMsg = "volume was not found in Ubiqutiy database."

type VolumeNotFoundError struct {
	VolName string
}

func (e *VolumeNotFoundError) Error() string {
	return fmt.Sprintf("[%s] "+VolumeNotFoundErrorMsg, e.VolName)
}

// volAlreadyExistsError error for Create interface if volume is already exist in the Ubiquity DB
type VolAlreadyExistsError struct {
	VolName string
}

func (e *VolAlreadyExistsError) Error() string {
	return fmt.Sprintf("Volume [%s] already exists.", e.VolName)
}

//go:generate counterfeiter -o ../fakes/fake_mounter.go . Mounter

type Mounter interface {
	Mount(mountRequest MountRequest) (string, error)
	Unmount(unmountRequest UnmountRequest) error
	ActionAfterDetach(request AfterDetachRequest) error
}

type ActivateRequest struct {
	CredentialInfo CredentialInfo
	Backends       []string
	Opts           map[string]string
	Context        RequestContext
}

type CreateVolumeRequest struct {
	CredentialInfo CredentialInfo
	Name           string
	Backend        string
	Opts           map[string]interface{}
	Context        RequestContext
}

type RemoveVolumeRequest struct {
	CredentialInfo CredentialInfo
	Name           string
	Context        RequestContext
}

type ListVolumesRequest struct {
	CredentialInfo CredentialInfo
	//TODO add filter
	Backends []string
	Context  RequestContext
}

type AttachRequest struct {
	CredentialInfo CredentialInfo
	Name           string
	Host           string
	Context        RequestContext
}

type DetachRequest struct {
	CredentialInfo CredentialInfo
	Name           string
	Host           string
	Context        RequestContext
}
type GetVolumeRequest struct {
	CredentialInfo CredentialInfo
	Name           string
	Context        RequestContext
}
type GetVolumeConfigRequest struct {
	CredentialInfo CredentialInfo
	Name           string
	Context        RequestContext
}
type ActivateResponse struct {
	Implements []string
	Err        string
}

type GenericResponse struct {
	Err string
}

type MountRequest struct {
	Mountpoint   string
	VolumeConfig map[string]interface{}
	Context      RequestContext
}
type UnmountRequest struct {
	// TODO missing Mountpoint string
	VolumeConfig map[string]interface{}
	Context      RequestContext
}
type AfterDetachRequest struct {
	VolumeConfig map[string]interface{}
	Context      RequestContext
}
type AttachResponse struct {
	Mountpoint string
	Err        string
}

type MountResponse struct {
	Mountpoint string
	Err        string
}

type GetResponse struct {
	Volume Volume
	Err    string
}

type DockerGetResponse struct {
	Volume map[string]interface{}
	Err    string
}

type Volume struct {
	gorm.Model
	Name       string
	Backend    string
	Mountpoint string
}

type GetConfigResponse struct {
	VolumeConfig map[string]interface{}
	Err          string
}

type ListResponse struct {
	Volumes []Volume
	Err     string
}

type FlexVolumeResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Device  string `json:"device"`
}

type FlexVolumeMountRequest struct {
	MountPath   string                 `json:"mountPath"`
	MountDevice string                 `json:"name"`
	Opts        map[string]interface{} `json:"opts"`
}

type FlexVolumeUnmountRequest struct {
	MountPath string `json:"mountPath"`
}

type FlexVolumeAttachRequest struct {
	Name string            `json:"name"`
	Host string            `json:"host"`
	Opts map[string]string `json:"opts"`
}

type FlexVolumeDetachRequest struct {
	Name string `json:"name"`
	Host string `json:"host"`
}

type RequestContext struct {
	Id string
	ActionName	string
}
