package spectrumscale

import (
	"fmt"
	"log"

	"strings"

	_ "github.com/mattn/go-sqlite3"
	"github.ibm.com/almaden-containers/ubiquity/utils"
	"database/sql"
)

//go:generate counterfeiter -o ../../fakes/fake_SpectrumDataModel.go . SpectrumDataModel
type SpectrumDataModel interface {
	CreateVolumeTable() error
	SetClusterId(string)
	GetClusterId() string
	//VolumeExists(name string) (bool, error)
	DeleteVolume(name string) error
	InsertFilesetVolume(fileset, volumeName string, filesystem string, opts map[string]interface{}) error
	InsertLightweightVolume(fileset, directory, volumeName string, filesystem string, opts map[string]interface{}) error
	InsertFilesetQuotaVolume(fileset, quota, volumeName string, filesystem string, opts map[string]interface{}) error
	GetVolume(name string) (Volume, bool, error)
	ListVolumes() ([]Volume, error)
}

type spectrumDataModel struct {
	log            *log.Logger
	databaseClient utils.DatabaseClient
	clusterId      string
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

type Volume struct {
	Id             int
	Name           string
	Type           VolumeType
	ClusterId      string
	FileSystem     string
	Fileset        string
	Directory      string
	AdditionalData map[string]string
}

func NewSpectrumDataModel(log *log.Logger, dbClient utils.DatabaseClient) SpectrumDataModel {
	return &spectrumDataModel{log: log, databaseClient: dbClient}
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

	volumes_table_create_stmt := `
	 CREATE TABLE IF NOT EXISTS Volumes (
	     Id       INTEGER PRIMARY KEY AUTOINCREMENT,
	     Name     TEXT NOT NULL,
	     Type     INTEGER NOT NULL,
	     ClusterId      TEXT NOT NULL,
             Filesystem     TEXT NOT NULL,
             Fileset        TEXT NOT NULL,
             Directory      TEXT,
             AdditionalData TEXT
         );
	`

	_, err := d.databaseClient.GetHandle().Exec(volumes_table_create_stmt)

	if err != nil {
		return fmt.Errorf("Failed To Create Volumes Table: %s", err.Error())
	}

	return nil
}

func (d *spectrumDataModel) DeleteVolume(name string) error {
	d.log.Println("SpectrumDataModel: DeleteVolume start")
	defer d.log.Println("SpectrumDataModel: DeleteVolume end")

	// Delete volume from table
	delete_volume_stmt := `
	DELETE FROM Volumes WHERE Name = ?
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

func (d *spectrumDataModel) InsertFilesetVolume(fileset, volumeName string, filesystem string, opts map[string]interface{}) error {
	d.log.Println("SpectrumDataModel: InsertFilesetVolume start")
	defer d.log.Println("SpectrumDataModel: InsertFilesetVolume end")

	volume := Volume{Name: volumeName, Type: FILESET, ClusterId: d.clusterId, FileSystem: filesystem,
		Fileset: fileset}

	addPermissionsForVolume(&volume, opts)

	return d.insertVolume(volume)
}

func (d *spectrumDataModel) InsertLightweightVolume(fileset, directory, volumeName string, filesystem string, opts map[string]interface{}) error {
	d.log.Println("SpectrumDataModel: InsertLightweightVolume start")
	defer d.log.Println("SpectrumDataModel: InsertLightweightVolume end")

	volume := Volume{Name: volumeName, Type: LIGHTWEIGHT, ClusterId: d.clusterId, FileSystem: filesystem,
		Fileset: fileset, Directory: directory}

	addPermissionsForVolume(&volume, opts)

	return d.insertVolume(volume)
}

func (d *spectrumDataModel) InsertFilesetQuotaVolume(fileset, quota, volumeName string, filesystem string, opts map[string]interface{}) error {
	d.log.Println("SpectrumDataModel: InsertFilesetQuotaVolume start")
	defer d.log.Println("SpectrumDataModel: InsertFilesetQuotaVolume end")

	volume := Volume{Name: volumeName, Type: FILESET_WITH_QUOTA, ClusterId: d.clusterId, FileSystem: filesystem,
		Fileset: fileset}

	volume.AdditionalData = make(map[string]string)
	volume.AdditionalData["quota"] = quota

	addPermissionsForVolume(&volume, opts)

	return d.insertVolume(volume)
}

func (d *spectrumDataModel) insertVolume(volume Volume) error {
	d.log.Println("SpectrumDataModel: insertVolume start")
	defer d.log.Println("SpectrumDataModel: insertVolume end")

	insert_volume_stmt := `
	INSERT INTO Volumes(Name, Type, ClusterId, Filesystem, Fileset, Directory, AdditionalData)
	values(?,?,?,?,?,?,?,?);
	`

	stmt, err := d.databaseClient.GetHandle().Prepare(insert_volume_stmt)

	if err != nil {
		return fmt.Errorf("Failed to create InsertVolume Stmt for %s: %s", volume.Name, err.Error())
	}

	defer stmt.Close()

	additionalData := getAdditionalData(&volume)

	_, err = stmt.Exec(volume.Name, volume.Type, volume.ClusterId, volume.FileSystem, volume.Fileset,
		volume.Directory, additionalData)

	if err != nil {
		return fmt.Errorf("Failed to Insert Volume %s : %s", volume.Name, err.Error())
	}

	return nil
}

func (d *spectrumDataModel) GetVolume(name string) (Volume, bool, error) {
	d.log.Println("SpectrumDataModel: GetVolume start")
	defer d.log.Println("SpectrumDataModel: GetVolume end")

	read_volume_stmt := `
        SELECT * FROM Volumes WHERE Name = ?
        `

	stmt, err := d.databaseClient.GetHandle().Prepare(read_volume_stmt)

	if err != nil {
		return Volume{}, false, fmt.Errorf("Failed to create GetVolume Stmt for %s : %s", name, err.Error())
	}

	defer stmt.Close()

	var volName, clusterId, filesystem, fileset, directory, addData string
	var volType, volId int

	err = stmt.QueryRow(name).Scan(&volId, &volName, &volType, &clusterId, &filesystem, &fileset, &directory, &addData)

	if err != nil {
		if err == sql.ErrNoRows {
			return Volume{}, false, nil
		}
		return Volume{}, false, fmt.Errorf("Failed to Get Volume for %s : %s", name, err.Error())
	}

	scannedVolume := Volume{Id: volId, Name: volName, Type: VolumeType(volType), ClusterId: clusterId, FileSystem: filesystem,
		Fileset: fileset, Directory: directory}

	setAdditionalData(addData, &scannedVolume)

	return scannedVolume, true, nil
}

func (d *spectrumDataModel) ListVolumes() ([]Volume, error) {
	d.log.Println("SpectrumDataModel: ListVolumes start")
	defer d.log.Println("SpectrumDataModel: ListVolumes end")

	list_volumes_stmt := `
        SELECT *
        FROM Volumes
        `

	rows, err := d.databaseClient.GetHandle().Query(list_volumes_stmt)
	defer rows.Close()

	if err != nil {
		return nil, fmt.Errorf("Failed to List Volumes : %s", err.Error())
	}

	var volumes []Volume
	var volName, clusterId, filesystem, fileset, directory, addData string
	var volType, volId int

	for rows.Next() {

		err = rows.Scan(&volId, &volName, &volType, &clusterId, &filesystem, &fileset, &directory, &addData)

		if err != nil {
			return nil, fmt.Errorf("Failed to scan rows while listing volumes: %s", err.Error())
		}

		scannedVolume := Volume{Id: volId, Name: volName, Type: VolumeType(volType), ClusterId: clusterId,
			FileSystem: filesystem, Fileset: fileset, Directory: directory}

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
			addData += key + "=" + value + ","
		}
		addData = strings.TrimSuffix(addData, ",")
	}
	return addData
}

func setAdditionalData(addData string, volume *Volume) {

	if len(addData) > 0 {
		volume.AdditionalData = make(map[string]string)

		lines := strings.Split(addData, ",")

		for _, line := range lines {
			tokens := strings.Split(line, "=")
			volume.AdditionalData[tokens[0]] = tokens[1]
		}
	}
}

func addPermissionsForVolume(volume *Volume, opts map[string]interface{}) {

	if len(opts) > 0 {
		uid, uidSpecified := opts[USER_SPECIFIED_UID]
		gid, gidSpecified := opts[USER_SPECIFIED_GID]

		if uidSpecified && gidSpecified {

			if volume.AdditionalData == nil {
				volume.AdditionalData = make(map[string]string)
			}

			volume.AdditionalData["uid"] = uid.(string)
			volume.AdditionalData["gid"] = gid.(string)
		}
	}
}
