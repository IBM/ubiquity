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
	"github.com/IBM/ubiquity/resources"
	"github.com/IBM/ubiquity/utils/logs"
	"strings"
	"github.com/IBM/ubiquity/database"
)

//go:generate counterfeiter -o ../../fakes/fake_ScbeDataModelWrapper.go . ScbeDataModelWrapper
type ScbeDataModelWrapper interface {
	GetVolume(name string, mustExist bool) (ScbeVolume, error)
	DeleteVolume(name string) error
	InsertVolume(volumeName string, wwn string, attachTo string, fstype string) error
	ListVolumes() ([]ScbeVolume, error)
	UpdateVolumeAttachTo(volumeName string, scbeVolume ScbeVolume, host2attach string) error
	UpdateDatabaseVolume(newVolume *ScbeVolume)
}

type scbeDataModelWrapper struct {
	logger   logs.Logger
	isTableCreated bool
	dbVolume *ScbeVolume
}

func NewScbeDataModelWrapper() ScbeDataModelWrapper {
	return &scbeDataModelWrapper{logger: logs.GetLogger()}
}

func isDatabaseVolume(volName string) bool {
	defer logs.GetLogger().Trace(logs.DEBUG)()
	isDatabaseVolume := strings.HasSuffix(volName, resources.DatabseSuffix)
	logs.GetLogger().Debug("", logs.Args{{volName, isDatabaseVolume}})
	return isDatabaseVolume
}

func (d *scbeDataModelWrapper) UpdateDatabaseVolume(newVolume *ScbeVolume) {
	defer d.logger.Trace(logs.DEBUG)()
	d.logger.Debug("", logs.Args{{"dbVolume", d.dbVolume}, {"newVolume", newVolume}})
	d.dbVolume = newVolume
}

func (d *scbeDataModelWrapper) createTableIfNeeded() error {
	defer d.logger.Trace(logs.DEBUG)()
	var err error

	// create table only if not created yet
	d.logger.Debug("check", logs.Args{{"isTableCreated", d.isTableCreated}})
	if !d.isTableCreated {

		// open db connection
		dbConnection := database.NewConnection()
		if err = dbConnection.Open(); err != nil {
			return d.logger.ErrorRet(err, "dbConnection.Open failed")
		}
		defer dbConnection.Close()

		// create volume table
		dataModel := NewScbeDataModel(dbConnection.GetDb())
		if err = dataModel.CreateVolumeTable(); err != nil {
			return d.logger.ErrorRet(err, "dataModel.CreateVolumeTable failed")
		}

		// mark table as created
		d.isTableCreated = true
		d.logger.Debug("set", logs.Args{{"isTableCreated", d.isTableCreated}})
 	}

	return nil
}

func (d *scbeDataModelWrapper) GetVolume(name string, mustExist bool) (ScbeVolume, error) {
	defer d.logger.Trace(logs.DEBUG)()
	var err error
	var volume ScbeVolume
	var exists bool

	if isDatabaseVolume(name) {

		// work with memory object
		exists = d.dbVolume != nil
		if exists {
			volume = *d.dbVolume
		}

	} else {

		// create table if needed
		if err = d.createTableIfNeeded(); err != nil {
			return ScbeVolume{}, d.logger.ErrorRet(err, "createTableIfNeeded failed")
		}

		// open db connection
		dbConnection := database.NewConnection()
		if err = dbConnection.Open(); err != nil {
			return ScbeVolume{}, d.logger.ErrorRet(err, "dbConnection.Open failed")
		}
		defer dbConnection.Close()

		// get volume
		dataModel := NewScbeDataModel(dbConnection.GetDb())
		if volume, exists, err = dataModel.GetVolume(name); err != nil {
			return ScbeVolume{}, d.logger.ErrorRet(err, "dataModel.GetVolume failed")
		}
	}

	// verify existence
	if mustExist != exists {
		if exists {
			err = &volAlreadyExistsError{name}
		} else {
			err = &volumeNotFoundError{name}
		}
		return ScbeVolume{}, d.logger.ErrorRet(err, "failed", logs.Args{{"mustExist", mustExist}, {"exists", exists}})
	}

	return volume, nil
}

func (d *scbeDataModelWrapper) DeleteVolume(name string) error {
	defer d.logger.Trace(logs.DEBUG)()
	var err error

	if isDatabaseVolume(name) {

		// sanity
		if d.dbVolume == nil {
			return d.logger.ErrorRet(&volumeNotFoundError{name}, "failed")
		}

		// work with memory object
		d.UpdateDatabaseVolume(nil)

	} else {

		// create table if needed
		if err = d.createTableIfNeeded(); err != nil {
			return d.logger.ErrorRet(err, "createTableIfNeeded failed")
		}

		// open db connection
		dbConnection := database.NewConnection()
		if err = dbConnection.Open(); err != nil {
			return d.logger.ErrorRet(err, "dbConnection.Open failed")
		}
		defer dbConnection.Close()

		// delete volume
		dataModel := NewScbeDataModel(dbConnection.GetDb())
		if err = dataModel.DeleteVolume(name); err != nil {
			return d.logger.ErrorRet(err, "dataModel.DeleteVolume failed")
		}
	}

	return nil
}

func (d *scbeDataModelWrapper) InsertVolume(volumeName string, wwn string, attachTo string, fstype string) error {
	defer d.logger.Trace(logs.DEBUG)()
	var err error

	if isDatabaseVolume(volumeName) {

		// sanity
		if d.dbVolume != nil {
			return d.logger.ErrorRet(&volAlreadyExistsError{volumeName}, "failed")
		}

		// work with memory object
		d.UpdateDatabaseVolume(&ScbeVolume{Volume: resources.Volume{Name: volumeName, Backend: resources.SCBE}, WWN: wwn, AttachTo: attachTo})

	} else {

		// create table if needed
		if err = d.createTableIfNeeded(); err != nil {
			return d.logger.ErrorRet(err, "createTableIfNeeded failed")
		}

		// open db connection
		dbConnection := database.NewConnection()
		if err = dbConnection.Open(); err != nil {
			return d.logger.ErrorRet(err, "dbConnection.Open failed")
		}
		defer dbConnection.Close()

		// insert volume
		dataModel := NewScbeDataModel(dbConnection.GetDb())
		if err = dataModel.InsertVolume(volumeName, wwn, attachTo, fstype); err != nil {
			return d.logger.ErrorRet(err, "dataModel.InsertVolume failed")
		}
	}

	return nil
}

func (d *scbeDataModelWrapper) ListVolumes() ([]ScbeVolume, error) {
	defer d.logger.Trace(logs.DEBUG)()
	var err error
	var volumes []ScbeVolume

	if d.dbVolume != nil {

		// get volumes from db if its attached
		if d.dbVolume.AttachTo != EmptyHost {

			// create table if needed
			if err = d.createTableIfNeeded(); err != nil {
				return nil, d.logger.ErrorRet(err, "createTableIfNeeded failed")
			}

			// open db connection
			dbConnection := database.NewConnection()
			if err = dbConnection.Open(); err != nil {
				return nil, d.logger.ErrorRet(err, "dbConnection.Open failed")
			}
			defer dbConnection.Close()

			// list volumes
			dataModel := NewScbeDataModel(dbConnection.GetDb())
			if volumes, err = dataModel.ListVolumes(); err != nil {
				return nil, d.logger.ErrorRet(err, "dataModel.ListVolumes failed")
			}
		}

		volumes = append(volumes, *d.dbVolume)
	}

	return volumes, nil
}

func (d *scbeDataModelWrapper) UpdateVolumeAttachTo(volumeName string, scbeVolume ScbeVolume, host2attach string) error {
	defer d.logger.Trace(logs.DEBUG)()
	var err error

	if isDatabaseVolume(volumeName) {

		// sanity
		if d.dbVolume == nil {
			return d.logger.ErrorRet(&volAlreadyExistsError{volumeName}, "failed")
		}

		// work with memory object
		d.dbVolume.AttachTo = host2attach

	} else {

		// create table if needed
		if err = d.createTableIfNeeded(); err != nil {
			return d.logger.ErrorRet(err, "createTableIfNeeded failed")
		}

		// open db connection
		dbConnection := database.NewConnection()
		if err = dbConnection.Open(); err != nil {
			return d.logger.ErrorRet(err, "dbConnection.Open failed")
		}
		defer dbConnection.Close()

		// delete volume
		dataModel := NewScbeDataModel(dbConnection.GetDb())
		if err = dataModel.UpdateVolumeAttachTo(volumeName, scbeVolume, host2attach); err != nil {
			return d.logger.ErrorRet(err, "dataModel.UpdateVolumeAttachTo failed")
		}
	}

	return nil
}