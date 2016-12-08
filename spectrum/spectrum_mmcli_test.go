package spectrum_test

import (
	"fmt"
	"log"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.ibm.com/almaden-containers/ubiquity/fakes"
	"github.ibm.com/almaden-containers/ubiquity/spectrum"
)

var _ = Describe("spectrum_mmcli", func() {
	var (
		spectrumMMCLI spectrum.Spectrum
		logger        *log.Logger
		fakeExec      *fakes.FakeExecutor
		opts          map[string]interface{}
		err           error
	)
	BeforeEach(func() {
		logger = log.New(os.Stdout, "spectrum: ", log.Lshortfile|log.LstdFlags)
		fakeExec = new(fakes.FakeExecutor)
		spectrumMMCLI, err = spectrum.NewSpectrumMMCLIWithExecutor(logger, fakeExec, opts)
		Expect(err).ToNot(HaveOccurred())

	})

	Context(".GetClusterID", func() {

		It("should fail when execute command errors", func() {
			fakeExec.ExecuteReturns(nil, fmt.Errorf("failed to execute command"))
			_, err := spectrumMMCLI.GetClusterId()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("failed to execute command"))
		})

		It("should fail when execute command returns a non expected value", func() {
			fakeExec.ExecuteReturns(nil, nil)
			_, err := spectrumMMCLI.GetClusterId()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Error determining cluster id"))
		})

		It("should fail when execute command returns a wrong number of info", func() {
			stringOutput := "a \n GPFS cluster id: 123 \n"
			fakeExec.ExecuteReturns([]byte(stringOutput), nil)
			_, err := spectrumMMCLI.GetClusterId()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Error determining cluster id"))
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
			_, err := spectrumMMCLI.IsFilesystemMounted("fake-filesystem")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("failed to execute command"))
		})
	})
})
