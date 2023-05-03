package handler

import (
	"github.com/go-chi/chi/v5"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go-tsv-watcher/internal/events"
	"go-tsv-watcher/internal/storage/service"
	"go-tsv-watcher/internal/usecase"
	mocks "go-tsv-watcher/internal/usecase/mocks"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandler_GetEvent(t *testing.T) {
	type mockBehavior func(r *mocks.MockIUseCase)

	url := "http://localhost:8080/api/v1/event"
	tests := []struct {
		name               string
		unitGUID           string
		num                int
		body               string
		expectedBody       string
		mockBehavior       mockBehavior
		expectedStatusCode int
	}{
		{
			name:               "Ok",
			body:               `{"unit_guid": "01749246-95f6-57db-b7c3-2ae0e8be6715","page": 1}`,
			expectedBody:       "{\n  \"ID\": \"123\",\n  \"Number\": 0,\n  \"MQTT\": \"\",\n  \"InventoryID\": \"\",\n  \"UnitGUID\": \"01749246-95f6-57db-b7c3-2ae0e8be6715\",\n  \"MessageID\": \"\",\n  \"MessageText\": \"\",\n  \"Context\": \"\",\n  \"MessageClass\": \"\",\n  \"Level\": 0,\n  \"Area\": \"\",\n  \"Address\": \"\",\n  \"Block\": false,\n  \"Type\": \"\",\n  \"Bit\": 0,\n  \"InvertBit\": 0\n}",
			expectedStatusCode: 202,
			mockBehavior: func(r *mocks.MockIUseCase) {
				r.EXPECT().GetEventByNumber(gomock.Any(), "01749246-95f6-57db-b7c3-2ae0e8be6715", 1).
					Return(events.Event{UnitGUID: "01749246-95f6-57db-b7c3-2ae0e8be6715", ID: "123"}, nil).AnyTimes()
			},
		},
		{
			name:               "Bad Request",
			body:               ``,
			expectedStatusCode: 400,
			mockBehavior: func(r *mocks.MockIUseCase) {
				r.EXPECT().GetEventByNumber(gomock.Any(), "", 0).
					Return(events.Event{}, nil).AnyTimes()
			},
		},
		{
			name:               "NotFound",
			body:               `{"unit_guid": "01749246-95f6-57db-b7c3-2ae0e8be6716","page": 1}`,
			expectedStatusCode: 404,
			mockBehavior: func(r *mocks.MockIUseCase) {
				r.EXPECT().GetEventByNumber(gomock.Any(), "01749246-95f6-57db-b7c3-2ae0e8be6716", 1).
					Return(events.Event{}, service.ErrEventNotFound).AnyTimes()
			},
		},
		{
			name:               "Storage Error",
			body:               `{"unit_guid": "01749246-95f6-57db-b7c3-2ae0e8be6716","page": 1}`,
			expectedStatusCode: 500,
			mockBehavior: func(r *mocks.MockIUseCase) {
				r.EXPECT().GetEventByNumber(gomock.Any(), "01749246-95f6-57db-b7c3-2ae0e8be6716", 1).
					Return(events.Event{}, usecase.ErrStorageIsUnavailable).AnyTimes()
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()
			logic := mocks.NewMockIUseCase(c)
			test.mockBehavior(logic)

			h := New(logic)

			r := httptest.NewRequest(http.MethodPost, url, strings.NewReader(test.body))
			w := httptest.NewRecorder()

			router := chi.NewRouter()
			router.Group(h.PublicRoutes)
			router.ServeHTTP(w, r)

			// Assert
			assert.Equal(t, test.expectedStatusCode, w.Code)
			if test.expectedBody != "" {
				assert.Equal(t, test.expectedBody, w.Body.String())
			}
		})
	}
}
