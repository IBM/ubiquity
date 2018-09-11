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
	"crypto/x509"
	"fmt"
	"github.com/IBM/ubiquity/utils/logs"
	"net/http"
	"net/url"
	"path"
	"time"
	"strings"
    "io/ioutil"
	"github.com/IBM/ubiquity/resources"
	"github.com/IBM/ubiquity/utils"
	"os"
)

type spectrumRestV2 struct {
	logger     logs.Logger
	httpClient *http.Client
	endpoint   string
	user       string
	password   string
}

type SslModeValueInvalid struct {
        sslModeInValid string
}

func (e *SslModeValueInvalid) Error() string {
        return fmt.Sprintf("Illegal SSL mode value [%s]. The allowed values are [%s, %s]",
                e.sslModeInValid, resources.SslModeRequire, resources.SslModeVerifyFull)
}

type SslModeFullVerifyWithoutCAfile struct {
        VerifyCaEnvName string
}

func (e *SslModeFullVerifyWithoutCAfile) Error() string {
        return fmt.Sprintf("Environment variable [%s] must be set for the SSL mode [%s]",
                e.VerifyCaEnvName, resources.SslModeVerifyFull)
}

func (s *spectrumRestV2) isStatusOK(statusCode int) bool {
    defer s.logger.Trace(logs.DEBUG)()

	if (statusCode == http.StatusOK) ||
		(statusCode == http.StatusCreated) ||
		(statusCode == http.StatusAccepted) {
		return true
	}
	return false
}

func (s *spectrumRestV2) checkAsynchronousJob(statusCode int) bool {
    defer s.logger.Trace(logs.DEBUG)()

	if (statusCode == http.StatusAccepted) ||
		(statusCode == http.StatusCreated) {
		return true
	}
	return false
}

func (s *spectrumRestV2) isRequestAccepted(response GenericResponse, url string) error {
    defer s.logger.Trace(logs.DEBUG)()

	if !s.isStatusOK(response.Status.Code) {
		return fmt.Errorf("error %v for url %v", response, url)
	}

	if len(response.Jobs) == 0 {
		return fmt.Errorf("Unable to get Job details for %v request", url)
	}
	return nil
}

func (s *spectrumRestV2) waitForJobCompletion(statusCode int, jobID uint64) error {
    defer s.logger.Trace(logs.DEBUG)()

	if s.checkAsynchronousJob(statusCode) {
		jobURL := utils.FormatURL(s.endpoint, fmt.Sprintf("scalemgmt/v2/jobs?filter=jobId=%d&fields=:all:", jobID))
		s.logger.Debug("Job URL: ", logs.Args{{"jobUrl", jobURL}})
		err := s.AsyncJobCompletion(jobURL)
		if err != nil {
			s.logger.Debug("Error", logs.Args{{"Error", err}})
			return err
		}
	}
	return nil
}

func (s *spectrumRestV2) AsyncJobCompletion(jobURL string) error {
    defer s.logger.Trace(logs.DEBUG)()

	jobQueryResponse := GenericResponse{}
	for {
		s.logger.Debug("jobUrl ", logs.Args{{"JobUrl", jobURL}})
		err := s.doHTTP(jobURL, "GET", &jobQueryResponse, nil)
		if err != nil {
			return err
		}
		if len(jobQueryResponse.Jobs) == 0 {
			return fmt.Errorf("Unable to get Job %v details", jobURL)
		}

		if jobQueryResponse.Jobs[0].Status == "RUNNING" {
			time.Sleep(2000 * time.Millisecond)
			continue
		}
		break
	}
	if jobQueryResponse.Jobs[0].Status == "COMPLETED" {
		s.logger.Debug("Job Completed Successfully\n", logs.Args{{"jobUrl", jobURL}, {"response" , jobQueryResponse.Jobs[0].Result}})
		return nil
	} else {
	        return fmt.Errorf("%v", jobQueryResponse.Jobs[0].Result.Stderr)
	}
}

func NewSpectrumRestV2(logger logs.Logger, restConfig resources.RestConfig) (SpectrumScaleConnector, error) {

	endpoint := fmt.Sprintf("https://%s:%d/", restConfig.ManagementIP, restConfig.Port)
	user := restConfig.User
	password := restConfig.Password

	sslMode := strings.ToLower(os.Getenv(resources.KeySpectrumScaleSslMode))
	exec := utils.NewExecutor()

	if sslMode == "" {
		sslMode = resources.DefaultSpectrumScaleSslMode
	}

	if sslMode == resources.SslModeVerifyFull {
		verifyFileCA := os.Getenv("UBIQUITY_SERVER_VERIFY_SPECTRUMSCALE_CERT")

               if verifyFileCA != "" {
                    if _, err := exec.Stat(verifyFileCA); err != nil {
                        return &spectrumRestV2{}, logger.ErrorRet(err, "failed")
                    }
                    caCert, err := ioutil.ReadFile(verifyFileCA)
                    if err != nil {
                        return &spectrumRestV2{}, logger.ErrorRet(err, "failed")
                    }
                    caCertPool := x509.NewCertPool()
                    if ok := caCertPool.AppendCertsFromPEM(caCert); !ok {
                        return &spectrumRestV2{}, fmt.Errorf("parse %v failed", verifyFileCA)
                    }
                    tr = &http.Transport{TLSClientConfig: &tls.Config{RootCAs: caCertPool}}
                    logger.Debug("", logs.Args{{"UBIQUITY_SERVER_VERIFY_SPECTRUMSCALE_CERT", verifyFileCA}})
		} else {
		     return &spectrumRestV2{}, logger.ErrorRet(&SslModeFullVerifyWithoutCAfile{"UBIQUITY_SERVER_VERIFY_SPECTRUMSCALE_CERT"}, "failed")
		}
	} else if sslMode == resources.SslModeRequire {
		logger.Debug(
                    fmt.Sprintf("Client SSL Mode set to [%s]. Means the communication to ubiquity is InsecureSkipVerify", sslMode))
		tr = &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	} else {
		return &spectrumRestV2{}, logger.ErrorRet(&SslModeValueInvalid{sslMode}, "failed")
	}

	return &spectrumRestV2{logger: logger, httpClient: &http.Client{Transport: tr}, endpoint: endpoint, user: user, password: password, hostname: hostname}, nil
}

func NewspectrumRestV2WithClient(logger logs.Logger, restConfig resources.RestConfig) (SpectrumScaleConnector, *http.Client, error) {

	var tr * http.Transport
	endpoint := fmt.Sprintf("https://%s:%d/", restConfig.ManagementIP, restConfig.Port)
	user := restConfig.User
	password := restConfig.Password

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	return &spectrumRestV2{logger: logger, httpClient: client, endpoint: endpoint, user: user, password: password}, client, nil

}

func (s *spectrumRestV2) GetClusterId() (string, error) {
    defer s.logger.Trace(logs.DEBUG)()

	getClusterURL := utils.FormatURL(s.endpoint, "scalemgmt/v2/cluster")
	getClusterResponse := GetClusterResponse{}

	s.logger.Debug("", logs.Args{{"ClusterUrl", getClusterURL}})

	err := s.doHTTP(getClusterURL, "GET", &getClusterResponse, nil)
	if err != nil {
		s.logger.Debug("error in executing remote call", logs.Args{{"Error", err}})
		return "", fmt.Errorf("Unable to get cluster id. Please refer Ubiquity server logs for more details")
	}
	cid_str := fmt.Sprintf("%v", getClusterResponse.Cluster.ClusterSummary.ClusterID)
	return cid_str, nil
}

func (s *spectrumRestV2) IsFilesystemMounted(filesystemName string) (bool, error) {
    defer s.logger.Trace(logs.DEBUG)()

	ownerResp := OwnerResp_v2{}
	ownerUrl := utils.FormatURL(s.endpoint,fmt.Sprintf("scalemgmt/v2/filesystems/%s/owner/%s", filesystemName, url.QueryEscape("/")))
	err := s.doHTTP(ownerUrl, "GET", &ownerResp, nil)
    if err != nil {
		s.logger.Debug("Filesystem not mounted", logs.Args{{"Filesystem", filesystemName}, {"Url", ownerUrl}})
		return false, err
	}
	s.logger.Debug("", logs.Args{{"Response", ownerResp}})
	return true, nil
}

func (s *spectrumRestV2) MountFileSystem(filesystemName string) error {
	fmt.Printf("This method is not yet implemented")
	return nil
}

func (s *spectrumRestV2) ListFilesystems() ([]string, error) {
    defer s.logger.Trace(logs.DEBUG)()

	listFilesystemsURL := utils.FormatURL(s.endpoint, "scalemgmt/v2/filesystems")
	getFilesystemResponse := GetFilesystemResponse_v2{}

	s.logger.Debug("List Filesystem", logs.Args{{"ListFilesystemUrl", listFilesystemsURL}})

	err := s.doHTTP(listFilesystemsURL, "GET", &getFilesystemResponse, nil)
	if err != nil {
		s.logger.Debug("error in executing remote call", logs.Args{{"Error", err}})
		return nil, fmt.Errorf("Unable to list filesystems. Please refer Ubiquity server logs for more details")
	}
	fsNumber := len(getFilesystemResponse.FileSystems)
	filesystems := make([]string, fsNumber)
	for i := 0; i < fsNumber; i++ {
		filesystems[i] = getFilesystemResponse.FileSystems[i].Name
	}
	return filesystems, nil
}

func (s *spectrumRestV2) GetFilesystemMountpoint(filesystemName string) (string, error) {
    defer s.logger.Trace(logs.DEBUG)()

	getFilesystemURL := utils.FormatURL(s.endpoint, fmt.Sprintf("scalemgmt/v2/filesystems/%s", filesystemName))
	getFilesystemResponse := GetFilesystemResponse_v2{}

	s.logger.Debug("Get Filesystem Mount ", logs.Args{{"getFilesystemURL", getFilesystemURL}})

	err := s.doHTTP(getFilesystemURL, "GET", &getFilesystemResponse, nil)
	if err != nil {
		s.logger.Debug("error in executing remote call", logs.Args{{"Error", err}})
		return "", fmt.Errorf("Unable to fetch mount point for %v. Please refer Ubiquity server logs for more details", filesystemName)
	}

	if len(getFilesystemResponse.FileSystems) > 0 {
		return getFilesystemResponse.FileSystems[0].Mount.MountPoint, nil
	} else {
		return "", fmt.Errorf("Unable to fetch mount point for %v. Please refer Ubiquity server logs for more details", filesystemName)
	}
}

func (s *spectrumRestV2) CreateFileset(filesystemName string, filesetName string, opts map[string]interface{}) error {
    defer s.logger.Trace(logs.DEBUG)()

	filesetreq := CreateFilesetRequest{}
	filesetreq.FilesetName = filesetName
	filesetreq.Comment = "fileset for container volume"

	filesetType, filesetTypeSpecified := opts[UserSpecifiedFilesetType]
	inodeLimit, inodeLimitSpecified := opts[UserSpecifiedInodeLimit]
	if filesetTypeSpecified && filesetType.(string) == "independent" {
		filesetreq.InodeSpace = "new"
		if inodeLimitSpecified {
			filesetreq.MaxNumInodes = inodeLimit.(string)
			filesetreq.AllocInodes = inodeLimit.(string)
		}
	} else {
		filesetreq.InodeSpace = "root"
	}

	s.logger.Debug("filesetreq ", logs.Args{{"filesetreq", filesetreq}})
	createFilesetURL := utils.FormatURL(s.endpoint, fmt.Sprintf("scalemgmt/v2/filesystems/%s/filesets", filesystemName))
	createFilesetResponse := GenericResponse{}

	s.logger.Debug("Create Fileset URL", logs.Args{{"createFilesetURL", createFilesetURL}})

	err := s.doHTTP(createFilesetURL, "POST", &createFilesetResponse, filesetreq)
	if err != nil {
		s.logger.Debug("error in remote call", logs.Args{{"Error", err}})
		return fmt.Errorf("Unable to create fileset %v. Please refer Ubiquity server logs for more details", filesetName)
	}

	err = s.isRequestAccepted(createFilesetResponse, createFilesetURL)
	if err != nil {
		return err
	}

	err = s.waitForJobCompletion(createFilesetResponse.Status.Code, createFilesetResponse.Jobs[0].JobID)
	if err != nil {
		return fmt.Errorf("Unable to create fileset %v:%v Please refer Ubiquity server logs for more details", filesetName, err)
	}
	return nil
}

func (s *spectrumRestV2) DeleteFileset(filesystemName string, filesetName string) error {
    defer s.logger.Trace(logs.DEBUG)()

	deleteFilesetURL := utils.FormatURL(s.endpoint, fmt.Sprintf("scalemgmt/v2/filesystems/%s/filesets/%s", filesystemName, filesetName))
	deleteFilesetResponse := GenericResponse{}

	s.logger.Debug("Delete Fileset ", logs.Args{{"deleteFilesetURL", deleteFilesetURL}})

	err := s.doHTTP(deleteFilesetURL, "DELETE", &deleteFilesetResponse, nil)
	if err != nil {
		s.logger.Debug("Error in delete remote call")
		return fmt.Errorf("Unable to delete fileset %v. Please refer Ubiquity server logs for more details", filesetName)
	}

	err = s.isRequestAccepted(deleteFilesetResponse, deleteFilesetURL)
	if err != nil {
		return err
	}

	err = s.waitForJobCompletion(deleteFilesetResponse.Status.Code, deleteFilesetResponse.Jobs[0].JobID)
	if err != nil {
		return fmt.Errorf("Unable to delete fileset %v:%v. Please refer Ubiquity server logs for more details", filesetName, err)
	}

	return nil
}

func (s *spectrumRestV2) LinkFileset(filesystemName string, filesetName string) error {
    defer s.logger.Trace(logs.DEBUG)()

	linkReq := LinkFilesetRequest{}
	fsMountpoint, err := s.GetFilesystemMountpoint(filesystemName)
	if err != nil {
		s.logger.Debug("error in linking fileset")
		return err
	}

	linkReq.Path = path.Join(fsMountpoint, filesetName)
	linkFilesetURL := utils.FormatURL(s.endpoint, fmt.Sprintf("scalemgmt/v2/filesystems/%s/filesets/%s/link", filesystemName, filesetName))
	linkFilesetResponse := GenericResponse{}

	s.logger.Debug("Link Fileset URL", logs.Args{{"linkFilesetURL",  linkFilesetURL}})

	err = s.doHTTP(linkFilesetURL, "POST", &linkFilesetResponse, linkReq)
	if err != nil {
		s.logger.Debug("error in remote call",logs.Args{{"Error", err}})
		return fmt.Errorf("Unable to link fileset %v. Please refer Ubiquity server logs for more details", filesetName)
	}

	err = s.isRequestAccepted(linkFilesetResponse, linkFilesetURL)
	if err != nil {
		return err
	}

	err = s.waitForJobCompletion(linkFilesetResponse.Status.Code, linkFilesetResponse.Jobs[0].JobID)
	if err != nil {
		return fmt.Errorf("Unable to link fileset %v:%v. Please refer Ubiquity server logs for more details", filesetName, err)
	}
	return nil
}

func (s *spectrumRestV2) UnlinkFileset(filesystemName string, filesetName string) error {
    defer s.logger.Trace(logs.DEBUG)()

	unlinkFilesetURL := utils.FormatURL(s.endpoint, fmt.Sprintf("scalemgmt/v2/filesystems/%s/filesets/%s/link?force=True", filesystemName, filesetName))
	unlinkFilesetResponse := GenericResponse{}

	s.logger.Debug("Unlink Fileset ", logs.Args{{"unlinkFilesetURL", unlinkFilesetURL}})

	err := s.doHTTP(unlinkFilesetURL, "DELETE", &unlinkFilesetResponse, nil)

	if err != nil {
		s.logger.Debug("error in remote call", logs.Args{{"Error", err}})
		return fmt.Errorf("Unable to unlink fileset %v. Please refer Ubiquity server logs for more details", filesetName)
	}

	err = s.isRequestAccepted(unlinkFilesetResponse, unlinkFilesetURL)
	if err != nil {
		return err
	}

	err = s.waitForJobCompletion(unlinkFilesetResponse.Status.Code, unlinkFilesetResponse.Jobs[0].JobID)
	if err != nil {
		return fmt.Errorf("Unable to unlink fileset %v:%v. Please refer Ubiquity server logs for more details", filesetName, err)
	}

	return nil
}

func (s *spectrumRestV2) ListFileset(filesystemName string, filesetName string) (resources.Volume, error) {
    defer s.logger.Trace(logs.DEBUG)()

	getFilesetURL := utils.FormatURL(s.endpoint, fmt.Sprintf("scalemgmt/v2/filesystems/%s/filesets/%s", filesystemName, filesetName))
	getFilesetResponse := GetFilesetResponse_v2{}

	s.logger.Debug("List Fileset URL", logs.Args{{"getFilesetURL", getFilesetURL}})

	err := s.doHTTP(getFilesetURL, "GET", &getFilesetResponse, nil)
	if err != nil {
		s.logger.Debug("error in processing remote call", logs.Args{{"Error", err}})
		return resources.Volume{}, fmt.Errorf("Unable to list fileset %v. Please refer Ubiquity server logs for more details", filesetName)
	}

	if len(getFilesetResponse.Filesets) == 0 {
		return resources.Volume{}, fmt.Errorf("Unable to list fileset %v. Please refer Ubiquity server logs for more details", filesetName)
	}

	name := getFilesetResponse.Filesets[0].Config.FilesetName
	mountpoint := getFilesetResponse.Filesets[0].Config.Path

	return resources.Volume{Name: name, Mountpoint: mountpoint}, nil
}

func (s *spectrumRestV2) ListFilesets(filesystemName string) ([]resources.Volume, error) {
    defer s.logger.Trace(logs.DEBUG)()

	listFilesetURL := utils.FormatURL(s.endpoint, fmt.Sprintf("scalemgmt/v2/filesystems/%s/filesets", filesystemName))
	listFilesetResponse := GetFilesetResponse_v2{}

	s.logger.Debug("List Filesets URL", logs.Args{{"listFilesetURL", listFilesetURL}})

	var response []resources.Volume
	var responseSize int
	for {
		err := s.doHTTP(listFilesetURL, "GET", &listFilesetResponse, nil)
		if err != nil {
			s.logger.Debug("error in processing remote call", logs.Args{{"Error", err}})
			return nil, fmt.Errorf("Unable to list filesets for %v. Please refer Ubiquity server logs for more details", filesystemName)
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
    defer s.logger.Trace(logs.DEBUG)()

	fileset, err := s.ListFileset(filesystemName, filesetName)
	if err != nil {
		s.logger.Debug("error retrieving fileset data")
		return false, err
	}

	if (fileset.Mountpoint == "") ||
		(fileset.Mountpoint == "--") {
		return false, nil
	}
	return true, nil
}

func (s *spectrumRestV2) SetFilesetQuota(filesystemName string, filesetName string, quota string) error {
    defer s.logger.Trace(logs.DEBUG)()

	setQuotaURL := utils.FormatURL(s.endpoint, fmt.Sprintf("scalemgmt/v2/filesystems/%s/quotas", filesystemName))
	quotaRequest := SetQuotaRequest_v2{}

	s.logger.Debug("Set Quota URL: ", logs.Args{{"setQuotaURL", setQuotaURL}})

	quotaRequest.BlockHardLimit = quota
	quotaRequest.BlockSoftLimit = quota
	quotaRequest.OperationType = "setQuota"
	quotaRequest.QuotaType = "fileset"
	quotaRequest.ObjectName = filesetName

	setQuotaResponse := GenericResponse{}

	err := s.doHTTP(setQuotaURL, "POST", &setQuotaResponse, quotaRequest)
	if err != nil {
		s.logger.Debug("error setting quota for fileset", logs.Args{{"Error", err}})
		return fmt.Errorf("Unable to set quota for fileset %v. Please refer Ubiquity server logs for more details", filesetName)
	}

	err = s.isRequestAccepted(setQuotaResponse, setQuotaURL)
	if err != nil {
		return err
	}

	err = s.waitForJobCompletion(setQuotaResponse.Status.Code, setQuotaResponse.Jobs[0].JobID)
	if err != nil {
		return fmt.Errorf("Unable to set quota for fileset %v:%v. Please refer Ubiquity server logs for more details", filesetName, err)
	}
	return nil
}

func (s *spectrumRestV2) ListFilesetQuota(filesystemName string, filesetName string) (string, error) {
    defer s.logger.Trace(logs.DEBUG)()

	listQuotaURL := utils.FormatURL(s.endpoint, fmt.Sprintf("scalemgmt/v2/filesystems/%s/quotas?filter=objectName=%s", filesystemName, filesetName))
	listQuotaResponse := GetQuotaResponse_v2{}

	s.logger.Debug("List Quota URL", logs.Args{{"listQuotaURL", listQuotaURL}})

	err := s.doHTTP(listQuotaURL, "GET", &listQuotaResponse, nil)
	if err != nil {
		s.logger.Debug("error in processing remote call", logs.Args{{"Error", err}})
		return "", fmt.Errorf("Unable to fetch quota information %v. Please refer Ubiquity server logs for more details", filesystemName)
	}

	//TODO check which quota in quotas[] and which attribute
	if len(listQuotaResponse.Quotas) > 0 {
		return fmt.Sprintf("%dK", listQuotaResponse.Quotas[0].BlockQuota), nil
	} else {
		return "", fmt.Errorf("Unable to fetch quota information %v. Please refer Ubiquity server logs for more details", filesystemName)
	}
}

func (s *spectrumRestV2) ExportNfs(volumeMountpoint string, clientConfig string) error {
    defer s.logger.Trace(logs.DEBUG)()

	exportNfsURL := utils.FormatURL(s.endpoint, fmt.Sprintf("scalemgmt/v2/nfs/exports"))
	nfsExportReq := nfsExportRequest{}
	nfsExportReq.Path = volumeMountpoint
	nfsExportReq.ClientDetail = append(nfsExportReq.ClientDetail, clientConfig)

	s.logger.Debug("Export NFS URL", logs.Args{{"exportNfsURL", exportNfsURL}})
	s.logger.Debug("", logs.Args{{"nfsExportReq.Path", nfsExportReq.Path} , {"nfsExportReq.ClientDetail", nfsExportReq.ClientDetail}})

	nfsExportResp := GenericResponse{}
	err := s.doHTTP(exportNfsURL, "POST", &nfsExportResp, nfsExportReq)
	if err != nil {
		s.logger.Debug("error during NFS export", logs.Args{{"Error", err}})
		return fmt.Errorf("Unable to export %v. Please refer Ubiquity server logs for more details", volumeMountpoint)
	}

	err = s.isRequestAccepted(nfsExportResp, exportNfsURL)
	if err != nil {
		return err
	}

	err = s.waitForJobCompletion(nfsExportResp.Status.Code, nfsExportResp.Jobs[0].JobID)
	if err != nil {
		return fmt.Errorf("Unable to export %v:%v. Please refer Ubiquity server logs for more details", volumeMountpoint, err)
	}
	return nil
}

func (s *spectrumRestV2) UnexportNfs(volumeMountpoint string) error {
    defer s.logger.Trace(logs.DEBUG)()

	volumeMountpoint = url.QueryEscape(volumeMountpoint)
	unexportNfsURL := utils.FormatURL(s.endpoint, "scalemgmt/v2/nfs/exports/", volumeMountpoint)
	unexportNfsResp := GenericResponse{}

	s.logger.Debug("NFS export DELETE URL", logs.Args{{"unexportNfsURL", unexportNfsURL}})

	err := s.doHTTP(unexportNfsURL, "DELETE", &unexportNfsResp, nil)
	if err != nil {
		s.logger.Debug("Error while deleting NFS export", logs.Args{{"Error", err}})
		return fmt.Errorf("Unable to remove export %v. Please refer Ubiquity server logs for more details", volumeMountpoint)
	}

	err = s.isRequestAccepted(unexportNfsResp, unexportNfsURL)
	if err != nil {
		return err
	}

	err = s.waitForJobCompletion(unexportNfsResp.Status.Code, unexportNfsResp.Jobs[0].JobID)

	if err != nil {
		return fmt.Errorf("Unable to remove export %v:%v. Please refer Ubiquity server logs for more details", volumeMountpoint, err)
	}
	return nil
}

func (s *spectrumRestV2) doHTTP(endpoint string, method string, responseObject interface{}, param interface{}) error {
	response, err := utils.HttpExecuteUserAuth(s.httpClient, method, endpoint, s.user, s.password, param)
	if err != nil {
		s.logger.Debug("Error in remote call", logs.Args{{"Method", method}, {"endpoint", endpoint}, {"Error", err}})

		return err
	}

	if !s.isStatusOK(response.StatusCode) {
		s.logger.Debug("Remote call completed with error", logs.Args{{"Response", response}})
		return fmt.Errorf("Remote call completed with error")
	}
	err = utils.UnmarshalResponse(response, responseObject)
	if err != nil {
		s.logger.Debug("Error in unmarshalling response for get remote call", logs.Args{{"Error", err},{"response", response}})
		return err

	}

	return nil
}
