package document

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

// blobServer is implemented by a BlobStore that needs to serve the raw
// upload/download bytes itself (the local-disk store). The S3 store hands
// out URLs pointing straight at the bucket and implements nothing here.
type blobServer interface {
	ServeBlob(http.ResponseWriter, *http.Request)
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/documents/folders", h.listFolders)
	mux.HandleFunc("POST /api/documents/folders", h.createFolder)
	mux.HandleFunc("PATCH /api/documents/folders/{id}", h.updateFolder)
	mux.HandleFunc("DELETE /api/documents/folders/{id}", h.deleteFolder)

	mux.HandleFunc("GET /api/documents", h.listDocuments)
	mux.HandleFunc("POST /api/documents", h.createDocument)
	mux.HandleFunc("POST /api/documents/{id}/complete", h.completeDocument)
	mux.HandleFunc("GET /api/documents/{id}/download", h.downloadDocument)
	mux.HandleFunc("PATCH /api/documents/{id}", h.updateDocument)
	mux.HandleFunc("DELETE /api/documents/{id}", h.deleteDocument)

	if bs, ok := h.service.blob.(blobServer); ok {
		mux.HandleFunc(blobRoutePrefix, bs.ServeBlob)
	}
}

// --- folders ---

func (h *Handler) listFolders(w http.ResponseWriter, r *http.Request) {
	folders, err := h.service.ListFolders(r.Context())
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, folders)
}

func (h *Handler) createFolder(w http.ResponseWriter, r *http.Request) {
	var input CreateFolderInput
	if err := httpx.DecodeJSON(r, &input); err != nil {
		httpx.WriteError(w, err)
		return
	}
	folder, err := h.service.CreateFolder(r.Context(), input)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, folder)
}

func (h *Handler) updateFolder(w http.ResponseWriter, r *http.Request) {
	var input UpdateFolderInput
	if err := httpx.DecodeJSON(r, &input); err != nil {
		httpx.WriteError(w, err)
		return
	}
	folder, err := h.service.UpdateFolder(r.Context(), r.PathValue("id"), input)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, folder)
}

func (h *Handler) deleteFolder(w http.ResponseWriter, r *http.Request) {
	if err := h.service.DeleteFolder(r.Context(), r.PathValue("id")); err != nil {
		httpx.WriteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- documents ---

func (h *Handler) listDocuments(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	filter := ListFilter{Query: q.Get("q")}
	if v := q.Get("folderId"); v != "" {
		filter.FolderID = &v
	}
	if v := q.Get("labelId"); v != "" {
		filter.LabelID = &v
	}

	docs, err := h.service.ListDocuments(r.Context(), filter)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, docs)
}

func (h *Handler) createDocument(w http.ResponseWriter, r *http.Request) {
	var input CreateDocumentInput
	if err := httpx.DecodeJSON(r, &input); err != nil {
		httpx.WriteError(w, err)
		return
	}
	created, err := h.service.CreateDocument(r.Context(), input)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, created)
}

func (h *Handler) completeDocument(w http.ResponseWriter, r *http.Request) {
	doc, err := h.service.CompleteUpload(r.Context(), r.PathValue("id"))
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, doc)
}

func (h *Handler) downloadDocument(w http.ResponseWriter, r *http.Request) {
	url, err := h.service.DownloadURL(r.Context(), r.PathValue("id"))
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"url": url})
}

func (h *Handler) updateDocument(w http.ResponseWriter, r *http.Request) {
	var input UpdateDocumentInput
	if err := httpx.DecodeJSON(r, &input); err != nil {
		httpx.WriteError(w, err)
		return
	}
	doc, err := h.service.UpdateDocument(r.Context(), r.PathValue("id"), input)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, doc)
}

func (h *Handler) deleteDocument(w http.ResponseWriter, r *http.Request) {
	if err := h.service.DeleteDocument(r.Context(), r.PathValue("id")); err != nil {
		httpx.WriteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
