package spotify

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Sam-Frost/portfolio/internal/apperr"
)

// fakeClient is a spotifyClient test double — every method is backed by an
// optional function field so each test only wires up what it exercises.
type fakeClient struct {
	exchangeCodeFn   func(ctx context.Context, code string) (TokenSet, error)
	refreshTokenFn   func(ctx context.Context, refreshToken string) (TokenSet, error)
	playlistTracksFn func(ctx context.Context, accessToken, playlistID string) ([]Track, error)
	playFn           func(ctx context.Context, accessToken string, input playInput) error

	refreshCalls int
	playedInput  playInput
}

func (f *fakeClient) authorizeURL(state string) string {
	return "https://accounts.spotify.com/authorize?state=" + state
}

func (f *fakeClient) exchangeCode(ctx context.Context, code string) (TokenSet, error) {
	return f.exchangeCodeFn(ctx, code)
}

func (f *fakeClient) refreshToken(ctx context.Context, refreshToken string) (TokenSet, error) {
	f.refreshCalls++
	return f.refreshTokenFn(ctx, refreshToken)
}

func (f *fakeClient) playbackState(context.Context, string) (*PlaybackState, error) { return nil, nil }
func (f *fakeClient) devices(context.Context, string) ([]Device, error)             { return nil, nil }
func (f *fakeClient) transferPlayback(context.Context, string, string) error        { return nil }

func (f *fakeClient) play(ctx context.Context, accessToken string, input playInput) error {
	f.playedInput = input
	if f.playFn != nil {
		return f.playFn(ctx, accessToken, input)
	}
	return nil
}

func (f *fakeClient) pause(context.Context, string) error                     { return nil }
func (f *fakeClient) skipNext(context.Context, string) error                  { return nil }
func (f *fakeClient) skipPrevious(context.Context, string) error              { return nil }
func (f *fakeClient) seek(context.Context, string, int) error                 { return nil }
func (f *fakeClient) setVolume(context.Context, string, int) error            { return nil }
func (f *fakeClient) search(context.Context, string, string) ([]Track, error) { return nil, nil }
func (f *fakeClient) playlists(context.Context, string) ([]Playlist, error)   { return nil, nil }

func (f *fakeClient) playlistTracks(ctx context.Context, accessToken, playlistID string) ([]Track, error) {
	return f.playlistTracksFn(ctx, accessToken, playlistID)
}

func assertInvalidInput(t *testing.T, err error) {
	t.Helper()
	var appErr *apperr.Error
	if !errors.As(err, &appErr) || appErr.Kind != apperr.KindInvalidInput {
		t.Fatalf("err = %v, want apperr.InvalidInput", err)
	}
}

func connectedRepo(accessToken string, expiresAt time.Time) Repository {
	repo := NewMemoryRepository()
	_ = repo.Save(context.Background(), TokenSet{
		RefreshToken:         "refresh-1",
		AccessToken:          accessToken,
		AccessTokenExpiresAt: expiresAt,
	})
	return repo
}

func TestService_ValidAccessToken_UsesCachedTokenWhenNotExpired(t *testing.T) {
	repo := connectedRepo("cached-access", time.Now().Add(time.Hour))
	client := &fakeClient{}
	svc := NewService(repo, client, []byte("secret"))

	token, err := svc.validAccessToken(context.Background())
	if err != nil {
		t.Fatalf("validAccessToken: %v", err)
	}
	if token != "cached-access" {
		t.Errorf("token = %q, want %q", token, "cached-access")
	}
	if client.refreshCalls != 0 {
		t.Errorf("refreshCalls = %d, want 0", client.refreshCalls)
	}
}

func TestService_ValidAccessToken_RefreshesWhenExpired(t *testing.T) {
	repo := connectedRepo("stale-access", time.Now().Add(-time.Minute))
	client := &fakeClient{
		refreshTokenFn: func(_ context.Context, refreshToken string) (TokenSet, error) {
			if refreshToken != "refresh-1" {
				t.Errorf("refreshToken = %q, want %q", refreshToken, "refresh-1")
			}
			return TokenSet{RefreshToken: "refresh-1", AccessToken: "fresh-access", AccessTokenExpiresAt: time.Now().Add(time.Hour)}, nil
		},
	}
	svc := NewService(repo, client, []byte("secret"))

	token, err := svc.validAccessToken(context.Background())
	if err != nil {
		t.Fatalf("validAccessToken: %v", err)
	}
	if token != "fresh-access" {
		t.Errorf("token = %q, want %q", token, "fresh-access")
	}
	if client.refreshCalls != 1 {
		t.Errorf("refreshCalls = %d, want 1", client.refreshCalls)
	}

	stored, ok, _ := repo.Get(context.Background())
	if !ok || stored.AccessToken != "fresh-access" {
		t.Errorf("refreshed token wasn't persisted: %+v", stored)
	}
}

func TestService_ValidAccessToken_RefreshesWithinSkewWindow(t *testing.T) {
	// Expires in 30s — inside refreshSkew (60s) — should refresh proactively
	// rather than hand back a token that may be rejected mid-request.
	repo := connectedRepo("about-to-expire", time.Now().Add(30*time.Second))
	client := &fakeClient{
		refreshTokenFn: func(context.Context, string) (TokenSet, error) {
			return TokenSet{RefreshToken: "refresh-1", AccessToken: "fresh-access", AccessTokenExpiresAt: time.Now().Add(time.Hour)}, nil
		},
	}
	svc := NewService(repo, client, []byte("secret"))

	token, err := svc.validAccessToken(context.Background())
	if err != nil {
		t.Fatalf("validAccessToken: %v", err)
	}
	if token != "fresh-access" {
		t.Errorf("token = %q, want %q", token, "fresh-access")
	}
}

func TestService_ValidAccessToken_NotConnectedReturnsInvalidInput(t *testing.T) {
	svc := NewService(NewMemoryRepository(), &fakeClient{}, []byte("secret"))

	_, err := svc.validAccessToken(context.Background())
	assertInvalidInput(t, err)
}

func TestService_SeekRejectsNegativePosition(t *testing.T) {
	svc := NewService(NewMemoryRepository(), &fakeClient{}, []byte("secret"))
	assertInvalidInput(t, svc.Seek(context.Background(), -1))
}

func TestService_SetVolumeRejectsOutOfRange(t *testing.T) {
	svc := NewService(NewMemoryRepository(), &fakeClient{}, []byte("secret"))
	assertInvalidInput(t, svc.SetVolume(context.Background(), 101))
	assertInvalidInput(t, svc.SetVolume(context.Background(), -1))
}

func TestService_SearchRejectsBlankQuery(t *testing.T) {
	svc := NewService(NewMemoryRepository(), &fakeClient{}, []byte("secret"))
	_, err := svc.Search(context.Background(), "   ")
	assertInvalidInput(t, err)
}

func TestService_PlayPlaylist_LikedSongsUsesExplicitTrackURIs(t *testing.T) {
	repo := connectedRepo("access", time.Now().Add(time.Hour))
	client := &fakeClient{
		playlistTracksFn: func(_ context.Context, _, playlistID string) ([]Track, error) {
			if playlistID != likedSongsID {
				t.Errorf("playlistID = %q, want %q", playlistID, likedSongsID)
			}
			return []Track{{URI: "spotify:track:1"}, {URI: "spotify:track:2"}}, nil
		},
	}
	svc := NewService(repo, client, []byte("secret"))

	if err := svc.PlayPlaylist(context.Background(), likedSongsID, "spotify:track:2", "device-1"); err != nil {
		t.Fatalf("PlayPlaylist: %v", err)
	}
	if len(client.playedInput.URIs) != 2 || client.playedInput.ContextURI != "" {
		t.Errorf("playedInput = %+v, want two explicit URIs and no context", client.playedInput)
	}
	if client.playedInput.OffsetURI != "spotify:track:2" {
		t.Errorf("playedInput.OffsetURI = %q, want %q", client.playedInput.OffsetURI, "spotify:track:2")
	}
	if client.playedInput.DeviceID != "device-1" {
		t.Errorf("playedInput.DeviceID = %q, want %q", client.playedInput.DeviceID, "device-1")
	}
}

func TestService_PlayPlaylist_LikedSongsEmptyIsInvalidInput(t *testing.T) {
	repo := connectedRepo("access", time.Now().Add(time.Hour))
	client := &fakeClient{
		playlistTracksFn: func(context.Context, string, string) ([]Track, error) { return nil, nil },
	}
	svc := NewService(repo, client, []byte("secret"))

	err := svc.PlayPlaylist(context.Background(), likedSongsID, "", "")
	assertInvalidInput(t, err)
}

func TestService_PlayPlaylist_RegularPlaylistUsesContextURI(t *testing.T) {
	repo := connectedRepo("access", time.Now().Add(time.Hour))
	client := &fakeClient{}
	svc := NewService(repo, client, []byte("secret"))

	if err := svc.PlayPlaylist(context.Background(), "abc123", "spotify:track:9", ""); err != nil {
		t.Fatalf("PlayPlaylist: %v", err)
	}
	if client.playedInput.ContextURI != "spotify:playlist:abc123" || len(client.playedInput.URIs) != 0 {
		t.Errorf("playedInput = %+v, want playlist context and no explicit URIs", client.playedInput)
	}
	if client.playedInput.OffsetURI != "spotify:track:9" {
		t.Errorf("playedInput.OffsetURI = %q, want %q", client.playedInput.OffsetURI, "spotify:track:9")
	}
}

func TestService_StateRoundTrip(t *testing.T) {
	svc := NewService(NewMemoryRepository(), &fakeClient{}, []byte("secret"))

	state, err := svc.signState()
	if err != nil {
		t.Fatalf("signState: %v", err)
	}
	if !svc.verifyState(state) {
		t.Error("verifyState rejected a state it just signed")
	}
}

func TestService_VerifyStateRejectsTamperedSignature(t *testing.T) {
	svc := NewService(NewMemoryRepository(), &fakeClient{}, []byte("secret"))

	state, _ := svc.signState()
	// Flip the last hex digit of the signature to something guaranteed
	// different, rather than a fixed "0" that could coincidentally match.
	last := state[len(state)-1]
	replacement := byte('0')
	if last == '0' {
		replacement = '1'
	}
	tampered := state[:len(state)-1] + string(replacement)

	if svc.verifyState(tampered) {
		t.Error("verifyState accepted a tampered state")
	}
}

func TestService_VerifyStateRejectsDifferentSecret(t *testing.T) {
	signer := NewService(NewMemoryRepository(), &fakeClient{}, []byte("secret-a"))
	verifier := NewService(NewMemoryRepository(), &fakeClient{}, []byte("secret-b"))

	state, _ := signer.signState()
	if verifier.verifyState(state) {
		t.Error("verifyState accepted a state signed with a different secret")
	}
}

func TestService_VerifyStateRejectsMalformedState(t *testing.T) {
	svc := NewService(NewMemoryRepository(), &fakeClient{}, []byte("secret"))
	if svc.verifyState("not-a-real-state") {
		t.Error("verifyState accepted a malformed state")
	}
}

func TestService_HandleCallback_RejectsInvalidState(t *testing.T) {
	svc := NewService(NewMemoryRepository(), &fakeClient{}, []byte("secret"))

	err := svc.HandleCallback(context.Background(), "code", "not-a-real-state")
	var appErr *apperr.Error
	if !errors.As(err, &appErr) || appErr.Kind != apperr.KindUnauthorized {
		t.Fatalf("err = %v, want apperr.Unauthorized", err)
	}
}

func TestService_HandleCallback_SavesTokensForValidState(t *testing.T) {
	repo := NewMemoryRepository()
	client := &fakeClient{
		exchangeCodeFn: func(_ context.Context, code string) (TokenSet, error) {
			if code != "auth-code" {
				t.Errorf("code = %q, want %q", code, "auth-code")
			}
			return TokenSet{RefreshToken: "r", AccessToken: "a", AccessTokenExpiresAt: time.Now().Add(time.Hour)}, nil
		},
	}
	svc := NewService(repo, client, []byte("secret"))

	state, err := svc.signState()
	if err != nil {
		t.Fatalf("signState: %v", err)
	}

	if err := svc.HandleCallback(context.Background(), "auth-code", state); err != nil {
		t.Fatalf("HandleCallback: %v", err)
	}

	_, ok, _ := repo.Get(context.Background())
	if !ok {
		t.Error("HandleCallback didn't persist tokens")
	}
}
