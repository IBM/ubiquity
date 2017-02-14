package service_broker_test

import (
	"fmt"
	"io"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestCFStorageBroker(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ibm-storage-broker Broker Local Test Suite")
}

type errCloser struct{ io.Reader }

func (errCloser) Close() error                     { return nil }
func (errCloser) Read(p []byte) (n int, err error) { return 0, fmt.Errorf("any") }
