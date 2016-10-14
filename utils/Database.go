package utils

import (
	"database/sql"
	"fmt"
	"log"
	"path"

	"strings"

	_ "github.com/mattn/go-sqlite3"
)

type DatabaseClient struct {
	Filesystem string
	Mountpoint string
	log        *log.Logger
	Db         *sql.DB
	ClusterId  string
}

const (
	FILESET = iota
	LIGHTWEIGHT
	FILESET_QUOTA
)

type Volume struct {
	VolumeId       int
	VolumeName     string
	VolumeType     int
	ClusterId      string
	FileSystem     string
	Fileset        string
	Directory      string
	Mountpoint     string
	AdditionalData map[string]string
}

func NewDatabaseClient(log *log.Logger, filesystem, mountpoint string) *DatabaseClient {
	return &DatabaseClient{log: log, Filesystem: filesystem, Mountpoint: mountpoint}
}

func (d *DatabaseClient) Init() error {

	d.log.Println("DatabaseClient: DB Init start")
	defer d.log.Println("DatabaseClient: DB Init end")

	dbFile := "spectrum-scale-" + d.Filesystem + ".db"
	dbPath := path.Join(d.Mountpoint, dbFile)

	Db, err := sql.Open("sqlite3", dbPath)

	if err != nil {
		return fmt.Errorf("Failed to Initialize DB connection to %s: %s\n", dbPath, err.Error())
	}
	d.Db = Db
	d.log.Printf("Established Database connection to %s via go-sqlite3 driver", dbPath)
	return nil
}

func (d *DatabaseClient) Close() error {

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

func (d *DatabaseClient) CreateVolumeTable() error {
	d.log.Println("DatabaseClient: Create Volumes Table start")
	defer d.log.Println("DatabaseClient: Create Volumes Table end")

	volumes_table_create_stmt := `
	 CREATE TABLE IF NOT EXISTS Volumes (
	     VolumeId       INTEGER PRIMARY KEY AUTOINCREMENT,
	     VolumeName     TEXT NOT NULL,
	     VolumeType     INTEGER NOT NULL,
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

func (d *DatabaseClient) VolumeExists(name string) (bool, error) {
	d.log.Println("DatabaseClient: VolumeExists start")
	defer d.log.Println("DatabaseClient: VolumeExists end")

	volume_exists_stmt := `
	SELECT EXISTS ( SELECT VolumeName FROM Volumes WHERE VolumeName = ? )
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

func (d *DatabaseClient) DeleteVolume(name string) error {
	d.log.Println("DatabaseClient: DeleteVolume start")
	defer d.log.Println("DatabaseClient: DeleteVolume end")

	// Delete volume from table
	delete_volume_stmt := `
	DELETE FROM Volumes WHERE VolumeName = ?
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

func (d *DatabaseClient) InsertFilesetVolume(fileset, volumeName string) error {
	d.log.Println("DatabaseClient: InsertFilesetVolume start")
	defer d.log.Println("DatabaseClient: InsertFilesetVolume end")

	volume := &Volume{VolumeName: volumeName, VolumeType: FILESET, ClusterId: d.ClusterId, FileSystem: d.Filesystem,
		Fileset: fileset}

	return d.insertVolume(volume)
}

func (d *DatabaseClient) InsertLightweightVolume(fileset, directory, volumeName string) error {
	d.log.Println("DatabaseClient: InsertLightweightVolume start")
	defer d.log.Println("DatabaseClient: InsertLightweightVolume end")

	volume := &Volume{VolumeName: volumeName, VolumeType: LIGHTWEIGHT, ClusterId: d.ClusterId, FileSystem: d.Filesystem,
		Fileset: fileset, Directory: directory}

	return d.insertVolume(volume)
}

func (d *DatabaseClient) InsertFilesetQuotaVolume(fileset, quota, volumeName string) error {
	d.log.Println("DatabaseClient: InsertFilesetQuotaVolume start")
	defer d.log.Println("DatabaseClient: InsertFilesetQuotaVolume end")

	volume := &Volume{VolumeName: volumeName, VolumeType: FILESET_QUOTA, ClusterId: d.ClusterId, FileSystem: d.Filesystem,
		Fileset: fileset}

	volume.AdditionalData = make(map[string]string)
	volume.AdditionalData["quota"] = quota

	return d.insertVolume(volume)
}

func (d *DatabaseClient) insertVolume(volume *Volume) error {
	d.log.Println("DatabaseClient: insertVolume start")
	defer d.log.Println("DatabaseClient: insertVolume end")

	insert_volume_stmt := `
	INSERT INTO Volumes(VolumeName, VolumeType, ClusterId, Filesystem, Fileset, Directory, MountPoint, AdditionalData)
	values(?,?,?,?,?,?,?,?);
	`

	stmt, err := d.Db.Prepare(insert_volume_stmt)

	if err != nil {
		return fmt.Errorf("Failed to create InsertVolume Stmt for %s: %s", volume.VolumeName, err.Error())
	}

	defer stmt.Close()

	additionalData := getAdditionalData(volume)

	_, err = stmt.Exec(volume.VolumeName, volume.VolumeType, volume.ClusterId, volume.FileSystem, volume.Fileset,
		volume.Directory, volume.Mountpoint, additionalData)

	if err != nil {
		return fmt.Errorf("Failed to Insert Volume %s : %s", volume.VolumeName, err.Error())
	}

	return nil
}

func (d *DatabaseClient) UpdateVolumeMountpoint(name, mountpoint string) error {
	d.log.Println("DatabaseClient: UpdateVolumeMountpoint start")
	defer d.log.Println("DatabaseClient: UpdateVolumeMountpoint end")

	update_volume_stmt := `
	UPDATE Volumes
	SET MountPoint = ?
	WHERE VolumeName = ?
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

func (d *DatabaseClient) GetVolume(name string) (*Volume, error) {
	d.log.Println("DatabaseClient: GetVolume start")
	defer d.log.Println("DatabaseClient: GetVolume end")

	read_volume_stmt := `
        SELECT * FROM Volumes WHERE VolumeName = ?
        `

	stmt, err := d.Db.Prepare(read_volume_stmt)

	if err != nil {
		return nil, fmt.Errorf("Failed to create GetVolume Stmt for %s : %s", name, err.Error())
	}

	defer stmt.Close()

	var volName, clusterId, filesystem, fileset, directory, mountpoint, addData string
	var volType, volId int

	err = stmt.QueryRow(name).Scan(&volId, &volName, &volType, &clusterId, &filesystem, &fileset, &directory, &mountpoint, &addData)

	if err != nil {
		return nil, fmt.Errorf("Failed to Get Volume for %s : %s", name, err.Error())
	}

	scannedVolume := &Volume{VolumeId: volId, VolumeName: volName, VolumeType: volType, ClusterId: clusterId, FileSystem: filesystem,
		Fileset: fileset, Directory: directory, Mountpoint: mountpoint}

	setAdditionalData(addData, scannedVolume)

	return scannedVolume, nil
}

func (d *DatabaseClient) GetVolumeForMountPoint(mountpoint string) (string, error) {
	d.log.Println("DatabaseClient: GetVolumeForMountPoint start")
	defer d.log.Println("DatabaseClient: GetVolumeForMountPoint end")

	read_volume_stmt := `
        SELECT VolumeName FROM Volumes WHERE MountPoint = ?
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

func (d *DatabaseClient) ListVolumes() ([]Volume, error) {
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

		scannedVolume := Volume{VolumeId: volId, VolumeName: volName, VolumeType: volType, ClusterId: clusterId,
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
