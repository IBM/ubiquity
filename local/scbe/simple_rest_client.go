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
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/IBM/ubiquity/resources"
	"github.com/IBM/ubiquity/utils"
	"github.com/IBM/ubiquity/utils/logs"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"
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
	KEY_VERIFY_SCBE_CERT = "UBIQUITY_SERVER_VERIFY_SCBE_CERT"
)

// simpleRestClient implements SimpleRestClient interface.
// The implementation of each interface simplify the use of REST API by doing all the rest and json ops,
// like pars the response result, handling json, marshaling, and token expire handling.
type simpleRestClient struct {
	logger         logs.Logger
	baseURL        string
	referrer       string
	connectionInfo resources.ConnectionInfo
	httpClient     *http.Client
	headers        *sync.Map
}

func NewSimpleRestClient(conInfo resources.ConnectionInfo, baseURL string, referrer string) (SimpleRestClient, error) {
	client := &simpleRestClient{logger: logs.GetLogger(), connectionInfo: conInfo, baseURL: baseURL, referrer: referrer, httpClient: &http.Client{}}
	client.initHeader()
	if err := client.initTransport(); err != nil {
		return nil, client.logger.ErrorRet(err, "client.initTransport failed")
	}
	return client, nil
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
	s.headers.Delete(HTTP_AUTH_KEY) // because no need token to get the token only user\password
	var loginResponse = LoginResponse{}
	credentials, err := json.Marshal(s.connectionInfo.CredentialInfo)
	if err != nil {
		return s.logger.ErrorRet(err, "json.Marshal failed")
	}
	if err = s.Post(UrlScbeResourceGetAuth, credentials, HTTP_SUCCEED, &loginResponse); err != nil {
		return s.logger.ErrorRet(err, "Post failed")
	}
	if loginResponse.Token == "" {
		return s.logger.ErrorRet(errors.New("Token is empty"), "Post failed")
	}
	s.headers.Store(HTTP_AUTH_KEY, "Token "+loginResponse.Token)
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
	var err error
	var request *http.Request

	url := utils.FormatURL(s.baseURL, resource_url)
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
	s.addHeader(request)

	response, err := s.httpClient.Do(request)
	if err != nil {
		return s.logger.ErrorRet(err, "httpClient.Do failed", logs.Args{{actionName, request.URL}})
	}

	defer response.Body.Close()

	// check if client sent a token and it expired
	if response.StatusCode == http.StatusUnauthorized {
		if retryUnauthorized && resource_url != UrlScbeResourceGetAuth {

			// login
			if err = s.Login(); err != nil {
				return s.logger.ErrorRet(err, "Login failed")
			}

			// retry
			return s.genericActionInternal(actionName, resource_url, payload, params, exitStatus, v, false)
		}
	}

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return s.logger.ErrorRet(err, "ioutil.ReadAll failed")
	}

	httpDataStr := string(data[:])
	s.logger.Debug(actionName+" "+url, logs.Args{{"data", httpDataStr}})
	if response.StatusCode != exitStatus {
		return s.logger.ErrorRet(&BadHttpStatusCodeError{
			httpStatusCode:         response.StatusCode,
			httpExpectedStatusCode: exitStatus,
			httpDataStr:            httpDataStr,
			httpAction:             actionName,
			httpUrl:                url,
		}, "failed")
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

func (s *simpleRestClient) initTransport() error {
	defer s.logger.Trace(logs.DEBUG)()
	exec := utils.NewExecutor()

	emptyConnection := resources.ConnectionInfo{}
	if s.connectionInfo != emptyConnection {
		sslMode := strings.ToLower(os.Getenv(resources.KeyScbeSslMode))
		if sslMode == "" {
			sslMode = resources.DefaultScbeSslMode
		}

		if sslMode == resources.SslModeVerifyFull {
			verifyFileCA := os.Getenv(KEY_VERIFY_SCBE_CERT)
			if verifyFileCA != "" {
				if _, err := exec.Stat(verifyFileCA); err != nil {
					return s.logger.ErrorRet(err, "failed")
				}
				caCert, err := ioutil.ReadFile(verifyFileCA)
				if err != nil {
					return s.logger.ErrorRet(err, "failed")
				}
				caCertPool := x509.NewCertPool()
				if ok := caCertPool.AppendCertsFromPEM(caCert); !ok {
					return fmt.Errorf("parse %v failed", verifyFileCA)
				}
				s.httpClient.Transport = &http.Transport{TLSClientConfig: &tls.Config{RootCAs: caCertPool}}
				s.logger.Info("", logs.Args{{KEY_VERIFY_SCBE_CERT, verifyFileCA}})
			} else {
				return s.logger.ErrorRet(&SslModeFullVerifyWithoutCAfile{KEY_VERIFY_SCBE_CERT}, "failed")
			}
		} else if sslMode == resources.SslModeRequire {
			s.logger.Info(
				fmt.Sprintf("Client SSL Mode set to [%s]. Means the communication to ubiquity is InsecureSkipVerify", sslMode))
			s.httpClient.Transport = &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
		} else {
			return s.logger.ErrorRet(&SslModeValueInvalid{sslMode}, "failed")
		}
	}
	return nil
}

func (s *simpleRestClient) addHeader(request *http.Request) error {
	s.headers.Range(func(k, v interface{}) bool {
		request.Header.Add(k.(string), v.(string))
		return true
	})
	return nil
}

func (s *simpleRestClient) initHeader() {
	s.headers = new(sync.Map)
	s.headers.Store("Content-Type", "application/json")
	s.headers.Store("referer", s.referrer)
}
