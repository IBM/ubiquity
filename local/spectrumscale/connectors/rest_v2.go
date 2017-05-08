package connectors

import (
        "fmt"
        "log"
        "net/http"
	"time" 
        "os"
        "crypto/tls"
        "github.com/IBM/ubiquity/resources"
        "github.com/IBM/ubiquity/utils"
)

type spectrum_rest_v2 struct {
        logger     *log.Logger
        httpClient *http.Client
        endpoint   string
        user       string
        password   string
}


func IsStatusOK(StatusCode int) bool {

	if ((StatusCode == http.StatusOK) ||
	    (StatusCode == http.StatusCreated) ||
	    (StatusCode == http.StatusAccepted)) { 
                return true
        }
        return false
}


func CheckAsynchronousJob(StatusCode int) bool {
	if ((StatusCode == http.StatusAccepted) ||
            (StatusCode == http.StatusCreated)) {
		return true
	}
	return false
 
}

func (s *spectrum_rest_v2) WaitForJobCompletion(statuscode int, jobID uint64) error {
        if (CheckAsynchronousJob(statuscode)) {
                JobID := jobID
                jobURL := utils.FormatURL(s.endpoint, fmt.Sprintf("scalemgmt/v2/jobs?filter=jobId=%d",JobID))
                async_status,err := s.AsyncJobCompletion(jobURL)
                if err != nil {
                        return fmt.Errorf("error getting jobid %v",err)
                }
                if (async_status == "SUCCESS") {
                        return nil
                } else {
			return fmt.Errorf("Job Failed")
		}
		
        }
	return nil
}


func (s *spectrum_rest_v2) AsyncJobCompletion(jobURL string) (status string, err error) {
	createFilesetResponse := GenericResponse{}
        for {
                s.logger.Printf("jobUrl  %v", jobURL)
                err = s.doHTTP(jobURL, "GET", &createFilesetResponse, nil)
                if err != nil {
                    return "FAILED", err;
                }
                if (createFilesetResponse.Jobs[0].Status == "RUNNING") {
                        time.Sleep(5000 * time.Millisecond)
                        continue
                }
                break;
        } 
        if (createFilesetResponse.Jobs[0].Status == "COMPLETED") {
	    return "SUCCESS", nil
	}
	return "FAILED", err
}


func NewSpectrumRest_v2(logger *log.Logger, restConfig resources.RestConfig) (SpectrumScaleConnector, error) {
        endpoint := restConfig.Endpoint
        user := restConfig.User
        password := restConfig.Password

        tr := &http.Transport{
                TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
        }
        return &spectrum_rest_v2{logger: logger, httpClient: &http.Client{Transport: tr}, endpoint: endpoint, user: user, password: password}, nil
}


func (s *spectrum_rest_v2) GetClusterId() (string, error) {
	getClusterURL := utils.FormatURL(s.endpoint, "scalemgmt/v2/cluster")

        getClusterResponse := GetClusterResponse{}
        err := s.doHTTP(getClusterURL, "GET", &getClusterResponse, nil)
        if err != nil {
                s.logger.Printf("error in executing remote call: %v", err)
                return "", err
        }
        cid_str := fmt.Sprintf("%v", getClusterResponse.Cluster.ClusterSummary.ClusterID)
        return cid_str, nil
}


func (s *spectrum_rest_v2) IsFilesystemMounted(filesystemName string) (bool, error) {
        getNodesURL := utils.FormatURL(s.endpoint, "scalemgmt/v2/nodes")
        getNodesResponse := GetNodesResponse_v2{}

        for {
	        err := s.doHTTP(getNodesURL, "GET", &getNodesResponse, nil)
                if err != nil {
                	s.logger.Printf("error in executing remote call: %v", err)
                        return false, err
                }
                currentNode, _ := os.Hostname()
                s.logger.Printf("spectrum rest Client: node name: %s\n", currentNode)
                for _, node := range getNodesResponse.Nodes {
                        if node.AdminNodename == currentNode {
                                return true, nil
                        }
                }
                if (getNodesResponse.Paging.Next == "") {
                        break
                } else {
                        getNodesURL = getNodesResponse.Paging.Next
                }
        }
        return false, nil
}		

func (s *spectrum_rest_v2) MountFileSystem(filesystemName string) error {
        fmt.Printf("This method is not yet implemented")
        return nil
}

func (s *spectrum_rest_v2) ListFilesystems() ([]string, error) {
        listFilesystemsURL := utils.FormatURL(s.endpoint, "scalemgmt/v2/filesystems")
        getFilesystemResponse := GetFilesystemResponse_v2{}
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

func (s *spectrum_rest_v2) GetFilesystemMountpoint(filesystemName string) (string, error) {
        getFilesystemURL := utils.FormatURL(s.endpoint, fmt.Sprintf("scalemgmt/v2/filesystems/%s", filesystemName))

        getFilesystemResponse := GetFilesystemResponse_v2{}

        err := s.doHTTP(getFilesystemURL, "GET", &getFilesystemResponse, nil)
        if err != nil {
                s.logger.Printf("error in executing remote call: %v", err)
                return "", err
        }

        return getFilesystemResponse.FileSystems[0].Mount.MountPoint, nil
}

func (s *spectrum_rest_v2) CreateFileset(filesystemName string, filesetName string, opts map[string]interface{}) error {

        filesetreq := CreateFilesetRequest{}
        filesetreq.FilesetName = filesetName
        filesetType, filesetTypeSpecified := opts[USER_SPECIFIED_FILESET_TYPE]
        inodeLimit, inodeLimitSpecified := opts[USER_SPECIFIED_INODE_LIMIT]
        if filesetTypeSpecified && filesetType.(string) == "independent" {
                filesetreq.InodeSpace = "new"
                if inodeLimitSpecified {
                        filesetreq.MaxNumInodes = inodeLimit.(string)
                }
        }
        createFilesetURL := utils.FormatURL(s.endpoint, fmt.Sprintf("scalemgmt/v2/filesystems/%s/filesets",filesystemName))

        createFilesetResponse := GenericResponse{}
        err := s.doHTTP(createFilesetURL, "POST", &createFilesetResponse, filesetreq)
        if err != nil {
                s.logger.Printf("error in remote call %v", err)
                return err
        }
        //TODO check the response message content and code
        if !IsStatusOK(createFilesetResponse.Status.Code)  {
                return fmt.Errorf("error creating fileset %v", createFilesetResponse)
        }

	err = s.WaitForJobCompletion(createFilesetResponse.Status.Code,createFilesetResponse.Jobs[0].JobID)
	if  err != nil {
		return err
	}
        return nil
}

func (s *spectrum_rest_v2) DeleteFileset(filesystemName string, filesetName string) error {

        deleteFilesetURL := utils.FormatURL(s.endpoint, fmt.Sprintf("scalemgmt/v2/filesystems/%s/filesets/%s", filesystemName, filesetName))
        deleteFilesetResponse := GenericResponse{}
        err := s.doHTTP(deleteFilesetURL, "DELETE", &deleteFilesetResponse, nil)
        if err != nil {
                s.logger.Printf("Error in delete remote call")
                return err
        }

        if !IsStatusOK(deleteFilesetResponse.Status.Code) {
                return fmt.Errorf("error deleting fileset %v", deleteFilesetResponse)
        }

       err = s.WaitForJobCompletion(deleteFilesetResponse.Status.Code, deleteFilesetResponse.Jobs[0].JobID)
        if  err != nil {
                return err
        }

        return nil
}


func (s *spectrum_rest_v2) LinkFileset(filesystemName string, filesetName string) error {
        LinkReq := LinkFilesetRequest{}
        fsMountpoint, err := s.GetFilesystemMountpoint(filesystemName)
        if err != nil {
                s.logger.Printf("error in linking fileset")
		return err
        }

        LinkReq.Path = fmt.Sprintf("%s/%s",fsMountpoint,filesetName)
        linkFilesetURL := utils.FormatURL(s.endpoint, fmt.Sprintf("scalemgmt/v2/filesystems/%s/filesets/%s/link",filesystemName, filesetName))
        linkFilesetResponse := GenericResponse{}

        err = s.doHTTP(linkFilesetURL, "POST", &linkFilesetResponse, LinkReq)
        if err != nil {
                s.logger.Printf("error in remote call %v", err)
                return err
        }

        if !IsStatusOK(linkFilesetResponse.Status.Code) {
                return fmt.Errorf("error linking fileset %v", linkFilesetResponse)
        }

        err = s.WaitForJobCompletion(linkFilesetResponse.Status.Code, linkFilesetResponse.Jobs[0].JobID)
        if  err != nil {
                return err
        }
        return nil
}


func (s *spectrum_rest_v2) UnlinkFileset(filesystemName string, filesetName string) error {

	UnlinkReq := UnlinkFilesetRequest{}
	UnlinkReq.Force = true 

	linkFilesetURL := utils.FormatURL(s.endpoint, fmt.Sprintf("scalemgmt/v2/filesystems/%s/filesets/%s/link",filesystemName,filesetName))
	unlinkFilesetResponse := GenericResponse{}

	err := s.doHTTP(linkFilesetURL, "DELETE", &unlinkFilesetResponse, UnlinkReq)

        if err != nil {
                s.logger.Printf("error in remote call %v", err)
                return err
        }

	if !IsStatusOK(unlinkFilesetResponse.Status.Code) {
		return fmt.Errorf("error unlinking fileset %v", unlinkFilesetResponse)
	}

        err = s.WaitForJobCompletion(unlinkFilesetResponse.Status.Code, unlinkFilesetResponse.Jobs[0].JobID)
        if  err != nil {
                return err
        }


	return nil
}

func (s *spectrum_rest_v2) ListFileset(filesystemName string, filesetName string) (resources.VolumeMetadata, error) {
        getFilesetURL := utils.FormatURL(s.endpoint, fmt.Sprintf("scalemgmt/v2/filesystems/%s/filesets/%s", filesystemName,filesetName))

        getFilesetResponse := GetFilesetResponse_v2{}
        err := s.doHTTP(getFilesetURL, "GET", &getFilesetResponse, nil)
        if err != nil {
                s.logger.Printf("error in processing remote call %v", err)
                return resources.VolumeMetadata{}, err
        }


        name := getFilesetResponse.Filesets[0].Config.FilesetName
        mountpoint := getFilesetResponse.Filesets[0].Config.Path

        return resources.VolumeMetadata{Name: name, Mountpoint: mountpoint}, nil
}

func (s *spectrum_rest_v2) ListFilesets(filesystemName string) ([]resources.VolumeMetadata, error) {
	listFilesetURL := utils.FormatURL(s.endpoint, "scalemgmt/v2/filesystems/%s/filesets",filesystemName)
	listFilesetResponse := GetFilesetResponse_v2{}

	var response []resources.VolumeMetadata 	
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
			response = append(response,resources.VolumeMetadata{Name: name, Mountpoint: mountpoint})
		}
		if (listFilesetResponse.Paging.Next == "") {
			break;
		} else {
			listFilesetURL = listFilesetResponse.Paging.Next
		}
	}
	return response, nil
}

func (s *spectrum_rest_v2) IsFilesetLinked(filesystemName string, filesetName string) (bool, error) {
        fileset, err := s.ListFileset(filesystemName, filesetName)
        if err != nil {
                s.logger.Printf("error retrieving fileset data")
                return false, err
        }

        if ((fileset.Mountpoint == "") ||
                (fileset.Mountpoint == "--")) {
                return false, nil
        }
        return true, nil
}

func (s *spectrum_rest_v2) SetFilesetQuota(filesystemName string, filesetName string, quota string) error {

	setQuotaURL := utils.FormatURL(s.endpoint, fmt.Sprintf("scalemgmt/v2/filesystems/%s/filesets/%s/quotas",filesystemName,filesetName))
	quotaRequest := SetQuotaRequest_v2{}	

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
        if !IsStatusOK(setQuotaResponse.Status.Code) {
                return fmt.Errorf("error unlinking fileset %v", setQuotaResponse)
        }

        err = s.WaitForJobCompletion(setQuotaResponse.Status.Code, setQuotaResponse.Jobs[0].JobID)
        if  err != nil {
                return err
        }


        return nil
}


func (s *spectrum_rest_v2) ListFilesetQuota(filesystemName string, filesetName string) (string, error) {

	listQuotaURL := utils.FormatURL(s.endpoint, fmt.Sprintf("scalemgmt/v2/filesystems/%s/filesets/%s/quotas",filesystemName,filesetName))
	listQuotaResponse := GetQuotaResponse_v2{}


        err := s.doHTTP(listQuotaURL, "GET", &listQuotaResponse, nil)
        if err != nil {
                s.logger.Printf("error in processing remote call %v", err)
                return "", err
        }

        //TODO check which quota in quotas[] and which attribute
        return listQuotaResponse.Quotas[0].BlockQuota, nil
}

func (s *spectrum_rest_v2) doHTTP(endpoint string, method string, responseObject interface{}, param interface{}) (error) {
        response, err := utils.HttpExecute(s.httpClient, s.logger, method, endpoint, s.user, s.password, param)
        if err != nil {
                s.logger.Printf("Error in %s: %s remote call %#v", method, endpoint, err)

                return err
        }

        if (!IsStatusOK(response.StatusCode))  {
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


