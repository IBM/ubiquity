package utils

import (
	"database/sql"
	"fmt"
	"log"
	"path"

	"strings"

	_ "github.com/mattn/go-sqlite3"
)

//go:generate counterfeiter -o ../fakes/fake_database_client.go . DatabaseClient
type DatabaseClient interface {
	Init() error
	Close() error
	CreateVolumeTable() error
	SetClusterId(string)
	GetClusterId() string
	VolumeExists(name string) (bool, error)
	DeleteVolume(name string) error
	InsertFilesetVolume(fileset, volumeName string, filesystem string, opts map[string]interface{}) error
	InsertLightweightVolume(fileset, directory, volumeName string, filesystem string, opts map[string]interface{}) error
	InsertFilesetQuotaVolume(fileset, quota, volumeName string, filesystem string, opts map[string]interface{}) error
	UpdateVolumeMountpoint(name, mountpoint string) error
	GetVolume(name string) (Volume, error)
	GetVolumeForMountPoint(mountpoint string) (string, error)
	ListVolumes() ([]Volume, error)
}

type dbClient struct {
	Mountpoint string
	log        *log.Logger
	Db         *sql.DB
	ClusterId  string
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
	Mountpoint     string
	AdditionalData map[string]string
}

func NewDatabaseClient(log *log.Logger, mountpoint string) DatabaseClient {
	return &dbClient{log: log, Mountpoint: mountpoint}
}

func (d *dbClient) Init() error {

	d.log.Println("DatabaseClient: DB Init start")
	defer d.log.Println("DatabaseClient: DB Init end")

	dbFile := "spectrum-scale.db"
	dbPath := path.Join(d.Mountpoint, dbFile)

	Db, err := sql.Open("sqlite3", dbPath)

	if err != nil {
		return fmt.Errorf("Failed to Initialize DB connection to %s: %s\n", dbPath, err.Error())
	}
	d.Db = Db
	d.log.Printf("Established Database connection to %s via go-sqlite3 driver", dbPath)
	return nil
}

func (d *dbClient) Close() error {

	d.log.Println("DatabaseClient: DB Close start")
	defer d.log.Println("DatabaseClient: DB Close end")

	if d.Db != nil {
		err := d.Db.Close()
		if err != nil {
			return fmt.Errorf("Failed to close DB connection: %s\n", err.Error())
		}
	}
	return nil
}
func (d *dbClient) SetClusterId(id string) {
	d.ClusterId = id
}
func (d *dbClient) GetClusterId() string {
	return d.ClusterId
}
func (d *dbClient) CreateVolumeTable() error {
	d.log.Println("DatabaseClient: Create Volumes Table start")
	defer d.log.Println("DatabaseClient: Create Volumes Table end")

	volumes_table_create_stmt := `
	 CREATE TABLE IF NOT EXISTS Volumes (
	     Id       INTEGER PRIMARY KEY AUTOINCREMENT,
	     Name     TEXT NOT NULL,
	     Type     INTEGER NOT NULL,
	     ClusterId      TEXT NOT NULL,
             Filesystem     TEXT NOT NULL,
             Fileset        TEXT NOT NULL,
             Directory      TEXT,
             MountPoint     TEXT,
             AdditionalData TEXT
         );
	`
	_, err := d.Db.Exec(volumes_table_create_stmt)

	if err != nil {
		return fmt.Errorf("Failed To Create Volumes Table: %s", err.Error())
	}

	return nil
}

func (d *dbClient) VolumeExists(name string) (bool, error) {
	d.log.Println("DatabaseClient: VolumeExists start")
	defer d.log.Println("DatabaseClient: VolumeExists end")

	volume_exists_stmt := `
	SELECT EXISTS ( SELECT Name FROM Volumes WHERE Name = ? )
	`

	stmt, err := d.Db.Prepare(volume_exists_stmt)

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

func (d *dbClient) DeleteVolume(name string) error {
	d.log.Println("DatabaseClient: DeleteVolume start")
	defer d.log.Println("DatabaseClient: DeleteVolume end")

	// Delete volume from table
	delete_volume_stmt := `
	DELETE FROM Volumes WHERE Name = ?
	`

	stmt, err := d.Db.Prepare(delete_volume_stmt)

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

func (d *dbClient) InsertFilesetVolume(fileset, volumeName string, filesystem string, opts map[string]interface{}) error {
	d.log.Println("DatabaseClient: InsertFilesetVolume start")
	defer d.log.Println("DatabaseClient: InsertFilesetVolume end")

	volume := Volume{Name: volumeName, Type: FILESET, ClusterId: d.ClusterId, FileSystem: filesystem,
		Fileset: fileset}

	addPermissionsForVolume(&volume, opts)

	return d.insertVolume(volume)
}

func (d *dbClient) InsertLightweightVolume(fileset, directory, volumeName string, filesystem string, opts map[string]interface{}) error {
	d.log.Println("DatabaseClient: InsertLightweightVolume start")
	defer d.log.Println("DatabaseClient: InsertLightweightVolume end")

	volume := Volume{Name: volumeName, Type: LIGHTWEIGHT, ClusterId: d.ClusterId, FileSystem: filesystem,
		Fileset: fileset, Directory: directory}

	addPermissionsForVolume(&volume, opts)

	return d.insertVolume(volume)
}

func (d *dbClient) InsertFilesetQuotaVolume(fileset, quota, volumeName string, filesystem string, opts map[string]interface{}) error {
	d.log.Println("DatabaseClient: InsertFilesetQuotaVolume start")
	defer d.log.Println("DatabaseClient: InsertFilesetQuotaVolume end")

	volume := Volume{Name: volumeName, Type: FILESET_WITH_QUOTA, ClusterId: d.ClusterId, FileSystem: filesystem,
		Fileset: fileset}

	volume.AdditionalData = make(map[string]string)
	volume.AdditionalData["quota"] = quota

	addPermissionsForVolume(&volume, opts)

	return d.insertVolume(volume)
}

func (d *dbClient) insertVolume(volume Volume) error {
	d.log.Println("DatabaseClient: insertVolume start")
	defer d.log.Println("DatabaseClient: insertVolume end")

	insert_volume_stmt := `
	INSERT INTO Volumes(Name, Type, ClusterId, Filesystem, Fileset, Directory, MountPoint, AdditionalData)
	values(?,?,?,?,?,?,?,?);
	`

	stmt, err := d.Db.Prepare(insert_volume_stmt)

	if err != nil {
		return fmt.Errorf("Failed to create InsertVolume Stmt for %s: %s", volume.Name, err.Error())
	}

	defer stmt.Close()

	additionalData := getAdditionalData(&volume)

	_, err = stmt.Exec(volume.Name, volume.Type, volume.ClusterId, volume.FileSystem, volume.Fileset,
		volume.Directory, volume.Mountpoint, additionalData)

	if err != nil {
		return fmt.Errorf("Failed to Insert Volume %s : %s", volume.Name, err.Error())
	}

	return nil
}

func (d *dbClient) UpdateVolumeMountpoint(name, mountpoint string) error {
	d.log.Println("DatabaseClient: UpdateVolumeMountpoint start")
	defer d.log.Println("DatabaseClient: UpdateVolumeMountpoint end")

	update_volume_stmt := `
	UPDATE Volumes
	SET MountPoint = ?
	WHERE Name = ?
	`

	stmt, err := d.Db.Prepare(update_volume_stmt)

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

func (d *dbClient) GetVolume(name string) (Volume, error) {
	d.log.Println("DatabaseClient: GetVolume start")
	defer d.log.Println("DatabaseClient: GetVolume end")

	read_volume_stmt := `
        SELECT * FROM Volumes WHERE Name = ?
        `

	stmt, err := d.Db.Prepare(read_volume_stmt)

	if err != nil {
		return Volume{}, fmt.Errorf("Failed to create GetVolume Stmt for %s : %s", name, err.Error())
	}

	defer stmt.Close()

	var volName, clusterId, filesystem, fileset, directory, mountpoint, addData string
	var volType, volId int

	err = stmt.QueryRow(name).Scan(&volId, &volName, &volType, &clusterId, &filesystem, &fileset, &directory, &mountpoint, &addData)

	if err != nil {
		return Volume{}, fmt.Errorf("Failed to Get Volume for %s : %s", name, err.Error())
	}

	scannedVolume := Volume{Id: volId, Name: volName, Type: VolumeType(volType), ClusterId: clusterId, FileSystem: filesystem,
		Fileset: fileset, Directory: directory, Mountpoint: mountpoint}

	setAdditionalData(addData, &scannedVolume)

	return scannedVolume, nil
}

func (d *dbClient) GetVolumeForMountPoint(mountpoint string) (string, error) {
	d.log.Println("DatabaseClient: GetVolumeForMountPoint start")
	defer d.log.Println("DatabaseClient: GetVolumeForMountPoint end")

	read_volume_stmt := `
        SELECT Name FROM Volumes WHERE MountPoint = ?
        `

	stmt, err := d.Db.Prepare(read_volume_stmt)

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

func (d *dbClient) ListVolumes() ([]Volume, error) {
	d.log.Println("DatabaseClient: ListVolumes start")
	defer d.log.Println("DatabaseClient: ListVolumes end")

	list_volumes_stmt := `
        SELECT *
        FROM Volumes
        `

	rows, err := d.Db.Query(list_volumes_stmt)
	defer rows.Close()

	if err != nil {
		return nil, fmt.Errorf("Failed to List Volumes : %s", err.Error())
	}

	var volumes []Volume
	var volName, clusterId, filesystem, fileset, directory, mountpoint, addData string
	var volType, volId int

	for rows.Next() {

		err = rows.Scan(&volId, &volName, &volType, &clusterId, &filesystem, &fileset, &directory, &mountpoint, &addData)

		if err != nil {
			return nil, fmt.Errorf("Failed to scan rows while listing volumes: %s", err.Error())
		}

		scannedVolume := Volume{Id: volId, Name: volName, Type: VolumeType(volType), ClusterId: clusterId,
			FileSystem: filesystem, Fileset: fileset, Directory: directory, Mountpoint: mountpoint}

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
