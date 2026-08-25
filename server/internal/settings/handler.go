package settings

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
	mux.HandleFunc("GET /api/settings", h.get)
	mux.HandleFunc("PATCH /api/settings", h.update)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	s, err := h.service.Get(r.Context())
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, s)
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	var input UpdateInput
	if err := httpx.DecodeJSON(r, &input); err != nil {
		httpx.WriteError(w, err)
		return
	}

	s, err := h.service.Update(r.Context(), input)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, s)
}
