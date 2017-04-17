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
	"fmt"
	"gopkg.in/jarcoal/httpmock.v1"
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
		client, err = scbe.NewRestClient(logger, conInfo, "http://scbe.fake.com", "users/get-auth-token", "https://{scbe_ip}:{scbe_port}/")
		Expect(err).ToNot(HaveOccurred())
	})

	Context(".Login", func() {
		It("should fail when httpClient returns http.StatusNotAcceptable", func() {
			loginResponse := scbe.LoginResponse{Token: "fake-token"}
			marshalledResponse, err := json.Marshal(loginResponse)
			Expect(err).ToNot(HaveOccurred())
			httpmock.RegisterResponder("POST", "http://scbe.fake.com/users/get-auth-token", httpmock.NewStringResponder(http.StatusAccepted, string(marshalledResponse)))
			token, err := client.Login()
			Expect(err).ToNot(HaveOccurred())
			Expect(token).To(Equal("fake-token"))

		})
	})

})

var _ = Describe("restClient integration", func() {
	var (
		logger         *log.Logger
		conInfo        scbe.ConnectionInfo
		client         scbe.RestClient
		err            error
		credentialInfo scbe.CredentialInfo
	)
	BeforeEach(func() {
		logger = log.New(os.Stdout, "ubiquity scbe: ", log.Lshortfile|log.LstdFlags)
		//conInfo = scbe.ConnectionInfo{}
		credentialInfo = scbe.CredentialInfo{"user", "password", "flocker"}
		conInfo = scbe.ConnectionInfo{credentialInfo, "8440", "9.151.162.17", false}
		//client, err = scbe.NewRestClient(logger, conInfo, "http://scbe.fake.com", "users/get-auth-token", "https://{scbe_ip}:{scbe_port}/")
		client, err = scbe.NewRestClient(logger, conInfo, "https://9.151.162.17:8440/api/v1", "users/get-auth-token", "https://9.151.162.17:8440/")
		Expect(err).ToNot(HaveOccurred())
	})

	Context(".Login", func() {
		It("Integration testing", func() {
			token, err := client.Login()
			fmt.Println("response %#s", err)
			Expect(err).ToNot(HaveOccurred())
			Expect(token).To(Equal("real-token"))

		})
	})

})
