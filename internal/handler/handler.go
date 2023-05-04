package handler

import (
	"encoding/json"
	"errors"
	bettererror "github.com/egorgasay/bettererrors"
	"github.com/go-chi/httplog"
	"go-tsv-watcher/internal/schema"
	"go-tsv-watcher/internal/storage/service"
	"go-tsv-watcher/internal/usecase"
	"net/http"
)

// Handler struct for handler
type Handler struct {
	logic usecase.IUseCase
}

// New Handler constructor
func New(logic usecase.IUseCase) *Handler {
	return &Handler{logic: logic}
}

// BindJSON godoc
// @Summary Bind JSON
// @Description Bind JSON
// @Tags event
// @Accept  json
// @Produce  error
// @Param event body schema.EventRequest
func BindJSON(r *http.Request, obj any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(obj)
}

// PostEvent godoc
// @Summary Post event
// @Description Post event
// @Tags event
// @Accept  json
// @Produce  json
// @Param event body schema.EventRequest
// @Success 202 {object} schema.EventResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/event [post]
func (h Handler) PostEvent() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var eventRequest schema.EventRequest
		err := BindJSON(r, &eventRequest)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write(bettererror.New(err).SetAppLayer(bettererror.Handler).JSONPretty())
			return
		}
		defer r.Body.Close()

		event, err := h.logic.GetEventByNumber(r.Context(), eventRequest.UnitGUID, eventRequest.Page)
		if err != nil {
			if errors.Is(err, service.ErrEventNotFound) {
				w.WriteHeader(http.StatusNotFound)
				w.Write(bettererror.New(err).SetAppLayer(bettererror.Storage).JSONPretty())
				return
			}
			oplog := httplog.LogEntry(r.Context())
			oplog.Error().Msg(err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(bettererror.New(err).SetAppLayer(bettererror.Storage).JSONPretty())
			return
		}

		// marshal response
		response, err := json.MarshalIndent(event, "", "  ")
		if err != nil {
			oplog := httplog.LogEntry(r.Context())
			oplog.Error().Msg(err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(bettererror.New(err).SetAppLayer(bettererror.Handler).JSON())
			return
		}

		w.WriteHeader(http.StatusAccepted)
		w.Write(response)
	}
}
