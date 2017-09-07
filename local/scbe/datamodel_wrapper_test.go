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
	"github.com/IBM/ubiquity/resources"
	"os"
)

var _ = Describe("ScbeDataModelWrapper test", func() {

	var (
		dataModelWrapper   scbe.ScbeDataModelWrapper
		err                error
		sqliteDbPath       string = "/tmp/ScbeDataModelWrapper_test"
		volumeName         string = "volumeName"
		volumeWwn          string = "wwn"
	    volumeAttachTo     string = "host"
		volumeFsType       string = "volumeFsType"
		volumeNameDb       string = volumeName + resources.DatabseSuffix
		volumeWwnDb        string = volumeWwn + resources.DatabseSuffix
		volumeAttachToDb   string = volumeAttachTo + resources.DatabseSuffix
		volumeFsTypeDb     string = volumeFsType + resources.DatabseSuffix
		scbeVolume         scbe.ScbeVolume
        listVolumes        []scbe.ScbeVolume
	)

	BeforeEach(func() {
		dataModelWrapper = scbe.NewScbeDataModelWrapper()
		os.Remove(sqliteDbPath)
	})

	Context("Database cannot be accessed yet", func() {
        Context("InsertVolume", func() {
            It("succeed for db volume", func() {
                defer database.InitTestError()()
                err = dataModelWrapper.InsertVolume(volumeNameDb, volumeWwnDb, volumeAttachToDb, volumeFsTypeDb)
                Expect(err).To(Not(HaveOccurred()))
                scbeVolume, err = dataModelWrapper.GetVolume(volumeNameDb, true)
                Expect(err).To(Not(HaveOccurred()))
            })
            It("fail for non db volume", func() {
                defer database.InitTestError()()
                err = dataModelWrapper.InsertVolume(volumeName, volumeWwn, volumeAttachTo, volumeFsType)
                Expect(err).To(HaveOccurred())
                scbeVolume, err = dataModelWrapper.GetVolume(volumeName, true)
                Expect(err).To(HaveOccurred())
            })
        })
        Context("DeleteVolume", func() {
            It("succeed for db volume", func() {
                defer database.InitTestError()()
                err = dataModelWrapper.InsertVolume(volumeNameDb, volumeWwnDb, volumeAttachToDb, volumeFsTypeDb)
                Expect(err).To(Not(HaveOccurred()))
                scbeVolume, err = dataModelWrapper.GetVolume(volumeNameDb, true)
                Expect(err).To(Not(HaveOccurred()))
                err = dataModelWrapper.DeleteVolume(volumeNameDb)
                Expect(err).To(Not(HaveOccurred()))
                scbeVolume, err = dataModelWrapper.GetVolume(volumeNameDb, false)
                Expect(err).To(Not(HaveOccurred()))
            })
            It("fail for non db volume", func() {
                sqliteDbCloseFunc := database.InitSqlite(sqliteDbPath)
                err = dataModelWrapper.InsertVolume(volumeName, volumeWwn, volumeAttachTo, volumeFsType)
                Expect(err).To(Not(HaveOccurred()))
                scbeVolume, err = dataModelWrapper.GetVolume(volumeName, true)
                Expect(err).To(Not(HaveOccurred()))
                sqliteDbCloseFunc()
                defer database.InitTestError()()
                err = dataModelWrapper.DeleteVolume(volumeName)
                Expect(err).To(HaveOccurred())
            })
        })
        Context("UpdateDatabaseVolume", func() {
            It("succeed", func() {
                defer database.InitTestError()()
                err = dataModelWrapper.InsertVolume(volumeNameDb, volumeWwnDb, volumeAttachToDb, volumeFsTypeDb)
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
        Context("UpdateVolumeAttachTo", func() {
            It("succeed for db volume", func() {
                defer database.InitTestError()()
                err = dataModelWrapper.InsertVolume(volumeNameDb, volumeWwnDb, volumeAttachToDb, volumeFsTypeDb)
                Expect(err).To(Not(HaveOccurred()))
                scbeVolume, err = dataModelWrapper.GetVolume(volumeNameDb, true)
                Expect(err).To(Not(HaveOccurred()))
                Expect(scbeVolume.AttachTo).To(Equal(volumeAttachToDb))
                err = dataModelWrapper.UpdateVolumeAttachTo(volumeNameDb, scbeVolume, scbe.EmptyHost)
                Expect(err).To(Not(HaveOccurred()))
                scbeVolume, err = dataModelWrapper.GetVolume(volumeNameDb, true)
                Expect(err).To(Not(HaveOccurred()))
                Expect(scbeVolume.AttachTo).To(Equal(scbe.EmptyHost))
            })
            It("fail for non db volume", func() {
                sqliteDbCloseFunc := database.InitSqlite(sqliteDbPath)
                err = dataModelWrapper.InsertVolume(volumeName, volumeWwn, volumeAttachTo, volumeFsType)
                Expect(err).To(Not(HaveOccurred()))
                scbeVolume, err = dataModelWrapper.GetVolume(volumeName, true)
                Expect(err).To(Not(HaveOccurred()))
                sqliteDbCloseFunc()
                defer database.InitTestError()()
                err = dataModelWrapper.UpdateVolumeAttachTo(volumeName, scbeVolume, scbe.EmptyHost)
                Expect(err).To(HaveOccurred())
            })
        })
        Context("ListVolumes", func() {
            It("empty when there is no db volume", func() {
                sqliteDbCloseFunc := database.InitSqlite(sqliteDbPath)
                err = dataModelWrapper.InsertVolume(volumeName, volumeWwn, volumeAttachTo, volumeFsType)
                Expect(err).To(Not(HaveOccurred()))
                scbeVolume, err = dataModelWrapper.GetVolume(volumeName, true)
                Expect(err).To(Not(HaveOccurred()))
                sqliteDbCloseFunc()
                defer database.InitTestError()()
                var emptyList []scbe.ScbeVolume
                listVolumes, err = dataModelWrapper.ListVolumes()
                Expect(err).To(Not(HaveOccurred()))
                Expect(listVolumes).To(Equal(emptyList))
            })
            It("only db volume is returned when it exists", func() {
                sqliteDbCloseFunc := database.InitSqlite(sqliteDbPath)
                err = dataModelWrapper.InsertVolume(volumeName, volumeWwn, volumeAttachTo, volumeFsType)
                Expect(err).To(Not(HaveOccurred()))
                scbeVolume, err = dataModelWrapper.GetVolume(volumeName, true)
                Expect(err).To(Not(HaveOccurred()))
                sqliteDbCloseFunc()
                defer database.InitTestError()()
                err = dataModelWrapper.InsertVolume(volumeNameDb, volumeWwnDb, scbe.EmptyHost, volumeFsTypeDb)
                Expect(err).To(Not(HaveOccurred()))
                scbeVolume, err = dataModelWrapper.GetVolume(volumeNameDb, true)
                Expect(err).To(Not(HaveOccurred()))
                scbeVolumes := []scbe.ScbeVolume{scbeVolume}
                listVolumes, err = dataModelWrapper.ListVolumes()
                Expect(err).To(Not(HaveOccurred()))
                Expect(isEqualScbeVolumes(listVolumes, scbeVolumes)).To(Equal(true))
            })
        })
    })
	Context("Database can be accessed", func() {
		Context("InsertVolume", func() {
			It("succeed for db volume", func() {
				defer database.InitSqlite(sqliteDbPath)()
				err = dataModelWrapper.InsertVolume(volumeNameDb, volumeWwnDb, volumeAttachToDb, volumeFsTypeDb)
				Expect(err).To(Not(HaveOccurred()))
				scbeVolume, err = dataModelWrapper.GetVolume(volumeNameDb, true)
				Expect(err).To(Not(HaveOccurred()))
			})
			It("succeed for non db volume", func() {
				defer database.InitSqlite(sqliteDbPath)()
				err = dataModelWrapper.InsertVolume(volumeName, volumeWwn, volumeAttachTo, volumeFsType)
				Expect(err).To(Not(HaveOccurred()))
				scbeVolume, err = dataModelWrapper.GetVolume(volumeName, true)
				Expect(err).To(Not(HaveOccurred()))
			})
		})
		Context("DeleteVolume", func() {
			It("succeed for db volume", func() {
				defer database.InitSqlite(sqliteDbPath)()
				err = dataModelWrapper.InsertVolume(volumeNameDb, volumeWwnDb, volumeAttachToDb, volumeFsTypeDb)
				Expect(err).To(Not(HaveOccurred()))
				scbeVolume, err = dataModelWrapper.GetVolume(volumeNameDb, true)
				Expect(err).To(Not(HaveOccurred()))
				err = dataModelWrapper.DeleteVolume(volumeNameDb)
				Expect(err).To(Not(HaveOccurred()))
				scbeVolume, err = dataModelWrapper.GetVolume(volumeNameDb, false)
				Expect(err).To(Not(HaveOccurred()))
			})
			It("succeed for non db volume", func() {
				defer database.InitSqlite(sqliteDbPath)()
				err = dataModelWrapper.InsertVolume(volumeName, volumeWwn, volumeAttachTo, volumeFsType)
				Expect(err).To(Not(HaveOccurred()))
				scbeVolume, err = dataModelWrapper.GetVolume(volumeName, true)
				Expect(err).To(Not(HaveOccurred()))
				err = dataModelWrapper.DeleteVolume(volumeName)
				Expect(err).To(Not(HaveOccurred()))
				scbeVolume, err = dataModelWrapper.GetVolume(volumeName, false)
				Expect(err).To(Not(HaveOccurred()))
			})
		})
        Context("UpdateVolumeAttachTo", func() {
            It("succeed for non db volume", func() {
                defer database.InitSqlite(sqliteDbPath)()
                err = dataModelWrapper.InsertVolume(volumeNameDb, volumeWwnDb, volumeAttachToDb, volumeFsTypeDb)
                Expect(err).To(Not(HaveOccurred()))
                scbeVolume, err = dataModelWrapper.GetVolume(volumeNameDb, true)
                Expect(err).To(Not(HaveOccurred()))
                Expect(scbeVolume.AttachTo).To(Equal(volumeAttachToDb))
                err = dataModelWrapper.UpdateVolumeAttachTo(volumeNameDb, scbeVolume, scbe.EmptyHost)
                Expect(err).To(Not(HaveOccurred()))
                scbeVolume, err = dataModelWrapper.GetVolume(volumeNameDb, true)
                Expect(err).To(Not(HaveOccurred()))
                Expect(scbeVolume.AttachTo).To(Equal(scbe.EmptyHost))
            })
            It("succeed for non db volume", func() {
                defer database.InitSqlite(sqliteDbPath)()
                err = dataModelWrapper.InsertVolume(volumeName, volumeWwn, volumeAttachTo, volumeFsType)
                Expect(err).To(Not(HaveOccurred()))
                scbeVolume, err = dataModelWrapper.GetVolume(volumeName, true)
                Expect(err).To(Not(HaveOccurred()))
                Expect(scbeVolume.AttachTo).To(Equal(volumeAttachTo))
                err = dataModelWrapper.UpdateVolumeAttachTo(volumeName, scbeVolume, scbe.EmptyHost)
                Expect(err).To(Not(HaveOccurred()))
                scbeVolume, err = dataModelWrapper.GetVolume(volumeName, true)
                Expect(err).To(Not(HaveOccurred()))
                Expect(scbeVolume.AttachTo).To(Equal(scbe.EmptyHost))
            })
        })
        Context("ListVolumes", func() {
            It("empty when there is no volume", func() {
                defer database.InitSqlite(sqliteDbPath)()
                var emptyList []scbe.ScbeVolume
                listVolumes, err = dataModelWrapper.ListVolumes()
                Expect(err).To(Not(HaveOccurred()))
                Expect(listVolumes).To(Equal(emptyList))
            })
            It("only db volume is returned if only db volume exists", func() {
                defer database.InitSqlite(sqliteDbPath)()
                err = dataModelWrapper.InsertVolume(volumeNameDb, volumeWwnDb, volumeAttachToDb, volumeFsTypeDb)
                Expect(err).To(Not(HaveOccurred()))
                dbVolume, err := dataModelWrapper.GetVolume(volumeNameDb, true)
                Expect(err).To(Not(HaveOccurred()))
                scbeVolumes := []scbe.ScbeVolume{dbVolume}
                listVolumes, err = dataModelWrapper.ListVolumes()
                Expect(err).To(Not(HaveOccurred()))
                Expect(isEqualScbeVolumes(listVolumes, scbeVolumes)).To(Equal(true))
            })
            It("only non db volume is returned if only non db volume exists", func() {
                defer database.InitSqlite(sqliteDbPath)()
                err = dataModelWrapper.InsertVolume(volumeName, volumeWwn, volumeAttachTo, volumeFsType)
                Expect(err).To(Not(HaveOccurred()))
                nonDbVolume, err := dataModelWrapper.GetVolume(volumeName, true)
                Expect(err).To(Not(HaveOccurred()))
                scbeVolumes := []scbe.ScbeVolume{nonDbVolume}
                listVolumes, err = dataModelWrapper.ListVolumes()
                Expect(err).To(Not(HaveOccurred()))
                Expect(isEqualScbeVolumes(listVolumes, scbeVolumes)).To(Equal(true))
            })
            It("all volumes", func() {
                defer database.InitSqlite(sqliteDbPath)()
                err = dataModelWrapper.InsertVolume(volumeNameDb, volumeWwnDb, volumeAttachToDb, volumeFsTypeDb)
                Expect(err).To(Not(HaveOccurred()))
                dbVolume, err := dataModelWrapper.GetVolume(volumeNameDb, true)
                Expect(err).To(Not(HaveOccurred()))
                err = dataModelWrapper.InsertVolume(volumeName, volumeWwn, volumeAttachTo, volumeFsType)
                Expect(err).To(Not(HaveOccurred()))
                scbeVolume, err = dataModelWrapper.GetVolume(volumeName, true)
                scbeVolumes := []scbe.ScbeVolume{scbeVolume, dbVolume}
                listVolumes, err = dataModelWrapper.ListVolumes()
                Expect(err).To(Not(HaveOccurred()))
                Expect(isEqualScbeVolumes(listVolumes, scbeVolumes)).To(Equal(true))
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
    return vol1.Volume.Name == vol2.Volume.Name && vol1.WWN == vol2.WWN && vol1.AttachTo == vol2.AttachTo
}