package utils

import (
	"database/sql"
	"fmt"
	"log"
	"path"

	_ "github.com/mattn/go-sqlite3"
)

//go:generate counterfeiter -o ../fakes/fake_database_client.go . DatabaseClient
type DatabaseClient interface {
	Init() error
	Close() error
	GetHandle() *sql.DB
}

type dbClient struct {
	mountpoint string // where DB file(s) is stored
	log        *log.Logger
	db         *sql.DB
}

func NewDatabaseClient(log *log.Logger, mountpoint string) DatabaseClient {
	return &dbClient{log: log, mountpoint: mountpoint}
}

func (d *dbClient) Init() error {

	d.log.Println("DatabaseClient: DB Init start")
	defer d.log.Println("DatabaseClient: DB Init end")

	dbFile := "spectrum-scale.db"
	dbPath := path.Join(d.mountpoint, dbFile)

	db, err := sql.Open("sqlite3", dbPath)

	if err != nil {
		return fmt.Errorf("Failed to Initialize DB connection to %s: %s\n", dbPath, err.Error())
	}
	d.db = db
	d.log.Printf("Established Database connection to %s via go-sqlite3 driver", dbPath)
	return nil
}

func (d *dbClient) Close() error {

	d.log.Println("DatabaseClient: DB Close start")
	defer d.log.Println("DatabaseClient: DB Close end")

	if d.db != nil {
		err := d.db.Close()
		if err != nil {
			return fmt.Errorf("Failed to close DB connection: %s\n", err.Error())
		}
	}
	return nil
}

func (d *dbClient) GetHandle() *sql.DB {
	return d.db
}
