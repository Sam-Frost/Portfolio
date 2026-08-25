package upskill

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
	mux.HandleFunc("GET /api/upskill/topics", h.listTopics)
	mux.HandleFunc("POST /api/upskill/topics", h.createTopic)
	mux.HandleFunc("GET /api/upskill/topics/{id}", h.getTopic)
	mux.HandleFunc("PATCH /api/upskill/topics/{id}", h.updateTopic)
	mux.HandleFunc("DELETE /api/upskill/topics/{id}", h.deleteTopic)

	mux.HandleFunc("GET /api/upskill/topics/{id}/subtopics", h.listSubtopics)
	mux.HandleFunc("POST /api/upskill/topics/{id}/subtopics", h.createSubtopic)

	mux.HandleFunc("PATCH /api/upskill/subtopics/{id}", h.updateSubtopic)
	mux.HandleFunc("DELETE /api/upskill/subtopics/{id}", h.deleteSubtopic)
	mux.HandleFunc("POST /api/upskill/subtopics/{id}/resources", h.addResource)

	mux.HandleFunc("DELETE /api/upskill/resources/{id}", h.deleteResource)
}

func (h *Handler) listTopics(w http.ResponseWriter, r *http.Request) {
	topics, err := h.service.ListTopics(r.Context())
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, topics)
}

func (h *Handler) createTopic(w http.ResponseWriter, r *http.Request) {
	var input CreateTopicInput
	if err := httpx.DecodeJSON(r, &input); err != nil {
		httpx.WriteError(w, err)
		return
	}

	topic, err := h.service.CreateTopic(r.Context(), input)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, topic)
}

func (h *Handler) getTopic(w http.ResponseWriter, r *http.Request) {
	topic, err := h.service.GetTopic(r.Context(), r.PathValue("id"))
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, topic)
}

func (h *Handler) updateTopic(w http.ResponseWriter, r *http.Request) {
	var input UpdateTopicInput
	if err := httpx.DecodeJSON(r, &input); err != nil {
		httpx.WriteError(w, err)
		return
	}

	topic, err := h.service.UpdateTopic(r.Context(), r.PathValue("id"), input)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, topic)
}

func (h *Handler) deleteTopic(w http.ResponseWriter, r *http.Request) {
	if err := h.service.DeleteTopic(r.Context(), r.PathValue("id")); err != nil {
		httpx.WriteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) listSubtopics(w http.ResponseWriter, r *http.Request) {
	subtopics, err := h.service.ListSubtopics(r.Context(), r.PathValue("id"))
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, subtopics)
}

func (h *Handler) createSubtopic(w http.ResponseWriter, r *http.Request) {
	var input CreateSubtopicInput
	if err := httpx.DecodeJSON(r, &input); err != nil {
		httpx.WriteError(w, err)
		return
	}

	subtopic, err := h.service.CreateSubtopic(r.Context(), r.PathValue("id"), input)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, subtopic)
}

func (h *Handler) updateSubtopic(w http.ResponseWriter, r *http.Request) {
	var input UpdateSubtopicInput
	if err := httpx.DecodeJSON(r, &input); err != nil {
		httpx.WriteError(w, err)
		return
	}

	subtopic, err := h.service.UpdateSubtopic(r.Context(), r.PathValue("id"), input)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, subtopic)
}

func (h *Handler) deleteSubtopic(w http.ResponseWriter, r *http.Request) {
	if err := h.service.DeleteSubtopic(r.Context(), r.PathValue("id")); err != nil {
		httpx.WriteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) addResource(w http.ResponseWriter, r *http.Request) {
	var input CreateResourceInput
	if err := httpx.DecodeJSON(r, &input); err != nil {
		httpx.WriteError(w, err)
		return
	}

	resource, err := h.service.AddResource(r.Context(), r.PathValue("id"), input)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, resource)
}

func (h *Handler) deleteResource(w http.ResponseWriter, r *http.Request) {
	if err := h.service.DeleteResource(r.Context(), r.PathValue("id")); err != nil {
		httpx.WriteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
