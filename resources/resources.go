package resources

import "github.com/jinzhu/gorm"

const (
	SpectrumScale    string = "spectrum-scale"
	SpectrumScaleNFS string = "spectrum-scale-nfs"
	SoftlayerNFS     string = "softlayer-nfs"
)

type UbiquityServerConfig struct {
	Port                int
	LogPath             string
	SpectrumScaleConfig SpectrumScaleConfig
	BrokerConfig        BrokerConfig
	DefaultBackend      string
}

type SpectrumScaleConfig struct {
	DefaultFilesystemName string
	ConfigPath            string
	NfsServerAddr         string
	SshConfig             SshConfig
	RestConfig            RestConfig
	ForceDelete           bool
}

type SshConfig struct {
	User string
	Host string
	Port string
}

type RestConfig struct {
	Endpoint string
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
	UbiquityServer          UbiquityServerConnectionInfo
	SpectrumNfsRemoteConfig SpectrumNfsRemoteConfig
	Backends                []string
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

//go:generate counterfeiter -o ../fakes/fake_mounter.go . Mounter

type Mounter interface {
	Mount(mountRequest MountRequest) (string, error)
	Unmount(unmountRequest UnmountRequest) error
}

type ActivateRequest struct {
	Backends []string
	Opts     map[string]string
}

type CreateVolumeRequest struct {
	Name    string
	Backend string
	Opts    map[string]interface{}
}

type RemoveVolumeRequest struct {
	Name string
}

type ListVolumesRequest struct {
	//TODO add filter
	Backends []string
}

type AttachRequest struct {
	Name string
	Host string
}

type DetachRequest struct {
	Name string
	Host string
}
type GetVolumeRequest struct {
	Name string
}
type GetVolumeConfigRequest struct {
	Name string
}
type ActivateResponse struct {
	Implements []string
	Err        string
}

type GenericResponse struct {
	Err string
}

type GenericRequest struct {
	Name string
}

type MountRequest struct {
	Mountpoint   string
	VolumeConfig map[string]interface{}
}
type UnmountRequest struct {
	VolumeConfig map[string]interface{}
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
	Volume Volume
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
