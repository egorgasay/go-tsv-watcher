package handler

import (
	"github.com/go-chi/chi/v5"
)

// PublicRoutes - Routes for public endpoints
func (h Handler) PublicRoutes(r chi.Router) {
	r.Post("/api/v1/event", h.PostEvent())
}
