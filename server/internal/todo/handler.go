package todo

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
	mux.HandleFunc("GET /api/todos", h.list)
	mux.HandleFunc("POST /api/todos", h.create)
	mux.HandleFunc("GET /api/todos/count", h.countActive)
	mux.HandleFunc("PATCH /api/todos/{id}", h.update)
	mux.HandleFunc("DELETE /api/todos/{id}", h.delete)
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	sortField := SortField(r.URL.Query().Get("sortBy"))
	if sortField != SortByTargetDate && sortField != SortByCompletedAt {
		sortField = SortByDateAdded
	}

	order := SortOrder(r.URL.Query().Get("order"))
	if order != SortAsc {
		order = SortDesc
	}

	var labelID *string
	if v := r.URL.Query().Get("labelId"); v != "" {
		labelID = &v
	}

	todos, err := h.service.List(r.Context(), sortField, order, labelID)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, todos)
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var input CreateInput
	if err := httpx.DecodeJSON(r, &input); err != nil {
		httpx.WriteError(w, err)
		return
	}

	todo, err := h.service.Create(r.Context(), input)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, todo)
}

func (h *Handler) countActive(w http.ResponseWriter, r *http.Request) {
	count, err := h.service.CountActive(r.Context())
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]int{"active": count})
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	var input UpdateInput
	if err := httpx.DecodeJSON(r, &input); err != nil {
		httpx.WriteError(w, err)
		return
	}

	todo, err := h.service.Update(r.Context(), r.PathValue("id"), input)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, todo)
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	if err := h.service.Delete(r.Context(), r.PathValue("id")); err != nil {
		httpx.WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
