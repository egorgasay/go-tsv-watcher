package sqlite_test

import (
	"database/sql"
	"go-tsv-watcher/internal/storage/queries"
	"go-tsv-watcher/internal/storage/sqlite"
	"testing"
)

type addStub map[string]struct{}

func (s *addStub) AddFile(filename string) {
	(*s)[filename] = struct{}{}
}

// TestAddFilename
func TestAddFilename(t *testing.T) {
	db, err := sql.Open("sqlite3", "test.db")
	if err != nil {
		t.Fatalf("error opening db: %v", err)
	}
	defer db.Close()

	st := sqlite.New(db, "file://..//..//..//migrations/sqlite3")

	err = queries.Prepare(db, "sqlite3")
	if err != nil {
		t.Fatalf("error preparing db: %v", err)
	}

	err = st.AddFilename("test.tsv", nil)
	if err != nil {
		t.Fatalf("error adding filename: %v", err)
	}

	a := &addStub{}

	err = st.LoadFilenames(a)
	if err != nil {
		t.Fatalf("error loading filenames: %v", err)
	}

	if _, ok := (*a)["test.tsv"]; !ok {
		t.Fatalf("filename not found")
	}
}
