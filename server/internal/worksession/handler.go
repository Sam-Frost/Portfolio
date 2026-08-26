package worksession

import (
	"net/http"
	"time"

	"github.com/Sam-Frost/portfolio/internal/apperr"
	"github.com/Sam-Frost/portfolio/internal/httpx"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/work-sessions", h.start)
	mux.HandleFunc("GET /api/work-sessions/current", h.current)
	mux.HandleFunc("GET /api/work-sessions/daily-summary", h.dailySummary)
	mux.HandleFunc("GET /api/work-sessions", h.list)
	mux.HandleFunc("POST /api/work-sessions/{id}/complete", h.complete)
	mux.HandleFunc("POST /api/work-sessions/{id}/cancel", h.cancel)
}

func (h *Handler) start(w http.ResponseWriter, r *http.Request) {
	var input StartInput
	if err := httpx.DecodeJSON(r, &input); err != nil {
		httpx.WriteError(w, err)
		return
	}

	session, err := h.service.Start(r.Context(), input)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, session)
}

// current returns 204 (rather than a JSON null) when nothing is running,
// matching how the rest of the API signals "nothing here" (see
// internal/spotify's /state).
func (h *Handler) current(w http.ResponseWriter, r *http.Request) {
	session, ok, err := h.service.Current(r.Context())
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	if !ok {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, session)
}

func (h *Handler) complete(w http.ResponseWriter, r *http.Request) {
	var input CompleteInput
	if err := httpx.DecodeJSON(r, &input); err != nil {
		httpx.WriteError(w, err)
		return
	}

	session, err := h.service.Complete(r.Context(), r.PathValue("id"), input)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, session)
}

func (h *Handler) cancel(w http.ResponseWriter, r *http.Request) {
	var input CancelInput
	if err := httpx.DecodeJSON(r, &input); err != nil {
		httpx.WriteError(w, err)
		return
	}

	session, err := h.service.Cancel(r.Context(), r.PathValue("id"), input)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, session)
}

// parseRange reads the required "from"/"to" YYYY-MM-DD (IST) query params
// shared by list and dailySummary, returning the [from, to) UTC instant
// bounds — to is exclusive, i.e. the start of the IST day *after* the
// "to" param, so that date is included in full.
func parseRange(r *http.Request) (from, to time.Time, err error) {
	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")
	if fromStr == "" || toStr == "" {
		return time.Time{}, time.Time{}, apperr.InvalidInput("from and to query params are required")
	}

	from, err = ParseISTDayStart(fromStr)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	toDayStart, err := ParseISTDayStart(toStr)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	return from, toDayStart.AddDate(0, 0, 1), nil
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	from, to, err := parseRange(r)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	sessions, err := h.service.ListRange(r.Context(), from, to)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, sessions)
}

func (h *Handler) dailySummary(w http.ResponseWriter, r *http.Request) {
	from, to, err := parseRange(r)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	summary, err := h.service.DailySummary(r.Context(), from, to)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, summary)
}
