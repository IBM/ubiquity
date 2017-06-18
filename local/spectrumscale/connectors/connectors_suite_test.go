package connectors_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"

	"github.com/jarcoal/httpmock"
)

func TestConnectors(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Connectors Suite")
}

var _ = BeforeSuite(func() {
	// block all HTTP requests
	httpmock.Activate()
})

var _ = AfterSuite(func() {
	httpmock.DeactivateAndReset()
})
