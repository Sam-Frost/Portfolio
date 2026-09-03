package workprofile

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
	mux.HandleFunc("GET /api/work/tabs", h.listTabs)
	mux.HandleFunc("POST /api/work/tabs", h.createTab)
	mux.HandleFunc("PATCH /api/work/tabs/{id}", h.updateTab)
	mux.HandleFunc("DELETE /api/work/tabs/{id}", h.deleteTab)

	mux.HandleFunc("GET /api/work/tabs/{id}/tasks", h.listTasks)
	mux.HandleFunc("POST /api/work/tabs/{id}/tasks", h.createTask)
	mux.HandleFunc("PATCH /api/work/tasks/{id}", h.updateTask)
	mux.HandleFunc("DELETE /api/work/tasks/{id}", h.deleteTask)

	mux.HandleFunc("GET /api/work/overview", h.overview)
}

func (h *Handler) listTabs(w http.ResponseWriter, r *http.Request) {
	tabs, err := h.service.ListTabs(r.Context())
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, tabs)
}

func (h *Handler) createTab(w http.ResponseWriter, r *http.Request) {
	var input CreateTabInput
	if err := httpx.DecodeJSON(r, &input); err != nil {
		httpx.WriteError(w, err)
		return
	}
	tab, err := h.service.CreateTab(r.Context(), input)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, tab)
}

func (h *Handler) updateTab(w http.ResponseWriter, r *http.Request) {
	var input UpdateTabInput
	if err := httpx.DecodeJSON(r, &input); err != nil {
		httpx.WriteError(w, err)
		return
	}
	tab, err := h.service.UpdateTab(r.Context(), r.PathValue("id"), input)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, tab)
}

func (h *Handler) deleteTab(w http.ResponseWriter, r *http.Request) {
	if err := h.service.DeleteTab(r.Context(), r.PathValue("id")); err != nil {
		httpx.WriteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) listTasks(w http.ResponseWriter, r *http.Request) {
	tasks, err := h.service.ListTasks(r.Context(), r.PathValue("id"))
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, tasks)
}

func (h *Handler) createTask(w http.ResponseWriter, r *http.Request) {
	var input CreateTaskInput
	if err := httpx.DecodeJSON(r, &input); err != nil {
		httpx.WriteError(w, err)
		return
	}
	task, err := h.service.CreateTask(r.Context(), r.PathValue("id"), input)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, task)
}

func (h *Handler) updateTask(w http.ResponseWriter, r *http.Request) {
	var input UpdateTaskInput
	if err := httpx.DecodeJSON(r, &input); err != nil {
		httpx.WriteError(w, err)
		return
	}
	task, err := h.service.UpdateTask(r.Context(), r.PathValue("id"), input)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, task)
}

func (h *Handler) deleteTask(w http.ResponseWriter, r *http.Request) {
	if err := h.service.DeleteTask(r.Context(), r.PathValue("id")); err != nil {
		httpx.WriteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) overview(w http.ResponseWriter, r *http.Request) {
	ov, err := h.service.Overview(r.Context())
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, ov)
}
