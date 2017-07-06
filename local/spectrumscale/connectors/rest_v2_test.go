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

package connectors_test

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/IBM/ubiquity/local/spectrumscale/connectors"
	"github.com/IBM/ubiquity/resources"
	"github.com/jarcoal/httpmock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("spectrumRestV2", func() {
	var (
		spectrumRestV2 connectors.SpectrumScaleConnector
		logger         *log.Logger
		fakeurl        string
		restConfig     resources.RestConfig
		opts           map[string]interface{}
		err            error
		fileset        string
		filesystem     string
		client         *http.Client
	)

	BeforeEach(func() {
		logger = log.New(os.Stdout, "spectrum: ", log.Lshortfile|log.LstdFlags)
		httpmock.Activate()
		fakeurl = "http://1.1.1.1:443"
		restConfig.Endpoint = fakeurl
		restConfig.User = "fakeuser"
		restConfig.Password = "fakepassword"
		restConfig.Hostname = "fakehostname"
		spectrumRestV2, client, err = connectors.NewspectrumRestV2WithClient(logger, restConfig)
		Expect(err).ToNot(HaveOccurred())
		httpmock.ActivateNonDefault(client)
		fileset = "fake-fileset"
		filesystem = "fake-filesystem"
	})

	Context(".GetClusterID", func() {
		It("Should Pass and Cluster ID should be valid", func() {
			getClusterRespo := connectors.GetClusterResponse{}
			getClusterRespo.Status.Code = 200
			getClusterRespo.Status.Message = "passed to get cluster id"
			getClusterRespo.Cluster.ClusterSummary.ClusterID = 12345
			marshalledResponse, err := json.Marshal(getClusterRespo)
			Expect(err).ToNot(HaveOccurred())
			registerurl := fakeurl + "/scalemgmt/v2/cluster"
			httpmock.RegisterResponder(
				"GET",
				registerurl,
				httpmock.NewStringResponder(200, string(marshalledResponse)),
			)
			str, err := spectrumRestV2.GetClusterId()
			Expect(str).To(Equal("12345"))
		})

		It("Should fail with blank Clusterid", func() {
			getClusterRespo := connectors.GetClusterResponse{}
			getClusterRespo.Status.Code = 500
			getClusterRespo.Status.Message = "passed to get cluster id"
			marshalledResponse, err := json.Marshal(getClusterRespo)
			Expect(err).ToNot(HaveOccurred())
			registerurl := fakeurl + "/scalemgmt/v2/cluster"
			httpmock.RegisterResponder(
				"GET",
				registerurl,
				httpmock.NewStringResponder(500, string(marshalledResponse)),
			)
			str, err := spectrumRestV2.GetClusterId()
			Expect(err).To(HaveOccurred())
			Expect(str).To(Equal(""))
		})
	})

	Context(".IsFilesystemMounted", func() {
		It("Should pass and return true", func() {
			getnodeResp := connectors.GetNodesResponse_v2{}
			getnodeResp.Nodes = make([]connectors.Node_v2, 1)
			getnodeResp.Nodes[0].AdminNodename = "fakehostname"
			getnodeResp.Status.Code = 200
			marshalledResponse, err := json.Marshal(getnodeResp)
			Expect(err).ToNot(HaveOccurred())
			registerurl := fakeurl + "/scalemgmt/v2/nodes"
			httpmock.RegisterResponder(
				"GET",
				registerurl,
				httpmock.NewStringResponder(200, string(marshalledResponse)),
			)
			ismounted, err := spectrumRestV2.IsFilesystemMounted(filesystem)
			Expect(err).ToNot(HaveOccurred())
			Expect(ismounted).To(Equal(true))
		})

		It("Should pass but return false as node name is not matching", func() {
			getnodeResp := connectors.GetNodesResponse_v2{}
			getnodeResp.Nodes = make([]connectors.Node_v2, 1)
			getnodeResp.Nodes[0].AdminNodename = "fakehostname1"
			getnodeResp.Status.Code = 200
			marshalledResponse, err := json.Marshal(getnodeResp)
			Expect(err).ToNot(HaveOccurred())
			registerurl := fakeurl + "/scalemgmt/v2/nodes"
			httpmock.RegisterResponder(
				"GET",
				registerurl,
				httpmock.NewStringResponder(200, string(marshalledResponse)),
			)
			ismounted, err := spectrumRestV2.IsFilesystemMounted(filesystem)
			Expect(err).ToNot(HaveOccurred())
			Expect(ismounted).To(Equal(false))
		})

		It("Should fail with http error and return false", func() {
			getnodeResp := connectors.GetNodesResponse_v2{}
			getnodeResp.Nodes = make([]connectors.Node_v2, 1)
			getnodeResp.Nodes[0].AdminNodename = "fakehostname"
			getnodeResp.Status.Code = 500
			marshalledResponse, err := json.Marshal(getnodeResp)
			Expect(err).ToNot(HaveOccurred())
			registerurl := fakeurl + "/scalemgmt/v2/nodes"
			httpmock.RegisterResponder(
				"GET",
				registerurl,
				httpmock.NewStringResponder(500, string(marshalledResponse)),
			)
			ismounted, err := spectrumRestV2.IsFilesystemMounted(filesystem)
			Expect(err).To(HaveOccurred())
			Expect(ismounted).To(Equal(false))
		})

		It("Should pass by getting hostname from system ", func() {
			getnodeResp := connectors.GetNodesResponse_v2{}
			getnodeResp.Nodes = make([]connectors.Node_v2, 1)
			hostn, err := os.Hostname()
			Expect(err).ToNot(HaveOccurred())
			getnodeResp.Nodes[0].AdminNodename = hostn
			getnodeResp.Status.Code = 200
			marshalledResponse, err := json.Marshal(getnodeResp)
			Expect(err).ToNot(HaveOccurred())
			registerurl := fakeurl + "/scalemgmt/v2/nodes"
			httpmock.RegisterResponder(
				"GET",
				registerurl,
				httpmock.NewStringResponder(200, string(marshalledResponse)),
			)

			restConfig.Hostname = ""

			nspectrumRestV2, nclient, err := connectors.NewspectrumRestV2WithClient(logger, restConfig)
			Expect(err).ToNot(HaveOccurred())
			httpmock.ActivateNonDefault(nclient)

			ismounted, err := nspectrumRestV2.IsFilesystemMounted(filesystem)
			Expect(err).ToNot(HaveOccurred())
			Expect(ismounted).To(Equal(true))
		})
	})

	Context(".ListFilesystems", func() {
		It("Should pass by getting proper filesystem", func() {
			getfilesysResp := connectors.GetFilesystemResponse_v2{}
			getfilesysResp.FileSystems = make([]connectors.FileSystem_v2, 1)
			getfilesysResp.FileSystems[0].Name = filesystem
			getfilesysResp.Status.Code = 200
			marshalledResponse, err := json.Marshal(getfilesysResp)
			Expect(err).ToNot(HaveOccurred())
			registerurl := fakeurl + "/scalemgmt/v2/filesystems"
			httpmock.RegisterResponder(
				"GET",
				registerurl,
				httpmock.NewStringResponder(200, string(marshalledResponse)),
			)
			arr_str, err := spectrumRestV2.ListFilesystems()

			Expect(err).ToNot(HaveOccurred())
			Expect(len(arr_str)).To(Equal(1))
			Expect(arr_str[0]).To(Equal(filesystem))
		})

		It("Should fail due to http error", func() {
			getfilesysResp := connectors.GetFilesystemResponse_v2{}
			getfilesysResp.FileSystems = make([]connectors.FileSystem_v2, 1)
			getfilesysResp.FileSystems[0].Name = filesystem
			getfilesysResp.Status.Code = 400
			marshalledResponse, err := json.Marshal(getfilesysResp)
			Expect(err).ToNot(HaveOccurred())
			registerurl := fakeurl + "/scalemgmt/v2/filesystems"
			httpmock.RegisterResponder(
				"GET",
				registerurl,
				httpmock.NewStringResponder(400, string(marshalledResponse)),
			)
			arr_str, err := spectrumRestV2.ListFilesystems()
			Expect(err).To(HaveOccurred())
			Expect(len(arr_str)).To(Equal(0))
		})
	})

	Context(".GetFilesystemMountpoint", func() {
		It("Should pass by getting mount name", func() {
			getfilesysResp := connectors.GetFilesystemResponse_v2{}
			getfilesysResp.FileSystems = make([]connectors.FileSystem_v2, 1)
			getfilesysResp.FileSystems[0].Mount.MountPoint = "fakemount"
			getfilesysResp.Status.Code = 200
			marshalledResponse, err := json.Marshal(getfilesysResp)
			Expect(err).ToNot(HaveOccurred())
			registerurl := fakeurl + "/scalemgmt/v2/filesystems/" + filesystem
			httpmock.RegisterResponder(
				"GET",
				registerurl,
				httpmock.NewStringResponder(200, string(marshalledResponse)),
			)
			arr_str, err := spectrumRestV2.GetFilesystemMountpoint(filesystem)

			Expect(err).ToNot(HaveOccurred())
			Expect(arr_str).To(Equal("fakemount"))
		})

		It("Should fail with http error", func() {
			getfilesysResp := connectors.GetFilesystemResponse_v2{}
			getfilesysResp.FileSystems = make([]connectors.FileSystem_v2, 1)
			getfilesysResp.FileSystems[0].Mount.MountPoint = "fakemount"
			getfilesysResp.Status.Code = 500
			marshalledResponse, err := json.Marshal(getfilesysResp)
			Expect(err).ToNot(HaveOccurred())
			registerurl := fakeurl + "/scalemgmt/v2/filesystems/" + filesystem
			httpmock.RegisterResponder(
				"GET",
				registerurl,
				httpmock.NewStringResponder(500, string(marshalledResponse)),
			)
			arr_str, err := spectrumRestV2.GetFilesystemMountpoint(filesystem)

			Expect(err).To(HaveOccurred())
			Expect(arr_str).To(Equal(""))
		})

		It("Should fail since we were unable to get filesystem", func() {
			getfilesysResp := connectors.GetFilesystemResponse_v2{}
			getfilesysResp.FileSystems = make([]connectors.FileSystem_v2, 0)
			getfilesysResp.Status.Code = 200
			marshalledResponse, err := json.Marshal(getfilesysResp)
			Expect(err).ToNot(HaveOccurred())
			registerurl := fakeurl + "/scalemgmt/v2/filesystems/" + filesystem
			httpmock.RegisterResponder(
				"GET",
				registerurl,
				httpmock.NewStringResponder(200, string(marshalledResponse)),
			)
			arr_str, err := spectrumRestV2.GetFilesystemMountpoint(filesystem)

			Expect(err).To(HaveOccurred())
			Expect(arr_str).To(Equal(""))
		})
	})

	Context(".CreateFileset", func() {
		It("Should pass while creating a fileset", func() {
			Skip("TODO need to fix")
			createFilesetResp := connectors.GenericResponse{}
			createFilesetResp.Status.Code = 202
			createFilesetResp.Jobs = make([]connectors.Job, 1)
			createFilesetResp.Jobs[0].JobID = 1234
			createFilesetResp.Jobs[0].Status = "COMPLETED"
			marshalledResponse, err := json.Marshal(createFilesetResp)
			Expect(err).ToNot(HaveOccurred())
			registerurl := fakeurl + "/scalemgmt/v2/filesystems/" + filesystem + "/filesets"
			joburl := fakeurl + "/scalemgmt/v2/jobs?filter=jobId=1234"

			httpmock.RegisterResponder(
				"POST",
				registerurl,
				httpmock.NewStringResponder(202, string(marshalledResponse)),
			)

			httpmock.RegisterResponder(
				"GET",
				joburl,
				httpmock.NewStringResponder(200, string(marshalledResponse)),
			)
			err = spectrumRestV2.CreateFileset(filesystem, fileset, opts)
			Expect(err).ToNot(HaveOccurred())
		})

		It("Should fail with http error", func() {
			createFilesetResp := connectors.GenericResponse{}
			createFilesetResp.Status.Code = 500
			createFilesetResp.Jobs = make([]connectors.Job, 1)
			createFilesetResp.Jobs[0].JobID = 1234
			createFilesetResp.Jobs[0].Status = "COMPLETED"
			marshalledResponse, err := json.Marshal(createFilesetResp)
			Expect(err).ToNot(HaveOccurred())
			registerurl := fakeurl + "/scalemgmt/v2/filesystems/" + filesystem + "/filesets"
			joburl := fakeurl + "/scalemgmt/v2/jobs?filter=jobId=1234"

			httpmock.RegisterResponder(
				"POST",
				registerurl,
				httpmock.NewStringResponder(500, string(marshalledResponse)),
			)

			httpmock.RegisterResponder(
				"GET",
				joburl,
				httpmock.NewStringResponder(200, string(marshalledResponse)),
			)
			err = spectrumRestV2.CreateFileset(filesystem, fileset, opts)
			Expect(err).To(HaveOccurred())
		})

		It("Should fail since it is unable to fetch job details due to http error", func() {
			createFilesetResp := connectors.GenericResponse{}
			createFilesetResp.Status.Code = 202
			createFilesetResp.Jobs = make([]connectors.Job, 1)
			createFilesetResp.Jobs[0].JobID = 1234
			createFilesetResp.Jobs[0].Status = "COMPLETED"
			marshalledResponse, err := json.Marshal(createFilesetResp)
			Expect(err).ToNot(HaveOccurred())
			registerurl := fakeurl + "/scalemgmt/v2/filesystems/" + filesystem + "/filesets"
			joburl := fakeurl + "/scalemgmt/v2/jobs?filter=jobId=1234"

			httpmock.RegisterResponder(
				"POST",
				registerurl,
				httpmock.NewStringResponder(202, string(marshalledResponse)),
			)

			httpmock.RegisterResponder(
				"GET",
				joburl,
				httpmock.NewStringResponder(400, string(marshalledResponse)),
			)
			err = spectrumRestV2.CreateFileset(filesystem, fileset, opts)
			Expect(err).To(HaveOccurred())
		})

		It("Should fail to do zero length job array", func() {
			createFilesetResp := connectors.GenericResponse{}
			createFilesetResp.Status.Code = 202
			createFilesetResp.Jobs = make([]connectors.Job, 0)
			marshalledResponse, err := json.Marshal(createFilesetResp)
			Expect(err).ToNot(HaveOccurred())
			registerurl := fakeurl + "/scalemgmt/v2/filesystems/" + filesystem + "/filesets"
			joburl := fakeurl + "/scalemgmt/v2/jobs?filter=jobId=1234"

			httpmock.RegisterResponder(
				"POST",
				registerurl,
				httpmock.NewStringResponder(202, string(marshalledResponse)),
			)

			httpmock.RegisterResponder(
				"GET",
				joburl,
				httpmock.NewStringResponder(200, string(marshalledResponse)),
			)
			err = spectrumRestV2.CreateFileset(filesystem, fileset, opts)
			Expect(err).To(HaveOccurred())

		})
	})

	Context(".DeleteFileset", func() {
		It("Should pass while deleting a fileset", func() {
			Skip("TODO need to fix")
			deleteFilesetResp := connectors.GenericResponse{}
			deleteFilesetResp.Status.Code = 202
			deleteFilesetResp.Jobs = make([]connectors.Job, 1)
			deleteFilesetResp.Jobs[0].JobID = 1234
			deleteFilesetResp.Jobs[0].Status = "COMPLETED"
			marshalledResponse, err := json.Marshal(deleteFilesetResp)
			Expect(err).ToNot(HaveOccurred())
			registerurl := fakeurl + "/scalemgmt/v2/filesystems/" + filesystem + "/filesets/" + fileset
			joburl := fakeurl + "/scalemgmt/v2/jobs?filter=jobId=1234"

			httpmock.RegisterResponder(
				"DELETE",
				registerurl,
				httpmock.NewStringResponder(202, string(marshalledResponse)),
			)

			httpmock.RegisterResponder(
				"GET",
				joburl,
				httpmock.NewStringResponder(200, string(marshalledResponse)),
			)
			err = spectrumRestV2.DeleteFileset(filesystem, fileset)
			Expect(err).ToNot(HaveOccurred())
		})

		It("Should fail with http error", func() {
			deleteFilesetResp := connectors.GenericResponse{}
			deleteFilesetResp.Status.Code = 500
			deleteFilesetResp.Jobs = make([]connectors.Job, 1)
			deleteFilesetResp.Jobs[0].JobID = 1234
			deleteFilesetResp.Jobs[0].Status = "COMPLETED"
			marshalledResponse, err := json.Marshal(deleteFilesetResp)
			Expect(err).ToNot(HaveOccurred())
			registerurl := fakeurl + "/scalemgmt/v2/filesystems/" + filesystem + "/filesets/" + fileset
			joburl := fakeurl + "/scalemgmt/v2/jobs?filter=jobId=1234"

			httpmock.RegisterResponder(
				"DELETE",
				registerurl,
				httpmock.NewStringResponder(500, string(marshalledResponse)),
			)

			httpmock.RegisterResponder(
				"GET",
				joburl,
				httpmock.NewStringResponder(200, string(marshalledResponse)),
			)
			err = spectrumRestV2.DeleteFileset(filesystem, fileset)
			Expect(err).To(HaveOccurred())
		})

		It("Should fail since it is unable to fetch job details due to http error", func() {
			deleteFilesetResp := connectors.GenericResponse{}
			deleteFilesetResp.Status.Code = 202
			deleteFilesetResp.Jobs = make([]connectors.Job, 1)
			deleteFilesetResp.Jobs[0].JobID = 1234
			deleteFilesetResp.Jobs[0].Status = "COMPLETED"
			marshalledResponse, err := json.Marshal(deleteFilesetResp)
			Expect(err).ToNot(HaveOccurred())

			registerurl := fakeurl + "/scalemgmt/v2/filesystems/" + filesystem + "/filesets/" + fileset
			joburl := fakeurl + "/scalemgmt/v2/jobs?filter=jobId=1234"
			httpmock.RegisterResponder(
				"DELETE",
				registerurl,
				httpmock.NewStringResponder(202, string(marshalledResponse)),
			)

			httpmock.RegisterResponder(
				"GET",
				joburl,
				httpmock.NewStringResponder(400, string(marshalledResponse)),
			)
			err = spectrumRestV2.DeleteFileset(filesystem, fileset)
			Expect(err).To(HaveOccurred())
		})

		It("Should fail to do zero length job array", func() {
			deleteFilesetResp := connectors.GenericResponse{}
			deleteFilesetResp.Status.Code = 202
			deleteFilesetResp.Jobs = make([]connectors.Job, 0)
			marshalledResponse, err := json.Marshal(deleteFilesetResp)
			Expect(err).ToNot(HaveOccurred())
			registerurl := fakeurl + "/scalemgmt/v2/filesystems/" + filesystem + "/filesets"
			joburl := fakeurl + "/scalemgmt/v2/jobs?filter=jobId=1234"

			httpmock.RegisterResponder(
				"DELETE",
				registerurl,
				httpmock.NewStringResponder(202, string(marshalledResponse)),
			)

			httpmock.RegisterResponder(
				"GET",
				joburl,
				httpmock.NewStringResponder(200, string(marshalledResponse)),
			)
			err = spectrumRestV2.DeleteFileset(filesystem, fileset)
			Expect(err).To(HaveOccurred())

		})
	})

	Context(".LinkFileset", func() {
		It("should pass while linking a fileset", func() {
			Skip("TODO need to fix")
			linkFilesetResp := connectors.GenericResponse{}
			linkFilesetResp.Status.Code = 202
			linkFilesetResp.Jobs = make([]connectors.Job, 1)
			linkFilesetResp.Jobs[0].JobID = 1234
			linkFilesetResp.Jobs[0].Status = "COMPLETED"
			marshalledResponse, err := json.Marshal(linkFilesetResp)
			Expect(err).ToNot(HaveOccurred())
			registerurl := fakeurl + "/scalemgmt/v2/filesystems/" + filesystem + "/filesets/" + fileset + "/link"
			joburl := fakeurl + "/scalemgmt/v2/jobs?filter=jobId=1234"
			httpmock.RegisterResponder(
				"POST",
				registerurl,
				httpmock.NewStringResponder(202, string(marshalledResponse)),
			)

			httpmock.RegisterResponder(
				"GET",
				joburl,
				httpmock.NewStringResponder(200, string(marshalledResponse)),
			)

			getfilesysResp := connectors.GetFilesystemResponse_v2{}
			getfilesysResp.FileSystems = make([]connectors.FileSystem_v2, 1)
			getfilesysResp.FileSystems[0].Mount.MountPoint = "fakemount"
			getfilesysResp.Status.Code = 200
			marshalledResponse_filesys, err := json.Marshal(getfilesysResp)
			Expect(err).ToNot(HaveOccurred())
			registerurl = fakeurl + "/scalemgmt/v2/filesystems/" + filesystem
			httpmock.RegisterResponder(
				"GET",
				registerurl,
				httpmock.NewStringResponder(200, string(marshalledResponse_filesys)),
			)

			err = spectrumRestV2.LinkFileset(filesystem, fileset)
			Expect(err).ToNot(HaveOccurred())
		})

		It("Should fail with http error", func() {
			linkFilesetResp := connectors.GenericResponse{}
			linkFilesetResp.Status.Code = 500
			linkFilesetResp.Jobs = make([]connectors.Job, 1)
			linkFilesetResp.Jobs[0].JobID = 1234
			linkFilesetResp.Jobs[0].Status = "COMPLETED"
			marshalledResponse, err := json.Marshal(linkFilesetResp)
			Expect(err).ToNot(HaveOccurred())
			registerurl := fakeurl + "/scalemgmt/v2/filesystems/" + filesystem + "/filesets/" + fileset + "/link"
			joburl := fakeurl + "/scalemgmt/v2/jobs?filter=jobId=1234"

			httpmock.RegisterResponder(
				"POST",
				registerurl,
				httpmock.NewStringResponder(500, string(marshalledResponse)),
			)

			httpmock.RegisterResponder(
				"GET",
				joburl,
				httpmock.NewStringResponder(200, string(marshalledResponse)),
			)

			getfilesysResp := connectors.GetFilesystemResponse_v2{}
			getfilesysResp.FileSystems = make([]connectors.FileSystem_v2, 1)
			getfilesysResp.FileSystems[0].Mount.MountPoint = "fakemount"
			getfilesysResp.Status.Code = 200
			marshalledResponse_filesys, err := json.Marshal(getfilesysResp)
			Expect(err).ToNot(HaveOccurred())
			registerurl = fakeurl + "/scalemgmt/v2/filesystems/" + filesystem
			httpmock.RegisterResponder(
				"GET",
				registerurl,
				httpmock.NewStringResponder(200, string(marshalledResponse_filesys)),
			)
			err = spectrumRestV2.LinkFileset(filesystem, fileset)
			Expect(err).To(HaveOccurred())
		})

		It("Should fail since it is unable to fetch job details due to http error", func() {
			linkFilesetResp := connectors.GenericResponse{}
			linkFilesetResp.Status.Code = 202
			linkFilesetResp.Jobs = make([]connectors.Job, 1)
			linkFilesetResp.Jobs[0].JobID = 1234
			linkFilesetResp.Jobs[0].Status = "COMPLETED"
			marshalledResponse, err := json.Marshal(linkFilesetResp)
			Expect(err).ToNot(HaveOccurred())
			registerurl := fakeurl + "/scalemgmt/v2/filesystems/" + filesystem + "/filesets/" + fileset + "/link"
			joburl := fakeurl + "/scalemgmt/v2/jobs?filter=jobId=1234"

			httpmock.RegisterResponder(
				"POST",
				registerurl,
				httpmock.NewStringResponder(202, string(marshalledResponse)),
			)

			httpmock.RegisterResponder(
				"GET",
				joburl,
				httpmock.NewStringResponder(400, string(marshalledResponse)),
			)
			getfilesysResp := connectors.GetFilesystemResponse_v2{}
			getfilesysResp.FileSystems = make([]connectors.FileSystem_v2, 1)
			getfilesysResp.FileSystems[0].Mount.MountPoint = "fakemount"
			getfilesysResp.Status.Code = 200
			marshalledResponse_filesys, err := json.Marshal(getfilesysResp)
			Expect(err).ToNot(HaveOccurred())
			registerurl = fakeurl + "/scalemgmt/v2/filesystems/" + filesystem
			httpmock.RegisterResponder(
				"GET",
				registerurl,
				httpmock.NewStringResponder(200, string(marshalledResponse_filesys)),
			)
			err = spectrumRestV2.LinkFileset(filesystem, fileset)
			Expect(err).To(HaveOccurred())
		})

		It("Should fail to do zero length job array", func() {
			linkFilesetResp := connectors.GenericResponse{}
			linkFilesetResp.Status.Code = 202
			linkFilesetResp.Jobs = make([]connectors.Job, 0)
			marshalledResponse, err := json.Marshal(linkFilesetResp)
			Expect(err).ToNot(HaveOccurred())
			registerurl := fakeurl + "/scalemgmt/v2/filesystems/" + filesystem + "/filesets/" + fileset + "/link"
			joburl := fakeurl + "/scalemgmt/v2/jobs?filter=jobId=1234"

			httpmock.RegisterResponder(
				"POST",
				registerurl,
				httpmock.NewStringResponder(202, string(marshalledResponse)),
			)

			httpmock.RegisterResponder(
				"GET",
				joburl,
				httpmock.NewStringResponder(200, string(marshalledResponse)),
			)
			getfilesysResp := connectors.GetFilesystemResponse_v2{}
			getfilesysResp.FileSystems = make([]connectors.FileSystem_v2, 1)
			getfilesysResp.FileSystems[0].Mount.MountPoint = "fakemount"
			getfilesysResp.Status.Code = 200
			marshalledResponse_filesys, err := json.Marshal(getfilesysResp)
			Expect(err).ToNot(HaveOccurred())
			registerurl = fakeurl + "/scalemgmt/v2/filesystems/" + filesystem
			httpmock.RegisterResponder(
				"GET",
				registerurl,
				httpmock.NewStringResponder(200, string(marshalledResponse_filesys)),
			)

			err = spectrumRestV2.LinkFileset(filesystem, fileset)
			Expect(err).To(HaveOccurred())

		})
	})

	Context(".UnlinkFileset", func() {
		It("should pass while deleting a fileset", func() {
			Skip("TODO need to fix")
			unlinkFilesetResp := connectors.GenericResponse{}
			unlinkFilesetResp.Status.Code = 202
			unlinkFilesetResp.Jobs = make([]connectors.Job, 1)
			unlinkFilesetResp.Jobs[0].JobID = 1234
			unlinkFilesetResp.Jobs[0].Status = "COMPLETED"
			marshalledResponse, err := json.Marshal(unlinkFilesetResp)
			Expect(err).ToNot(HaveOccurred())
			registerurl := fakeurl + "/scalemgmt/v2/filesystems/" + filesystem + "/filesets/" + fileset + "/link"
			joburl := fakeurl + "/scalemgmt/v2/jobs?filter=jobId=1234"

			httpmock.RegisterResponder(
				"DELETE",
				registerurl,
				httpmock.NewStringResponder(202, string(marshalledResponse)),
			)

			httpmock.RegisterResponder(
				"GET",
				joburl,
				httpmock.NewStringResponder(200, string(marshalledResponse)),
			)
			err = spectrumRestV2.UnlinkFileset(filesystem, fileset)
			Expect(err).ToNot(HaveOccurred())
		})

		It("Should fail with http error", func() {
			unlinkFilesetResp := connectors.GenericResponse{}
			unlinkFilesetResp.Status.Code = 500
			unlinkFilesetResp.Jobs = make([]connectors.Job, 1)
			unlinkFilesetResp.Jobs[0].JobID = 1234
			unlinkFilesetResp.Jobs[0].Status = "COMPLETED"
			marshalledResponse, err := json.Marshal(unlinkFilesetResp)
			Expect(err).ToNot(HaveOccurred())
			registerurl := fakeurl + "/scalemgmt/v2/filesystems/" + filesystem + "/filesets/" + fileset + "/link"
			joburl := fakeurl + "/scalemgmt/v2/jobs?filter=jobId=1234"

			httpmock.RegisterResponder(
				"DELETE",
				registerurl,
				httpmock.NewStringResponder(500, string(marshalledResponse)),
			)

			httpmock.RegisterResponder(
				"GET",
				joburl,
				httpmock.NewStringResponder(200, string(marshalledResponse)),
			)
			err = spectrumRestV2.UnlinkFileset(filesystem, fileset)
			Expect(err).To(HaveOccurred())
		})

		It("Should fail since it is unable to fetch job details due to http error", func() {
			unlinkFilesetResp := connectors.GenericResponse{}
			unlinkFilesetResp.Status.Code = 202
			unlinkFilesetResp.Jobs = make([]connectors.Job, 1)
			unlinkFilesetResp.Jobs[0].JobID = 1234
			unlinkFilesetResp.Jobs[0].Status = "COMPLETED"
			marshalledResponse, err := json.Marshal(unlinkFilesetResp)
			Expect(err).ToNot(HaveOccurred())

			registerurl := fakeurl + "/scalemgmt/v2/filesystems/" + filesystem + "/filesets/" + fileset + "/link"
			joburl := fakeurl + "/scalemgmt/v2/jobs?filter=jobId=1234"
			httpmock.RegisterResponder(
				"DELETE",
				registerurl,
				httpmock.NewStringResponder(202, string(marshalledResponse)),
			)

			httpmock.RegisterResponder(
				"GET",
				joburl,
				httpmock.NewStringResponder(400, string(marshalledResponse)),
			)
			err = spectrumRestV2.UnlinkFileset(filesystem, fileset)
			Expect(err).To(HaveOccurred())
		})

		It("Should fail to do zero length job array", func() {
			unlinkFilesetResp := connectors.GenericResponse{}
			unlinkFilesetResp.Status.Code = 202
			unlinkFilesetResp.Jobs = make([]connectors.Job, 0)
			marshalledResponse, err := json.Marshal(unlinkFilesetResp)
			Expect(err).ToNot(HaveOccurred())
			registerurl := fakeurl + "/scalemgmt/v2/filesystems/" + filesystem + "/filesets/" + fileset + "/link"
			joburl := fakeurl + "/scalemgmt/v2/jobs?filter=jobId=1234"

			httpmock.RegisterResponder(
				"DELETE",
				registerurl,
				httpmock.NewStringResponder(202, string(marshalledResponse)),
			)

			httpmock.RegisterResponder(
				"GET",
				joburl,
				httpmock.NewStringResponder(200, string(marshalledResponse)),
			)
			err = spectrumRestV2.UnlinkFileset(filesystem, fileset)
			Expect(err).To(HaveOccurred())

		})
	})

	Context(".ListFileset", func() {
		It("Should pass with fileset info", func() {
			getFilesetRespo := connectors.GetFilesetResponse_v2{}
			getFilesetRespo.Status.Code = 200
			getFilesetRespo.Filesets = make([]connectors.Fileset_v2, 1)
			getFilesetRespo.Filesets[0].Config.FilesetName = fileset
			getFilesetRespo.Filesets[0].Config.Path = "fakemount"
			marshalledResponse, err := json.Marshal(getFilesetRespo)
			Expect(err).ToNot(HaveOccurred())
			registerurl := fakeurl + "/scalemgmt/v2/filesystems/" + filesystem + "/filesets/" + fileset
			httpmock.RegisterResponder(
				"GET",
				registerurl,
				httpmock.NewStringResponder(200, string(marshalledResponse)),
			)
			volumemeta, err := spectrumRestV2.ListFileset(filesystem, fileset)
			Expect(err).ToNot(HaveOccurred())
			Expect(volumemeta.Name).To(Equal(fileset))
			Expect(volumemeta.Mountpoint).To(Equal("fakemount"))
		})

		It("Should fail with http error", func() {
			getFilesetRespo := connectors.GetFilesetResponse_v2{}
			getFilesetRespo.Status.Code = 500
			getFilesetRespo.Filesets = make([]connectors.Fileset_v2, 1)
			getFilesetRespo.Filesets[0].Config.FilesetName = fileset
			getFilesetRespo.Filesets[0].Config.Path = "fakemount"
			marshalledResponse, err := json.Marshal(getFilesetRespo)
			Expect(err).ToNot(HaveOccurred())
			registerurl := fakeurl + "/scalemgmt/v2/filesystems/" + filesystem + "/filesets/" + fileset
			httpmock.RegisterResponder(
				"GET",
				registerurl,
				httpmock.NewStringResponder(500, string(marshalledResponse)),
			)
			volumemeta, err := spectrumRestV2.ListFileset(filesystem, fileset)
			Expect(err).To(HaveOccurred())
			Expect(volumemeta.Name).To(Equal(""))
			Expect(volumemeta.Mountpoint).To(Equal(""))
		})

		It("Should fail due to zero length fileset", func() {
			getFilesetRespo := connectors.GetFilesetResponse_v2{}
			getFilesetRespo.Status.Code = 500
			getFilesetRespo.Filesets = make([]connectors.Fileset_v2, 0)
			marshalledResponse, err := json.Marshal(getFilesetRespo)
			Expect(err).ToNot(HaveOccurred())
			registerurl := fakeurl + "/scalemgmt/v2/filesystems/" + filesystem + "/filesets/" + fileset
			httpmock.RegisterResponder(
				"GET",
				registerurl,
				httpmock.NewStringResponder(500, string(marshalledResponse)),
			)
			volumemeta, err := spectrumRestV2.ListFileset(filesystem, fileset)
			Expect(err).To(HaveOccurred())
			Expect(volumemeta.Name).To(Equal(""))
			Expect(volumemeta.Mountpoint).To(Equal(""))
		})

		It("Should pass with fileset info", func() {
			getFilesetRespo := connectors.GetFilesetResponse_v2{}
			getFilesetRespo.Status.Code = 200
			getFilesetRespo.Filesets = make([]connectors.Fileset_v2, 4)
			getFilesetRespo.Filesets[0].Config.FilesetName = fileset
			getFilesetRespo.Filesets[0].Config.Path = "fakemount"

			getFilesetRespo.Filesets[1].Config.FilesetName = fileset
			getFilesetRespo.Filesets[1].Config.Path = "fakemount"

			getFilesetRespo.Filesets[2].Config.FilesetName = fileset
			getFilesetRespo.Filesets[2].Config.Path = "fakemount"

			getFilesetRespo.Filesets[3].Config.FilesetName = fileset
			getFilesetRespo.Filesets[3].Config.Path = "fakemount"

			marshalledResponse, err := json.Marshal(getFilesetRespo)
			Expect(err).ToNot(HaveOccurred())
			registerurl := fakeurl + "/scalemgmt/v2/filesystems/" + filesystem + "/filesets"
			httpmock.RegisterResponder(
				"GET",
				registerurl,
				httpmock.NewStringResponder(200, string(marshalledResponse)),
			)
			volumemeta_arr, err := spectrumRestV2.ListFilesets(filesystem)
			Expect(err).ToNot(HaveOccurred())
			Expect(len(volumemeta_arr)).To(Equal(4))
			for iter := 0; iter < len(volumemeta_arr); iter++ {
				Expect(volumemeta_arr[iter].Name).To(Equal(fileset))
				Expect(volumemeta_arr[iter].Mountpoint).To(Equal("fakemount"))
			}
		})

		It("Should fail with http error", func() {
			getFilesetRespo := connectors.GetFilesetResponse_v2{}
			getFilesetRespo.Status.Code = 500
			getFilesetRespo.Filesets = make([]connectors.Fileset_v2, 4)
			getFilesetRespo.Filesets[0].Config.FilesetName = fileset
			getFilesetRespo.Filesets[0].Config.Path = "fakemount"

			getFilesetRespo.Filesets[1].Config.FilesetName = fileset
			getFilesetRespo.Filesets[1].Config.Path = "fakemount"

			getFilesetRespo.Filesets[2].Config.FilesetName = fileset
			getFilesetRespo.Filesets[2].Config.Path = "fakemount"

			getFilesetRespo.Filesets[3].Config.FilesetName = fileset
			getFilesetRespo.Filesets[3].Config.Path = "fakemount"

			marshalledResponse, err := json.Marshal(getFilesetRespo)
			Expect(err).ToNot(HaveOccurred())
			registerurl := fakeurl + "/scalemgmt/v2/filesystems/" + filesystem + "/filesets"
			httpmock.RegisterResponder(
				"GET",
				registerurl,
				httpmock.NewStringResponder(500, string(marshalledResponse)),
			)
			volumemeta_arr, err := spectrumRestV2.ListFilesets(filesystem)
			Expect(err).To(HaveOccurred())
			Expect(len(volumemeta_arr)).To(Equal(0))
		})

		It("Should fail with zero length", func() {
			getFilesetRespo := connectors.GetFilesetResponse_v2{}
			getFilesetRespo.Status.Code = 200
			getFilesetRespo.Filesets = make([]connectors.Fileset_v2, 0)

			marshalledResponse, err := json.Marshal(getFilesetRespo)
			Expect(err).ToNot(HaveOccurred())
			registerurl := fakeurl + "/scalemgmt/v2/filesystems/" + filesystem + "/filesets"
			httpmock.RegisterResponder(
				"GET",
				registerurl,
				httpmock.NewStringResponder(200, string(marshalledResponse)),
			)
			volumemeta_arr, err := spectrumRestV2.ListFilesets(filesystem)
			Expect(err).ToNot(HaveOccurred())
			Expect(len(volumemeta_arr)).To(Equal(0))
		})
	})

	Context(".IsFilesetLinked", func() {
		It("Should pass since fileset is linked", func() {
			getFilesetRespo := connectors.GetFilesetResponse_v2{}
			getFilesetRespo.Status.Code = 200
			getFilesetRespo.Filesets = make([]connectors.Fileset_v2, 1)
			getFilesetRespo.Filesets[0].Config.FilesetName = fileset
			getFilesetRespo.Filesets[0].Config.Path = "fakemount"
			marshalledResponse, err := json.Marshal(getFilesetRespo)
			Expect(err).ToNot(HaveOccurred())
			registerurl := fakeurl + "/scalemgmt/v2/filesystems/" + filesystem + "/filesets/" + fileset
			httpmock.RegisterResponder(
				"GET",
				registerurl,
				httpmock.NewStringResponder(200, string(marshalledResponse)),
			)
			islinked, err := spectrumRestV2.IsFilesetLinked(filesystem, fileset)
			Expect(err).ToNot(HaveOccurred())
			Expect(islinked).To(Equal(true))

		})

		It("Should fail with http error", func() {
			getFilesetRespo := connectors.GetFilesetResponse_v2{}
			getFilesetRespo.Status.Code = 500
			getFilesetRespo.Filesets = make([]connectors.Fileset_v2, 1)
			getFilesetRespo.Filesets[0].Config.FilesetName = fileset
			getFilesetRespo.Filesets[0].Config.Path = "fakemount"
			marshalledResponse, err := json.Marshal(getFilesetRespo)
			Expect(err).ToNot(HaveOccurred())
			registerurl := fakeurl + "/scalemgmt/v2/filesystems/" + filesystem + "/filesets/" + fileset
			httpmock.RegisterResponder(
				"GET",
				registerurl,
				httpmock.NewStringResponder(500, string(marshalledResponse)),
			)
			islinked, err := spectrumRestV2.IsFilesetLinked(filesystem, fileset)
			Expect(err).To(HaveOccurred())
			Expect(islinked).To(Equal(false))
		})

		It("Should return false as mount point is empty", func() {
			getFilesetRespo := connectors.GetFilesetResponse_v2{}
			getFilesetRespo.Status.Code = 200
			getFilesetRespo.Filesets = make([]connectors.Fileset_v2, 1)
			getFilesetRespo.Filesets[0].Config.FilesetName = fileset
			getFilesetRespo.Filesets[0].Config.Path = ""
			marshalledResponse, err := json.Marshal(getFilesetRespo)
			Expect(err).ToNot(HaveOccurred())
			registerurl := fakeurl + "/scalemgmt/v2/filesystems/" + filesystem + "/filesets/" + fileset
			httpmock.RegisterResponder(
				"GET",
				registerurl,
				httpmock.NewStringResponder(200, string(marshalledResponse)),
			)
			islinked, err := spectrumRestV2.IsFilesetLinked(filesystem, fileset)
			Expect(err).ToNot(HaveOccurred())
			Expect(islinked).To(Equal(false))
		})

		It("Should fail as mount point contains --", func() {
			getFilesetRespo := connectors.GetFilesetResponse_v2{}
			getFilesetRespo.Status.Code = 200
			getFilesetRespo.Filesets = make([]connectors.Fileset_v2, 1)
			getFilesetRespo.Filesets[0].Config.FilesetName = fileset
			getFilesetRespo.Filesets[0].Config.Path = "--"
			marshalledResponse, err := json.Marshal(getFilesetRespo)
			Expect(err).ToNot(HaveOccurred())
			registerurl := fakeurl + "/scalemgmt/v2/filesystems/" + filesystem + "/filesets/" + fileset
			httpmock.RegisterResponder(
				"GET",
				registerurl,
				httpmock.NewStringResponder(200, string(marshalledResponse)),
			)
			islinked, err := spectrumRestV2.IsFilesetLinked(filesystem, fileset)
			Expect(err).ToNot(HaveOccurred())
			Expect(islinked).To(Equal(false))
		})
	})

	Context(".SetFilesetQuota", func() {
		It("Should pass while creating a fileset", func() {
			Skip("TODO need to fix")
			setFilesetResp := connectors.GenericResponse{}
			setFilesetResp.Status.Code = 202
			setFilesetResp.Jobs = make([]connectors.Job, 1)
			setFilesetResp.Jobs[0].JobID = 1234
			setFilesetResp.Jobs[0].Status = "COMPLETED"
			marshalledResponse, err := json.Marshal(setFilesetResp)
			Expect(err).ToNot(HaveOccurred())
			registerurl := fakeurl + "/scalemgmt/v2/filesystems/" + filesystem + "/filesets/" + fileset + "/quotas"
			joburl := fakeurl + "/scalemgmt/v2/jobs?filter=jobId=1234"

			httpmock.RegisterResponder(
				"POST",
				registerurl,
				httpmock.NewStringResponder(202, string(marshalledResponse)),
			)

			httpmock.RegisterResponder(
				"GET",
				joburl,
				httpmock.NewStringResponder(200, string(marshalledResponse)),
			)
			err = spectrumRestV2.SetFilesetQuota(filesystem, fileset, "10G")
			Expect(err).ToNot(HaveOccurred())
		})

		It("Should fail with http error", func() {
			setFilesetResp := connectors.GenericResponse{}
			setFilesetResp.Status.Code = 500
			setFilesetResp.Jobs = make([]connectors.Job, 1)
			setFilesetResp.Jobs[0].JobID = 1234
			setFilesetResp.Jobs[0].Status = "COMPLETED"
			marshalledResponse, err := json.Marshal(setFilesetResp)
			Expect(err).ToNot(HaveOccurred())
			registerurl := fakeurl + "/scalemgmt/v2/filesystems/" + filesystem + "/filesets/" + fileset + "/quotas"
			joburl := fakeurl + "/scalemgmt/v2/jobs?filter=jobId=1234"

			httpmock.RegisterResponder(
				"POST",
				registerurl,
				httpmock.NewStringResponder(500, string(marshalledResponse)),
			)

			httpmock.RegisterResponder(
				"GET",
				joburl,
				httpmock.NewStringResponder(200, string(marshalledResponse)),
			)
			err = spectrumRestV2.SetFilesetQuota(filesystem, fileset, "10G")
			Expect(err).To(HaveOccurred())
		})

		It("Should fail since it is unable to fetch job details due to http error", func() {
			setFilesetResp := connectors.GenericResponse{}
			setFilesetResp.Status.Code = 202
			setFilesetResp.Jobs = make([]connectors.Job, 1)
			setFilesetResp.Jobs[0].JobID = 1234
			setFilesetResp.Jobs[0].Status = "COMPLETED"
			marshalledResponse, err := json.Marshal(setFilesetResp)
			Expect(err).ToNot(HaveOccurred())
			registerurl := fakeurl + "/scalemgmt/v2/filesystems/" + filesystem + "/filesets/" + fileset + "/quotas"
			joburl := fakeurl + "/scalemgmt/v2/jobs?filter=jobId=1234"

			httpmock.RegisterResponder(
				"POST",
				registerurl,
				httpmock.NewStringResponder(202, string(marshalledResponse)),
			)

			httpmock.RegisterResponder(
				"GET",
				joburl,
				httpmock.NewStringResponder(400, string(marshalledResponse)),
			)
			err = spectrumRestV2.SetFilesetQuota(filesystem, fileset, "10G")
			Expect(err).To(HaveOccurred())
		})

		It("Should fail to do zero length job array", func() {
			setFilesetResp := connectors.GenericResponse{}
			setFilesetResp.Status.Code = 202
			setFilesetResp.Jobs = make([]connectors.Job, 0)
			marshalledResponse, err := json.Marshal(setFilesetResp)
			Expect(err).ToNot(HaveOccurred())
			registerurl := fakeurl + "/scalemgmt/v2/filesystems/" + filesystem + "/filesets/" + fileset + "/quotas"
			joburl := fakeurl + "/scalemgmt/v2/jobs?filter=jobId=1234"

			httpmock.RegisterResponder(
				"POST",
				registerurl,
				httpmock.NewStringResponder(202, string(marshalledResponse)),
			)

			httpmock.RegisterResponder(
				"GET",
				joburl,
				httpmock.NewStringResponder(200, string(marshalledResponse)),
			)
			err = spectrumRestV2.SetFilesetQuota(filesystem, fileset, "10G")
			Expect(err).To(HaveOccurred())

		})
	})

	Context(".ListFilesetQuota", func() {
		It("Should pass after getting quota", func() {
			getFilesetquota := connectors.GetQuotaResponse_v2{}
			getFilesetquota.Status.Code = 200
			getFilesetquota.Quotas = make([]connectors.Quota_v2, 1)
			getFilesetquota.Quotas[0].BlockQuota = 1024

			marshalledResponse, err := json.Marshal(getFilesetquota)
			Expect(err).ToNot(HaveOccurred())
			registerurl := fakeurl + "/scalemgmt/v2/filesystems/" + filesystem + "/filesets/" + fileset + "/quotas"
			httpmock.RegisterResponder(
				"GET",
				registerurl,
				httpmock.NewStringResponder(200, string(marshalledResponse)),
			)
			quota, err := spectrumRestV2.ListFilesetQuota(filesystem, fileset)
			Expect(err).ToNot(HaveOccurred())
			Expect(quota).To(Equal("1024"))
		})

		It("Should fail with http error", func() {
			getFilesetquota := connectors.GetQuotaResponse_v2{}
			getFilesetquota.Status.Code = 500
			getFilesetquota.Quotas = make([]connectors.Quota_v2, 1)
			getFilesetquota.Quotas[0].BlockQuota = 1024

			marshalledResponse, err := json.Marshal(getFilesetquota)
			Expect(err).ToNot(HaveOccurred())
			registerurl := fakeurl + "/scalemgmt/v2/filesystems/" + filesystem + "/filesets/" + fileset + "/quotas"
			httpmock.RegisterResponder(
				"GET",
				registerurl,
				httpmock.NewStringResponder(500, string(marshalledResponse)),
			)
			quota, err := spectrumRestV2.ListFilesetQuota(filesystem, fileset)
			Expect(err).To(HaveOccurred())
			Expect(quota).To(Equal(""))
		})

		It("should fail with zero length quota array", func() {
			getFilesetquota := connectors.GetQuotaResponse_v2{}
			getFilesetquota.Status.Code = 200
			getFilesetquota.Quotas = make([]connectors.Quota_v2, 0)

			marshalledResponse, err := json.Marshal(getFilesetquota)
			Expect(err).ToNot(HaveOccurred())
			registerurl := fakeurl + "/scalemgmt/v2/filesystems/" + filesystem + "/filesets/" + fileset + "/quotas"
			httpmock.RegisterResponder(
				"GET",
				registerurl,
				httpmock.NewStringResponder(200, string(marshalledResponse)),
			)
			quota, err := spectrumRestV2.ListFilesetQuota(filesystem, fileset)
			Expect(err).To(HaveOccurred())
			Expect(quota).To(Equal(""))
		})
	})

	Context(".ExportNfs", func() {
		It("Should pass while creating a fileset", func() {
			Skip("TODO need to fix")
			exportNfsResp := connectors.GenericResponse{}
			exportNfsResp.Status.Code = 202
			exportNfsResp.Jobs = make([]connectors.Job, 1)
			exportNfsResp.Jobs[0].JobID = 1234
			exportNfsResp.Jobs[0].Status = "COMPLETED"
			marshalledResponse, err := json.Marshal(exportNfsResp)
			Expect(err).ToNot(HaveOccurred())
			registerurl := fakeurl + "/scalemgmt/v2/nfs/exports"
			joburl := fakeurl + "/scalemgmt/v2/jobs?filter=jobId=1234"

			httpmock.RegisterResponder(
				"POST",
				registerurl,
				httpmock.NewStringResponder(202, string(marshalledResponse)),
			)

			httpmock.RegisterResponder(
				"GET",
				joburl,
				httpmock.NewStringResponder(200, string(marshalledResponse)),
			)
			err = spectrumRestV2.ExportNfs(filesystem, filesystem)
			Expect(err).ToNot(HaveOccurred())
		})

		It("Should fail with http error", func() {
			exportNfsResp := connectors.GenericResponse{}
			exportNfsResp.Status.Code = 500
			exportNfsResp.Jobs = make([]connectors.Job, 1)
			exportNfsResp.Jobs[0].JobID = 1234
			exportNfsResp.Jobs[0].Status = "COMPLETED"
			marshalledResponse, err := json.Marshal(exportNfsResp)
			Expect(err).ToNot(HaveOccurred())
			registerurl := fakeurl + "/scalemgmt/v2/nfs/exports"
			joburl := fakeurl + "/scalemgmt/v2/jobs?filter=jobId=1234"

			httpmock.RegisterResponder(
				"POST",
				registerurl,
				httpmock.NewStringResponder(500, string(marshalledResponse)),
			)

			httpmock.RegisterResponder(
				"GET",
				joburl,
				httpmock.NewStringResponder(200, string(marshalledResponse)),
			)
			err = spectrumRestV2.ExportNfs(filesystem, fileset)
			Expect(err).To(HaveOccurred())
		})

		It("Should fail since it is unable to fetch job details due to http error", func() {
			exportNfsResp := connectors.GenericResponse{}
			exportNfsResp.Status.Code = 202
			exportNfsResp.Jobs = make([]connectors.Job, 1)
			exportNfsResp.Jobs[0].JobID = 1234
			exportNfsResp.Jobs[0].Status = "COMPLETED"
			marshalledResponse, err := json.Marshal(exportNfsResp)
			Expect(err).ToNot(HaveOccurred())
			registerurl := fakeurl + "/scalemgmt/v2/nfs/exports"
			joburl := fakeurl + "/scalemgmt/v2/jobs?filter=jobId=1234"

			httpmock.RegisterResponder(
				"POST",
				registerurl,
				httpmock.NewStringResponder(202, string(marshalledResponse)),
			)

			httpmock.RegisterResponder(
				"GET",
				joburl,
				httpmock.NewStringResponder(400, string(marshalledResponse)),
			)
			err = spectrumRestV2.ExportNfs(filesystem, fileset)
			Expect(err).To(HaveOccurred())
		})

		It("Should fail to do zero length job array", func() {
			exportNfsResp := connectors.GenericResponse{}
			exportNfsResp.Status.Code = 202
			exportNfsResp.Jobs = make([]connectors.Job, 0)
			marshalledResponse, err := json.Marshal(exportNfsResp)
			Expect(err).ToNot(HaveOccurred())
			registerurl := fakeurl + "/scalemgmt/v2/nfs/exports"
			joburl := fakeurl + "/scalemgmt/v2/jobs?filter=jobId=1234"

			httpmock.RegisterResponder(
				"POST",
				registerurl,
				httpmock.NewStringResponder(202, string(marshalledResponse)),
			)

			httpmock.RegisterResponder(
				"GET",
				joburl,
				httpmock.NewStringResponder(200, string(marshalledResponse)),
			)
			err = spectrumRestV2.ExportNfs(filesystem, fileset)
			Expect(err).To(HaveOccurred())

		})
	})

	Context(".UnexportNfs", func() {
		It("Should pass while deleting a fileset", func() {
			Skip("TODO need to fix")
			unexportNfs := connectors.GenericResponse{}
			unexportNfs.Status.Code = 202
			unexportNfs.Jobs = make([]connectors.Job, 1)
			unexportNfs.Jobs[0].JobID = 1234
			unexportNfs.Jobs[0].Status = "COMPLETED"
			marshalledResponse, err := json.Marshal(unexportNfs)
			Expect(err).ToNot(HaveOccurred())
			registerurl := fakeurl + "/scalemgmt/v2/nfs/exports/" + fileset
			joburl := fakeurl + "/scalemgmt/v2/jobs?filter=jobId=1234"

			httpmock.RegisterResponder(
				"DELETE",
				registerurl,
				httpmock.NewStringResponder(202, string(marshalledResponse)),
			)

			httpmock.RegisterResponder(
				"GET",
				joburl,
				httpmock.NewStringResponder(200, string(marshalledResponse)),
			)
			err = spectrumRestV2.UnexportNfs(fileset)
			Expect(err).ToNot(HaveOccurred())
		})

		It("Should fail with http error", func() {
			unexportNfs := connectors.GenericResponse{}
			unexportNfs.Status.Code = 500
			unexportNfs.Jobs = make([]connectors.Job, 1)
			unexportNfs.Jobs[0].JobID = 1234
			unexportNfs.Jobs[0].Status = "COMPLETED"
			marshalledResponse, err := json.Marshal(unexportNfs)
			Expect(err).ToNot(HaveOccurred())
			registerurl := fakeurl + "/scalemgmt/v2/nfs/exports/" + fileset
			joburl := fakeurl + "/scalemgmt/v2/jobs?filter=jobId=1234"

			httpmock.RegisterResponder(
				"DELETE",
				registerurl,
				httpmock.NewStringResponder(500, string(marshalledResponse)),
			)

			httpmock.RegisterResponder(
				"GET",
				joburl,
				httpmock.NewStringResponder(200, string(marshalledResponse)),
			)
			err = spectrumRestV2.UnexportNfs(fileset)
			Expect(err).To(HaveOccurred())
		})

		It("Should fail since it is unable to fetch job details due to http error", func() {
			unexportNfs := connectors.GenericResponse{}
			unexportNfs.Status.Code = 202
			unexportNfs.Jobs = make([]connectors.Job, 1)
			unexportNfs.Jobs[0].JobID = 1234
			unexportNfs.Jobs[0].Status = "COMPLETED"
			marshalledResponse, err := json.Marshal(unexportNfs)
			Expect(err).ToNot(HaveOccurred())

			registerurl := fakeurl + "/scalemgmt/v2/nfs/exports/" + fileset
			joburl := fakeurl + "/scalemgmt/v2/jobs?filter=jobId=1234"
			httpmock.RegisterResponder(
				"DELETE",
				registerurl,
				httpmock.NewStringResponder(202, string(marshalledResponse)),
			)

			httpmock.RegisterResponder(
				"GET",
				joburl,
				httpmock.NewStringResponder(400, string(marshalledResponse)),
			)
			err = spectrumRestV2.UnexportNfs(fileset)
			Expect(err).To(HaveOccurred())
		})

		It("Should fail to do zero length job array", func() {
			unexportNfs := connectors.GenericResponse{}
			unexportNfs.Status.Code = 202
			unexportNfs.Jobs = make([]connectors.Job, 0)
			marshalledResponse, err := json.Marshal(unexportNfs)
			Expect(err).ToNot(HaveOccurred())
			registerurl := fakeurl + "/scalemgmt/v2/nfs/exports/" + fileset
			joburl := fakeurl + "/scalemgmt/v2/jobs?filter=jobId=1234"

			httpmock.RegisterResponder(
				"DELETE",
				registerurl,
				httpmock.NewStringResponder(202, string(marshalledResponse)),
			)

			httpmock.RegisterResponder(
				"GET",
				joburl,
				httpmock.NewStringResponder(200, string(marshalledResponse)),
			)
			err = spectrumRestV2.UnexportNfs(fileset)
			Expect(err).To(HaveOccurred())

		})
	})
})
