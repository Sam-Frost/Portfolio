import { useState, type KeyboardEvent, type ReactNode } from "react";

interface ConfirmDialogProps {
  title: string;
  body: ReactNode;
  confirmLabel?: string;
  onCancel: () => void;
  onConfirm: () => Promise<void> | void;
}

// Destructive-action confirm modal, shared by "delete document" and
// "delete folder" (which passes a cascade warning as `body`).
export function ConfirmDialog({ title, body, confirmLabel = "Delete", onCancel, onConfirm }: ConfirmDialogProps) {
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function confirm() {
    if (busy) return;
    setBusy(true);
    setError(null);
    try {
      await onConfirm();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Something went wrong.");
      setBusy(false);
    }
  }

  function handleKeyDown(e: KeyboardEvent<HTMLDivElement>) {
    if (e.key === "Escape") onCancel();
  }

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 px-4"
      onClick={onCancel}
      onKeyDown={handleKeyDown}
      role="presentation"
    >
      <div
        className="w-full max-w-sm rounded-xl border-(--line) border-[0.5px] border-solid bg-(--card) p-4 shadow-lg"
        onClick={(e) => e.stopPropagation()}
        role="dialog"
        aria-modal="true"
        aria-label={title}
      >
        <h2 className="text-[length:var(--text-caption)] text-(--fg) mb-1.5">{title}</h2>
        <div className="text-[length:var(--text-pill)] text-(--text-muted) mb-4">{body}</div>

        {error && <p className="mb-3 text-[length:var(--text-pill)] text-red-400">{error}</p>}

        <div className="flex items-center justify-end gap-1.5">
          <button
            type="button"
            onClick={onCancel}
            className="rounded-md px-2.5 py-1 text-[length:var(--text-pill)] text-(--text-muted) hover:text-(--fg) hover:bg-(--card-alt) transition-colors cursor-pointer"
          >
            Cancel
          </button>
          <button
            type="button"
            disabled={busy}
            onClick={confirm}
            className="rounded-md bg-(--label-red) text-(--bg) px-2.5 py-1 text-[length:var(--text-pill)] cursor-pointer disabled:opacity-50"
          >
            {confirmLabel}
          </button>
        </div>
      </div>
    </div>
  );
}
