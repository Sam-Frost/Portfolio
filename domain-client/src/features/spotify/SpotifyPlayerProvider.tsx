import { createContext, useContext, useEffect, useRef, useState, type ReactNode } from "react";
import * as api from "./api";
import type { PlaybackState, SpotifyDevice } from "./types";

const SDK_SCRIPT_SRC = "https://sdk.scdn.co/spotify-player.js";
const POLL_INTERVAL_MS = 5000;

interface SpotifyContextValue {
  // null while the initial /status check is in flight.
  connected: boolean | null;
  connectError: boolean;
  dismissConnectError: () => void;
  connect: () => Promise<void>;

  playback: PlaybackState | null;
  devices: SpotifyDevice[];
  thisDeviceId: string | null;
  refreshDevices: () => Promise<void>;

  playTrack: (uri: string) => Promise<void>;
  playPlaylist: (playlistId: string) => Promise<void>;
  togglePlayPause: () => Promise<void>;
  next: () => Promise<void>;
  previous: () => Promise<void>;
  seek: (positionMs: number) => Promise<void>;
  setVolume: (percent: number) => Promise<void>;
  transferTo: (deviceId: string) => Promise<void>;
}

const SpotifyContext = createContext<SpotifyContextValue | null>(null);

export function useSpotifyPlayer(): SpotifyContextValue {
  const ctx = useContext(SpotifyContext);
  if (!ctx) throw new Error("useSpotifyPlayer must be used within SpotifyPlayerProvider");
  return ctx;
}

function loadSdkScript(): void {
  if (document.getElementById("spotify-web-playback-sdk")) return;
  const script = document.createElement("script");
  script.id = "spotify-web-playback-sdk";
  script.src = SDK_SCRIPT_SRC;
  script.async = true;
  document.body.appendChild(script);
}

export function SpotifyPlayerProvider({ children }: { children: ReactNode }) {
  const [connected, setConnected] = useState<boolean | null>(null);
  const [connectError, setConnectError] = useState(false);
  const [playback, setPlayback] = useState<PlaybackState | null>(null);
  const [devices, setDevices] = useState<SpotifyDevice[]>([]);
  const [thisDeviceId, setThisDeviceId] = useState<string | null>(null);
  const thisDeviceIdRef = useRef<string | null>(null);

  // Surfaces the ?spotify=connected|error redirect from the OAuth callback
  // (see server/internal/spotify's handler) once, then cleans the URL so a
  // refresh doesn't re-trigger it.
  useEffect(() => {
    const params = new URLSearchParams(window.location.search);
    const result = params.get("spotify");
    if (!result) return;

    if (result === "error") setConnectError(true);
    params.delete("spotify");
    const cleaned = window.location.pathname + (params.toString() ? `?${params}` : "");
    window.history.replaceState(null, "", cleaned);
  }, []);

  const refreshState = async () => {
    try {
      const state = await api.fetchState();
      setPlayback(state ?? null);
    } catch {
      // Transient failures just leave the last known state on screen.
    }
  };

  const refreshDevices = async () => {
    try {
      setDevices(await api.fetchDevices());
    } catch {
      // Ignored — devices list is best-effort UI, not load-bearing.
    }
  };

  useEffect(() => {
    api
      .fetchStatus()
      .then((s) => setConnected(s.connected))
      .catch(() => setConnected(false));
  }, []);

  // Poll for playback state (covers "playing on another device") whenever
  // connected; the SDK's own player_state_changed event (below) also feeds
  // this and fires much faster for this specific browser tab.
  useEffect(() => {
    if (!connected) return;

    refreshState();
    refreshDevices();
    const interval = window.setInterval(refreshState, POLL_INTERVAL_MS);
    return () => window.clearInterval(interval);
  }, [connected]);

  // Registers this browser tab as a Spotify Connect device via the Web
  // Playback SDK. Every control action still goes through the backend (see
  // api.ts) — the SDK's only jobs are making this tab a playable device and
  // providing lower-latency state updates when it's the active one.
  useEffect(() => {
    if (!connected) return;

    let player: SpotifySdkPlayer | null = null;

    window.onSpotifyWebPlaybackSDKReady = () => {
      if (!window.Spotify) return;

      player = new window.Spotify.Player({
        name: "sat0ru.dev",
        getOAuthToken: (callback) => {
          api
            .fetchSdkToken()
            .then((t) => callback(t.accessToken))
            .catch(() => {
              // Leaving the SDK without a token just keeps this tab
              // unavailable as a device; the rest of the widget still works
              // against whatever device is already active.
            });
        },
        volume: 0.5,
      });

      player.addListener("ready", ({ device_id }) => {
        thisDeviceIdRef.current = device_id;
        setThisDeviceId(device_id);
        refreshDevices();
      });

      player.addListener("player_state_changed", (state) => {
        if (!state) return;
        const track = state.track_window.current_track;
        setPlayback({
          isPlaying: !state.paused,
          progressMs: state.position,
          track: {
            id: track.uri,
            name: track.name,
            artists: track.artists.map((a) => a.name).join(", "),
            albumArt: track.album.images[0]?.url ?? "",
            uri: track.uri,
            durationMs: track.duration_ms,
          },
          device: thisDeviceIdRef.current
            ? { id: thisDeviceIdRef.current, name: "This browser", type: "Computer", isActive: true, volumePercent: 50 }
            : null,
        });
      });

      player.connect();
    };

    loadSdkScript();
    // If the script tag already existed from a previous mount (e.g. fast
    // refresh in dev), the ready callback above was already assigned but
    // the SDK already fired it before this effect ran — nudge it directly.
    if (window.Spotify) window.onSpotifyWebPlaybackSDKReady?.();

    return () => {
      player?.disconnect();
    };
  }, [connected]);

  const connect = async () => {
    setConnectError(false);
    const url = await api.connectUrl();
    window.location.href = url;
  };

  // Every control below applies the change through the backend, then
  // immediately re-polls rather than waiting up to POLL_INTERVAL_MS for the
  // next tick, so the widget doesn't feel laggy on this tab or elsewhere.
  const withRefresh = async (action: () => Promise<void>) => {
    await action();
    await refreshState();
  };

  // Nothing is marked "active" on Spotify's side until something has
  // actually been transferred to or started on a device — a freshly
  // registered Web Playback SDK device (this tab) doesn't count yet, and
  // /play rejects requests with nowhere to start. In that case, target this
  // tab directly so the very first play click just works instead of
  // erroring; once some device is active, leave targeting alone so an
  // already-active device (e.g. the phone) isn't silently hijacked.
  const noActiveDevice = !devices.some((d) => d.isActive);
  const bootstrapDeviceId = noActiveDevice && thisDeviceId ? thisDeviceId : undefined;

  const value: SpotifyContextValue = {
    connected,
    connectError,
    dismissConnectError: () => setConnectError(false),
    connect,
    playback,
    devices,
    thisDeviceId,
    refreshDevices,
    playTrack: (uri) => withRefresh(() => api.playTrack(uri, bootstrapDeviceId)),
    playPlaylist: (playlistId) => withRefresh(() => api.playPlaylist(playlistId, bootstrapDeviceId)),
    togglePlayPause: () =>
      withRefresh(() => (playback?.isPlaying ? api.pause() : api.resume(bootstrapDeviceId))),
    next: () => withRefresh(api.next),
    previous: () => withRefresh(api.previous),
    seek: (positionMs) => withRefresh(() => api.seek(positionMs)),
    setVolume: (percent) => api.setVolume(percent),
    transferTo: (deviceId) => withRefresh(() => api.transferDevice(deviceId).then(refreshDevices)),
  };

  return <SpotifyContext.Provider value={value}>{children}</SpotifyContext.Provider>;
}
