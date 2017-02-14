package local_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestLocalStorage(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Storage Local Test Suite")
}
