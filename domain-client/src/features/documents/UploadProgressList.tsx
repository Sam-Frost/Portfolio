import { AlertCircle, CheckCircle2, X } from "lucide-react";
import type { Upload } from "./useUploads";

interface UploadProgressListProps {
  uploads: Upload[];
  onDismiss: () => void;
}

export function UploadProgressList({ uploads, onDismiss }: UploadProgressListProps) {
  if (uploads.length === 0) return null;

  const inFlight = uploads.filter((u) => u.status === "uploading").length;
  const anyFinished = uploads.some((u) => u.status !== "uploading");

  return (
    <div className="mb-3 rounded-lg border-(--line) border-[0.5px] border-solid bg-(--card) p-2.5">
      <div className="mb-1.5 flex items-center justify-between">
        <span className="text-[length:var(--text-pill)] text-(--text-muted)">
          {inFlight > 0 ? `Uploading ${inFlight}…` : "Uploads"}
        </span>
        {anyFinished && (
          <button
            onClick={onDismiss}
            aria-label="Dismiss finished uploads"
            className="text-(--text-faint) hover:text-(--fg) transition-colors cursor-pointer"
          >
            <X size={12} />
          </button>
        )}
      </div>

      <ul className="flex flex-col gap-1.5">
        {uploads.map((u) => (
          <li key={u.id} className="flex items-center gap-2">
            <span className="flex-1 min-w-0 truncate text-[length:var(--text-pill)] text-(--fg)">{u.name}</span>
            {u.status === "uploading" && (
              <div className="h-1 w-20 shrink-0 overflow-hidden rounded-full bg-(--card-alt)">
                <div className="h-full bg-(--fg) transition-all" style={{ width: `${Math.round(u.progress * 100)}%` }} />
              </div>
            )}
            {u.status === "done" && <CheckCircle2 size={13} className="shrink-0 text-(--label-green)" />}
            {u.status === "error" && (
              <span className="flex items-center gap-1 text-[length:var(--text-pill)] text-red-400" title={u.error}>
                <AlertCircle size={13} className="shrink-0" />
              </span>
            )}
          </li>
        ))}
      </ul>
    </div>
  );
}
