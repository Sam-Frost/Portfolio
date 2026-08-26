package spotify

import (
	"net/http"

	"github.com/Sam-Frost/portfolio/internal/httpx"
)

type Handler struct {
	service     *Service
	frontendURL string
}

// frontendURL is where the browser lands after the OAuth round-trip
// finishes (success or failure) — see SPOTIFY_FRONTEND_REDIRECT_URL in
// .env.example. It's an HTTP concern (where to send the browser), not
// business logic, so it lives on the handler rather than the service.
func NewHandler(service *Service, frontendURL string) *Handler {
	return &Handler{service: service, frontendURL: frontendURL}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/spotify/auth-url", h.authURL)
	mux.HandleFunc("GET /api/spotify/callback", h.callback)
	mux.HandleFunc("GET /api/spotify/status", h.status)
	mux.HandleFunc("GET /api/spotify/sdk-token", h.sdkToken)
	mux.HandleFunc("GET /api/spotify/state", h.state)
	mux.HandleFunc("GET /api/spotify/devices", h.devices)
	mux.HandleFunc("POST /api/spotify/transfer", h.transfer)
	mux.HandleFunc("POST /api/spotify/play", h.play)
	mux.HandleFunc("POST /api/spotify/pause", h.pause)
	mux.HandleFunc("POST /api/spotify/next", h.next)
	mux.HandleFunc("POST /api/spotify/previous", h.previous)
	mux.HandleFunc("POST /api/spotify/seek", h.seek)
	mux.HandleFunc("POST /api/spotify/volume", h.volume)
	mux.HandleFunc("GET /api/spotify/search", h.search)
	mux.HandleFunc("GET /api/spotify/playlists", h.playlists)
	mux.HandleFunc("GET /api/spotify/playlists/{id}/tracks", h.playlistTracks)
	mux.HandleFunc("POST /api/spotify/playlists/{id}/play", h.playPlaylist)
}

func (h *Handler) authURL(w http.ResponseWriter, r *http.Request) {
	url, err := h.service.AuthURL(r.Context())
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"url": url})
}

// callback is hit by Spotify's own redirect, not our frontend, so it
// carries no bearer token — that's why /api/spotify/callback is in
// cmd/main.go's publicPaths. state is HMAC-signed by Service.AuthURL and
// verified inside HandleCallback instead.
func (h *Handler) callback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")

	if code == "" || r.URL.Query().Get("error") != "" {
		http.Redirect(w, r, h.frontendURL+"?spotify=error", http.StatusFound)
		return
	}

	if err := h.service.HandleCallback(r.Context(), code, state); err != nil {
		http.Redirect(w, r, h.frontendURL+"?spotify=error", http.StatusFound)
		return
	}

	http.Redirect(w, r, h.frontendURL+"?spotify=connected", http.StatusFound)
}

func (h *Handler) status(w http.ResponseWriter, r *http.Request) {
	connected, err := h.service.Status(r.Context())
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]bool{"connected": connected})
}

func (h *Handler) sdkToken(w http.ResponseWriter, r *http.Request) {
	token, expiresAt, err := h.service.SDKToken(r.Context())
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"accessToken": token,
		"expiresAt":   expiresAt,
	})
}

// state returns 204 (rather than a JSON null) when nothing is playing
// anywhere, matching how the rest of the API signals "nothing here".
func (h *Handler) state(w http.ResponseWriter, r *http.Request) {
	state, err := h.service.State(r.Context())
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	if state == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, state)
}

func (h *Handler) devices(w http.ResponseWriter, r *http.Request) {
	devices, err := h.service.Devices(r.Context())
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, devices)
}

type transferRequest struct {
	DeviceID string `json:"deviceId"`
}

func (h *Handler) transfer(w http.ResponseWriter, r *http.Request) {
	var input transferRequest
	if err := httpx.DecodeJSON(r, &input); err != nil {
		httpx.WriteError(w, err)
		return
	}
	if err := h.service.Transfer(r.Context(), input.DeviceID); err != nil {
		httpx.WriteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type playRequest struct {
	URI      string `json:"uri"`
	DeviceID string `json:"deviceId"`
}

// play plays a specific track (uri set — e.g. a search result) or resumes
// whatever was last playing (uri omitted — the collapsed widget's
// play/pause button once state is already Paused). deviceId is optional —
// the frontend sends its own Web Playback SDK device id whenever no device
// is currently active, so playback has somewhere to start.
func (h *Handler) play(w http.ResponseWriter, r *http.Request) {
	var input playRequest
	if err := httpx.DecodeJSON(r, &input); err != nil {
		httpx.WriteError(w, err)
		return
	}

	var err error
	if input.URI != "" {
		err = h.service.PlayTrack(r.Context(), input.URI, input.DeviceID)
	} else {
		err = h.service.Resume(r.Context(), input.DeviceID)
	}
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) pause(w http.ResponseWriter, r *http.Request) {
	if err := h.service.Pause(r.Context()); err != nil {
		httpx.WriteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) next(w http.ResponseWriter, r *http.Request) {
	if err := h.service.Next(r.Context()); err != nil {
		httpx.WriteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) previous(w http.ResponseWriter, r *http.Request) {
	if err := h.service.Previous(r.Context()); err != nil {
		httpx.WriteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type seekRequest struct {
	PositionMs int `json:"positionMs"`
}

func (h *Handler) seek(w http.ResponseWriter, r *http.Request) {
	var input seekRequest
	if err := httpx.DecodeJSON(r, &input); err != nil {
		httpx.WriteError(w, err)
		return
	}
	if err := h.service.Seek(r.Context(), input.PositionMs); err != nil {
		httpx.WriteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type volumeRequest struct {
	Percent int `json:"percent"`
}

func (h *Handler) volume(w http.ResponseWriter, r *http.Request) {
	var input volumeRequest
	if err := httpx.DecodeJSON(r, &input); err != nil {
		httpx.WriteError(w, err)
		return
	}
	if err := h.service.SetVolume(r.Context(), input.Percent); err != nil {
		httpx.WriteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) search(w http.ResponseWriter, r *http.Request) {
	tracks, err := h.service.Search(r.Context(), r.URL.Query().Get("q"))
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, tracks)
}

func (h *Handler) playlists(w http.ResponseWriter, r *http.Request) {
	playlists, err := h.service.Playlists(r.Context())
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, playlists)
}

func (h *Handler) playlistTracks(w http.ResponseWriter, r *http.Request) {
	tracks, err := h.service.PlaylistTracks(r.Context(), r.PathValue("id"))
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, tracks)
}

// playPlaylist accepts the same optional deviceId body as play, for the
// same reason (nothing active yet to target).
func (h *Handler) playPlaylist(w http.ResponseWriter, r *http.Request) {
	var input playRequest
	if err := httpx.DecodeJSON(r, &input); err != nil {
		httpx.WriteError(w, err)
		return
	}

	if err := h.service.PlayPlaylist(r.Context(), r.PathValue("id"), input.DeviceID); err != nil {
		httpx.WriteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
