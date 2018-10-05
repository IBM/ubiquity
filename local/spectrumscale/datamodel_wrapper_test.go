/**
 * Copyright 2018 IBM Corp.
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

package spectrumscale_test

import (
  . "github.com/onsi/ginkgo"
  . "github.com/onsi/gomega"
	"github.com/IBM/ubiquity/database"
	"github.com/IBM/ubiquity/local/spectrumscale"
)

var _ = Describe("SpectrumDataModelWrapper test", func() {

	var (
		dataModelWrapper   spectrumscale.SpectrumDataModelWrapper
		err                error
		ispreexists		   bool
		volumeName         string = "volumeName"
		fileset            string = "fileset"
		filesystem         string = "gpfs"
		volumeNameDb       string = volumeName + database.VolumeNameSuffix
		spectrumscaleVolume spectrumscale.SpectrumScaleVolume
		opts				map[string]interface{}
		backend				string = "spectrum-scale"
		quota				string = "2Gi"
	)

	BeforeEach(func() {
		dataModelWrapper = spectrumscale.NewSpectrumDataModelWrapper(backend)
		opts = make(map[string]interface{})
	})
	AfterEach(func() {
		database.UnregisterAllMigrations()
	})

	Context("Database cannot be accessed yet", func() {
        Context("InsertFilesetVolume", func() {
            It("succeed for db volume", func() {
				opts := make(map[string]interface{})
                err = dataModelWrapper.InsertFilesetVolume(fileset, volumeNameDb,filesystem, false, opts)
                Expect(err).To(Not(HaveOccurred()))
                spectrumscaleVolume, ispreexists, err = dataModelWrapper.GetVolume(volumeNameDb)
                Expect(err).To(Not(HaveOccurred()))
            })
            It("fail for non db volume", func() {
                defer database.InitTestError()()
                err = dataModelWrapper.InsertFilesetVolume(fileset, volumeName,filesystem, false, opts)
                Expect(err).To(HaveOccurred())
				spectrumscaleVolume, ispreexists, err = dataModelWrapper.GetVolume(volumeName)
                Expect(err).To(HaveOccurred())
            })
        })

        Context("InsertFilesetQuotaVolume", func() {
            It("succeed for db volume", func() {
				opts := make(map[string]interface{})
                err = dataModelWrapper.InsertFilesetQuotaVolume(fileset, quota, volumeNameDb,filesystem, false, opts)
                Expect(err).To(Not(HaveOccurred()))
                spectrumscaleVolume, ispreexists, err = dataModelWrapper.GetVolume(volumeNameDb)
                Expect(err).To(Not(HaveOccurred()))
            })
            It("fail for non db volume", func() {
                defer database.InitTestError()()
                err = dataModelWrapper.InsertFilesetQuotaVolume(fileset, quota, volumeName,filesystem, false, opts)
                Expect(err).To(HaveOccurred())
				spectrumscaleVolume, ispreexists, err = dataModelWrapper.GetVolume(volumeName)
                Expect(err).To(HaveOccurred())
            })
        })
        Context("DeleteVolume", func() {
            It("succeed for db volume", func() {
                defer database.InitTestError()()
                err = dataModelWrapper.InsertFilesetVolume(fileset, volumeNameDb,filesystem, false, opts)
                Expect(err).To(Not(HaveOccurred()))
				spectrumscaleVolume, ispreexists, err = dataModelWrapper.GetVolume(volumeNameDb)
                Expect(err).To(Not(HaveOccurred()))
                err = dataModelWrapper.DeleteVolume(volumeNameDb)
                Expect(err).To(Not(HaveOccurred()))
				spectrumscaleVolume, ispreexists, err = dataModelWrapper.GetVolume(volumeNameDb)
                Expect(err).To(Not(HaveOccurred()))
            })
        })
        Context("UpdateDatabaseVolume", func() {
            It("succeed", func() {
                defer database.InitTestError()()
				err = dataModelWrapper.InsertFilesetVolume(fileset, volumeNameDb,filesystem, false, opts)
                Expect(err).To(Not(HaveOccurred()))
				spectrumscaleVolume, ispreexists, err = dataModelWrapper.GetVolume(volumeNameDb)
                Expect(err).To(Not(HaveOccurred()))
                err = dataModelWrapper.DeleteVolume(volumeNameDb)
                Expect(err).To(Not(HaveOccurred()))
                dataModelWrapper.UpdateDatabaseVolume(&spectrumscaleVolume)
				spectrumscaleVolume, ispreexists, err = dataModelWrapper.GetVolume(volumeNameDb)
                Expect(err).To(Not(HaveOccurred()))
                dataModelWrapper.UpdateDatabaseVolume(nil)
				spectrumscaleVolume, ispreexists, err = dataModelWrapper.GetVolume(volumeNameDb)
                Expect(err).To(Not(HaveOccurred()))
            })
        })
    })
})
