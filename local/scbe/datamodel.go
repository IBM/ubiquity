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

package scbe

import (
	"fmt"
	"github.com/IBM/ubiquity/model"
	"github.com/IBM/ubiquity/resources"
	"github.com/IBM/ubiquity/utils/logs"
	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
)

//go:generate counterfeiter -o ../../fakes/fake_ScbeDataModel.go . ScbeDataModel
type ScbeDataModel interface {
	CreateVolumeTable() error
	DeleteVolume(name string) error
	InsertVolume(volumeName string, wwn string, attachTo string, fstype string) error
	GetVolume(name string) (ScbeVolume, bool, error)
	ListVolumes() ([]ScbeVolume, error)
	UpdateVolumeAttachTo(volumeName string, scbeVolume ScbeVolume, host2attach string) error
}

type scbeDataModel struct {
	logger   logs.Logger
	database *gorm.DB
	backend  string
}

type ScbeVolume struct {
	ID       uint
	Volume   resources.Volume
	VolumeID uint
	WWN      string
	AttachTo string
	FSType   string
}

func NewScbeDataModel(db *gorm.DB, backend string) ScbeDataModel {
	return &scbeDataModel{logger: logs.GetLogger(), database: db, backend: backend}
}

// CreateVolumeTable create the SCBE backend table
func (d *scbeDataModel) CreateVolumeTable() error {
	defer d.logger.Trace(logs.DEBUG)()

	if err := d.database.AutoMigrate(&ScbeVolume{}).Error; err != nil {
		return d.logger.ErrorRet(err, "failed")
	}
	return nil
}

// DeleteVolume if vol exist in DB then delete it (both in the generic table and the specific one)
func (d *scbeDataModel) DeleteVolume(name string) error {
	defer d.logger.Trace(logs.DEBUG)()

	volume, exists, err := d.GetVolume(name)

	if err != nil {
		return err
	}
	if exists == false {
		return d.logger.ErrorRet(&volumeNotFoundError{name}, "failed")
	}

	if err := d.database.Delete(&volume).Error; err != nil {
		return d.logger.ErrorRet(err, "database.Delete failed")
	}

	if err := model.DeleteVolume(d.database, &volume.Volume).Error; err != nil {
		return d.logger.ErrorRet(err, "model.DeleteVolume failed")
	}
	return nil
}

// InsertVolume volume name and its details given in opts
func (d *scbeDataModel) InsertVolume(volumeName string, wwn string, attachTo string, fstype string) error {
	defer d.logger.Trace(logs.DEBUG)()

	volume := ScbeVolume{
		Volume: resources.Volume{Name: volumeName,
			Backend: fmt.Sprintf("%s", d.backend)},
		WWN:      wwn,
		AttachTo: attachTo,
		FSType:   fstype,
	}

	if err := d.database.Create(&volume).Error; err != nil {
		return d.logger.ErrorRet(err, "database.Create failed")
	}
	return nil
}

// GetVolume return ScbeVolume if exist in DB, else return false and err
func (d *scbeDataModel) GetVolume(name string) (ScbeVolume, bool, error) {
	defer d.logger.Trace(logs.DEBUG)()

	volume, err := model.GetVolume(d.database, name, d.backend)
	if err != nil {
		if err.Error() == "record not found" {
			return ScbeVolume{}, false, nil
		}
		return ScbeVolume{}, false, d.logger.ErrorRet(err, "model.GetVolume failed")

	}

	var scbeVolume ScbeVolume
	if err := d.database.Where("volume_id = ?", volume.ID).Preload("Volume").First(&scbeVolume).Error; err != nil {
		if err.Error() == "record not found" {
			return ScbeVolume{}, false, nil
		}
		return ScbeVolume{}, false, d.logger.ErrorRet(err, "failed")
	}
	return scbeVolume, true, nil
}

func (d *scbeDataModel) ListVolumes() ([]ScbeVolume, error) {
	defer d.logger.Trace(logs.DEBUG)()

	var volumesInDb []ScbeVolume
	if err := d.database.Preload("Volume").Find(&volumesInDb).Error; err != nil {
		return nil, d.logger.ErrorRet(err, "failed")
	}
	// hack: to be replaced by proper DB filtering (joins)
	var volumes []ScbeVolume
	for _, volume := range volumesInDb {
		if volume.Volume.Backend == d.backend {
			volumes = append(volumes, volume)
		}
	}

	return volumes, nil
}
func (d *scbeDataModel) UpdateVolumeAttachTo(volumeName string, scbeVolume ScbeVolume, host2attach string) error {
	defer d.logger.Trace(logs.DEBUG)()

	err := d.database.Table("scbe_volumes").Where("volume_id = ?", scbeVolume.ID).Update("attach_to", host2attach).Error
	if err != nil {
		return d.logger.ErrorRet(err, "failed", logs.Args{{"volumeName", volumeName}})
	}
	return nil
}
