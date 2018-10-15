package local_test

import (
	"os"
	"log"
	"github.com/IBM/ubiquity/resources"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/IBM/ubiquity/local"
)

var _ = Describe("Clients", func() {
	var (
		fakeConfig          resources.UbiquityServerConfig
		fakeScbeConfig      resources.ScbeConfig
		fakeConnectionInfo  resources.ConnectionInfo
		err                 error
		logger              *log.Logger
		client		        map[string]resources.StorageClient
	)
	BeforeEach(func() {
		logger = log.New(os.Stdout, "ubiquity: ", log.Lshortfile|log.LstdFlags)
	})

	Context(".GetLocalClients", func() {
	It("should fail when ManagementIP is empty for SCBE backend", func() {
		fakeConnectionInfo = resources.ConnectionInfo{}
		fakeScbeConfig	   = resources.ScbeConfig{ConnectionInfo: fakeConnectionInfo,}
		fakeConfig	   = resources.UbiquityServerConfig{ScbeConfig: fakeScbeConfig,}
		client, err = local.GetLocalClients(logger, fakeConfig)
		Expect(client).To(BeNil())
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("No client can be initialized. Please check ubiquity-configmap parameters"))
	})
	})
})
