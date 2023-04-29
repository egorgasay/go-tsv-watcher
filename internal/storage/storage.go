package storage

import (
	"database/sql"
	"go-tsv-watcher/internal/devices"
	"go-tsv-watcher/internal/storage/base"
	"go-tsv-watcher/internal/storage/postgres"
	"go-tsv-watcher/internal/storage/sqlite"
)

type Database interface {
	Prepare(filename string) error

	LoadFilenames(putter base.Adder) error
	AddFilename(filename string, err error) error

	AddRelations(filename string, number []string) error

	SaveDevices(devs *devices.Devices) error
	GetEventByNumber(guid string, number int) (devices.Device, error)
}

type Storage Database

type Config struct {
	Type           string
	DataSourceCred string
}

func New(cfg *Config) Storage {
	var st Storage

	switch cfg.Type {
	case "postgres":
		db, err := sql.Open("postgres", cfg.DataSourceCred)
		if err != nil {
			panic(err)
		}

		st = postgres.New(db, "file://migrations/sqlite3")
	case "sqlite3":
		db, err := sql.Open("sqlite3", cfg.DataSourceCred)
		if err != nil {
			panic(err)
		}

		st = sqlite.New(db, "file://migrations/sqlite3")
	default:
		panic("unknown database type")
	}

	return st
}
