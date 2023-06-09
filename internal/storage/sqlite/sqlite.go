package sqlite

import (
	"database/sql"
	"github.com/pkg/errors"
	"go-tsv-watcher/internal/storage/sqllike"
	"go-tsv-watcher/pkg/logger"
	"log"

	// SQLite driver
	_ "modernc.org/sqlite"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"

	// file driver
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// Sqlite3 struct for the sqlite3 database.
type Sqlite3 struct {
	sqllike.DB
}

// New Sqlite3 constructor.
func New(db *sql.DB, path string, logger logger.ILogger) *Sqlite3 {
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

	bdb := sqllike.New(db, logger)

	return &Sqlite3{DB: *bdb}
}
