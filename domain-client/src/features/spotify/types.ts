export interface Track {
  id: string;
  name: string;
  artists: string;
  albumArt: string;
  uri: string;
  durationMs: number;
}

export interface SpotifyDevice {
  id: string;
  name: string;
  type: string;
  isActive: boolean;
  volumePercent: number;
}

export interface PlaybackState {
  isPlaying: boolean;
  progressMs: number;
  track: Track | null;
  device: SpotifyDevice | null;
}

export interface Playlist {
  id: string;
  name: string;
  imageUrl: string;
  trackCount: number;
  uri: string;
}
