import { useState, type KeyboardEvent } from "react";
import type { Exercise, ExerciseInput } from "./types";

interface EditExerciseDialogProps {
  exercise: Exercise;
  onClose: () => void;
  onSave: (input: ExerciseInput) => void;
}

export function EditExerciseDialog({ exercise, onClose, onSave }: EditExerciseDialogProps) {
  const [name, setName] = useState(exercise.name);
  const [unit, setUnit] = useState(exercise.unit ?? "");
  const [goalQuantity, setGoalQuantity] = useState(exercise.goalQuantity?.toString() ?? "");
  const [goalDate, setGoalDate] = useState(exercise.goalDate ?? "");

  function handleSave() {
    const trimmed = name.trim();
    if (!trimmed) return;
    const qty = Number(goalQuantity);
    onSave({
      name: trimmed,
      unit: unit.trim() || null,
      goalQuantity: goalQuantity.trim() !== "" && Number.isFinite(qty) && qty > 0 ? qty : null,
      goalDate: goalDate || null,
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
        className="w-full max-w-sm rounded-xl border-(--line) border-[0.5px] border-solid bg-(--card) p-4 shadow-lg"
        onClick={(e) => e.stopPropagation()}
        role="dialog"
        aria-modal="true"
        aria-label="Edit exercise"
      >
        <h2 className="text-[length:var(--text-caption)] text-(--fg) mb-3">Edit exercise</h2>

        <div className="flex flex-col gap-3">
          <label className="flex flex-col gap-1">
            <span className="text-[length:var(--text-pill)] text-(--text-muted)">Name</span>
            <input autoFocus value={name} onChange={(e) => setName(e.target.value)} className={field} />
          </label>
          <div className="flex gap-3">
            <label className="flex flex-col gap-1 flex-1">
              <span className="text-[length:var(--text-pill)] text-(--text-muted)">Unit</span>
              <input value={unit} onChange={(e) => setUnit(e.target.value)} className={field} />
            </label>
            <label className="flex flex-col gap-1 flex-1">
              <span className="text-[length:var(--text-pill)] text-(--text-muted)">Goal qty</span>
              <input type="number" step={1} min={0} value={goalQuantity} onChange={(e) => setGoalQuantity(e.target.value)} className={field} />
            </label>
          </div>
          <label className="flex flex-col gap-1">
            <span className="text-[length:var(--text-pill)] text-(--text-muted)">Goal date</span>
            <input type="date" value={goalDate} onChange={(e) => setGoalDate(e.target.value)} className={field} />
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
