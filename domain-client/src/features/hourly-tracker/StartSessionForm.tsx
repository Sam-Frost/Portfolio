import { useState } from "react";

interface StartSessionFormProps {
  onStart: (plannedMinutes: number) => Promise<void>;
}

const PRESETS_MINUTES = [25, 50, 90];

export function StartSessionForm({ onStart }: StartSessionFormProps) {
  const [hours, setHours] = useState(0);
  const [minutes, setMinutes] = useState(25);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const totalMinutes = hours * 60 + minutes;

  async function handleStart() {
    if (totalMinutes <= 0) {
      setError("Pick a duration greater than zero.");
      return;
    }

    setSubmitting(true);
    setError(null);
    try {
      await onStart(totalMinutes);
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
        <button
          type="button"
          onClick={handleStart}
          disabled={submitting}
          className="rounded-md bg-(--fg) text-(--bg) px-3 py-1.5 text-[length:var(--text-pill)] cursor-pointer disabled:opacity-50"
        >
          {submitting ? "Starting..." : "Start"}
        </button>
      </div>

      <div className="flex items-center gap-1.5">
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

      {error && <p className="mt-3 text-[length:var(--text-pill)] text-red-400">{error}</p>}
    </div>
  );
}
