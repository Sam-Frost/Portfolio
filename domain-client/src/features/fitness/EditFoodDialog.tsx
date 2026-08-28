import { useState, type KeyboardEvent } from "react";
import type { Food } from "./types";

interface EditFoodDialogProps {
  food: Food;
  onClose: () => void;
  onSave: (input: { name: string; unit: string; proteinPerUnit: number }) => void;
}

export function EditFoodDialog({ food, onClose, onSave }: EditFoodDialogProps) {
  const [name, setName] = useState(food.name);
  const [unit, setUnit] = useState(food.unit);
  const [proteinPerUnit, setProteinPerUnit] = useState(food.proteinPerUnit.toString());

  function handleSave() {
    const n = Number(proteinPerUnit);
    if (!name.trim() || !unit.trim() || !Number.isFinite(n) || n <= 0) return;
    onSave({ name: name.trim(), unit: unit.trim(), proteinPerUnit: n });
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
        aria-label="Edit food"
      >
        <h2 className="text-[length:var(--text-caption)] text-(--fg) mb-3">Edit food</h2>

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
              <span className="text-[length:var(--text-pill)] text-(--text-muted)">Protein / unit (g)</span>
              <input type="number" step={0.1} min={0} value={proteinPerUnit} onChange={(e) => setProteinPerUnit(e.target.value)} className={field} />
            </label>
          </div>
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
