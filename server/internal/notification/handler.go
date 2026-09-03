package notification

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
	mux.HandleFunc("GET /api/notifications/vapid-public-key", h.vapidKey)
	mux.HandleFunc("POST /api/notifications/subscriptions", h.subscribe)
	mux.HandleFunc("POST /api/notifications/subscriptions/sync", h.resync)
	mux.HandleFunc("DELETE /api/notifications/subscriptions", h.unsubscribe)
	mux.HandleFunc("POST /api/notifications/test", h.test)
}

func (h *Handler) vapidKey(w http.ResponseWriter, _ *http.Request) {
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"key": h.service.VAPIDPublicKey()})
}

func (h *Handler) subscribe(w http.ResponseWriter, r *http.Request) {
	var input SubscribeInput
	if err := httpx.DecodeJSON(r, &input); err != nil {
		httpx.WriteError(w, err)
		return
	}
	if err := h.service.Subscribe(r.Context(), input); err != nil {
		httpx.WriteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// resync is unauthenticated (see publicPaths in cmd/main.go): the service
// worker's pushsubscriptionchange handler has no bearer token. It can only
// move/refresh a subscription — knowing the (secret, high-entropy) push
// endpoint URL is the capability check.
func (h *Handler) resync(w http.ResponseWriter, r *http.Request) {
	var input struct {
		OldEndpoint string `json:"oldEndpoint"`
		SubscribeInput
	}
	if err := httpx.DecodeJSON(r, &input); err != nil {
		httpx.WriteError(w, err)
		return
	}
	if err := h.service.Resync(r.Context(), input.OldEndpoint, input.SubscribeInput); err != nil {
		httpx.WriteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) unsubscribe(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Endpoint string `json:"endpoint"`
	}
	if err := httpx.DecodeJSON(r, &input); err != nil {
		httpx.WriteError(w, err)
		return
	}
	if err := h.service.Unsubscribe(r.Context(), input.Endpoint); err != nil {
		httpx.WriteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) test(w http.ResponseWriter, r *http.Request) {
	if err := h.service.SendTest(r.Context()); err != nil {
		httpx.WriteError(w, err)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}
