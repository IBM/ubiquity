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

package mounter_test

import (
	"github.com/IBM/ubiquity/utils/logs"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"testing"
	"github.com/IBM/ubiquity/remote/mounter"
	"github.com/IBM/ubiquity/resources"
)

var _ = Describe("mounter_factory_test", func() {
	var (
		mounterFactory mounter.MounterFactory
	)

	BeforeEach(func() {
		mounterFactory = mounter.NewMounterFactory()
	})

	Context(".GetMounterPerBackend", func() {
		It("should fail due to no backend type found", func() {
			_, err := mounterFactory.GetMounterPerBackend(
				"fakebackend type",
				nil,
				resources.UbiquityPluginConfig{},
			)
			Expect(err).To(HaveOccurred())
			_, ok := err.(*mounter.NoMounterForVolumeError)
			Expect(ok).To(Equal(true))
		})
		It("should succeed get scbe backend", func() {
			backendMounter, err := mounterFactory.GetMounterPerBackend(
				resources.SCBE,
				nil,
				resources.UbiquityPluginConfig{},
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(backendMounter).NotTo(Equal(nil))
		})
	})
})

func TestMounterFactory(t *testing.T) {
	RegisterFailHandler(Fail)
	defer logs.InitStdoutLogger(logs.DEBUG)()
	RunSpecs(t, "MountFactory Test Suite")
}
