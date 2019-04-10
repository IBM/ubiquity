package utils

/*
Example multipath output:
{
   "major_version": 0,
   "minor_version": 1,
   "maps": [{
      "name" : "mpathp",
      "uuid" : "360050768029b8168e000000000006247",
      "sysfs" : "dm-3",
      "failback" : "immediate",
      "queueing" : "5 chk",
      "paths" : 0,
      "write_prot" : "rw",
      "dm_st" : "active",
      "features" : "0",
      "hwhandler" : "0",
      "action" : "create",
      "path_faults" : 1,
      "vend" : "IBM     ",
      "prod" : "2145            ",
      "rev" : "0000",
      "switch_grp" : 0,
      "map_loads" : 1,
      "total_q_time" : 26,
      "q_timeouts" : 1,
      "path_groups": [{
         "selector" : "round-robin 0",
         "pri" : 0,
         "dm_st" : "active",
         "group" : 1,
         "paths": [{
            "dev" : "sdb",
            "dev_t" : "8:16",
            "dm_st" : "failed",
            "dev_st" : "running",
            "chk_st" : "faulty",
            "checker" : "tur",
            "pri" : 50,
            "host_wwnn" : "[undef]",
            "target_wwnn" : "iqn.1986-03.com.ibm:2145.v7k60.node1",
            "host_wwpn" : "[undef]",
            "target_wwpn" : "[undef]",
            "host_adapter" : "9.115.240.253"
         }]
      }]
   }]
}
*/

type MultipathOutputAll struct {
	Maps []*MultipathDevice `json:"maps"`
}

type MultipathOutput struct {
	Map *MultipathDevice `json:"map"`
}

type MultipathDevice struct {
	Name       string                      `json:"name"`
	Uuid       string                      `json:"uuid"`
	Sysfs      string                      `json:"sysfs"`
	PathFaults int                         `json:"path_faults"`
	Vend       string                      `json:"vend"`
	Prod       string                      `json:"prod"`
	PathGroups []*MultipathDevicePathGroup `json:"path_groups"`
}

type MultipathDevicePathGroup struct {
	Selector string                 `json:"selector"`
	Pri      int                    `json:"pri"`
	DmSt     string                 `json:"dm_st"`
	Group    int                    `json:"group"`
	Paths    []*MultipathDevicePath `json:"paths"`
}

type MultipathDevicePath struct {
	Dev   string `json:"dev"`
	DevT  string `json:"dev_t"`
	DmSt  string `json:"dm_st"`
	DevSt string `json:"dev_st"`
	ChkSt string `json:"chk_st"`
}
