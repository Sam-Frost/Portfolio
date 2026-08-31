package spotify

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"strings"
	"time"

	"github.com/Sam-Frost/portfolio/internal/apperr"
)

// refreshSkew is how far ahead of AccessTokenExpiresAt validTokens treats
// the cached access token as already-expired, so a request never races a
// token that's valid when checked but rejected by Spotify a moment later.
const refreshSkew = 60 * time.Second

// stateTTL bounds how long a signed OAuth state (see signState) is
// accepted, so a stale/replayed callback URL can't be reused indefinitely.
const stateTTL = 10 * time.Minute

type Service struct {
	repo        Repository
	client      spotifyClient
	stateSecret []byte
}

func NewService(repo Repository, client spotifyClient, stateSecret []byte) *Service {
	return &Service{repo: repo, client: client, stateSecret: stateSecret}
}

// AuthURL returns the Spotify authorize URL to redirect the browser to,
// carrying an HMAC-signed state so the public callback (see HandleCallback)
// can verify the request originated here rather than being forged.
func (s *Service) AuthURL(_ context.Context) (string, error) {
	state, err := s.signState()
	if err != nil {
		return "", apperr.Internal("failed to build spotify auth url")
	}
	return s.client.authorizeURL(state), nil
}

// HandleCallback completes the OAuth round-trip: verifies state, exchanges
// the code Spotify issued for tokens, and persists them as the domain
// area's one Spotify connection.
func (s *Service) HandleCallback(ctx context.Context, code, state string) error {
	if !s.verifyState(state) {
		return apperr.Unauthorized("invalid or expired spotify auth request")
	}

	tokens, err := s.client.exchangeCode(ctx, code)
	if err != nil {
		return err
	}
	return s.repo.Save(ctx, tokens)
}

func (s *Service) Status(ctx context.Context) (bool, error) {
	_, ok, err := s.repo.Get(ctx)
	return ok, err
}

// validTokens returns a TokenSet with a currently-valid access token,
// refreshing (and persisting the result) first if the cached one is
// missing or within refreshSkew of expiring.
func (s *Service) validTokens(ctx context.Context) (TokenSet, error) {
	tokens, ok, err := s.repo.Get(ctx)
	if err != nil {
		return TokenSet{}, err
	}
	if !ok {
		return TokenSet{}, apperr.InvalidInput("spotify isn't connected yet")
	}

	if tokens.AccessToken != "" && time.Now().Before(tokens.AccessTokenExpiresAt.Add(-refreshSkew)) {
		return tokens, nil
	}

	refreshed, err := s.client.refreshToken(ctx, tokens.RefreshToken)
	if err != nil {
		return TokenSet{}, err
	}
	if err := s.repo.Save(ctx, refreshed); err != nil {
		return TokenSet{}, err
	}
	return refreshed, nil
}

func (s *Service) validAccessToken(ctx context.Context) (string, error) {
	tokens, err := s.validTokens(ctx)
	if err != nil {
		return "", err
	}
	return tokens.AccessToken, nil
}

// SDKToken is the one place a raw access token leaves the backend — the
// Web Playback SDK running in the browser needs it directly to authenticate
// its own streaming connection. Every other frontend action goes through a
// backend endpoint instead of touching the token.
func (s *Service) SDKToken(ctx context.Context) (string, time.Time, error) {
	tokens, err := s.validTokens(ctx)
	if err != nil {
		return "", time.Time{}, err
	}
	return tokens.AccessToken, tokens.AccessTokenExpiresAt, nil
}

func (s *Service) State(ctx context.Context) (*PlaybackState, error) {
	token, err := s.validAccessToken(ctx)
	if err != nil {
		return nil, err
	}
	return s.client.playbackState(ctx, token)
}

func (s *Service) Devices(ctx context.Context) ([]Device, error) {
	token, err := s.validAccessToken(ctx)
	if err != nil {
		return nil, err
	}
	return s.client.devices(ctx, token)
}

func (s *Service) Transfer(ctx context.Context, deviceID string) error {
	if strings.TrimSpace(deviceID) == "" {
		return apperr.InvalidInput("deviceId is required")
	}
	token, err := s.validAccessToken(ctx)
	if err != nil {
		return err
	}
	return s.client.transferPlayback(ctx, token, deviceID)
}

// PlayTrack plays a single track immediately (used by search results),
// replacing whatever context was previously playing. deviceID is optional —
// pass it when no device is currently active (see PlaybackState.Device) so
// playback has somewhere to start instead of Spotify rejecting the request;
// leave it empty to target whatever's already active.
func (s *Service) PlayTrack(ctx context.Context, uri, deviceID string) error {
	if strings.TrimSpace(uri) == "" {
		return apperr.InvalidInput("uri is required")
	}
	token, err := s.validAccessToken(ctx)
	if err != nil {
		return err
	}
	return s.client.play(ctx, token, playInput{URIs: []string{uri}, DeviceID: deviceID})
}

// PlayTracks plays an explicit list of tracks as a queue (used by search
// results — the clicked track plus the rest of the results), starting at
// offsetURI when set, so playback continues to the next result instead of
// stopping once one track ends. See PlayTrack for deviceID.
func (s *Service) PlayTracks(ctx context.Context, uris []string, offsetURI, deviceID string) error {
	if len(uris) == 0 {
		return apperr.InvalidInput("uris is required")
	}
	token, err := s.validAccessToken(ctx)
	if err != nil {
		return err
	}
	return s.client.play(ctx, token, playInput{URIs: uris, OffsetURI: offsetURI, DeviceID: deviceID})
}

// Resume resumes whatever was last playing, without changing what's queued.
// See PlayTrack for deviceID.
func (s *Service) Resume(ctx context.Context, deviceID string) error {
	token, err := s.validAccessToken(ctx)
	if err != nil {
		return err
	}
	return s.client.play(ctx, token, playInput{DeviceID: deviceID})
}

func (s *Service) Pause(ctx context.Context) error {
	token, err := s.validAccessToken(ctx)
	if err != nil {
		return err
	}
	return s.client.pause(ctx, token)
}

func (s *Service) Next(ctx context.Context) error {
	token, err := s.validAccessToken(ctx)
	if err != nil {
		return err
	}
	return s.client.skipNext(ctx, token)
}

func (s *Service) Previous(ctx context.Context) error {
	token, err := s.validAccessToken(ctx)
	if err != nil {
		return err
	}
	return s.client.skipPrevious(ctx, token)
}

func (s *Service) Seek(ctx context.Context, positionMs int) error {
	if positionMs < 0 {
		return apperr.InvalidInput("positionMs must be >= 0")
	}
	token, err := s.validAccessToken(ctx)
	if err != nil {
		return err
	}
	return s.client.seek(ctx, token, positionMs)
}

func (s *Service) SetVolume(ctx context.Context, percent int) error {
	if percent < 0 || percent > 100 {
		return apperr.InvalidInput("percent must be between 0 and 100")
	}
	token, err := s.validAccessToken(ctx)
	if err != nil {
		return err
	}
	return s.client.setVolume(ctx, token, percent)
}

func (s *Service) Search(ctx context.Context, query string) ([]Track, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, apperr.InvalidInput("q is required")
	}
	token, err := s.validAccessToken(ctx)
	if err != nil {
		return nil, err
	}
	return s.client.search(ctx, token, query)
}

func (s *Service) Playlists(ctx context.Context) ([]Playlist, error) {
	token, err := s.validAccessToken(ctx)
	if err != nil {
		return nil, err
	}
	return s.client.playlists(ctx, token)
}

func (s *Service) PlaylistTracks(ctx context.Context, playlistID string) ([]Track, error) {
	if strings.TrimSpace(playlistID) == "" {
		return nil, apperr.InvalidInput("playlist id is required")
	}
	token, err := s.validAccessToken(ctx)
	if err != nil {
		return nil, err
	}
	return s.client.playlistTracks(ctx, token, playlistID)
}

// PlayPlaylist plays a playlist, starting at offsetURI when set (a track
// clicked in the playlist view) or from the top otherwise. Either way the
// whole playlist stays queued, so it keeps playing after that track ends.
// Liked Songs has no context URI in Spotify's API, so it's resolved to an
// explicit list of track URIs instead of a context — everything else plays
// by context_uri. See PlayTrack for deviceID.
func (s *Service) PlayPlaylist(ctx context.Context, playlistID, offsetURI, deviceID string) error {
	if strings.TrimSpace(playlistID) == "" {
		return apperr.InvalidInput("playlist id is required")
	}
	token, err := s.validAccessToken(ctx)
	if err != nil {
		return err
	}

	if playlistID == likedSongsID {
		tracks, err := s.client.playlistTracks(ctx, token, likedSongsID)
		if err != nil {
			return err
		}
		if len(tracks) == 0 {
			return apperr.InvalidInput("liked songs is empty")
		}
		uris := make([]string, 0, len(tracks))
		for _, t := range tracks {
			uris = append(uris, t.URI)
		}
		return s.client.play(ctx, token, playInput{URIs: uris, OffsetURI: offsetURI, DeviceID: deviceID})
	}

	return s.client.play(ctx, token, playInput{
		ContextURI: "spotify:playlist:" + playlistID,
		OffsetURI:  offsetURI,
		DeviceID:   deviceID,
	})
}

// signState/verifyState implement the OAuth "state" parameter as an
// HMAC-signed, timestamped opaque token rather than a server-side session:
// there's no session store elsewhere in this codebase, and one signed
// string is enough to make the public callback endpoint (see
// cmd/main.go's publicPaths) resistant to being hit directly/forged.
func (s *Service) signState() (string, error) {
	nonce := make([]byte, 16)
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}

	payload := strconv.FormatInt(time.Now().Unix(), 10) + "." + hex.EncodeToString(nonce)
	mac := hmac.New(sha256.New, s.stateSecret)
	mac.Write([]byte(payload))
	return payload + "." + hex.EncodeToString(mac.Sum(nil)), nil
}

func (s *Service) verifyState(state string) bool {
	parts := strings.SplitN(state, ".", 3)
	if len(parts) != 3 {
		return false
	}

	payload := parts[0] + "." + parts[1]
	mac := hmac.New(sha256.New, s.stateSecret)
	mac.Write([]byte(payload))
	expected := hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(expected), []byte(parts[2])) {
		return false
	}

	ts, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return false
	}
	return time.Since(time.Unix(ts, 0)) < stateTTL
}
