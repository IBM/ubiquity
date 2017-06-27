package utils_test

import (
	"github.com/IBM/ubiquity/utils/logs"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestLocalStorage(t *testing.T) {
	RegisterFailHandler(Fail)
	defer logs.InitStdoutLogger(logs.DEBUG)()
	RunSpecs(t, "Utils Test Suite")
}
