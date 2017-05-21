package scbe_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	//	httpmock "gopkg.in/jarcoal/httpmock.v1"
	"testing"

	"github.com/jarcoal/httpmock"
)

func TestSpectrum(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "SCBE Test Suite")
}

var _ = BeforeSuite(func() {
	// block all HTTP requests
	httpmock.Activate()
})

var _ = AfterSuite(func() {
	httpmock.DeactivateAndReset()
})
