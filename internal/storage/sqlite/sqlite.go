package sqlite

import (
	"database/sql"
	"github.com/pkg/errors"
	"go-tsv-watcher/internal/storage/sqllike"
	"log"

	// SQLite driver
	_ "github.com/mattn/go-sqlite3"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"

	// file driver
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type Sqlite3 struct {
	sqllike.DB
}

// New Sqlite3 constructor.
func New(db *sql.DB, path string) *Sqlite3 {
	driver, err := sqlite.WithInstance(db, &sqlite.Config{})
	if err != nil {
		log.Fatal(err)
		return nil
	}

	m, err := migrate.NewWithDatabaseInstance(
		path,
		"sqlite", driver)
	if err != nil {
		log.Fatal(err)
		return nil
	}

	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatal(err)
	}

	return &Sqlite3{DB: sqllike.DB{DB: db}}
}
