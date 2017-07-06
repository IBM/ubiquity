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
	"log"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/IBM/ubiquity/local/spectrumscale"

	"github.com/IBM/ubiquity/fakes"
	"github.com/IBM/ubiquity/resources"
)

var _ = Describe("local-client", func() {
	var (
		client                     resources.StorageClient
		logger                     *log.Logger
		fakeSpectrumScaleConnector *fakes.FakeSpectrumScaleConnector
		fakeSpectrumDataModel      *fakes.FakeSpectrumDataModel
		fakeExec                   *fakes.FakeExecutor
		fakeConfig                 resources.SpectrumScaleConfig
		activateRequest            resources.ActivateRequest
		createVolumeRequest        resources.CreateVolumeRequest
		removeVolumeRequest        resources.RemoveVolumeRequest
		attachRequest              resources.AttachRequest
		detachRequest              resources.DetachRequest
		getVolumeRequest           resources.GetVolumeRequest
		backends                   []string
		// getVolumeConfigRequest     resources.GetVolumeConfigRequest
		listVolumesRequest resources.ListVolumesRequest
		err                error
	)
	BeforeEach(func() {
		logger = log.New(os.Stdout, "ubiquity: ", log.Lshortfile|log.LstdFlags)
		fakeSpectrumScaleConnector = new(fakes.FakeSpectrumScaleConnector)
		backends = []string{resources.SpectrumScale}
		fakeExec = new(fakes.FakeExecutor)
		fakeSpectrumDataModel = new(fakes.FakeSpectrumDataModel)
		fakeConfig = resources.SpectrumScaleConfig{}
		activateRequest = resources.ActivateRequest{Backends: backends}
		client, err = spectrumscale.NewSpectrumLocalClientWithConnectors(logger, fakeSpectrumScaleConnector, fakeExec, fakeConfig, fakeSpectrumDataModel)
		Expect(err).ToNot(HaveOccurred())

	})

	Context(".Activate", func() {
		It("should fail when spectrum client IsFilesystemMounted errors", func() {
			fakeSpectrumScaleConnector.IsFilesystemMountedReturns(false, fmt.Errorf("error in isFilesystemMounted"))
			err = client.Activate(activateRequest)
			Expect(err).To(HaveOccurred())
			Expect(fakeSpectrumScaleConnector.IsFilesystemMountedCallCount()).To(Equal(1))
			Expect(fakeSpectrumScaleConnector.MountFileSystemCallCount()).To(Equal(0))
		})

		It("should fail when spectrum client MountFileSystem errors", func() {
			fakeSpectrumScaleConnector.IsFilesystemMountedReturns(false, nil)
			fakeSpectrumScaleConnector.MountFileSystemReturns(fmt.Errorf("error in mount filesystem"))
			err = client.Activate(activateRequest)
			Expect(err).To(HaveOccurred())
			Expect(fakeSpectrumScaleConnector.IsFilesystemMountedCallCount()).To(Equal(1))
			Expect(fakeSpectrumScaleConnector.MountFileSystemCallCount()).To(Equal(1))
			Expect(fakeSpectrumScaleConnector.GetClusterIdCallCount()).To(Equal(0))
		})

		It("should fail when spectrum client GetClusterID errors", func() {
			fakeSpectrumScaleConnector.IsFilesystemMountedReturns(true, nil)
			fakeSpectrumScaleConnector.GetClusterIdReturns("", fmt.Errorf("error getting the cluster ID"))
			err = client.Activate(activateRequest)
			Expect(err).To(HaveOccurred())
			Expect(fakeSpectrumScaleConnector.IsFilesystemMountedCallCount()).To(Equal(1))
			Expect(fakeSpectrumScaleConnector.MountFileSystemCallCount()).To(Equal(0))
			Expect(fakeSpectrumScaleConnector.GetClusterIdCallCount()).To(Equal(1))
		})

		It("should fail when spectrum client GetClusterID return empty ID", func() {
			fakeSpectrumScaleConnector.IsFilesystemMountedReturns(true, nil)
			fakeSpectrumScaleConnector.GetClusterIdReturns("", nil)
			err = client.Activate(activateRequest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Unable to retrieve clusterId: clusterId is empty"))
			Expect(fakeSpectrumScaleConnector.IsFilesystemMountedCallCount()).To(Equal(1))
			Expect(fakeSpectrumScaleConnector.MountFileSystemCallCount()).To(Equal(0))
			Expect(fakeSpectrumScaleConnector.GetClusterIdCallCount()).To(Equal(1))
		})

		It("should succeed when everything is fine with no mounting", func() {
			fakeSpectrumScaleConnector.IsFilesystemMountedReturns(true, nil)
			fakeSpectrumScaleConnector.GetClusterIdReturns("fake-cluster", nil)
			fakeSpectrumDataModel.CreateVolumeTableReturns(nil)
			err = client.Activate(activateRequest)
			Expect(err).ToNot(HaveOccurred())
			Expect(fakeSpectrumScaleConnector.IsFilesystemMountedCallCount()).To(Equal(1))
			Expect(fakeSpectrumScaleConnector.MountFileSystemCallCount()).To(Equal(0))
			Expect(fakeSpectrumScaleConnector.GetClusterIdCallCount()).To(Equal(1))
			Expect(fakeSpectrumDataModel.CreateVolumeTableCallCount()).To(Equal(1))
		})

		It("should succeed when everything is fine with mounting", func() {
			fakeSpectrumScaleConnector.IsFilesystemMountedReturns(false, nil)
			fakeSpectrumScaleConnector.MountFileSystemReturns(nil)
			fakeSpectrumScaleConnector.GetClusterIdReturns("fake-cluster", nil)
			fakeSpectrumDataModel.CreateVolumeTableReturns(nil)
			err = client.Activate(activateRequest)
			Expect(err).ToNot(HaveOccurred())
			Expect(fakeSpectrumScaleConnector.IsFilesystemMountedCallCount()).To(Equal(1))
			Expect(fakeSpectrumScaleConnector.MountFileSystemCallCount()).To(Equal(1))
			Expect(fakeSpectrumScaleConnector.GetClusterIdCallCount()).To(Equal(1))
			Expect(fakeSpectrumDataModel.CreateVolumeTableCallCount()).To(Equal(1))
		})
		It("should succeed on subsequent activate without duplicating execution", func() {
			fakeSpectrumScaleConnector.IsFilesystemMountedReturns(false, nil)
			fakeSpectrumScaleConnector.MountFileSystemReturns(nil)
			fakeSpectrumScaleConnector.GetClusterIdReturns("fake-cluster", nil)
			fakeSpectrumDataModel.CreateVolumeTableReturns(nil)
			err = client.Activate(activateRequest)
			Expect(err).ToNot(HaveOccurred())
			Expect(fakeSpectrumScaleConnector.IsFilesystemMountedCallCount()).To(Equal(1))
			Expect(fakeSpectrumScaleConnector.MountFileSystemCallCount()).To(Equal(1))
			Expect(fakeSpectrumScaleConnector.GetClusterIdCallCount()).To(Equal(1))
			Expect(fakeSpectrumDataModel.CreateVolumeTableCallCount()).To(Equal(1))

			err = client.Activate(activateRequest)
			Expect(err).ToNot(HaveOccurred())
			Expect(fakeSpectrumScaleConnector.IsFilesystemMountedCallCount()).To(Equal(1))
			Expect(fakeSpectrumScaleConnector.MountFileSystemCallCount()).To(Equal(1))
			Expect(fakeSpectrumScaleConnector.GetClusterIdCallCount()).To(Equal(1))

		})

	})

	Context(".CreateVolume", func() {
		var (
			opts map[string]interface{}
		)
		BeforeEach(func() {
			client, err = spectrumscale.NewSpectrumLocalClientWithConnectors(logger, fakeSpectrumScaleConnector, fakeExec, fakeConfig, fakeSpectrumDataModel)
			Expect(err).ToNot(HaveOccurred())
			fakeSpectrumScaleConnector.IsFilesystemMountedReturns(false, nil)
			fakeSpectrumScaleConnector.MountFileSystemReturns(nil)
			fakeSpectrumScaleConnector.GetClusterIdReturns("fake-cluster", nil)
			fakeSpectrumDataModel.CreateVolumeTableReturns(nil)
			err = client.Activate(activateRequest)
			Expect(err).ToNot(HaveOccurred())

			createVolumeRequest = resources.CreateVolumeRequest{Name: "fake-volume", Opts: opts}
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
				err = client.CreateVolume(createVolumeRequest)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("error creating fileset"))
				Expect(fakeSpectrumScaleConnector.CreateFilesetCallCount()).To(Equal(1))
				Expect(fakeSpectrumDataModel.InsertFilesetVolumeCallCount()).To(Equal(0))
			})

			It("should fail when dbClient fails to insert fileset record", func() {
				fakeSpectrumScaleConnector.CreateFilesetReturns(nil)
				fakeSpectrumDataModel.InsertFilesetVolumeReturns(fmt.Errorf("error inserting fileset"))

				err = client.CreateVolume(createVolumeRequest)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("error inserting fileset"))
				Expect(fakeSpectrumScaleConnector.CreateFilesetCallCount()).To(Equal(1))
				Expect(fakeSpectrumDataModel.InsertFilesetVolumeCallCount()).To(Equal(1))
			})

			It("should succeed to create fileset", func() {
				fakeSpectrumScaleConnector.CreateFilesetReturns(nil)
				fakeSpectrumDataModel.InsertFilesetVolumeReturns(nil)

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
					opts["quota"] = "1Gi"
					createVolumeRequest = resources.CreateVolumeRequest{Name: "fake-fileset", Opts: opts}
				})
				It("should fail when spectrum client fails to list fileset quota", func() {
					fakeSpectrumScaleConnector.ListFilesetQuotaReturns("", fmt.Errorf("error in list quota"))
					err = client.CreateVolume(createVolumeRequest)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("error in list quota"))
					Expect(fakeSpectrumDataModel.InsertFilesetQuotaVolumeCallCount()).To(Equal(0))
				})
				It("should fail when spectrum client returns a missmatching fileset quota", func() {
					fakeSpectrumScaleConnector.ListFilesetQuotaReturns("2Gi", nil)
					err = client.CreateVolume(createVolumeRequest)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("Mismatch between user-specified and listed quota for fileset fake-fileset"))
					Expect(fakeSpectrumDataModel.InsertFilesetQuotaVolumeCallCount()).To(Equal(0))
				})
				It("should fail when dbClient fails to insert Fileset quota volume", func() {
					fakeSpectrumScaleConnector.ListFilesetQuotaReturns("1Gi", nil)
					fakeSpectrumDataModel.InsertFilesetQuotaVolumeReturns(fmt.Errorf("error inserting filesetquotavolume"))
					err = client.CreateVolume(createVolumeRequest)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("error inserting filesetquotavolume"))
					Expect(fakeSpectrumDataModel.InsertFilesetQuotaVolumeCallCount()).To(Equal(1))
				})
				It("should succeed when the options are well specified", func() {
					fakeSpectrumScaleConnector.ListFilesetQuotaReturns("1Gi", nil)
					fakeSpectrumDataModel.InsertFilesetQuotaVolumeReturns(nil)
					err = client.CreateVolume(createVolumeRequest)
					Expect(err).ToNot(HaveOccurred())
					Expect(fakeSpectrumScaleConnector.ListFilesetQuotaCallCount()).To(Equal(1))
					Expect(fakeSpectrumDataModel.InsertFilesetQuotaVolumeCallCount()).To(Equal(1))
				})

			})
		})

		Context(".LightWeightVolume", func() {
			BeforeEach(func() {
				opts = make(map[string]interface{})
				opts["fileset"] = "fake-fileset"
				opts["filesystem"] = "fake-filesystem"
				opts["type"] = "lightweight"

				createVolumeRequest = resources.CreateVolumeRequest{Name: "fake-lightweight", Opts: opts}
			})
			It("should fail when spectrum client IsfilesetLinked errors", func() {
				fakeSpectrumScaleConnector.IsFilesetLinkedReturns(false, fmt.Errorf("error in checking fileset linked"))
				err = client.CreateVolume(createVolumeRequest)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("error in checking fileset linked"))
				Expect(fakeSpectrumScaleConnector.IsFilesetLinkedCallCount()).To(Equal(1))
				Expect(fakeSpectrumScaleConnector.LinkFilesetCallCount()).To(Equal(0))
			})
			It("should fail when spectrum client LinkFileset errors", func() {
				fakeSpectrumScaleConnector.IsFilesetLinkedReturns(false, nil)
				fakeSpectrumScaleConnector.LinkFilesetReturns(fmt.Errorf("error linking fileset"))
				err = client.CreateVolume(createVolumeRequest)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("error linking fileset"))
				Expect(fakeSpectrumScaleConnector.IsFilesetLinkedCallCount()).To(Equal(1))
				Expect(fakeSpectrumScaleConnector.LinkFilesetCallCount()).To(Equal(1))
			})

			It("should fail when spectrum client GetFilesystemMountpoint errors", func() {
				fakeSpectrumScaleConnector.IsFilesetLinkedReturns(true, nil)
				fakeSpectrumScaleConnector.GetFilesystemMountpointReturns("", fmt.Errorf("error getting mountpoint"))
				err = client.CreateVolume(createVolumeRequest)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("error getting mountpoint"))
				Expect(fakeSpectrumScaleConnector.IsFilesetLinkedCallCount()).To(Equal(1))
				Expect(fakeSpectrumScaleConnector.LinkFilesetCallCount()).To(Equal(0))
				Expect(fakeSpectrumScaleConnector.GetFilesystemMountpointCallCount()).To(Equal(1))
			})

			It("should fail when spectrum client GetFilesystemMountpoint errors", func() {
				fakeSpectrumScaleConnector.IsFilesetLinkedReturns(true, nil)
				fakeSpectrumScaleConnector.GetFilesystemMountpointReturns("fake-mountpoint", nil)
				fakeExec.StatReturns(nil, fmt.Errorf("error in os.Stat"))
				fakeExec.ExecuteReturns(nil, fmt.Errorf("error in mkdir"))
				//fakeExec.MkdirReturns(fmt.Errorf("error in mkdir"))
				err = client.CreateVolume(createVolumeRequest)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("error in mkdir"))
				Expect(fakeSpectrumScaleConnector.IsFilesetLinkedCallCount()).To(Equal(1))
				Expect(fakeSpectrumScaleConnector.LinkFilesetCallCount()).To(Equal(0))
				Expect(fakeSpectrumScaleConnector.GetFilesystemMountpointCallCount()).To(Equal(1))
				// Expect(fakeExec.StatCallCount()).To(Equal(1))
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
			Expect(err.Error()).To(Equal("Volume not found"))
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

		It("should fail when type is lightweight and dbClient fails to delete the volume", func() {
			volume := spectrumscale.SpectrumScaleVolume{Volume: resources.Volume{Name: "fake-volume"}, Type: 1}
			fakeSpectrumDataModel.GetVolumeReturns(volume, true, nil)
			fakeSpectrumDataModel.DeleteVolumeReturns(fmt.Errorf("error deleting volume"))
			err = client.RemoveVolume(removeVolumeRequest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("error deleting volume"))
			Expect(fakeSpectrumDataModel.GetVolumeCallCount()).To(Equal(1))
			Expect(fakeSpectrumDataModel.DeleteVolumeCallCount()).To(Equal(1))
			Expect(fakeSpectrumScaleConnector.GetFilesystemMountpointCallCount()).To(Equal(0))
		})
		Context("When forcedelete is set to true", func() {
			BeforeEach(func() {
				fakeConfig = resources.SpectrumScaleConfig{ForceDelete: true}
				client, err = spectrumscale.NewSpectrumLocalClientWithConnectors(logger, fakeSpectrumScaleConnector, fakeExec, fakeConfig, fakeSpectrumDataModel)
				Expect(err).ToNot(HaveOccurred())

			})
			It("should fail when type is lightweight and forcedelete is true and spectrumClient fails to get filesystem mountpoint", func() {
				volume := spectrumscale.SpectrumScaleVolume{Volume: resources.Volume{Name: "fake-volume"}, Type: 1}
				fakeSpectrumDataModel.GetVolumeReturns(volume, true, nil)
				fakeSpectrumDataModel.DeleteVolumeReturns(nil)
				fakeSpectrumScaleConnector.GetFilesystemMountpointReturns("", fmt.Errorf("error getting fs mountpoint"))
				err = client.RemoveVolume(removeVolumeRequest)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("error getting fs mountpoint"))
				Expect(fakeSpectrumDataModel.GetVolumeCallCount()).To(Equal(1))
				Expect(fakeSpectrumDataModel.DeleteVolumeCallCount()).To(Equal(1))
				Expect(fakeSpectrumScaleConnector.GetFilesystemMountpointCallCount()).To(Equal(1))
				Expect(fakeExec.RemoveAllCallCount()).To(Equal(0))
			})

			It("should fail when type is lightweight and forcedelete is true and executor fails to remove volume folder", func() {
				volume := spectrumscale.SpectrumScaleVolume{Volume: resources.Volume{Name: "fake-volume"}, Type: 1}
				fakeSpectrumDataModel.GetVolumeReturns(volume, true, nil)
				fakeSpectrumDataModel.DeleteVolumeReturns(nil)
				fakeSpectrumScaleConnector.GetFilesystemMountpointReturns("fake-mountpoint", nil)
				fakeExec.RemoveAllReturns(fmt.Errorf("error removing path"))
				err = client.RemoveVolume(removeVolumeRequest)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("error removing path"))
				Expect(fakeSpectrumDataModel.GetVolumeCallCount()).To(Equal(1))
				Expect(fakeSpectrumDataModel.DeleteVolumeCallCount()).To(Equal(1))
				Expect(fakeSpectrumScaleConnector.GetFilesystemMountpointCallCount()).To(Equal(1))
				Expect(fakeExec.RemoveAllCallCount()).To(Equal(1))
			})

			It("should succeed when type is lightweight and forcedelete is true", func() {
				volume := spectrumscale.SpectrumScaleVolume{Volume: resources.Volume{Name: "fake-volume"}, Type: 1}
				fakeSpectrumDataModel.GetVolumeReturns(volume, true, nil)
				fakeSpectrumDataModel.DeleteVolumeReturns(nil)
				fakeSpectrumScaleConnector.GetFilesystemMountpointReturns("fake-mountpoint", nil)
				fakeExec.RemoveAllReturns(nil)
				err = client.RemoveVolume(removeVolumeRequest)
				Expect(err).ToNot(HaveOccurred())
				Expect(fakeSpectrumDataModel.GetVolumeCallCount()).To(Equal(1))
				Expect(fakeSpectrumDataModel.DeleteVolumeCallCount()).To(Equal(1))
				Expect(fakeSpectrumScaleConnector.GetFilesystemMountpointCallCount()).To(Equal(1))
				Expect(fakeExec.RemoveAllCallCount()).To(Equal(1))
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
		It("should succeed when type is lightweight and forcedelete is false", func() {
			volume := spectrumscale.SpectrumScaleVolume{Volume: resources.Volume{Name: "fake-volume"}, Type: 1}
			fakeSpectrumDataModel.GetVolumeReturns(volume, true, nil)
			fakeSpectrumDataModel.DeleteVolumeReturns(nil)
			err = client.RemoveVolume(removeVolumeRequest)
			Expect(err).ToNot(HaveOccurred())
			Expect(fakeSpectrumDataModel.GetVolumeCallCount()).To(Equal(1))
			Expect(fakeSpectrumDataModel.DeleteVolumeCallCount()).To(Equal(1))
			Expect(fakeSpectrumScaleConnector.GetFilesystemMountpointCallCount()).To(Equal(0))
			Expect(fakeExec.RemoveAllCallCount()).To(Equal(0))
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
			Expect(err.Error()).To(Equal("Volume not found"))
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

	Context(".Attach", func() {
		BeforeEach(func() {
			attachRequest = resources.AttachRequest{Name: "fake-volume"}
		})

		It("should fail when dbClient fails to check volumeExists", func() {
			fakeSpectrumDataModel.GetVolumeReturns(spectrumscale.SpectrumScaleVolume{}, false, fmt.Errorf("error in checking volume"))
			mountpath, err := client.Attach(attachRequest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("error in checking volume"))
			Expect(mountpath).To(Equal(""))
			Expect(fakeSpectrumDataModel.GetVolumeCallCount()).To(Equal(1))
		})

		It("should fail when volume does not exist", func() {
			fakeSpectrumDataModel.GetVolumeReturns(spectrumscale.SpectrumScaleVolume{}, false, nil)
			mountpath, err := client.Attach(attachRequest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Volume not found"))
			Expect(mountpath).To(Equal(""))
			Expect(fakeSpectrumDataModel.GetVolumeCallCount()).To(Equal(1))
		})

		It("should fail when volume exists and dbClient fails to getVolume", func() {
			fakeSpectrumDataModel.GetVolumeReturns(spectrumscale.SpectrumScaleVolume{}, false, fmt.Errorf("error getting volume"))
			mountpath, err := client.Attach(attachRequest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("error getting volume"))
			Expect(mountpath).To(Equal(""))
			Expect(fakeSpectrumDataModel.GetVolumeCallCount()).To(Equal(1))
			Expect(fakeSpectrumScaleConnector.GetFilesystemMountpointCallCount()).To(Equal(0))
		})

		It("should fail when volume is not attached and dbClient fails to get filesystem mountpoint", func() {
			volume := spectrumscale.SpectrumScaleVolume{Volume: resources.Volume{Name: "fake-volume"}, FileSystem: "fake-filesystem"}
			fakeSpectrumDataModel.GetVolumeReturns(volume, true, nil)
			fakeSpectrumScaleConnector.GetFilesystemMountpointReturns("", fmt.Errorf("error getting mountpoint"))
			mountpath, err := client.Attach(attachRequest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("error getting mountpoint"))
			Expect(mountpath).To(Equal(""))
			Expect(fakeSpectrumDataModel.GetVolumeCallCount()).To(Equal(1))
			Expect(fakeSpectrumScaleConnector.GetFilesystemMountpointCallCount()).To(Equal(1))
			Expect(fakeSpectrumScaleConnector.IsFilesetLinkedCallCount()).To(Equal(0))
		})

		It("should fail when volume is fileset volume and spectrumClient fails to check fileset linked", func() {
			volume := spectrumscale.SpectrumScaleVolume{Volume: resources.Volume{Name: "fake-volume"}, FileSystem: "fake-filesystem", Type: spectrumscale.Fileset}
			fakeSpectrumDataModel.GetVolumeReturns(volume, true, nil)
			fakeSpectrumScaleConnector.GetFilesystemMountpointReturns("fake-mountpoint", nil)
			fakeSpectrumScaleConnector.IsFilesetLinkedReturns(false, fmt.Errorf("error checking filesetlinked"))
			mountpath, err := client.Attach(attachRequest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("error checking filesetlinked"))
			Expect(mountpath).To(Equal(""))
			Expect(fakeSpectrumDataModel.GetVolumeCallCount()).To(Equal(1))
			Expect(fakeSpectrumScaleConnector.GetFilesystemMountpointCallCount()).To(Equal(1))
			Expect(fakeSpectrumScaleConnector.IsFilesetLinkedCallCount()).To(Equal(1))
			Expect(fakeSpectrumScaleConnector.LinkFilesetCallCount()).To(Equal(0))
		})

		It("should fail when volume is fileset volume and spectrumClient fails to link it", func() {
			volume := spectrumscale.SpectrumScaleVolume{Volume: resources.Volume{Name: "fake-volume"}, FileSystem: "fake-filesystem", Type: spectrumscale.Fileset}
			fakeSpectrumDataModel.GetVolumeReturns(volume, true, nil)
			fakeSpectrumScaleConnector.GetFilesystemMountpointReturns("fake-mountpoint", nil)
			fakeSpectrumScaleConnector.IsFilesetLinkedReturns(false, nil)
			fakeSpectrumScaleConnector.LinkFilesetReturns(fmt.Errorf("error linking fileset"))
			mountpath, err := client.Attach(attachRequest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("error linking fileset"))
			Expect(mountpath).To(Equal(""))
			Expect(fakeSpectrumDataModel.GetVolumeCallCount()).To(Equal(1))
			Expect(fakeSpectrumScaleConnector.GetFilesystemMountpointCallCount()).To(Equal(1))
			Expect(fakeSpectrumScaleConnector.IsFilesetLinkedCallCount()).To(Equal(1))
			Expect(fakeSpectrumScaleConnector.LinkFilesetCallCount()).To(Equal(1))
		})

		It("should succeed when volume is lightweight volume with permissions", func() {
			volume := spectrumscale.SpectrumScaleVolume{Volume: resources.Volume{Name: "fake-volume"}, FileSystem: "fake-filesystem", Type: spectrumscale.Lightweight, UID: "fake-uid", GID: "gid"}
			fakeSpectrumDataModel.GetVolumeReturns(volume, true, nil)
			fakeSpectrumScaleConnector.GetFilesystemMountpointReturns("fake-mountpoint", nil)
			fakeSpectrumScaleConnector.IsFilesetLinkedReturns(false, nil)
			fakeSpectrumScaleConnector.LinkFilesetReturns(nil)
			fakeExec.ExecuteReturns(nil, nil)
			mountpath, err := client.Attach(attachRequest)
			Expect(err).ToNot(HaveOccurred())
			Expect(mountpath).To(Equal("fake-mountpoint"))
			Expect(fakeSpectrumDataModel.GetVolumeCallCount()).To(Equal(1))
			Expect(fakeSpectrumScaleConnector.GetFilesystemMountpointCallCount()).To(Equal(1))
			Expect(fakeSpectrumScaleConnector.IsFilesetLinkedCallCount()).To(Equal(1))
			Expect(fakeSpectrumScaleConnector.LinkFilesetCallCount()).To(Equal(1))
		})

		It("should succeed when volume is fileset volume with permissions", func() {
			volume := spectrumscale.SpectrumScaleVolume{Volume: resources.Volume{Name: "fake-volume"}, FileSystem: "fake-filesystem", Type: spectrumscale.Fileset, UID: "fake-uid", GID: "gid"}
			fakeSpectrumDataModel.GetVolumeReturns(volume, true, nil)
			fakeSpectrumScaleConnector.GetFilesystemMountpointReturns("fake-mountpoint", nil)
			fakeSpectrumScaleConnector.IsFilesetLinkedReturns(false, nil)
			fakeSpectrumScaleConnector.LinkFilesetReturns(nil)
			fakeExec.ExecuteReturns(nil, nil)
			mountpath, err := client.Attach(attachRequest)
			Expect(err).ToNot(HaveOccurred())
			Expect(mountpath).To(Equal("fake-mountpoint"))
			Expect(fakeSpectrumDataModel.GetVolumeCallCount()).To(Equal(1))
			Expect(fakeSpectrumScaleConnector.GetFilesystemMountpointCallCount()).To(Equal(1))
			Expect(fakeSpectrumScaleConnector.IsFilesetLinkedCallCount()).To(Equal(1))
			Expect(fakeSpectrumScaleConnector.LinkFilesetCallCount()).To(Equal(1))
		})

	})

	Context(".Detach", func() {
		BeforeEach(func() {
			detachRequest = resources.DetachRequest{Name: "fake-volume"}
		})

		It("should fail when dbClient fails to check volumeExists", func() {
			fakeSpectrumDataModel.GetVolumeReturns(spectrumscale.SpectrumScaleVolume{}, false, fmt.Errorf("error in checking volume"))
			err = client.Detach(detachRequest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("error in checking volume"))
			Expect(fakeSpectrumDataModel.GetVolumeCallCount()).To(Equal(1))
		})

		It("should fail when volume does not exist", func() {
			fakeSpectrumDataModel.GetVolumeReturns(spectrumscale.SpectrumScaleVolume{}, false, nil)
			err = client.Detach(detachRequest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Volume not found"))
			Expect(fakeSpectrumDataModel.GetVolumeCallCount()).To(Equal(1))
		})

		It("should fail when volume exists and dbClient fails to getVolume", func() {
			fakeSpectrumDataModel.GetVolumeReturns(spectrumscale.SpectrumScaleVolume{}, false, fmt.Errorf("error getting volume"))
			err = client.Detach(detachRequest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("error getting volume"))
			Expect(fakeSpectrumDataModel.GetVolumeCallCount()).To(Equal(1))
		})

		It("should fail when volume exists but not attached", func() {
			volume := spectrumscale.SpectrumScaleVolume{Volume: resources.Volume{Name: "fake-volume"}, FileSystem: "fake-filesystem"}
			fakeSpectrumDataModel.GetVolumeReturns(volume, true, nil)
			err := client.Detach(detachRequest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("volume not attached"))
			Expect(fakeSpectrumDataModel.GetVolumeCallCount()).To(Equal(1))
		})

		It("should succeed when everything is all right", func() {
			fakeSpectrumScaleConnector.IsFilesystemMountedReturns(true, nil)
			fakeSpectrumScaleConnector.IsFilesetLinkedReturns(true, nil)

			volume := spectrumscale.SpectrumScaleVolume{Volume: resources.Volume{Name: "fake-volume"}, FileSystem: "fake-filesystem"}
			fakeSpectrumDataModel.GetVolumeReturns(volume, true, nil)
			err := client.Detach(detachRequest)
			Expect(err).ToNot(HaveOccurred())
			Expect(fakeSpectrumDataModel.GetVolumeCallCount()).To(Equal(1))
		})

	})

})
