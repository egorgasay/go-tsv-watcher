package usecase

import (
	"context"
	"errors"
	"github.com/go-chi/httplog"
	"github.com/golang/mock/gomock"
	"go-tsv-watcher/internal/events"
	"go-tsv-watcher/internal/storage/mocks"
	"go-tsv-watcher/internal/storage/service"
	"go-tsv-watcher/pkg/logger"
	"os"
	"reflect"
	"strings"
	"testing"
)

type mockBehavior func(r *mocks.MockStorage)

func TestUseCase_GetEventByNumber(t *testing.T) {
	type args struct {
		ctx      context.Context
		unitGUID string
		number   int
	}
	tests := []struct {
		name         string
		args         args
		want         events.Event
		wantErr      bool
		mockBehavior func(r *mocks.MockStorage)
	}{
		{
			name: "Ok",
			args: args{
				ctx:      context.Background(),
				unitGUID: "testUseCase",
				number:   1,
			},
			want: events.Event{
				Number:   3,
				UnitGUID: "testUseCase",
				ID:       "testID",
			},
			mockBehavior: func(r *mocks.MockStorage) {
				r.EXPECT().GetEventByNumber(gomock.Any(), gomock.Any(), gomock.Any()).Return(events.Event{
					Number:   3,
					UnitGUID: "testUseCase",
					ID:       "testID",
				}, nil)
			},
		},
		{
			name: "not found",
			args: args{
				ctx:      context.Background(),
				unitGUID: "testUseCase",
				number:   2,
			},
			want: events.Event{},
			mockBehavior: func(r *mocks.MockStorage) {
				r.EXPECT().GetEventByNumber(gomock.Any(), gomock.Any(), gomock.Any()).Return(events.Event{}, service.ErrEventNotFound)
			},
			wantErr: true,
		},
		{
			name: "storage error",
			args: args{
				ctx:      context.Background(),
				unitGUID: "testStorage",
				number:   54,
			},
			want: events.Event{},
			mockBehavior: func(r *mocks.MockStorage) {
				r.EXPECT().GetEventByNumber(gomock.Any(), gomock.Any(), gomock.Any()).Return(events.Event{}, errors.New("test error"))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()
			st := mocks.NewMockStorage(c)
			tt.mockBehavior(st)

			loggerInstance := httplog.NewLogger("watcher", httplog.Options{
				Concise: true,
			})

			u := &UseCase{storage: st, logger: logger.New(loggerInstance)}
			got, err := u.GetEventByNumber(tt.args.ctx, tt.args.unitGUID, tt.args.number)
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

type eventStub struct {
	events []events.Event
}

func (i eventStub) Fill() error {
	return nil
}

func (i eventStub) Print() {

}

func (i eventStub) Iter(cb func(d events.Event) (stop bool)) {
	for _, d := range i.events {
		if stop := cb(d); stop {
			return
		}
	}
}

func TestUseCase_savePDF(t *testing.T) {
	loggerInstance := httplog.NewLogger("watcher", httplog.Options{
		Concise: true,
	})

	dir := "test_dir/testSavePDF"

	err := os.Chdir("../../")
	if err != nil {
		return
	}

	err = os.Mkdir(dir, 0666)
	if err != nil {
		if !os.IsExist(err) {
			t.Fatalf("create() error = %v", err)
		}
	}

	uc := New(nil, dir, logger.New(loggerInstance))

	es := eventStub{events: []events.Event{
		{UnitGUID: "1"}, {UnitGUID: "2"}, {UnitGUID: "3"}, {UnitGUID: "4"}, {UnitGUID: "5"},
	}}

	err = uc.savePDF(es)
	if err != nil {
		t.Fatalf("savePDF() error = %v", err)
	}

	directory, err := os.Open(dir)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}

	fis, err := directory.Readdir(-1)
	if err != nil {
		t.Fatalf("Readdir() error = %v", err)
	}

loop:
	for _, f := range fis {
		if f.IsDir() {
			continue
		}
		for i, e := range es.events {
			longFilename := f.Name()
			shortFilename := strings.Trim(longFilename, ".pdf")
			if shortFilename == e.UnitGUID {
				es.events = append(es.events[:i], es.events[i+1:]...)
				continue loop
			}
		}
	}

	if len(es.events) != 0 {
		t.Fatalf("savePDF() missing files")
	}

	err = os.RemoveAll(dir)
	if err != nil {
		t.Fatalf("remove() error = %v", err)
	}
}
