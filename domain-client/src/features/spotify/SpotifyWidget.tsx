import { ListMusic, Music2, Pause, Play, Search, SkipBack, SkipForward } from "lucide-react";
import { useEffect, useRef, useState } from "react";
import { NowPlayingTab } from "./NowPlayingTab";
import { PlaylistsTab } from "./PlaylistsTab";
import { SearchTab } from "./SearchTab";
import { useSpotifyPlayer } from "./SpotifyPlayerProvider";

type Tab = "playing" | "playlists" | "search";

function TabButton({ label, active, onClick }: { label: string; active: boolean; onClick: () => void }) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={`flex-1 py-2 text-center text-[10px] font-space font-semibold tracking-wide transition-colors ${
        active ? "text-(--fg) shadow-[inset_0_-2px_0_var(--green)]" : "text-(--text-faint) hover:text-(--text-muted)"
      }`}
    >
      {label}
    </button>
  );
}

export function SpotifyWidget() {
  const { connected, connectError, dismissConnectError, connect, playback, togglePlayPause, next, previous } =
    useSpotifyPlayer();
  const [open, setOpen] = useState(false);
  const [tab, setTab] = useState<Tab>("playing");
  const containerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!open) return;

    const onPointerDown = (e: PointerEvent) => {
      if (containerRef.current && !containerRef.current.contains(e.target as Node)) setOpen(false);
    };
    const onKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape") setOpen(false);
    };

    document.addEventListener("pointerdown", onPointerDown);
    document.addEventListener("keydown", onKeyDown);
    return () => {
      document.removeEventListener("pointerdown", onPointerDown);
      document.removeEventListener("keydown", onKeyDown);
    };
  }, [open]);

  const openTab = (t: Tab) => {
    setTab(t);
    setOpen(true);
  };

  if (connected === null) return null;

  if (!connected) {
    return (
      <div className="mt-auto rounded-lg border-(--line-strong) border border-dashed px-3 py-3.5 text-center bg-(--card-alt)">
        <div className="mx-auto mb-2.5 flex h-7 w-7 items-center justify-center rounded-full bg-(--green-bg) text-(--green)">
          <Music2 size={13} />
        </div>
        <p className="mb-2.5 text-[11px] leading-relaxed text-(--text-muted)">
          Connect your Spotify account to play music here.
        </p>
        {connectError && (
          <p className="mb-2 text-[10px] text-red-400">Couldn't connect — try again.</p>
        )}
        <button
          type="button"
          onClick={() => {
            dismissConnectError();
            connect();
          }}
          className="w-full rounded-md bg-(--green) py-1.5 text-[10.5px] font-space font-semibold text-(--bg)"
        >
          Connect Spotify
        </button>
      </div>
    );
  }

  const track = playback?.track;

  return (
    <div className="relative mt-auto" ref={containerRef}>
      {open && (
        <div className="absolute bottom-full left-0 z-20 mb-2 w-60 overflow-hidden rounded-xl border-(--line-strong) border-[0.5px] border-solid bg-(--card) shadow-2xl">
          <div className="flex border-b-(--line-soft) border-b-[0.5px] border-solid">
            <TabButton label="Playing" active={tab === "playing"} onClick={() => setTab("playing")} />
            <TabButton label="Playlists" active={tab === "playlists"} onClick={() => setTab("playlists")} />
            <TabButton label="Search" active={tab === "search"} onClick={() => setTab("search")} />
          </div>
          <div className="p-3">
            {tab === "playing" && <NowPlayingTab />}
            {tab === "playlists" && <PlaylistsTab onPlayed={() => setOpen(false)} />}
            {tab === "search" && <SearchTab onPlayed={() => setOpen(false)} />}
          </div>
        </div>
      )}

      <div className="overflow-hidden rounded-lg border-(--line-soft) border-[0.5px] border-solid bg-(--card-alt)">
        <div className="flex justify-end gap-1 px-1.5 pt-1.5">
          <button
            type="button"
            onClick={() => openTab("search")}
            aria-label="Search"
            className="rounded p-1 text-(--text-faint) hover:text-(--fg) hover:bg-(--card)"
          >
            <Search size={11} />
          </button>
          <button
            type="button"
            onClick={() => openTab("playlists")}
            aria-label="Playlists"
            className="rounded p-1 text-(--text-faint) hover:text-(--fg) hover:bg-(--card)"
          >
            <ListMusic size={11} />
          </button>
        </div>

        <button
          type="button"
          onClick={() => openTab("playing")}
          className="flex w-full items-center gap-2 px-2 py-1.5 text-left"
        >
          <div className="flex h-7 w-7 shrink-0 items-center justify-center overflow-hidden rounded bg-(--card) text-(--green)">
            {track?.albumArt ? <img src={track.albumArt} alt="" className="h-full w-full object-cover" /> : <Music2 size={12} />}
          </div>
          <div className="min-w-0 flex-1">
            <p className="truncate text-[10.5px] text-(--fg)">{track?.name ?? "Nothing playing"}</p>
            <p className="truncate text-[9px] text-(--text-faint)">{track?.artists ?? "Search or pick a playlist"}</p>
          </div>
        </button>

        <div className="flex items-center justify-center gap-3.5 pb-1.5">
          <button
            type="button"
            onClick={(e) => {
              e.stopPropagation();
              previous();
            }}
            className="text-(--text-muted) hover:text-(--fg)"
          >
            <SkipBack size={12} />
          </button>
          <button
            type="button"
            onClick={(e) => {
              e.stopPropagation();
              togglePlayPause();
            }}
            className="flex h-5 w-5 items-center justify-center rounded-full bg-(--fg) text-(--bg)"
          >
            {playback?.isPlaying ? <Pause size={9} /> : <Play size={9} className="ml-0.5" />}
          </button>
          <button
            type="button"
            onClick={(e) => {
              e.stopPropagation();
              next();
            }}
            className="text-(--text-muted) hover:text-(--fg)"
          >
            <SkipForward size={12} />
          </button>
        </div>

        {track && (
          <div className="mx-2 mb-1.5 h-[2px] rounded-full bg-(--ring-track)">
            <div
              className="h-full rounded-full bg-(--green)"
              style={{ width: `${track.durationMs ? Math.min(100, ((playback?.progressMs ?? 0) / track.durationMs) * 100) : 0}%` }}
            />
          </div>
        )}
      </div>
    </div>
  );
}
