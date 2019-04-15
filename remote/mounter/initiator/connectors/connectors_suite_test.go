package connectors_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/IBM/ubiquity/utils"
)

func TestConnectors(t *testing.T) {
	RegisterFailHandler(Fail)
	defer utils.InitUbiquityServerTestLogger()()
	RunSpecs(t, "Connectors Suite")
}
