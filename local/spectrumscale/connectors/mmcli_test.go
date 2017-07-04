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
	"fmt"
	"log"
	"os"

	"github.com/IBM/ubiquity/fakes"
	"github.com/IBM/ubiquity/local/spectrumscale/connectors"
	"github.com/IBM/ubiquity/resources"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("spectrum_mmcli", func() {
	var (
		spectrumMMCLI connectors.SpectrumScaleConnector
		logger        *log.Logger
		fakeExec      *fakes.FakeExecutor
		opts          map[string]interface{}
		err           error
		fileset       string
		filesystem    string
	)

	BeforeEach(func() {
		logger = log.New(os.Stdout, "spectrum: ", log.Lshortfile|log.LstdFlags)
		fakeExec = new(fakes.FakeExecutor)
		spectrumMMCLI, err = connectors.NewSpectrumMMCLIWithExecutor(logger, fakeExec)
		Expect(err).ToNot(HaveOccurred())
		fileset = "fake-fileset"
		filesystem = "fake-filesystem"
	})

	Context(".GetClusterID", func() {
		It("should fail when execute command errors", func() {
			fakeExec.ExecuteReturns(nil, fmt.Errorf("failed to execute command"))

			id, err := spectrumMMCLI.GetClusterId()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("failed to execute command"))
			Expect(id).To(Equal(""))
		})

		It("should fail when execute command returns a non expected value", func() {
			fakeExec.ExecuteReturns(nil, nil)

			id, err := spectrumMMCLI.GetClusterId()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Error determining cluster id"))
			Expect(id).To(Equal(""))
		})

		It("should fail when execute command returns a wrong number of info", func() {
			stringOutput := "a \n GPFS cluster id: 123 \n"
			fakeExec.ExecuteReturns([]byte(stringOutput), nil)

			id, err := spectrumMMCLI.GetClusterId()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Error determining cluster id"))
			Expect(id).To(Equal(""))
		})

		It("should succeed when execute command returns the right info", func() {
			stringOutput := "a \n b\n c \n d \nGPFS cluster id:123\n"
			fakeExec.ExecuteReturns([]byte(stringOutput), nil)

			clusterID, err := spectrumMMCLI.GetClusterId()
			Expect(err).ToNot(HaveOccurred())
			Expect(clusterID).To(Equal("123"))
		})
	})

	Context(".IsFilesystemMounted", func() {
		It("should fail when execute command errors", func() {
			fakeExec.ExecuteReturns(nil, fmt.Errorf("failed to execute command"))

			response, err := spectrumMMCLI.IsFilesystemMounted("fake-filesystem")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("failed to execute command"))
			Expect(response).To(Equal(false))
		})

		It("should fail when execute hostname errors", func() {
			fakeExec.HostnameReturns("", fmt.Errorf("failed to execute hostname"))
			stringOutput := "a \n1:2:3:4:5:6:7:8:9:10:11:12\n"
			fakeExec.ExecuteReturns([]byte(stringOutput), nil)

			response, err := spectrumMMCLI.IsFilesystemMounted("fake-filesystem")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("failed to execute hostname"))
			Expect(response).To(Equal(false))
		})

		It("should succeed when execute command returns response that does not contain the actual hostname", func() {
			fakeExec.HostnameReturns("fake-host", nil)
			stringOutput := "a \n1:2:3:4:5:6:7:8:9:10:11:12\n"
			fakeExec.ExecuteReturns([]byte(stringOutput), nil)

			response, err := spectrumMMCLI.IsFilesystemMounted("fake-filesystem")
			Expect(err).ToNot(HaveOccurred())
			Expect(response).To(Equal(false))
		})

		It("should succeed when execute command returns response that contains the actual hostname", func() {
			fakeExec.HostnameReturns("fake-host", nil)
			stringOutput := "a \n1:2:3:4:5:6:7:8:9:10:11:fake-host\n 12:21:32"
			fakeExec.ExecuteReturns([]byte(stringOutput), nil)

			response, err := spectrumMMCLI.IsFilesystemMounted("fake-filesystem")
			Expect(err).ToNot(HaveOccurred())
			Expect(response).To(Equal(true))
		})
	})

	Context(".MountFileSystem", func() {
		It("should fail when execute command errors", func() {
			fakeExec.ExecuteReturns(nil, fmt.Errorf("failed to execute command"))

			err = spectrumMMCLI.MountFileSystem("fake-filesystem")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("failed to execute command"))
		})

		It("should succeed when execute command does not error", func() {
			fakeExec.ExecuteReturns(nil, nil)

			err = spectrumMMCLI.MountFileSystem("fake-filesystem")
			Expect(err).ToNot(HaveOccurred())
		})
	})

	//TODO when listfilesystems is implemented
	Context(".ListFileSystems", func() {
		It("", func() {})
	})

	Context(".GetFileSystemMountpoint", func() {
		It("should fail when execute command errors", func() {
			fakeExec.ExecuteReturns(nil, fmt.Errorf("failed to execute command"))

			mountpoint, err := spectrumMMCLI.GetFilesystemMountpoint("fake-filesystem")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("failed to execute command"))
			Expect(mountpoint).To(Equal(""))
		})

		It("should fail when execute command returns non valid output", func() {
			stringOutput := "a"
			fakeExec.ExecuteReturns([]byte(stringOutput), nil)

			mountpoint, err := spectrumMMCLI.GetFilesystemMountpoint("fake-filesystem")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Cannot determine filesystem mountpoint"))
			Expect(mountpoint).To(Equal(""))
		})

		It("should fail when execute command returns non valid output", func() {
			stringOutput := "a \nfake-filesystem:8:fake-mountpoint:10:11:fake-host\n 12:21:32"
			fakeExec.ExecuteReturns([]byte(stringOutput), nil)

			mountpoint, err := spectrumMMCLI.GetFilesystemMountpoint("fake-filesystem")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Cannot determine filesystem mountpoint"))
			Expect(mountpoint).To(Equal(""))
		})

		It("should succeed when execute command returns valid output", func() {
			stringOutput := "a \n1:2:3:4:5:6:fake-filesystem:8:fake-mountpoint:10:11:fake-host\n 12:21:32"
			fakeExec.ExecuteReturns([]byte(stringOutput), nil)

			mountpoint, err := spectrumMMCLI.GetFilesystemMountpoint("fake-filesystem")
			Expect(err).ToNot(HaveOccurred())
			Expect(mountpoint).To(Equal("fake-mountpoint"))
		})
	})

	Context(".CreateFileset", func() {
		It("should fail when execute command errors", func() {
			errorMsg := fmt.Sprintf("Failed to create fileset %s on filesystem %s. Please check that filesystem specified is correct and healthy", fileset, filesystem)
			fakeExec.ExecuteReturns(nil, fmt.Errorf(errorMsg))

			err = spectrumMMCLI.CreateFileset(filesystem, fileset, opts)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(errorMsg))
		})
		It("should succeed when execute command does not error", func() {
			fakeExec.ExecuteReturns(nil, nil)

			err = spectrumMMCLI.CreateFileset(filesystem, fileset, opts)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context(".DeleteFileset", func() {
		It("should fail when execute command errors", func() {
			errorMsg := fmt.Sprintf("error executing command")
			fakeExec.ExecuteReturns(nil, fmt.Errorf(errorMsg))

			err = spectrumMMCLI.DeleteFileset(filesystem, fileset)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(fmt.Sprintf("Failed to remove fileset %s: %s ", fileset, errorMsg)))
		})

		It("should succeed when execute command does not error", func() {
			fakeExec.ExecuteReturns(nil, nil)

			err = spectrumMMCLI.DeleteFileset(filesystem, fileset)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context(".IsFilesetLinked", func() {
		It("should fail when execute command errors", func() {
			errorMsg := fmt.Sprintf("error executing command")
			fakeExec.ExecuteReturns(nil, fmt.Errorf(errorMsg))

			response, err := spectrumMMCLI.IsFilesetLinked(filesystem, fileset)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("error executing command"))
			Expect(response).To(Equal(false))
		})

		It("should fail when execute command returns one line output", func() {
			stringOutput := "a"
			fakeExec.ExecuteReturns([]byte(stringOutput), nil)

			response, err := spectrumMMCLI.IsFilesetLinked(filesystem, fileset)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(fmt.Sprintf("Error listing fileset %s", fileset)))
			Expect(response).To(Equal(false))
		})

		It("should fail when execute command returns non valid output", func() {
			stringOutput := "a \n int:10:11:fake-host\n 12:21:32"
			fakeExec.ExecuteReturns([]byte(stringOutput), nil)

			response, err := spectrumMMCLI.IsFilesetLinked(filesystem, fileset)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(fmt.Sprintf("Error listing fileset %s after parsing", fileset)))
			Expect(response).To(Equal(false))
		})

		It("should succeed when execute command returns valid output", func() {
			stringOutput := "a \n1:2:3:4:5:6:fake-filesystem:8:fake-mountpoint:10:Linked:fake-host"
			fakeExec.ExecuteReturns([]byte(stringOutput), nil)

			response, err := spectrumMMCLI.IsFilesetLinked(filesystem, fileset)
			Expect(err).ToNot(HaveOccurred())
			Expect(response).To(Equal(true))
		})

		It("should succeed when execute command returns valid output", func() {
			stringOutput := "a \n1:2:3:4:5:6:fake-filesystem:8:fake-mountpoint:10:nothing:fake-host"
			fakeExec.ExecuteReturns([]byte(stringOutput), nil)

			response, err := spectrumMMCLI.IsFilesetLinked(filesystem, fileset)
			Expect(err).ToNot(HaveOccurred())
			Expect(response).To(Equal(false))
		})
	})

	Context(".LinkFileset", func() {
		It("should fail when execute command errors", func() {
			errorMsg := fmt.Sprintf("error executing command")
			fakeExec.ExecuteReturns(nil, fmt.Errorf(errorMsg))

			err = spectrumMMCLI.LinkFileset(filesystem, fileset)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(errorMsg))
			Expect(fakeExec.ExecuteCallCount()).To(Equal(1))
		})

		It("should succeed when execute command does not error", func() {
			returnMsg := "a \n1:2:3:4:5:6:fake-filesystem:8:fake-mountpoint:10:11:fake-host\n 12:21:32"
			fakeExec.ExecuteReturns([]byte(returnMsg), nil)

			err = spectrumMMCLI.LinkFileset(filesystem, fileset)
			Expect(err).ToNot(HaveOccurred())
			Expect(fakeExec.ExecuteCallCount()).To(Equal(2))
		})
	})

	Context(".UnlinkFileset", func() {
		It("should fail when execute command errors", func() {
			errorMsg := fmt.Sprintf("error executing command")
			fakeExec.ExecuteReturns(nil, fmt.Errorf(errorMsg))

			err = spectrumMMCLI.UnlinkFileset(filesystem, fileset)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(fmt.Sprintf("Failed to unlink fileset %s: %s", fileset, errorMsg)))
		})

		It("should succeed when execute command does not error", func() {
			fakeExec.ExecuteReturns(nil, nil)

			err = spectrumMMCLI.UnlinkFileset(filesystem, fileset)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	//TODO when listfilesets is implemented
	Context(".ListFilesets", func() {
		It("", func() {})
	})

	Context(".ListFileset", func() {
		It("should fail when execute command errors", func() {
			errorMsg := fmt.Sprintf("error executing command")
			fakeExec.ExecuteReturns(nil, fmt.Errorf(errorMsg))

			volume, err := spectrumMMCLI.ListFileset(filesystem, fileset)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(errorMsg))
			Expect(volume).To(Equal(resources.Volume{}))
		})

		It("should succeed when execute command does not error", func() {
			fakeExec.ExecuteReturns(nil, nil)

			volume, err := spectrumMMCLI.ListFileset(filesystem, fileset)
			Expect(err).ToNot(HaveOccurred())
			Expect(volume.Name).To(Equal(fileset))
		})
	})

	Context(".ListFilesetQuota", func() {
		It("should fail when execute command errors", func() {
			errorMsg := fmt.Sprintf("error executing command")
			fakeExec.ExecuteReturns(nil, fmt.Errorf(errorMsg))

			quota, err := spectrumMMCLI.ListFilesetQuota(filesystem, fileset)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(fmt.Sprintf("Failed to list quota for fileset %s: %s", fileset, errorMsg)))
			Expect(quota).To(Equal(""))
		})

		It("should fail when execute command returns an invalid output", func() {
			returnMsg := "a"
			fakeExec.ExecuteReturns([]byte(returnMsg), nil)

			quota, err := spectrumMMCLI.ListFilesetQuota(filesystem, fileset)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(fmt.Sprintf("Error listing quota for fileset %s", fileset)))
			Expect(quota).To(Equal(""))
		})

		It("should fail when execute command returns an invalid output", func() {
			returnMsg := "a \n b \n c"
			fakeExec.ExecuteReturns([]byte(returnMsg), nil)

			quota, err := spectrumMMCLI.ListFilesetQuota(filesystem, fileset)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(fmt.Sprintf("Error parsing tokens while listing quota for fileset %s", fileset)))
			Expect(quota).To(Equal(""))
		})

		It("should succeed when execute command returns a valid output", func() {
			returnMsg := "a \n b \n1 2 3 fake-quota 5 6 fake-filesystem 8 fake-mountpoint 10 11 fake-host\n 12:21:32"
			fakeExec.ExecuteReturns([]byte(returnMsg), nil)

			quota, err := spectrumMMCLI.ListFilesetQuota(filesystem, fileset)
			Expect(err).ToNot(HaveOccurred())
			Expect(quota).To(Equal("fake-quota"))
		})
	})

	Context(".SetFilesetQuota", func() {
		It("should fail when execute command errors", func() {
			errorMsg := fmt.Sprintf("error executing command")
			fakeExec.ExecuteReturns(nil, fmt.Errorf(errorMsg))

			err = spectrumMMCLI.SetFilesetQuota(filesystem, fileset, "fake-quota")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(fmt.Sprintf("Failed to set quota 'fake-quota' for fileset '%s': %s", fileset, errorMsg)))
		})

		It("should succeed when execute command does not error", func() {
			fakeExec.ExecuteReturns(nil, nil)

			err := spectrumMMCLI.SetFilesetQuota(filesystem, fileset, "fake-quota")
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
