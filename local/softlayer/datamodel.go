package softlayer

import (
	"log"

	"fmt"

	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
	"github.ibm.com/almaden-containers/ubiquity/model"
	"github.ibm.com/almaden-containers/ubiquity/resources"
)

//go:generate counterfeiter -o ../../fakes/fake_SoftlayerDataModel.go . SoftlayerDataModel
type SoftlayerDataModel interface {
	CreateVolumeTable() error
	DeleteVolume(name string) error
	InsertFileshare(fileshareID int, volumeName string, mountPath string, opts map[string]interface{}) error
	GetVolume(name string) (SoftlayerVolume, bool, error)
	GetVolumeForMountPoint(mountpoint string) (string, error)
	ListVolumes() ([]SoftlayerVolume, error)
}

const (
	SERVER_USER_NAME  = "serverUserName"
	SERVER_BACKEND_IP = "serverBackendIP"
)

type softlayerDataModel struct {
	log      *log.Logger
	database *gorm.DB
	backend  resources.Backend
}

type SoftlayerVolume struct {
	ID          uint
	Volume      model.Volume
	VolumeID    uint
	FileshareID int
	Mountpoint  string
}

func NewSoftlayerDataModel(log *log.Logger, db *gorm.DB, backend resources.Backend) SoftlayerDataModel {
	return &softlayerDataModel{log: log, database: db, backend: backend}
}

func (d *softlayerDataModel) CreateVolumeTable() error {
	d.log.Println("SoftlayerDataModel: Create SLVolumes Table start")
	defer d.log.Println("SoftlayerDataModel: Create SLVolumes Table end")

	if err := d.database.AutoMigrate(&SoftlayerVolume{}).Error; err != nil {
		return err
	}
	return nil
}

func (d *softlayerDataModel) DeleteVolume(name string) error {
	d.log.Println("SoftlayerDataModel: DeleteVolume start")
	defer d.log.Println("SoftlayerDataModel: DeleteVolume end")

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

func (d *softlayerDataModel) InsertFileshare(fileshareID int, volumeName string, mountPath string, opts map[string]interface{}) error {
	d.log.Println("SoftlayerDataModel: InsertFilesetVolume start")
	defer d.log.Println("SoftlayerDataModel: InsertFilesetVolume end")

	volume := SoftlayerVolume{Volume: model.Volume{Name: volumeName, Backend: fmt.Sprintf("%s", d.backend)}, FileshareID: fileshareID, Mountpoint: mountPath}

	return d.insertVolume(volume)
}

func (d *softlayerDataModel) insertVolume(volume SoftlayerVolume) error {
	d.log.Println("SoftlayerDataModel: insertVolume start")
	defer d.log.Println("SoftlayerDataModel: insertVolume end")

	if err := d.database.Create(&volume).Error; err != nil {
		return err
	}

	return nil
}

func (d *softlayerDataModel) GetVolume(name string) (SoftlayerVolume, bool, error) {
	d.log.Println("SoftlayerDataModel: GetVolume start")
	defer d.log.Println("SoftlayerDataModel: GetVolume end")

	volume, err := model.GetVolume(d.database, name, d.backend)
	if err != nil {
		if err.Error() == "record not found" {
			return SoftlayerVolume{}, false, nil
		}
		return SoftlayerVolume{}, false, err
	}

	var softlayerVolume SoftlayerVolume
	if err := d.database.Where("volume_id = ?", volume.ID).Preload("Volume").First(&softlayerVolume).Error; err != nil {
		if err.Error() == "record not found" {
			return SoftlayerVolume{}, false, nil
		}
		return SoftlayerVolume{}, false, err
	}

	return softlayerVolume, true, nil
}

func (d *softlayerDataModel) GetVolumeForMountPoint(mountpoint string) (string, error) {
	d.log.Println("SoftlayerDataModel: GetVolumeForMountPoint start")
	defer d.log.Println("SoftlayerDataModel: GetVolumeForMountPoint end")

	//read_volume_stmt := `
	//SELECT Name FROM SLVolumes WHERE MountPoint = ?
	//`
	//
	//stmt, err := d.databaseClient.GetHandle().Prepare(read_volume_stmt)
	//
	//if err != nil {
	//	return "", fmt.Errorf("Failed to create GetVolumeForMountPoint Stmt for %s : %s", mountpoint, err.Error())
	//}
	//
	//defer stmt.Close()
	//
	//var volName string
	//
	//err = stmt.QueryRow(mountpoint).Scan(&volName)
	//
	//if err != nil {
	//	return "", fmt.Errorf("Failed to Get Volume for %s : %s", mountpoint, err.Error())
	//}

	return "", nil
}

func (d *softlayerDataModel) ListVolumes() ([]SoftlayerVolume, error) {
	d.log.Println("SoftlayerDataModel: ListSLVolumesstart")
	defer d.log.Println("SoftlayerDataModel: ListSLVolumesend")

	var volumes []SoftlayerVolume
	if err := d.database.Preload("Volume").Find(&volumes).Error; err != nil {
		return nil, err
	}
	return volumes, nil
}
