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

package model

import (
	"fmt"

	"github.com/IBM/ubiquity/resources"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

func GetVolume(db *gorm.DB, name string, backend string) (resources.Volume, error) {
	var volume resources.Volume
	err := db.Where("name = ? AND backend = ?", name, fmt.Sprintf("%s", backend)).First(&volume).Error
	return volume, err
}
func GetBackendForVolume(db *gorm.DB, name string) (string, error) {
	var volume resources.Volume
	err := db.Where("name = ? ", name).First(&volume).Error
	if err != nil {
		return "", err
	}
	return volume.Backend, err
}
func VolumeExists(db *gorm.DB, name string) (bool, error) {
	var volume resources.Volume
	err := db.Where("name = ? ", name).First(&volume).Error
	if err != nil {
		return false, err
	}
	return true, err
}
func DeleteVolume(db *gorm.DB, volume *resources.Volume) *gorm.DB {
	return db.Delete(volume)
}

func UpdateVolumeMountpoint(db *gorm.DB, volume *resources.Volume, mountpoint string) error {
	err := db.Model(volume).Update("mountpoint", mountpoint).Error
	return err
}
