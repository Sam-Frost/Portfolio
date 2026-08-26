import { ArrowLeft } from "lucide-react";
import { useEffect, useState } from "react";
import { fetchPlaylists, fetchPlaylistTracks } from "./api";
import { useSpotifyPlayer } from "./SpotifyPlayerProvider";
import type { Playlist, Track } from "./types";

function PlaylistList({ onSelect }: { onSelect: (playlist: Playlist) => void }) {
  const [playlists, setPlaylists] = useState<Playlist[] | null>(null);
  const [error, setError] = useState(false);

  useEffect(() => {
    fetchPlaylists()
      .then(setPlaylists)
      .catch(() => setError(true));
  }, []);

  if (error) {
    return <p className="text-[11px] text-(--text-muted)">Couldn't load playlists.</p>;
  }

  if (!playlists) {
    return <p className="text-[11px] text-(--text-faint)">Loading playlists…</p>;
  }

  return (
    <div className="flex flex-col max-h-64 overflow-y-auto themed-scrollbar -mx-1">
      {playlists.map((p) => (
        <button
          key={p.id}
          type="button"
          onClick={() => onSelect(p)}
          className="flex items-center gap-2 rounded-md px-1 py-1.5 text-left hover:bg-(--card-alt)"
        >
          <div className="h-6 w-6 shrink-0 rounded bg-(--ring-track) overflow-hidden">
            {p.imageUrl && <img src={p.imageUrl} alt="" className="h-full w-full object-cover" />}
          </div>
          <div className="min-w-0 flex-1">
            <p className="truncate text-[11px] text-(--fg)">{p.name}</p>
            <p className="truncate text-[9px] text-(--text-faint)">{p.trackCount} tracks</p>
          </div>
        </button>
      ))}
    </div>
  );
}

function PlaylistTrackList({ playlist, onBack, onPlayed }: { playlist: Playlist; onBack: () => void; onPlayed: () => void }) {
  const { playTrack } = useSpotifyPlayer();
  const [tracks, setTracks] = useState<Track[] | null>(null);
  const [error, setError] = useState(false);

  useEffect(() => {
    fetchPlaylistTracks(playlist.id)
      .then(setTracks)
      .catch(() => setError(true));
  }, [playlist.id]);

  return (
    <div className="flex flex-col gap-2">
      <button
        type="button"
        onClick={onBack}
        className="flex items-center gap-1 text-[10px] text-(--text-muted) hover:text-(--fg) -mt-0.5"
      >
        <ArrowLeft size={11} />
        <span className="truncate">{playlist.name}</span>
      </button>

      {error && <p className="text-[11px] text-(--text-muted)">Couldn't load tracks.</p>}
      {!error && !tracks && <p className="text-[11px] text-(--text-faint)">Loading tracks…</p>}

      {tracks && (
        <div className="flex flex-col max-h-56 overflow-y-auto themed-scrollbar -mx-1">
          {tracks.length === 0 && <p className="px-1 py-1 text-[10px] text-(--text-faint)">This playlist is empty.</p>}
          {tracks.map((t, i) => (
            <button
              key={`${t.id}-${i}`}
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
      )}
    </div>
  );
}

export function PlaylistsTab({ onPlayed }: { onPlayed: () => void }) {
  const [selected, setSelected] = useState<Playlist | null>(null);

  return selected ? (
    <PlaylistTrackList playlist={selected} onBack={() => setSelected(null)} onPlayed={onPlayed} />
  ) : (
    <PlaylistList onSelect={setSelected} />
  );
}
