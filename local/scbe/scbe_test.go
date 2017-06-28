package scbe_test

import (
	"errors"
	"fmt"
	"github.com/IBM/ubiquity/fakes"
	"github.com/IBM/ubiquity/local/scbe"
	"github.com/IBM/ubiquity/resources"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"reflect"
	"strings"
)

const (
	fakeDefaultProfile = "defaultProfile"
	fakeHost           = "fakehost"
	fakeHost2          = "fakehost2"
	fakeError          = "error"
	fakeVol            = "fakevol"
)

var (
	fakeAttachRequest = resources.AttachRequest{Name: fakeVol, Host: fakeHost}
	fakeDetachRequest = resources.DetachRequest{Name: fakeVol, Host: fakeHost}
	fakeRemoveRequest = resources.RemoveVolumeRequest{Name: fakeVol}
)

var _ = Describe("scbeLocalClient init", func() {
	var (
		client             resources.StorageClient
		fakeScbeDataModel  *fakes.FakeScbeDataModel
		fakeScbeRestClient *fakes.FakeScbeRestClient
		fakeConfig         resources.ScbeConfig
		err                error
	)
	BeforeEach(func() {
		fakeScbeDataModel = new(fakes.FakeScbeDataModel)
		fakeScbeRestClient = new(fakes.FakeScbeRestClient)
	})
	Context(".init", func() {
		It("should fail because DefaultVolumeSize is not int", func() {
			fakeConfig = resources.ScbeConfig{
				DefaultVolumeSize: "badint",
			}
			client, err = scbe.NewScbeLocalClientWithNewScbeRestClientAndDataModel(
				fakeConfig,
				fakeScbeDataModel,
				fakeScbeRestClient)
			Expect(err).To(HaveOccurred())
			_, ok := err.(*scbe.ConfigDefaultSizeNotNumError)
			Expect(ok).To(Equal(true))
		})
		It("should fail because UbiquityInstanceName lenth is too long", func() {
			fakeConfig = resources.ScbeConfig{
				UbiquityInstanceName: "1234567890123456",
				DefaultVolumeSize:    "1",
			}
			client, err = scbe.NewScbeLocalClientWithNewScbeRestClientAndDataModel(
				fakeConfig,
				fakeScbeDataModel,
				fakeScbeRestClient)
			Expect(err).To(HaveOccurred())
			_, ok := err.(*scbe.ConfigScbeUbiquityInstanceNameWrongSize)
			Expect(ok).To(Equal(true))
		})
		It("should succeed to init because config is ok", func() {
			fakeConfig = resources.ScbeConfig{
				UbiquityInstanceName: "123456789012345",
				DefaultVolumeSize:    "1",
			}
			fakeScbeRestClient.LoginReturns(nil)
			fakeScbeRestClient.ServiceExistReturns(true, nil)

			client, err = scbe.NewScbeLocalClientWithNewScbeRestClientAndDataModel(
				fakeConfig,
				fakeScbeDataModel,
				fakeScbeRestClient)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})

var _ = Describe("scbeLocalClient", func() {
	var (
		client             resources.StorageClient
		fakeScbeDataModel  *fakes.FakeScbeDataModel
		fakeScbeRestClient *fakes.FakeScbeRestClient
		fakeConfig         resources.ScbeConfig
		err                error
	)
	BeforeEach(func() {
		fakeScbeDataModel = new(fakes.FakeScbeDataModel)
		fakeScbeRestClient = new(fakes.FakeScbeRestClient)
		fakeConfig = resources.ScbeConfig{
			ConfigPath:           "/tmp",
			DefaultService:       fakeDefaultProfile,
			UbiquityInstanceName: "fakeInstance1",
		}
	})

	Context(".Activate", func() {
		It("should fail login to SCBE during activation", func() {
			fakeScbeRestClient.LoginReturns(fmt.Errorf("Fail to SCBE login during activation"))
			client, err = scbe.NewScbeLocalClientWithNewScbeRestClientAndDataModel(
				fakeConfig,
				fakeScbeDataModel,
				fakeScbeRestClient)
			Expect(err).To(HaveOccurred())
			Expect(fakeScbeRestClient.LoginCallCount()).To(Equal(1))
			Expect(fakeScbeRestClient.ServiceExistCallCount()).To(Equal(0))
		})

		It("should fail when service exist fail", func() {
			fakeScbeRestClient.LoginReturns(nil)
			fakeScbeRestClient.ServiceExistReturns(false, fmt.Errorf("Fail to run service exist"))
			client, err = scbe.NewScbeLocalClientWithNewScbeRestClientAndDataModel(
				fakeConfig,
				fakeScbeDataModel,
				fakeScbeRestClient)
			Expect(err).To(HaveOccurred())
			Expect(fakeScbeRestClient.LoginCallCount()).To(Equal(1))
			Expect(fakeScbeRestClient.ServiceExistCallCount()).To(Equal(1))

		})
		It("should fail when service does NOT exist", func() {
			fakeScbeRestClient.LoginReturns(nil)
			fakeScbeRestClient.ServiceExistReturns(false, nil)
			client, err = scbe.NewScbeLocalClientWithNewScbeRestClientAndDataModel(
				fakeConfig,
				fakeScbeDataModel,
				fakeScbeRestClient)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(MatchRegexp("^Error in activate .* does not exist in SCBE"))
			Expect(fakeScbeRestClient.LoginCallCount()).To(Equal(1))
			Expect(fakeScbeRestClient.ServiceExistCallCount()).To(Equal(1))

		})
		It("should succeed when ServiceExist returns true", func() {
			fakeScbeRestClient.LoginReturns(nil)
			fakeScbeRestClient.ServiceExistReturns(true, nil)
			client, err = scbe.NewScbeLocalClientWithNewScbeRestClientAndDataModel(
				fakeConfig,
				fakeScbeDataModel,
				fakeScbeRestClient)
			Expect(err).ToNot(HaveOccurred())
			Expect(fakeScbeRestClient.LoginCallCount()).To(Equal(1))
			Expect(fakeScbeRestClient.ServiceExistCallCount()).To(Equal(1))

		})
		It("should succeed when ServiceExist returns true", func() {
			fakeScbeRestClient.LoginReturns(nil)
			fakeScbeRestClient.ServiceExistReturns(true, nil)
			client, err = scbe.NewScbeLocalClientWithNewScbeRestClientAndDataModel(
				fakeConfig,
				fakeScbeDataModel,
				fakeScbeRestClient)
			Expect(err).ToNot(HaveOccurred())
			Expect(fakeScbeRestClient.LoginCallCount()).To(Equal(1))
			Expect(fakeScbeRestClient.ServiceExistCallCount()).To(Equal(1))

		})

	})
})

var _ = Describe("scbeLocalClient", func() {
	var (
		client             resources.StorageClient
		fakeScbeDataModel  *fakes.FakeScbeDataModel
		fakeScbeRestClient *fakes.FakeScbeRestClient
		fakeConfig         resources.ScbeConfig
		fakeErr            error = errors.New("fake error")
		err                error
	)
	BeforeEach(func() {
		fakeScbeDataModel = new(fakes.FakeScbeDataModel)
		fakeScbeRestClient = new(fakes.FakeScbeRestClient)
		fakeConfig = resources.ScbeConfig{
			ConfigPath:           "/tmp",
			DefaultService:       fakeDefaultProfile,
			UbiquityInstanceName: "fakeInstance1",
		}
		fakeScbeRestClient.LoginReturns(nil)
		fakeScbeRestClient.ServiceExistReturns(true, nil)
		client, err = scbe.NewScbeLocalClientWithNewScbeRestClientAndDataModel(
			fakeConfig,
			fakeScbeDataModel,
			fakeScbeRestClient)
		Expect(err).ToNot(HaveOccurred())
	})
	Context(".CreateVolume", func() {
		It("should fail create volume if error to get vol from DB", func() {
			fakeScbeDataModel.GetVolumeReturns(scbe.ScbeVolume{}, false, fakeErr)
			req := resources.CreateVolumeRequest{Name: "fakevol", Backend: resources.SCBE, Opts: nil}
			err = client.CreateVolume(req)
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(fakeErr))
		})
		It("should fail create volume if volume already exist in DB", func() {
			fakeScbeDataModel.GetVolumeReturns(scbe.ScbeVolume{}, true, nil)
			req := resources.CreateVolumeRequest{Name: "fakevol", Backend: resources.SCBE, Opts: nil}
			err = client.CreateVolume(req)
			Expect(err).To(HaveOccurred())
		})
		It("should fail create volume if vol sise not provided in opts", func() {
			fakeScbeDataModel.GetVolumeReturns(scbe.ScbeVolume{}, false, nil)
			req := resources.CreateVolumeRequest{Name: "fakevol", Backend: resources.SCBE, Opts: nil}
			err = client.CreateVolume(req)
			Expect(err).To(HaveOccurred())
		})

		It("should fail create volume if vol size is not number", func() {
			fakeScbeDataModel.GetVolumeReturns(scbe.ScbeVolume{}, false, nil)
			opts := make(map[string]interface{})
			opts[scbe.OptionNameForVolumeSize] = "aaa"
			req := resources.CreateVolumeRequest{Name: "fakevol", Backend: resources.SCBE, Opts: opts}
			err = client.CreateVolume(req)
			Expect(err).To(HaveOccurred())
		})
		It("should fail create volume if vol len exeeded", func() {
			fakeScbeDataModel.GetVolumeReturns(scbe.ScbeVolume{}, false, nil)
			opts := make(map[string]interface{})
			opts[scbe.OptionNameForVolumeSize] = "100"
			maxVolNameCapable := scbe.MaxVolumeNameLength - (len(fakeConfig.UbiquityInstanceName) + 3)
			volname := strings.Repeat("x", maxVolNameCapable+1)
			req := resources.CreateVolumeRequest{Name: volname, Backend: resources.SCBE, Opts: opts}
			err = client.CreateVolume(req)
			Expect(err).To(HaveOccurred())
			_, ok := err.(*scbe.VolumeNameExceededMaxLengthError)
			Expect(ok).To(Equal(true))

		})
		It("should fail in create volume but succeed to validate vol name len", func() {
			fakeScbeDataModel.GetVolumeReturns(scbe.ScbeVolume{}, false, nil)
			fakeScbeRestClient.CreateVolumeReturns(scbe.ScbeVolumeInfo{}, fmt.Errorf("error"))

			opts := make(map[string]interface{})
			opts[scbe.OptionNameForVolumeSize] = "100"
			maxVolNameCapable := scbe.MaxVolumeNameLength - (len(fakeConfig.UbiquityInstanceName) + 3)

			volName := strings.Repeat("x", maxVolNameCapable)
			req := resources.CreateVolumeRequest{Name: volName, Backend: resources.SCBE, Opts: opts}
			err = client.CreateVolume(req)
			Expect(err).To(HaveOccurred())
			_, ok := err.(*scbe.VolumeNameExceededMaxLengthError)
			Expect(ok).To(Equal(false))
			Expect(err.Error()).To(Equal("error"))

		})

		It("should fail create volume if vol creation failed with err", func() {
			fakeScbeDataModel.GetVolumeReturns(scbe.ScbeVolume{}, false, nil)
			fakeScbeRestClient.CreateVolumeReturns(scbe.ScbeVolumeInfo{}, fmt.Errorf("error"))
			opts := make(map[string]interface{})
			opts[scbe.OptionNameForVolumeSize] = "100"

			volFake := "fakevol"
			req := resources.CreateVolumeRequest{Name: volFake, Backend: resources.SCBE, Opts: opts}
			err = client.CreateVolume(req)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("error"))
			Expect(fakeScbeRestClient.CreateVolumeCallCount()).To(Equal(1))
			volname, profile, size := fakeScbeRestClient.CreateVolumeArgsForCall(0)
			Expect(profile).To(Equal(fakeDefaultProfile))
			Expect(size).To(Equal(100))
			expectedVolName := fmt.Sprintf(scbe.ComposeVolumeName, fakeConfig.UbiquityInstanceName, volFake)
			Expect(volname).To(Equal(expectedVolName))
		})
		It("should fail create volume if vol creation failed with err", func() {
			fakeScbeDataModel.GetVolumeReturns(scbe.ScbeVolume{}, false, nil)
			fakeScbeRestClient.CreateVolumeReturns(scbe.ScbeVolumeInfo{}, fmt.Errorf("error"))
			opts := make(map[string]interface{})
			opts[scbe.OptionNameForVolumeSize] = "100"
			opts[scbe.OptionNameForServiceName] = "gold"

			volFake := "fakevol"
			req := resources.CreateVolumeRequest{Name: volFake, Backend: resources.SCBE, Opts: opts}
			err = client.CreateVolume(req)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("error"))
			Expect(fakeScbeRestClient.CreateVolumeCallCount()).To(Equal(1))
			volname, profile, size := fakeScbeRestClient.CreateVolumeArgsForCall(0)
			Expect(profile).To(Equal("gold"))
			Expect(size).To(Equal(100))
			expectedVolName := fmt.Sprintf(scbe.ComposeVolumeName, fakeConfig.UbiquityInstanceName, volFake)
			Expect(volname).To(Equal(expectedVolName))
		})

		It("should fail to insert vol to DB after create it", func() {
			fakeScbeDataModel.GetVolumeReturns(scbe.ScbeVolume{}, false, nil)
			fakeScbeRestClient.CreateVolumeReturns(scbe.ScbeVolumeInfo{
				Name: "v1", Wwn: "wwn1", Profile: "gold"}, nil)
			fakeScbeDataModel.InsertVolumeReturns(fmt.Errorf("error"))
			opts := make(map[string]interface{})
			opts[scbe.OptionNameForVolumeSize] = "100"
			opts[scbe.OptionNameForServiceName] = "gold"

			volFake := "fakevol"
			req := resources.CreateVolumeRequest{Name: volFake, Backend: resources.SCBE, Opts: opts}
			err = client.CreateVolume(req)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("error"))
			Expect(fakeScbeDataModel.InsertVolumeCallCount()).To(Equal(1))
			name, wwn, host := fakeScbeDataModel.InsertVolumeArgsForCall(0)
			Expect(name).To(Equal(volFake))
			Expect(wwn).To(Equal("wwn1"))
			Expect(host).To(Equal(scbe.AttachedToNothing))
		})

		It("should succeed to insert vol to DB after create it", func() {
			fakeScbeDataModel.GetVolumeReturns(scbe.ScbeVolume{}, false, nil)
			fakeScbeRestClient.CreateVolumeReturns(scbe.ScbeVolumeInfo{Name: "v1", Wwn: "wwn1", Profile: "gold"}, nil)
			fakeScbeDataModel.InsertVolumeReturns(nil)
			opts := make(map[string]interface{})
			opts[scbe.OptionNameForVolumeSize] = "100"
			opts[scbe.OptionNameForServiceName] = "gold"

			volFake := "fakevol"
			req := resources.CreateVolumeRequest{Name: volFake, Backend: resources.SCBE, Opts: opts}
			err = client.CreateVolume(req)
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeScbeDataModel.InsertVolumeCallCount()).To(Equal(1))
			name, wwn, host := fakeScbeDataModel.InsertVolumeArgsForCall(0)
			Expect(name).To(Equal(volFake))
			Expect(wwn).To(Equal("wwn1"))
			Expect(host).To(Equal(scbe.AttachedToNothing))
		})
	})
})

var _ = Describe("scbeLocalClient", func() {
	var (
		client             resources.StorageClient
		fakeScbeDataModel  *fakes.FakeScbeDataModel
		fakeScbeRestClient *fakes.FakeScbeRestClient
		fakeConfig         resources.ScbeConfig
		fakeErr            error = errors.New("fake error")
		err                error
	)
	BeforeEach(func() {
		fakeScbeDataModel = new(fakes.FakeScbeDataModel)
		fakeScbeRestClient = new(fakes.FakeScbeRestClient)
		fakeConfig = resources.ScbeConfig{
			ConfigPath:     "/tmp",
			DefaultService: fakeDefaultProfile}

		fakeScbeRestClient.LoginReturns(nil)
		fakeScbeRestClient.ServiceExistReturns(true, nil)
		client, err = scbe.NewScbeLocalClientWithNewScbeRestClientAndDataModel(
			fakeConfig,
			fakeScbeDataModel,
			fakeScbeRestClient)
		Expect(err).ToNot(HaveOccurred())
		Expect(fakeScbeRestClient.LoginCallCount()).To(Equal(1))
		Expect(fakeScbeRestClient.ServiceExistCallCount()).To(Equal(1))
	})

	Context(".Attach", func() {
		It("should fail to attach request is bad", func() {
			_, err := client.Attach(resources.AttachRequest{Name: "AAA", Host: scbe.EmptyHost})
			Expect(err).To(HaveOccurred())
			_, ok := err.(*scbe.InValidRequestError)
			Expect(ok).To(Equal(true))
			_, err = client.Attach(resources.AttachRequest{Name: "", Host: fakeHost})
			Expect(err).To(HaveOccurred())
			_, ok = err.(*scbe.InValidRequestError)
			Expect(ok).To(Equal(true))
		})

		It("should fail to attach the volume if GetVolume failed", func() {
			fakeScbeDataModel.GetVolumeReturns(scbe.ScbeVolume{}, false, fakeErr)
			_, err := client.Attach(fakeAttachRequest)
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(fakeErr))
		})
		It("should fail to attach the volume if vol not exist in DB", func() {
			fakeScbeDataModel.GetVolumeReturns(
				scbe.ScbeVolume{}, false, nil)
			_, err := client.Attach(fakeAttachRequest)
			Expect(err).To(HaveOccurred())
		})
		It("should fail to attach the volume if vol already attached in DB", func() {
			fakeScbeDataModel.GetVolumeReturns(
				scbe.ScbeVolume{AttachTo: fakeHost2}, true, nil)
			_, err := client.Attach(fakeAttachRequest)
			Expect(err).To(HaveOccurred())
		})
		It("should fail to attach the volume if MapVolume failed", func() {
			fakeScbeDataModel.GetVolumeReturns(scbe.ScbeVolume{AttachTo: ""}, true, nil)
			fakeScbeRestClient.MapVolumeReturns(scbe.ScbeResponseMapping{}, fakeErr)
			_, err := client.Attach(fakeAttachRequest)
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(fakeErr))
			Expect(fakeScbeRestClient.MapVolumeCallCount()).To(Equal(1))
		})
		It("should fail to attach the volume if update the vol in the DB failed", func() {
			fakeScbeDataModel.GetVolumeReturns(scbe.ScbeVolume{AttachTo: ""}, true, nil)
			fakeScbeRestClient.MapVolumeReturns(scbe.ScbeResponseMapping{}, nil)
			fakeScbeDataModel.UpdateVolumeAttachToReturns(fakeErr)
			_, err := client.Attach(fakeAttachRequest)
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(fakeErr))
			Expect(fakeScbeDataModel.UpdateVolumeAttachToCallCount()).To(Equal(1))
		})
		It("should succeed to attach the volume when everything is cool", func() {
			fakeScbeDataModel.GetVolumeReturns(
				scbe.ScbeVolume{AttachTo: ""}, true, nil)
			fakeScbeRestClient.MapVolumeReturns(scbe.ScbeResponseMapping{}, nil)
			fakeScbeDataModel.UpdateVolumeAttachToReturns(nil)
			_, err := client.Attach(fakeAttachRequest)
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeScbeDataModel.UpdateVolumeAttachToCallCount()).To(Equal(1))
		})
		It("should succeed to attach the volume if vol already attach to this host", func() {
			fakeScbeDataModel.GetVolumeReturns(
				scbe.ScbeVolume{AttachTo: fakeHost}, true, nil)
			_, err := client.Attach(fakeAttachRequest)
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeScbeDataModel.UpdateVolumeAttachToCallCount()).To(Equal(0))
		})

	})
	Context(".Detach", func() {
		It("should fail to detach the volume if GetVolume failed", func() {
			fakeScbeDataModel.GetVolumeReturns(scbe.ScbeVolume{}, false, fakeErr)
			err := client.Detach(fakeDetachRequest)
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(fakeErr))
		})
		It("should fail to detach the volume if vol not exist in DB", func() {
			fakeScbeDataModel.GetVolumeReturns(scbe.ScbeVolume{}, false, nil)
			err := client.Detach(fakeDetachRequest)
			Expect(err).To(HaveOccurred())
		})
		It("should fail to detach the volume if vol already detached in DB", func() {
			fakeScbeDataModel.GetVolumeReturns(
				scbe.ScbeVolume{AttachTo: scbe.EmptyHost}, true, nil)
			err := client.Detach(fakeDetachRequest)
			Expect(err).To(HaveOccurred())
		})
		It("should fail to detach the volume if MapVolume failed", func() {
			fakeScbeDataModel.GetVolumeReturns(scbe.ScbeVolume{AttachTo: fakeHost}, true, nil)
			fakeScbeRestClient.UnmapVolumeReturns(fakeErr)
			err := client.Detach(fakeDetachRequest)
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(fakeErr))
			Expect(fakeScbeRestClient.UnmapVolumeCallCount()).To(Equal(1))
		})
		It("should fail to detach the volume if update the vol in the DB failed", func() {
			fakeScbeDataModel.GetVolumeReturns(scbe.ScbeVolume{AttachTo: fakeHost}, true, nil)
			fakeScbeRestClient.UnmapVolumeReturns(nil)
			fakeScbeDataModel.UpdateVolumeAttachToReturns(fakeErr)
			err := client.Detach(fakeDetachRequest)
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(fakeErr))
			Expect(fakeScbeDataModel.UpdateVolumeAttachToCallCount()).To(Equal(1))
		})
		It("should succeed to detach the volume when everything is cool", func() {
			fakeScbeDataModel.GetVolumeReturns(
				scbe.ScbeVolume{AttachTo: fakeHost}, true, nil)
			fakeScbeRestClient.UnmapVolumeReturns(nil)
			fakeScbeDataModel.UpdateVolumeAttachToReturns(nil)
			err := client.Detach(fakeDetachRequest)
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeScbeDataModel.UpdateVolumeAttachToCallCount()).To(Equal(1))
		})

	})
	Context(".GetVolumeConfig", func() {
		It("succeed and return volume info", func() {
			volumes := make([]scbe.ScbeVolumeInfo, 1)
			vol := &volumes[0]
			val := reflect.Indirect(reflect.ValueOf(vol))
			for i := 0; i < val.Type().NumField(); i++ {
				reflect.ValueOf(vol).Elem().Field(i).SetString(val.Type().Field(i).Name)
			}
			fakeScbeDataModel.GetVolumeReturns(scbe.ScbeVolume{WWN: "wwn"}, true, nil)
			fakeScbeRestClient.GetVolumesReturns(volumes, nil)
			volConfig, err := client.GetVolumeConfig(resources.GetVolumeConfigRequest{"name"})
			Expect(err).To(Not(HaveOccurred()))
			Expect(len(volConfig)).To(Equal(val.Type().NumField()))
			for k, v := range volConfig {
				Expect(k).To(Not(Equal("")))
				Expect(k).To(Equal(v))
			}
		})
		It("fail upon GetVolume error", func() {
			fakeScbeDataModel.GetVolumeReturns(scbe.ScbeVolume{}, false, fakeErr)
			_, err := client.GetVolumeConfig(resources.GetVolumeConfigRequest{"name"})
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(fakeErr))
		})
		It("fail if GetVolume returns false", func() {
			fakeScbeDataModel.GetVolumeReturns(scbe.ScbeVolume{WWN: "wwn"}, false, nil)
			_, err := client.GetVolumeConfig(resources.GetVolumeConfigRequest{"name"})
			Expect(err).To(HaveOccurred())
		})
		It("fail upon GetVolumes error", func() {
			fakeScbeDataModel.GetVolumeReturns(scbe.ScbeVolume{WWN: "wwn"}, true, nil)
			fakeScbeRestClient.GetVolumesReturns(nil, fakeErr)
			_, err := client.GetVolumeConfig(resources.GetVolumeConfigRequest{"name"})
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(fakeErr))
		})
		It("fail if GetVolumes returned 0 volumes", func() {
			volumes := make([]scbe.ScbeVolumeInfo, 0)
			fakeScbeDataModel.GetVolumeReturns(scbe.ScbeVolume{WWN: "wwn"}, true, nil)
			fakeScbeRestClient.GetVolumesReturns(volumes, nil)
			_, err := client.GetVolumeConfig(resources.GetVolumeConfigRequest{"name"})
			Expect(err).To(HaveOccurred())
		})
	})
	Context(".Remove", func() {
		It("should fail to remove the volume if GetVolume failed", func() {
			fakeScbeDataModel.GetVolumeReturns(scbe.ScbeVolume{}, false, fakeErr)
			err := client.RemoveVolume(fakeRemoveRequest)
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(fakeErr))
		})
		It("should fail to detach the volume if vol not exist in DB", func() {
			fakeScbeDataModel.GetVolumeReturns(scbe.ScbeVolume{}, false, nil)
			err := client.RemoveVolume(fakeRemoveRequest)
			Expect(err).To(HaveOccurred())
		})
		It("should fail to remove the volume if vol already attached in DB", func() {
			fakeScbeDataModel.GetVolumeReturns(
				scbe.ScbeVolume{AttachTo: fakeHost}, true, nil)
			err := client.RemoveVolume(fakeRemoveRequest)
			Expect(err).To(HaveOccurred())
			_, ok := err.(*scbe.CannotDeleteVolWhichAttachedToHostError)
			Expect(ok).To(Equal(true))
		})
		It("should fail to remove the volume if fail to delete vol from system", func() {
			fakeScbeDataModel.GetVolumeReturns(
				scbe.ScbeVolume{AttachTo: scbe.EmptyHost}, true, nil)
			fakeScbeRestClient.DeleteVolumeReturns(fakeErr)

			err := client.RemoveVolume(fakeRemoveRequest)
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(fakeErr))
			Expect(fakeScbeRestClient.DeleteVolumeCallCount()).To(Equal(1))
		})
		It("should fail to remove the volume if fail to delete from DB", func() {
			fakeScbeDataModel.GetVolumeReturns(
				scbe.ScbeVolume{AttachTo: scbe.EmptyHost}, true, nil)
			fakeScbeRestClient.DeleteVolumeReturns(nil)
			fakeScbeDataModel.DeleteVolumeReturns(fakeErr)
			err := client.RemoveVolume(fakeRemoveRequest)
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(fakeErr))
			Expect(fakeScbeRestClient.DeleteVolumeCallCount()).To(Equal(1))
			Expect(fakeScbeDataModel.DeleteVolumeCallCount()).To(Equal(1))
		})
		It("should succeed to remove the volume if all is cool", func() {
			fakeScbeDataModel.GetVolumeReturns(
				scbe.ScbeVolume{AttachTo: scbe.EmptyHost}, true, nil)
			fakeScbeRestClient.DeleteVolumeReturns(nil)
			fakeScbeDataModel.DeleteVolumeReturns(nil)
			err := client.RemoveVolume(fakeRemoveRequest)
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeScbeRestClient.DeleteVolumeCallCount()).To(Equal(1))
			Expect(fakeScbeDataModel.DeleteVolumeCallCount()).To(Equal(1))
		})
	})

})
