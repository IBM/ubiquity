package web_server_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestCFStorageHandlers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CF-Storage Broker Web Server Handlers Suite")
}
