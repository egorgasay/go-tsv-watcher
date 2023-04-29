package postgres

import (
	"database/sql"
	"errors"
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/jackc/pgx"
	"go-tsv-watcher/internal/storage/base"
)

type Postgres struct {
	base.DB
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
	if !errors.Is(err, migrate.ErrNoChange) {
		log.Fatal(err)
	}

	return &Postgres{DB: base.DB{DB: db}}
}
