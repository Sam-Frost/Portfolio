import { useState } from "react";
import { Plus, X } from "lucide-react";
import type { StartSessionInput } from "./api";
import type { SessionCategory } from "./types";

interface StartSessionFormProps {
  onStart: (input: StartSessionInput) => Promise<void>;
}

const PRESETS_MINUTES = [25, 50, 90];

const CATEGORIES: { value: SessionCategory; label: string }[] = [
  { value: "professional", label: "Professional" },
  { value: "personal", label: "Personal" },
];

export function StartSessionForm({ onStart }: StartSessionFormProps) {
  const [hours, setHours] = useState(0);
  const [minutes, setMinutes] = useState(25);
  const [category, setCategory] = useState<SessionCategory>("professional");
  const [goals, setGoals] = useState<string[]>([""]);
  const [startNote, setStartNote] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const totalMinutes = hours * 60 + minutes;

  function updateGoal(index: number, value: string) {
    setGoals((prev) => prev.map((g, i) => (i === index ? value : g)));
  }

  function addGoal() {
    setGoals((prev) => [...prev, ""]);
  }

  function removeGoal(index: number) {
    setGoals((prev) => (prev.length === 1 ? [""] : prev.filter((_, i) => i !== index)));
  }

  async function handleStart() {
    if (totalMinutes <= 0) {
      setError("Pick a duration greater than zero.");
      return;
    }

    setSubmitting(true);
    setError(null);
    try {
      await onStart({
        plannedMinutes: totalMinutes,
        category,
        goals: goals.map((g) => g.trim()).filter(Boolean),
        startNote: startNote.trim(),
      });
    } catch (err) {
      setError(err instanceof Error ? err.message : "Couldn't start the session.");
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <div className="bg-(--card) border-(--line) border-[0.5px] border-solid rounded-xl p-5">
      <h2 className="text-[length:var(--text-pill)] font-medium uppercase tracking-wide text-(--text-faint) mb-3">
        Start a session
      </h2>

      <div className="flex flex-wrap items-end gap-3 mb-3">
        <div>
          <label className="block text-[length:var(--text-pill)] text-(--text-muted) mb-1">Hours</label>
          <input
            type="number"
            min={0}
            max={12}
            value={hours}
            onChange={(e) => setHours(Math.min(12, Math.max(0, Number(e.target.value) || 0)))}
            className="w-20 rounded-lg bg-(--card-alt) border-(--line) border-[0.5px] border-solid px-2.5 py-1.5 text-[length:var(--text-caption)] text-(--fg) focus:outline-none"
          />
        </div>
        <div>
          <label className="block text-[length:var(--text-pill)] text-(--text-muted) mb-1">Minutes</label>
          <input
            type="number"
            min={0}
            max={59}
            value={minutes}
            onChange={(e) => setMinutes(Math.min(59, Math.max(0, Number(e.target.value) || 0)))}
            className="w-20 rounded-lg bg-(--card-alt) border-(--line) border-[0.5px] border-solid px-2.5 py-1.5 text-[length:var(--text-caption)] text-(--fg) focus:outline-none"
          />
        </div>
      </div>

      <div className="flex items-center gap-1.5 mb-4">
        {PRESETS_MINUTES.map((preset) => (
          <button
            key={preset}
            type="button"
            onClick={() => {
              setHours(Math.floor(preset / 60));
              setMinutes(preset % 60);
            }}
            className="rounded-full px-2.5 py-1 text-[length:var(--text-pill)] text-(--text-muted) hover:text-(--fg) hover:bg-(--card-alt) transition-colors cursor-pointer"
          >
            {preset < 60 ? `${preset}m` : `${preset / 60}h`}
          </button>
        ))}
      </div>

      <label className="block text-[length:var(--text-pill)] text-(--text-muted) mb-1.5">Category</label>
      <div className="flex items-center gap-1.5 mb-4">
        {CATEGORIES.map((c) => (
          <button
            key={c.value}
            type="button"
            onClick={() => setCategory(c.value)}
            aria-pressed={category === c.value}
            className={`rounded-full px-3 py-1 text-[length:var(--text-pill)] border-[0.5px] border-solid transition-colors cursor-pointer ${
              category === c.value
                ? "border-(--line-strong) bg-(--card-alt) text-(--fg)"
                : "border-(--line) text-(--text-muted) hover:text-(--fg)"
            }`}
          >
            {c.label}
          </button>
        ))}
      </div>

      <label className="block text-[length:var(--text-pill)] text-(--text-muted) mb-1.5">Goals</label>
      <div className="flex flex-col gap-1.5 mb-2">
        {goals.map((goal, i) => (
          <div key={i} className="flex items-center gap-1.5">
            <span className="text-(--text-faint) select-none">•</span>
            <input
              value={goal}
              onChange={(e) => updateGoal(i, e.target.value)}
              onKeyDown={(e) => {
                if (e.key === "Enter") {
                  e.preventDefault();
                  if (i === goals.length - 1 && goal.trim()) addGoal();
                }
              }}
              placeholder={`Goal ${i + 1}`}
              className="flex-1 min-w-0 rounded-lg bg-(--card-alt) border-(--line) border-[0.5px] border-solid px-2.5 py-1.5 text-[length:var(--text-caption)] text-(--fg) focus:outline-none"
            />
            <button
              type="button"
              onClick={() => removeGoal(i)}
              aria-label={`Remove goal ${i + 1}`}
              className="shrink-0 text-(--text-faint) hover:text-(--fg) transition-colors cursor-pointer"
            >
              <X size={13} />
            </button>
          </div>
        ))}
      </div>
      <button
        type="button"
        onClick={addGoal}
        className="mb-4 inline-flex items-center gap-1 text-[length:var(--text-pill)] text-(--text-muted) hover:text-(--fg) transition-colors cursor-pointer"
      >
        <Plus size={12} /> Add goal
      </button>

      <label className="block text-[length:var(--text-pill)] text-(--text-muted) mb-1.5">Remarks (optional)</label>
      <textarea
        value={startNote}
        onChange={(e) => setStartNote(e.target.value)}
        rows={2}
        placeholder="Anything else worth noting before you start"
        className="w-full mb-4 resize-none rounded-lg bg-(--card-alt) border-(--line) border-[0.5px] border-solid px-2.5 py-1.5 text-[length:var(--text-caption)] text-(--fg) focus:outline-none"
      />

      <button
        type="button"
        onClick={handleStart}
        disabled={submitting}
        className="rounded-md bg-(--fg) text-(--bg) px-3 py-1.5 text-[length:var(--text-pill)] cursor-pointer disabled:opacity-50"
      >
        {submitting ? "Starting..." : "Start"}
      </button>

      {error && <p className="mt-3 text-[length:var(--text-pill)] text-red-400">{error}</p>}
    </div>
  );
}
