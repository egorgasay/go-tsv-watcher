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
	"reflect"
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
