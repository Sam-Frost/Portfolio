package notepad

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
	mux.HandleFunc("GET /api/notes", h.list)
	mux.HandleFunc("POST /api/notes", h.create)
	// More specific than GET /api/notes/{id}, so ServeMux routes it here.
	mux.HandleFunc("GET /api/notes/scratch", h.scratch)
	mux.HandleFunc("GET /api/notes/{id}", h.get)
	mux.HandleFunc("PATCH /api/notes/{id}", h.update)
	mux.HandleFunc("DELETE /api/notes/{id}", h.delete)
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	filter := ListFilter{Archived: r.URL.Query().Get("archived") == "true"}

	notes, err := h.service.List(r.Context(), filter)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, notes)
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var input CreateInput
	if err := httpx.DecodeJSON(r, &input); err != nil {
		httpx.WriteError(w, err)
		return
	}

	n, err := h.service.Create(r.Context(), input)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, n)
}

func (h *Handler) scratch(w http.ResponseWriter, r *http.Request) {
	n, err := h.service.Scratch(r.Context())
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, n)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	n, err := h.service.Get(r.Context(), r.PathValue("id"))
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, n)
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	var input UpdateInput
	if err := httpx.DecodeJSON(r, &input); err != nil {
		httpx.WriteError(w, err)
		return
	}

	n, err := h.service.Update(r.Context(), r.PathValue("id"), input)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, n)
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	if err := h.service.Delete(r.Context(), r.PathValue("id")); err != nil {
		httpx.WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
