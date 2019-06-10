/* Copyright 2018 IBM Corp.
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
        "github.com/IBM/ubiquity/resources"
        "github.com/IBM/ubiquity/utils/logs"
        "github.com/IBM/ubiquity/database"
)

type SpectrumDataModelWrapper interface {
	DeleteVolume(name string) error
	InsertFilesetVolume(fileset, volumeName string, filesystem string, isPreexisting bool, opts map[string]interface{}) error
	InsertFilesetQuotaVolume(fileset, quota, volumeName string, filesystem string, isPreexisting bool, opts map[string]interface{}) error
	GetVolume(name string) (SpectrumScaleVolume, bool, error)
	ListVolumes() ([]resources.Volume, error)
	UpdateDatabaseVolume(newVolume *SpectrumScaleVolume)
	IsDbVolume(name string) bool
	GetDbName() string 
}

type spectrumDataModelWrapper struct {
	logger logs.Logger
	dbVolume *SpectrumScaleVolume
	backend string
}

func NewSpectrumDataModelWrapper(backend string) SpectrumDataModelWrapper {
	database.RegisterMigration(resources.Volume{})
	database.RegisterMigration(&SpectrumScaleVolume{})
	return &spectrumDataModelWrapper{logger: logs.GetLogger(), backend: backend}
}

func (d *spectrumDataModelWrapper) GetDbName() string {
        d.logger.Debug("DB volume Name:", logs.Args{{"DbVolume:",database.VolumeNameSuffix}})
        return database.VolumeNameSuffix
}


func (d *spectrumDataModelWrapper) IsDbVolume(name string) bool {
	if database.IsDatabaseVolume(name) {
		return true
	}
	return false
}

func (d *spectrumDataModelWrapper) UpdateDatabaseVolume(newVolume *SpectrumScaleVolume) {
        defer d.logger.Trace(logs.DEBUG)()
        d.logger.Debug("", logs.Args{{"dbVolume", d.dbVolume}, {"newVolume", newVolume}})
        d.dbVolume = newVolume
}

func (d *spectrumDataModelWrapper) GetVolume(name string) (SpectrumScaleVolume, bool, error) {
	defer d.logger.Trace(logs.DEBUG)()
	var volume SpectrumScaleVolume
	var isExists bool
	var err error

	isExists = false
	volume = SpectrumScaleVolume{}

    if database.IsDatabaseVolume(name) {
		// work with memory object
		exists := d.dbVolume != nil
		if exists {
			volume = *d.dbVolume
			isExists = true
			d.logger.Debug("DEEBUG Wrapper GetVolume exists", logs.Args{{"isExists", isExists})
		}
	} else {
		// open db connection
		dbConnection := database.NewConnection()
		if err := dbConnection.Open(); err != nil {
			return SpectrumScaleVolume{}, false, d.logger.ErrorRet(err, "dbConnection.Open failed")
		}

		defer dbConnection.Close()
		dataModel := NewSpectrumDataModel(d.logger, dbConnection.GetDb(), d.backend)
		if volume, isExists, err = dataModel.GetVolume(name); err != nil {
			return SpectrumScaleVolume{}, false, d.logger.ErrorRet(err, "dataModel.GetVolume failed")
		}
    }
	d.logger.Debug("Got volume", logs.Args{{"VolumeName", name}, {"Volume Details",volume}})
	d.logger.Debug("DEEBUG Wrapper GetVolume", logs.Args{{"isExists", isExists})
	return volume, isExists, nil
}

func (d *spectrumDataModelWrapper) DeleteVolume(volName string) error {

	defer d.logger.Trace(logs.DEBUG)()
	var err error

	if database.IsDatabaseVolume(volName) {
		// sanity
		if d.dbVolume == nil {
			return d.logger.ErrorRet(&resources.VolumeNotFoundError{volName}, "failed")
		}
		// work with memory object
		d.UpdateDatabaseVolume(nil)
	} else {
		// open db connection
		dbConnection := database.NewConnection()
		if err = dbConnection.Open(); err != nil {
			return d.logger.ErrorRet(err, "dbConnection.Open failed")
		}

	   defer dbConnection.Close()
	   dataModel := NewSpectrumDataModel(d.logger, dbConnection.GetDb(), d.backend)
		if err = dataModel.DeleteVolume(volName); err != nil {
			return d.logger.ErrorRet(err, "dataModel.DeleteVolume failed")
		}
	}
	return nil
}

func (d *spectrumDataModelWrapper)addPermissionsForVolume(volume *SpectrumScaleVolume, opts map[string]interface{}) {

	if len(opts) > 0 {
		uid, uidSpecified := opts[UserSpecifiedUID]
		gid, gidSpecified := opts[UserSpecifiedGID]
		if uidSpecified && gidSpecified {
			volume.UID = uid.(string)
			volume.GID = gid.(string)

		}
    }
}

func (d *spectrumDataModelWrapper) InsertFilesetVolume(fileset string, volumeName string, filesystem string, isPreexisting bool, opts map[string]interface{}) error {
	defer d.logger.Trace(logs.DEBUG)()
	var err error

	if database.IsDatabaseVolume(volumeName) {
		if d.dbVolume != nil {
			return d.logger.ErrorRet(&resources.VolAlreadyExistsError{volumeName}, "failed")
		}
		volume := &SpectrumScaleVolume{Volume: resources.Volume{Name: volumeName, Backend: d.backend}, Type: Fileset, FileSystem: filesystem, Fileset: fileset, IsPreexisting: isPreexisting}
		d.addPermissionsForVolume(volume, opts)
		d.UpdateDatabaseVolume(volume)
	} else {
		dbConnection := database.NewConnection()
		if err = dbConnection.Open(); err != nil {
			return d.logger.ErrorRet(err, "dbConnection.Open failed")
		}

		defer dbConnection.Close()
		dataModel := NewSpectrumDataModel(d.logger, dbConnection.GetDb(), d.backend)
		if err = dataModel.InsertFilesetVolume(fileset, volumeName, filesystem, isPreexisting, opts); err != nil {
			return d.logger.ErrorRet(err, "dataModel.InsertFilesetVolume failed")
		}
    }
	return nil
}

func (d *spectrumDataModelWrapper) InsertFilesetQuotaVolume(fileset string, quota, volumeName string, filesystem string, isPreexisting bool, opts map[string]interface{}) error {
	defer d.logger.Trace(logs.DEBUG)()
	var err error

	if database.IsDatabaseVolume(volumeName) {
		if d.dbVolume != nil {
			return d.logger.ErrorRet(&resources.VolAlreadyExistsError{volumeName}, "failed")
		}
		volume := &SpectrumScaleVolume{Volume: resources.Volume{Name: volumeName, Backend: d.backend}, Type: FilesetWithQuota, FileSystem: filesystem, Fileset: fileset, Quota: quota, IsPreexisting: isPreexisting}
		d.addPermissionsForVolume(volume, opts)
		d.UpdateDatabaseVolume(volume)
	} else {
		dbConnection := database.NewConnection()
		if err = dbConnection.Open(); err != nil {
			return d.logger.ErrorRet(err, "dbConnection.Open failed")
		}

		defer dbConnection.Close()
		dataModel := NewSpectrumDataModel(d.logger, dbConnection.GetDb(), d.backend)
		if err = dataModel.InsertFilesetQuotaVolume(fileset, quota,volumeName, filesystem, isPreexisting, opts); err != nil {
			return d.logger.ErrorRet(err, "dataModel.InsertFilesetQuotaVolume failed")
		}
	}
	return nil
}

func (d *spectrumDataModelWrapper) ListVolumes() ([]resources.Volume, error) {
	defer d.logger.Trace(logs.DEBUG)()
	var err error
	var volumes []resources.Volume
    // open db connection
    dbConnection := database.NewConnection()
    err = dbConnection.Open()
    if err == nil {
		defer dbConnection.Close()
		// list volumes
		dataModel := NewSpectrumDataModel(d.logger, dbConnection.GetDb(), d.backend)
		if volumes, err = dataModel.ListVolumes(); err != nil {
			return nil, d.logger.ErrorRet(err, "dataModel.ListVolumes failed")
		}
	}
	if d.dbVolume != nil {
		volumes = append(volumes,(*d.dbVolume).Volume)
	}
	return volumes, nil
}
