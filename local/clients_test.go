package local_test

import (
	"github.com/IBM/ubiquity/utils/logs"
	"github.com/IBM/ubiquity/resources"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/IBM/ubiquity/local"
	"fmt"
)

var _ = Describe("Clients", func() {
	var (
		fakeConfig          resources.UbiquityServerConfig
		fakeScbeConfig      resources.ScbeConfig
		fakeConnectionInfo  resources.ConnectionInfo
		fakeRestConfig	    resources.RestConfig
        fakeSpectrumScaleConfig	resources.SpectrumScaleConfig
		err                 error
		logger              logs.Logger
		client		    map[string]resources.StorageClient

	)
	BeforeEach(func() {
		logger = logs.GetLogger()
	})

	Context(".GetLocalClients", func() {
	It("Should Pass when ManagementIP present for SCBE backend, not present for SpectrumScale backend and initialization for SCBE backend is successful", func() {
		fakeConnectionInfo = resources.ConnectionInfo{}
		fakeScbeConfig	   = resources.ScbeConfig{ConnectionInfo: fakeConnectionInfo,}
		fakeScbeConfig.ConnectionInfo.ManagementIP="1.1.1.1"

		fakeRestConfig = resources.RestConfig{}
		fakeSpectrumScaleConfig = resources.SpectrumScaleConfig{RestConfig: fakeRestConfig,}
		fakeSpectrumScaleConfig.RestConfig.ManagementIP = ""

		fakeConfig = resources.UbiquityServerConfig{ScbeConfig: fakeScbeConfig, SpectrumScaleConfig: fakeSpectrumScaleConfig}


		oldNewScbeLocalClient := local.NewScbeLocalClient

		defer func () { local.NewScbeLocalClient = oldNewScbeLocalClient }()

		local.NewScbeLocalClient = func (scbeConfig resources.ScbeConfig) (resources.StorageClient, error) {
			return  nil, nil
		}

		client, err = local.GetLocalClients(logger, fakeConfig)
		Expect(err).ToNot(HaveOccurred())
	})

	It("Should Fail when ManagementIP present for SCBE backend, not present for SpectrumScale backend and SCBE backend intialization fails", func() {
		fakeConnectionInfo = resources.ConnectionInfo{}
		fakeScbeConfig	   = resources.ScbeConfig{ConnectionInfo: fakeConnectionInfo,}
		fakeScbeConfig.ConnectionInfo.ManagementIP="1.1.1.1"

		fakeRestConfig = resources.RestConfig{}
		fakeSpectrumScaleConfig = resources.SpectrumScaleConfig{RestConfig: fakeRestConfig,}
		fakeSpectrumScaleConfig.RestConfig.ManagementIP = ""

		fakeConfig = resources.UbiquityServerConfig{ScbeConfig: fakeScbeConfig, SpectrumScaleConfig: fakeSpectrumScaleConfig}


		oldNewScbeLocalClient := local.NewScbeLocalClient

		defer func () { local.NewScbeLocalClient = oldNewScbeLocalClient }()

		local.NewScbeLocalClient = func (scbeConfig resources.ScbeConfig) (resources.StorageClient, error) {
			return  nil, fmt.Errorf("SCBE Initialization failed")
		}

		client, err = local.GetLocalClients(logger, fakeConfig)

		Expect(err.Error()).To(Equal("Error while initializing scbe client:[SCBE Initialization failed]"))
	})

	It("Should Pass when ManagementIP present for SpectrumScale backend, not present for SCBE backend and SpectrumScale initialization is successful", func() {
		fakeConnectionInfo = resources.ConnectionInfo{}
		fakeScbeConfig	   = resources.ScbeConfig{ConnectionInfo: fakeConnectionInfo,}
		fakeScbeConfig.ConnectionInfo.ManagementIP=""

		fakeRestConfig = resources.RestConfig{}
		fakeSpectrumScaleConfig = resources.SpectrumScaleConfig{RestConfig: fakeRestConfig,}
		fakeSpectrumScaleConfig.RestConfig.ManagementIP = "1.1.1.1"

		fakeConfig = resources.UbiquityServerConfig{ScbeConfig: fakeScbeConfig, SpectrumScaleConfig: fakeSpectrumScaleConfig}


		oldNewSpectrumScaleLocalClient := local.NewSpectrumLocalClient

		defer func () { local.NewSpectrumLocalClient = oldNewSpectrumScaleLocalClient }()

		local.NewSpectrumLocalClient = func (Config resources.UbiquityServerConfig) (resources.StorageClient, error) {
			return  nil, nil
		}

		client, err = local.GetLocalClients(logger, fakeConfig)
		Expect(err).ToNot(HaveOccurred())
	})

	It("Should Fail when ManagementIP present for SpectrumScale backend, not present for SCBE backend and SpectrumScale Backend Initialization fails", func() {
		fakeConnectionInfo = resources.ConnectionInfo{}
		fakeScbeConfig	   = resources.ScbeConfig{ConnectionInfo: fakeConnectionInfo,}
		fakeScbeConfig.ConnectionInfo.ManagementIP=""

		fakeRestConfig = resources.RestConfig{}
		fakeSpectrumScaleConfig = resources.SpectrumScaleConfig{RestConfig: fakeRestConfig,}
		fakeSpectrumScaleConfig.RestConfig.ManagementIP = "1.1.1.1"

		fakeConfig = resources.UbiquityServerConfig{ScbeConfig: fakeScbeConfig, SpectrumScaleConfig: fakeSpectrumScaleConfig}


		oldNewSpectrumScaleLocalClient := local.NewSpectrumLocalClient

		defer func () { local.NewSpectrumLocalClient = oldNewSpectrumScaleLocalClient }()

		local.NewSpectrumLocalClient = func (Config resources.UbiquityServerConfig) (resources.StorageClient, error) {
			return  nil, fmt.Errorf("SpectrumScale Initialization failed") 
		}

		client, err = local.GetLocalClients(logger, fakeConfig)

		Expect(err.Error()).To(Equal("Error while initializing spectrum-scale client:[SpectrumScale Initialization failed]"))
	})

	It("Should Pass when ManagementIP present for both backends and Initialization is successfull for both backends", func() {
		fakeConnectionInfo = resources.ConnectionInfo{}
		fakeScbeConfig	   = resources.ScbeConfig{ConnectionInfo: fakeConnectionInfo,}
		fakeScbeConfig.ConnectionInfo.ManagementIP=""

		fakeRestConfig = resources.RestConfig{}
		fakeSpectrumScaleConfig = resources.SpectrumScaleConfig{RestConfig: fakeRestConfig,}
		fakeSpectrumScaleConfig.RestConfig.ManagementIP = "1.1.1.1"

		fakeConfig = resources.UbiquityServerConfig{ScbeConfig: fakeScbeConfig, SpectrumScaleConfig: fakeSpectrumScaleConfig}


		oldNewSpectrumScaleLocalClient := local.NewSpectrumLocalClient

		defer func () { local.NewSpectrumLocalClient = oldNewSpectrumScaleLocalClient }()

		local.NewSpectrumLocalClient = func (Config resources.UbiquityServerConfig) (resources.StorageClient, error) {
			return  nil, nil
		}

		oldNewScbeLocalClient := local.NewScbeLocalClient

		defer func () { local.NewScbeLocalClient = oldNewScbeLocalClient }()

		local.NewScbeLocalClient = func (scbeConfig resources.ScbeConfig) (resources.StorageClient, error) {
			return  nil, nil
		}

		client, err = local.GetLocalClients(logger, fakeConfig)
		Expect(err).ToNot(HaveOccurred())
	})


	It("Should fail when ManagementIP is empty for both backend", func() {
		fakeConnectionInfo = resources.ConnectionInfo{}
		fakeScbeConfig	   = resources.ScbeConfig{ConnectionInfo: fakeConnectionInfo,}
		fakeRestConfig = resources.RestConfig{}
		fakeSpectrumScaleConfig = resources.SpectrumScaleConfig{RestConfig: fakeRestConfig,}
		fakeConfig	   = resources.UbiquityServerConfig{ScbeConfig: fakeScbeConfig,SpectrumScaleConfig: fakeSpectrumScaleConfig,}
		client, err = local.GetLocalClients(logger, fakeConfig)
		Expect(client).To(BeNil())
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal(resources.ClientInitializationErrorStr))
	})
	})
})
