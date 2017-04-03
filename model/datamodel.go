package model

import (
	"fmt"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/ibm/ubiquity/resources"
)

type Volume struct {
	gorm.Model
	Name    string
	Backend string
}

func GetVolume(db *gorm.DB, name string, backend resources.Backend) (Volume, error) {
	var volume Volume
	err := db.Where("name = ? AND backend = ?", name, fmt.Sprintf("%s", backend)).First(&volume).Error
	return volume, err
}
func GetBackendForVolume(db *gorm.DB, name string) (resources.Backend, error) {
	var volume Volume
	err := db.Where("name = ? ", name).First(&volume).Error
	if err != nil {
		return resources.Backend(""), err
	}
	return resources.Backend(volume.Backend), err
}
func VolumeExists(db *gorm.DB, name string) (bool, error) {
	var volume Volume
	err := db.Where("name = ? ", name).First(&volume).Error
	if err != nil {
		return false, err
	}
	return true, err
}
func DeleteVolume(db *gorm.DB, volume *Volume) *gorm.DB {
	return db.Delete(volume)
}
