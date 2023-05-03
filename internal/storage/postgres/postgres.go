package postgres

import (
	"database/sql"
	"errors"
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	// File driver
	_ "github.com/golang-migrate/migrate/v4/source/file"
	// Postgres driver
	_ "github.com/jackc/pgx"
	"go-tsv-watcher/internal/storage/sqllike"
)

// Postgres struct for the postgres db
type Postgres struct {
	sqllike.DB
}

// New Postgres constructor.
func New(db *sql.DB, path string) *Postgres {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Fatal(err)
		return nil
	}

	m, err := migrate.NewWithDatabaseInstance(
		path,
		"postgres", driver)
	if err != nil {
		log.Fatal(err)
		return nil
	}

	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatal(err)
	}

	return &Postgres{DB: sqllike.DB{DB: db}}
}
