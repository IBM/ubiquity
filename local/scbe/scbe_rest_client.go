package scbe

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
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
	Post(resource_url string, payload []byte, exit_status int) ([]byte, error)
	// Paper the payload, send get request and check expected status response and returned parsed response
	Get(resource_url string, payload []byte, exit_status int) ([]byte, error)
	// Paper the payload, send delete request and check expected status respon		se and returned parsed response
	Delete(resource_url string, payload []byte, exit_status int) ([]byte, error)
}

type restClient struct {
	logger         *log.Logger
	baseURL        string
	authURL        string
	referrer       string
	connectionInfo ConnectionInfo
	httpClient     *http.Client
	token          string
	headers        map[string]string
}

func NewRestClient(logger *log.Logger, conInfo ConnectionInfo, baseURL string, authURL string, referrer string) (RestClient, error) {
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

	credentials, err := json.Marshal(s.connectionInfo.CredentialInfo)
	if err != nil {
		s.logger.Printf("Error in marshalling CredentialInfo %#v", err)
		return "", fmt.Errorf("Error in marshalling CredentialInfo")
	}

	responseBody, err := s.Post(s.authURL, credentials, 200)
	if err != nil {
		s.logger.Printf("Error posting to url %#v to get a token %#v", s.authURL, err)
		return "", err
	}

	var loginResponse = LoginResponse{}
	err = json.Unmarshal(responseBody, &loginResponse)
	if err != nil {
		return "", err
	}

	if loginResponse.Token == "" {
		return "", fmt.Errorf("Token is empty")
	}

	return loginResponse.Token, nil
}

func (s *restClient) Post(resource_url string, payload []byte, exitStatus int) ([]byte, error) {
	url := utils.FormatURL(s.baseURL, resource_url)

	reader := bytes.NewReader(payload)
	request, err := http.NewRequest("POST", url, reader)
	if err != nil {
		s.logger.Printf("Error in creating request %#v", err)
		return nil, fmt.Errorf("Error in creating request")
	}
	// append all the headers to the request
	for key, value := range s.headers {
		request.Header.Add(key, value)
	}

	response, err := s.httpClient.Do(request)
	//defer response.Body.Close()
	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		//TODO logging
		return nil, err
	}

	err = s.verifyStatusCode(*response, exitStatus) // &dereference
	if err != nil {
		//TODO logging
		return nil, err
	}
	return data, nil
}

func (s *restClient) Get(resource_url string, payload []byte, exit_status int) ([]byte, error) {
	return nil, nil
}

func (s *restClient) Delete(resource_url string, payload []byte, exit_status int) ([]byte, error) {
	return nil, nil
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
	ServiceExist(serviceName string) bool
}

type scbeRestClient struct {
	logger         *log.Logger
	connectionInfo ConnectionInfo
	client         RestClient
}

const DEFAULT_SCBE_PORT = 8440
const URL_SCBE_REFERER = "https://%s:%d/"
const URL_SCBE_BASE_SUFFIX = "api/v1"
const SCBE_FLOCKER_GROUP_PARAM = "flocker"
const URL_SCBE_RESOURCE_GET_AUTH = "/users/get-auth-token"

func NewScbeRestClient(logger *log.Logger, conInfo ConnectionInfo) (ScbeRestClient, error) {
	// Set default SCBE port if not mentioned
	if conInfo.Port == "" {
		conInfo.Port = string(DEFAULT_SCBE_PORT)
	}
	// Add the default SCBE Flocker group to the credentials
	conInfo.CredentialInfo.Group = SCBE_FLOCKER_GROUP_PARAM
	referrer := fmt.Sprintf(URL_SCBE_REFERER, conInfo.ManagementIP, conInfo.Port)
	baseUrl := referrer + URL_SCBE_BASE_SUFFIX

	client, _ := NewRestClient(logger, conInfo, baseUrl, URL_SCBE_RESOURCE_GET_AUTH, referrer)

	return &scbeRestClient{logger, conInfo, client}, nil
}

func (s *scbeRestClient) Login() error {
	return nil
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

func (s *scbeRestClient) ServiceExist(serviceName string) bool {
	return true
}
