package scbe

import (
    "github.com/IBM/ubiquity/resources"
    "github.com/IBM/ubiquity/utils"
    "crypto/tls"
    "fmt"
    "log"
    "bytes"
    "io/ioutil"
    "net/http"
    "encoding/json"
)

// SimpleRestClient is an interface that wrapper the http requests to provide easy REST API operations,
//go:generate counterfeiter -o ../fakes/fake_simple_rest_client.go . SimpleRestClient
type SimpleRestClient interface {
    // Authenticate the server, prepare headers and save the token
    Login() error

    // send POST request with optional payload and check expected status of response
    Post(resource_url string, payload []byte, exitStatus int, v interface{}) error

    // send GET request with optional params and check expected status of response
    Get(resource_url string, params map[string]string, exitStatus int, v interface{}) error

    // send DELETE request with optional payload and check expected status of response
    Delete(resource_url string, payload []byte, exitStatus int) error
}

const (
    HTTP_SUCCEED         = 200
    HTTP_SUCCEED_POST    = 201
    HTTP_SUCCEED_DELETED = 204
    HTTP_AUTH_KEY        = "Authorization"
)

// simpleRestClient implements SimpleRestClient interface.
// The implementation of each interface simplify the use of REST API by doing all the rest and json ops,
// like pars the response result, handling json, marshaling, and token expire handling.
type simpleRestClient struct {
    logger         *log.Logger
    baseURL        string
    authURL        string
    referrer       string
    connectionInfo resources.ConnectionInfo
    httpClient     *http.Client
    headers        map[string]string
}

func NewSimpleRestClient(logger *log.Logger, conInfo resources.ConnectionInfo, baseURL string, authURL string, referrer string) SimpleRestClient {
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
    return &simpleRestClient{logger: logger, connectionInfo: conInfo, baseURL: baseURL, authURL: authURL, referrer: referrer, httpClient: client, headers: headers}
}

func (s *simpleRestClient) Login() error {
    err := s.getToken()
    if err != nil {
        s.logger.Printf("Error in getting token %#v", err)
        return err
    }

    return nil
}

func (s *simpleRestClient) getToken() error {
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
func (s *simpleRestClient) genericAction(actionName string, resource_url string, payload []byte, params map[string]string, exitStatus int, v interface{}) error {
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

    msg := fmt.Sprintf("The response Data from request url [%s] with action [%s] is : %#v", url, actionName, string(data[:]))
    s.logger.Printf(msg)
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
func (s *simpleRestClient) Post(resource_url string, payload []byte, exitStatus int, v interface{}) error {
    if exitStatus < 0 {
        exitStatus = HTTP_SUCCEED_POST // Default value
    }
    return s.genericAction("POST", resource_url, payload, nil, exitStatus, v)
}

// Get http request
func (s *simpleRestClient) Get(resource_url string, params map[string]string, exitStatus int, v interface{}) error {
    if exitStatus < 0 {
        exitStatus = HTTP_SUCCEED // Default value
    }
    return s.genericAction("GET", resource_url, nil, params, exitStatus, v)
}

// Delete request
func (s *simpleRestClient) Delete(resource_url string, payload []byte, exitStatus int) error {
    if exitStatus < 0 {
        exitStatus = HTTP_SUCCEED_DELETED // Default value
    }
    return s.genericAction("DELETE", resource_url, payload, nil, exitStatus, nil)
}
