package scbe_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/IBM/ubiquity/logutil"
)

func TestScbe(t *testing.T) {
	RegisterFailHandler(Fail)
	defer logutil.InitStdoutLogger(logutil.DEBUG)()
	RunSpecs(t, "SCBE Test Suite")
}

var _ = BeforeSuite(func() {
	// block all HTTP requests
	httpmock.Activate()
})

var _ = AfterSuite(func() {
	httpmock.DeactivateAndReset()
})
