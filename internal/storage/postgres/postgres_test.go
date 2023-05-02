package postgres_test

import (
	"context"
	"database/sql"
	"errors"
	"github.com/egorgasay/dockerdb/v2"
	"go-tsv-watcher/internal/events"
	"go-tsv-watcher/internal/storage/postgres"
	"go-tsv-watcher/internal/storage/queries"
	"log"
	"testing"
)

var st *postgres.Postgres

func TestMain(m *testing.M) {
	cfg := dockerdb.CustomDB{
		DB: dockerdb.DB{
			Name:     "admin",
			User:     "admin",
			Password: "XXXXX",
		},
		Port:   "1234",
		Vendor: dockerdb.Postgres15,
	}

	err := dockerdb.Pull(context.Background(), dockerdb.Postgres15)
	if err != nil {
		return
	}

	ddb, err := dockerdb.New(context.Background(), cfg)
	if err != nil {
		log.Fatalf("can't create db: %v", err)
	}

	db, err := sql.Open("sqlite3", ddb.ConnString)
	if err != nil {
		log.Fatalf("can't opening the db: %v", err)
	}
	defer cleanup(db)
	defer db.Close()

	st = postgres.New(db, "file://..//..//..//migrations/postgres")

	err = queries.Prepare(db, "sqlite3")
	if err != nil {
		log.Fatalf("error preparing db: %v", err)
	}

	m.Run()

}

func cleanup(db *sql.DB) {
	_, err := db.Exec("DROP TABLE IF EXISTS events")
	if err != nil {
		log.Fatalf("error dropping table: %v", err)
	}
	_, err = db.Exec("DROP TABLE IF EXISTS files")
	if err != nil {
		log.Fatalf("error dropping table: %v", err)
	}
}

type addStub map[string]struct{}

func (s *addStub) AddFile(filename string) {
	(*s)[filename] = struct{}{}
}

// TestAddFilename
func TestAddFilename(t *testing.T) {
	tests := []struct {
		name      string
		filename  string
		err       error
		wantError bool
	}{
		{
			name:      "ok #1",
			filename:  "test.tsv",
			err:       nil,
			wantError: false,
		},
		{
			name:      "duplicate",
			filename:  "test.tsv",
			err:       nil,
			wantError: true,
		},
		{
			name:      "ok #2",
			filename:  "test2.tsv",
			err:       errors.New("testError"),
			wantError: false,
		},
		{
			name:      "ok #3",
			filename:  "test3.tsv",
			wantError: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := st.AddFilename(tt.filename, tt.err)
			if (err != nil) != tt.wantError {
				t.Errorf("error adding filename: %v", err)
			}
		})

		if tt.wantError {
			continue
		}

		a := &addStub{}

		err := st.LoadFilenames(a)
		if err != nil {
			t.Fatalf("error loading filenames: %v", err)
		}

		if _, ok := (*a)["test.tsv"]; !ok {
			t.Fatalf("filename not found")
		}
	}

}

// TestLoadFilenames
func TestLoadFilenames(t *testing.T) {
	tests := []struct {
		name      string
		filename  string
		err       error
		wantError bool
	}{
		{
			name:      "ok #1",
			filename:  "test.tsv",
			err:       nil,
			wantError: false,
		},
		{
			name:      "duplicate",
			filename:  "test.tsv",
			err:       nil,
			wantError: true,
		},
		{
			name:      "ok #2",
			filename:  "test2.tsv",
			err:       errors.New("testError"),
			wantError: false,
		},
		{
			name:      "ok #3",
			filename:  "test3.tsv",
			wantError: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := st.AddFilename(tt.filename, tt.err)
			if (err != nil) != tt.wantError {
				t.Errorf("error adding filename: %v", err)
			}
		})
	}

	a := &addStub{}
	err := st.LoadFilenames(a)
	if err != nil {
		t.Fatalf("error loading filenames: %v", err)
	}

	for _, tt := range tests {
		if !tt.wantError {
			continue
		}

		if _, ok := (*a)[tt.filename]; !ok {
			t.Fatalf("filename not found")
		}
	}

	if len(*a) != 3 {
		t.Fatalf("unexpected number of filenames %d", len(*a))
	}
}

type ieventsStub struct {
	events []events.Event
}

func (i ieventsStub) Fill() error {
	return nil
}

func (i ieventsStub) Print() {
	return
}

func (i ieventsStub) Iter(cb func(d events.Event) (stop bool)) {
	for _, d := range i.events {
		if stop := cb(d); stop {
			return
		}
	}
}

func TestDB_SaveEvents(t *testing.T) {
	tests := []struct {
		name    string
		evs     *ieventsStub
		wantErr bool
	}{
		{
			name: "ok #1",
			evs: &ieventsStub{
				events: []events.Event{
					{
						ID:       "1",
						UnitGUID: "1",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "ok #2",
			evs: &ieventsStub{
				events: []events.Event{
					{
						ID:       "2",
						UnitGUID: "2",
					},
					{
						ID:       "3",
						UnitGUID: "3",
					},
					{
						ID:       "4",
						UnitGUID: "4",
					},
					{
						ID:       "5",
						UnitGUID: "5",
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := st.SaveEvents(tt.evs); (err != nil) != tt.wantErr {
				t.Errorf("SaveEvents() error = %v, wantErr %v", err, tt.wantErr)
			}
			for _, ev := range tt.evs.events {
				var id string
				if err := st.DB.QueryRow("SELECT ID FROM events WHERE UnitGUID = $1", ev.UnitGUID).Scan(&id); err != nil {
					t.Errorf("error getting id: %v", err)
				}
				if id != ev.ID {
					t.Errorf("unexpected id: %s want %s", id, ev.ID)
				}
			}
		})
	}
}
