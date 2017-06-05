package scbe

// go standard for all the structures in the project.

type LoginResponse struct {
	Token string `json:"token"`
}

type ScbeStorageService struct {
	Id                                 string `json:"id"`
	UniqueIdentifier                   string `json:"unique_identifier"`
	Name                               string `json:"name"`
	Description                        string `json:"description"`
	Container                          string `json:"container"`
	CapabilityValues                   string `json:"capability_values"`
	Type                               string `json:"type"`
	PhysicalSize                       int    `json:""`
	Logical_size                       int    `json:"logical_size"`
	Physical_free                      int    `json:"physical_free"`
	Logical_free                       int    `json:"logical_free"`
	Total_capacity                     int    `json:"total_capacity"`
	Used_capacity                      int    `json:"used_capacity"`
	MaxResourceLogicalFree             int    `json:"max_resource_logical_free"`
	MaxResourceFreeSizeForProvisioning int    `json:"max_resource_free_size_for_provisioning"`
	NumVolumes                         int    `json:"num_volumes"`
	HasAdmin                           bool   `json:"has_admin"`
	QosMaxIops                         int    `json:"qos_max_iops"`
	QosMaxMbps                         int    `json:"qos_max_mbps"`
}

/*
Example of services response from SCBE
[
  {
    "id": "cc4c1254-d551-4a51-81f5-ffffffffffff",
    "unique_identifier": "cc4c1254-d551-4a51-81f5-ffffffffffff",
    "name": "gold",
    "description": " ",
    "container": "23c380fc-fe1e-4c02-9d1e-ffffffffffff",
    "capability_values": "",
    "type": "regular",
    "physical_size": 413457711104,
    "logical_size": 413457711104,
    "physical_free": 310093283328,
    "logical_free": 310093283328,
    "total_capacity": 413457711104,
    "used_capacity": 103364427776,
    "max_resource_logical_free": 310093283328,
    "max_resource_free_size_for_provisioning": 310093283328,
    "num_volumes": 0,
    "has_admin": true,
    "qos_max_iops": 0,
    "qos_max_mbps": 0
  }
]
*/

type ScbeResponseVolume struct {
	ScsiIdentifier          string `json:"scsi_identifier"`
	ArrayType               string `json:"array_type"`
	Array                   string `json:"array"`
	ArrayName               string `json:"array_name"`
	Id                      string `json:"id"`
	PoolName                string `json:"pool_name"`
	MaxExtendable_size      int    `json:"max_extendable_size"`
	ServiceCompliance       string `json:"service_compliance"`
	DomainName              string `json:"domain_name"`
	ServiceName             string `json:"service_name"`
	ContainerName           string `json:"container_name"`
	ServiceId               string `json:"service_id"`
	ContainerId             string `json:"container_id"`
	StorageModel            string `json:""`
	VolumeId                string `json:"volume_id"`
	Name                    string `json:"name"`
	LogicalCapacity         int    `json:"logical_capacity"`
	PhysicalCapacity        int    `json:"physical_capacity"`
	UsedCapacity            int    `json:"used_capacity"`
	LastUpdateTime          string `json:"last_update_time"`
	IsPendingDeletion       bool   `json:"is_pending_deletion"`
	Serial                  string `json:"serial"`
	CgId                    string `json:"cg_id"`
	Snapshot_time           string `json:"snapshot_time"` //not sure about the type
	Compressed              bool   `json:"compressed"`
	Ratio                   int    `json:"ratio"`
	Saving                  int    `json:"cg_id"`
	ThinProvisioningSavings string `json:"thin_provisioning_savings"`
	EstCompressionFactor    string `json:"est_compression_factor"`
	Pool                    int    `json:"pool"`
	PerfClass               string `json:"perf_class"`
}

type ScbeCreateVolumePostParams struct {
	Service  string `json:"service"`
	Name     string `json:"name"`
	Size     int    `json:"size"`
	SizeUnit string `json:"size_unit"`
}

type ScbeMapVolumePostParams struct {
	VolumeId string `json:"volume_id"`
	HostId   int    `json:"host_id"`
}

// Its the same as ScbeMapVolumePostParams but later on we may add more attr to the map struct
type ScbeUnMapVolumePostParams struct {
	VolumeId string `json:"volume_id"`
	HostId   int    `json:"host_id"`
}

type ScbeResponseMapping struct {
	Id        int    `json:"id"`
	Volume    string `json:"volume"`
	Host      int    `json:"host"`
	LunNumber int    `json:"lun_number"`
}

type ScbeResponseMappings struct {
	Mappings []ScbeResponseMapping `json:"mappings"`
}

/*
"{\"mappings\":[{\"id\":1,\"volume\":\"6001738CFC9035E8000000000091XXXX\",\"host\":2,\"lun_number\":1}]}"
*/

type ScbeVolumeInfo struct {
	Name        string
	Wwn         string
	ServiceName string
	// TODO later on we will want also size and maybe other stuff
}

type ScbeResponseHost struct {
	Id             int    `json:"id"`
	ArrayType      string `json:"array_type"`
	Array          string `json:"array"`
	HostId         string `json:"host_id"`
	Name           string `json:"name"`
	StorageCluster string `json:"storage_cluster"`
	PhysicalHost   int    `json:"physical_host"`
}
