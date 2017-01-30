package softlayer

import (
	"fmt"
	"log"

	"strings"

	_ "github.com/mattn/go-sqlite3"
	"github.ibm.com/almaden-containers/ubiquity/utils"
)

//go:generate counterfeiter -o ../../fakes/fake_SoftlayerDataModel.go . SoftlayerDataModel
type SoftlayerDataModel interface {
	CreateVolumeTable() error
	VolumeExists(name string) (bool, error)
	DeleteVolume(name string) error
	InsertFileshare(fileshareID int, volumeName string, mountPath string, opts map[string]interface{}) error
	UpdateVolumeMountpoint(name, mountpoint string) error
	GetVolume(name string) (Volume, error)
	GetVolumeForMountPoint(mountpoint string) (string, error)
	ListVolumes() ([]Volume, error)
}

const (
	SERVER_USER_NAME  = "serverUserName"
	SERVER_BACKEND_IP = "serverBackendIP"
)

type softlayerDataModel struct {
	log            *log.Logger
	databaseClient utils.DatabaseClient
}

type Volume struct {
	Id             int
	Name           string
	FileshareID    int
	Mountpoint     string
	AdditionalData map[string]interface{}
}

func NewSoftlayerDataModel(log *log.Logger, dbClient utils.DatabaseClient) SoftlayerDataModel {
	return &softlayerDataModel{log: log, databaseClient: dbClient}
}

func (d *softlayerDataModel) CreateVolumeTable() error {
	d.log.Println("SoftlayerDataModel: Create SLVolumes Table start")
	defer d.log.Println("SoftlayerDataModel: Create SLVolumes Table end")

	volumes_table_create_stmt := `
	 CREATE TABLE IF NOT EXISTS SLVolumes(
	     Id       INTEGER PRIMARY KEY AUTOINCREMENT,
	     Name     TEXT NOT NULL,
	     FileshareID    TEXT NOT NULL,
             Directory      TEXT,
             MountPoint     TEXT,
             AdditionalData TEXT
         );
	`

	_, err := d.databaseClient.GetHandle().Exec(volumes_table_create_stmt)

	if err != nil {
		return fmt.Errorf("Failed To Create SLVolumesTable: %s", err.Error())
	}

	return nil
}

func (d *softlayerDataModel) VolumeExists(name string) (bool, error) {
	d.log.Println("SoftlayerDataModel: VolumeExists start")
	defer d.log.Println("SoftlayerDataModel: VolumeExists end")

	volume_exists_stmt := `
	SELECT EXISTS ( SELECT Name FROM SLVolumes WHERE Name = ? )
	`

	stmt, err := d.databaseClient.GetHandle().Prepare(volume_exists_stmt)

	if err != nil {
		return false, fmt.Errorf("Failed to create VolumeExists Stmt for %s: %s", name, err.Error())
	}

	defer stmt.Close()

	var exists int
	err = stmt.QueryRow(name).Scan(&exists)

	if err != nil {
		return false, fmt.Errorf("Failed to query VolumeExists stmt for %s: %s", name, err.Error())
	}

	if exists == 1 {
		return true, nil
	}

	return false, nil
}

func (d *softlayerDataModel) DeleteVolume(name string) error {
	d.log.Println("SoftlayerDataModel: DeleteVolume start")
	defer d.log.Println("SoftlayerDataModel: DeleteVolume end")

	// Delete volume from table
	delete_volume_stmt := `
	DELETE FROM SLVolumes WHERE Name = ?
	`

	stmt, err := d.databaseClient.GetHandle().Prepare(delete_volume_stmt)

	if err != nil {
		return fmt.Errorf("Failed to create DeleteVolume Stmt for %s: %s", name, err.Error())
	}

	defer stmt.Close()

	_, err = stmt.Exec(name)

	if err != nil {
		return fmt.Errorf("Failed to Delete Volume %s : %s", name, err.Error())
	}

	return nil
}

func (d *softlayerDataModel) InsertFileshare(fileshareID int, volumeName string, mountPath string, opts map[string]interface{}) error {
	d.log.Println("SoftlayerDataModel: InsertFilesetVolume start")
	defer d.log.Println("SoftlayerDataModel: InsertFilesetVolume end")

	volume := Volume{Name: volumeName, FileshareID: fileshareID, Mountpoint: mountPath, AdditionalData: opts}

	return d.insertVolume(volume)
}

func (d *softlayerDataModel) insertVolume(volume Volume) error {
	d.log.Println("SoftlayerDataModel: insertVolume start")
	defer d.log.Println("SoftlayerDataModel: insertVolume end")

	insert_volume_stmt := `
	INSERT INTO SLVolumes(Name, Type, ClusterId, Filesystem, Fileset, Directory, MountPoint, AdditionalData)
	values(?,?,?,?,?,?,?,?);
	`

	stmt, err := d.databaseClient.GetHandle().Prepare(insert_volume_stmt)

	if err != nil {
		return fmt.Errorf("Failed to create InsertVolume Stmt for %s: %s", volume.Name, err.Error())
	}

	defer stmt.Close()

	additionalData := getAdditionalData(&volume)

	_, err = stmt.Exec(volume.Name, volume.FileshareID, volume.Mountpoint, additionalData)

	if err != nil {
		return fmt.Errorf("Failed to Insert Volume %s : %s", volume.Name, err.Error())
	}

	return nil
}

func (d *softlayerDataModel) UpdateVolumeMountpoint(name, mountpoint string) error {
	d.log.Println("SoftlayerDataModel: UpdateVolumeMountpoint start")
	defer d.log.Println("SoftlayerDataModel: UpdateVolumeMountpoint end")

	update_volume_stmt := `
	UPDATE SLVolumes
	SET MountPoint = ?
	WHERE Name = ?
	`

	stmt, err := d.databaseClient.GetHandle().Prepare(update_volume_stmt)

	if err != nil {
		return fmt.Errorf("Failed to create UpdateVolume Stmt for %s: %s", name, err.Error())
	}

	defer stmt.Close()

	_, err = stmt.Exec(mountpoint, name)

	if err != nil {
		return fmt.Errorf("Failed to Update Volume %s : %s", name, err.Error())
	}

	return nil
}

func (d *softlayerDataModel) GetVolume(name string) (Volume, error) {
	d.log.Println("SoftlayerDataModel: GetVolume start")
	defer d.log.Println("SoftlayerDataModel: GetVolume end")

	read_volume_stmt := `
        SELECT * FROM SLVolumes WHERE Name = ?
        `

	stmt, err := d.databaseClient.GetHandle().Prepare(read_volume_stmt)

	if err != nil {
		return Volume{}, fmt.Errorf("Failed to create GetVolume Stmt for %s : %s", name, err.Error())
	}

	defer stmt.Close()

	var volName, mountpoint, addData string
	var volId, fileshareID int

	err = stmt.QueryRow(name).Scan(&volId, &volName, &fileshareID, &mountpoint, &addData)

	if err != nil {
		return Volume{}, fmt.Errorf("Failed to Get Volume for %s : %s", name, err.Error())
	}

	scannedVolume := Volume{Id: volId, Name: volName, FileshareID: fileshareID, Mountpoint: mountpoint}

	setAdditionalData(addData, &scannedVolume)

	return scannedVolume, nil
}

func (d *softlayerDataModel) GetVolumeForMountPoint(mountpoint string) (string, error) {
	d.log.Println("SoftlayerDataModel: GetVolumeForMountPoint start")
	defer d.log.Println("SoftlayerDataModel: GetVolumeForMountPoint end")

	read_volume_stmt := `
        SELECT Name FROM SLVolumes WHERE MountPoint = ?
        `

	stmt, err := d.databaseClient.GetHandle().Prepare(read_volume_stmt)

	if err != nil {
		return "", fmt.Errorf("Failed to create GetVolumeForMountPoint Stmt for %s : %s", mountpoint, err.Error())
	}

	defer stmt.Close()

	var volName string

	err = stmt.QueryRow(mountpoint).Scan(&volName)

	if err != nil {
		return "", fmt.Errorf("Failed to Get Volume for %s : %s", mountpoint, err.Error())
	}

	return volName, nil
}

func (d *softlayerDataModel) ListVolumes() ([]Volume, error) {
	d.log.Println("SoftlayerDataModel: ListSLVolumesstart")
	defer d.log.Println("SoftlayerDataModel: ListSLVolumesend")

	list_volumes_stmt := `
        SELECT *
        FROM SLVolumes
        `

	rows, err := d.databaseClient.GetHandle().Query(list_volumes_stmt)
	defer rows.Close()

	if err != nil {
		return nil, fmt.Errorf("Failed to List SLVolumes: %s", err.Error())
	}

	var volumes []Volume
	var volName, mountpoint, addData string
	var volId, fileshareId int

	for rows.Next() {

		err = rows.Scan(&volId, &volName, &fileshareId, &mountpoint, &addData)

		if err != nil {
			return nil, fmt.Errorf("Failed to scan rows while listing volumes: %s", err.Error())
		}

		scannedVolume := Volume{Id: volId, Name: volName, FileshareID: fileshareId, Mountpoint: mountpoint}

		setAdditionalData(addData, &scannedVolume)

		volumes = append(volumes, scannedVolume)
	}

	err = rows.Err()

	if err != nil {
		return nil, fmt.Errorf("Failure while iterating rows : %s", err.Error())
	}

	return volumes, nil
}

func getAdditionalData(volume *Volume) string {

	var addData string

	if len(volume.AdditionalData) > 0 {

		for key, value := range volume.AdditionalData {
			addData += key + "=" + value.(string) + ","
		}
		addData = strings.TrimSuffix(addData, ",")
	}
	return addData
}

func setAdditionalData(addData string, volume *Volume) {

	if len(addData) > 0 {
		volume.AdditionalData = make(map[string]interface{})

		lines := strings.Split(addData, ",")

		for _, line := range lines {
			tokens := strings.Split(line, "=")
			volume.AdditionalData[tokens[0]] = tokens[1]
		}
	}
}
