package spotify

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Sam-Frost/portfolio/internal/apperr"
)

const (
	accountsBaseURL = "https://accounts.spotify.com"
	apiBaseURL      = "https://api.spotify.com/v1"

	// likedSongsID is a synthetic playlist ID: Spotify's Web API exposes
	// saved tracks through /me/tracks rather than as a real playlist, so
	// it's surfaced here as if it were one and special-cased below.
	likedSongsID = "liked-songs"
)

var scopes = []string{
	"streaming",
	"user-read-email",
	"user-read-private",
	"user-read-playback-state",
	"user-modify-playback-state",
	"user-read-currently-playing",
	"playlist-read-private",
	// Required for GET /me/tracks, which the Liked Songs entry in the
	// Playlists tab is built from (see likedSongsTotal/playlistTracks).
	"user-library-read",
}

// playInput selects what PUT /me/player/play should start: either a
// context (playlist/album) or an explicit list of track URIs. At most one
// of ContextURI/URIs should be set; neither set means "resume whatever was
// last playing". OffsetURI, when set, starts playback at that track within
// the context/URIs instead of the first — so the rest of the playlist (or
// list) stays queued and keeps playing after it, rather than stopping once
// a single track ends. DeviceID targets a specific device directly —
// Spotify otherwise requires some device to already be marked *active*
// (playing natively, or explicitly transferred to) before /play will
// accept anything, which a freshly-registered Web Playback SDK device
// isn't yet.
type playInput struct {
	ContextURI string
	URIs       []string
	OffsetURI  string
	DeviceID   string
}

// spotifyClient is the subset of Spotify's accounts + Web API the service
// needs. apiClient implements it against the real network; service_test.go
// substitutes a fake so token-refresh/caching and playlist logic are
// verifiable without hitting Spotify.
type spotifyClient interface {
	authorizeURL(state string) string
	exchangeCode(ctx context.Context, code string) (TokenSet, error)
	refreshToken(ctx context.Context, refreshToken string) (TokenSet, error)
	playbackState(ctx context.Context, accessToken string) (*PlaybackState, error)
	devices(ctx context.Context, accessToken string) ([]Device, error)
	transferPlayback(ctx context.Context, accessToken, deviceID string) error
	play(ctx context.Context, accessToken string, input playInput) error
	pause(ctx context.Context, accessToken string) error
	skipNext(ctx context.Context, accessToken string) error
	skipPrevious(ctx context.Context, accessToken string) error
	seek(ctx context.Context, accessToken string, positionMs int) error
	setVolume(ctx context.Context, accessToken string, percent int) error
	search(ctx context.Context, accessToken, query string) ([]Track, error)
	playlists(ctx context.Context, accessToken string) ([]Playlist, error)
	playlistTracks(ctx context.Context, accessToken, playlistID string) ([]Track, error)
}

// apiClient is the real spotifyClient, talking to accounts.spotify.com for
// OAuth and api.spotify.com for playback/search/playlists.
type apiClient struct {
	clientID     string
	clientSecret string
	redirectURI  string
	httpClient   *http.Client
}

func NewAPIClient(clientID, clientSecret, redirectURI string) *apiClient {
	return &apiClient{
		clientID:     clientID,
		clientSecret: clientSecret,
		redirectURI:  redirectURI,
		httpClient:   &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *apiClient) authorizeURL(state string) string {
	q := url.Values{
		"client_id":     {c.clientID},
		"response_type": {"code"},
		"redirect_uri":  {c.redirectURI},
		"scope":         {strings.Join(scopes, " ")},
		"state":         {state},
	}
	return accountsBaseURL + "/authorize?" + q.Encode()
}

func (c *apiClient) basicAuthHeader() string {
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(c.clientID+":"+c.clientSecret))
}

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

func (c *apiClient) postToken(ctx context.Context, form url.Values) (TokenSet, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, accountsBaseURL+"/api/token", strings.NewReader(form.Encode()))
	if err != nil {
		return TokenSet{}, apperr.Internal("failed to build spotify token request")
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", c.basicAuthHeader())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return TokenSet{}, apperr.Internal("failed to reach spotify")
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return TokenSet{}, apperr.Internal(fmt.Sprintf("spotify token request failed: %s", string(body)))
	}

	var parsed tokenResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return TokenSet{}, apperr.Internal("failed to parse spotify token response")
	}

	return TokenSet{
		AccessToken:          parsed.AccessToken,
		RefreshToken:         parsed.RefreshToken,
		AccessTokenExpiresAt: time.Now().Add(time.Duration(parsed.ExpiresIn) * time.Second),
	}, nil
}

func (c *apiClient) exchangeCode(ctx context.Context, code string) (TokenSet, error) {
	form := url.Values{
		"grant_type":   {"authorization_code"},
		"code":         {code},
		"redirect_uri": {c.redirectURI},
	}
	return c.postToken(ctx, form)
}

// refreshToken exchanges a refresh token for a new access token. Spotify
// doesn't always rotate the refresh token in the response, so the caller's
// existing one is preserved when the response omits it.
func (c *apiClient) refreshToken(ctx context.Context, refreshToken string) (TokenSet, error) {
	form := url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {refreshToken},
	}
	tokens, err := c.postToken(ctx, form)
	if err != nil {
		return TokenSet{}, err
	}
	if tokens.RefreshToken == "" {
		tokens.RefreshToken = refreshToken
	}
	return tokens, nil
}

func (c *apiClient) apiRequest(ctx context.Context, method, path, accessToken string, body any) (*http.Response, error) {
	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, apperr.Internal("failed to encode spotify request")
		}
		reader = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, apiBaseURL+path, reader)
	if err != nil {
		return nil, apperr.Internal("failed to build spotify request")
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, apperr.Internal("failed to reach spotify")
	}
	return resp, nil
}

// okOrErr turns a non-2xx Web API response into an *apperr.Error. A 404 on
// a playback-control call almost always means no device is currently
// active, which is a user-actionable state (pick a device), not a server
// failure — everything else collapses to Internal with Spotify's own body
// for debugging.
func okOrErr(resp *http.Response, msg string) error {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}
	if resp.StatusCode == http.StatusNotFound {
		return apperr.InvalidInput("no active Spotify device — open the player or pick a device first")
	}
	body, _ := io.ReadAll(resp.Body)
	return apperr.Internal(fmt.Sprintf("%s: %s", msg, string(body)))
}

type rawImage struct {
	URL string `json:"url"`
}

type rawArtist struct {
	Name string `json:"name"`
}

type rawAlbum struct {
	Images []rawImage `json:"images"`
}

type rawTrack struct {
	ID         string      `json:"id"`
	Name       string      `json:"name"`
	URI        string      `json:"uri"`
	DurationMs int         `json:"duration_ms"`
	Artists    []rawArtist `json:"artists"`
	Album      rawAlbum    `json:"album"`
}

func (rt rawTrack) toTrack() Track {
	names := make([]string, 0, len(rt.Artists))
	for _, a := range rt.Artists {
		names = append(names, a.Name)
	}
	art := ""
	if len(rt.Album.Images) > 0 {
		art = rt.Album.Images[0].URL
	}
	return Track{
		ID:         rt.ID,
		Name:       rt.Name,
		Artists:    strings.Join(names, ", "),
		AlbumArt:   art,
		URI:        rt.URI,
		DurationMs: rt.DurationMs,
	}
}

type rawDevice struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Type          string `json:"type"`
	IsActive      bool   `json:"is_active"`
	VolumePercent int    `json:"volume_percent"`
}

func (rd rawDevice) toDevice() Device {
	return Device{ID: rd.ID, Name: rd.Name, Type: rd.Type, IsActive: rd.IsActive, VolumePercent: rd.VolumePercent}
}

func (c *apiClient) playbackState(ctx context.Context, accessToken string) (*PlaybackState, error) {
	resp, err := c.apiRequest(ctx, http.MethodGet, "/me/player", accessToken, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, apperr.Internal("failed to fetch spotify playback state")
	}

	var raw struct {
		Device     *rawDevice `json:"device"`
		ProgressMs int        `json:"progress_ms"`
		IsPlaying  bool       `json:"is_playing"`
		Item       *rawTrack  `json:"item"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, apperr.Internal("failed to parse spotify playback state")
	}

	state := &PlaybackState{IsPlaying: raw.IsPlaying, ProgressMs: raw.ProgressMs}
	if raw.Item != nil {
		t := raw.Item.toTrack()
		state.Track = &t
	}
	if raw.Device != nil {
		d := raw.Device.toDevice()
		state.Device = &d
	}
	return state, nil
}

func (c *apiClient) devices(ctx context.Context, accessToken string) ([]Device, error) {
	resp, err := c.apiRequest(ctx, http.MethodGet, "/me/player/devices", accessToken, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, apperr.Internal("failed to fetch spotify devices")
	}

	var raw struct {
		Devices []rawDevice `json:"devices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, apperr.Internal("failed to parse spotify devices")
	}

	out := make([]Device, 0, len(raw.Devices))
	for _, d := range raw.Devices {
		out = append(out, d.toDevice())
	}
	return out, nil
}

func (c *apiClient) transferPlayback(ctx context.Context, accessToken, deviceID string) error {
	body := map[string]any{"device_ids": []string{deviceID}, "play": true}
	resp, err := c.apiRequest(ctx, http.MethodPut, "/me/player", accessToken, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return okOrErr(resp, "failed to transfer spotify playback")
}

func (c *apiClient) play(ctx context.Context, accessToken string, input playInput) error {
	var body map[string]any
	switch {
	case len(input.URIs) > 0:
		body = map[string]any{"uris": input.URIs}
	case input.ContextURI != "":
		body = map[string]any{"context_uri": input.ContextURI}
	}
	if body != nil && input.OffsetURI != "" {
		body["offset"] = map[string]any{"uri": input.OffsetURI}
	}

	path := "/me/player/play"
	if input.DeviceID != "" {
		path += "?device_id=" + url.QueryEscape(input.DeviceID)
	}

	// Pass a genuinely nil body (not a nil map boxed in an interface) for the
	// resume case, so apiRequest sends no request body at all.
	var reqBody any
	if body != nil {
		reqBody = body
	}
	resp, err := c.apiRequest(ctx, http.MethodPut, path, accessToken, reqBody)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return okOrErr(resp, "failed to start spotify playback")
}

func (c *apiClient) pause(ctx context.Context, accessToken string) error {
	resp, err := c.apiRequest(ctx, http.MethodPut, "/me/player/pause", accessToken, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return okOrErr(resp, "failed to pause spotify playback")
}

func (c *apiClient) skipNext(ctx context.Context, accessToken string) error {
	resp, err := c.apiRequest(ctx, http.MethodPost, "/me/player/next", accessToken, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return okOrErr(resp, "failed to skip to next track")
}

func (c *apiClient) skipPrevious(ctx context.Context, accessToken string) error {
	resp, err := c.apiRequest(ctx, http.MethodPost, "/me/player/previous", accessToken, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return okOrErr(resp, "failed to skip to previous track")
}

func (c *apiClient) seek(ctx context.Context, accessToken string, positionMs int) error {
	path := "/me/player/seek?position_ms=" + strconv.Itoa(positionMs)
	resp, err := c.apiRequest(ctx, http.MethodPut, path, accessToken, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return okOrErr(resp, "failed to seek spotify playback")
}

func (c *apiClient) setVolume(ctx context.Context, accessToken string, percent int) error {
	path := "/me/player/volume?volume_percent=" + strconv.Itoa(percent)
	resp, err := c.apiRequest(ctx, http.MethodPut, path, accessToken, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return okOrErr(resp, "failed to set spotify volume")
}

func (c *apiClient) search(ctx context.Context, accessToken, query string) ([]Track, error) {
	path := "/search?type=track&limit=10&q=" + url.QueryEscape(query)
	resp, err := c.apiRequest(ctx, http.MethodGet, path, accessToken, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, apperr.Internal("failed to search spotify")
	}

	var raw struct {
		Tracks struct {
			Items []rawTrack `json:"items"`
		} `json:"tracks"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, apperr.Internal("failed to parse spotify search results")
	}

	out := make([]Track, 0, len(raw.Tracks.Items))
	for _, t := range raw.Tracks.Items {
		out = append(out, t.toTrack())
	}
	return out, nil
}

type rawPlaylist struct {
	ID     string     `json:"id"`
	Name   string     `json:"name"`
	Images []rawImage `json:"images"`
	URI    string     `json:"uri"`
	// Spotify nests the track count under "items", not "tracks" as older
	// docs/examples suggest — confirmed against a live response.
	Items struct {
		Total int `json:"total"`
	} `json:"items"`
}

func (rp rawPlaylist) toPlaylist() Playlist {
	img := ""
	if len(rp.Images) > 0 {
		img = rp.Images[0].URL
	}
	return Playlist{ID: rp.ID, Name: rp.Name, ImageURL: img, TrackCount: rp.Items.Total, URI: rp.URI}
}

func (c *apiClient) likedSongsTotal(ctx context.Context, accessToken string) (int, error) {
	resp, err := c.apiRequest(ctx, http.MethodGet, "/me/tracks?limit=1", accessToken, nil)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return 0, apperr.Internal(fmt.Sprintf("failed to fetch liked songs total: %s", string(body)))
	}

	var raw struct {
		Total int `json:"total"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return 0, apperr.Internal("failed to parse liked songs total")
	}
	return raw.Total, nil
}

// playlists prepends the synthetic "Liked Songs" entry ahead of the user's
// real playlists, matching how Spotify's own clients present it.
func (c *apiClient) playlists(ctx context.Context, accessToken string) ([]Playlist, error) {
	likedTotal, err := c.likedSongsTotal(ctx, accessToken)
	if err != nil {
		return nil, err
	}

	resp, err := c.apiRequest(ctx, http.MethodGet, "/me/playlists?limit=50", accessToken, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, apperr.Internal(fmt.Sprintf("failed to fetch spotify playlists: %s", string(body)))
	}

	var raw struct {
		Items []rawPlaylist `json:"items"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, apperr.Internal("failed to parse spotify playlists")
	}

	out := make([]Playlist, 0, len(raw.Items)+1)
	out = append(out, Playlist{ID: likedSongsID, Name: "Liked Songs", TrackCount: likedTotal})
	for _, p := range raw.Items {
		out = append(out, p.toPlaylist())
	}
	return out, nil
}

func (c *apiClient) playlistTracks(ctx context.Context, accessToken, playlistID string) ([]Track, error) {
	path := "/playlists/" + playlistID + "/tracks?limit=50"
	if playlistID == likedSongsID {
		path = "/me/tracks?limit=50"
	}

	resp, err := c.apiRequest(ctx, http.MethodGet, path, accessToken, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, apperr.Internal("failed to fetch playlist tracks")
	}

	var raw struct {
		Items []struct {
			Track rawTrack `json:"track"`
		} `json:"items"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, apperr.Internal("failed to parse playlist tracks")
	}

	out := make([]Track, 0, len(raw.Items))
	for _, item := range raw.Items {
		out = append(out, item.Track.toTrack())
	}
	return out, nil
}
