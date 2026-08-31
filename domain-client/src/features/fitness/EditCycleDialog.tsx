import { useState, type KeyboardEvent } from "react";
import type { Cycle, CycleInput } from "./types";

interface EditCycleDialogProps {
  cycle: Cycle;
  onClose: () => void;
  onSave: (input: CycleInput) => void;
}

function numOrNull(v: string): number | null {
  const n = Number(v);
  return v.trim() !== "" && Number.isFinite(n) && n > 0 ? n : null;
}

export function EditCycleDialog({ cycle, onClose, onSave }: EditCycleDialogProps) {
  const [name, setName] = useState(cycle.name);
  const [startDate, setStartDate] = useState(cycle.startDate);
  const [weightStart, setWeightStart] = useState(cycle.weightStart?.toString() ?? "");
  const [weightTarget, setWeightTarget] = useState(cycle.weightTarget?.toString() ?? "");
  const [proteinTarget, setProteinTarget] = useState(cycle.proteinTarget?.toString() ?? "");

  function handleSave() {
    const trimmed = name.trim();
    if (!trimmed) return;
    onSave({
      name: trimmed,
      startDate,
      weightStart: numOrNull(weightStart),
      weightTarget: numOrNull(weightTarget),
      proteinTarget: numOrNull(proteinTarget),
    });
  }

  function handleKeyDown(e: KeyboardEvent<HTMLDivElement>) {
    if (e.key === "Escape") onClose();
  }

  const field =
    "w-full rounded-lg bg-(--card-alt) border-(--line) border-[0.5px] border-solid px-2.5 py-1.5 text-[length:var(--text-caption)] text-(--fg) focus:outline-none";

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 px-4"
      onClick={onClose}
      onKeyDown={handleKeyDown}
      role="presentation"
    >
      <div
        className="w-full max-w-sm max-h-[90vh] overflow-y-auto themed-scrollbar rounded-xl border-(--line) border-[0.5px] border-solid bg-(--card) p-4 shadow-lg"
        onClick={(e) => e.stopPropagation()}
        role="dialog"
        aria-modal="true"
        aria-label="Edit cycle"
      >
        <h2 className="text-[length:var(--text-caption)] text-(--fg) mb-3">Edit cycle</h2>

        <div className="flex flex-col gap-3">
          <label className="flex flex-col gap-1">
            <span className="text-[length:var(--text-pill)] text-(--text-muted)">Name</span>
            <input autoFocus value={name} onChange={(e) => setName(e.target.value)} className={field} />
          </label>
          <label className="flex flex-col gap-1">
            <span className="text-[length:var(--text-pill)] text-(--text-muted)">Start date</span>
            <input type="date" value={startDate} onChange={(e) => setStartDate(e.target.value)} className={field} />
          </label>
          <div className="flex gap-3">
            <label className="flex flex-col gap-1 flex-1">
              <span className="text-[length:var(--text-pill)] text-(--text-muted)">Start wt (kg)</span>
              <input type="number" step={0.1} min={0} value={weightStart} onChange={(e) => setWeightStart(e.target.value)} className={field} />
            </label>
            <label className="flex flex-col gap-1 flex-1">
              <span className="text-[length:var(--text-pill)] text-(--text-muted)">Target wt (kg)</span>
              <input type="number" step={0.1} min={0} value={weightTarget} onChange={(e) => setWeightTarget(e.target.value)} className={field} />
            </label>
          </div>
          <label className="flex flex-col gap-1">
            <span className="text-[length:var(--text-pill)] text-(--text-muted)">Protein target (g/day)</span>
            <input type="number" step={1} min={0} value={proteinTarget} onChange={(e) => setProteinTarget(e.target.value)} className={field} />
          </label>
        </div>

        <div className="flex items-center justify-end gap-1.5 mt-4">
          <button
            type="button"
            onClick={onClose}
            className="rounded-md px-2.5 py-1 text-[length:var(--text-pill)] text-(--text-muted) hover:text-(--fg) hover:bg-(--card-alt) transition-colors cursor-pointer"
          >
            Cancel
          </button>
          <button
            type="button"
            onClick={handleSave}
            className="rounded-md bg-(--fg) text-(--bg) px-2.5 py-1 text-[length:var(--text-pill)] cursor-pointer"
          >
            Save
          </button>
        </div>
      </div>
    </div>
  );
}
