import { useEffect, useRef } from "react";
import { X, Circle, Square, RotateCcw, AlertTriangle } from "lucide-react";
import { useVideoRecorder } from "./useVideoRecorder";
import type { DiaryVideo } from "./videoTypes";

interface VideoRecorderDialogProps {
  date: string;
  onClose: () => void;
  onSaved: (video: DiaryVideo) => void;
}

function formatClock(totalSec: number): string {
  const m = Math.floor(totalSec / 60);
  const s = totalSec % 60;
  return `${m}:${String(s).padStart(2, "0")}`;
}

// A full-screen recorder, mirroring DiagramDialog's shell. Records from the
// webcam/mic and streams the clip to the blob store as it goes (see
// useVideoRecorder / uploader).
export function VideoRecorderDialog({ date, onClose, onSaved }: VideoRecorderDialogProps) {
  const rec = useVideoRecorder(date);
  const videoElRef = useRef<HTMLVideoElement | null>(null);

  // Open the camera as soon as the dialog mounts (unless there's an
  // unfinished upload to deal with first).
  useEffect(() => {
    if (rec.phase === "idle" && !rec.resumable) void rec.beginPreview();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [rec.phase, rec.resumable]);

  useEffect(() => {
    const el = videoElRef.current;
    if (el && rec.previewStream) {
      el.srcObject = rec.previewStream;
      void el.play().catch(() => {});
    }
  }, [rec.previewStream]);

  useEffect(() => {
    if (rec.phase === "done" && rec.result) onSaved(rec.result);
  }, [rec.phase, rec.result, onSaved]);

  const busy = rec.phase === "finalizing";

  async function handleClose() {
    if (rec.phase === "recording" || rec.phase === "preview") await rec.cancel();
    onClose();
  }

  return (
    <div className="fixed inset-0 z-50 flex flex-col bg-(--bg)">
      <div className="shrink-0 flex items-center justify-between gap-3 px-4 py-2.5 border-b-(--line) border-b-[0.5px] border-solid">
        <span className="text-[length:var(--text-caption)] text-(--fg)">
          Video diary · {date}
        </span>
        <button
          type="button"
          onClick={handleClose}
          disabled={busy}
          className="flex items-center gap-1.5 rounded-md px-2.5 py-1 text-[length:var(--text-pill)] text-(--text-muted) hover:text-(--fg) hover:bg-(--card-alt) transition-colors cursor-pointer disabled:opacity-60"
        >
          <X size={13} />
          {rec.phase === "done" ? "Close" : "Cancel"}
        </button>
      </div>

      <div className="flex-1 min-h-0 flex flex-col items-center justify-center gap-4 p-4">
        {rec.resumable && rec.phase !== "done" && (
          <div className="w-full max-w-md rounded-lg border-(--line) border-[0.5px] border-solid bg-(--card) p-4 flex flex-col gap-3">
            <div className="flex items-center gap-2 text-[length:var(--text-caption)] text-(--fg)">
              <AlertTriangle size={14} className="text-(--label-orange)" />
              Unfinished recording from {rec.resumable.date}
            </div>
            <p className="text-[length:var(--text-pill)] text-(--text-muted)">
              A recording didn't finish uploading last time. You can finish uploading what was
              captured, or discard it.
            </p>
            <div className="flex items-center gap-2">
              <button
                type="button"
                onClick={() => void rec.resumeUpload()}
                className="rounded-md bg-(--fg) text-(--bg) px-2.5 py-1 text-[length:var(--text-pill)] cursor-pointer"
              >
                Finish upload
              </button>
              <button
                type="button"
                onClick={() => void rec.discardResumable()}
                className="rounded-md px-2.5 py-1 text-[length:var(--text-pill)] text-(--text-muted) hover:text-(--fg) hover:bg-(--card-alt) transition-colors cursor-pointer"
              >
                Discard
              </button>
            </div>
          </div>
        )}

        {rec.phase === "error" && (
          <div className="w-full max-w-md rounded-lg border-(--line) border-[0.5px] border-solid bg-(--card) px-3 py-2 text-[length:var(--text-pill)] text-(--label-red)">
            {rec.error ?? "Something went wrong."}
          </div>
        )}

        {(rec.phase === "preview" || rec.phase === "recording" || rec.phase === "finalizing") && (
          <div className="relative w-full max-w-2xl aspect-video rounded-xl overflow-hidden bg-black border-(--line) border-[0.5px] border-solid">
            <video ref={videoElRef} muted playsInline className="size-full object-cover" />
            {rec.phase === "recording" && (
              <div className="absolute top-3 left-3 flex items-center gap-1.5 rounded-full bg-black/60 px-2.5 py-1 text-[length:var(--text-pill)] text-white">
                <span className="size-2 rounded-full bg-(--label-red) animate-pulse" />
                {formatClock(rec.elapsedSec)}
              </div>
            )}
          </div>
        )}

        {rec.phase === "done" && (
          <div className="flex flex-col items-center gap-2 text-center">
            <div className="text-[length:var(--text-caption)] text-(--fg)">Recording saved.</div>
            <p className="text-[length:var(--text-pill)] text-(--text-muted)">
              It's in this day's video log below.
            </p>
          </div>
        )}

        {(rec.phase === "recording" || rec.phase === "finalizing") && (
          <div className="w-full max-w-2xl">
            <div className="h-1 rounded-full bg-(--card-alt) overflow-hidden">
              <div
                className="h-full bg-(--green) transition-[width] duration-500"
                style={{ width: `${Math.round(rec.uploadProgress * 100)}%` }}
              />
            </div>
            <div className="mt-1 text-[length:var(--text-pill)] text-(--text-faint)">
              {rec.phase === "finalizing"
                ? "Finishing upload…"
                : "Uploading as you record — a crash loses only the last few seconds."}
            </div>
          </div>
        )}

        <div className="flex items-center gap-2">
          {rec.phase === "preview" && (
            <button
              type="button"
              onClick={() => void rec.startRecording()}
              className="flex items-center gap-1.5 rounded-full bg-(--label-red) text-white px-4 py-2 text-[length:var(--text-caption)] cursor-pointer hover:opacity-90 transition-opacity"
            >
              <Circle size={13} fill="currentColor" />
              Start recording
            </button>
          )}
          {rec.phase === "recording" && (
            <button
              type="button"
              onClick={() => rec.stopRecording()}
              className="flex items-center gap-1.5 rounded-full bg-(--fg) text-(--bg) px-4 py-2 text-[length:var(--text-caption)] cursor-pointer hover:opacity-90 transition-opacity"
            >
              <Square size={13} fill="currentColor" />
              Stop & save
            </button>
          )}
          {rec.phase === "error" && !rec.resumable && (
            <button
              type="button"
              onClick={() => void rec.beginPreview()}
              className="flex items-center gap-1.5 rounded-full border-(--line) border-[0.5px] border-solid px-4 py-2 text-[length:var(--text-caption)] text-(--fg) cursor-pointer hover:bg-(--card-alt) transition-colors"
            >
              <RotateCcw size={13} />
              Try again
            </button>
          )}
        </div>
      </div>
    </div>
  );
}
