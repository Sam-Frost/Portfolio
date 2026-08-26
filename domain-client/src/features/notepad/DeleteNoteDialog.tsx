import type { KeyboardEvent } from "react";

interface DeleteNoteDialogProps {
  title: string;
  onCancel: () => void;
  onConfirm: () => void;
}

export function DeleteNoteDialog({ title, onCancel, onConfirm }: DeleteNoteDialogProps) {
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
        aria-label="Delete note"
      >
        <h2 className="text-[length:var(--text-caption)] text-(--fg) mb-1.5">Delete note?</h2>
        <p className="text-[length:var(--text-pill)] text-(--text-muted) mb-4">
          "{title}" will be removed from your notes.
        </p>

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
            onClick={onConfirm}
            className="rounded-md bg-(--label-red) text-(--bg) px-2.5 py-1 text-[length:var(--text-pill)] cursor-pointer"
          >
            Delete
          </button>
        </div>
      </div>
    </div>
  );
}
