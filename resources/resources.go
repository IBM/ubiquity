package resources

import (
	"encoding/json"
	"fmt"
	"net/http"
)

const (
	SPECTRUM_SCALE     Backend = "spectrum-scale"
	SPECTRUM_SCALE_NFS Backend = "spectrum-scale-nfs"
	SOFTLAYER_NFS      Backend = "softlayer-nfs"
	SCBE               Backend = "scbe"
)

type Backend string

type UbiquityServerConfig struct {
	Port                int
	LogPath             string
	SpectrumScaleConfig SpectrumScaleConfig
	ScbeConfig          ScbeConfig
	BrokerConfig        BrokerConfig
	DefaultBackend      string
}

// TODO we should consider to move dedicated backend structs to the backend resource file instead of this one.
type SpectrumScaleConfig struct {
	DefaultFilesystem string
	ConfigPath        string
	NfsServerAddr     string
	SshConfig         SshConfig
	RestConfig        RestConfig
	ForceDelete       bool
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
	ConfigPath        string // TODO consider to remove later
	ConnectionInfo    ConnectionInfo
	DefaultService    string // SCBE storage service to be used by default if not mentioned by plugin
	DefaultVolumeSize string // The default volume size in case not specified by user
	DefaultFilesystem string // The default filesystem to create on new volumes
	HostnameTmp	  string // TODO this is a temp config param that workaround issue #23 (remove it when #23 will be fixed)
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
	Activate() error
	CreateVolume(name string, opts map[string]interface{}) error
	RemoveVolume(name string) error
	ListVolumes() ([]VolumeMetadata, error)
	GetVolume(name string) (Volume, error)
	GetVolumeConfig(name string) (map[string]interface{}, error)
	Attach(name string) (string, error)
	Detach(name string) error
}

type CreateRequest struct {
	Name string
	Opts map[string]interface{}
}

type RemoveRequest struct {
	Name string
}

type AttachRequest struct {
	Name string
}

type DetachRequest struct {
	Name string
}

type ActivateResponse struct {
	Implements []string
}

func (r *ActivateResponse) WriteResponse(w http.ResponseWriter) {
	data, err := json.Marshal(r)
	if err != nil {
		fmt.Errorf("Error marshalling response: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, string(data))
}

type GenericResponse struct {
	Err string
}

func (r *GenericResponse) WriteResponse(w http.ResponseWriter) {
	if r.Err != "" {
		w.WriteHeader(http.StatusBadRequest)
	}
	data, err := json.Marshal(r)
	if err != nil {
		fmt.Errorf("Error marshalling response: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, string(data))
}

type GenericRequest struct {
	Name string
}

type MountResponse struct {
	Mountpoint string
	Err        string
}

func (r *MountResponse) WriteResponse(w http.ResponseWriter) {
	if r.Err != "" {
		w.WriteHeader(http.StatusBadRequest)
	}
	data, err := json.Marshal(r)
	if err != nil {
		fmt.Errorf("Error marshalling Get response: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, string(data))
}

type VolumeMetadata struct {
	Name       string
	Mountpoint string
}

type GetResponse struct {
	Volume Volume
	Err    string
}
type DockerGetResponse struct {
	Volume VolumeMetadata
	Err    string
}

type Volume struct {
	Name    string
	Backend Backend
}

type GetConfigResponse struct {
	VolumeConfig map[string]interface{}
	Err          string
}

func (r *GetResponse) WriteResponse(w http.ResponseWriter) {
	if r.Err != "" {
		w.WriteHeader(http.StatusBadRequest)
	}
	data, err := json.Marshal(r)
	if err != nil {
		fmt.Errorf("Error marshalling Get response: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, string(data))
}
func (r *DockerGetResponse) WriteResponse(w http.ResponseWriter) {
	if r.Err != "" {
		w.WriteHeader(http.StatusBadRequest)
	}
	data, err := json.Marshal(r)
	if err != nil {
		fmt.Errorf("Error marshalling DockerGetResponse: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, string(data))
}

type ListResponse struct {
	Volumes []VolumeMetadata
	Err     string
}

func (r *ListResponse) WriteResponse(w http.ResponseWriter) {
	if r.Err != "" {
		w.WriteHeader(http.StatusBadRequest)
	}
	data, err := json.Marshal(r)
	if err != nil {
		fmt.Errorf("Error marshalling Get response: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, string(data))
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

type FlexVolumeDetachRequest struct {
	Name string `json:"name"`
}
