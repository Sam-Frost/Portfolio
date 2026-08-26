package spotify

import "time"

// TokenSet is the OAuth token state for the domain area's one Spotify
// connection (see Repository) — AccessToken/AccessTokenExpiresAt are a
// cache to avoid refreshing on every request; RefreshToken is the
// long-lived credential everything else is rebuilt from.
type TokenSet struct {
	RefreshToken         string
	AccessToken          string
	AccessTokenExpiresAt time.Time
}

type Track struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Artists    string `json:"artists"`
	AlbumArt   string `json:"albumArt"`
	URI        string `json:"uri"`
	DurationMs int    `json:"durationMs"`
}

type Device struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Type          string `json:"type"`
	IsActive      bool   `json:"isActive"`
	VolumePercent int    `json:"volumePercent"`
}

type PlaybackState struct {
	IsPlaying  bool    `json:"isPlaying"`
	ProgressMs int     `json:"progressMs"`
	Track      *Track  `json:"track"`
	Device     *Device `json:"device"`
}

// Playlist represents both real Spotify playlists and the synthetic
// "Liked Songs" entry (ID == likedSongsID) — Spotify exposes saved tracks
// through a separate endpoint rather than as a real playlist, so it's
// synthesized in client.go to appear alongside the rest.
type Playlist struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	ImageURL   string `json:"imageUrl"`
	TrackCount int    `json:"trackCount"`
	URI        string `json:"uri"`
}
