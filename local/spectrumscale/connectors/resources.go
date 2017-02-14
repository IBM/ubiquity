package connectors

type GetClusterResponse struct {
	Cluster Cluster `json:"cluster"`
	Status  Status  `json:"status"`
}
type Cluster struct {
	CesSummary     CesSummary        `json:"cesSummary"`
	ClusterSummary ClusterSummary    `json:clusterSummary`
	CnfsSummary    CnfsSummary       `json:"cnfsSummary"`
	Links          map[string]string `json:"links"`
	Nodes          []ClusterNode     `json:"nodes"`
}

type CesSummary struct {
	EnabledServices string `json:enabledServices`
	AddressPolicy   string `json:"addressPolicy"`
	CesSharedRoot   string `json:"cesSharedRoot"`
	LogLevel        string `json:"logLevel"`
}
type ClusterSummary struct {
	ClusterID       string `json:clusterId`
	ClusterName     string `json:"clusterName"`
	PrimaryServer   string `json:"primaryServer"`
	RcpPath         string `json:"rcpPath"`
	RcpSudoWrapper  string `json:"rcpSudoWrapper"`
	RepositoryType  string `json:"repositoryType"`
	RshPath         string `json:"rshPath"`
	SecondaryServer string `json:"secondaryServer"`
	UIDDomain       string `json:"uidDomain"`
}

type CnfsSummary struct {
	CNFSGanesha        string `json:"cnfsGanesha"`
	CNFSMonitorEnabled string `json:"cnfsMonitorEnabled"`
	CNFSMountdPort     string `json:"cnfsMountdPort"`
	CNFSNFSDprocs      string `json:"cnfsNFSDprocs"`
	CNFSReboot         string `json:"cnfsReboot"`
	CNFSSharedReboot   string `json:"cnfsSharedReboot"`
}

type ClusterNode struct {
	AdminLoginName string            `json:"adminLoginName"`
	AdminNodeName  string            `json:"adminNodeName"`
	CESNode        CESNode           `json:"cesNode"`
	CNFSNode       CNFSNode          `json:"cnfsNode"`
	DaemonNodeName string            `json:"daemonNodeName"`
	Designation    string            `json:"designation"`
	IPAddress      string            `json:"ipAddress"`
	Links          map[string]string `json:"links"`
	NodeNumber     string            `json:"nodeNumber"`
	OtherNodeRoles string            `json:"otherNodeRoles"`
}

type CESNode struct {
	CESGroup  string `json:"cesGroup"`
	CESIPList string `json:"cesIpList"`
	CESState  string `json:"cesState"`
}
type CNFSNode struct {
	CNFSGroupID string `json:"cnfsGroupID"`
	CNFSIPList  string `json:"cnfsIpList"`
	CNFSState   string `json:"cnfsState"`
}

type Status struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type GetFilesystemResponse struct {
	FileSystems []FileSystem `json:"filesystems"`
	Status      Status       `json:"status"`
}

type FileSystem struct {
	ACLSemantics               string            `json:"ACLSemantics"`
	DMAPIEnabled               string            `json:"DMAPIEnabled"`
	UID                        string            `json:"UID"`
	AdditionalMountOptions     string            `json:"additionalMountOptions"`
	AutomaticMountOption       string            `json:"automaticMountOption"`
	BlockAllocationType        string            `json:"blockAllocationType"`
	BlockSize                  string            `json:"blockSize"`
	CreateTime                 string            `json:"createTime"`
	DefaultMetadataReplicas    string            `json:"defaultMetadataReplicas"`
	DefaultMountPoint          string            `json:"defaultMountPoint"`
	DefaultQuotasEnabled       string            `json:"defaultQuotasEnabled"`
	Disks                      string            `json:"disks"`
	Encryption                 string            `json:"encryption"`
	ExactMTime                 string            `json:"exactMTime"`
	FastEAEnabled              string            `json:"FastEAEnabled"`
	FileLockingSemantics       string            `json:"fileLockingSemantics"`
	FilesetdfEnabled           string            `json:"filesetdfEnabled"`
	FilesystemHighestSupported string            `json:"filesystemHighestSupported"`
	FilesystemName             string            `json:"filesystemName"`
	FilesystemVersion          string            `json:"filesystemVersion"`
	FilesystemVersionLocal     string            `json:"filesystemVersionLocal"`
	FilesystemVersionManager   string            `json:"filesystemVersionManager"`
	FilesystemVersionOriginal  string            `json:"filesystemVersionOriginal"`
	IndirectBlockSize          string            `json:"indirectBlockSize"`
	InodeSize                  string            `json:"inodeSize"`
	Is4KAligned                string            `json:"is4KAligned"`
	Links                      map[string]string `json:"links"`
	LogReplicas                string            `json:"logReplicas"`
	LogfileSize                string            `json:"logfileSize"`
	MaxDataReplicas            string            `json:"maxDataReplicas"`
	MaxMetadataReplicas        string            `json:"maxMetadataReplicas"`
	MaxNumberOfInodes          string            `json:"maxNumberOfInodes"`
	MaxSnapshotId              string            `json:"maxSnapshotId"`
	MinFragmentSize            string            `json:"minFragmentSize"`
	MountPriority              string            `json:"mountPriority"`
	NumNodes                   string            `json:"numNodes"`
	OtherPools                 []Pool            `json:"otherPools"`
	PerfilesetQuotas           string            `json:"perfilesetQuotas"`
	QuotasAccountingEnable     string            `json:"quotasAccountingEnable"`
	QuotasEnforced             string            `json:"quotasEnforced"`
	RapidRepairEnabled         string            `json:"rapidRepairEnabled"`
	StoragePools               string            `json:"storagePools"`
	StrictReplication          string            `json:"strictReplication"`
	SuppressATime              string            `json:"suppressATime"`
	WriteCacheThreshold        string            `json:"writeCacheThreshold"`
}
type Pool struct {
	BlockSize       string `json:"blockSize"`
	MinFragmentSize string `json:"minFragmentSize"`
}

type GetFilesetResponse struct {
	Filesets []Fileset `json:"filesets"`
	Status   Status    `json:"status"`
}

type Fileset struct {
	AFM    AFM               `json:"afm"`
	Config FilesetConfig     `json:"config"`
	Links  map[string]string `json:"links"`
	State  FilesetState      `json:"state"`
}

type AFM struct {
	AFMAsyncDelay                string `json:"afmAsyncDelay"`
	AFMDirLookupRefreshInterval  string `json:"afmDirLookupRefreshInterval"`
	AFMDirOpenRefreshInterval    string `json:"afmDirOpenRefreshInterval"`
	AFMEnableAutoEviction        string `json:"afmEnableAutoEviction"`
	AFMExpirationTimeout         string `json:"afmExpirationTimeout"`
	AFMFileLookupRefreshInterval string `json:"afmFileLookupRefreshInterval"`
	AFMMode                      string `json:"afmMode"`
	AFMNumFlushThreads           string `json:"afmNumFlushThreads"`
	AFMParallelReadChunkSize     string `json:"afmParallelReadChunkSize"`
	AFMParallelReadThreshold     string `json:"afmParallelReadThreshold"`
	AFMParallelWriteChunkSize    string `json:"afmParallelWriteChunkSize"`
	AFMParallelWriteThreshold    string `json:"afmParallelWriteThreshold"`
	AFMPrefetchThreshold         string `json:"afmPrefetchThreshold"`
	AFMPrimaryID                 string `json:"afmPrimaryID"`
	AFMRPO                       string `json:"afmRPO"`
	AFMShowHomeSnapshots         string `json:"afmShowHomeSnapshots"`
	AFMTarget                    string `json:"afmTarget"`
}
type FilesetConfig struct {
	Comment              string `json:"comment"`
	FilesetName          string `json:"filesetName"`
	FilesystemName       string `json:"filesystemName"`
	IAMMode              string `json:"iamMode"`
	INodeSpace           string `json:"inodeSpace"`
	MaxNumInodes         string `json:"maxNumInodes"`
	AllocInodes          string `json:"allocInodes"`
	Owner                string `json:"owner"`
	Path                 string `json:"path"`
	PermissionChangeMode string `json:"permissionChangeMode"`
	Permissions          string `json:"permissions"`
}
type FilesetState struct {
	AFMState          string `json:"afmState"`
	Created           string `json:"created"`
	DataInKB          string `json:"dataInKB"`
	FreeInodes        string `json:"freeInodes"`
	ID                string `json:"id"`
	InodeSpaceMAsk    string `json:"inodeSpaceMask"`
	Inodes            string `json:"inodes"`
	IsInodeSpaceOwner string `json:"isInodeSpaceOwner"`
	ParentID          string `json:"parentId"`
	RootInode         string `json:"rootInode"`
	SnapID            string `json:"snapID"`
	Status            string `json:"status"`
}

type CreateFilesetResquest struct {
	AFM    AFM           `json:"afm"`
	Config FilesetConfig `json:"config"`
}
type CreateFilesetResponse struct {
	Status Status `json:"status"`
}

type DeleteFilesetResponse struct {
	Status Status `json:"status"`
}

type GetQuotaResponse struct {
	Links  map[string]string `json:"links"`
	Quotas []Quota           `json:"quotas"`
	Status Status            `json:"status"`
}
type SetQuotaRequest struct {
	BlockGracePeriod string `json:"blockGracePeriod"`
	BlockHardLimit   string `json:"blockHardLimit"`
	BlockSoftLimit   string `json:"blockSoftLimit"`
	FilesGracePeriod string `json:"filesGracePeriod"`
	FilesHardLimit   string `json:"filesHardLimit"`
	FilesSoftLimit   string `json:"filesSoftLimit"`
	FilesetName      string `json:"filesetName"`
	FilesystemName   string `json:"filesystemName"`
	ObjectName       string `json:"objectName"`
	OperationType    string `json:"operationType"`
	QuotaType        string `json:"quotaType"`
}

type SetQuotaResponse struct {
	Status Status `json:"status"`
}
type Quota struct {
	BlockGrace     string `json:"blockGrace"`
	BlockInDoubt   string `json:"blockInDoubt"`
	BlockLimit     string `json:"blockLimit"`
	BlockQuota     string `json:"blockQuota"`
	BlockUsage     string `json:"blockUsage"`
	DefQuota       string `json:"defQuota"`
	FilesGrace     string `json:"filesGrace"`
	FilesInDoubt   string `json:"filesInDoubt"`
	FilesLimit     string `json:"filesLimit"`
	FilesQuota     string `json:"filesQuota"`
	FilesUsage     string `json:"filesUsage"`
	FilesetId      string `json:"filesetId"`
	FilesetName    string `json:"filesetName"`
	FilesystemName string `json:"filesystemName"`
	ObjectID       string `json:"objectID"`
	ObjectName     string `json:"objectName"`
	QuotaType      string `json:"quotaType"`
}

type GetNodesResponse struct {
	Nodes  []Node `json:"nodes"`
	Status Status `json:"status"`
}
type Node struct {
	AdminLogin        string            `json:"adminLogin"`
	CESNode           string            `json:"cesNode"`
	ClientNode        string            `json:"clientNode"`
	CNFSInterface     string            `json:"cnfsInterface"`
	CNFSState         string            `json:"cnfsState"`
	DaemonIPAddress   string            `json:"daemonIPAddress"`
	DaemonInterface   string            `json:"daemonInterface"`
	DesignatedLicence string            `json:"designatedLicense"`
	GatewayNode       string            `json:"gatewayNode"`
	GPFSState         string            `json:"gpfsState"`
	Links             map[string]string `json:"links"`
	ManagerNode       string            `json:"managerNode"`
	NodeName          string            `json:"nodeName"`
	NodeNumber        string            `json:"nodeNumber"`
	OSName            string            `json:"osName"`
	ProductVersion    string            `json:"productVersion"`
	QuorumNode        string            `json:"quorumNode"`
	SNMPNode          string            `json:"snmpNode"`
}
