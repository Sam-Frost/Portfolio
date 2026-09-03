package reminder

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
	mux.HandleFunc("GET /api/todos/{todoId}/reminders", h.listByTodo)
	mux.HandleFunc("POST /api/todos/{todoId}/reminders", h.create)
	mux.HandleFunc("DELETE /api/reminders/{id}", h.delete)
}

func (h *Handler) listByTodo(w http.ResponseWriter, r *http.Request) {
	reminders, err := h.service.ListByTodo(r.Context(), r.PathValue("todoId"))
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, reminders)
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var input CreateInput
	if err := httpx.DecodeJSON(r, &input); err != nil {
		httpx.WriteError(w, err)
		return
	}
	rem, err := h.service.Create(r.Context(), r.PathValue("todoId"), input)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, rem)
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	if err := h.service.Delete(r.Context(), r.PathValue("id")); err != nil {
		httpx.WriteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
