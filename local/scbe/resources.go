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

/*
'{"scsi_identifier":"6001738CFC9035E8000000000001348E",'\
    '"array_type":"2810XIV",'\
    '"array":"9835-415-6013800",' \
    '"array_name":"a9000",' \
    '"id":"6001738CFC9035E8000000000001348E",' \
    '"pool_name":"s1pool2",' \
    '"pool_id":"701a18900038",' \
    '"max_extendable_size":113364427776,' \
    '"service_compliance":"True",' \
    '"domain_name":"roei_domain",' \
    '"service_name":"s1",' \
    '"container_name":"Default_Space",' \
    '"service_id":"145d5b94-d573-45da-abac-6d625cc6970d",' \
    '"container_id":"5bba448a-6b9a-4d91-a369-758194a88c42",' \
    '"storage_model":"FlashSystem A9000R","' \
    'volume_id":"d55718f00053",' \
    '"name":"f_manual_vol",' \
    '"logical_capacity":10000000000,' \
    '"physical_capacity":10234101760,' \
    '"used_capacity":0,"last_update_time":"2016-10-13T10:45:52.588271",' \
    '"is_pending_deletion":false,' \
    '"serial":"",' \
    '"cg_id":"0",' \
    '"snapshot_time":null,' \
    '"compressed":true,' \
    '"ratio":0,' \
    '"saving":0,' \
    '"thin_provisioning_savings":"100",' \
    '"est_compression_factor":"1.00",' \
    '"pool":49,' \
    '"perf_class":null}
*/

type ScbeCreateVolumePostParams struct {
	Service  string `json:"service"`
	Name     string `json:"name"`
	Size     int    `json:"size"`
	SizeUnit string `json:"size_unit"`
}

type ScbeVolumeInfo struct {
	Name        string
	Wwn         string
	ServiceName string
	// TODO later on we will want also size and maybe other stuff
}
