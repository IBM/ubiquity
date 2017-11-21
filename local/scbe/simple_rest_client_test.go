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

package scbe_test

import (
	"encoding/json"
	"github.com/IBM/ubiquity/local/scbe"
	"github.com/IBM/ubiquity/resources"
	"github.com/jarcoal/httpmock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega" // including the whole package inside the file
	"io/ioutil"
	"net/http"
	"os"
)

const (
	fakeScbeQfdn        = "1.1.1.1"
	fakeScbeUrlBase     = "https://" + fakeScbeQfdn + ":6666"
	suffix              = "api/v1"
	fakeScbeUrlAuth     = "users/get-auth-token"
	fakeScbeUrlAuthFull = fakeScbeUrlBase + "/" + suffix + "/" + fakeScbeUrlAuth
	fakeScbeUrlReferer  = fakeScbeUrlBase + "/"
	fakeScbeUrlApi      = fakeScbeUrlBase + "/" + suffix
	fakeCert            = "/tmp/fake_cert.crt"
)

var fakeServiceJsonResponse string = `
[
{
"id": "cc4c1254-d551-4a51-81f5-ffffffffffff",
"unique_identifier": "cc4c1254-d551-4a51-81f5-ffffffffffff",
"name": "gold",
"description": " ",
"container": "23c380fc-fe1e-4c02-9d1e-ffffffffffff",
"capability_values": "",
"type": "regular",
"physical_size": 413457711104,
"logical_size": 413457711104,
"physical_free": 310093283328,
"logical_free": 310093283328,
"total_capacity": 413457711104,
"used_capacity": 103364427776,
"max_resource_logical_free": 310093283328,
"max_resource_free_size_for_provisioning": 310093283328,
"num_volumes": 0,
"has_admin": true,
"qos_max_iops": 0,
"qos_max_mbps": 0
}
]`

var _ = Describe("restClient", func() {
	var (
		client scbe.SimpleRestClient
		err    error
	)
	BeforeEach(func() {
		client, err = scbe.NewSimpleRestClient(resources.ConnectionInfo{}, fakeScbeUrlBase+"/"+suffix, fakeScbeUrlAuth, fakeScbeUrlReferer)
		Expect(err).ToNot(HaveOccurred())
	})

	Context(".Login", func() {
		It("should succeed when httpClient POST succeeds and returns a token", func() {
			loginResponse := scbe.LoginResponse{Token: "fake-token"}
			marshalledResponse, err := json.Marshal(loginResponse)
			Expect(err).ToNot(HaveOccurred())
			httpmock.RegisterResponder(
				"POST",
				fakeScbeUrlAuthFull,
				httpmock.NewStringResponder(http.StatusOK, string(marshalledResponse)),
			)
			err = client.Login()
			Expect(err).ToNot(HaveOccurred())
		})
		It("should fail when httpClient.Post returns an empty token", func() {
			loginResponse := scbe.LoginResponse{Token: ""}
			marshalledResponse, err := json.Marshal(loginResponse)
			Expect(err).ToNot(HaveOccurred())
			httpmock.RegisterResponder("POST", fakeScbeUrlAuthFull, httpmock.NewStringResponder(http.StatusOK, string(marshalledResponse)))
			err = client.Login()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Token is empty"))
		})
		It("should fail when httpClient.Post returns error status", func() {
			httpmock.RegisterResponder("POST", fakeScbeUrlAuthFull, httpmock.NewStringResponder(http.StatusBadRequest, "{}"))
			err = client.Login()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(MatchRegexp("^bad status code"))
		})
		It("should fail when httpClient.Post returns invalid json", func() {
			httpmock.RegisterResponder("POST", fakeScbeUrlAuthFull, httpmock.NewStringResponder(http.StatusOK, "yyy"))
			err = client.Login()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(MatchRegexp("^invalid character"))
		})

	})
})

var _ = Describe("restClient", func() {
	var (
		client scbe.SimpleRestClient
		err    error
	)
	BeforeEach(func() {
		client, err = scbe.NewSimpleRestClient(resources.ConnectionInfo{}, fakeScbeUrlBase+"/"+suffix, fakeScbeUrlAuth, fakeScbeUrlReferer)
		Expect(err).ToNot(HaveOccurred())
		loginResponse := scbe.LoginResponse{Token: "fake-token"}
		marshalledResponse, err := json.Marshal(loginResponse)
		Expect(err).ToNot(HaveOccurred())
		httpmock.RegisterResponder(
			"POST",
			fakeScbeUrlAuthFull,
			httpmock.NewStringResponder(http.StatusOK, string(marshalledResponse)),
		)
		err = client.Login()
		Expect(err).ToNot(HaveOccurred())

	})

	Context(".Get", func() {
		It("should succeed when httpClient.Get succeeds and returns valid data", func() {
			httpmock.RegisterResponder(
				"GET",
				fakeScbeUrlApi+"/"+scbe.UrlScbeResourceService,
				httpmock.NewStringResponder(http.StatusOK, fakeServiceJsonResponse),
			)
			var services []scbe.ScbeStorageService
			err = client.Get(scbe.UrlScbeResourceService, nil, -1, &services)
			Expect(err).ToNot(HaveOccurred())
			Expect(services[0].Name).To(Equal("gold"))
		})
		It("should fail when httpClient.Get returns error status", func() {
			httpmock.RegisterResponder(
				"GET",
				fakeScbeUrlApi+"/"+scbe.UrlScbeResourceService,
				httpmock.NewStringResponder(http.StatusBadRequest, fakeServiceJsonResponse),
			)
			var services []scbe.ScbeStorageService
			err = client.Get(scbe.UrlScbeResourceService, nil, -1, &services)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(MatchRegexp("^bad status code"))
		})
		It("should fail when httpClient.Get returns invalid json", func() {
			httpmock.RegisterResponder(
				"GET",
				fakeScbeUrlApi+"/"+scbe.UrlScbeResourceService,
				httpmock.NewStringResponder(http.StatusOK, "invalid data"),
			)
			var services []scbe.ScbeStorageService
			err = client.Get(scbe.UrlScbeResourceService, nil, -1, &services)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(MatchRegexp("^invalid character"))
		})
		It("Login and retry rest call if token expired", func() {
			var numLogin, numGetServices int
			fakeTokenRetry := "fake-token-retry"
			httpmock.RegisterResponder("POST", fakeScbeUrlAuthFull, CountLoginResponder(&numLogin, fakeTokenRetry))
			httpmock.RegisterResponder("GET", fakeScbeUrlApi+"/"+scbe.UrlScbeResourceService, TokenExpiredResponder(&numGetServices, fakeTokenRetry))
			var services []scbe.ScbeStorageService
			err = client.Get(scbe.UrlScbeResourceService, nil, http.StatusOK, &services)
			Expect(err).ToNot(HaveOccurred())
			Expect(numLogin).To(Equal(1))
			Expect(numGetServices).To(Equal(2))
		})
	})
})

func CountLoginResponder(num *int, retryToken string) httpmock.Responder {
	*num = 0
	loginResponse := scbe.LoginResponse{Token: retryToken}
	marshResponse, err := json.Marshal(loginResponse)
	Expect(err).ToNot(HaveOccurred())
	return func(req *http.Request) (*http.Response, error) {
		*num++
		return httpmock.NewStringResponse(http.StatusOK, string(marshResponse)), nil
	}
}

func TokenExpiredResponder(num *int, retryToken string) httpmock.Responder {
	*num = 0
	return func(req *http.Request) (*http.Response, error) {
		*num++
		auth := req.Header[scbe.HTTP_AUTH_KEY]
		if len(auth) == 1 && auth[0] == "Token "+retryToken {
			return httpmock.NewStringResponse(http.StatusOK, fakeServiceJsonResponse), nil
		} else {
			return httpmock.NewStringResponse(http.StatusUnauthorized, ""), nil
		}
	}
}

var _ = Describe("restClient", func() {
	var (
		client scbe.SimpleRestClient
		err    error
	)
	BeforeEach(func() {
	})
	Context(".initTransport", func() {
		It("should fail with default (verify-full) and no certificates", func() {
			client, err = scbe.NewSimpleRestClient(resources.ConnectionInfo{ManagementIP: fakeScbeQfdn}, fakeScbeUrlBase+"/"+suffix, fakeScbeUrlAuth, fakeScbeUrlReferer)
			Expect(err).To(HaveOccurred())
			_, ok := err.(*scbe.SslModeFullVerifyWithoutCAfile)
			Expect(ok).To(Equal(true))
		})
		It("should fail if wrong ssl mode", func() {
			os.Setenv(resources.KeyScbeSslMode, "fake ssl mode")
			client, err = scbe.NewSimpleRestClient(resources.ConnectionInfo{ManagementIP: fakeScbeQfdn}, fakeScbeUrlBase+"/"+suffix, fakeScbeUrlAuth, fakeScbeUrlReferer)
			os.Unsetenv(resources.KeyScbeSslMode)
			Expect(err).To(HaveOccurred())
			_, ok := err.(*scbe.SslModeValueInvalid)
			Expect(ok).To(Equal(true))
		})
		It("fails if no certificate", func() {
			os.Remove(fakeCert)
			os.Setenv(scbe.KEY_VERIFY_SCBE_CERT, fakeCert)
			os.Setenv(resources.KeyScbeSslMode, resources.SslModeVerifyFull)
			client, err = scbe.NewSimpleRestClient(resources.ConnectionInfo{ManagementIP: fakeScbeQfdn}, fakeScbeUrlBase+"/"+suffix, fakeScbeUrlAuth, fakeScbeUrlReferer)
			os.Unsetenv(resources.KeyScbeSslMode)
			os.Setenv(scbe.KEY_VERIFY_SCBE_CERT, "")
			Expect(err).To(HaveOccurred())
		})
		It("fails if certificate parse error", func() {
			os.Remove(fakeCert)
			err := ioutil.WriteFile(fakeCert, []byte("fake\n"), 0644)
			Expect(err).ToNot(HaveOccurred())
			os.Setenv(scbe.KEY_VERIFY_SCBE_CERT, fakeCert)
			os.Setenv(resources.KeyScbeSslMode, resources.SslModeVerifyFull)
			client, err = scbe.NewSimpleRestClient(resources.ConnectionInfo{ManagementIP: fakeScbeQfdn}, fakeScbeUrlBase+"/"+suffix, fakeScbeUrlAuth, fakeScbeUrlReferer)
			os.Unsetenv(resources.KeyScbeSslMode)
			os.Setenv(scbe.KEY_VERIFY_SCBE_CERT, "")
			Expect(err).To(HaveOccurred())
		})
		It("should succeed with require ssl mode", func() {
			os.Setenv(resources.KeyScbeSslMode, resources.SslModeRequire)
			client, err = scbe.NewSimpleRestClient(resources.ConnectionInfo{ManagementIP: fakeScbeQfdn}, fakeScbeUrlBase+"/"+suffix, fakeScbeUrlAuth, fakeScbeUrlReferer)
			os.Unsetenv(resources.KeyScbeSslMode)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
