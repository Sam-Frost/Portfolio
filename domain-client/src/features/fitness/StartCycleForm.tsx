import { useState, type FormEvent } from "react";
import { todayISTKey } from "./dateUtils";
import type { CycleInput } from "./types";

interface StartCycleFormProps {
  onStart: (input: CycleInput) => void;
}

function numOrNull(v: string): number | null {
  const n = Number(v);
  return v.trim() !== "" && Number.isFinite(n) && n > 0 ? n : null;
}

// Starting a cycle is like starting a project — it also archives whatever
// cycle is currently active (handled server-side).
export function StartCycleForm({ onStart }: StartCycleFormProps) {
  const [open, setOpen] = useState(false);
  const [name, setName] = useState("");
  const [startDate, setStartDate] = useState(todayISTKey());
  const [weightStart, setWeightStart] = useState("");
  const [weightTarget, setWeightTarget] = useState("");
  const [proteinTarget, setProteinTarget] = useState("");

  function submit(e: FormEvent) {
    e.preventDefault();
    const trimmed = name.trim();
    if (!trimmed) return;
    onStart({
      name: trimmed,
      startDate,
      weightStart: numOrNull(weightStart),
      weightTarget: numOrNull(weightTarget),
      proteinTarget: numOrNull(proteinTarget),
    });
    setName("");
    setStartDate(todayISTKey());
    setWeightStart("");
    setWeightTarget("");
    setProteinTarget("");
    setOpen(false);
  }

  if (!open) {
    return (
      <button
        type="button"
        onClick={() => setOpen(true)}
        className="mb-4 rounded-lg bg-(--fg) text-(--bg) px-3 py-2 text-[length:var(--text-pill)] cursor-pointer"
      >
        Start a fitness cycle
      </button>
    );
  }

  const field =
    "rounded-lg bg-(--card-alt) border-(--line) border-[0.5px] border-solid px-2.5 py-1.5 text-[length:var(--text-caption)] text-(--fg) placeholder:text-(--text-faint) focus:outline-none";

  return (
    <form onSubmit={submit} className="bg-(--card) border-(--line) border-[0.5px] border-solid rounded-xl p-4 mb-4 flex flex-col gap-3">
      <div className="flex flex-wrap gap-3">
        <label className="flex flex-col gap-1 flex-1 min-w-40">
          <span className="text-[length:var(--text-pill)] text-(--text-muted)">Cycle name</span>
          <input autoFocus value={name} onChange={(e) => setName(e.target.value)} placeholder="e.g. Summer cut" className={field} />
        </label>
        <label className="flex flex-col gap-1">
          <span className="text-[length:var(--text-pill)] text-(--text-muted)">Start date</span>
          <input type="date" value={startDate} onChange={(e) => setStartDate(e.target.value)} className={field} />
        </label>
      </div>

      <div className="flex flex-wrap gap-3">
        <label className="flex flex-col gap-1">
          <span className="text-[length:var(--text-pill)] text-(--text-muted)">Starting weight (kg)</span>
          <input type="number" inputMode="decimal" step={0.1} min={0} value={weightStart} onChange={(e) => setWeightStart(e.target.value)} placeholder="optional" className={`${field} w-32`} />
        </label>
        <label className="flex flex-col gap-1">
          <span className="text-[length:var(--text-pill)] text-(--text-muted)">Target weight (kg)</span>
          <input type="number" inputMode="decimal" step={0.1} min={0} value={weightTarget} onChange={(e) => setWeightTarget(e.target.value)} placeholder="optional" className={`${field} w-32`} />
        </label>
        <label className="flex flex-col gap-1">
          <span className="text-[length:var(--text-pill)] text-(--text-muted)">Protein target (g/day)</span>
          <input type="number" inputMode="decimal" step={1} min={0} value={proteinTarget} onChange={(e) => setProteinTarget(e.target.value)} placeholder="optional" className={`${field} w-32`} />
        </label>
      </div>

      <div className="flex items-center justify-end gap-1.5">
        <button type="button" onClick={() => setOpen(false)} className="rounded-md px-2.5 py-1 text-[length:var(--text-pill)] text-(--text-muted) hover:text-(--fg) hover:bg-(--card-alt) transition-colors cursor-pointer">
          Cancel
        </button>
        <button type="submit" className="rounded-md bg-(--fg) text-(--bg) px-2.5 py-1 text-[length:var(--text-pill)] cursor-pointer">
          Start cycle
        </button>
      </div>
    </form>
  );
}
