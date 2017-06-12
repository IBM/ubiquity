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
	fakeHost           = "fakehost"
	fakeError          = "error"
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
			fakeScbeRestClient.CreateVolumeReturns(scbe.ScbeVolumeInfo{"v1", "wwn1", "gold"}, nil)
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
			fakeScbeRestClient.CreateVolumeReturns(scbe.ScbeVolumeInfo{"v1", "wwn1", "gold"}, nil)
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
		fakeConfig = resources.ScbeConfig{
			ConfigPath:     "/tmp",
			DefaultService: fakeDefaultProfile,
			HostnameTmp:    fakeHost} // TODO its workaround to issue #23
		client, err = scbe.NewScbeLocalClientWithNewScbeRestClientAndDataModel(
			logger,
			fakeConfig,
			fakeScbeDataModel,
			fakeScbeRestClient)
		Expect(err).ToNot(HaveOccurred())

		fakeScbeRestClient.LoginReturns(nil)
		fakeScbeRestClient.ServiceExistReturns(true, nil)
		err = client.Activate()
		Expect(err).ToNot(HaveOccurred())
		Expect(fakeScbeRestClient.LoginCallCount()).To(Equal(1))
		Expect(fakeScbeRestClient.ServiceExistCallCount()).To(Equal(1))
	})

	Context(".Attach", func() {
		It("should fail to attach the volume if GetVolume failed", func() {
			fakeScbeDataModel.GetVolumeReturns(
				scbe.ScbeVolume{}, false, fmt.Errorf(fakeError))
			_, err := client.Attach("fakevol")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(fakeError))
		})
		It("should fail to attach the volume if vol not exist in DB", func() {
			fakeScbeDataModel.GetVolumeReturns(
				scbe.ScbeVolume{}, false, nil)
			_, err := client.Attach("fakevol")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(MatchRegexp(scbe.MsgVolumeNotInUbiquityDB))
		})
		It("should fail to attach the volume if vol already attached in DB", func() {
			fakeScbeDataModel.GetVolumeReturns(
				scbe.ScbeVolume{AttachTo: "fakevol1"}, true, nil)
			_, err := client.Attach("fakevol")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(MatchRegexp("Cannot attach volume"))
		})
		It("should fail to attach the volume if MapVolume failed", func() {
			fakeScbeDataModel.GetVolumeReturns(
				scbe.ScbeVolume{AttachTo: ""}, true, nil)
			fakeScbeRestClient.MapVolumeReturns(scbe.ScbeResponseMapping{}, fmt.Errorf(fakeError))
			_, err := client.Attach("fakevol")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(fakeError))
			Expect(fakeScbeRestClient.MapVolumeCallCount()).To(Equal(1))

		})
		It("should fail to attach the volume if update the vol in the DB failed", func() {
			fakeScbeDataModel.GetVolumeReturns(
				scbe.ScbeVolume{AttachTo: ""}, true, nil)
			fakeScbeRestClient.MapVolumeReturns(scbe.ScbeResponseMapping{}, nil)
			fakeScbeDataModel.UpdateVolumeAttachToReturns(fmt.Errorf(fakeError))
			_, err := client.Attach("fakevol")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(fakeError))
			Expect(fakeScbeDataModel.UpdateVolumeAttachToCallCount()).To(Equal(1))
		})
		It("should succeed to attach the volume when everything is cool", func() {
			fakeScbeDataModel.GetVolumeReturns(
				scbe.ScbeVolume{AttachTo: ""}, true, nil)
			fakeScbeRestClient.MapVolumeReturns(scbe.ScbeResponseMapping{}, nil)
			fakeScbeDataModel.UpdateVolumeAttachToReturns(nil)
			_, err := client.Attach("fakevol")
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeScbeDataModel.UpdateVolumeAttachToCallCount()).To(Equal(1))
		})
		It("should succeed to attach the volume if vol already attach to this host", func() {
			fakeScbeDataModel.GetVolumeReturns(
				scbe.ScbeVolume{AttachTo: fakeHost}, true, nil)
			_, err := client.Attach("fakevol")
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeScbeDataModel.UpdateVolumeAttachToCallCount()).To(Equal(0))
		})

	})
	Context(".Detach", func() {
		It("should fail to attach the volume if GetVolume failed", func() {
			fakeScbeDataModel.GetVolumeReturns(
				scbe.ScbeVolume{}, false, fmt.Errorf(fakeError))
			_, err := client.Attach("fakevol")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(fakeError))
		})
		It("should fail to attach the volume if vol not exist in DB", func() {
			fakeScbeDataModel.GetVolumeReturns(
				scbe.ScbeVolume{}, false, nil)
			_, err := client.Attach("fakevol")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(MatchRegexp(scbe.MsgVolumeNotInUbiquityDB))
		})
		It("should fail to detach the volume if vol already detached in DB", func() {
			fakeScbeDataModel.GetVolumeReturns(
				scbe.ScbeVolume{AttachTo: scbe.EmptyHost}, true, nil)
			err := client.Detach("fakevol")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(MatchRegexp("Cannot detach volume"))
		})
		It("should fail to detach the volume if MapVolume failed", func() {
			fakeScbeDataModel.GetVolumeReturns(
				scbe.ScbeVolume{AttachTo: fakeHost}, true, nil)
			fakeScbeRestClient.UnmapVolumeReturns(fmt.Errorf(fakeError))
			err := client.Detach("fakevol")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(fakeError))
			Expect(fakeScbeRestClient.UnmapVolumeCallCount()).To(Equal(1))
		})
		It("should fail to detach the volume if update the vol in the DB failed", func() {
			fakeScbeDataModel.GetVolumeReturns(
				scbe.ScbeVolume{AttachTo: fakeHost}, true, nil)
			fakeScbeRestClient.UnmapVolumeReturns(nil)
			fakeScbeDataModel.UpdateVolumeAttachToReturns(fmt.Errorf(fakeError))
			err := client.Detach("fakevol")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(fakeError))
			Expect(fakeScbeDataModel.UpdateVolumeAttachToCallCount()).To(Equal(1))
		})
		It("should succeed to detach the volume when everything is cool", func() {
			fakeScbeDataModel.GetVolumeReturns(
				scbe.ScbeVolume{AttachTo: fakeHost}, true, nil)
			fakeScbeRestClient.UnmapVolumeReturns(nil)
			fakeScbeDataModel.UpdateVolumeAttachToReturns(nil)
			err := client.Detach("fakevol")
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeScbeDataModel.UpdateVolumeAttachToCallCount()).To(Equal(1))
		})

	})

})
