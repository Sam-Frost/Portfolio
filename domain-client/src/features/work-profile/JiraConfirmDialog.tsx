import type { KeyboardEvent } from "react";

interface Props {
  taskName: string;
  onCancel: () => void;
  onConfirm: () => void;
}

// Shown when marking a Work Profile task done. The backend also rejects a
// completion without the acknowledgement, so this can't be bypassed by
// racing the dialog.
export function JiraConfirmDialog({ taskName, onCancel, onConfirm }: Props) {
  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 px-4"
      onClick={onCancel}
      onKeyDown={(e: KeyboardEvent<HTMLDivElement>) => e.key === "Escape" && onCancel()}
      role="presentation"
    >
      <div
        className="w-full max-w-xs rounded-xl border-(--line) border-[0.5px] border-solid bg-(--card) p-4 shadow-lg"
        onClick={(e) => e.stopPropagation()}
        role="dialog"
        aria-modal="true"
        aria-label="Confirm Jira entry"
      >
        <h2 className="text-[length:var(--text-caption)] text-(--fg) mb-1">Added to Jira?</h2>
        <p className="text-[length:var(--text-pill)] text-(--text-muted) mb-4">
          Confirm <span className="text-(--fg)">{taskName}</span> has been logged in Jira before
          marking it done.
        </p>
        <div className="flex items-center justify-end gap-1.5">
          <button
            type="button"
            onClick={onCancel}
            className="rounded-md px-2.5 py-1 text-[length:var(--text-pill)] text-(--text-muted) hover:text-(--fg) hover:bg-(--card-alt) transition-colors cursor-pointer"
          >
            Not yet
          </button>
          <button
            type="button"
            onClick={onConfirm}
            className="rounded-md bg-(--fg) text-(--bg) px-2.5 py-1 text-[length:var(--text-pill)] cursor-pointer"
          >
            Yes, mark done
          </button>
        </div>
      </div>
    </div>
  );
}
