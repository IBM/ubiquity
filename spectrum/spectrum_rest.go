package spectrum

import (
	"fmt"
	"log"
	"net/http"

	"github.ibm.com/almaden-containers/ubiquity/model"
	"github.ibm.com/almaden-containers/ubiquity/utils"
)

type spectrum_rest struct {
	logger     *log.Logger
	httpClient *http.Client
	endpoint   string
}

func NewSpectrumRest(logger *log.Logger, opts map[string]interface{}) Spectrum {
	return &spectrum_rest{logger: logger, httpClient: &http.Client{}, endpoint: opts["endpoint"].(string)}
}
func (s *spectrum_rest) GetClusterId() (string, error) {
	getClusterIDURL := utils.FormatURL(s.endpoint, "scalemgmt/v1/cluster")
	response, err := utils.HttpExecute(s.httpClient, s.logger, "GET", getClusterIDURL, nil)
	if err != nil {
		s.logger.Printf("Error in get cluster ID remote call %#v", err)
		return "", fmt.Errorf("Error in get Cluster ID remote call")
	}

	if response.StatusCode != http.StatusOK {
		s.logger.Printf("Error in get cluster ID remote call %#v\n", response)
		return "", utils.ExtractErrorResponse(response)
	}

	getClusterResponse := GetClusterResponse{}
	err = utils.UnmarshalResponse(response, &getClusterResponse)
	if err != nil {
		s.logger.Printf("Error in unmarshalling response for get remote call %#v for response %#v", err, response)
		return "", fmt.Errorf("Error in unmarshalling response for get remote call")
	}

	return getClusterResponse.Cluster.ClusterSummary.ClusterID, nil
}

func (s *spectrum_rest) IsFilesystemMounted(filesystemName string) (bool, error) {
	return true, nil
}
func (s *spectrum_rest) MountFileSystem(filesystemName string) error {
	return nil
}
func (s *spectrum_rest) ListFilesystems() ([]string, error) {
	return nil, nil
}
func (s *spectrum_rest) GetFilesystemMountpoint(filesystemName string) (string, error) {
	return "", nil
}
func (s *spectrum_rest) CreateFileset(filesystemName string, filesetName string, opts map[string]interface{}) error {
	return nil
}
func (s *spectrum_rest) DeleteFileset(filesystemName string, filesetName string) error { return nil }
func (s *spectrum_rest) LinkFileset(filesystemName string, filesetName string) error   { return nil }
func (s *spectrum_rest) UnlinkFileset(filesystemName string, filesetName string) error { return nil }
func (s *spectrum_rest) ListFilesets(filesystemName string) ([]model.VolumeMetadata, error) {
	return nil, nil
}
func (s *spectrum_rest) ListFileset(filesystemName string, filesetName string) (model.VolumeMetadata, error) {
	return model.VolumeMetadata{}, nil
}
func (s *spectrum_rest) IsFilesetLinked(filesystemName string, filesetName string) (bool, error) {
	return true, nil
}

//TODO modify quota from string to Capacity (see kubernetes)
func (s *spectrum_rest) ListFilesetQuota(filesystemName string, filesetName string) (string, error) {
	return "", nil
}
func (s *spectrum_rest) SetFilesetQuota(filesystemName string, filesetName string, quota string) error {
	return nil
}
