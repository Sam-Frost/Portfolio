import { useRef, useState, type KeyboardEvent } from "react";
import { formatMinutes } from "./dateUtils";
import { GoalChecklist } from "./GoalChecklist";
import type { FinishPayload, Goal, WorkSession } from "./types";

interface CompleteSessionDialogProps {
  session: WorkSession;
  onSubmit: (payload: FinishPayload) => Promise<void>;
}

// Not dismissable — no backdrop click, Escape is swallowed — a completed
// session must be logged with what came of it, so this is the one dialog in
// the app with no cancel path. "Logged" means either ticking the goals set
// at start or leaving a remark (matching the server's rule).
export function CompleteSessionDialog({ session, onSubmit }: CompleteSessionDialogProps) {
  const [goals, setGoals] = useState<Goal[]>(() =>
    (session.goals ?? []).map((g) => ({ text: g.text, done: g.done })),
  );
  const [note, setNote] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const textareaRef = useRef<HTMLTextAreaElement>(null);

  const hasGoals = goals.length > 0;

  function handleKeyDown(e: KeyboardEvent<HTMLDivElement>) {
    if (e.key === "Escape") e.stopPropagation();
  }

  function toggleGoal(index: number) {
    setGoals((prev) => prev.map((g, i) => (i === index ? { ...g, done: !g.done } : g)));
  }

  async function handleSubmit() {
    const trimmed = note.trim();
    if (!hasGoals && !trimmed) {
      textareaRef.current?.focus();
      setError("Tick a goal or add a remark to log this session.");
      return;
    }

    setSubmitting(true);
    setError(null);
    try {
      await onSubmit({ goals, note: trimmed });
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
        className="w-full max-w-sm max-h-[90vh] overflow-y-auto themed-scrollbar rounded-xl border-(--line) border-[0.5px] border-solid bg-(--card) p-4 shadow-lg"
        onClick={(e) => e.stopPropagation()}
        role="dialog"
        aria-modal="true"
        aria-label="Log session"
      >
        <h2 className="text-[length:var(--text-caption)] text-(--fg) mb-1.5">Time's up</h2>
        <p className="text-[length:var(--text-pill)] text-(--text-muted) mb-4">
          Your {formatMinutes(session.plannedMinutes)} session just finished. How did it go?
        </p>

        {hasGoals && (
          <>
            <p className="text-[length:var(--text-pill)] text-(--text-muted) mb-1.5">Goals</p>
            <div className="mb-4">
              <GoalChecklist goals={goals} onToggle={toggleGoal} disabled={submitting} />
            </div>
          </>
        )}

        <p className="text-[length:var(--text-pill)] text-(--text-muted) mb-1.5">
          Remarks{hasGoals ? " (optional)" : ""}
        </p>
        <textarea
          ref={textareaRef}
          autoFocus={!hasGoals}
          value={note}
          onChange={(e) => setNote(e.target.value)}
          rows={3}
          disabled={submitting}
          placeholder="What got done, what didn't, anything to carry over"
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
