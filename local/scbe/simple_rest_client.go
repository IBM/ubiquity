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

package scbe

import (
    "github.com/IBM/ubiquity/resources"
    "github.com/IBM/ubiquity/utils"
    "github.com/IBM/ubiquity/utils/logs"
    "crypto/tls"
    "bytes"
    "io/ioutil"
    "net/http"
    "encoding/json"
    "errors"
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
    logger         logs.Logger
    baseURL        string
    authURL        string
    referrer       string
    connectionInfo resources.ConnectionInfo
    httpClient     *http.Client
    headers        map[string]string
}

func NewSimpleRestClient(conInfo resources.ConnectionInfo, baseURL string, authURL string, referrer string) SimpleRestClient {
    headers := map[string]string{"Content-Type": "application/json", "referer": referrer}
    client := &http.Client{}
    if conInfo.SkipVerifySSL {
        client.Transport = &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
    }
    return &simpleRestClient{logger: logs.GetLogger(), connectionInfo: conInfo, baseURL: baseURL, authURL: authURL, referrer: referrer, httpClient: client, headers: headers}
}

func (s *simpleRestClient) Login() error {
    defer s.logger.Trace(logs.DEBUG)()
    if err := s.getToken(); err != nil {
        return s.logger.ErrorRet(err, "getToken failed")
    }
    return nil
}

func (s *simpleRestClient) getToken() error {
    defer s.logger.Trace(logs.DEBUG)()
    delete(s.headers, HTTP_AUTH_KEY) // because no need token to get the token only user\password
    var loginResponse = LoginResponse{}
    credentials, err := json.Marshal(s.connectionInfo.CredentialInfo)
    if err != nil {
        return s.logger.ErrorRet(err, "json.Marshal failed")
    }
    if err = s.Post(s.authURL, credentials, HTTP_SUCCEED, &loginResponse); err != nil {
        return s.logger.ErrorRet(err, "Post failed")
    }
    if loginResponse.Token == "" {
        return s.logger.ErrorRet(errors.New("Token is empty"), "Post failed")
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
    return s.genericActionInternal(actionName, resource_url, payload, params, exitStatus, v, true)
}

func (s *simpleRestClient) genericActionInternal(actionName string, resource_url string, payload []byte, params map[string]string, exitStatus int, v interface{}, retryUnauthorized bool) error {
    defer s.logger.Trace(logs.DEBUG)()
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
        return s.logger.ErrorRet(err, "http.NewRequest failed", logs.Args{{actionName, url}})
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
        return s.logger.ErrorRet(err, "httpClient.Do failed", logs.Args{{actionName, request.URL}})
    }

    // check if client sent a token and it expired
    if response.StatusCode == http.StatusUnauthorized && s.headers[HTTP_AUTH_KEY] != "" && retryUnauthorized {

        // login
        if err = s.Login(); err != nil {
            return s.logger.ErrorRet(err, "Login failed")
        }

        // retry
        return s.genericActionInternal(actionName, resource_url, payload, params, exitStatus, v, false)
    }

    defer response.Body.Close()
    data, err := ioutil.ReadAll(response.Body)
    if err != nil {
        return s.logger.ErrorRet(err, "ioutil.ReadAll failed")
    }

    s.logger.Debug(actionName + " " + url, logs.Args{{"data", string(data[:])}})
    if response.StatusCode != exitStatus {
        return s.logger.ErrorRet(errors.New("bad status code " + response.Status), "failed", logs.Args{{actionName, url}})
    }

    if v != nil {
        if err = json.Unmarshal(data, v); err != nil {
            return s.logger.ErrorRet(err, "json.Unmarshal failed", logs.Args{{actionName, url}})
        }
    }

    return nil
}

// Post http request
func (s *simpleRestClient) Post(resource_url string, payload []byte, exitStatus int, v interface{}) error {
    defer s.logger.Trace(logs.DEBUG)()
    if exitStatus < 0 {
        exitStatus = HTTP_SUCCEED_POST // Default value
    }
    return s.genericAction("POST", resource_url, payload, nil, exitStatus, v)
}

// Get http request
func (s *simpleRestClient) Get(resource_url string, params map[string]string, exitStatus int, v interface{}) error {
    defer s.logger.Trace(logs.DEBUG)()
    if exitStatus < 0 {
        exitStatus = HTTP_SUCCEED // Default value
    }
    return s.genericAction("GET", resource_url, nil, params, exitStatus, v)
}

// Delete request
func (s *simpleRestClient) Delete(resource_url string, payload []byte, exitStatus int) error {
    defer s.logger.Trace(logs.DEBUG)()
    if exitStatus < 0 {
        exitStatus = HTTP_SUCCEED_DELETED // Default value
    }
    return s.genericAction("DELETE", resource_url, payload, nil, exitStatus, nil)
}
