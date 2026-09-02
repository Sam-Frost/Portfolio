package diary

import (
	"net/http"

	"github.com/Sam-Frost/portfolio/internal/httpx"
)

type VideoHandler struct {
	service *VideoService
}

func NewVideoHandler(service *VideoService) *VideoHandler {
	return &VideoHandler{service: service}
}

func (h *VideoHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/diary/videos", h.counts)
	mux.HandleFunc("GET /api/diary/videos/{date}", h.listByDate)
	mux.HandleFunc("POST /api/diary/videos/{date}", h.createUpload)
	mux.HandleFunc("POST /api/diary/videos/{id}/parts", h.partURL)
	mux.HandleFunc("POST /api/diary/videos/{id}/complete", h.complete)
	mux.HandleFunc("GET /api/diary/videos/{id}/play", h.play)
	mux.HandleFunc("DELETE /api/diary/videos/{id}", h.delete)
}

// counts powers the calendar: how many ready videos each date in [from, to]
// has, so a day with only a video (no written entry) is still marked.
func (h *VideoHandler) counts(w http.ResponseWriter, r *http.Request) {
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")

	counts, err := h.service.CountsByDateRange(r.Context(), from, to)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, counts)
}

func (h *VideoHandler) listByDate(w http.ResponseWriter, r *http.Request) {
	videos, err := h.service.ListByDate(r.Context(), r.PathValue("date"))
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, videos)
}

func (h *VideoHandler) createUpload(w http.ResponseWriter, r *http.Request) {
	var input CreateVideoInput
	if err := httpx.DecodeJSON(r, &input); err != nil {
		httpx.WriteError(w, err)
		return
	}
	created, err := h.service.CreateUpload(r.Context(), r.PathValue("date"), input)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, created)
}

func (h *VideoHandler) partURL(w http.ResponseWriter, r *http.Request) {
	var input struct {
		PartNumber int `json:"partNumber"`
	}
	if err := httpx.DecodeJSON(r, &input); err != nil {
		httpx.WriteError(w, err)
		return
	}
	url, err := h.service.PartURL(r.Context(), r.PathValue("id"), input.PartNumber)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"url": url})
}

func (h *VideoHandler) complete(w http.ResponseWriter, r *http.Request) {
	var input CompleteVideoInput
	if err := httpx.DecodeJSON(r, &input); err != nil {
		httpx.WriteError(w, err)
		return
	}
	video, err := h.service.Complete(r.Context(), r.PathValue("id"), input)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, video)
}

func (h *VideoHandler) play(w http.ResponseWriter, r *http.Request) {
	url, err := h.service.PlaybackURL(r.Context(), r.PathValue("id"))
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"url": url})
}

func (h *VideoHandler) delete(w http.ResponseWriter, r *http.Request) {
	if err := h.service.Delete(r.Context(), r.PathValue("id")); err != nil {
		httpx.WriteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
