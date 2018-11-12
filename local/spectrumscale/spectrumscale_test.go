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

package spectrumscale_test

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/IBM/ubiquity/local/spectrumscale"

	"github.com/IBM/ubiquity/fakes"
	"github.com/IBM/ubiquity/resources"
	"github.com/IBM/ubiquity/utils/logs"
)

var _ = Describe("local-client", func() {
	var (
		client                     resources.StorageClient
		logger                     logs.Logger
		fakeSpectrumScaleConnector *fakes.FakeSpectrumScaleConnector
		fakeSpectrumDataModel      *fakes.FakeSpectrumDataModelWrapper
		fakeExec                   *fakes.FakeExecutor
		fakeConfig                 resources.SpectrumScaleConfig
		activateRequest            resources.ActivateRequest
		createVolumeRequest        resources.CreateVolumeRequest
		removeVolumeRequest        resources.RemoveVolumeRequest
		getVolumeRequest           resources.GetVolumeRequest
		backends                   []string
		// getVolumeConfigRequest     resources.GetVolumeConfigRequest
		listVolumesRequest resources.ListVolumesRequest
		err                error
	)
	BeforeEach(func() {
		logger = logs.GetLogger()
		fakeSpectrumScaleConnector = new(fakes.FakeSpectrumScaleConnector)
		backends = []string{resources.SpectrumScale}
		fakeExec = new(fakes.FakeExecutor)
		fakeSpectrumDataModel = new(fakes.FakeSpectrumDataModelWrapper)
		fakeConfig = resources.SpectrumScaleConfig{}
		activateRequest = resources.ActivateRequest{Backends: backends}
		client, err = spectrumscale.NewSpectrumLocalClientWithConnectors(logger, fakeSpectrumScaleConnector, fakeExec, fakeConfig, fakeSpectrumDataModel)
		Expect(err).ToNot(HaveOccurred())

	})

	Context(".Activate", func() {
		It("should succeed when everything is fine", func() {
			err = client.Activate(activateRequest)
			Expect(err).ToNot(HaveOccurred())
		})
		It("should succeed on subsequent activate without duplicating execution", func() {
			err = client.Activate(activateRequest)
			Expect(err).ToNot(HaveOccurred())

			err = client.Activate(activateRequest)
			Expect(err).ToNot(HaveOccurred())

		})

	})

	Context(".CreateVolume", func() {
		var (
			opts map[string]interface{}
		)
		BeforeEach(func() {
			fakeConfig.DefaultFilesystemName = "fake-config-filesystem"
			client, err = spectrumscale.NewSpectrumLocalClientWithConnectors(logger, fakeSpectrumScaleConnector, fakeExec, fakeConfig, fakeSpectrumDataModel)
			Expect(err).ToNot(HaveOccurred())
			fakeSpectrumScaleConnector.IsFilesystemMountedReturns(false, nil)
			err = client.Activate(activateRequest)
			Expect(err).ToNot(HaveOccurred())
			createVolumeRequest = resources.CreateVolumeRequest{Name: "fake-volume", Opts: opts}
		})


		It("should pass when filesystem is mounted", func() {
			fakeSpectrumDataModel.GetVolumeReturns(spectrumscale.SpectrumScaleVolume{}, false, nil)
			fakeSpectrumScaleConnector.IsFilesystemMountedReturns(true, nil)
			err = client.CreateVolume(createVolumeRequest)
			Expect(err).ToNot(HaveOccurred())
			Expect(fakeSpectrumDataModel.GetVolumeCallCount()).To(Equal(1))
			Expect(fakeSpectrumScaleConnector.CreateFilesetCallCount()).To(Equal(1))
		})

		It("should fail since we unable to fetch filesystem mounted status", func() {
			fakeSpectrumDataModel.GetVolumeReturns(spectrumscale.SpectrumScaleVolume{}, false, nil)
			fakeSpectrumScaleConnector.IsFilesystemMountedReturns(false, fmt.Errorf("Failed to fetch filesystem"))
			err = client.CreateVolume(createVolumeRequest)
			Expect(err.Error()).To(Equal("Failed to check if Filesystem [fake-config-filesystem] is mounted"))
			Expect(fakeSpectrumDataModel.GetVolumeCallCount()).To(Equal(1))
		})

		It("should fail since filesystem is not mounted", func() {
			fakeSpectrumDataModel.GetVolumeReturns(spectrumscale.SpectrumScaleVolume{}, false, nil)
			fakeSpectrumScaleConnector.IsFilesystemMountedReturns(false, nil)
			err = client.CreateVolume(createVolumeRequest)
			Expect(err.Error()).To(Equal("SpectrumScale filesystem [fake-config-filesystem] is not mounted"))
			Expect(fakeSpectrumDataModel.GetVolumeCallCount()).To(Equal(1))
		})

		It("should pass when filesystem (present in options) is mounted", func() {
			fakeSpectrumDataModel.GetVolumeReturns(spectrumscale.SpectrumScaleVolume{}, false, nil)
			fakeSpectrumScaleConnector.IsFilesystemMountedReturns(true, nil)
			createVolumeRequest.Opts = make(map[string]interface{})
			createVolumeRequest.Opts["filesystem"] = "fake-opt-filesystem"
			err = client.CreateVolume(createVolumeRequest)
			Expect(err).ToNot(HaveOccurred())
			Expect(fakeSpectrumDataModel.GetVolumeCallCount()).To(Equal(1))
			Expect(fakeSpectrumScaleConnector.CreateFilesetCallCount()).To(Equal(1))
		})

		It("should fail since we unable to fetch filesystem (present in options) mounted status", func() {
			fakeSpectrumDataModel.GetVolumeReturns(spectrumscale.SpectrumScaleVolume{}, false, nil)
			fakeSpectrumScaleConnector.IsFilesystemMountedReturns(false, fmt.Errorf("Failed to fetch filesystem"))
			createVolumeRequest.Opts = make(map[string]interface{})
			createVolumeRequest.Opts["filesystem"] = "fake-opt-filesystem"
			err = client.CreateVolume(createVolumeRequest)
			Expect(err.Error()).To(Equal("Failed to check if Filesystem [fake-opt-filesystem] is mounted"))
			Expect(fakeSpectrumDataModel.GetVolumeCallCount()).To(Equal(1))
		})

		It("should fail since filesystem (present in options) is not mounted", func() {
			fakeSpectrumDataModel.GetVolumeReturns(spectrumscale.SpectrumScaleVolume{}, false, nil)
			fakeSpectrumScaleConnector.IsFilesystemMountedReturns(false, nil)
			createVolumeRequest.Opts = make(map[string]interface{})
			createVolumeRequest.Opts["filesystem"] = "fake-opt-filesystem"
			err = client.CreateVolume(createVolumeRequest)
			Expect(err.Error()).To(Equal("SpectrumScale filesystem [fake-opt-filesystem] is not mounted"))
			Expect(fakeSpectrumDataModel.GetVolumeCallCount()).To(Equal(1))
		})

		It("should fail when dbClient volumeExists errors", func() {
			fakeSpectrumDataModel.GetVolumeReturns(spectrumscale.SpectrumScaleVolume{}, false, fmt.Errorf("error checking if volume exists"))
			err = client.CreateVolume(createVolumeRequest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("error checking if volume exists"))
			Expect(fakeSpectrumDataModel.GetVolumeCallCount()).To(Equal(1))
			Expect(fakeSpectrumScaleConnector.CreateFilesetCallCount()).To(Equal(0))
		})

		It("should fail when dbClient volumeExists returns true", func() {
			fakeSpectrumDataModel.GetVolumeReturns(spectrumscale.SpectrumScaleVolume{}, true, nil)
			err = client.CreateVolume(createVolumeRequest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Volume already exists"))
			Expect(fakeSpectrumDataModel.GetVolumeCallCount()).To(Equal(1))
			Expect(fakeSpectrumScaleConnector.CreateFilesetCallCount()).To(Equal(0))
		})

		Context(".FilesetVolume", func() {
			BeforeEach(func() {
				opts = make(map[string]interface{})
				opts[""] = ""
				createVolumeRequest = resources.CreateVolumeRequest{Name: "fake-fileset", Opts: opts}
			})

			It("should fail when spectrum client fails to create fileset", func() {
				fakeSpectrumScaleConnector.CreateFilesetReturns(fmt.Errorf("error creating fileset"))
				fakeSpectrumScaleConnector.IsFilesystemMountedReturns(true, nil)
				err = client.CreateVolume(createVolumeRequest)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("error creating fileset"))
				Expect(fakeSpectrumScaleConnector.CreateFilesetCallCount()).To(Equal(1))
				Expect(fakeSpectrumDataModel.InsertFilesetVolumeCallCount()).To(Equal(0))
			})

			It("should fail when dbClient fails to insert fileset record", func() {
				fakeSpectrumScaleConnector.CreateFilesetReturns(nil)
				fakeSpectrumDataModel.InsertFilesetVolumeReturns(fmt.Errorf("error inserting fileset"))
				fakeSpectrumScaleConnector.IsFilesystemMountedReturns(true, nil)
				err = client.CreateVolume(createVolumeRequest)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("error inserting fileset"))
				Expect(fakeSpectrumScaleConnector.CreateFilesetCallCount()).To(Equal(1))
				Expect(fakeSpectrumDataModel.InsertFilesetVolumeCallCount()).To(Equal(1))
			})

			It("should succeed to create fileset", func() {
				fakeSpectrumScaleConnector.CreateFilesetReturns(nil)
				fakeSpectrumDataModel.InsertFilesetVolumeReturns(nil)
				fakeSpectrumScaleConnector.IsFilesystemMountedReturns(true, nil)
				err = client.CreateVolume(createVolumeRequest)
				Expect(err).ToNot(HaveOccurred())
				Expect(fakeSpectrumScaleConnector.CreateFilesetCallCount()).To(Equal(1))
				Expect(fakeSpectrumDataModel.InsertFilesetVolumeCallCount()).To(Equal(1))
			})

		})

		Context(".FilesetVolume", func() {
			BeforeEach(func() {
				opts = make(map[string]interface{})
				opts["fileset"] = "fake-fileset"
				opts["type"] = "fileset"
				opts["filesystem"] = "fake-filesystem"

			})

			Context(".WithoutFilesetWithQuota", func() {
				BeforeEach(func() {
					delete(opts, "fileset")
					opts["quota"] = "1gb"
                    createVolumeRequest = resources.CreateVolumeRequest{Name: "fake-fileset", Opts: opts}
				})

				It("should fail since quota is not enabled for filesystem", func() {
					fakeSpectrumScaleConnector.CheckIfFSQuotaEnabledReturns(fmt.Errorf("Quota not enabled"))
					fakeSpectrumScaleConnector.IsFilesystemMountedReturns(true, nil)
					err = client.CreateVolume(createVolumeRequest)
					Expect(err.Error()).To(Equal("Quota not enabled for Filesystem [fake-filesystem]"))
				})
				It("should pass since quota is enabled for filesystem", func() {
					fakeSpectrumScaleConnector.CheckIfFSQuotaEnabledReturns(nil)
					fakeSpectrumDataModel.InsertFilesetQuotaVolumeReturns(nil)
					fakeSpectrumScaleConnector.IsFilesystemMountedReturns(true, nil)
					err = client.CreateVolume(createVolumeRequest)
					Expect(err).ToNot(HaveOccurred())
				})
			})

			Context(".WithNoQuota", func() {
				BeforeEach(func() {
					createVolumeRequest = resources.CreateVolumeRequest{Name: "fake-fileset", Opts: opts}
				})

				It("should fail when spectrum client fails to list fileset quota", func() {
					fakeSpectrumScaleConnector.ListFilesetReturns(resources.Volume{}, fmt.Errorf("error in list fileset"))
					err = client.CreateVolume(createVolumeRequest)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("error in list fileset"))
					Expect(fakeSpectrumDataModel.InsertFilesetVolumeCallCount()).To(Equal(0))
				})
				It("should fail when dbClient fails to insert Fileset quota volume", func() {
					fakeVolume := resources.Volume{Name: "fake-fileset", Mountpoint: "fake-mountpoint"}
					fakeSpectrumScaleConnector.ListFilesetReturns(fakeVolume, nil)
					fakeSpectrumDataModel.InsertFilesetVolumeReturns(fmt.Errorf("error inserting filesetvolume"))
					err = client.CreateVolume(createVolumeRequest)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("error inserting filesetvolume"))
					Expect(fakeSpectrumDataModel.InsertFilesetVolumeCallCount()).To(Equal(1))
				})
				It("should succeed when parameters are well specified", func() {
					fakeVolume := resources.Volume{Name: "fake-fileset", Mountpoint: "fake-mountpoint"}
					fakeSpectrumScaleConnector.ListFilesetReturns(fakeVolume, nil)
					fakeSpectrumDataModel.InsertFilesetVolumeReturns(nil)
					err = client.CreateVolume(createVolumeRequest)
					Expect(err).ToNot(HaveOccurred())
					Expect(fakeSpectrumDataModel.InsertFilesetVolumeCallCount()).To(Equal(1))
				})

			})

			Context(".WithQuota", func() {
				BeforeEach(func() {
					opts["quota"] = "1gb"
					createVolumeRequest = resources.CreateVolumeRequest{Name: "fake-fileset", Opts: opts}
				})
				It("should fail when spectrum client fails to list fileset quota", func() {
					fakeSpectrumScaleConnector.ListFilesetQuotaReturns("", fmt.Errorf("error in list quota"))
					err = client.CreateVolume(createVolumeRequest)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("error in list quota"))
					Expect(fakeSpectrumDataModel.InsertFilesetQuotaVolumeCallCount()).To(Equal(0))
				})
				It("should pass since user specified quota is less than spectrumscale fileset quota", func() {
					fakeSpectrumScaleConnector.ListFilesetQuotaReturns("2gb", nil)
					err = client.CreateVolume(createVolumeRequest)
					Expect(err).ToNot(HaveOccurred())
					Expect(fakeSpectrumDataModel.InsertFilesetQuotaVolumeCallCount()).To(Equal(1))
				})

				It("should fail when dbClient fails to insert Fileset quota volume", func() {
					fakeSpectrumScaleConnector.ListFilesetQuotaReturns("1gb", nil)
					fakeSpectrumDataModel.InsertFilesetQuotaVolumeReturns(fmt.Errorf("error inserting filesetquotavolume"))
					err = client.CreateVolume(createVolumeRequest)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("error inserting filesetquotavolume"))
					Expect(fakeSpectrumDataModel.InsertFilesetQuotaVolumeCallCount()).To(Equal(1))
				})
				It("should succeed when the options are well specified", func() {
					fakeSpectrumScaleConnector.ListFilesetQuotaReturns("1gb", nil)
					fakeSpectrumDataModel.InsertFilesetQuotaVolumeReturns(nil)
					err = client.CreateVolume(createVolumeRequest)
					Expect(err).ToNot(HaveOccurred())
					Expect(fakeSpectrumScaleConnector.ListFilesetQuotaCallCount()).To(Equal(1))
					Expect(fakeSpectrumDataModel.InsertFilesetQuotaVolumeCallCount()).To(Equal(1))
				})

			})
		})

	})

	Context(".RemoveVolume", func() {
		BeforeEach(func() {
			removeVolumeRequest = resources.RemoveVolumeRequest{Name: "fake-volume"}
		})

		It("should fail when the dbClient fails to check the volume", func() {
			fakeSpectrumDataModel.GetVolumeReturns(spectrumscale.SpectrumScaleVolume{}, false, fmt.Errorf("failed checking volume"))
			err = client.RemoveVolume(removeVolumeRequest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("failed checking volume"))
			Expect(fakeSpectrumDataModel.GetVolumeCallCount()).To(Equal(1))
		})

		It("should fail when the dbClient does not find the volume", func() {
			fakeSpectrumDataModel.GetVolumeReturns(spectrumscale.SpectrumScaleVolume{}, false, nil)
			err = client.RemoveVolume(removeVolumeRequest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("[fake-volume] "+resources.VolumeNotFoundErrorMsg))
			Expect(fakeSpectrumDataModel.GetVolumeCallCount()).To(Equal(1))
		})

		It("should fail when the dbClient fails to get the volume", func() {
			fakeSpectrumDataModel.GetVolumeReturns(spectrumscale.SpectrumScaleVolume{}, false, fmt.Errorf("error getting volume"))
			err = client.RemoveVolume(removeVolumeRequest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("error getting volume"))
			Expect(fakeSpectrumDataModel.GetVolumeCallCount()).To(Equal(1))
			Expect(fakeSpectrumDataModel.DeleteVolumeCallCount()).To(Equal(0))
		})

		Context("When forcedelete is set to true", func() {
			BeforeEach(func() {
				fakeConfig = resources.SpectrumScaleConfig{ForceDelete: true}
				client, err = spectrumscale.NewSpectrumLocalClientWithConnectors(logger, fakeSpectrumScaleConnector, fakeExec, fakeConfig, fakeSpectrumDataModel)
				Expect(err).ToNot(HaveOccurred())

			})

			It("should fail when type is fileset and forceDelete is true and spectrumClient fails to delete fileset", func() {
				volume := spectrumscale.SpectrumScaleVolume{Volume: resources.Volume{Name: "fake-volume"}, FileSystem: "fake-filesystem", Type: 0}
				fakeSpectrumDataModel.GetVolumeReturns(volume, true, nil)
				fakeConfig = resources.SpectrumScaleConfig{ForceDelete: true}
				fakeSpectrumScaleConnector.IsFilesetLinkedReturns(false, nil)
				fakeSpectrumDataModel.DeleteVolumeReturns(nil)
				fakeSpectrumScaleConnector.DeleteFilesetReturns(fmt.Errorf("error deleting fileset"))
				err = client.RemoveVolume(removeVolumeRequest)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("error deleting fileset"))
				Expect(fakeSpectrumDataModel.GetVolumeCallCount()).To(Equal(1))
				Expect(fakeSpectrumDataModel.DeleteVolumeCallCount()).To(Equal(1))
				Expect(fakeSpectrumScaleConnector.DeleteFilesetCallCount()).To(Equal(1))
			})

			It("should succeed when type is fileset and forceDelete is true", func() {
				volume := spectrumscale.SpectrumScaleVolume{Volume: resources.Volume{Name: "fake-volume"}, FileSystem: "fake-filesystem", Type: 0}
				fakeSpectrumDataModel.GetVolumeReturns(volume, true, nil)
				fakeConfig = resources.SpectrumScaleConfig{ForceDelete: true}
				fakeSpectrumScaleConnector.IsFilesetLinkedReturns(false, nil)
				fakeSpectrumDataModel.DeleteVolumeReturns(nil)
				fakeSpectrumScaleConnector.DeleteFilesetReturns(nil)
				err = client.RemoveVolume(removeVolumeRequest)
				Expect(err).ToNot(HaveOccurred())
				Expect(fakeSpectrumDataModel.GetVolumeCallCount()).To(Equal(1))
				Expect(fakeSpectrumDataModel.DeleteVolumeCallCount()).To(Equal(1))
				Expect(fakeSpectrumScaleConnector.DeleteFilesetCallCount()).To(Equal(1))
			})

		})

		It("should fail when type is fileset and spectrumClient fails to check filesetLinked", func() {
			volume := spectrumscale.SpectrumScaleVolume{Volume: resources.Volume{Name: "fake-volume"}, FileSystem: "fake-filesystem", Type: 0}
			fakeSpectrumDataModel.GetVolumeReturns(volume, true, nil)
			fakeSpectrumScaleConnector.IsFilesetLinkedReturns(false, fmt.Errorf("error in IsFilesetLinked"))
			err = client.RemoveVolume(removeVolumeRequest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("error in IsFilesetLinked"))
			Expect(fakeSpectrumDataModel.GetVolumeCallCount()).To(Equal(1))
			Expect(fakeSpectrumDataModel.DeleteVolumeCallCount()).To(Equal(0))
			Expect(fakeSpectrumScaleConnector.IsFilesetLinkedCallCount()).To(Equal(1))
			Expect(fakeSpectrumScaleConnector.UnlinkFilesetCallCount()).To(Equal(0))
		})

		It("should fail when type is fileset and fileset is linked and spectrumClient fails to unlink fileset", func() {
			volume := spectrumscale.SpectrumScaleVolume{Volume: resources.Volume{Name: "fake-volume"}, FileSystem: "fake-filesystem", Type: 0}
			fakeSpectrumDataModel.GetVolumeReturns(volume, true, nil)
			fakeSpectrumScaleConnector.IsFilesetLinkedReturns(true, nil)
			fakeSpectrumScaleConnector.UnlinkFilesetReturns(fmt.Errorf("error in UnlinkFileset"))
			err = client.RemoveVolume(removeVolumeRequest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("error in UnlinkFileset"))
			Expect(fakeSpectrumDataModel.GetVolumeCallCount()).To(Equal(1))
			Expect(fakeSpectrumDataModel.DeleteVolumeCallCount()).To(Equal(0))
			Expect(fakeSpectrumScaleConnector.IsFilesetLinkedCallCount()).To(Equal(1))
			Expect(fakeSpectrumScaleConnector.UnlinkFilesetCallCount()).To(Equal(1))
		})

		It("should fail when type is fileset and dbClient fails to delete fileset", func() {
			volume := spectrumscale.SpectrumScaleVolume{Volume: resources.Volume{Name: "fake-volume"}, FileSystem: "fake-filesystem", Type: 0}
			fakeSpectrumDataModel.GetVolumeReturns(volume, true, nil)
			fakeSpectrumScaleConnector.IsFilesetLinkedReturns(false, nil)
			fakeSpectrumDataModel.DeleteVolumeReturns(fmt.Errorf("error deleting volume"))
			err = client.RemoveVolume(removeVolumeRequest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("error deleting volume"))
			Expect(fakeSpectrumDataModel.GetVolumeCallCount()).To(Equal(1))
			Expect(fakeSpectrumDataModel.DeleteVolumeCallCount()).To(Equal(1))
			Expect(fakeSpectrumScaleConnector.DeleteFilesetCallCount()).To(Equal(0))
		})

		It("should succeed when type is fileset and forceDelete is false", func() {
			volume := spectrumscale.SpectrumScaleVolume{Volume: resources.Volume{Name: "fake-volume"}, FileSystem: "fake-filesystem", Type: 0}
			fakeSpectrumDataModel.GetVolumeReturns(volume, true, nil)
			fakeSpectrumScaleConnector.IsFilesetLinkedReturns(false, nil)
			fakeSpectrumDataModel.DeleteVolumeReturns(nil)
			fakeSpectrumScaleConnector.DeleteFilesetReturns(nil)
			err = client.RemoveVolume(removeVolumeRequest)
			Expect(err).ToNot(HaveOccurred())
			Expect(fakeSpectrumDataModel.GetVolumeCallCount()).To(Equal(1))
			Expect(fakeSpectrumDataModel.DeleteVolumeCallCount()).To(Equal(1))
			Expect(fakeSpectrumScaleConnector.DeleteFilesetCallCount()).To(Equal(0))
		})

	})

	Context(".ListVolumes", func() {
		BeforeEach(func() {
			listVolumesRequest = resources.ListVolumesRequest{Backends: backends}
		})

		It("should fail when dbClient fails to list volumes", func() {
			fakeSpectrumDataModel.ListVolumesReturns(nil, fmt.Errorf("error listing volumes"))
			volumes, err := client.ListVolumes(listVolumesRequest)
			Expect(err).To(HaveOccurred())
			Expect(len(volumes)).To(Equal(0))
			Expect(err.Error()).To(Equal("error listing volumes"))
			Expect(fakeSpectrumDataModel.ListVolumesCallCount()).To(Equal(1))
		})
		It("should succeed to list volumes", func() {

			volume1 := resources.Volume{Name: "fake-volume-1"}
			volume2 := resources.Volume{Name: "fake-volume-2"}
			volumesList := make([]resources.Volume, 2)
			volumesList[0] = volume1
			volumesList[1] = volume2
			fakeSpectrumDataModel.ListVolumesReturns(volumesList, nil)
			volumes, err := client.ListVolumes(listVolumesRequest)
			Expect(err).ToNot(HaveOccurred())
			Expect(len(volumes)).To(Equal(2))
			Expect(fakeSpectrumDataModel.ListVolumesCallCount()).To(Equal(1))
		})

	})

	Context("GetVolume", func() {
		BeforeEach(func() {
			getVolumeRequest = resources.GetVolumeRequest{Name: "fake-volume"}
		})

		It("should fail when dbClient fails to check if the volume exists", func() {
			fakeSpectrumDataModel.GetVolumeReturns(spectrumscale.SpectrumScaleVolume{}, false, fmt.Errorf("error checking volume"))
			_, err = client.GetVolume(getVolumeRequest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("error checking volume"))
			Expect(fakeSpectrumDataModel.GetVolumeCallCount()).To(Equal(1))
		})

		It("should fail when volume exists and dbClient fails to getVolume", func() {
			fakeSpectrumDataModel.GetVolumeReturns(spectrumscale.SpectrumScaleVolume{}, false, fmt.Errorf("error getting volume"))
			_, err = client.GetVolume(getVolumeRequest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("error getting volume"))
			Expect(fakeSpectrumDataModel.GetVolumeCallCount()).To(Equal(1))
		})

		It("should fail when volume does not exist", func() {
			fakeSpectrumDataModel.GetVolumeReturns(spectrumscale.SpectrumScaleVolume{}, false, nil)
			_, err = client.GetVolume(getVolumeRequest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("[fake-volume] "+resources.VolumeNotFoundErrorMsg))
			Expect(fakeSpectrumDataModel.GetVolumeCallCount()).To(Equal(1))
		})

		It("should succeed  when volume exists", func() {
			volume := spectrumscale.SpectrumScaleVolume{Volume: resources.Volume{Name: "fake-volume"}, FileSystem: "fake-filesystem"}
			fakeSpectrumDataModel.GetVolumeReturns(volume, true, nil)
			vol, err := client.GetVolume(getVolumeRequest)
			Expect(err).ToNot(HaveOccurred())
			Expect(vol.Name).To(Equal("fake-volume"))
			Expect(fakeSpectrumDataModel.GetVolumeCallCount()).To(Equal(1))
		})

	})

})
