package resources

import "github.com/jinzhu/gorm"

const (
	SpectrumScale    string = "spectrum-scale"
	SpectrumScaleNFS string = "spectrum-scale-nfs"
	SoftlayerNFS     string = "softlayer-nfs"
	SCBE             string = "scbe"
)

type UbiquityServerConfig struct {
	Port                int
	LogPath             string
	ConfigPath          string
	SpectrumScaleConfig SpectrumScaleConfig
	ScbeConfig          ScbeConfig
	BrokerConfig        BrokerConfig
	DefaultBackend      string
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
	SkipVerifySSL  bool
}

type ScbeConfig struct {
	ConfigPath           string // TODO consider to remove later
	ConnectionInfo       ConnectionInfo
	DefaultService       string // SCBE storage service to be used by default if not mentioned by plugin
	DefaultVolumeSize    string // The default volume size in case not specified by user
	DefaultFilesystem    string // The default filesystem to create on new volumes
	UbiquityInstanceName string // Prefix for the volume name in the storage side (max length 15 char)
}

const UbiquityInstanceNameMaxSize = 15
const DefaultForScbeConfigParamDefaultVolumeSize = "1"    // if customer don't mention size, then the default is 1gb
const DefaultForScbeConfigParamDefaultFilesystem = "ext4" // if customer don't mention fstype, then the default is ext4
const PathToMountUbiquityBlockDevices = "/ubiquity/%s"    // %s is the WWN of the volume # TODO this should be moved to docker plugin side

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
	ActionAfterDetach(request AfterDetachRequest) error
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
type AfterDetachRequest struct {
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
