package cms

import (
	"net/http"
	"strconv"

	"github.com/Sam-Frost/portfolio/internal/httpx"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/cms/content", h.content)
	mux.HandleFunc("GET /api/cms/status", h.status)
	mux.HandleFunc("POST /api/cms/publish", h.publish)
	mux.HandleFunc("GET /api/cms/publications", h.publications)

	mux.HandleFunc("GET /api/cms/projects", h.listProjects)
	mux.HandleFunc("POST /api/cms/projects", h.createProject)
	mux.HandleFunc("PATCH /api/cms/projects/{id}", h.updateProject)
	mux.HandleFunc("DELETE /api/cms/projects/{id}", h.deleteProject)

	mux.HandleFunc("GET /api/cms/experiences", h.listExperiences)
	mux.HandleFunc("POST /api/cms/experiences", h.createExperience)
	mux.HandleFunc("PATCH /api/cms/experiences/{id}", h.updateExperience)
	mux.HandleFunc("DELETE /api/cms/experiences/{id}", h.deleteExperience)

	mux.HandleFunc("GET /api/cms/blogs", h.listBlogs)
	mux.HandleFunc("POST /api/cms/blogs", h.createBlog)
	mux.HandleFunc("PATCH /api/cms/blogs/{id}", h.updateBlog)
	mux.HandleFunc("DELETE /api/cms/blogs/{id}", h.deleteBlog)

	mux.HandleFunc("GET /api/cms/summary", h.getSummary)
	mux.HandleFunc("PATCH /api/cms/summary", h.updateSummary)
}

// ─────────────────────────────────────────────
// content / status / publish
// ─────────────────────────────────────────────

func (h *Handler) content(w http.ResponseWriter, r *http.Request) {
	c, err := h.service.Content(r.Context())
	writeResult(w, http.StatusOK, c, err)
}

func (h *Handler) status(w http.ResponseWriter, r *http.Request) {
	cs, err := h.service.Status(r.Context())
	writeResult(w, http.StatusOK, cs, err)
}

func (h *Handler) publish(w http.ResponseWriter, r *http.Request) {
	p, err := h.service.Publish(r.Context())
	writeResult(w, http.StatusOK, p, err)
}

func (h *Handler) publications(w http.ResponseWriter, r *http.Request) {
	limit := 20
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			limit = n
		}
	}
	list, err := h.service.Publications(r.Context(), limit)
	writeResult(w, http.StatusOK, list, err)
}

// ─────────────────────────────────────────────
// projects
// ─────────────────────────────────────────────

func (h *Handler) listProjects(w http.ResponseWriter, r *http.Request) {
	list, err := h.service.ListProjects(r.Context())
	writeResult(w, http.StatusOK, list, err)
}

func (h *Handler) createProject(w http.ResponseWriter, r *http.Request) {
	var in CreateProjectInput
	if err := httpx.DecodeJSON(r, &in); err != nil {
		httpx.WriteError(w, err)
		return
	}
	p, err := h.service.CreateProject(r.Context(), in)
	writeResult(w, http.StatusCreated, p, err)
}

func (h *Handler) updateProject(w http.ResponseWriter, r *http.Request) {
	var in UpdateProjectInput
	if err := httpx.DecodeJSON(r, &in); err != nil {
		httpx.WriteError(w, err)
		return
	}
	p, err := h.service.UpdateProject(r.Context(), r.PathValue("id"), in)
	writeResult(w, http.StatusOK, p, err)
}

func (h *Handler) deleteProject(w http.ResponseWriter, r *http.Request) {
	writeDelete(w, h.service.DeleteProject(r.Context(), r.PathValue("id")))
}

// ─────────────────────────────────────────────
// experiences
// ─────────────────────────────────────────────

func (h *Handler) listExperiences(w http.ResponseWriter, r *http.Request) {
	list, err := h.service.ListExperiences(r.Context())
	writeResult(w, http.StatusOK, list, err)
}

func (h *Handler) createExperience(w http.ResponseWriter, r *http.Request) {
	var in CreateExperienceInput
	if err := httpx.DecodeJSON(r, &in); err != nil {
		httpx.WriteError(w, err)
		return
	}
	e, err := h.service.CreateExperience(r.Context(), in)
	writeResult(w, http.StatusCreated, e, err)
}

func (h *Handler) updateExperience(w http.ResponseWriter, r *http.Request) {
	var in UpdateExperienceInput
	if err := httpx.DecodeJSON(r, &in); err != nil {
		httpx.WriteError(w, err)
		return
	}
	e, err := h.service.UpdateExperience(r.Context(), r.PathValue("id"), in)
	writeResult(w, http.StatusOK, e, err)
}

func (h *Handler) deleteExperience(w http.ResponseWriter, r *http.Request) {
	writeDelete(w, h.service.DeleteExperience(r.Context(), r.PathValue("id")))
}

// ─────────────────────────────────────────────
// blogs
// ─────────────────────────────────────────────

func (h *Handler) listBlogs(w http.ResponseWriter, r *http.Request) {
	list, err := h.service.ListBlogs(r.Context())
	writeResult(w, http.StatusOK, list, err)
}

func (h *Handler) createBlog(w http.ResponseWriter, r *http.Request) {
	var in CreateBlogInput
	if err := httpx.DecodeJSON(r, &in); err != nil {
		httpx.WriteError(w, err)
		return
	}
	b, err := h.service.CreateBlog(r.Context(), in)
	writeResult(w, http.StatusCreated, b, err)
}

func (h *Handler) updateBlog(w http.ResponseWriter, r *http.Request) {
	var in UpdateBlogInput
	if err := httpx.DecodeJSON(r, &in); err != nil {
		httpx.WriteError(w, err)
		return
	}
	b, err := h.service.UpdateBlog(r.Context(), r.PathValue("id"), in)
	writeResult(w, http.StatusOK, b, err)
}

func (h *Handler) deleteBlog(w http.ResponseWriter, r *http.Request) {
	writeDelete(w, h.service.DeleteBlog(r.Context(), r.PathValue("id")))
}

// ─────────────────────────────────────────────
// summary
// ─────────────────────────────────────────────

func (h *Handler) getSummary(w http.ResponseWriter, r *http.Request) {
	s, err := h.service.GetSummary(r.Context())
	writeResult(w, http.StatusOK, s, err)
}

func (h *Handler) updateSummary(w http.ResponseWriter, r *http.Request) {
	var in UpdateSummaryInput
	if err := httpx.DecodeJSON(r, &in); err != nil {
		httpx.WriteError(w, err)
		return
	}
	s, err := h.service.UpdateSummary(r.Context(), in)
	writeResult(w, http.StatusOK, s, err)
}

// ─────────────────────────────────────────────
// shared response helpers
// ─────────────────────────────────────────────

func writeResult(w http.ResponseWriter, status int, body any, err error) {
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, status, body)
}

func writeDelete(w http.ResponseWriter, err error) {
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
