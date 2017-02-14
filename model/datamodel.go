package model

import (
	"fmt"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.ibm.com/almaden-containers/ubiquity/resources"
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
