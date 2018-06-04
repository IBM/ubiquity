/**
 * Copyright 2017 IBM Corp.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package remote_test

import (
	"github.com/IBM/ubiquity/remote"
	"github.com/IBM/ubiquity/resources"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega" // including the whole package inside the file
	"log"
	"os"
	//"io/ioutil"
	"io/ioutil"
)

const fakeCert = "/tmp/fake_cert.crt"

var _ = Describe("initialize", func() {
	var (
		client resources.StorageClient
		err    error
	)
	BeforeEach(func() {
	})
	Context(".initTransport", func() {
		It("should fail with default (verify-full) and no certificates", func() {
			logger := log.New(os.Stdout, "ubiquity: ", log.Lshortfile|log.LstdFlags)
			fakeConfig := resources.UbiquityPluginConfig{}
			client, err = remote.NewRemoteClientSecure(logger, fakeConfig)
			Expect(err).To(HaveOccurred())
			_, ok := err.(*remote.SslModeFullVerifyWithoutCAfile)
			Expect(ok).To(Equal(true))
		})
		It("should fail if wrong ssl mode", func() {
			logger := log.New(os.Stdout, "ubiquity: ", log.Lshortfile|log.LstdFlags)
			fakeConfig := resources.UbiquityPluginConfig{}
			os.Setenv(resources.KeySslMode, "fake ssl mode")
			client, err = remote.NewRemoteClientSecure(logger, fakeConfig)
			os.Unsetenv(resources.KeyScbeSslMode)
			Expect(err).To(HaveOccurred())
			_, ok := err.(*remote.SslModeValueInvalid)
			Expect(ok).To(Equal(true))
		})
		It("fails if no certificate", func() {
			logger := log.New(os.Stdout, "ubiquity: ", log.Lshortfile|log.LstdFlags)
			fakeConfig := resources.UbiquityPluginConfig{}
			os.Remove(fakeCert)
			os.Setenv(remote.KeyVerifyCA, fakeCert)
			os.Setenv(resources.KeySslMode, resources.SslModeVerifyFull)
			client, err = remote.NewRemoteClientSecure(logger, fakeConfig)
			os.Unsetenv(remote.KeyVerifyCA)
			os.Setenv(resources.KeySslMode, "")
			Expect(err).To(HaveOccurred())
		})
		It("fails if certificate parse error", func() {
			logger := log.New(os.Stdout, "ubiquity: ", log.Lshortfile|log.LstdFlags)
			fakeConfig := resources.UbiquityPluginConfig{}
			os.Remove(fakeCert)
			err := ioutil.WriteFile(fakeCert, []byte("fake\n"), 0644)
			Expect(err).ToNot(HaveOccurred())
			os.Setenv(remote.KeyVerifyCA, fakeCert)
			os.Setenv(resources.KeySslMode, resources.SslModeVerifyFull)
			client, err = remote.NewRemoteClientSecure(logger, fakeConfig)
			os.Unsetenv(resources.KeySslMode)
			os.Setenv(remote.KeyVerifyCA, "")
			Expect(err).To(HaveOccurred())
		})
		It("should succeed with require ssl mode", func() {
			logger := log.New(os.Stdout, "ubiquity: ", log.Lshortfile|log.LstdFlags)
			fakeConfig := resources.UbiquityPluginConfig{}
			os.Setenv(resources.KeySslMode, resources.SslModeRequire)
			client, err = remote.NewRemoteClientSecure(logger, fakeConfig)
			os.Unsetenv(resources.KeySslMode)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
