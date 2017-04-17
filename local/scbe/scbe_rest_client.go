package scbe

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"bytes"

	"github.com/IBM/ubiquity/utils"
)

// TODO change _ to camel case
/// Wrapper for http requests to provide easy REST API operations, help with parsing, exist status and token handling
type RestClient interface {
	// Authenticate the server, prepare headers and save the token
	Login() (string, error)
	// Obtain a new token and return it as string
	GetToken() (string, error)
	// Paper the payload, send post request and check expected status response and returned parsed response
	Post(resource_url string, payload map[string]string, exit_status int) ([]byte, error)
	// Paper the payload, send get request and check expected status response and returned parsed response
	Get(resource_url string, payload map[string]string, exit_status int) ([]byte, error)
	// Paper the payload, send delete request and check expected status response and returned parsed response
	Delete(resource_url string, payload map[string]string, exit_status int) ([]byte, error)
	// check that the status code of the response is as expected
	verifyStatusCode(response http.Response, expected_status_code int, action_name string) error
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
	headers := make(map[string]string)
	headers["Content-Type"] = "application/json"
	headers["referer"] = referrer

	return &restClient{logger: logger, connectionInfo: conInfo, baseURL: baseURL, authURL: authURL, referrer: referrer, httpClient: &http.Client{}, headers: headers}, nil
}

const HTTP_AUTH_KEY = "Authorization"

func (s *restClient) Login() (string, error) {
	token, err := s.GetToken()
	if err != nil {
		s.logger.Printf("Error in getting token %#v", err)
		return "", fmt.Errorf("Error in getting token")
	}
	s.headers[HTTP_AUTH_KEY] = "Token " + token
}

func (s *restClient) GetToken() (string, error) {
	delete(s.headers, HTTP_AUTH_KEY) // because no need token to get the token only user\password
	response, err := s.Post(s.authURL, s.connectionInfo.CredentialInfo, 200)
	var loginResponse = LoginResponse{}
	err = utils.UnmarshalResponse(response, &loginResponse)
	if err != nil {
		s.logger.Printf("Error in unmarshalling response %#v", err)
		return "", fmt.Errorf("Error in unmarshalling response")
	}
	return loginResponse.Token, nil
}

func (s *restClient) Post(resource_url string, payload map[string]string, exit_status int) ([]byte, error) {
	// TODO not sure about the payload type
	url := utils.FormatURL(s.baseURL, resource_url)

	payload, err := json.MarshalIndent(payload, "", " ")
	if err != nil {

		s.logger.Printf(fmt.Sprintf("Internal error marshalling credentialInfo %#v", err))
		return nil, fmt.Errorf("Internal error marshalling credentialInfo")
	}

	request, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		s.logger.Printf("Error in creating request %#v", err)
		return nil, fmt.Errorf("Error in creating request")
	}
	// append all the headers to the request
	for key, value := range s.headers {
		request.Header.Add(key, value)
	}

	response, err := s.httpClient.Do(request)
	fmt.Printf("response %#v", response)
	if err != nil {
		s.logger.Printf("Error in executing remote call %#v", err)
		return nil, fmt.Errorf("Error in executing remote call")
	}
	// TODO need to pars the json respons
	jsonDataFromHttp, err := ioutil.ReadAll(response.Body)
	if err != nil {
		s.logger.Printf("Error reading the response %#v", err)
		return "", fmt.Errorf("Error reading the response")
	}
	return jsonDataFromHttp, nil
}

func (s *restClient) Get(resource_url string, payload map[string]string, exit_status int) (map[string]string, error) {
	return nil, nil
}

func (s *restClient) Delete(resource_url string, payload map[string]string, exit_status int) (map[string]string, error) {
	return nil, nil
}
func (s *restClient) verifyStatusCode(response http.Response, expected_status_code int, action_name string) error {
	return nil
}

type ScbeVolumeInfo struct {
	name string
	wwn  string
	// TODO later on we will want also size and maybe other stuff
}

/// SCBE rest client
type ScbeRestClient interface {
	Login() (string, error)

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
	client         *restClient // TODO does it have to be * or not?
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
