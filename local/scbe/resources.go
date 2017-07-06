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
	LogicalSize                        int    `json:"logical_size"`
	PhysicalFree                       int    `json:"physical_free"`
	LogicalFree                        int    `json:"logical_free"`
	TotalCapacity                      int    `json:"total_capacity"`
	UsedCapacity                       int    `json:"used_capacity"`
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
	ScsiIdentifier     string `json:"scsi_identifier"`
	ArrayType          string `json:"array_type"`
	Array              string `json:"array"`
	ArrayName          string `json:"array_name"`
	Id                 string `json:"id"`
	PoolName           string `json:"pool_name"`
	MaxExtendable_size int    `json:"max_extendable_size"`
	ServiceCompliance  string `json:"service_compliance"`
	DomainName         string `json:"domain_name"`
	ServiceName        string `json:"service_name"`
	ContainerName      string `json:"container_name"`
	ServiceId          string `json:"service_id"`
	ContainerId        string `json:"container_id"`
	StorageModel       string `json:""`
	VolumeId           string `json:"volume_id"`
	Name               string `json:"name"`
	LogicalCapacity    int    `json:"logical_capacity"`
	PhysicalCapacity   int    `json:"physical_capacity"`
	UsedCapacity       int    `json:"used_capacity"`
	LastUpdateTime     string `json:"last_update_time"`
	IsPendingDeletion  bool   `json:"is_pending_deletion"`
	Pool               int    `json:"pool"`

	// A9000 related attrs
	Serial        string `json:"serial"`
	CgId          string `json:"cg_id"`
	Snapshot_time string `json:"snapshot_time"` //not sure about the type
	Compressed    bool   `json:"compressed"`
	//Ratio                   int `json:"ratio"` // for A9000 its int but for v9000 its string "N/A", so comment out to avoid marshaling issue
	Saving                  int    `json:"cg_id"`
	ThinProvisioningSavings string `json:"thin_provisioning_savings"`
	EstCompressionFactor    string `json:"est_compression_factor"`
	PerfClass               string `json:"perf_class"`

	// V9000 related attrs
	Status           string `json:"status"`
	CopyCount        int    `json:"copy_count"`
	FcMapCount       int    `json:"fc_map_count"`
	FastWriteState   string `json:"fast_write_state"`
	OwnerType        string `json:"owner_type"`
	FlashcopyTime    string `json:"flashcopy_time"`
	ThrottleId       string `json:"throttle_id"`
	IOPsLimit        string `json:"IOPs_limit"`
	BandwidthLimitMB string `json:"bandwidth_limit_MB"`
	IoGrp            int    `json:"io_grp"`
}

/*
Example of SCBE response on A9000 volume
    {
        "scsi_identifier": "6001738CFC9035E8000000000091D081",
        "array_type": "2810XIV",
        "array": "9835-415-6013800",
        "array_name": "XIV Gen4d-67a",
        "id": "6001738CFC9035E8000000000091D081",
        "pool_name": "gold_pool_ubiquity_scbe_9.151.162.17",
        "pool_id": "5b819800004",
        "max_extendable_size": 108364427776,
        "service_compliance": "True",
        "domain_name": "ilyak_domain",
        "service_name": "gold",
        "container_name": "Default_Space",
        "service_id": "cc4c1254-d551-4a51-81f5-4732f9e0833e",
        "container_id": "23c380fc-fe1e-4c02-9d1e-0781e9442394",
        "storage_model": "FlashSystem A9000R",
        "ratio": 0, // on SVC\V9000 it returns string "N/A"
        "update_status": "",
        "volume_id": "478319e0012e",
        "name": "u_ubiquity_instance1_vol1",
        "logical_capacity": 5000000000,
        "physical_capacity": 5117050880,
        "used_capacity": 0,
        "last_update_time": "2017-06-16T11:14:16.876891",
        "is_pending_deletion": false,
        "pool": 28, // was below
    // From here its specific A9000
        "serial": "",
        "cg_id": "0",
        "snapshot_time": null,
        "compressed": true,
        "saving": 0,
        "thin_provisioning_savings": "100",
        "est_compression_factor": "1.00",
        "pool": 274,
        "perf_class": null
    // From here its specific V9000 related
        "status": "online",
        "copy_count": 1,
        "fc_map_count": 0,
        "fast_write_state": "empty",
        "owner_type": null,
        "flashcopy_time": null,
        "throttle_id": null,
        "IOPs_limit": null,
        "bandwidth_limit_MB": null,
        "io_grp": 1
    }
*/

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
	Name             string
	Wwn              string
	Profile          string
	StorageType      string
	StorageName      string
	PoolName         string
	LogicalCapacity  string
	PhysicalCapacity string
	UsedCapacity     string
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
