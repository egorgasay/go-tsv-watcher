package postgres_test

import (
	"context"
	"database/sql"
	"errors"
	"github.com/docker/distribution/uuid"
	"github.com/egorgasay/dockerdb/v2"
	"github.com/go-chi/httplog"
	"go-tsv-watcher/internal/events"
	"go-tsv-watcher/internal/storage/postgres"
	"go-tsv-watcher/internal/storage/queries"
	"go-tsv-watcher/pkg/logger"
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
		log.Fatalf("can't pull postgres: %v", err)
	}

	ddb, err := dockerdb.New(context.Background(), cfg)
	if err != nil {
		log.Fatalf("can't create db: %v", err)
	}

	db, err := sql.Open("postgres", ddb.ConnString)
	if err != nil {
		log.Fatalf("can't opening the db: %v", err)
	}
	defer db.Close()
	defer cleanup(db)

	loggerInstance := httplog.NewLogger("watcher", httplog.Options{
		Concise: true,
	})

	st = postgres.New(db, "file://..//..//..//migrations/postgres", logger.New(loggerInstance))

	err = queries.Prepare(db, "postgres")
	if err != nil {
		log.Fatalf("error preparing db: %v", err)
	}

	m.Run()

}

func cleanup(db *sql.DB) {
	_, err := db.Exec("SELECT 'drop table if exists \"' || tablename || '\" cascade;' as pg_tbl_drop FROM pg_tables WHERE schemaname='public';")
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
	_, err := st.DB.Exec("DELETE FROM files")
	if err != nil {
		t.Fatalf("error deleting files: %v", err)
	}

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
			err := st.AddFilename(context.Background(), tt.filename, tt.err)
			if (err != nil) != tt.wantError {
				t.Errorf("error adding filename: %v", err)
			}
		})

		if tt.wantError {
			continue
		}

		a := &addStub{}

		err := st.LoadFilenames(context.Background(), a)
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
	_, err := st.DB.Exec("DELETE FROM files")
	if err != nil {
		t.Fatalf("error deleting files: %v", err)
	}

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
			err := st.AddFilename(context.Background(), tt.filename, tt.err)
			if (err != nil) != tt.wantError {
				t.Errorf("error adding filename: %v", err)
			}
		})
	}

	a := &addStub{}

	err = st.LoadFilenames(context.Background(), a)

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

// ieventsStub is a stub for events.
type ieventsStub struct {
	events []events.Event
}

// Fill unused
func (i ieventsStub) Fill() error {
	return nil
}

// Print unused
func (i ieventsStub) Print() {

}

// Iter iterates over all the events.
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
						UnitGUID: "1",
						ID:       "3992bf73-76af-438b-9e75-085348da7f6a",
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
						UnitGUID: "2",
						ID:       "268cb81b-c82f-4c0c-bf4a-cb6f5fb89ceb",
					},
					{
						UnitGUID: "3",
						ID:       "9132dbdf-5991-4a56-bc0c-5b3e3d6777bf",
					},
					{
						UnitGUID: "4",
						ID:       "fbc6af89-a89c-4fd7-8d08-c8d9e39adde2",
					},
					{
						UnitGUID: "5",
						ID:       "cb33fc64-94e0-4fa6-9a64-d267e56c1c91",
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := st.SaveEvents(context.Background(), tt.evs); (err != nil) != tt.wantErr {
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

func TestDB_GetEventByNumber(t *testing.T) {
	type args struct {
		ctx    context.Context
		guid   string
		number int
	}
	tests := []struct {
		name    string
		args    args
		want    events.Event
		wantErr bool
	}{
		{
			name: "ok #1",
			args: args{
				ctx:    context.Background(),
				guid:   "3992bf73-76af-438b-9e75-085348da7f61",
				number: 1,
			},

			want: events.Event{
				UnitGUID:    "3992bf73-76af-438b-9e75-085348da7f61",
				MessageText: "Test2",
			},

			wantErr: false,
		},

		{
			name: "ok #2",
			args: args{
				ctx:    context.Background(),
				guid:   "3992bf73-76af-438b-9e75-085348da7f61",
				number: 2,
			},

			want: events.Event{
				UnitGUID:    "3992bf73-76af-438b-9e75-085348da7f61",
				MessageText: "Test2",
			},

			wantErr: false,
		},
		{
			name: "not found",
			args: args{
				ctx:    context.Background(),
				guid:   "3992bf73-76af-438b-9e75-085348da",
				number: 3,
			},
			wantErr: true,
		},
	}

	query := `
INSERT INTO events (
                    ID, UnitGUID, MessageText, Number, Level, Bit, 
                    InvertBit, MQTT, InventoryID, MessageID, MessageClass, 
                    Context, Area, Address, Block, Type) 
VALUES ($1, $2, $3, 0, 0, 0, 0, '', '', '', '', '','', '', false, '')`

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := st.DB.Exec(query, uuid.Generate().String(), tt.want.UnitGUID, tt.want.MessageText)
			if err != nil {
				t.Fatalf("error adding mock event: %v", err)
			}

			got, err := st.GetEventByNumber(tt.args.ctx, tt.args.guid, tt.args.number)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetEventByNumber() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got.UnitGUID != tt.want.UnitGUID {
				t.Errorf("GetEventByNumber() got = %v, want %v", got, tt.want)
			} else if got.MessageText != tt.want.MessageText {
				t.Errorf("GetEventByNumber() got = %v, want %v", got, tt.want)
			}
		})
	}
}
