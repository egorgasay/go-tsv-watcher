package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"go-tsv-watcher/internal/events"
	"go-tsv-watcher/internal/storage/itisadb"
	"go-tsv-watcher/internal/storage/postgres"
	"go-tsv-watcher/internal/storage/queries"
	"go-tsv-watcher/internal/storage/service"
	"go-tsv-watcher/internal/storage/sqlite"
)

type Database interface {
	LoadFilenames(putter service.Adder) error
	AddFilename(filename string, err error) error

	SaveEvents(evs service.IEvents) error
	GetEventByNumber(guid string, number int) (events.Event, error) // TODO: ADD CONTEXT
}

type Storage Database

type Config struct {
	Type           string
	DataSourceCred string
}

func New(cfg *Config) (Storage, error) {
	var st Storage
	var err error
	var db *sql.DB

	switch cfg.Type {
	case "postgres":
		db, err = sql.Open("postgres", cfg.DataSourceCred)
		if err != nil {
			panic(err)
		}

		st = postgres.New(db, "file://migrations/sqlite3")
	case "sqlite3":
		db, err = sql.Open("sqlite3", cfg.DataSourceCred)
		if err != nil {
			panic(err)
		}

		st = sqlite.New(db, "file://migrations/sqlite3")
	case "itisadb":
		nosql, err := itisadb.New(cfg.DataSourceCred)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to itisadb: %w", err)
		}

		return nosql, nil
	default:
		return nil, errors.New("unknown database type")
	}

	if err = queries.Prepare(db, cfg.Type); err != nil {
		return nil, err
	}

	return st, nil
}
