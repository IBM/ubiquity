package scbe

import (
	"encoding/json"
	"fmt"
	"github.com/IBM/ubiquity/resources"
	"log"
)

//go:generate counterfeiter -o ../fakes/fake_scbe_rest_client.go . ScbeRestClient
type ScbeRestClient interface {
	Login() error
	CreateVolume(volName string, serviceName string, size int) (ScbeVolumeInfo, error)
	GetAllVolumes() ([]ScbeVolumeInfo, error)
	GetVolume(wwn string) (ScbeVolumeInfo, error)
	DeleteVolume(wwn string) error
	MapVolume(wwn string, host string) (ScbeResponseMapping, error)
	UnmapVolume(wwn string, host string) error
	GetVolMapping(wwn string) (string, error)
	ServiceExist(serviceName string) (bool, error)
}

type scbeRestClient struct {
	logger         *log.Logger
	connectionInfo resources.ConnectionInfo
	client         SimpleRestClient
}

const (
	DEFAULT_SCBE_PORT          = 8440
	URL_SCBE_REFERER           = "https://%s:%d/"
	URL_SCBE_BASE_SUFFIX       = "api/v1"
	URL_SCBE_RESOURCE_GET_AUTH = "users/get-auth-token"
	SCBE_FLOCKER_GROUP_PARAM   = "flocker"
	UrlScbeResourceService     = "services"
	UrlScbeResourceVolume      = "volumes"
	UrlScbeResourceMapping     = "mappings"
	UrlScbeResourceHost        = "hosts"
	DefaultSizeUnit            = "gb"
)


func NewScbeRestClient(logger *log.Logger, conInfo resources.ConnectionInfo) ScbeRestClient {
	return newScbeRestClient(logger, conInfo, nil)
}

// NewScbeRestClientWithNewRestClient for mocking during test # TODO consider to remove it to test file
func NewScbeRestClientWithSimpleRestClient(logger *log.Logger, conInfo resources.ConnectionInfo, simpleClient SimpleRestClient) ScbeRestClient {
	return newScbeRestClient(logger, conInfo, simpleClient)
}

func newScbeRestClient(logger *log.Logger, conInfo resources.ConnectionInfo, simpleClient SimpleRestClient) ScbeRestClient {
	// Set default SCBE port if not mentioned
	if conInfo.Port == 0 {
		conInfo.Port = DEFAULT_SCBE_PORT
	}
	// Add the default SCBE Flocker group to the credentials  # TODO change to ubiquity interface
	conInfo.CredentialInfo.Group = SCBE_FLOCKER_GROUP_PARAM

	if simpleClient == nil {
		referrer := fmt.Sprintf(URL_SCBE_REFERER, conInfo.ManagementIP, conInfo.Port)
		baseUrl := referrer + URL_SCBE_BASE_SUFFIX
		simpleClient = NewSimpleRestClient(logger, conInfo, baseUrl, URL_SCBE_RESOURCE_GET_AUTH, referrer)
	}
	return &scbeRestClient{logger, conInfo, simpleClient}
}


func (s *scbeRestClient) Login() error {
	return s.client.Login()
}

// CreateVolume provision new volume on SCBE storage service.
// Return ScbeVolumeInfo of the new volume that was created
// Errors:
//	if service don't exist
//	if fail to create the volume
func (s *scbeRestClient) CreateVolume(volName string, serviceName string, size int) (ScbeVolumeInfo, error) {
	// find the service in order to validate and also to get the service id
	services, err := s.serviceList(serviceName)
	if err != nil {
		return ScbeVolumeInfo{}, err
	}
	// check existence of the service
	if len(services) <= 0 || services[0].Name != serviceName {
		err = fmt.Errorf(fmt.Sprintf(
			MsgVolumeCreateFailBecauseNoServicesExist, volName, serviceName, s.connectionInfo.ManagementIP))
		return ScbeVolumeInfo{}, err
	}

	payload := ScbeCreateVolumePostParams{
		services[0].Id,
		volName,
		size,
		DefaultSizeUnit, // TODO lets support different type of unit size, for now only gb
	}

	payloadMarshaled, err := json.Marshal(payload)
	if err != nil {
		msg := fmt.Sprintf("Error in marshalling payload for create volume %#v", err)
		s.logger.Printf(msg)
		return ScbeVolumeInfo{}, fmt.Errorf(msg)
	}
	volResponse := ScbeResponseVolume{}
	err = s.client.Post(UrlScbeResourceVolume, payloadMarshaled, HTTP_SUCCEED_POST, &volResponse)
	if err != nil {
		msg := fmt.Sprintf("Fail to provision volume %#v on service %s, due to error: %#v", volName, serviceName, err)
		s.logger.Printf(msg)
		return ScbeVolumeInfo{}, fmt.Errorf(msg)
	}

	return ScbeVolumeInfo{volResponse.Name, volResponse.ScsiIdentifier, serviceName}, nil
}

func (s *scbeRestClient) GetAllVolumes() ([]ScbeVolumeInfo, error) {
	return nil, nil
}
func (s *scbeRestClient) GetVolume(wwn string) (ScbeVolumeInfo, error) {
	return ScbeVolumeInfo{}, nil
}
func (s *scbeRestClient) DeleteVolume(wwn string) error {
	urlToDelete := fmt.Sprintf("%s/%s", UrlScbeResourceVolume, wwn)
	err := s.client.Delete(urlToDelete, []byte{}, HTTP_SUCCEED_DELETED)
	if err != nil {
		msg := fmt.Sprintf("Fail to delete volume WWN %#v, due to error: %#v", wwn, err)
		s.logger.Printf(msg)
		return fmt.Errorf(msg)
	}
	return nil
}
func (s *scbeRestClient) MapVolume(wwn string, host string) (ScbeResponseMapping, error) {
	hostId, err := s.getHostIdByVol(wwn, host)
	if err != nil {
		return ScbeResponseMapping{}, err
	}
	payload := ScbeMapVolumePostParams{VolumeId: wwn, HostId: hostId}
	payloadMarshal, err := json.Marshal(payload)
	if err != nil {
		s.logger.Printf("Error in marshalling ScbeMapVolumePostParams %#v", err)
		return ScbeResponseMapping{}, fmt.Errorf("Error in marshalling ScbeMapVolumePostParams")
	}
	mappingsResponse := ScbeResponseMappings{}
	err = s.client.Post(UrlScbeResourceMapping, payloadMarshal, HTTP_SUCCEED_POST, &mappingsResponse)
	if err != nil {
		return ScbeResponseMapping{}, err
	}
	if len(mappingsResponse.Mappings) != 1 {
		msg := fmt.Sprintf(MsgMappingDoneButResponseNotOk, mappingsResponse.Mappings)
		s.logger.Printf(msg)
		return ScbeResponseMapping{}, fmt.Errorf(msg)
	}
	return mappingsResponse.Mappings[0], nil
}

func (s *scbeRestClient) UnmapVolume(wwn string, host string) error {
	// TODO consider to return the unmap SCBE response
	hostId, err := s.getHostIdByVol(wwn, host)
	if err != nil {
		return err
	}
	payload := ScbeUnMapVolumePostParams{VolumeId: wwn, HostId: hostId}
	payloadMarshal, err := json.Marshal(payload)
	if err != nil {
		msg := fmt.Sprintf("Error in marshalling ScbeMapVolumePostParams %#v", err)
		s.logger.Printf(msg)
		return fmt.Errorf(msg)
	}
	err = s.client.Delete(UrlScbeResourceMapping, payloadMarshal, HTTP_SUCCEED_DELETED)
	if err != nil {
		return err
	}
	return nil
}

func (s *scbeRestClient) GetVolMapping(wwn string) (string, error) {
	return "", nil
}

func (s *scbeRestClient) ServiceExist(serviceName string) (exist bool, err error) {
	var services []ScbeStorageService
	services, err = s.serviceList(serviceName)
	if err == nil {
		return len(services) > 0, err
	}
	return false, err
}

func (s *scbeRestClient) serviceList(serviceName string) ([]ScbeStorageService, error) {
	payload := make(map[string]string)
	var err error
	if serviceName == "" {
		payload = nil
	} else {
		payload["name"] = serviceName
	}
	var services []ScbeStorageService
	err = s.client.Get(UrlScbeResourceService, payload, -1, &services)
	if err != nil {
		return nil, err
	}

	return services, nil
}
func (s *scbeRestClient) volumeList(wwn string) ([]ScbeResponseVolume, error) {
	payload := make(map[string]string)
	var err error
	if wwn == "" {
		payload = nil
	} else {
		payload["scsi_identifier"] = wwn
	}
	var volumes []ScbeResponseVolume
	err = s.client.Get(UrlScbeResourceVolume, payload, -1, &volumes)
	if err != nil {
		return nil, err
	}

	return volumes, nil
}

func (s *scbeRestClient) hostList(payload map[string]string) ([]ScbeResponseHost, error) {
	var hosts []ScbeResponseHost
	err := s.client.Get(UrlScbeResourceHost, payload, -1, &hosts)
	if err != nil {
		return nil, err
	}
	return hosts, nil
}

//getHostIdByVol return the host ID from the storage system of the given volume(wwn)
func (s *scbeRestClient) getHostIdByVol(wwn string, host string) (int, error) {
	vols, err := s.volumeList(wwn)
	if err != nil {
		s.logger.Printf("Fail to get volume of wwn ", wwn, " due to error ", err)
		return 0, err
	}

	if len(vols) == 0 {
		msg := fmt.Sprintf(MsgVolumeWWNNotFound, wwn)
		s.logger.Printf(msg)
		return 0, fmt.Errorf(msg)
	}
	vol := vols[0]
	payload := make(map[string]string)
	payload["array_id"] = vol.Array
	payload["name"] = host
	hosts, err := s.hostList(payload)
	if len(hosts) != 1 {
		msg := fmt.Sprintf(MsgHostIDNotFoundByVolWWNOnArray, host, vol.Array, wwn, hosts)
		s.logger.Printf(msg)
		return 0, fmt.Errorf(msg)
	}

	return hosts[0].Id, nil
}
