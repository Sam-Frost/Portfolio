import { useEffect, useState } from "react";
import { search } from "./api";
import { useSpotifyPlayer } from "./SpotifyPlayerProvider";
import type { Track } from "./types";

const DEBOUNCE_MS = 350;

export function SearchTab({ onPlayed }: { onPlayed: () => void }) {
  const { playTrack } = useSpotifyPlayer();
  const [query, setQuery] = useState("");
  const [results, setResults] = useState<Track[]>([]);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    const trimmed = query.trim();
    if (!trimmed) {
      setResults([]);
      return;
    }

    setLoading(true);
    const timeout = window.setTimeout(() => {
      search(trimmed)
        .then(setResults)
        .catch(() => setResults([]))
        .finally(() => setLoading(false));
    }, DEBOUNCE_MS);

    return () => window.clearTimeout(timeout);
  }, [query]);

  return (
    <div className="flex flex-col gap-2">
      <input
        autoFocus
        type="text"
        value={query}
        onChange={(e) => setQuery(e.target.value)}
        placeholder="Search tracks…"
        className="w-full rounded-md border-(--line) border-[0.5px] border-solid bg-(--bg) px-2 py-1.5 text-[11px] text-(--fg) placeholder:text-(--text-faint) outline-none focus:border-(--gold)"
      />

      <div className="flex flex-col max-h-56 overflow-y-auto themed-scrollbar -mx-1">
        {loading && <p className="px-1 py-1 text-[10px] text-(--text-faint)">Searching…</p>}
        {!loading &&
          results.map((t) => (
            <button
              key={t.id}
              type="button"
              onClick={() => {
                playTrack(t.uri);
                onPlayed();
              }}
              className="flex items-center gap-2 rounded-md px-1 py-1.5 text-left hover:bg-(--card-alt)"
            >
              <div className="h-6 w-6 shrink-0 rounded bg-(--ring-track) overflow-hidden">
                {t.albumArt && <img src={t.albumArt} alt="" className="h-full w-full object-cover" />}
              </div>
              <div className="min-w-0 flex-1">
                <p className="truncate text-[11px] text-(--fg)">{t.name}</p>
                <p className="truncate text-[9px] text-(--text-faint)">{t.artists}</p>
              </div>
            </button>
          ))}
      </div>
    </div>
  );
}
