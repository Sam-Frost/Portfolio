import { useState, type KeyboardEvent } from "react";

interface PromptDialogProps {
  title: string;
  label: string;
  initialValue?: string;
  confirmLabel?: string;
  onCancel: () => void;
  onConfirm: (value: string) => Promise<void> | void;
}

// A single-text-field modal, shared by "New folder" and the rename flows.
export function PromptDialog({
  title,
  label,
  initialValue = "",
  confirmLabel = "Save",
  onCancel,
  onConfirm,
}: PromptDialogProps) {
  const [value, setValue] = useState(initialValue);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function submit() {
    const trimmed = value.trim();
    if (!trimmed || saving) return;
    setSaving(true);
    setError(null);
    try {
      await onConfirm(trimmed);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Something went wrong.");
      setSaving(false);
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
        <h2 className="text-[length:var(--text-caption)] text-(--fg) mb-3">{title}</h2>

        <label className="flex flex-col gap-1.5">
          <span className="text-[length:var(--text-pill)] text-(--text-muted)">{label}</span>
          <input
            autoFocus
            value={value}
            onChange={(e) => setValue(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === "Enter") submit();
            }}
            className="rounded-md border-(--line) border-[0.5px] border-solid bg-(--bg) px-2.5 py-1.5 text-[length:var(--text-caption)] text-(--fg) outline-none focus:border-(--line-strong)"
          />
        </label>

        {error && <p className="mt-2 text-[length:var(--text-pill)] text-red-400">{error}</p>}

        <div className="mt-4 flex items-center justify-end gap-1.5">
          <button
            type="button"
            onClick={onCancel}
            className="rounded-md px-2.5 py-1 text-[length:var(--text-pill)] text-(--text-muted) hover:text-(--fg) hover:bg-(--card-alt) transition-colors cursor-pointer"
          >
            Cancel
          </button>
          <button
            type="button"
            disabled={saving || !value.trim()}
            onClick={submit}
            className="rounded-md bg-(--fg) text-(--bg) px-2.5 py-1 text-[length:var(--text-pill)] cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {confirmLabel}
          </button>
        </div>
      </div>
    </div>
  );
}
