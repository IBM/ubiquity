package scbe

import (
	"log"

	"fmt"

	"github.com/IBM/ubiquity/model"
	"github.com/IBM/ubiquity/resources"
	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
)

//go:generate counterfeiter -o ../../fakes/fake_ScbeDataModel.go . ScbeDataModel
type ScbeDataModel interface {
	CreateVolumeTable() error
	DeleteVolume(name string) error
	InsertVolume(volumeName string, opts map[string]interface{}) error
	GetVolume(name string) (ScbeVolume, bool, error)
	ListVolumes() ([]ScbeVolume, error)
}

type scbeDataModel struct {
	log       *log.Logger
	database  *gorm.DB
	clusterId string
	backend   resources.Backend
}

const (
	USER_SPECIFIED_UID string = "uid"
	USER_SPECIFIED_GID string = "gid"
)

type ScbeVolume struct {
	ID     uint
	Volume model.Volume
	WWN    string
	Quota  string
}

func NewScbeDataModel(log *log.Logger, db *gorm.DB, backend resources.Backend) ScbeDataModel {
	return &scbeDataModel{log: log, database: db, backend: backend}
}

func (d *scbeDataModel) CreateVolumeTable() error {
	d.log.Println("ScbeDataModel: Create Volumes Table start")
	defer d.log.Println("ScbeDataModel: Create Volumes Table end")

	if err := d.database.AutoMigrate(&ScbeVolume{}).Error; err != nil {
		return err
	}
	return nil
}

func (d *scbeDataModel) DeleteVolume(name string) error {
	d.log.Println("ScbeDataModel: DeleteVolume start")
	defer d.log.Println("SpectrumDataModel: DeleteVolume end")

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

func (d *scbeDataModel) InsertVolume(volumeName string, opts map[string]interface{}) error {
	d.log.Println("ScbeDataModel: InsertVolume start")
	defer d.log.Println("ScbeDataModel: InsertVolume end")

	volume := ScbeVolume{Volume: model.Volume{Name: volumeName, Backend: fmt.Sprintf("%s", d.backend)}}

	return d.insertVolume(volume)
}

func (d *scbeDataModel) insertVolume(volume ScbeVolume) error {
	d.log.Println("ScbeDataModel: insertVolume start")
	defer d.log.Println("ScbeDataModel: insertVolume end")

	if err := d.database.Create(&volume).Error; err != nil {
		return err
	}
	return nil
}

func (d *scbeDataModel) GetVolume(name string) (ScbeVolume, bool, error) {
	d.log.Println("ScbeDataModel: GetVolume start")
	defer d.log.Println("ScbeDataModel: GetVolume end")

	volume, err := model.GetVolume(d.database, name, d.backend)
	if err != nil {
		if err.Error() == "record not found" {
			return ScbeVolume{}, false, nil
		}
		return ScbeVolume{}, false, err
	}

	var scbeVolume ScbeVolume
	if err := d.database.Where("volume_id = ?", volume.ID).Preload("Volume").First(&scbeVolume).Error; err != nil {
		if err.Error() == "record not found" {
			return ScbeVolume{}, false, nil
		}
		return ScbeVolume{}, false, err
	}
	return scbeVolume, true, nil
}

func (d *scbeDataModel) ListVolumes() ([]ScbeVolume, error) {
	d.log.Println("ScbeDataModel: ListVolumes start")
	defer d.log.Println("ScbeDataModel: ListVolumes end")

	var volumesInDb []ScbeVolume
	if err := d.database.Preload("Volume").Find(&volumesInDb).Error; err != nil {
		return nil, err
	}
	// hack: to be replaced by proper DB filtering (joins)
	var volumes []ScbeVolume
	for _, volume := range volumesInDb {
		if resources.Backend(volume.Volume.Backend) == d.backend {
			volumes = append(volumes, volume)
		}
	}

	return volumes, nil
}
