/**
 * Copyright 2017 IBM Corp.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

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
		fakeScbeDataModel  *fakes.FakeScbeDataModelWrapper
		fakeScbeRestClient *fakes.FakeScbeRestClient
		fakeConfig         resources.ScbeConfig
		err                error
	)
	BeforeEach(func() {
		fakeScbeDataModel = new(fakes.FakeScbeDataModelWrapper)
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
		It("should fail because DefaultFilesystemType is not supported", func() {
			fakeConfig = resources.ScbeConfig{
				DefaultFilesystemType: "bad fstype",
			}
			client, err = scbe.NewScbeLocalClientWithNewScbeRestClientAndDataModel(
				fakeConfig,
				fakeScbeDataModel,
				fakeScbeRestClient)
			Expect(err).To(HaveOccurred())
			_, ok := err.(*scbe.ConfigDefaultFilesystemTypeNotSupported)
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
		It("should succeed to init because config is ok (ext4)", func() {
			fakeConfig = resources.ScbeConfig{
				UbiquityInstanceName:  "123456789012345",
				DefaultVolumeSize:     "1",
				DefaultFilesystemType: "ext4",
			}
			fakeScbeRestClient.LoginReturns(nil)
			fakeScbeRestClient.ServiceExistReturns(true, nil)

			client, err = scbe.NewScbeLocalClientWithNewScbeRestClientAndDataModel(
				fakeConfig,
				fakeScbeDataModel,
				fakeScbeRestClient)
			Expect(err).NotTo(HaveOccurred())
		})
		It("should succeed to init because config is ok (xsf)", func() {
			fakeConfig = resources.ScbeConfig{
				DefaultFilesystemType: "xfs",
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
		fakeScbeDataModel  *fakes.FakeScbeDataModelWrapper
		fakeScbeRestClient *fakes.FakeScbeRestClient
		fakeConfig         resources.ScbeConfig
		err                error
	)
	BeforeEach(func() {
		fakeScbeDataModel = new(fakes.FakeScbeDataModelWrapper)
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
			Expect(err.Error()).To(MatchRegexp("^SCBE backend activation error. The default service .* does not exist on SCBE"))
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
		fakeScbeDataModel  *fakes.FakeScbeDataModelWrapper
		fakeScbeRestClient *fakes.FakeScbeRestClient
		fakeConfig         resources.ScbeConfig
		fakeErr            error = errors.New("fake error")
		err                error
	)
	BeforeEach(func() {
		fakeScbeDataModel = new(fakes.FakeScbeDataModelWrapper)
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
			fakeScbeDataModel.GetVolumeReturns(scbe.ScbeVolume{}, fakeErr)
			req := resources.CreateVolumeRequest{Name: "fakevol", Backend: resources.SCBE, Opts: nil}
			err = client.CreateVolume(req)
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(fakeErr))
		})
		It("should fail create volume if volume already exist in DB", func() {
			fakeScbeDataModel.GetVolumeReturns(scbe.ScbeVolume{}, fakeErr)
			req := resources.CreateVolumeRequest{Name: "fakevol", Backend: resources.SCBE, Opts: nil}
			err = client.CreateVolume(req)
			Expect(err).To(HaveOccurred())
		})
		It("should fail create volume if vol size is not number", func() {
			fakeScbeDataModel.GetVolumeReturns(scbe.ScbeVolume{}, nil)
			opts := make(map[string]interface{})
			opts[scbe.OptionNameForVolumeSize] = "aaa"
			req := resources.CreateVolumeRequest{Name: "fakevol", Backend: resources.SCBE, Opts: opts}
			err = client.CreateVolume(req)
			Expect(err).To(HaveOccurred())
		})
		It("should fail create volume if vol size is not number", func() {
			fakeScbeDataModel.GetVolumeReturns(scbe.ScbeVolume{}, nil)
			opts := make(map[string]interface{})
			opts[resources.OptionNameForVolumeFsType] = "bad-fs-type"
			req := resources.CreateVolumeRequest{Name: "fakevol", Backend: resources.SCBE, Opts: opts}
			err = client.CreateVolume(req)
			Expect(err).To(HaveOccurred())
			_, ok := err.(*scbe.FsTypeNotSupportedError)
			Expect(ok).To(Equal(true))
		})

		It("should fail create volume if vol len exeeded", func() {
			fakeScbeDataModel.GetVolumeReturns(scbe.ScbeVolume{}, nil)
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
			fakeScbeDataModel.GetVolumeReturns(scbe.ScbeVolume{}, nil)
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
			fakeScbeDataModel.GetVolumeReturns(scbe.ScbeVolume{}, nil)
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
			fakeScbeDataModel.GetVolumeReturns(scbe.ScbeVolume{}, nil)
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
			fakeScbeDataModel.GetVolumeReturns(scbe.ScbeVolume{}, nil)
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
			name, wwn, fstype := fakeScbeDataModel.InsertVolumeArgsForCall(0)
			Expect(name).To(Equal(volFake))
			Expect(wwn).To(Equal("wwn1"))
			Expect(fstype).To(Equal("ext4"))
		})

		It("should succeed to insert vol to DB after create it", func() {
			fakeScbeDataModel.GetVolumeReturns(scbe.ScbeVolume{}, nil)
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
			name, wwn, fstype := fakeScbeDataModel.InsertVolumeArgsForCall(0)
			Expect(name).To(Equal(volFake))
			Expect(wwn).To(Equal("wwn1"))
			Expect(fstype).To(Equal("ext4"))
		})
		It("should succeed to insert vol to DB even if size not provided", func() {
			fakeScbeDataModel.GetVolumeReturns(scbe.ScbeVolume{}, nil)
			fakeScbeRestClient.CreateVolumeReturns(scbe.ScbeVolumeInfo{Name: "v1", Wwn: "wwn1", Profile: "gold"}, nil)
			fakeScbeDataModel.InsertVolumeReturns(nil)
			opts := make(map[string]interface{})
			//opts[scbe.OptionNameForVolumeSize] = "10"
			opts[scbe.OptionNameForServiceName] = "gold"

			volFake := "fakevol"
			req := resources.CreateVolumeRequest{Name: volFake, Backend: resources.SCBE, Opts: opts}
			err = client.CreateVolume(req)
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeScbeDataModel.InsertVolumeCallCount()).To(Equal(1))
			name, wwn, fstype := fakeScbeDataModel.InsertVolumeArgsForCall(0)
			Expect(name).To(Equal(volFake))
			Expect(wwn).To(Equal("wwn1"))
			Expect(fstype).To(Equal("ext4"))
		})
		It("should succeed to insert vol to DB even if size not provided (xfs)", func() {
			fakeScbeDataModel.GetVolumeReturns(scbe.ScbeVolume{}, nil)
			fakeScbeRestClient.CreateVolumeReturns(scbe.ScbeVolumeInfo{Name: "v1", Wwn: "wwn1", Profile: "gold"}, nil)
			fakeScbeDataModel.InsertVolumeReturns(nil)
			opts := make(map[string]interface{})
			opts[resources.OptionNameForVolumeFsType] = "xfs"
			opts[scbe.OptionNameForServiceName] = "gold"

			volFake := "fakevol"
			req := resources.CreateVolumeRequest{Name: volFake, Backend: resources.SCBE, Opts: opts}
			err = client.CreateVolume(req)
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeScbeDataModel.InsertVolumeCallCount()).To(Equal(1))
			name, wwn, fstype := fakeScbeDataModel.InsertVolumeArgsForCall(0)
			Expect(name).To(Equal(volFake))
			Expect(wwn).To(Equal("wwn1"))
			Expect(fstype).To(Equal("xfs"))
		})

	})
})

var _ = Describe("scbeLocalClient", func() {
	var (
		client             resources.StorageClient
		fakeScbeDataModel  *fakes.FakeScbeDataModelWrapper
		fakeScbeRestClient *fakes.FakeScbeRestClient
		fakeConfig         resources.ScbeConfig
		fakeErr            error = errors.New("fake error")
		err                error
	)
	BeforeEach(func() {
		fakeScbeDataModel = new(fakes.FakeScbeDataModelWrapper)
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
	Context(".Activate", func() {
		It("should succeed", func() {
			err := client.Activate(resources.ActivateRequest{})
			Expect(err).To(Not(HaveOccurred()))
		})
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
			fakeScbeDataModel.GetVolumeReturns(scbe.ScbeVolume{}, fakeErr)
			_, err := client.Attach(fakeAttachRequest)
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(fakeErr))
		})
		It("should fail to attach the volume if vol not exist in DB", func() {
			fakeScbeDataModel.GetVolumeReturns(
				scbe.ScbeVolume{}, fakeErr)
			_, err := client.Attach(fakeAttachRequest)
			Expect(err).To(HaveOccurred())
		})
		It("should fail to attach the volume if MapVolume failed", func() {
			fakeScbeDataModel.GetVolumeReturns(scbe.ScbeVolume{}, nil)
			fakeScbeRestClient.MapVolumeReturns(scbe.ScbeResponseMapping{}, fakeErr)
			_, err := client.Attach(fakeAttachRequest)
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(fakeErr))
			Expect(fakeScbeRestClient.MapVolumeCallCount()).To(Equal(1))
		})
	})
	Context(".Detach", func() {
		It("should fail to detach the volume if GetVolume failed", func() {
			fakeScbeDataModel.GetVolumeReturns(scbe.ScbeVolume{}, fakeErr)
			err := client.Detach(fakeDetachRequest)
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(fakeErr))
		})
		It("should fail to detach the volume if vol not exist in DB", func() {
			fakeScbeDataModel.GetVolumeReturns(scbe.ScbeVolume{}, nil)
			err := client.Detach(fakeDetachRequest)
			Expect(err).To(HaveOccurred())
		})
		It("should fail to detach the volume if vol already detached in DB", func() {
			fakeScbeDataModel.GetVolumeReturns(scbe.ScbeVolume{}, nil)
			err := client.Detach(fakeDetachRequest)
			Expect(err).To(HaveOccurred())
		})
		It("should fail to detach the volume if MapVolume failed", func() {
			fakeScbeDataModel.GetVolumeReturns(scbe.ScbeVolume{}, nil)
			fakeScbeRestClient.GetVolMappingReturns(fakeHost, nil)
			fakeScbeRestClient.UnmapVolumeReturns(fakeErr)
			err := client.Detach(fakeDetachRequest)
			Expect(err).To(HaveOccurred())
			fmt.Println(err)
			fmt.Println(fakeErr)
			Expect(err).To(MatchError(fakeErr))
			Expect(fakeScbeRestClient.UnmapVolumeCallCount()).To(Equal(1))
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
			fakeScbeDataModel.GetVolumeReturns(scbe.ScbeVolume{WWN: "wwn", FSType: "ext4"}, nil)
			fakeScbeRestClient.GetVolumesReturns(volumes, nil)
			fakeScbeRestClient.GetVolMappingReturns(fakeHost, nil)
			volConfig, err := client.GetVolumeConfig(resources.GetVolumeConfigRequest{Name:"name"})
			Expect(err).To(Not(HaveOccurred()))
			Expect(len(volConfig)).To(Equal(val.Type().NumField() + scbe.GetVolumeConfigExtraParams))
			fstype, ok := volConfig[resources.OptionNameForVolumeFsType]
			Expect(ok).To(Equal(true))
			Expect(fstype).To(Equal("ext4"))
			attachTo, ok := volConfig[resources.ScbeKeyVolAttachToHost]
			Expect(ok).To(Equal(true))
			Expect(attachTo).To(Equal(fakeHost))

			for k, v := range volConfig {
				if k == resources.OptionNameForVolumeFsType || k == resources.ScbeKeyVolAttachToHost {
					continue
				}
				Expect(k).To(Not(Equal("")))
				Expect(k).To(Equal(v))
			}
		})
		It("fail upon GetVolume error", func() {
			fakeScbeDataModel.GetVolumeReturns(scbe.ScbeVolume{}, fakeErr)
			_, err := client.GetVolumeConfig(resources.GetVolumeConfigRequest{Name:"name"})
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(fakeErr))
		})
		It("fail if GetVolume returns false", func() {
			fakeScbeDataModel.GetVolumeReturns(scbe.ScbeVolume{WWN: "wwn"}, nil)
			_, err := client.GetVolumeConfig(resources.GetVolumeConfigRequest{Name:"name"})
			Expect(err).To(HaveOccurred())
		})
		It("fail upon GetVolumes error", func() {
			fakeScbeDataModel.GetVolumeReturns(scbe.ScbeVolume{WWN: "wwn"}, nil)
			fakeScbeRestClient.GetVolumesReturns(nil, fakeErr)
			_, err := client.GetVolumeConfig(resources.GetVolumeConfigRequest{Name:"name"})
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(fakeErr))
		})
		It("fail if GetVolumes returned 0 volumes", func() {
			volumes := make([]scbe.ScbeVolumeInfo, 0)
			fakeScbeDataModel.GetVolumeReturns(scbe.ScbeVolume{WWN: "wwn"}, nil)
			fakeScbeRestClient.GetVolumesReturns(volumes, nil)
			_, err := client.GetVolumeConfig(resources.GetVolumeConfigRequest{Name:"name"})
			Expect(err).To(HaveOccurred())
		})
	})
	Context(".Remove", func() {
		It("should fail to remove the volume if GetVolume failed", func() {
			fakeScbeDataModel.GetVolumeReturns(scbe.ScbeVolume{}, fakeErr)
			err := client.RemoveVolume(fakeRemoveRequest)
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(fakeErr))
		})
		It("should fail to detach the volume if vol not exist in DB", func() {
			fakeScbeDataModel.GetVolumeReturns(scbe.ScbeVolume{}, fakeErr)
			err := client.RemoveVolume(fakeRemoveRequest)
			Expect(err).To(HaveOccurred())
		})
		It("should fail to remove the volume if fail to delete vol from system", func() {
			fakeScbeDataModel.GetVolumeReturns(scbe.ScbeVolume{}, nil)
			fakeScbeRestClient.DeleteVolumeReturns(fakeErr)

			err := client.RemoveVolume(fakeRemoveRequest)
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(fakeErr))
			Expect(fakeScbeRestClient.DeleteVolumeCallCount()).To(Equal(1))
		})
		It("should fail to remove the volume if fail to delete from DB", func() {
			fakeScbeDataModel.GetVolumeReturns(scbe.ScbeVolume{}, nil)
			fakeScbeRestClient.DeleteVolumeReturns(nil)
			fakeScbeDataModel.DeleteVolumeReturns(fakeErr)
			err := client.RemoveVolume(fakeRemoveRequest)
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(fakeErr))
			Expect(fakeScbeRestClient.DeleteVolumeCallCount()).To(Equal(1))
			Expect(fakeScbeDataModel.DeleteVolumeCallCount()).To(Equal(1))
		})
		It("should succeed to remove the volume if all is cool", func() {
			fakeScbeDataModel.GetVolumeReturns(scbe.ScbeVolume{}, nil)
			fakeScbeRestClient.DeleteVolumeReturns(nil)
			fakeScbeDataModel.DeleteVolumeReturns(nil)
			err := client.RemoveVolume(fakeRemoveRequest)
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeScbeRestClient.DeleteVolumeCallCount()).To(Equal(1))
			Expect(fakeScbeDataModel.DeleteVolumeCallCount()).To(Equal(1))
		})
	})

})

var _ = Describe("scbeLocalClient", func() {
	var (
		client             resources.StorageClient
		fakeScbeDataModel  *fakes.FakeScbeDataModelWrapper
		fakeScbeRestClient *fakes.FakeScbeRestClient
		fakeConfig         resources.ScbeConfig
		fakeCredentialInfo resources.CredentialInfo
		fakeConnectionInfo resources.ConnectionInfo
		err                error
	)
	BeforeEach(func() {
		fakeCredentialInfo = resources.CredentialInfo{UserName: "user1", Password: "pass1"}
		fakeConnectionInfo = resources.ConnectionInfo{CredentialInfo: fakeCredentialInfo}
		fakeConfig = resources.ScbeConfig{
			ConfigPath:     "/tmp",
			DefaultService: fakeDefaultProfile,
			ConnectionInfo: fakeConnectionInfo,
		}
	})
	Context(".getAuthenticatedScbeRestClient", func() {
		It("call with same credentialInfo should not login again", func() {
			fakeScbeDataModel = new(fakes.FakeScbeDataModelWrapper)
			fakeScbeRestClient = new(fakes.FakeScbeRestClient)
			fakeScbeRestClient.LoginReturns(nil)
			fakeScbeRestClient.ServiceExistReturns(true, nil)
			defer scbe.InitScbeRestClientGen(GenFakeScbeRestClient(nil))()
			client, err = scbe.NewScbeLocalClientWithNewScbeRestClientAndDataModel(
				fakeConfig,
				fakeScbeDataModel,
				fakeScbeRestClient)
			Expect(err).ToNot(HaveOccurred())
			Expect(fakeScbeRestClient.LoginCallCount()).To(Equal(1))
			Expect(fakeScbeRestClient.ServiceExistCallCount()).To(Equal(1))
			Expect(fakeScbeRestClient.LoginCallCount()).To(Equal(1))
			err = client.Activate(resources.ActivateRequest{CredentialInfo: fakeCredentialInfo})
			Expect(err).To(Not(HaveOccurred()))
			Expect(fakeScbeRestClient.LoginCallCount()).To(Equal(1))
		})
	})
	Context(".getAuthenticatedScbeRestClient", func() {
		It("call with other credentialInfo should login again and cache the new client", func() {
			fakeScbeDataModel = new(fakes.FakeScbeDataModelWrapper)
			fakeScbeRestClient = new(fakes.FakeScbeRestClient)
			fakeScbeRestClient.LoginReturns(nil)
			fakeScbeRestClient.ServiceExistReturns(true, nil)
			otherFakeScbeRestClient := new(fakes.FakeScbeRestClient)
			otherFakeScbeRestClient.LoginReturns(nil)
			otherFakeScbeRestClient.ServiceExistReturns(true, nil)
			defer scbe.InitScbeRestClientGen(GenFakeScbeRestClient(otherFakeScbeRestClient))()
			client, err = scbe.NewScbeLocalClientWithNewScbeRestClientAndDataModel(
				fakeConfig,
				fakeScbeDataModel,
				fakeScbeRestClient)
			Expect(err).ToNot(HaveOccurred())
			Expect(fakeScbeRestClient.LoginCallCount()).To(Equal(1))
			Expect(fakeScbeRestClient.ServiceExistCallCount()).To(Equal(1))
			Expect(otherFakeScbeRestClient.LoginCallCount()).To(Equal(0))
			newCredentialInfo := resources.CredentialInfo{UserName: "user2", Password: "pass2"}
			err = client.Activate(resources.ActivateRequest{CredentialInfo: newCredentialInfo})
			Expect(err).To(Not(HaveOccurred()))
			Expect(otherFakeScbeRestClient.LoginCallCount()).To(Equal(1))
			volumes := make([]scbe.ScbeVolumeInfo, 1)
			vol := &volumes[0]
			val := reflect.Indirect(reflect.ValueOf(vol))
			for i := 0; i < val.Type().NumField(); i++ {
				reflect.ValueOf(vol).Elem().Field(i).SetString(val.Type().Field(i).Name)
			}
			fakeScbeDataModel.GetVolumeReturns(scbe.ScbeVolume{WWN: "wwn", FSType: "ext4"}, nil)
			otherFakeScbeRestClient.GetVolumesReturns(volumes, nil)
			_, err := client.GetVolumeConfig(resources.GetVolumeConfigRequest{CredentialInfo: resources.CredentialInfo{UserName: "user2", Password: "pass2"}, Name:"name"})
			Expect(err).To(Not(HaveOccurred()))
			Expect(otherFakeScbeRestClient.LoginCallCount()).To(Equal(1))
			Expect(otherFakeScbeRestClient.GetVolumesCallCount()).To(Equal(1))
		})
	})
})


func GenFakeScbeRestClient(newClient scbe.ScbeRestClient) scbe.ScbeRestClientGen {
	return func(conInfo resources.ConnectionInfo) (scbe.ScbeRestClient, error) {
		return newClient, nil
	}
}
