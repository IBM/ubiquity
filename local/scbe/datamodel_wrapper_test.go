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

package scbe_test

import (
  . "github.com/onsi/ginkgo"
  . "github.com/onsi/gomega"
	"github.com/IBM/ubiquity/database"
	"github.com/IBM/ubiquity/local/scbe"
)

var _ = Describe("ScbeDataModelWrapper test", func() {

	var (
		dataModelWrapper   scbe.ScbeDataModelWrapper
		err                error
		volumeName         string = "volumeName"
		volumeWwn          string = "wwn"
		volumeFsType       string = "volumeFsType"
		volumeNameDb       string = volumeName + database.VolumeNameSuffix
		volumeWwnDb        string = volumeWwn + database.VolumeNameSuffix
		volumeFsTypeDb     string = volumeFsType + database.VolumeNameSuffix
		scbeVolume         scbe.ScbeVolume
	)

	BeforeEach(func() {
		dataModelWrapper = scbe.NewScbeDataModelWrapper()
	})
	AfterEach(func() {
		database.UnregisterAllMigrations()
	})

	Context("Database cannot be accessed yet", func() {
        Context("InsertVolume", func() {
            It("succeed for db volume", func() {
                defer database.InitTestError()()
                err = dataModelWrapper.InsertVolume(volumeNameDb, volumeWwnDb, volumeFsTypeDb)
                Expect(err).To(Not(HaveOccurred()))
                scbeVolume, err = dataModelWrapper.GetVolume(volumeNameDb, true)
                Expect(err).To(Not(HaveOccurred()))
            })
            It("fail for non db volume", func() {
                defer database.InitTestError()()
                err = dataModelWrapper.InsertVolume(volumeName, volumeWwn, volumeFsType)
                Expect(err).To(HaveOccurred())
                scbeVolume, err = dataModelWrapper.GetVolume(volumeName, true)
                Expect(err).To(HaveOccurred())
            })
        })
        Context("DeleteVolume", func() {
            It("succeed for db volume", func() {
                defer database.InitTestError()()
                err = dataModelWrapper.InsertVolume(volumeNameDb, volumeWwnDb, volumeFsTypeDb)
                Expect(err).To(Not(HaveOccurred()))
                scbeVolume, err = dataModelWrapper.GetVolume(volumeNameDb, true)
                Expect(err).To(Not(HaveOccurred()))
                err = dataModelWrapper.DeleteVolume(volumeNameDb)
                Expect(err).To(Not(HaveOccurred()))
                scbeVolume, err = dataModelWrapper.GetVolume(volumeNameDb, false)
                Expect(err).To(Not(HaveOccurred()))
            })
            It("succeed if db volume does not exist", func() {
				err = dataModelWrapper.DeleteVolume(volumeNameDb)
                Expect(err).To(Not(HaveOccurred()))
            })
        })
        Context("UpdateDatabaseVolume", func() {
            It("succeed", func() {
                defer database.InitTestError()()
                err = dataModelWrapper.InsertVolume(volumeNameDb, volumeWwnDb, volumeFsTypeDb)
                Expect(err).To(Not(HaveOccurred()))
                scbeVolume, err = dataModelWrapper.GetVolume(volumeNameDb, true)
                Expect(err).To(Not(HaveOccurred()))
                err = dataModelWrapper.DeleteVolume(volumeNameDb)
                Expect(err).To(Not(HaveOccurred()))
                dataModelWrapper.UpdateDatabaseVolume(&scbeVolume)
                scbeVolume, err = dataModelWrapper.GetVolume(volumeNameDb, true)
                Expect(err).To(Not(HaveOccurred()))
                dataModelWrapper.UpdateDatabaseVolume(nil)
                scbeVolume, err = dataModelWrapper.GetVolume(volumeNameDb, false)
                Expect(err).To(Not(HaveOccurred()))
            })
        })
    })
})

func isEqualScbeVolumes(list1 []scbe.ScbeVolume, list2 []scbe.ScbeVolume) bool {
	if len(list1) != len(list2) {
		return false
	}

	for i, v := range list1 {
		if !isEqualScbeVolume(v, list2[i]) {
			return false
		}
	}

	return true
}

func isEqualScbeVolume(vol1 scbe.ScbeVolume, vol2 scbe.ScbeVolume) bool {
	return vol1.Volume.Name == vol2.Volume.Name && vol1.WWN == vol2.WWN
}
