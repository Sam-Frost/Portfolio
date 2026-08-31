import { useState, type KeyboardEvent } from "react";
import { formatMinutes } from "./dateUtils";
import { GoalChecklist } from "./GoalChecklist";
import type { FinishPayload, Goal, WorkSession } from "./types";

interface CancelSessionDialogProps {
  session: WorkSession;
  elapsedSeconds: number;
  onCancel: () => void;
  onConfirm: (payload: FinishPayload) => Promise<void>;
}

export function CancelSessionDialog({ session, elapsedSeconds, onCancel, onConfirm }: CancelSessionDialogProps) {
  const [goals, setGoals] = useState<Goal[]>(() =>
    (session.goals ?? []).map((g) => ({ text: g.text, done: g.done })),
  );
  const [note, setNote] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const elapsedMinutes = Math.max(0, Math.round(elapsedSeconds / 60));

  function handleKeyDown(e: KeyboardEvent<HTMLDivElement>) {
    if (e.key === "Escape" && !submitting) onCancel();
  }

  function toggleGoal(index: number) {
    setGoals((prev) => prev.map((g, i) => (i === index ? { ...g, done: !g.done } : g)));
  }

  async function handleConfirm() {
    setSubmitting(true);
    setError(null);
    try {
      await onConfirm({ goals, note: note.trim() });
    } catch (err) {
      setError(err instanceof Error ? err.message : "Couldn't cancel the session.");
      setSubmitting(false);
    }
  }

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 px-4"
      onClick={submitting ? undefined : onCancel}
      onKeyDown={handleKeyDown}
      role="presentation"
    >
      <div
        className="w-full max-w-sm rounded-xl border-(--line) border-[0.5px] border-solid bg-(--card) p-4 shadow-lg"
        onClick={(e) => e.stopPropagation()}
        role="dialog"
        aria-modal="true"
        aria-label="Cancel session"
      >
        <h2 className="text-[length:var(--text-caption)] text-(--fg) mb-1.5">Cancel this session?</h2>
        <p className="text-[length:var(--text-pill)] text-(--text-muted) mb-4">
          Planned for {formatMinutes(session.plannedMinutes)}, ran for {formatMinutes(elapsedMinutes)} so far. This
          will be logged as cancelled, ending now.
        </p>

        {goals.length > 0 && (
          <>
            <p className="text-[length:var(--text-pill)] text-(--text-muted) mb-1.5">Goals</p>
            <div className="mb-4">
              <GoalChecklist goals={goals} onToggle={toggleGoal} disabled={submitting} />
            </div>
          </>
        )}

        <label className="block text-[length:var(--text-pill)] text-(--text-muted) mb-1">Remarks (optional)</label>
        <textarea
          value={note}
          onChange={(e) => setNote(e.target.value)}
          rows={2}
          disabled={submitting}
          className="w-full mb-3 resize-none rounded-lg bg-(--card-alt) border-(--line) border-[0.5px] border-solid px-2.5 py-1.5 text-[length:var(--text-caption)] text-(--fg) focus:outline-none"
        />

        {error && <p className="mb-3 text-[length:var(--text-pill)] text-red-400">{error}</p>}

        <div className="flex items-center justify-end gap-1.5">
          <button
            type="button"
            onClick={onCancel}
            disabled={submitting}
            className="rounded-md px-2.5 py-1 text-[length:var(--text-pill)] text-(--text-muted) hover:text-(--fg) hover:bg-(--card-alt) transition-colors cursor-pointer disabled:opacity-50"
          >
            Keep going
          </button>
          <button
            type="button"
            onClick={handleConfirm}
            disabled={submitting}
            className="rounded-md bg-(--fg) text-(--bg) px-2.5 py-1 text-[length:var(--text-pill)] cursor-pointer disabled:opacity-50"
          >
            {submitting ? "Cancelling..." : "Cancel session"}
          </button>
        </div>
      </div>
    </div>
  );
}
