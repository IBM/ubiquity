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

package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
	"strings"

	"github.com/IBM/ubiquity/resources"
	"github.com/IBM/ubiquity/utils/logs"
	"github.com/gorilla/mux"
)

func ExtractErrorResponse(response *http.Response) error {
	errorResponse := resources.GenericResponse{}
	err := UnmarshalResponse(response, &errorResponse)
	if err != nil {
		return logs.GetLogger().ErrorRet(err, "json.Unmarshal failed")
	}
	return fmt.Errorf(errorResponse.Err)
}

func UnmarshalResponse(r *http.Response, object interface{}) error {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return logs.GetLogger().ErrorRet(err, "ioutil.ReadAll failed")
	}

	err = json.Unmarshal(body, object)
	if err != nil {
		return logs.GetLogger().ErrorRet(err, "json.Unmarshal failed")
	}

	return nil
}

func FormatURL(url string, entries ...string) string {
	base := url
	if !strings.HasSuffix(url, "/") {
		base = fmt.Sprintf("%s/", url)
	}
	suffix := ""
	for _, entry := range entries {
		suffix = path.Join(suffix, entry)
	}
	return fmt.Sprintf("%s%s", base, suffix)
}

func HttpExecuteUserAuth(httpClient *http.Client, requestType string, requestURL string, user string, password string, rawPayload interface{}) (*http.Response, error) {
	logger := logs.GetLogger()
	payload, err := json.MarshalIndent(rawPayload, "", " ")
	if err != nil {
		err = fmt.Errorf("Internal error marshalling params %#v", err)
		return nil, logger.ErrorRet(err, "failed")
	}

	if user == "" {
		return nil, fmt.Errorf("Empty UserName passed")
	}

	request, err := http.NewRequest(requestType, requestURL, bytes.NewBuffer(payload))
	if err != nil {
		err = fmt.Errorf("Error in creating request %#v", err)
		return nil, logger.ErrorRet(err, "failed")
	}

	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("Accept", "application/json")

	request.SetBasicAuth(user, password)
	return httpClient.Do(request)

}

func HttpExecute(httpClient *http.Client, requestType string, requestURL string, rawPayload interface{}, request_context resources.RequestContext) (*http.Response, error) {
	logger := logs.GetLogger()
	payload, err := json.MarshalIndent(rawPayload, "", " ")

	if err != nil {
		err = fmt.Errorf("Internal error marshalling params %#v", err)
		return nil, logger.ErrorRet(err, "failed")
	}

	request, err := http.NewRequest(requestType, requestURL, bytes.NewBuffer(payload))

	if err != nil {
		err = fmt.Errorf("Error in creating request %#v", err)
		return nil, logger.ErrorRet(err, "failed")
	}

	return httpClient.Do(request)
}

func WriteResponse(w http.ResponseWriter, code int, object interface{}) {
	data, err := json.Marshal(object)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(code)
	fmt.Fprintf(w, string(data))
}

func Unmarshal(r *http.Request, object interface{}) error {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(body, object)
	if err != nil {
		return err
	}

	return nil
}

func UnmarshalDataFromRequest(r *http.Request, object interface{}) error {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(body, object)
	if err != nil {
		return err
	}

	return nil
}

func ExtractVarsFromRequest(r *http.Request, varName string) string {
	return mux.Vars(r)[varName]
}

type encoding int

const (
	encodePath encoding = 1 + iota

	encodePathSegment

	encodeHost

	encodeZone

	encodeUserPassword

	encodeQueryComponent

	encodeFragment
)
