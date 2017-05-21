package scbe_test

import (
	"log"
	"os"


	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/IBM/ubiquity/fakes"
	"github.com/IBM/ubiquity/resources"

	"github.com/IBM/ubiquity/local/scbe"
)

var _ = Describe("scbeLocalClient", func() {
	var (
		client            resources.StorageClient
		logger            *log.Logger
		fakeScbeDataModel *fakes.FakeScbeDataModel
		fakeConfig        resources.ScbeConfig
		err               error
	)
	BeforeEach(func() {
		logger = log.New(os.Stdout, "ubiquity scbe: ", log.Lshortfile|log.LstdFlags)
		fakeScbeDataModel = new(fakes.FakeScbeDataModel)
		fakeConfig = resources.ScbeConfig{ConfigPath: "/tmp"} // TODO add more details
		client, err = scbe.NewScbeLocalClientWithNewScbeRestClientAndDataModel(
			logger,
			fakeConfig,
			fakeScbeDataModel,
			nil) // TODO send mock client
		Expect(err).ToNot(HaveOccurred())

	})

	Context(".Activate", func() {
		
		})
})
