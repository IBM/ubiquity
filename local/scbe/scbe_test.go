package scbe_test

import (
	"fmt"
	"github.com/IBM/ubiquity/fakes"
	"github.com/IBM/ubiquity/local/scbe"
	"github.com/IBM/ubiquity/resources"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"log"
	"os"
)

const (
	fakeDefaultProfile = "defaultProfile"
)

var _ = Describe("scbeLocalClient", func() {
	var (
		client             resources.StorageClient
		logger             *log.Logger
		fakeScbeDataModel  *fakes.FakeScbeDataModel
		fakeScbeRestClient *fakes.FakeScbeRestClient
		fakeConfig         resources.ScbeConfig
		err                error
	)
	BeforeEach(func() {
		logger = log.New(os.Stdout, "ubiquity scbe: ", log.Lshortfile|log.LstdFlags)
		fakeScbeDataModel = new(fakes.FakeScbeDataModel)
		fakeScbeRestClient = new(fakes.FakeScbeRestClient)
		fakeConfig = resources.ScbeConfig{ConfigPath: "/tmp", DefaultService: fakeDefaultProfile} // TODO add more details
		client, err = scbe.NewScbeLocalClientWithNewScbeRestClientAndDataModel(
			logger,
			fakeConfig,
			fakeScbeDataModel,
			fakeScbeRestClient)
		Expect(err).ToNot(HaveOccurred())

	})

	Context(".Activate", func() {
		It("should fail login to SCBE during activation", func() {
			fakeScbeRestClient.LoginReturns(fmt.Errorf("Fail to SCBE login during activation"))
			err = client.Activate()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Error in login remote call"))
			Expect(fakeScbeRestClient.LoginCallCount()).To(Equal(1))
			Expect(fakeScbeRestClient.ServiceExistCallCount()).To(Equal(0))

		})

		It("should fail when service exist fail", func() {
			fakeScbeRestClient.LoginReturns(nil)
			fakeScbeRestClient.ServiceExistReturns(false, fmt.Errorf("Fail to run service exist"))
			err = client.Activate()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(MatchRegexp("^Error in activate SCBE backend while checking default service"))
			Expect(fakeScbeRestClient.LoginCallCount()).To(Equal(1))
			Expect(fakeScbeRestClient.ServiceExistCallCount()).To(Equal(1))

		})
		It("should fail when service does NOT exist", func() {
			fakeScbeRestClient.LoginReturns(nil)
			fakeScbeRestClient.ServiceExistReturns(false, nil)
			err = client.Activate()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(MatchRegexp("^Error in activate .* does not exist in SCBE"))
			Expect(fakeScbeRestClient.LoginCallCount()).To(Equal(1))
			Expect(fakeScbeRestClient.ServiceExistCallCount()).To(Equal(1))

		})
		It("should succeed when ServiceExist returns true", func() {
			fakeScbeRestClient.LoginReturns(nil)
			fakeScbeRestClient.ServiceExistReturns(true, nil)
			err = client.Activate()
			Expect(err).ToNot(HaveOccurred())
			Expect(fakeScbeRestClient.LoginCallCount()).To(Equal(1))
			Expect(fakeScbeRestClient.ServiceExistCallCount()).To(Equal(1))

		})

	})
	Context(".CreateVolume", func() {
		It("should fail create volume if error to get vol from DB", func() {
			fakeError := "error"
			fakeScbeDataModel.GetVolumeReturns(scbe.ScbeVolume{}, false, fmt.Errorf(fakeError))
			err = client.CreateVolume("fakevol", nil)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(fakeError))
		})
		It("should fail create volume if volume already exist in DB", func() {
			fakeScbeDataModel.GetVolumeReturns(scbe.ScbeVolume{}, true, nil)
			err = client.CreateVolume("fakevol", nil)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(fmt.Sprintf(scbe.MsgVolumeAlreadyExistInDB, "fakevol")))
		})
		It("should fail create volume if vol sise not provided in opts", func() {
			fakeScbeDataModel.GetVolumeReturns(scbe.ScbeVolume{}, false, nil)
			err = client.CreateVolume("fakevol", nil)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(scbe.MsgOptionSizeIsMissing))
		})

		It("should fail create volume if vol size is not number", func() {
			fakeScbeDataModel.GetVolumeReturns(scbe.ScbeVolume{}, false, nil)
			opts := make(map[string]interface{})
			opts[scbe.OptionNameForVolumeSize] = "aaa"

			err = client.CreateVolume("fakevol", opts)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(fmt.Sprintf(scbe.MsgOptionMustBeNumber, opts["size"])))
		})
		It("should fail create volume if vol creation failed with err", func() {
			fakeScbeDataModel.GetVolumeReturns(scbe.ScbeVolume{}, false, nil)
			fakeScbeRestClient.CreateVolumeReturns(scbe.ScbeVolumeInfo{}, fmt.Errorf("error"))
			opts := make(map[string]interface{})
			opts[scbe.OptionNameForVolumeSize] = "100"

			volFake := "fakevol"
			err = client.CreateVolume(volFake, opts)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("error"))
			Expect(fakeScbeRestClient.CreateVolumeCallCount()).To(Equal(1))
			volname, profile, size := fakeScbeRestClient.CreateVolumeArgsForCall(0)
			Expect(profile).To(Equal(fakeDefaultProfile))
			Expect(size).To(Equal(100))
			Expect(volname).To(Equal(fmt.Sprintf(scbe.ComposeVolumeName, scbe.DefaultUbiquityInstanceName, volFake)))
		})
		It("should fail create volume if vol creation failed with err", func() {
			fakeScbeDataModel.GetVolumeReturns(scbe.ScbeVolume{}, false, nil)
			fakeScbeRestClient.CreateVolumeReturns(scbe.ScbeVolumeInfo{}, fmt.Errorf("error"))
			opts := make(map[string]interface{})
			opts[scbe.OptionNameForVolumeSize] = "100"
			opts[scbe.OptionNameForServiceName] = "gold"

			volFake := "fakevol"
			err = client.CreateVolume(volFake, opts)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("error"))
			Expect(fakeScbeRestClient.CreateVolumeCallCount()).To(Equal(1))
			volname, profile, size := fakeScbeRestClient.CreateVolumeArgsForCall(0)
			Expect(profile).To(Equal("gold"))
			Expect(size).To(Equal(100))
			Expect(volname).To(Equal(fmt.Sprintf(scbe.ComposeVolumeName, scbe.DefaultUbiquityInstanceName, volFake)))
		})

		It("should fail to insert vol to DB after create it", func() {
			fakeScbeDataModel.GetVolumeReturns(scbe.ScbeVolume{}, false, nil)
			fakeScbeRestClient.CreateVolumeReturns(scbe.ScbeVolumeInfo{"v1", "wwn1"}, nil)
			fakeScbeDataModel.InsertVolumeReturns(fmt.Errorf("error"))
			opts := make(map[string]interface{})
			opts[scbe.OptionNameForVolumeSize] = "100"
			opts[scbe.OptionNameForServiceName] = "gold"

			volFake := "fakevol"
			err = client.CreateVolume(volFake, opts)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("error"))
			Expect(fakeScbeDataModel.InsertVolumeCallCount()).To(Equal(1))
			name, wwn, profile, host := fakeScbeDataModel.InsertVolumeArgsForCall(0)
			Expect(name).To(Equal(volFake))
			Expect(wwn).To(Equal("wwn1"))
			Expect(profile).To(Equal("gold"))
			Expect(host).To(Equal(scbe.AttachedToNothing))

		})

		It("should succeed to insert vol to DB after create it", func() {
			fakeScbeDataModel.GetVolumeReturns(scbe.ScbeVolume{}, false, nil)
			fakeScbeRestClient.CreateVolumeReturns(scbe.ScbeVolumeInfo{"v1", "wwn1"}, nil)
			fakeScbeDataModel.InsertVolumeReturns(nil)
			opts := make(map[string]interface{})
			opts[scbe.OptionNameForVolumeSize] = "100"
			opts[scbe.OptionNameForServiceName] = "gold"

			volFake := "fakevol"
			err = client.CreateVolume(volFake, opts)
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeScbeDataModel.InsertVolumeCallCount()).To(Equal(1))
			name, wwn, profile, host := fakeScbeDataModel.InsertVolumeArgsForCall(0)
			Expect(name).To(Equal(volFake))
			Expect(wwn).To(Equal("wwn1"))
			Expect(profile).To(Equal("gold"))
			Expect(host).To(Equal(scbe.AttachedToNothing))

		})

	})
})
