package spectrum

type GetClusterResponse struct {
	Cluster Cluster `json:"cluster"`
	Status  Status  `json:"status"`
}
type Cluster struct {
	CesSummary     CesSummary        `json:"cesSummary"`
	ClusterSummary ClusterSummary    `json:clusterSummary`
	CnfsSummary    CnfsSummary       `json:"cnfsSummary"`
	Links          map[string]string `json:"links"`
	Nodes          []Node            `json:"nodes"`
}

type CesSummary struct {
	EnabledServices string `json:enabledServices`
	AddressPolicy   string `json:"addressPolicy"`
	CesSharedRoot   string `json:"cesSharedRoot"`
	LogLevel        uint16 `json:"logLevel"`
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

type Node struct {
	AdminLoginName string            `json:"adminLoginName"`
	AdminNodeName  string            `json:"adminNodeName"`
	CESNode        CESNode           `json:"cesNode"`
	CNFSNode       CNFSNode          `json:"cnfsNode"`
	DaemonNodeName string            `json:"daemonNodeName"`
	Designation    string            `json:"designation"`
	IPAddress      string            `json:"ipAddress"`
	Links          map[string]string `json:"links"`
	NodeNumber     uint64            `json:"nodeNumber"`
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

type GetFileSystemResponse struct {
	FileSystems []FileSystem `json:"filesystems"`
	Status      Status       `json:"status"`
}

type FileSystem struct {
	ACLSemantics               string            `json:"ACLSemantics"`
	DMAPIEnabled               bool              `json:"DMAPIEnabled"`
	UID                        string            `json:"UID"`
	AdditionalMountOptions     string            `json:"additionalMountOptions"`
	AutomaticMountOption       bool              `json:"automaticMountOption"`
	BlockAllocationType        string            `json:"blockAllocationType"`
	BlockSize                  string            `json:"blockSize"`
	CreateTime                 string            `json:"createTime"`
	DefaultMetadataReplicas    uint16            `json:"defaultMetadataReplicas"`
	DefaultMountPoint          string            `json:"defaultMountPoint"`
	DefaultQuotasEnabled       string            `json:"defaultQuotasEnabled"`
	Disks                      string            `json:"disks"`
	Encryption                 bool              `json:"encryption"`
	ExactMTime                 bool              `json:"exactMTime"`
	FastEAEnabled              bool              `json:"FastEAEnabled"`
	FileLockingSemantics       string            `json:"fileLockingSemantics"`
	FilesetdfEnabled           bool              `json:"filesetdfEnabled"`
	FilesystemHighestSupported string            `json:"filesystemHighestSupported"`
	FilesystemName             string            `json:"filesystemName"`
	FilesystemVersion          string            `json:"filesystemVersion"`
	FilesystemVersionLocal     string            `json:"filesystemVersionLocal"`
	FilesystemVersionManager   string            `json:"filesystemVersionManager"`
	FilesystemVersionOriginal  string            `json:"filesystemVersionOriginal"`
	IndirectBlockSize          uint64            `json:"indirectBlockSize"`
	InodeSize                  uint64            `json:"inodeSize"`
	Is4KAligned                bool              `json:"is4KAligned"`
	Links                      map[string]string `json:"links"`
	LogReplicas                uint16            `json:"logReplicas"`
	LogfileSize                uint64            `json:"logfileSize"`
	MaxDataReplicas            uint16            `json:"maxDataReplicas"`
	MaxMetadataReplicas        uint16            `json:"maxMetadataReplicas"`
	MaxNumberOfInodes          uint16            `json:"maxNumberOfInodes"`
	MaxSnapshotId              uint16            `json:"maxSnapshotId"`
	MinFragmentSize            uint64            `json:"minFragmentSize"`
	MountPriority              string            `json:"mountPriority"`
	NumNodes                   uint16            `json:"numNodes"`
	OtherPools                 []Pool            `json:"otherPools"`
	PerfilesetQuotas           bool              `json:"perfilesetQuotas"`
	QuotasAccountingEnable     string            `json:"quotasAccountingEnable"`
	QuotasEnforced             string            `json:"quotasEnforced"`
	RapidRepairEnabled         bool              `json:"rapidRepairEnabled"`
	StoragePools               string            `json:"storagePools"`
	StrictReplication          string            `json:"strictReplication"`
	SuppressATime              bool              `json:"suppressATime"`
	WriteCacheThreshold        uint64            `json:"writeCacheThreshold"`
}
type Pool struct {
	BlockSize       uint64 `json:"blockSize"`
	MinFragmentSize uint64 `json:"minFragmentSize"`
}

type GetFileSetResponse struct {
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
	AFMAsyncDelay                uint64 `json:"afmAsyncDelay"`
	AFMDirLookupRefreshInterval  uint64 `json:"afmDirLookupRefreshInterval"`
	AFMDirOpenRefreshInterval    uint64 `json:"afmDirOpenRefreshInterval"`
	AFMEnableAutoEviction        bool   `json:"afmEnableAutoEviction"`
	AFMExpirationTimeout         uint64 `json:"afmExpirationTimeout"`
	AFMFileLookupRefreshInterval uint64 `json:"afmFileLookupRefreshInterval"`
	AFMMode                      string `json:"afmMode"`
	AFMNumFlushThreads           string `json:"afmNumFlushThreads"`
	AFMParallelReadChunkSize     uint64 `json:"afmParallelReadChunkSize"`
	AFMParallelReadThreshold     uint64 `json:"afmParallelReadThreshold"`
	AFMParallelWriteChunkSize    uint64 `json:"afmParallelWriteChunkSize"`
	AFMParallelWriteThreshold    uint64 `json:"afmParallelWriteThreshold"`
	AFMPrefetchThreshold         uint16 `json:"afmPrefetchThreshold"`
	AFMPrimaryID                 string `json:"afmPrimaryID"`
	AFMRPO                       uint64 `json:"afmRPO"`
	AFMShowHomeSnapshots         bool   `json:"afmShowHomeSnapshots"`
	AFMTarget                    string `json:"afmTarget"`
}
type FilesetConfig struct {
	Comment              string `json:"comment"`
	FilesetName          string `json:"filesetName"`
	FilesystemName       string `json:"filesystemName"`
	IAMMode              string `json:"iamMode"`
	INodeSpace           string `json:"inodeSpace"`
	MaxNumInodes         uint16 `json:"maxNumInodes"`
	AllocInodes          string `json:"allocInodes"`
	Owner                string `json:"owner"`
	Path                 string `json:"path"`
	PermissionChangeMode string `json:"permissionChangeMode"`
	Permissions          string `json:"permissions"`
}
type FilesetState struct {
	AFMState          string `json:"afmState"`
	Created           string `json:"created"`
	DataInKB          uint64 `json:"dataInKB"`
	FreeInodes        string `json:"freeInodes"`
	ID                string `json:"id"`
	InodeSpaceMAsk    string `json:"inodeSpaceMask"`
	Inodes            string `json:"inodes"`
	IsInodeSpaceOwner bool   `json:"isInodeSpaceOwner"`
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
