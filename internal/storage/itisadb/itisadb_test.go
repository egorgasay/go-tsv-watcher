package itisadb

import (
	"context"
	"github.com/egorgasay/itisadb-go-sdk"
	"go-tsv-watcher/internal/events"
	"go-tsv-watcher/internal/storage/service"
	"reflect"
	"testing"
)

func TestItisadb_AddFilename(t *testing.T) {
	client, err := itisadb.New(":800")
	if err != nil {
		t.Fatal(err)
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
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				ctx:      context.Background(),
				filename: "filename1",
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
			if err := i.AddFilename(tt.args.ctx, tt.args.filename, tt.args.err); (err != nil) != tt.wantErr {
				t.Errorf("AddFilename() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestItisadb_GetEventByNumber(t *testing.T) {
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
		want    events.Event
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &Itisadb{
				files:  tt.fields.files,
				client: tt.fields.client,
			}
			got, err := i.GetEventByNumber(tt.args.ctx, tt.args.guid, tt.args.number)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetEventByNumber() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetEventByNumber() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestItisadb_LoadFilenames(t *testing.T) {
	type fields struct {
		files  *itisadb.Index
		client *itisadb.Client
	}
	type args struct {
		ctx   context.Context
		adder service.Adder
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &Itisadb{
				files:  tt.fields.files,
				client: tt.fields.client,
			}
			if err := i.LoadFilenames(tt.args.ctx, tt.args.adder); (err != nil) != tt.wantErr {
				t.Errorf("LoadFilenames() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestItisadb_SaveEvents(t *testing.T) {
	type fields struct {
		files  *itisadb.Index
		client *itisadb.Client
	}
	type args struct {
		ctx context.Context
		evs service.IEvents
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &Itisadb{
				files:  tt.fields.files,
				client: tt.fields.client,
			}
			if err := i.SaveEvents(tt.args.ctx, tt.args.evs); (err != nil) != tt.wantErr {
				t.Errorf("SaveEvents() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNew(t *testing.T) {
	type args struct {
		ctx   context.Context
		creds string
	}
	tests := []struct {
		name    string
		args    args
		want    *Itisadb
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.args.ctx, tt.args.creds)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() got = %v, want %v", got, tt.want)
			}
		})
	}
}
