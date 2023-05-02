package handler

import (
	"encoding/json"
	bettererror "github.com/egorgasay/bettererrors"
	"github.com/go-chi/httplog"
	"go-tsv-watcher/internal/schema"
	"go-tsv-watcher/internal/usecase"
	"net/http"
)

type Handler struct {
	logic *usecase.UseCase
}

func New(logic *usecase.UseCase) *Handler {
	return &Handler{logic: logic}
}

func BindJSON(r *http.Request, obj any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(obj)
}

func (h Handler) PostEvent() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var eventRequest schema.EventRequest
		err := BindJSON(r, &eventRequest)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(bettererror.New(err).SetAppLayer(bettererror.Handler).JSONPretty())
			return
		}
		defer r.Body.Close()

		event, err := h.logic.GetEventByNumber(r.Context(), eventRequest.UnitGUID, eventRequest.Page)
		if err != nil {
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
