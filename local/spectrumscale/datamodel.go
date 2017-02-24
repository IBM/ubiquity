package spectrumscale

import (
	"log"

	"fmt"

	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
	"github.ibm.com/almaden-containers/ubiquity/model"
	"github.ibm.com/almaden-containers/ubiquity/resources"
)

//go:generate counterfeiter -o ../../fakes/fake_SpectrumDataModel.go . SpectrumDataModel
type SpectrumDataModel interface {
	CreateVolumeTable() error
	SetClusterId(string)
	GetClusterId() string
	DeleteVolume(name string) error
	InsertFilesetVolume(fileset, volumeName string, filesystem string, isPreexisting bool, opts map[string]interface{}) error
	InsertLightweightVolume(fileset, directory, volumeName string, filesystem string, isPreexisting bool, opts map[string]interface{}) error
	InsertFilesetQuotaVolume(fileset, quota, volumeName string, filesystem string, isPreexisting bool, opts map[string]interface{}) error
	GetVolume(name string) (SpectrumScaleVolume, bool, error)
	ListVolumes() ([]SpectrumScaleVolume, error)
}

type spectrumDataModel struct {
	log       *log.Logger
	database  *gorm.DB
	clusterId string
	backend   resources.Backend
}

type VolumeType int

const (
	FILESET VolumeType = iota
	LIGHTWEIGHT
	FILESET_WITH_QUOTA
)

const (
	USER_SPECIFIED_UID string = "uid"
	USER_SPECIFIED_GID string = "gid"
)

type SpectrumScaleVolume struct {
	ID            uint
	Volume        model.Volume
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

func NewSpectrumDataModel(log *log.Logger, db *gorm.DB, backend resources.Backend) SpectrumDataModel {
	return &spectrumDataModel{log: log, database: db, backend: backend}
}

func (d *spectrumDataModel) SetClusterId(id string) {
	d.clusterId = id
}
func (d *spectrumDataModel) GetClusterId() string {
	return d.clusterId
}
func (d *spectrumDataModel) CreateVolumeTable() error {
	d.log.Println("SpectrumDataModel: Create Volumes Table start")
	defer d.log.Println("SpectrumDataModel: Create Volumes Table end")

	if err := d.database.AutoMigrate(&SpectrumScaleVolume{}).Error; err != nil {
		return err
	}
	return nil
}

func (d *spectrumDataModel) DeleteVolume(name string) error {
	d.log.Println("SpectrumDataModel: DeleteVolume start")
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
	return nil
}

func (d *spectrumDataModel) InsertFilesetVolume(fileset, volumeName string, filesystem string, isPreexisting bool, opts map[string]interface{}) error {
	d.log.Println("SpectrumDataModel: InsertFilesetVolume start")
	defer d.log.Println("SpectrumDataModel: InsertFilesetVolume end")

	volume := SpectrumScaleVolume{Volume: model.Volume{Name: volumeName, Backend: fmt.Sprintf("%s", d.backend)}, Type: FILESET, ClusterId: d.clusterId, FileSystem: filesystem,
		Fileset: fileset, IsPreexisting: isPreexisting}

	addPermissionsForVolume(&volume, opts)

	return d.insertVolume(volume)
}

func (d *spectrumDataModel) InsertLightweightVolume(fileset, directory, volumeName string, filesystem string, isPreexisting bool, opts map[string]interface{}) error {
	d.log.Println("SpectrumDataModel: InsertLightweightVolume start")
	defer d.log.Println("SpectrumDataModel: InsertLightweightVolume end")

	volume := SpectrumScaleVolume{Volume: model.Volume{Name: volumeName, Backend: fmt.Sprintf("%s", d.backend)}, Type: LIGHTWEIGHT, ClusterId: d.clusterId, FileSystem: filesystem,
		Fileset: fileset, Directory: directory, IsPreexisting: isPreexisting}

	addPermissionsForVolume(&volume, opts)

	return d.insertVolume(volume)
}

func (d *spectrumDataModel) InsertFilesetQuotaVolume(fileset, quota, volumeName string, filesystem string, isPreexisting bool, opts map[string]interface{}) error {
	d.log.Println("SpectrumDataModel: InsertFilesetQuotaVolume start")
	defer d.log.Println("SpectrumDataModel: InsertFilesetQuotaVolume end")

	volume := SpectrumScaleVolume{Volume: model.Volume{Name: volumeName, Backend: fmt.Sprintf("%s", d.backend)}, Type: FILESET_WITH_QUOTA, ClusterId: d.clusterId, FileSystem: filesystem,
		Fileset: fileset, Quota: quota, IsPreexisting: isPreexisting}

	addPermissionsForVolume(&volume, opts)

	return d.insertVolume(volume)
}

func (d *spectrumDataModel) insertVolume(volume SpectrumScaleVolume) error {
	d.log.Println("SpectrumDataModel: insertVolume start")
	defer d.log.Println("SpectrumDataModel: insertVolume end")

	if err := d.database.Create(&volume).Error; err != nil {
		return err
	}
	return nil
}

func (d *spectrumDataModel) GetVolume(name string) (SpectrumScaleVolume, bool, error) {
	d.log.Println("SpectrumDataModel: GetVolume start")
	defer d.log.Println("SpectrumDataModel: GetVolume end")

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

func (d *spectrumDataModel) ListVolumes() ([]SpectrumScaleVolume, error) {
	d.log.Println("SpectrumDataModel: ListVolumes start")
	defer d.log.Println("SpectrumDataModel: ListVolumes end")

	var volumesInDb []SpectrumScaleVolume
	if err := d.database.Preload("Volume").Find(&volumesInDb).Error; err != nil {
		return nil, err
	}
	// hack: to be replaced by proper DB filtering (joins)
	var volumes []SpectrumScaleVolume
	for _, volume := range volumesInDb {
		if resources.Backend(volume.Volume.Backend) == d.backend {
			volumes = append(volumes, volume)
		}
	}

	return volumes, nil
}

func addPermissionsForVolume(volume *SpectrumScaleVolume, opts map[string]interface{}) {

	if len(opts) > 0 {
		uid, uidSpecified := opts[USER_SPECIFIED_UID]
		gid, gidSpecified := opts[USER_SPECIFIED_GID]

		if uidSpecified && gidSpecified {
			volume.UID = uid.(string)
			volume.GID = gid.(string)

		}
	}
}
