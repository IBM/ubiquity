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

package connectors

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"time"

	"github.com/IBM/ubiquity/resources"
	"github.com/IBM/ubiquity/utils"
)

type spectrumRestV2 struct {
	logger     *log.Logger
	httpClient *http.Client
	endpoint   string
	user       string
	password   string
	hostname   string
}

func (s *spectrumRestV2) isStatusOK(statusCode int) bool {
	s.logger.Println("spectrumRestConnector: isStatusOK")
	defer s.logger.Println("spectrumRestConnector: isStatusOK end")

	if (statusCode == http.StatusOK) ||
		(statusCode == http.StatusCreated) ||
		(statusCode == http.StatusAccepted) {
		return true
	}
	return false
}

func (s *spectrumRestV2) checkAsynchronousJob(statusCode int) bool {
	s.logger.Println("spectrumRestConnector: checkAsynchronousJob")
	defer s.logger.Println("spectrumRestConnector: checkAsynchronousJob end")

	if (statusCode == http.StatusAccepted) ||
		(statusCode == http.StatusCreated) {
		return true
	}
	return false
}

func (s *spectrumRestV2) isRequestAccepted(response GenericResponse, url string) error {
	s.logger.Println("spectrumRestConnector: isRequestAccepted")
	defer s.logger.Println("spectrumRestConnector: isRequestAccepted end")

	if !s.isStatusOK(response.Status.Code) {
		return fmt.Errorf("error %v for url %v", response, url)
	}

	if len(response.Jobs) == 0 {
		return fmt.Errorf("Unable to get Job details for %v request", url)
	}
	return nil
}

func (s *spectrumRestV2) waitForJobCompletion(statusCode int, jobID uint64) error {
	s.logger.Println("spectrumRestConnector: waitForJobCompletion")
	defer s.logger.Println("spectrumRestConnector: waitForJobCompletion end")

	if s.checkAsynchronousJob(statusCode) {
		jobURL := utils.FormatURL(s.endpoint, fmt.Sprintf("scalemgmt/v2/jobs?filter=jobId=%d&fields=:all:", jobID))
		s.logger.Println("Job URL: ", jobURL)
		err := s.AsyncJobCompletion(jobURL)
		if err != nil {
			s.logger.Printf("%v\n",err)
			return err
		}
	}
	return nil
}

func (s *spectrumRestV2) AsyncJobCompletion(jobURL string) error {
	s.logger.Println("spectrumRestConnector: AsyncJobCompletion")
	defer s.logger.Println("spectrumRestConnector: AsyncJobCompletion end")

	jobQueryResponse := GenericResponse{}
	for {
		s.logger.Printf("jobUrl  %v", jobURL)
		err := s.doHTTP(jobURL, "GET", &jobQueryResponse, nil)
		if err != nil {
			return err
		}
		if len(jobQueryResponse.Jobs) == 0 {
			return fmt.Errorf("Unable to get Job %v details", jobURL)
		}

		if jobQueryResponse.Jobs[0].Status == "RUNNING" {
			time.Sleep(5000 * time.Millisecond)
			continue
		}
		break
	}
	if jobQueryResponse.Jobs[0].Status == "COMPLETED" {
		s.logger.Printf("Job %v Completed Successfully: %v\n",jobURL,jobQueryResponse.Jobs[0].Result)
		return nil
	} else {
   	        return fmt.Errorf("Job %v Failed to Complete:\n %v",jobURL,jobQueryResponse.Jobs[0].Result)
	}
}

func NewSpectrumRestV2(logger *log.Logger, restConfig resources.RestConfig) (SpectrumScaleConnector, error) {

	endpoint := restConfig.Endpoint
	user := restConfig.User
	password := restConfig.Password
	hostname := restConfig.Hostname

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	return &spectrumRestV2{logger: logger, httpClient: &http.Client{Transport: tr}, endpoint: endpoint, user: user, password: password, hostname: hostname}, nil
}

func NewspectrumRestV2WithClient(logger *log.Logger, restConfig resources.RestConfig) (SpectrumScaleConnector, *http.Client, error) {
	endpoint := restConfig.Endpoint
	user := restConfig.User
	password := restConfig.Password
	hostname := restConfig.Hostname

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	return &spectrumRestV2{logger: logger, httpClient: client, endpoint: endpoint, user: user, password: password, hostname: hostname}, client, nil

}

func (s *spectrumRestV2) GetClusterId() (string, error) {
	s.logger.Println("spectrumRestConnector: GetClusterId")
	defer s.logger.Println("spectrumRestConnector: GetClusterId end")

	getClusterURL := utils.FormatURL(s.endpoint, "scalemgmt/v2/cluster")
	getClusterResponse := GetClusterResponse{}

	s.logger.Println("Get Cluster URL : %s", getClusterURL)

	err := s.doHTTP(getClusterURL, "GET", &getClusterResponse, nil)
	if err != nil {
		s.logger.Printf("error in executing remote call: %v", err)
		return "", err
	}
	cid_str := fmt.Sprintf("%v", getClusterResponse.Cluster.ClusterSummary.ClusterID)
	return cid_str, nil
}

func (s *spectrumRestV2) IsFilesystemMounted(filesystemName string) (bool, error) {
	s.logger.Println("spectrumRestConnector: IsFilesystemMounted")
	defer s.logger.Println("spectrumRestConnector: IsFilesystemMounted end")

	var currentNode string
	getNodesURL := utils.FormatURL(s.endpoint, "scalemgmt/v2/nodes")
	getNodesResponse := GetNodesResponse_v2{}

	s.logger.Println("Get Nodes URL %s", getNodesURL)

	for {
		err := s.doHTTP(getNodesURL, "GET", &getNodesResponse, nil)
		if err != nil {
			s.logger.Printf("error in executing remote call: %v", err)
			return false, err
		}

		if s.hostname != "" {
			s.logger.Printf("Got hostname from config %v", s.hostname)
			currentNode = s.hostname
		} else {
			currentNode, _ = os.Hostname()
		}
		s.logger.Printf("spectrum rest Client: node name: %s\n", currentNode)
		for _, node := range getNodesResponse.Nodes {
			if node.AdminNodename == currentNode {
				return true, nil
			}
		}
		if getNodesResponse.Paging.Next == "" {
			break
		} else {
			getNodesURL = getNodesResponse.Paging.Next
		}
	}
	return false, nil
}

func (s *spectrumRestV2) MountFileSystem(filesystemName string) error {
	fmt.Printf("This method is not yet implemented")
	return nil
}

func (s *spectrumRestV2) ListFilesystems() ([]string, error) {

	s.logger.Println("spectrumRestConnector: ListFilesystems")
	defer s.logger.Println("spectrumRestConnector: ListFilesystems end")

	listFilesystemsURL := utils.FormatURL(s.endpoint, "scalemgmt/v2/filesystems")
	getFilesystemResponse := GetFilesystemResponse_v2{}

	s.logger.Println("List Filesystem URL: ", listFilesystemsURL)

	err := s.doHTTP(listFilesystemsURL, "GET", &getFilesystemResponse, nil)
	if err != nil {
		s.logger.Printf("error in executing remote call: %v", err)
		return nil, err
	}
	fsNumber := len(getFilesystemResponse.FileSystems)
	filesystems := make([]string, fsNumber)
	for i := 0; i < fsNumber; i++ {
		filesystems[i] = getFilesystemResponse.FileSystems[i].Name
	}
	return filesystems, nil
}

func (s *spectrumRestV2) GetFilesystemMountpoint(filesystemName string) (string, error) {

	s.logger.Println("spectrumRestConnector: GetFilesystemMountpoint")
	defer s.logger.Println("spectrumRestConnector: GetFilesystemMountpoint end")

	getFilesystemURL := utils.FormatURL(s.endpoint, fmt.Sprintf("scalemgmt/v2/filesystems/%s", filesystemName))
	getFilesystemResponse := GetFilesystemResponse_v2{}

	s.logger.Println("Get Filesystem Mount URL: ", getFilesystemURL)

	err := s.doHTTP(getFilesystemURL, "GET", &getFilesystemResponse, nil)
	if err != nil {
		s.logger.Printf("error in executing remote call: %v", err)
		return "", err
	}

	if len(getFilesystemResponse.FileSystems) > 0 {
		return getFilesystemResponse.FileSystems[0].Mount.MountPoint, nil
	} else {
		return "", fmt.Errorf("Unable to get Filesystem")
	}
}

func (s *spectrumRestV2) CreateFileset(filesystemName string, filesetName string, opts map[string]interface{}) error {

	s.logger.Println("spectrumRestConnector: CreateFileset")
	defer s.logger.Println("spectrumRestConnector: CreateFileset end")

	filesetreq := CreateFilesetRequest{}
	filesetreq.FilesetName = filesetName
	filesetreq.Comment = "fileset for container volume"	
	filesetType, filesetTypeSpecified := opts[UserSpecifiedFilesetType]
	inodeLimit, inodeLimitSpecified := opts[UserSpecifiedInodeLimit]
	if filesetTypeSpecified && filesetType.(string) == "independent" {
		filesetreq.InodeSpace = "new"
		if inodeLimitSpecified {
			filesetreq.MaxNumInodes = inodeLimit.(string)
		}
	} else {
		filesetreq.InodeSpace = "root"
	}
	s.logger.Printf("filesetreq %v\n", filesetreq)
	createFilesetURL := utils.FormatURL(s.endpoint, fmt.Sprintf("scalemgmt/v2/filesystems/%s/filesets", filesystemName))
	createFilesetResponse := GenericResponse{}

	s.logger.Println("Create Fileset URL: ", createFilesetURL)

	err := s.doHTTP(createFilesetURL, "POST", &createFilesetResponse, filesetreq)
	if err != nil {
		s.logger.Printf("error in remote call %v", err)
		return err
	}

	err = s.isRequestAccepted(createFilesetResponse, createFilesetURL)
	if err != nil {
		return err
	}

	err = s.waitForJobCompletion(createFilesetResponse.Status.Code, createFilesetResponse.Jobs[0].JobID)
	if err != nil {
		return fmt.Errorf("Unable to create fileset %v. Please refer Ubiquity server logs for more details",filesetName)
	}
	return nil
}

func (s *spectrumRestV2) DeleteFileset(filesystemName string, filesetName string) error {

	s.logger.Println("spectrumRestConnector: DeleteFileset")
	defer s.logger.Println("spectrumRestConnector: DeleteFileset end")

	deleteFilesetURL := utils.FormatURL(s.endpoint, fmt.Sprintf("scalemgmt/v2/filesystems/%s/filesets/%s", filesystemName, filesetName))
	deleteFilesetResponse := GenericResponse{}

	s.logger.Println("Delete Fileset URL: ", deleteFilesetURL)

	err := s.doHTTP(deleteFilesetURL, "DELETE", &deleteFilesetResponse, nil)
	if err != nil {
		s.logger.Printf("Error in delete remote call")
		return err
	}

	err = s.isRequestAccepted(deleteFilesetResponse, deleteFilesetURL)
	if err != nil {
		return err
	}

	err = s.waitForJobCompletion(deleteFilesetResponse.Status.Code, deleteFilesetResponse.Jobs[0].JobID)
	if err != nil {
		return fmt.Errorf("Unable to delete fileset %v. Please refer Ubiquity server logs for more details",filesetName)
	}

	return nil
}

func (s *spectrumRestV2) LinkFileset(filesystemName string, filesetName string) error {

	s.logger.Println("spectrumRestConnector: LinkFileset")
	defer s.logger.Println("spectrumRestConnector: LinkFileset end")

	linkReq := LinkFilesetRequest{}
	fsMountpoint, err := s.GetFilesystemMountpoint(filesystemName)
	if err != nil {
		s.logger.Printf("error in linking fileset")
		return err
	}

	linkReq.Path = path.Join(fsMountpoint, filesetName)
	linkFilesetURL := utils.FormatURL(s.endpoint, fmt.Sprintf("scalemgmt/v2/filesystems/%s/filesets/%s/link", filesystemName, filesetName))
	linkFilesetResponse := GenericResponse{}

	s.logger.Println("Link Fileset URL: ", linkFilesetURL)

	err = s.doHTTP(linkFilesetURL, "POST", &linkFilesetResponse, linkReq)
	if err != nil {
		s.logger.Printf("error in remote call %v", err)
		return err
	}

	err = s.isRequestAccepted(linkFilesetResponse, linkFilesetURL)
	if err != nil {
		return err
	}

	err = s.waitForJobCompletion(linkFilesetResponse.Status.Code, linkFilesetResponse.Jobs[0].JobID)
	if err != nil {
		return fmt.Errorf("Unable to link fileset %v. Please refer Ubiquity server logs for more details",filesetName)
	}
	return nil
}

func (s *spectrumRestV2) UnlinkFileset(filesystemName string, filesetName string) error {

	s.logger.Println("spectrumRestConnector: UnlinkFileset")
	defer s.logger.Println("spectrumRestConnector: UnlinkFileset end")

	UnlinkReq := UnlinkFilesetRequest{}
	UnlinkReq.Force = true

	unlinkFilesetURL := utils.FormatURL(s.endpoint, fmt.Sprintf("scalemgmt/v2/filesystems/%s/filesets/%s/link", filesystemName, filesetName))
	unlinkFilesetResponse := GenericResponse{}

	s.logger.Println("Unlink Fileset URL: ", unlinkFilesetURL)

	err := s.doHTTP(unlinkFilesetURL, "DELETE", &unlinkFilesetResponse, UnlinkReq)

	if err != nil {
		s.logger.Printf("error in remote call %v", err)
		return err
	}

	err = s.isRequestAccepted(unlinkFilesetResponse, unlinkFilesetURL)
	if err != nil {
		return err
	}

	err = s.waitForJobCompletion(unlinkFilesetResponse.Status.Code, unlinkFilesetResponse.Jobs[0].JobID)
	if err != nil {
		return fmt.Errorf("Unable to unlink fileset %v. Please refer Ubiquity server logs for more details",filesetName)
	}

	return nil
}

func (s *spectrumRestV2) ListFileset(filesystemName string, filesetName string) (resources.Volume, error) {

	s.logger.Println("spectrumRestConnector: ListFileset")
	defer s.logger.Println("spectrumRestConnector: ListFileset end")

	getFilesetURL := utils.FormatURL(s.endpoint, fmt.Sprintf("scalemgmt/v2/filesystems/%s/filesets/%s", filesystemName, filesetName))
	getFilesetResponse := GetFilesetResponse_v2{}

	s.logger.Println("List Fileset URL: ", getFilesetURL)

	err := s.doHTTP(getFilesetURL, "GET", &getFilesetResponse, nil)
	if err != nil {
		s.logger.Printf("error in processing remote call %v", err)
		return resources.Volume{}, err
	}

	if len(getFilesetResponse.Filesets) == 0 {
		return resources.Volume{}, fmt.Errorf("Unable to get fileset %v", getFilesetURL)
	}

	name := getFilesetResponse.Filesets[0].Config.FilesetName
	mountpoint := getFilesetResponse.Filesets[0].Config.Path

	return resources.Volume{Name: name, Mountpoint: mountpoint}, nil
}

func (s *spectrumRestV2) ListFilesets(filesystemName string) ([]resources.Volume, error) {

	s.logger.Println("spectrumRestConnector: ListFilesets")
	defer s.logger.Println("spectrumRestConnector: ListFilesets end")

	listFilesetURL := utils.FormatURL(s.endpoint, fmt.Sprintf("scalemgmt/v2/filesystems/%s/filesets", filesystemName))
	listFilesetResponse := GetFilesetResponse_v2{}

	s.logger.Println("List Filesets URL: ", listFilesetURL)

	var response []resources.Volume
	var responseSize int
	for {
		err := s.doHTTP(listFilesetURL, "GET", &listFilesetResponse, nil)
		if err != nil {
			s.logger.Printf("error in processing remote call %v", err)
			return nil, err
		}
		responseSize = len(listFilesetResponse.Filesets)

		for i := 0; i < responseSize; i++ {
			name := listFilesetResponse.Filesets[i].Config.FilesetName
			mountpoint := listFilesetResponse.Filesets[i].Config.Path
			response = append(response, resources.Volume{Name: name, Mountpoint: mountpoint})
		}
		if listFilesetResponse.Paging.Next == "" {
			break
		} else {
			listFilesetURL = listFilesetResponse.Paging.Next
		}
	}
	return response, nil
}

func (s *spectrumRestV2) IsFilesetLinked(filesystemName string, filesetName string) (bool, error) {

	s.logger.Println("spectrumRestConnector: IsFilesetLinked")
	defer s.logger.Println("spectrumRestConnector: IsFilesetLinked end")

	fileset, err := s.ListFileset(filesystemName, filesetName)
	if err != nil {
		s.logger.Printf("error retrieving fileset data")
		return false, err
	}

	if (fileset.Mountpoint == "") ||
		(fileset.Mountpoint == "--") {
		return false, nil
	}
	return true, nil
}

func (s *spectrumRestV2) SetFilesetQuota(filesystemName string, filesetName string, quota string) error {

	s.logger.Println("spectrumRestConnector: SetFilesetQuota")
	defer s.logger.Println("spectrumRestConnector: SetFilesetQuota end")

	setQuotaURL := utils.FormatURL(s.endpoint, fmt.Sprintf("scalemgmt/v2/filesystems/%s/filesets/%s/quotas", filesystemName, filesetName))
	quotaRequest := SetQuotaRequest_v2{}

	s.logger.Println("Set Quota URL: ", setQuotaURL)

	quotaRequest.BlockHardLimit = quota
	quotaRequest.BlockSoftLimit = quota
	quotaRequest.OperationType = "setQuota"
	quotaRequest.QuotaType = "fileset"

	setQuotaResponse := GenericResponse{}

	err := s.doHTTP(setQuotaURL, "POST", &setQuotaResponse, quotaRequest)
	if err != nil {
		s.logger.Printf("error setting quota for fileset %v", err)
		return err
	}

	err = s.isRequestAccepted(setQuotaResponse, setQuotaURL)
	if err != nil {
		return err
	}

	err = s.waitForJobCompletion(setQuotaResponse.Status.Code, setQuotaResponse.Jobs[0].JobID)
	if err != nil {
		return fmt.Errorf("Unable to set quota for fileset %v. Please refer Ubiquity server logs for more details",filesetName)
	}
	return nil
}

func (s *spectrumRestV2) ListFilesetQuota(filesystemName string, filesetName string) (string, error) {

	s.logger.Println("spectrumRestConnector: ListFilesetQuota")
	defer s.logger.Println("spectrumRestConnector: ListFilesetQuota end")

	listQuotaURL := utils.FormatURL(s.endpoint, fmt.Sprintf("scalemgmt/v2/filesystems/%s/filesets/%s/quotas", filesystemName, filesetName))
	listQuotaResponse := GetQuotaResponse_v2{}

	s.logger.Println("List Quota URL: ", listQuotaURL)

	err := s.doHTTP(listQuotaURL, "GET", &listQuotaResponse, nil)
	if err != nil {
		s.logger.Printf("error in processing remote call %v", err)
		return "", err
	}

	//TODO check which quota in quotas[] and which attribute
	if len(listQuotaResponse.Quotas) > 0 {
		return fmt.Sprintf("%d", listQuotaResponse.Quotas[0].BlockQuota), nil
	} else {
		return "", fmt.Errorf("Unable to get Quota information")
	}
}

func (s *spectrumRestV2) ExportNfs(volumeMountpoint string, clientConfig string) error {

	s.logger.Println("spectrumRestConnector: ExportNfs")
	defer s.logger.Println("spectrumRestConnector: ExportNfs end")

	exportNfsURL := utils.FormatURL(s.endpoint, fmt.Sprintf("scalemgmt/v2/nfs/exports"))
	nfsExportReq := nfsExportRequest{}
	nfsExportReq.Path = volumeMountpoint
	nfsExportReq.ClientDetail = append(nfsExportReq.ClientDetail, clientConfig)

	s.logger.Println("Export NFS URL: ", exportNfsURL)
	s.logger.Printf("volumemount %s clientdetail %s\n", nfsExportReq.Path, nfsExportReq.ClientDetail)

	nfsExportResp := GenericResponse{}
	err := s.doHTTP(exportNfsURL, "POST", &nfsExportResp, nfsExportReq)
	if err != nil {
		s.logger.Printf("error during NFS export %v", err)
		return err
	}

	err = s.isRequestAccepted(nfsExportResp, exportNfsURL)
	if err != nil {
		return err
	}

	err = s.waitForJobCompletion(nfsExportResp.Status.Code, nfsExportResp.Jobs[0].JobID)
	if err != nil {
		return fmt.Errorf("Unable to export %v. Please refer Ubiquity server logs for more details",volumeMountpoint)
	}
	return nil
}

func (s *spectrumRestV2) UnexportNfs(volumeMountpoint string) error {

	s.logger.Println("spectrumRestConnector: UnexportNfs")
	defer s.logger.Println("spectrumRestConnector: UnexportNfs end")

	volumeMountpoint = url.QueryEscape(volumeMountpoint)
	unexportNfsURL := utils.FormatURL(s.endpoint, "scalemgmt/v2/nfs/exports/", volumeMountpoint)
	unexportNfsResp := GenericResponse{}

	s.logger.Printf("NFS export DELETE URL: \n", unexportNfsURL)

	err := s.doHTTP(unexportNfsURL, "DELETE", &unexportNfsResp, nil)
	if err != nil {
		s.logger.Printf("Error while deleting NFS export %v", err)
		return err
	}

	err = s.isRequestAccepted(unexportNfsResp, unexportNfsURL)
	if err != nil {
		return err
	}

	err = s.waitForJobCompletion(unexportNfsResp.Status.Code, unexportNfsResp.Jobs[0].JobID)

	if err != nil {
		return fmt.Errorf("Unable to remove export %v. Please refer Ubiquity server logs for more details",volumeMountpoint)
	}
	return nil
}

func (s *spectrumRestV2) doHTTP(endpoint string, method string, responseObject interface{}, param interface{}) error {
	response, err := utils.HttpExecuteUserAuth(s.httpClient, s.logger, method, endpoint, s.user, s.password, param)
	if err != nil {
		s.logger.Printf("Error in %s: %s remote call %#v", method, endpoint, err)

		return err
	}

	if !s.isStatusOK(response.StatusCode) {
		s.logger.Printf("Error in get filesystem remote call %#v\n", response)
		return utils.ExtractErrorResponse(response)
	}
	err = utils.UnmarshalResponse(response, responseObject)
	if err != nil {
		s.logger.Printf("Error in unmarshalling response for get remote call %#v for response %#v", err, response)
		return err

	}

	return nil
}
