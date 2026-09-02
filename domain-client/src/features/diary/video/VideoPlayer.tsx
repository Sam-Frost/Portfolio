import { useEffect, useRef, useState } from "react";
import { Play, Trash2, Loader2 } from "lucide-react";
import { deleteVideo, getPlaybackUrl } from "./videoApi";
import type { DiaryVideo } from "./videoTypes";

interface VideoPlayerProps {
  video: DiaryVideo;
  canDelete: boolean;
  onDeleted: (id: string) => void;
}

function formatDuration(sec: number | null): string {
  if (sec == null) return "";
  const m = Math.floor(sec / 60);
  const s = sec % 60;
  return `${m}:${String(s).padStart(2, "0")}`;
}

function formatSize(bytes: number): string {
  if (bytes < 1024 * 1024) return `${Math.max(1, Math.round(bytes / 1024))} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}

// One clip in the day's video log. The playback URL is fetched lazily (on
// the first Play) and is short-lived, so it isn't requested until wanted.
// The blob store serves Range requests, so <video> streams + seeks without
// downloading the whole file.
export function VideoPlayer({ video, canDelete, onDeleted }: VideoPlayerProps) {
  const [url, setUrl] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [deleting, setDeleting] = useState(false);
  const videoElRef = useRef<HTMLVideoElement | null>(null);

  async function loadAndPlay() {
    if (url) return;
    setLoading(true);
    setError(null);
    try {
      setUrl(await getPlaybackUrl(video.id));
    } catch {
      setError("Couldn't load this recording.");
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    if (url && videoElRef.current) void videoElRef.current.play().catch(() => {});
  }, [url]);

  async function handleDelete() {
    setDeleting(true);
    try {
      await deleteVideo(video.id);
      onDeleted(video.id);
    } catch {
      setError("Couldn't delete this recording.");
      setDeleting(false);
    }
  }

  const label = video.title?.trim() || (video.status === "pending" ? "Still uploading…" : "Recording");

  return (
    <div className="rounded-lg border-(--line) border-[0.5px] border-solid bg-(--card) overflow-hidden">
      <div className="relative aspect-video bg-black">
        {url ? (
          <video ref={videoElRef} src={url} controls preload="metadata" className="size-full object-contain" />
        ) : (
          <button
            type="button"
            onClick={() => void loadAndPlay()}
            disabled={loading || video.status === "pending"}
            className="size-full flex items-center justify-center text-white/90 hover:text-white transition-colors cursor-pointer disabled:cursor-default disabled:opacity-50"
          >
            {loading ? <Loader2 size={28} className="animate-spin" /> : <Play size={28} fill="currentColor" />}
          </button>
        )}
      </div>

      <div className="flex items-center justify-between gap-2 px-3 py-2">
        <div className="min-w-0">
          <div className="truncate text-[length:var(--text-pill)] text-(--fg)">{label}</div>
          <div className="text-[length:var(--text-pill)] text-(--text-faint)">
            {[formatDuration(video.durationSeconds), video.status === "ready" ? formatSize(video.sizeBytes) : null]
              .filter(Boolean)
              .join(" · ")}
          </div>
        </div>
        {canDelete && (
          <button
            type="button"
            onClick={() => void handleDelete()}
            disabled={deleting}
            aria-label="Delete recording"
            className="shrink-0 text-(--text-faint) hover:text-(--label-red) transition-colors cursor-pointer disabled:opacity-50"
          >
            <Trash2 size={14} />
          </button>
        )}
      </div>

      {error && <div className="px-3 pb-2 text-[length:var(--text-pill)] text-(--label-red)">{error}</div>}
    </div>
  );
}
