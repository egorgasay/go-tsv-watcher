package sqlite

import (
	"go-tsv-watcher/internal/storage/base"

	_ "github.com/mattn/go-sqlite3"
)

type Sqlite struct {
	db base.DB
}
