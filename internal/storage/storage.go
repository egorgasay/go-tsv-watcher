package storage

import (
	"context"
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

// Database interface
type Database interface {
	LoadFilenames(ctx context.Context, putter service.Adder) error
	AddFilename(ctx context.Context, filename string, err error) error

	SaveEvents(ctx context.Context, evs service.IEvents) error
	GetEventByNumber(ctx context.Context, guid string, number int) (events.Event, error)
}

// Storage interface for storage
//
//go:generate mockgen -destination=mocks/mock_storage.go -package=mocks go-tsv-watcher/internal/storage Storage
type Storage Database

// Config for storage
type Config struct {
	Type           string
	DataSourceCred string
}

// New storage
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
		nosql, err := itisadb.New(context.Background(), cfg.DataSourceCred)
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
