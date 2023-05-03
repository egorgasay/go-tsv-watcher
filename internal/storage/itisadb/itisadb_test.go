package itisadb

import (
	"context"
	"fmt"
	"github.com/egorgasay/itisadb-go-sdk"
	"github.com/pkg/errors"
	"go-tsv-watcher/internal/events"
	"reflect"
	"testing"
)

/*
	THERE IS NO WAY TO UP ITISADB IN GITHUB ACTIONS NOW
 	THAT IS WHY I HAVE TO SKIP IT
*/

func isWorking(client *itisadb.Client) bool {
	_, err := client.Index(context.Background(), "test")
	if err != nil {
		return false
	}
	return true
}

func TestItisadb_AddFilename(t *testing.T) {
	client, err := itisadb.New(":800")
	if err != nil {
		t.Logf("Can't create client %v", err)
	}

	if !isWorking(client) {
		t.Skip()
	}

	files, err := client.Index(context.Background(), "test_files")
	if err != nil {
		t.Fatal(err)
	}

	type args struct {
		ctx      context.Context
		filename string
		err      error
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "success",
			args: args{
				ctx:      context.Background(),
				filename: "filename1",
				err:      errors.New("Test success"),
			},
		},
		{
			name: "success #2",
			args: args{
				ctx:      context.Background(),
				filename: "filename2",
				err:      nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &Itisadb{
				files:  files,
				client: client,
			}
			if err = i.AddFilename(tt.args.ctx, tt.args.filename, tt.args.err); err != nil {
				t.Errorf("AddFilename() error = %v", err)
			}

			get, err := files.Get(context.Background(), tt.args.filename)
			if err != nil {
				t.Errorf("Can't get file %v", err)
			}

			if tt.args.err == nil {
				if get != "" {
					t.Errorf("AddFilename() = %v, want %v", get, tt.args.err.Error())
				} else {
					return
				}
			}

			if get != tt.args.err.Error() {
				t.Errorf("AddFilename() = %v, want %v", get, tt.args.err.Error())
			}
		})
	}
}

func TestItisadb_GetEventByNumber(t *testing.T) {
	client, err := itisadb.New(":800")
	if err != nil {
		t.Logf("Can't create client %v", err)
		t.Skip()
	}

	if !isWorking(client) {
		t.Skip()
	}

	type fields struct {
		files  *itisadb.Index
		client *itisadb.Client
	}
	type args struct {
		ctx    context.Context
		guid   string
		number int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		toSave  events.Event
		want    events.Event
		wantErr bool
	}{
		{
			name: "success",
			fields: fields{
				client: client,
			},
			args: args{
				ctx:    context.Background(),
				guid:   "guid1",
				number: 1,
			},
			want: events.Event{
				UnitGUID:    "guid1",
				MessageText: "test1",
			},
			toSave: events.Event{
				UnitGUID:    "guid1",
				MessageText: "test1",
			},
		},
		{
			name: "success second page",
			fields: fields{
				client: client,
			},
			args: args{
				ctx:    context.Background(),
				guid:   "guid1",
				number: 2,
			},
			want: events.Event{
				UnitGUID:    "guid1",
				MessageText: "test2",
			},
			toSave: events.Event{
				UnitGUID:    "guid1",
				MessageText: "test2",
			},
		},
		{
			name: "not found",
			fields: fields{
				client: client,
			},
			args: args{
				ctx:    context.Background(),
				guid:   "not found",
				number: 1,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &Itisadb{
				client: tt.fields.client,
			}

			g, err := i.client.Index(context.Background(), tt.toSave.UnitGUID)
			if err != nil {
				t.Fatalf("Can't create index %v", err)
			}

			numIndex, err := g.Index(context.Background(), fmt.Sprintf("%d", tt.args.number))
			if err != nil {
				t.Fatalf("Can't create index %v", err)
			}

			err = numIndex.Set(context.Background(), "MessageText", tt.toSave.MessageText, false)
			if err != nil {
				t.Errorf("Can't set index %v", err)
			}

			err = numIndex.Set(context.Background(), "UnitGUID", tt.toSave.UnitGUID, false)
			if err != nil {
				t.Errorf("Can't set index %v", err)
			}

			got, err := i.GetEventByNumber(tt.args.ctx, tt.args.guid, tt.args.number)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetEventByNumber() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetEventByNumber() got = %v, want %v", got, tt.want)
			}
		})
	}
}

type addStub map[string]struct{}

func (s *addStub) AddFile(filename string) {
	(*s)[filename] = struct{}{}
}

func TestItisadb_LoadFilenames(t *testing.T) {
	client, err := itisadb.New(":800")
	if err != nil {
		t.Logf("Can't create client %v", err)
		t.Skip()
	}

	if !isWorking(client) {
		t.Skip()
	}

	files, err := client.Index(context.Background(), "files")
	if err != nil {
		t.Fatal(err)
	}

	st := &Itisadb{
		client: client,
		files:  files,
	}

	tests := []struct {
		name      string
		filename  string
		err       error
		wantError bool
	}{
		{
			name:      "ok #1",
			filename:  "TestLoadFilenames.tsv",
			err:       nil,
			wantError: false,
		},
		{
			name:      "ok #2",
			filename:  "TestLoadFilenames2.tsv",
			err:       errors.New("testError"),
			wantError: false,
		},
		{
			name:      "ok #3",
			filename:  "TestLoadFilenames4.tsv",
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

type ieventsStub struct {
	events []events.Event
}

func (i ieventsStub) Fill() error {
	return nil
}

func (i ieventsStub) Print() {

}

func (i ieventsStub) Iter(cb func(d events.Event) (stop bool)) {
	for _, d := range i.events {
		if stop := cb(d); stop {
			return
		}
	}
}

func TestItisadb_SaveEvents(t *testing.T) {
	client, err := itisadb.New(":800")
	if err != nil {
		t.Logf("Can't create client %v", err)
		t.Skip()
	}

	if !isWorking(client) {
		t.Skip()
	}

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
						UnitGUID: "3992bf73-76af-438b-9e75-085348da7f6a",
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
						UnitGUID: "268cb81b-c82f-4c0c-bf4a-cb6f5fb89ceb",
					},
					{
						ID:       "3",
						UnitGUID: "9132dbdf-5991-4a56-bc0c-5b3e3d6777bf",
					},
					{
						ID:       "4",
						UnitGUID: "fbc6af89-a89c-4fd7-8d08-c8d9e39adde2",
					},
					{
						ID:       "5",
						UnitGUID: "cb33fc64-94e0-4fa6-9a64-d267e56c1c91",
					},
				},
			},
			wantErr: false,
		},
	}

	ctx := context.TODO()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &Itisadb{
				client: client,
			}
			if err := i.SaveEvents(ctx, tt.evs); (err != nil) != tt.wantErr {
				t.Errorf("SaveEvents() error = %v, wantErr %v", err, tt.wantErr)
			}

			for j, e := range tt.evs.events {
				guidIndex, err := i.client.Index(ctx, e.UnitGUID)
				if err != nil {
					t.Fatalf("failed to create or get guid index: %v", err)
				}

				numIndex, err := guidIndex.Index(ctx, fmt.Sprintf("1"))
				if err != nil {
					t.Fatalf("failed to create or get index: %v", err)
				}

				get, err := numIndex.Get(ctx, "ID")
				if err != nil {
					t.Fatalf("failed to get ID: %v", err)
				}

				if get != tt.evs.events[j].ID {
					t.Errorf("got = %v, want %v", get, tt.evs.events[j].ID)
				}

				guidIndex.Delete(ctx)
			}

		})
	}

}
