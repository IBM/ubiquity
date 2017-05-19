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

// TODO change _ to camel case
/// Wrapper for http requests to provide easy REST API operations, help with parsing, exist status and token handling
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

const (
	HTTP_SUCCEED         = 200
	HTTP_SUCCEED_POST    = 201
	HTTP_SUCCEED_DELETED = 204
)

type restClient struct {
	logger         *log.Logger
	baseURL        string
	authURL        string
	referrer       string
	connectionInfo resources.ConnectionInfo
	httpClient     *http.Client
	token          string
	headers        map[string]string
}

func NewRestClient(logger *log.Logger, conInfo resources.ConnectionInfo, baseURL string, authURL string, referrer string) (RestClient, error) {
	headers := map[string]string{
		"Content-Type": "application/json",
		"referer":      referrer,
	}
	transCfg := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // ignore expired SSL certificates TODO to use
	}
	client := &http.Client{Transport: transCfg}
	return &restClient{logger: logger, connectionInfo: conInfo, baseURL: baseURL, authURL: authURL, referrer: referrer, httpClient: client, headers: headers}, nil
}

const HTTP_AUTH_KEY = "Authorization"

func (s *restClient) Login() error {
	token, err := s.getToken()
	if err != nil {
		s.logger.Printf("Error in getting token %#v", err)
		return err
	}
	if token == "" {
		s.logger.Printf("Error, token is empty")
		return fmt.Errorf("Error, token is empty")
	}
	s.headers[HTTP_AUTH_KEY] = "Token " + token

	return nil
}

func (s *restClient) getToken() (string, error) {
	delete(s.headers, HTTP_AUTH_KEY) // because no need token to get the token only user\password
	var loginResponse = LoginResponse{}

	credentials, err := json.Marshal(s.connectionInfo.CredentialInfo)
	if err != nil {
		s.logger.Printf("Error in marshalling CredentialInfo %#v", err)
		return "", fmt.Errorf("Error in marshalling CredentialInfo")
	}

	err = s.Post(s.authURL, credentials, HTTP_SUCCEED, &loginResponse)
	if err != nil {
		s.logger.Printf("Error posting to url %#v to get a token %#v", s.authURL, err)
		return "", err
	}

	if loginResponse.Token == "" {
		return "", fmt.Errorf("Token is empty")
	}

	return loginResponse.Token, nil
}

func (s *restClient) genericAction(actionName string, resource_url string, payload []byte, exitStatus int, v interface{}) error {
	url := utils.FormatURL(s.baseURL, resource_url)

	reader := bytes.NewReader(payload)
	request, err := http.NewRequest(actionName, url, reader)
	if err != nil {
		s.logger.Printf("Error in creating %s request %#v", actionName, err)
		return fmt.Errorf("Error in creating request")
	}
	// append all the headers to the request
	for key, value := range s.headers {
		request.Header.Add(key, value)
	}

	response, err := s.httpClient.Do(request)
	if err != nil {
		s.logger.Printf("Error sending %s request : %#v", actionName, err)
		return err
	}

	defer response.Body.Close()
	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		s.logger.Printf("Fail to read the body %#v", err)
		return err
	}

	err = s.verifyStatusCode(*response, exitStatus) // &dereference
	if err != nil {
		s.logger.Printf("Status code is wrong %#v", err)
		return err
	}

	err = json.Unmarshal(data, v)
	if err != nil {
		s.logger.Printf("Error unmarshal %#v", err)
		return err
	}

	return nil
}

func (s *restClient) Post(resource_url string, payload []byte, exitStatus int, v interface{}) error {
	if exitStatus < 0 {
		exitStatus = HTTP_SUCCEED_POST // Default value
	}
	return s.genericAction("POST", resource_url, payload, exitStatus, v)
}

func (s *restClient) Get(resource_url string, params map[string]string, exitStatus int, v interface{}) error {
	if exitStatus < 0 {
		exitStatus = HTTP_SUCCEED // Default value
	}
	url := utils.FormatURL(s.baseURL, resource_url)

	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		s.logger.Printf("Error in creating GET request %#v", err)
		return fmt.Errorf("Error in creating request")
	}

	// append all the headers to the request
	for key, value := range s.headers {
		request.Header.Add(key, value)
	}

	// append all the params into the request
	q := request.URL.Query()
	for key, value := range params {
		q.Add(key, value)
	}
	request.URL.RawQuery = q.Encode()

	response, err := s.httpClient.Do(request)
	defer response.Body.Close()
	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		s.logger.Printf("Fail to read the body %#v", err)
		return err
	}

	err = s.verifyStatusCode(*response, exitStatus) // &dereference
	if err != nil {
		s.logger.Printf("Status code is wrong %#v", err)
		return err
	}

	err = json.Unmarshal(data, v)
	if err != nil {
		s.logger.Printf("Error unmarshal %#v", err)
		return err
	}

	return nil
}

func (s *restClient) Delete(resource_url string, payload []byte, exitStatus int, v interface{}) error {
	if exitStatus < 0 {
		exitStatus = HTTP_SUCCEED_DELETED // Default value
	}
	return s.genericAction("DELETE", resource_url, payload, exitStatus, v)
}

func (s *restClient) verifyStatusCode(response http.Response, expected_status_code int) error {
	if response.StatusCode != expected_status_code {
		fmt.Printf("------->response statuc code, %#v", response)
		s.logger.Printf("Error, bad status code of http response %#v", response.StatusCode)
		return fmt.Errorf("Error, bad status code of http response")
	}
	return nil
}

type ScbeVolumeInfo struct {
	name string
	wwn  string
	// TODO later on we will want also size and maybe other stuff
}

/// SCBE rest client
type ScbeRestClient interface {
	Login() error
	CreateVolume(volName string, serviceName string, size_byte int) (ScbeVolumeInfo, error)
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
	URL_SCBE_RESOURCE_GET_AUTH = "/users/get-auth-token"
	SCBE_FLOCKER_GROUP_PARAM   = "flocker"
	UrlScbeResourceService     = "/services"
	//UrlScbeResourceVolume = "/volumes"
	//UrlScbeResourceMapping = "/mappings"
	//UrlScbeResourceHost = "/hosts"
)

func NewScbeRestClient(logger *log.Logger, conInfo resources.ConnectionInfo) (ScbeRestClient, error) {
	// Set default SCBE port if not mentioned
	if conInfo.Port == 0 {
		conInfo.Port = DEFAULT_SCBE_PORT
	}
	// Add the default SCBE Flocker group to the credentials # TODO need to update with ubiquity group later on
	conInfo.CredentialInfo.Group = SCBE_FLOCKER_GROUP_PARAM
	referrer := fmt.Sprintf(URL_SCBE_REFERER, conInfo.ManagementIP, conInfo.Port)
	baseUrl := referrer + URL_SCBE_BASE_SUFFIX
	client, _ := NewRestClient(logger, conInfo, baseUrl, URL_SCBE_RESOURCE_GET_AUTH, referrer)

	return &scbeRestClient{logger, conInfo, client}, nil
}

func (s *scbeRestClient) Login() error {
	return s.client.Login()
}

func (s *scbeRestClient) CreateVolume(volName string, serviceName string, size_byte int) (ScbeVolumeInfo, error) {
	return ScbeVolumeInfo{}, nil
}
func (s *scbeRestClient) GetAllVolumes() ([]ScbeVolumeInfo, error) {
	return nil, nil
}
func (s *scbeRestClient) GetVolume(wwn string) (ScbeVolumeInfo, error) {
	return ScbeVolumeInfo{}, nil
}
func (s *scbeRestClient) DeleteVolume(wwn string) error {
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
		exist = len(services) > 0
	}
	return
}

func (s *scbeRestClient) serviceList(serviceName string) ([]ScbeStorageService, error) {
	payload := make(map[string]string)
	var err error
	if serviceName == "" {
		payload = nil // TODO else
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
