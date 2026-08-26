import { Music2, Pause, Play, SkipBack, SkipForward, Volume2 } from "lucide-react";
import { useState } from "react";
import { DeviceChips } from "./DeviceChips";
import { useSpotifyPlayer } from "./SpotifyPlayerProvider";

function formatTime(ms: number): string {
  const totalSeconds = Math.floor(ms / 1000);
  const minutes = Math.floor(totalSeconds / 60);
  const seconds = totalSeconds % 60;
  return `${minutes}:${seconds.toString().padStart(2, "0")}`;
}

export function NowPlayingTab() {
  const { playback, togglePlayPause, next, previous, seek, setVolume } = useSpotifyPlayer();
  const [volume, setVolumeLocal] = useState(50);

  const track = playback?.track;

  return (
    <div className="flex flex-col gap-2.5">
      <div className="aspect-square w-full rounded-lg bg-(--card-alt) flex items-center justify-center overflow-hidden">
        {track?.albumArt ? (
          <img src={track.albumArt} alt="" className="h-full w-full object-cover" />
        ) : (
          <Music2 size={28} className="text-(--green)" />
        )}
      </div>

      <div className="min-w-0">
        <p className="truncate text-[length:var(--text-caption)] text-(--fg)">{track?.name ?? "Nothing playing"}</p>
        <p className="truncate text-[10px] text-(--text-faint)">{track?.artists ?? "Pick a playlist or search a track"}</p>
      </div>

      {track && (
        <div
          className="h-[3px] rounded-full bg-(--ring-track) cursor-pointer"
          onClick={(e) => {
            const rect = e.currentTarget.getBoundingClientRect();
            const ratio = Math.min(1, Math.max(0, (e.clientX - rect.left) / rect.width));
            seek(Math.round(ratio * track.durationMs));
          }}
        >
          <div
            className="h-full rounded-full bg-(--green)"
            style={{ width: `${track.durationMs ? Math.min(100, (playback!.progressMs / track.durationMs) * 100) : 0}%` }}
          />
        </div>
      )}
      {track && (
        <div className="flex justify-between text-[9px] text-(--text-faint) tabular-nums -mt-1">
          <span>{formatTime(playback!.progressMs)}</span>
          <span>{formatTime(track.durationMs)}</span>
        </div>
      )}

      <div className="flex items-center justify-center gap-5">
        <button type="button" onClick={previous} className="text-(--text-muted) hover:text-(--fg)">
          <SkipBack size={16} />
        </button>
        <button
          type="button"
          onClick={togglePlayPause}
          className="h-8 w-8 rounded-full bg-(--fg) text-(--bg) flex items-center justify-center"
        >
          {playback?.isPlaying ? <Pause size={14} /> : <Play size={14} className="ml-0.5" />}
        </button>
        <button type="button" onClick={next} className="text-(--text-muted) hover:text-(--fg)">
          <SkipForward size={16} />
        </button>
      </div>

      <div className="flex items-center gap-2">
        <Volume2 size={12} className="text-(--text-faint) shrink-0" />
        <input
          type="range"
          min={0}
          max={100}
          value={volume}
          onChange={(e) => {
            const v = Number(e.target.value);
            setVolumeLocal(v);
            setVolume(v);
          }}
          className="w-full accent-(--green)"
        />
      </div>

      <DeviceChips />
    </div>
  );
}
