package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
}

func New() *Handler {
	return &Handler{}
}

func (h *Handler) Router() http.Handler {
	r := chi.NewRouter()

	r.Get("/health", h.Health)

	return r
}
