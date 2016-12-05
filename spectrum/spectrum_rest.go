package spectrum

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.ibm.com/almaden-containers/ubiquity/model"
	"github.ibm.com/almaden-containers/ubiquity/utils"
)

type spectrum_rest struct {
	logger     *log.Logger
	httpClient *http.Client
	endpoint   string
}

func NewSpectrumRest(logger *log.Logger, opts map[string]interface{}) Spectrum {
	endpoint, _ := opts["endpoint"]
	return &spectrum_rest{logger: logger, httpClient: &http.Client{}, endpoint: endpoint.(string)}
}
func (s *spectrum_rest) GetClusterId() (string, error) {
	getClusterURL := utils.FormatURL(s.endpoint, "scalemgmt/v1/cluster")
	getClusterResponse := GetClusterResponse{}
	cidResponse, err := s.doHTTP(getClusterURL, "GET", getClusterResponse)
	if err != nil {
		s.logger.Printf("error in executing remote call: %v", err)
		return "", err
	}

	getClusterResponse = cidResponse.(GetClusterResponse)

	return getClusterResponse.Cluster.ClusterSummary.ClusterID, nil
}

func (s *spectrum_rest) IsFilesystemMounted(filesystemName string) (bool, error) {
	getNodesURL := utils.FormatURL(s.endpoint, "scalemgmt/v1/nodes")

	getNodesResponse := GetNodesResponse{}
	nodesResponse, err := s.doHTTP(getNodesURL, "GET", getNodesResponse)
	if err != nil {
		s.logger.Printf("error in executing remote call: %v", err)
		return false, err
	}

	getNodesResponse = nodesResponse.(GetNodesResponse)

	currentNode, _ := os.Hostname()
	s.logger.Printf("spectrum rest Client: node name: %s\n", currentNode)
	for _, node := range getNodesResponse.Nodes {
		if node.NodeName == currentNode {
			return true, nil
		}
	}

	return false, nil
}

func (s *spectrum_rest) MountFileSystem(filesystemName string) error {
	return nil
}

func (s *spectrum_rest) ListFilesystems() ([]string, error) {
	listFilesystemsURL := utils.FormatURL(s.endpoint, "scalemgmt/v1/filesystems")
	getFilesystemResponse := GetFilesystemResponse{}
	fsResponse, err := s.doHTTP(listFilesystemsURL, "GET", getFilesystemResponse)
	if err != nil {
		s.logger.Printf("error in executing remote call: %v", err)
		return nil, err
	}

	getFilesystemResponse = fsResponse.(GetFilesystemResponse)

	fsNumber := len(getFilesystemResponse.FileSystems)
	filesystems := make([]string, fsNumber)
	for i := 0; i < fsNumber; i++ {
		filesystems[i] = getFilesystemResponse.FileSystems[i].FilesystemName
	}
	return filesystems, nil
}

func (s *spectrum_rest) GetFilesystemMountpoint(filesystemName string) (string, error) {
	getFilesystemURL := utils.FormatURL(s.endpoint, fmt.Sprintf("scalemgmt/v1/filesystems/%s", filesystemName))
	getFilesystemResponse := GetFilesystemResponse{}

	fsResponse, err := s.doHTTP(getFilesystemURL, "GET", getFilesystemResponse)
	if err != nil {
		s.logger.Printf("error in executing remote call: %v", err)
		return "", err
	}

	getFilesystemResponse = fsResponse.(GetFilesystemResponse)
	return getFilesystemResponse.FileSystems[0].DefaultMountPoint, nil
}

func (s *spectrum_rest) CreateFileset(filesystemName string, filesetName string, opts map[string]interface{}) error {
	return nil
}

func (s *spectrum_rest) DeleteFileset(filesystemName string, filesetName string) error {
	deleteFilesetURL := utils.FormatURL(s.endpoint, fmt.Sprintf("scalemgmt/v1/filesets/%s/filesystemName=%s&qosClass=other", filesetName, filesystemName))
	deleteFilesetResponse := DeleteFilesetResponse{}
	_, err := s.doHTTP(deleteFilesetURL, "DELETE", deleteFilesetResponse)
	if err != nil {
		s.logger.Printf("Error in delete remote call")
		return err
	}
	//TODO check the response message for errors
	return nil
}

func (s *spectrum_rest) LinkFileset(filesystemName string, filesetName string) error { return nil }

func (s *spectrum_rest) UnlinkFileset(filesystemName string, filesetName string) error { return nil }

func (s *spectrum_rest) ListFilesets(filesystemName string) ([]model.VolumeMetadata, error) {
	listFilesetURL := utils.FormatURL(s.endpoint, "scalemgmt/v1/filesets")
	listFilesetResponse := GetFilesetResponse{}
	lfsResponse, err := s.doHTTP(listFilesetURL, "GET", listFilesetResponse)
	if err != nil {
		s.logger.Printf("error in processing remote call %v", err)
		return nil, err
	}
	listFilesetResponse = lfsResponse.(GetFilesetResponse)
	responseSize := len(listFilesetResponse.Filesets)
	response := make([]model.VolumeMetadata, responseSize)
	for i := 0; i < responseSize; i++ {
		name := listFilesetResponse.Filesets[i].Config.FilesetName
		//TODO check the mountpoint
		mountpoint := listFilesetResponse.Filesets[i].Config.Path

		response[i] = model.VolumeMetadata{Name: name, Mountpoint: mountpoint}
	}
	return response, nil
}

func (s *spectrum_rest) ListFileset(filesystemName string, filesetName string) (model.VolumeMetadata, error) {
	getFilesetURL := utils.FormatURL(s.endpoint, fmt.Sprintf("scalemgmt/v1/filesets/%s?filesystemname=%s", filesetName, filesystemName))
	getFilesetResponse := GetFilesetResponse{}
	gfsResponse, err := s.doHTTP(getFilesetURL, "GET", getFilesetResponse)
	if err != nil {
		s.logger.Printf("error in processing remote call %v", err)
		return model.VolumeMetadata{}, err
	}
	getFilesetResponse = gfsResponse.(GetFilesetResponse)
	name := getFilesetResponse.Filesets[0].Config.FilesetName
	//TODO get the mountpoint
	mountpoint := getFilesetResponse.Filesets[0].Config.Path

	return model.VolumeMetadata{Name: name, Mountpoint: mountpoint}, nil
}
func (s *spectrum_rest) IsFilesetLinked(filesystemName string, filesetName string) (bool, error) {
	return true, nil
}

//TODO modify quota from string to Capacity (see kubernetes)
func (s *spectrum_rest) ListFilesetQuota(filesystemName string, filesetName string) (string, error) {
	listQuotaURL := utils.FormatURL(s.endpoint, fmt.Sprintf("scalemgmt/v1/filesets/%s?filesystemname=%s", filesetName, filesystemName))
	listQuotaResponse := GetQuotaResponse{}
	gqResponse, err := s.doHTTP(listQuotaURL, "GET", listQuotaResponse)
	if err != nil {
		s.logger.Printf("error in processing remote call %v", err)
		return "", err
	}
	listQuotaResponse = gqResponse.(GetQuotaResponse)
	//TODO check which quota in quotas[] and which attribute
	return listQuotaResponse.Quotas[0].BlockQuota, nil
}
func (s *spectrum_rest) SetFilesetQuota(filesystemName string, filesetName string, quota string) error {
	return nil
}

func (s *spectrum_rest) doHTTP(endpoint string, method string, responseObject interface{}) (interface{}, error) {
	response, err := utils.HttpExecute(s.httpClient, s.logger, method, endpoint, nil)
	if err != nil {
		s.logger.Printf("Error in %s: %s remote call %#v", method, endpoint, err)
		return nil, fmt.Errorf("Error in get filesystem remote call")
	}

	if response.StatusCode != http.StatusOK {
		s.logger.Printf("Error in get filesystem remote call %#v\n", response)
		return nil, utils.ExtractErrorResponse(response)
	}
	err = utils.UnmarshalResponse(response, &responseObject)
	if err != nil {
		s.logger.Printf("Error in unmarshalling response for get remote call %#v for response %#v", err, response)
		return nil, fmt.Errorf("Error in unmarshalling response for get remote call")
	}

	return responseObject, nil
}
