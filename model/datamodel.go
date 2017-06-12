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
