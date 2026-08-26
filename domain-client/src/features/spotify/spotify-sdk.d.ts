// Minimal ambient types for Spotify's Web Playback SDK (loaded at runtime
// from https://sdk.scdn.co/spotify-player.js — see SpotifyPlayerProvider).
// Only the surface SpotifyPlayerProvider actually uses is declared; every
// playback action still goes through our backend (see api.ts), so the SDK
// itself is only used to register this tab as a device and to receive
// push-based state updates for it.

interface SpotifySdkPlayerOptions {
  name: string;
  getOAuthToken: (callback: (token: string) => void) => void;
  volume?: number;
}

interface SpotifySdkTrack {
  name: string;
  uri: string;
  duration_ms: number;
  artists: { name: string }[];
  album: { images: { url: string }[] };
}

interface SpotifySdkPlaybackState {
  paused: boolean;
  position: number;
  duration: number;
  track_window: { current_track: SpotifySdkTrack };
}

interface SpotifySdkPlayer {
  addListener(event: "ready" | "not_ready", callback: (data: { device_id: string }) => void): void;
  addListener(event: "player_state_changed", callback: (state: SpotifySdkPlaybackState | null) => void): void;
  connect(): Promise<boolean>;
  disconnect(): void;
}

interface Window {
  onSpotifyWebPlaybackSDKReady?: () => void;
  Spotify?: {
    Player: new (options: SpotifySdkPlayerOptions) => SpotifySdkPlayer;
  };
}
