package local_test

import (
	"testing"
	"github.com/IBM/ubiquity/utils"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestLocal(t *testing.T) {
	RegisterFailHandler(Fail)
	defer utils.InitUbiquityServerTestLogger()()
	RunSpecs(t, "Local Suite")
}
