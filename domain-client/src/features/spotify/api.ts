import { apiRequest } from "../../lib/apiClient";
import type { PlaybackState, Playlist, SpotifyDevice, Track } from "./types";

export function fetchStatus(): Promise<{ connected: boolean }> {
  return apiRequest("/api/spotify/status");
}

// Not a fetch — the caller does `window.location.href = await connectUrl()`
// for a full-page handoff to Spotify's consent screen.
export async function connectUrl(): Promise<string> {
  const { url } = await apiRequest<{ url: string }>("/api/spotify/auth-url");
  return url;
}

export function fetchSdkToken(): Promise<{ accessToken: string; expiresAt: string }> {
  return apiRequest("/api/spotify/sdk-token");
}

// undefined (not just null) on 204 — see apiRequest — when nothing is
// playing anywhere.
export function fetchState(): Promise<PlaybackState | null> {
  return apiRequest("/api/spotify/state");
}

export function fetchDevices(): Promise<SpotifyDevice[]> {
  return apiRequest("/api/spotify/devices");
}

export function transferDevice(deviceId: string): Promise<void> {
  return apiRequest("/api/spotify/transfer", { method: "POST", body: JSON.stringify({ deviceId }) });
}

// deviceId targets a specific device directly — pass it (typically the SDK's
// own device id for this tab) when no device is currently active, since
// Spotify otherwise rejects /play with no device to start on.
export function playTrack(uri: string, deviceId?: string): Promise<void> {
  return apiRequest("/api/spotify/play", { method: "POST", body: JSON.stringify({ uri, deviceId }) });
}

export function resume(deviceId?: string): Promise<void> {
  return apiRequest("/api/spotify/play", { method: "POST", body: JSON.stringify({ deviceId }) });
}

export function pause(): Promise<void> {
  return apiRequest("/api/spotify/pause", { method: "POST" });
}

export function next(): Promise<void> {
  return apiRequest("/api/spotify/next", { method: "POST" });
}

export function previous(): Promise<void> {
  return apiRequest("/api/spotify/previous", { method: "POST" });
}

export function seek(positionMs: number): Promise<void> {
  return apiRequest("/api/spotify/seek", { method: "POST", body: JSON.stringify({ positionMs }) });
}

export function setVolume(percent: number): Promise<void> {
  return apiRequest("/api/spotify/volume", { method: "POST", body: JSON.stringify({ percent }) });
}

export function search(query: string): Promise<Track[]> {
  return apiRequest(`/api/spotify/search?q=${encodeURIComponent(query)}`);
}

export function fetchPlaylists(): Promise<Playlist[]> {
  return apiRequest("/api/spotify/playlists");
}

export function fetchPlaylistTracks(playlistId: string): Promise<Track[]> {
  return apiRequest(`/api/spotify/playlists/${playlistId}/tracks`);
}

export function playPlaylist(playlistId: string, deviceId?: string): Promise<void> {
  return apiRequest(`/api/spotify/playlists/${playlistId}/play`, {
    method: "POST",
    body: JSON.stringify({ deviceId }),
  });
}
