import { useState, type FormEvent } from "react";
import type { ExerciseInput } from "./types";

interface AddExerciseFormProps {
  onAdd: (input: ExerciseInput) => void;
}

export function AddExerciseForm({ onAdd }: AddExerciseFormProps) {
  const [name, setName] = useState("");
  const [unit, setUnit] = useState("");
  const [goalQuantity, setGoalQuantity] = useState("");
  const [goalDate, setGoalDate] = useState("");

  function submit(e: FormEvent) {
    e.preventDefault();
    const trimmed = name.trim();
    if (!trimmed) return;
    const qty = Number(goalQuantity);
    onAdd({
      name: trimmed,
      unit: unit.trim() || null,
      goalQuantity: goalQuantity.trim() !== "" && Number.isFinite(qty) && qty > 0 ? qty : null,
      goalDate: goalDate || null,
    });
    setName("");
    setUnit("");
    setGoalQuantity("");
    setGoalDate("");
  }

  const field =
    "rounded-lg bg-(--card-alt) border-(--line) border-[0.5px] border-solid px-2.5 py-1.5 text-[length:var(--text-pill)] text-(--fg) placeholder:text-(--text-faint) focus:outline-none";

  return (
    <form onSubmit={submit} className="bg-(--card) border-(--line) border-[0.5px] border-solid rounded-xl p-4 flex flex-wrap items-end gap-3">
      <label className="flex flex-col gap-1 flex-1 min-w-40">
        <span className="text-[length:var(--text-pill)] text-(--text-muted)">Exercise</span>
        <input autoFocus value={name} onChange={(e) => setName(e.target.value)} placeholder="e.g. Pull-ups" className={`${field} text-[length:var(--text-caption)]`} />
      </label>
      <label className="flex flex-col gap-1">
        <span className="text-[length:var(--text-pill)] text-(--text-muted)">Unit</span>
        <input value={unit} onChange={(e) => setUnit(e.target.value)} placeholder="reps / km" className={`${field} w-24`} />
      </label>
      <label className="flex flex-col gap-1">
        <span className="text-[length:var(--text-pill)] text-(--text-muted)">Goal qty</span>
        <input type="number" step={1} min={0} value={goalQuantity} onChange={(e) => setGoalQuantity(e.target.value)} placeholder="optional" className={`${field} w-24`} />
      </label>
      <label className="flex flex-col gap-1">
        <span className="text-[length:var(--text-pill)] text-(--text-muted)">Goal date</span>
        <input type="date" value={goalDate} onChange={(e) => setGoalDate(e.target.value)} className={field} />
      </label>
      <button type="submit" className="rounded-md bg-(--fg) text-(--bg) px-3 py-1.5 text-[length:var(--text-pill)] cursor-pointer">
        Add
      </button>
    </form>
  );
}
