package drawingboard

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
	mux.HandleFunc("GET /api/drawing-boards", h.list)
	mux.HandleFunc("POST /api/drawing-boards", h.create)
	mux.HandleFunc("GET /api/drawing-boards/{id}", h.get)
	mux.HandleFunc("PATCH /api/drawing-boards/{id}", h.update)
	mux.HandleFunc("DELETE /api/drawing-boards/{id}", h.delete)
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	boards, err := h.service.List(r.Context())
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, boards)
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var input CreateInput
	if err := httpx.DecodeJSON(r, &input); err != nil {
		httpx.WriteError(w, err)
		return
	}

	b, err := h.service.Create(r.Context(), input)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, b)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	b, err := h.service.Get(r.Context(), r.PathValue("id"))
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, b)
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	var input UpdateInput
	if err := httpx.DecodeJSON(r, &input); err != nil {
		httpx.WriteError(w, err)
		return
	}

	b, err := h.service.Update(r.Context(), r.PathValue("id"), input)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, b)
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	if err := h.service.Delete(r.Context(), r.PathValue("id")); err != nil {
		httpx.WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
