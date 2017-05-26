package scbe

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/IBM/ubiquity/resources"
	"github.com/IBM/ubiquity/utils"
	"io/ioutil"
	"log"
	"net/http"
)

// RestClient is an interface that wrapper the http requests to provide easy REST API operations,
//go:generate counterfeiter -o ../fakes/fake_scbe_rest_client.go . RestClient
type RestClient interface {
	// Authenticate the server, prepare headers and save the token
	Login() error

	// Paper the payload, send post request and check expected status response and returned parsed response
	Post(resource_url string, payload []byte, exitStatus int, v interface{}) error

	// Paper the payload, send get request and check expected status response and returned parsed response
	Get(resource_url string, params map[string]string, exitStatus int, v interface{}) error

	// Paper the payload, send delete request and check expected status respon		se and returned parsed response
	Delete(resource_url string, payload []byte, exitStatus int, v interface{}) error
}

// TODO consider to move this RestClient into different go file named restclient.go
const (
	HTTP_SUCCEED         = 200
	HTTP_SUCCEED_POST    = 201
	HTTP_SUCCEED_DELETED = 204
	HTTP_AUTH_KEY        = "Authorization"
)

// restClient implements RestClient interface.
// The implementation of each interface simplify the use of REST API by doing all the rest and json ops,
// like pars the response result, handling json, marshaling, and token expire handling.
type restClient struct {
	logger         *log.Logger
	baseURL        string
	authURL        string
	referrer       string
	connectionInfo resources.ConnectionInfo
	httpClient     *http.Client
	headers        map[string]string
}

func NewRestClient(logger *log.Logger, conInfo resources.ConnectionInfo, baseURL string, authURL string, referrer string) RestClient {
	headers := map[string]string{
		"Content-Type": "application/json",
		"referer":      referrer,
	}
	var client *http.Client

	if conInfo.SkipVerifySSL {
		transCfg := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
		client = &http.Client{Transport: transCfg}
	} else {
		client = &http.Client{}
	}
	return &restClient{logger: logger, connectionInfo: conInfo, baseURL: baseURL, authURL: authURL, referrer: referrer, httpClient: client, headers: headers}
}

func (s *restClient) Login() error {
	err := s.getToken()
	if err != nil {
		s.logger.Printf("Error in getting token %#v", err)
		return err
	}

	return nil
}

func (s *restClient) getToken() error {
	delete(s.headers, HTTP_AUTH_KEY) // because no need token to get the token only user\password
	var loginResponse = LoginResponse{}

	credentials, err := json.Marshal(s.connectionInfo.CredentialInfo)
	if err != nil {
		s.logger.Printf("Error in marshalling CredentialInfo %#v", err)
		return fmt.Errorf("Error in marshalling CredentialInfo")
	}

	err = s.Post(s.authURL, credentials, HTTP_SUCCEED, &loginResponse)
	if err != nil {
		s.logger.Printf("Error posting to url %#v to get a token %#v", s.authURL, err)
		return err
	}

	if loginResponse.Token == "" {
		return fmt.Errorf("Token is empty")
	}
	s.headers[HTTP_AUTH_KEY] = "Token " + loginResponse.Token

	return nil
}

// genericAction trigger the http actionName give.
// It first format the url, prepare the http.Request(if post\delete uses payload, if get uses params)
// Then it append all relevant the http headers and then trigger the http action by using Do interface.
// Then read the response, and if exist status as expacted it reads the body into the given struct(v)
// The function return only error if accured and of cause the object(v) loaded with the response.
func (s *restClient) genericAction(actionName string, resource_url string, payload []byte, params map[string]string, exitStatus int, v interface{}) error {
	url := utils.FormatURL(s.baseURL, resource_url)
	var err error
	var request *http.Request
	if actionName == "GET" {
		request, err = http.NewRequest(actionName, url, nil)
	} else {
		// TODO : consider to add
		request, err = http.NewRequest(actionName, url, bytes.NewReader(payload))
	}

	if err != nil {
		msg := fmt.Sprintf("Error in creating %s request %#v", actionName, err)
		s.logger.Printf(msg)
		return fmt.Errorf(msg)
	}
	if actionName == "GET" {
		// append all the params into the request
		q := request.URL.Query()
		for key, value := range params {
			q.Add(key, value)
		}
		request.URL.RawQuery = q.Encode()
	}

	// append all the headers to the request
	for key, value := range s.headers {
		request.Header.Add(key, value)
	}

	response, err := s.httpClient.Do(request)
	if err != nil {
		s.logger.Printf("Error sending %s request url [%s]: %#v", actionName, url, err)
		return err
	}

	// check if client sent a token and it expired
	if response.StatusCode == http.StatusUnauthorized && s.headers[HTTP_AUTH_KEY] != "" {

		// login
		err = s.Login()
		if err != nil {
			s.logger.Printf("Login error: %#v", err)
			return err
		}

		// retry
		response, err = s.httpClient.Do(request)
		if err != nil {
			s.logger.Printf("Error sending %s request url [%s]: %#v", actionName, url, err)
			return err
		}
	}

	defer response.Body.Close()
	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		s.logger.Printf("Fail to read the body %#v", err)
		return err
	}

	if response.StatusCode != exitStatus {
		msg := fmt.Sprintf("Error, bad status code [%#v] of http response to url [%s].", response.StatusCode, url)
		s.logger.Printf(msg)
		return fmt.Errorf(msg)
	}
	if v != nil {
		s.logger.Printf("Unmarsh the post data into the given interface")
		err = json.Unmarshal(data, v)
		if err != nil {
			s.logger.Printf("Error unmarshal %#v", err)
			return err
		}
	}

	return nil
}

// Post http request
func (s *restClient) Post(resource_url string, payload []byte, exitStatus int, v interface{}) error {
	if exitStatus < 0 {
		exitStatus = HTTP_SUCCEED_POST // Default value
	}
	return s.genericAction("POST", resource_url, payload, nil, exitStatus, v)
}

// Get http request
func (s *restClient) Get(resource_url string, params map[string]string, exitStatus int, v interface{}) error {
	if exitStatus < 0 {
		exitStatus = HTTP_SUCCEED // Default value
	}
	return s.genericAction("GET", resource_url, nil, params, exitStatus, v)
}

// Delete request
func (s *restClient) Delete(resource_url string, payload []byte, exitStatus int, v interface{}) error {
	if exitStatus < 0 {
		exitStatus = HTTP_SUCCEED_DELETED // Default value
	}
	return s.genericAction("DELETE", resource_url, payload, nil, exitStatus, v)
}

// ********************************
// ****** SCBE Rest Client ********
// ********************************

//go:generate counterfeiter -o ../fakes/fake_scbe_rest_client.go . ScbeRestClient
type ScbeRestClient interface {
	Login() error
	CreateVolume(volName string, serviceName string, size int) (ScbeVolumeInfo, error)
	GetAllVolumes() ([]ScbeVolumeInfo, error)
	GetVolume(wwn string) (ScbeVolumeInfo, error)
	DeleteVolume(wwn string) error
	MapVolume(wwn string, host string) error
	UnmapVolume(wwn string, host string) error
	GetVolMapping(wwn string) (string, error)
	ServiceExist(serviceName string) (bool, error)
}

type scbeRestClient struct {
	logger         *log.Logger
	connectionInfo resources.ConnectionInfo
	client         RestClient
}

const (
	DEFAULT_SCBE_PORT          = 8440
	URL_SCBE_REFERER           = "https://%s:%d/"
	URL_SCBE_BASE_SUFFIX       = "api/v1"
	URL_SCBE_RESOURCE_GET_AUTH = "users/get-auth-token"
	SCBE_FLOCKER_GROUP_PARAM   = "flocker"
	UrlScbeResourceService     = "services"
	UrlScbeResourceVolume      = "volumes"
	//UrlScbeResourceMapping = "/mappings"
	//UrlScbeResourceHost = "/hosts"
	DefaultSizeUnit = "gb"
)

func NewScbeRestClient(logger *log.Logger, conInfo resources.ConnectionInfo) (ScbeRestClient, error) {
	// Set default SCBE port if not mentioned
	if conInfo.Port == 0 {
		conInfo.Port = DEFAULT_SCBE_PORT
	}
	// Add the default SCBE Flocker group to the credentials
	conInfo.CredentialInfo.Group = SCBE_FLOCKER_GROUP_PARAM
	referrer := fmt.Sprintf(URL_SCBE_REFERER, conInfo.ManagementIP, conInfo.Port)
	baseUrl := referrer + URL_SCBE_BASE_SUFFIX
	client := NewRestClient(logger, conInfo, baseUrl, URL_SCBE_RESOURCE_GET_AUTH, referrer)
	return &scbeRestClient{logger, conInfo, client}, nil
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
	err := s.client.Delete(urlToDelete, nil, HTTP_SUCCEED_DELETED, nil)
	if err != nil {
		msg := fmt.Sprintf("Fail to delete volume WWN %#v, due to error: %#v", wwn, err)
		s.logger.Printf(msg)
		return fmt.Errorf(msg)
	}
	return nil
}

func (s *scbeRestClient) MapVolume(wwn string, host string) error {
	return nil

}
func (s *scbeRestClient) UnmapVolume(wwn string, host string) error {
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
