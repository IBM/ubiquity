package scbe_test

import (
	"github.com/IBM/ubiquity/local/scbe"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega" // including the whole package inside the file
	"log"
	"os"
	// httpmock is the referrer for this module
	"gopkg.in/jarcoal/httpmock.v1"
)

var _ = Describe("restClient integration testing with existing SCBE instance", func() {
	var (
		logger         *log.Logger
		conInfo        scbe.ConnectionInfo
		client         scbe.RestClient
		err            error
		credentialInfo scbe.CredentialInfo
	)
	BeforeEach(func() {
		logger = log.New(os.Stdout, "ubiquity scbe: ", log.Lshortfile|log.LstdFlags)
		// Get environment variable for the tests
		scbeUser := os.Getenv("SCBE_USER")
		scbePassword := os.Getenv("SCBE_PASSWORD")
		scbeIP := os.Getenv("SCBE_IP")
		scbePort := os.Getenv("SCBE_PORT")
		skipMsg := "environment is empty, skip the integration test."
		if scbeUser == "" {
			Skip("SCBE_USER " + skipMsg)
		}
		if scbePassword == "" {
			Skip("SCBE_PASSWORD " + skipMsg)
		}
		if scbeIP == "" {
			Skip("SCBE_IP " + skipMsg)
		}
		if scbePort == "" {
			Skip("SCBE_PORT " + skipMsg)
		}
		credentialInfo = scbe.CredentialInfo{scbeUser, scbePassword, "flocker"}
		conInfo = scbe.ConnectionInfo{credentialInfo, scbePort, scbeIP, false}
		client, err = scbe.NewRestClient(
			logger,
			conInfo,
			"https://"+scbeIP+":"+scbePort+"/api/v1",
			"users/get-auth-token",
			"https://"+scbeIP+":"+scbePort+"/")
		Expect(err).ToNot(HaveOccurred())
		httpmock.DeactivateAndReset()
	})

	Context(".Login", func() {
		It("Should succeed to login to SCBE", func() {
			err := client.Login()
			Expect(err).ToNot(HaveOccurred())
		})
	})

})
