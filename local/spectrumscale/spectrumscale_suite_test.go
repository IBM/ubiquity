package spectrumscale_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestSpectrum(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Spectrum Test Suite")
}
