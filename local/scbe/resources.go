package scbe

// go standard for all the structures in the project.

type CredentialInfo struct {
	UserName string `json:"username"`
	Password string `json:"password"`
	Group    string `json:"group"`
}

type ConnectionInfo struct {
	CredentialInfo CredentialInfo
	Port           int
	ManagementIP   string
	VerifySSL      bool
}

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
