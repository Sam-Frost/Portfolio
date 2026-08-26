package diary

import (
	"net/http"

	"github.com/Sam-Frost/portfolio/internal/httpx"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/diary/entries", h.listDates)
	mux.HandleFunc("GET /api/diary/entries/{date}", h.get)
	mux.HandleFunc("PUT /api/diary/entries/{date}", h.upsert)
}

// listDates powers the calendar view: which dates in [from, to] have an
// entry, without pulling every day's content over the wire.
func (h *Handler) listDates(w http.ResponseWriter, r *http.Request) {
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")

	dates, err := h.service.ListDates(r.Context(), from, to)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, dates)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	e, err := h.service.GetByDate(r.Context(), r.PathValue("date"))
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, e)
}

// upsert creates or updates the entry for a date ("one entry per day",
// edited over time) — rejected with a 409 once that date's edit window has
// closed, per Service.Upsert.
func (h *Handler) upsert(w http.ResponseWriter, r *http.Request) {
	var input UpsertInput
	if err := httpx.DecodeJSON(r, &input); err != nil {
		httpx.WriteError(w, err)
		return
	}

	e, err := h.service.Upsert(r.Context(), r.PathValue("date"), input.Content)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, e)
}
