package scbe

import (
	"encoding/json"
	"fmt"
	"github.com/IBM/ubiquity/utils/logs"
	"github.com/IBM/ubiquity/resources"
)

//go:generate counterfeiter -o ../fakes/fake_scbe_rest_client.go . ScbeRestClient
type ScbeRestClient interface {
	Login() error
	CreateVolume(volName string, serviceName string, size int) (ScbeVolumeInfo, error)
	GetVolumes(wwn string) ([]ScbeVolumeInfo, error)
	DeleteVolume(wwn string) error
	MapVolume(wwn string, host string) (ScbeResponseMapping, error)
	UnmapVolume(wwn string, host string) error
	GetVolMapping(wwn string) (string, error)
	ServiceExist(serviceName string) (bool, error)
}

type scbeRestClient struct {
	logger         logs.Logger
	connectionInfo resources.ConnectionInfo
	client         SimpleRestClient
}

const (
	DefaultScbePort        = 8440
	UrlScbeReferer         = "https://%s:%d/"
	UrlScbeBaseSuffix      = "api/v1"
	UrlScbeResourceGetAuth = "users/get-auth-token"
	ScbeFlockerGroupParam  = "flocker"
	UrlScbeResourceService = "services"
	UrlScbeResourceVolume  = "volumes"
	UrlScbeResourceMapping = "mappings"
	UrlScbeResourceHost    = "hosts"
	DefaultSizeUnit        = "gb"
)

func NewScbeRestClient(conInfo resources.ConnectionInfo) ScbeRestClient {
	return newScbeRestClient(conInfo, nil)
}

// NewScbeRestClientWithNewRestClient for mocking during test # TODO consider to remove it to test file
func NewScbeRestClientWithSimpleRestClient(conInfo resources.ConnectionInfo, simpleClient SimpleRestClient) ScbeRestClient {
	return newScbeRestClient(conInfo, simpleClient)
}

func newScbeRestClient(conInfo resources.ConnectionInfo, simpleClient SimpleRestClient) ScbeRestClient {
	// Set default SCBE port if not mentioned
	if conInfo.Port == 0 {
		conInfo.Port = DefaultScbePort
	}
	// Add the default SCBE Flocker group to the credentials  # TODO change to ubiquity interface
	conInfo.CredentialInfo.Group = ScbeFlockerGroupParam

	if simpleClient == nil {
		referrer := fmt.Sprintf(UrlScbeReferer, conInfo.ManagementIP, conInfo.Port)
		baseUrl := referrer + UrlScbeBaseSuffix
		simpleClient = NewSimpleRestClient(conInfo, baseUrl, UrlScbeResourceGetAuth, referrer)
	}
	return &scbeRestClient{logs.GetLogger(), conInfo, simpleClient}
}

func (s *scbeRestClient) Login() error {
	defer s.logger.Trace(logs.DEBUG)()
	return s.client.Login()
}

// CreateVolume provision new volume on SCBE storage service.
// Return ScbeVolumeInfo of the new volume that was created
// Errors:
//	if service don't exist
//	if fail to create the volume
func (s *scbeRestClient) CreateVolume(volName string, serviceName string, size int) (ScbeVolumeInfo, error) {
	defer s.logger.Trace(logs.DEBUG)()
	// find the service in order to validate and also to get the service id
	services, err := s.serviceList(serviceName)
	if err != nil {
		return ScbeVolumeInfo{}, s.logger.ErrorRet(err, "failed")
	}
	// check existence of the service
	if len(services) <= 0 || services[0].Name != serviceName {
		return ScbeVolumeInfo{}, s.logger.ErrorRet(&serviceDoesntExistError{volName, serviceName, s.connectionInfo.ManagementIP}, "failed")
	}

	payload := ScbeCreateVolumePostParams{
		services[0].Id,
		volName,
		size,
		DefaultSizeUnit, // TODO lets support different type of unit size, for now only gb
	}

	payloadMarshaled, err := json.Marshal(payload)
	if err != nil {
		return ScbeVolumeInfo{}, s.logger.ErrorRet(err, "json.Marshal failed", logs.Args{{"payload", payload}})
	}
	volResponse := ScbeResponseVolume{}
	if err = s.client.Post(UrlScbeResourceVolume, payloadMarshaled, HTTP_SUCCEED_POST, &volResponse); err != nil {
		return ScbeVolumeInfo{}, s.logger.ErrorRet(err, "client.Post failed", logs.Args{{"payload", payload}})
	}

	return NewScbeVolumeInfo(&volResponse), nil
}

func (s *scbeRestClient) GetVolumes(wwn string) ([]ScbeVolumeInfo, error) {
	defer s.logger.Trace(logs.DEBUG)()
	vols, err := s.volumeList(wwn)
	if err != nil {
		return nil, s.logger.ErrorRet(err, "volumeList failed", logs.Args{{"wwn", wwn}})
	}
	scbeVolumes := []ScbeVolumeInfo{}
	for _, volume := range vols {
		scbeVolumes = append(scbeVolumes, NewScbeVolumeInfo(&volume))
	}

	return scbeVolumes, nil
}

func (s *scbeRestClient) DeleteVolume(wwn string) error {
	defer s.logger.Trace(logs.DEBUG)()
	urlToDelete := fmt.Sprintf("%s/%s", UrlScbeResourceVolume, wwn)
	if err := s.client.Delete(urlToDelete, []byte{}, HTTP_SUCCEED_DELETED); err != nil {
		return s.logger.ErrorRet(err, "client.Delete failed", logs.Args{{"url", urlToDelete}})
	}
	return nil
}

func (s *scbeRestClient) MapVolume(wwn string, host string) (ScbeResponseMapping, error) {
	defer s.logger.Trace(logs.DEBUG)()
	hostId, err := s.getHostIdByVol(wwn, host)
	if err != nil {
		return ScbeResponseMapping{}, s.logger.ErrorRet(err, "getHostIdByVol failed")
	}
	payload := ScbeMapVolumePostParams{VolumeId: wwn, HostId: hostId}
	payloadMarshal, err := json.Marshal(payload)
	if err != nil {
		return ScbeResponseMapping{}, s.logger.ErrorRet(err, "json.Marshal failed", logs.Args{{"payload", payload}})
	}
	mappingsResponse := ScbeResponseMappings{}
	if err = s.client.Post(UrlScbeResourceMapping, payloadMarshal, HTTP_SUCCEED_POST, &mappingsResponse); err != nil {
		return ScbeResponseMapping{}, s.logger.ErrorRet(err, "client.Post failed", logs.Args{{"payload", payload}})
	}
	if len(mappingsResponse.Mappings) != 1 {
		return ScbeResponseMapping{}, s.logger.ErrorRet(&mappingResponseError{mappingsResponse}, "failed")
	}
	return mappingsResponse.Mappings[0], nil
}

func (s *scbeRestClient) UnmapVolume(wwn string, host string) error {
	defer s.logger.Trace(logs.DEBUG)()
	// TODO consider to return the unmap SCBE response
	hostId, err := s.getHostIdByVol(wwn, host)
	if err != nil {
		return err
	}
	payload := ScbeUnMapVolumePostParams{VolumeId: wwn, HostId: hostId}
	payloadMarshal, err := json.Marshal(payload)
	if err != nil {
		return s.logger.ErrorRet(err, "json.Marshal failed", logs.Args{{"payload", payload}})
	}
	if err = s.client.Delete(UrlScbeResourceMapping, payloadMarshal, HTTP_SUCCEED_DELETED); err != nil {
		return s.logger.ErrorRet(err, "client.Delete failed", logs.Args{{"url", UrlScbeResourceMapping}})
	}
	return nil
}

func (s *scbeRestClient) GetVolMapping(wwn string) (string, error) {
	defer s.logger.Trace(logs.DEBUG)()
	return "", nil
}

func (s *scbeRestClient) ServiceExist(serviceName string) (exist bool, err error) {
	defer s.logger.Trace(logs.DEBUG)()
	var services []ScbeStorageService
	services, err = s.serviceList(serviceName)
	if err == nil {
		return len(services) > 0, err
	}
	return false, err
}

func (s *scbeRestClient) serviceList(serviceName string) ([]ScbeStorageService, error) {
	defer s.logger.Trace(logs.DEBUG)()
	payload := map[string]string{}
	if serviceName != "" {
		payload["name"] = serviceName
	}

	var services []ScbeStorageService
	if err := s.client.Get(UrlScbeResourceService, payload, -1, &services); err != nil {
		return nil, s.logger.ErrorRet(err, "client.Get failed")
	}

	return services, nil
}
func (s *scbeRestClient) volumeList(wwn string) ([]ScbeResponseVolume, error) {
	defer s.logger.Trace(logs.DEBUG)()
	payload := map[string]string{}
	if wwn != "" {
		payload["scsi_identifier"] = wwn
	}
	var volumes []ScbeResponseVolume
	if err := s.client.Get(UrlScbeResourceVolume, payload, -1, &volumes); err != nil {
		return nil, s.logger.ErrorRet(err, "client.Get failed")
	}

	return volumes, nil
}

func (s *scbeRestClient) hostList(payload map[string]string) ([]ScbeResponseHost, error) {
	defer s.logger.Trace(logs.DEBUG)()
	var hosts []ScbeResponseHost
	err := s.client.Get(UrlScbeResourceHost, payload, -1, &hosts)
	if err != nil {
		return nil, s.logger.ErrorRet(err, "client.Get failed")
	}
	return hosts, nil
}

//getHostIdByVol return the host ID from the storage system of the given volume(wwn)
func (s *scbeRestClient) getHostIdByVol(wwn string, host string) (int, error) {
	defer s.logger.Trace(logs.DEBUG)()
	vols, err := s.volumeList(wwn)
	if err != nil {
		return 0, s.logger.ErrorRet(err, "volumeList failed", logs.Args{{"wwn", wwn}})
	}

	if len(vols) == 0 {
		return 0, s.logger.ErrorRet(&volumeNotFoundError{wwn}, "failed")
	}
	vol := vols[0]
	payload := make(map[string]string)
	payload["array_id"] = vol.Array
	payload["name"] = host
	hosts, err := s.hostList(payload)
	if err != nil {
		return 0, s.logger.ErrorRet(err, "hostList failed")
	}
	if len(hosts) != 1 {
		return 0, s.logger.ErrorRet(&hostNotFoundvolumeNotFoundError{wwn, vol.Array, host}, "failed")
	}

	return hosts[0].Id, nil
}
