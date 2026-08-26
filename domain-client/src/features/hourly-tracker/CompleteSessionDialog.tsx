import { useRef, useState, type KeyboardEvent } from "react";
import { formatMinutes } from "./dateUtils";
import type { WorkSession } from "./types";

interface CompleteSessionDialogProps {
  session: WorkSession;
  onSubmit: (note: string) => Promise<void>;
}

// Not dismissable — no backdrop click, Escape is swallowed — a completed
// session's note is required (per spec), so this is the one dialog in the
// app with no cancel path.
export function CompleteSessionDialog({ session, onSubmit }: CompleteSessionDialogProps) {
  const [note, setNote] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const textareaRef = useRef<HTMLTextAreaElement>(null);

  function handleKeyDown(e: KeyboardEvent<HTMLDivElement>) {
    if (e.key === "Escape") e.stopPropagation();
  }

  async function handleSubmit() {
    const trimmed = note.trim();
    if (!trimmed) {
      textareaRef.current?.focus();
      return;
    }

    setSubmitting(true);
    setError(null);
    try {
      await onSubmit(trimmed);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Couldn't save the session.");
      setSubmitting(false);
    }
  }

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 px-4"
      onKeyDown={handleKeyDown}
      role="presentation"
    >
      <div
        className="w-full max-w-sm rounded-xl border-(--line) border-[0.5px] border-solid bg-(--card) p-4 shadow-lg"
        onClick={(e) => e.stopPropagation()}
        role="dialog"
        aria-modal="true"
        aria-label="Log session"
      >
        <h2 className="text-[length:var(--text-caption)] text-(--fg) mb-1.5">Time's up</h2>
        <p className="text-[length:var(--text-pill)] text-(--text-muted) mb-4">
          Your {formatMinutes(session.plannedMinutes)} session just finished. What did you work on?
        </p>

        <textarea
          ref={textareaRef}
          autoFocus
          value={note}
          onChange={(e) => setNote(e.target.value)}
          rows={3}
          disabled={submitting}
          placeholder="What did you get done?"
          className="w-full mb-3 resize-none rounded-lg bg-(--card-alt) border-(--line) border-[0.5px] border-solid px-2.5 py-1.5 text-[length:var(--text-caption)] text-(--fg) focus:outline-none"
        />

        {error && <p className="mb-3 text-[length:var(--text-pill)] text-red-400">{error}</p>}

        <div className="flex items-center justify-end gap-1.5">
          <button
            type="button"
            onClick={handleSubmit}
            disabled={submitting}
            className="rounded-md bg-(--fg) text-(--bg) px-2.5 py-1 text-[length:var(--text-pill)] cursor-pointer disabled:opacity-50"
          >
            {submitting ? "Saving..." : "Log session"}
          </button>
        </div>
      </div>
    </div>
  );
}
