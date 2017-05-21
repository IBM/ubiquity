package scbe_test

import (
	"crypto/tls"
	"fmt"
	"github.com/IBM/ubiquity/local/scbe"
	"github.com/IBM/ubiquity/resources"
	"os"
	"strconv"
	//"github.com/jarcoal/httpmock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega" // including the whole package inside the file
	"log"
	"net/http"
)

var _ = Describe("restClient integration testing with existing SCBE instance", func() {
	var (
		logger         *log.Logger
		conInfo        resources.ConnectionInfo
		client         scbe.RestClient
		credentialInfo resources.CredentialInfo
	)
	BeforeEach(func() {
		logger = log.New(os.Stdout, "ubiquity scbe: ", log.Lshortfile|log.LstdFlags)
		// Get environment variable for the tests
		scbeUser, scbePassword, scbeIP, scbePort, err := getScbeEnvs()
		if err != nil {
			Skip(err.Error())
		}
		credentialInfo = resources.CredentialInfo{scbeUser, scbePassword, "flocker"}
		conInfo = resources.ConnectionInfo{credentialInfo, scbePort, scbeIP, false}
		transCfg := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // ignore expired SSL certificates TODO to use
		}

		client, err = scbe.NewRestClient(
			logger,
			conInfo,
			"https://"+scbeIP+":"+strconv.Itoa(scbePort)+"/api/v1",
			"users/get-auth-token",
			"https://"+scbeIP+":"+strconv.Itoa(scbePort)+"/", transCfg)
		Expect(err).ToNot(HaveOccurred())
		//httpmock.DeactivateAndReset()
	})

	Context(".Login", func() {
		It("Should succeed to login to SCBE", func() {
			err := client.Login()
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context(".Get", func() {
		It("Succeed if there are services available in SCBE", func() {
			var services []scbe.ScbeStorageService
			err := client.Login()
			Expect(err).ToNot(HaveOccurred())
			err = client.Get("/services", nil, 200, &services)
			fmt.Printf("%#v", services)
			Expect(len(services) > 0).To(Equal(true))
		})
	})

})

func getScbeEnvs() (scbeUser, scbePassword, scbeIP string, scbePort int, err error) {
	scbeUser = os.Getenv("SCBE_USER")
	scbePassword = os.Getenv("SCBE_PASSWORD")
	scbeIP = os.Getenv("SCBE_IP")
	scbePortStr := os.Getenv("SCBE_PORT")
	var missingEnvs string
	if scbeUser == "" {
		missingEnvs = missingEnvs + "SCBE_USER "
	}
	if scbePassword == "" {
		missingEnvs = missingEnvs + "SCBE_PASSWORD "
	}
	if scbeIP == "" {
		missingEnvs = missingEnvs + "SCBE_IP "
	}
	if scbePortStr == "" {
		missingEnvs = missingEnvs + "SCBE_PORT "
		scbePort = 0
	} else {
		scbePort, err = strconv.Atoi(scbePortStr)
		if err != nil {
			err = fmt.Errorf("SCBE_PORT environment must be a number")
			return
		}
	}
	if missingEnvs != "" {
		missingEnvs = missingEnvs + "environments are empty, skip the integration test."
		err = fmt.Errorf(missingEnvs)
	} else {
		//fmt.Printf("Using the following params : %s, %s, %s, %s\n", scbeUser, scbePassword, scbeIP, scbePort)
	}

	return
}

var _ = Describe("ScbeRestClient integration testing with existing SCBE instance", func() {
	var (
		logger         *log.Logger
		conInfo        resources.ConnectionInfo
		scbeRestClient scbe.ScbeRestClient
		credentialInfo resources.CredentialInfo
	)
	BeforeEach(func() {
		logger = log.New(os.Stdout, "ubiquity scbe: ", log.Lshortfile|log.LstdFlags)
		// Get environment variable for the tests
		scbeUser, scbePassword, scbeIP, scbePort, err := getScbeEnvs()
		if err != nil {
			Skip(err.Error())
		}
		credentialInfo = resources.CredentialInfo{scbeUser, scbePassword, "flocker"}
		conInfo = resources.ConnectionInfo{credentialInfo, scbePort, scbeIP, false}
		transCfg := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // ignore expired SSL certificates TODO to use
		}
		scbeRestClient, err = scbe.NewScbeRestClient(logger, conInfo, transCfg)
		Expect(err).ToNot(HaveOccurred())
		//httpmock.DeactivateAndReset()
	})

	Context(".Login", func() {
		It("Should succeed to login to SCBE", func() {

			err := scbeRestClient.Login()
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context(".ServiceExist", func() {
		It("Should succeed if gold service exist in SCBE", func() {
			err := scbeRestClient.Login()
			Expect(err).ToNot(HaveOccurred())
			var exist bool
			exist, err = scbeRestClient.ServiceExist("gold")
			Expect(err).NotTo(HaveOccurred())
			Expect(exist).To(Equal(true))
		})
	})
})
