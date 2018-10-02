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

package spectrumscale

import (
	"fmt"
	"github.com/IBM/ubiquity/utils/logs"
	"github.com/IBM/ubiquity/model"
	"github.com/IBM/ubiquity/resources"
	"github.com/jinzhu/gorm"
)

//go:generate counterfeiter -o ../../fakes/fake_SpectrumDataModel.go . SpectrumDataModel
type SpectrumDataModel interface {
	CreateVolumeTable() error
	SetClusterId(string)
	GetClusterId() string
	DeleteVolume(name string) error
	InsertFilesetVolume(fileset, volumeName string, filesystem string, isPreexisting bool, opts map[string]interface{}) error
	InsertFilesetQuotaVolume(fileset, quota, volumeName string, filesystem string, isPreexisting bool, opts map[string]interface{}) error
	GetVolume(name string) (SpectrumScaleVolume, bool, error)
	ListVolumes() ([]resources.Volume, error)
	UpdateVolumeMountpoint(name string, mountpoint string) error
}

type spectrumDataModel struct {
	log       logs.Logger
	database  *gorm.DB
	clusterId string
	backend   string
}

type VolumeType int

const (
	Fileset VolumeType = iota
	Lightweight
	FilesetWithQuota
)

const (
	UserSpecifiedUID string = "uid"
	UserSpecifiedGID string = "gid"
)

type SpectrumScaleVolume struct {
	ID            uint
	Volume        resources.Volume
	VolumeID      uint
	Type          VolumeType
	ClusterId     string
	FileSystem    string
	Fileset       string
	Directory     string
	UID           string
	GID           string
	Quota         string
	IsPreexisting bool
}

func NewSpectrumDataModel(log logs.Logger, db *gorm.DB, backend string) SpectrumDataModel {
	return &spectrumDataModel{log: log, database: db, backend: backend}
}

func (d *spectrumDataModel) SetClusterId(id string) {
	d.clusterId = id
}
func (d *spectrumDataModel) GetClusterId() string {
	return d.clusterId
}
func (d *spectrumDataModel) CreateVolumeTable() error {
	defer d.log.Trace(logs.DEBUG)()
	if err := d.database.AutoMigrate(&SpectrumScaleVolume{}).Error; err != nil {
		return err
	}
	return nil
}

func (d *spectrumDataModel) DeleteVolume(name string) error {
	defer d.log.Trace(logs.DEBUG)()
	volume, exists, err := d.GetVolume(name)

	if err != nil {
		return err
	}
	if exists == false {
		return fmt.Errorf("Volume : %s not found", name)
	}

	if err := d.database.Delete(&volume).Error; err != nil {
		return err
	}
	if err := model.DeleteVolume(d.database, &volume.Volume).Error; err != nil {
		return err
	}
	return nil
}

func (d *spectrumDataModel) InsertFilesetVolume(fileset, volumeName string, filesystem string, isPreexisting bool, opts map[string]interface{}) error {
	defer d.log.Trace(logs.DEBUG)()
	volume := SpectrumScaleVolume{Volume: resources.Volume{Name: volumeName, Backend: d.backend}, Type: Fileset, ClusterId: d.clusterId, FileSystem: filesystem,
		Fileset: fileset, IsPreexisting: isPreexisting}

	addPermissionsForVolume(&volume, opts)

	return d.insertVolume(volume)
}

func (d *spectrumDataModel) InsertFilesetQuotaVolume(fileset, quota, volumeName string, filesystem string, isPreexisting bool, opts map[string]interface{}) error {
	defer d.log.Trace(logs.DEBUG)()
	volume := SpectrumScaleVolume{Volume: resources.Volume{Name: volumeName, Backend: d.backend}, Type: FilesetWithQuota, ClusterId: d.clusterId, FileSystem: filesystem,
		Fileset: fileset, Quota: quota, IsPreexisting: isPreexisting}

	addPermissionsForVolume(&volume, opts)

	return d.insertVolume(volume)
}

func (d *spectrumDataModel) insertVolume(volume SpectrumScaleVolume) error {
	defer d.log.Trace(logs.DEBUG)()
	if err := d.database.Create(&volume).Error; err != nil {
		return err
	}
	return nil
}

func (d *spectrumDataModel) GetVolume(name string) (SpectrumScaleVolume, bool, error) {
	defer d.log.Trace(logs.DEBUG)()
	volume, err := model.GetVolume(d.database, name, d.backend)
	if err != nil {
		if err.Error() == "record not found" {
			return SpectrumScaleVolume{}, false, nil
		}
		return SpectrumScaleVolume{}, false, err
	}

	var spectrumVolume SpectrumScaleVolume
	if err := d.database.Where("volume_id = ?", volume.ID).Preload("Volume").First(&spectrumVolume).Error; err != nil {
		if err.Error() == "record not found" {
			return SpectrumScaleVolume{}, false, nil
		}
		return SpectrumScaleVolume{}, false, err
	}
	return spectrumVolume, true, nil
}

func (d *spectrumDataModel) ListVolumes() ([]resources.Volume, error) {
	defer d.log.Trace(logs.DEBUG)()
	var volumesInDb []SpectrumScaleVolume
	if err := d.database.Preload("Volume").Find(&volumesInDb).Error; err != nil {
		return nil, err
	}
	// hack: to be replaced by proper DB filtering (joins)
	var volumes []resources.Volume
	for _, volume := range volumesInDb {
		if volume.Volume.Backend == d.backend {
			volumes = append(volumes, volume.Volume)
		}
	}

	return volumes, nil
}

func (d *spectrumDataModel) UpdateVolumeMountpoint(name string, mountpoint string) error {
	defer d.log.Trace(logs.DEBUG)()
	volume, err := model.GetVolume(d.database, name, d.backend)
	if err != nil {
		return err
	}

	if err = model.UpdateVolumeMountpoint(d.database, &volume, mountpoint); err != nil {
		return fmt.Errorf("Error updating mountpoint of volume %s to %s: %s", volume.Name, mountpoint, err.Error())
	}
	return nil
}

func addPermissionsForVolume(volume *SpectrumScaleVolume, opts map[string]interface{}) {

	if len(opts) > 0 {
		uid, uidSpecified := opts[UserSpecifiedUID]
		gid, gidSpecified := opts[UserSpecifiedGID]

		if uidSpecified && gidSpecified {
			volume.UID = uid.(string)
			volume.GID = gid.(string)

		}
	}
}
