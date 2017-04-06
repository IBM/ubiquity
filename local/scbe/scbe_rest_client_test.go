package scbe_test

import (
	"log"
	"net/http"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega" // including the whole package inside the file

	"github.com/IBM/ubiquity/local/scbe"
	httpmock "gopkg.in/jarcoal/httpmock.v1" // httpmock is the referrer for this module
	"k8s.io/kubernetes/staging/src/k8s.io/client-go/_vendor/k8s.io/apimachinery/pkg/util/json"
)

var _ = Describe("scbeRestClient", func() {
	var (
		logger  *log.Logger
		conInfo scbe.ConnectionInfo
		client  scbe.ScbeRestClient
		err     error
	)
	BeforeEach(func() {
		logger = log.New(os.Stdout, "ubiquity scbe: ", log.Lshortfile|log.LstdFlags)
		conInfo = scbe.ConnectionInfo{}
		client, err = scbe.NewScbeRestClient(logger, conInfo, "http://scbe.fake.com", "users/get-auth-token", "https://{scbe_ip}:{scbe_port}/")
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
