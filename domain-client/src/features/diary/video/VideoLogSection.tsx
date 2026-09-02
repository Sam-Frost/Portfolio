import { useEffect, useState } from "react";
import { Video as VideoIcon } from "lucide-react";
import { listDayVideos } from "./videoApi";
import { VideoPlayer } from "./VideoPlayer";
import { VideoRecorderDialog } from "./VideoRecorderDialog";
import type { DiaryVideo } from "./videoTypes";

interface VideoLogSectionProps {
  date: string;
  locked: boolean;
}

// The "Video log" block under the written entry on DiaryEntryPage: the
// day's recorded clips plus a Record button (hidden once the day's edit
// window has closed, same rule as the text editor's read-only state).
export function VideoLogSection({ date, locked }: VideoLogSectionProps) {
  const [videos, setVideos] = useState<DiaryVideo[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [recording, setRecording] = useState(false);

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    setError(null);
    listDayVideos(date)
      .then((v) => {
        if (!cancelled) setVideos(v);
      })
      .catch(() => {
        if (!cancelled) setError("Couldn't load this day's recordings.");
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, [date]);

  function handleSaved(video: DiaryVideo) {
    setVideos((prev) => (prev.some((v) => v.id === video.id) ? prev : [...prev, video]));
  }

  function handleDeleted(id: string) {
    setVideos((prev) => prev.filter((v) => v.id !== id));
  }

  const ready = videos.filter((v) => v.status === "ready" || v.status === "pending");

  if (locked && !loading && ready.length === 0) return null;

  return (
    <div className="shrink-0 mt-6 border-t-(--line) border-t-[0.5px] border-solid pt-4">
      <div className="flex items-center justify-between gap-2 mb-3">
        <h2 className="flex items-center gap-1.5 text-[length:var(--text-caption)] font-medium text-(--fg)">
          <VideoIcon size={14} className="text-(--text-muted)" />
          Video log
        </h2>
        {!locked && (
          <button
            type="button"
            onClick={() => setRecording(true)}
            className="flex items-center gap-1.5 rounded-lg bg-(--fg) text-(--bg) px-3 py-1.5 text-[length:var(--text-pill)] cursor-pointer transition-opacity hover:opacity-90"
          >
            <VideoIcon size={13} />
            Record video
          </button>
        )}
      </div>

      {error && <div className="mb-3 text-[length:var(--text-pill)] text-(--label-red)">{error}</div>}

      {loading ? (
        <div className="text-[length:var(--text-pill)] text-(--text-faint)">Loading recordings…</div>
      ) : ready.length === 0 ? (
        <div className="text-[length:var(--text-pill)] text-(--text-faint)">
          {locked ? "No recordings for this day." : "No recordings yet — hit Record video to add one."}
        </div>
      ) : (
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-3 lg:max-h-[34vh] lg:overflow-y-auto themed-scrollbar lg:pr-1">
          {ready.map((v) => (
            <VideoPlayer key={v.id} video={v} canDelete={!locked} onDeleted={handleDeleted} />
          ))}
        </div>
      )}

      {recording && (
        <VideoRecorderDialog
          date={date}
          onClose={() => setRecording(false)}
          onSaved={handleSaved}
        />
      )}
    </div>
  );
}
