package scbe_test

import (
	"encoding/json"
	"log"
	"net/http"

	"os"

	"github.com/IBM/ubiquity/local/scbe"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega" // including the whole package inside the file
	// httpmock is the referrer for this module
	"gopkg.in/jarcoal/httpmock.v1"
)

const (
	fakeScbeQfdn        = "scbe.fake.com"
	fakeScbeUrlBase     = "https://" + fakeScbeQfdn + ":666"
	fakeScbeUrlAuth     = "users/get-auth-token"
	fakeScbeUrlAuthFull = fakeScbeUrlBase + "/" + fakeScbeUrlAuth
	fakeScbeUrlReferer  = fakeScbeUrlBase + "/"
)

var _ = Describe("restClient", func() {
	var (
		logger  *log.Logger
		conInfo scbe.ConnectionInfo
		client  scbe.RestClient
		err     error
	)
	BeforeEach(func() {
		logger = log.New(os.Stdout, "ubiquity scbe: ", log.Lshortfile|log.LstdFlags)
		conInfo = scbe.ConnectionInfo{}
		client, err = scbe.NewRestClient(logger, conInfo, fakeScbeUrlBase, fakeScbeUrlAuth, fakeScbeUrlReferer)
		Expect(err).ToNot(HaveOccurred())
	})

	Context(".Login", func() {
		It("should succeed when httpClient succeed and return a token", func() {
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
		It("should fail when httpClient succeed and return an empty token", func() {
			loginResponse := scbe.LoginResponse{Token: ""}
			marshalledResponse, err := json.Marshal(loginResponse)
			Expect(err).ToNot(HaveOccurred())
			httpmock.RegisterResponder("POST", fakeScbeUrlAuthFull, httpmock.NewStringResponder(http.StatusOK, string(marshalledResponse)))
			err = client.Login()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Error, token is empty"))
		})
		It("should fail when httpClient fails to login due to bad status of response", func() {
			httpmock.RegisterResponder("POST", fakeScbeUrlAuthFull, httpmock.NewStringResponder(http.StatusBadRequest, "{}"))
			err = client.Login()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Error, bad status code of http response"))
		})
		It("should fail when httpClient.post return bad structure that marshaling cannot work with", func() {
			httpmock.RegisterResponder("POST", fakeScbeUrlAuthFull, httpmock.NewStringResponder(http.StatusOK, "yyy"))
			err = client.Login()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Error in unmarshalling response"))
		})

	})

})
