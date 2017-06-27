package scbe_test

import (
	"encoding/json"
	"github.com/IBM/ubiquity/local/scbe"
	"github.com/IBM/ubiquity/resources"
	"github.com/jarcoal/httpmock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega" // including the whole package inside the file
	"net/http"
)

const (
	fakeScbeQfdn        = "1.1.1.1"
	fakeScbeUrlBase     = "https://" + fakeScbeQfdn + ":6666"
	suffix              = "api/v1"
	fakeScbeUrlAuth     = "users/get-auth-token"
	fakeScbeUrlAuthFull = fakeScbeUrlBase + "/" + suffix + "/" + fakeScbeUrlAuth
	fakeScbeUrlReferer  = fakeScbeUrlBase + "/"
	fakeScbeUrlApi      = fakeScbeUrlBase + "/" + suffix
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
		client = scbe.NewSimpleRestClient(resources.ConnectionInfo{}, fakeScbeUrlBase+"/"+suffix, fakeScbeUrlAuth, fakeScbeUrlReferer)
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
		client = scbe.NewSimpleRestClient(resources.ConnectionInfo{}, fakeScbeUrlBase+"/"+suffix, fakeScbeUrlAuth, fakeScbeUrlReferer)
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
		if len(auth) == 1 && auth[0] == "Token " + retryToken {
			return httpmock.NewStringResponse(http.StatusOK, fakeServiceJsonResponse), nil
		} else {
			return httpmock.NewStringResponse(http.StatusUnauthorized, ""), nil
		}
	}
}
