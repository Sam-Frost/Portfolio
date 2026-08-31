import { useEffect, useState } from "react";
import { Plus, X } from "lucide-react";
import { createExercise, fetchExercises } from "./api";
import { AddExerciseForm } from "./AddExerciseForm";
import { ExerciseDetail } from "./ExerciseDetail";
import type { Exercise, ExerciseInput } from "./types";

interface ExerciseTabProps {
  cycleId: string;
  onError: (message: string) => void;
}

// One exercise at a time: pick it from the dropdown and everything about it
// (goal progress, day logger, chart, history, edit/delete) renders below in
// <ExerciseDetail>. The add form is tucked behind a toggle so the whole tab
// fits on one screen without scrolling.
export function ExerciseTab({ cycleId, onError }: ExerciseTabProps) {
  const [exercises, setExercises] = useState<Exercise[]>([]);
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [adding, setAdding] = useState(false);

  useEffect(() => {
    setLoading(true);
    fetchExercises(cycleId)
      .then((list) => {
        setExercises(list);
        setSelectedId((prev) => (prev && list.some((e) => e.id === prev) ? prev : (list[0]?.id ?? null)));
      })
      .catch((err) => onError(err instanceof Error ? err.message : "Couldn't load exercises."))
      .finally(() => setLoading(false));
  }, [cycleId, onError]);

  function handleAdd(input: ExerciseInput) {
    createExercise(cycleId, input)
      .then((exercise) => {
        setExercises((prev) => [exercise, ...prev]);
        setSelectedId(exercise.id);
        setAdding(false);
      })
      .catch((err) => onError(err instanceof Error ? err.message : "Couldn't add exercise."));
  }

  function handleExerciseChange(updated: Exercise) {
    setExercises((prev) => prev.map((e) => (e.id === updated.id ? updated : e)));
  }

  function handleExerciseDelete(id: string) {
    const goneIndex = exercises.findIndex((e) => e.id === id);
    const next = exercises.filter((e) => e.id !== id);
    setExercises(next);
    if (selectedId === id) {
      setSelectedId(next[goneIndex]?.id ?? next[goneIndex - 1]?.id ?? next[0]?.id ?? null);
    }
  }

  const selectClass =
    "w-full rounded-lg bg-(--card-alt) border-(--line) border-[0.5px] border-solid px-2.5 py-2 text-[length:var(--text-caption)] text-(--fg) focus:outline-none";

  if (loading) {
    return (
      <div className="bg-(--card) border-(--line) border-[0.5px] border-solid rounded-xl p-8 text-center text-(--text-faint) text-[length:var(--text-caption)]">
        Loading exercises...
      </div>
    );
  }

  if (exercises.length === 0) {
    return (
      <div className="flex flex-col gap-3">
        <AddExerciseForm onAdd={handleAdd} />
        <div className="bg-(--card) border-(--line) border-[0.5px] border-solid rounded-xl p-8 text-center text-(--text-faint) text-[length:var(--text-caption)]">
          No exercises yet. Add one above to start logging your daily reps.
        </div>
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-3">
      <div className="flex items-end gap-2">
        <label className="flex flex-col gap-1 flex-1 min-w-0">
          <span className="text-[length:var(--text-pill)] text-(--text-muted)">Exercise</span>
          <select
            value={selectedId ?? ""}
            onChange={(e) => setSelectedId(e.target.value)}
            className={selectClass}
          >
            {exercises.map((e) => (
              <option key={e.id} value={e.id}>
                {e.name}
                {e.goalQuantity != null && e.goalQuantity > 0
                  ? ` — ${round1(e.totalLogged)} / ${round1(e.goalQuantity)}${e.unit ? ` ${e.unit}` : ""}`
                  : ""}
              </option>
            ))}
          </select>
        </label>
        <button
          type="button"
          onClick={() => setAdding((v) => !v)}
          className="shrink-0 inline-flex items-center gap-1.5 rounded-lg border-(--line) border-[0.5px] border-solid px-3 py-2 text-[length:var(--text-pill)] text-(--text-muted) hover:text-(--fg) hover:bg-(--card-alt) transition-colors cursor-pointer"
        >
          {adding ? <X size={13} /> : <Plus size={13} />}
          {adding ? "Close" : "Add exercise"}
        </button>
      </div>

      {adding && <AddExerciseForm onAdd={handleAdd} />}

      {selectedId && (
        <ExerciseDetail
          key={selectedId}
          exerciseId={selectedId}
          onError={onError}
          onExerciseChange={handleExerciseChange}
          onExerciseDelete={handleExerciseDelete}
        />
      )}
    </div>
  );
}

function round1(n: number): number {
  return Math.round(n * 10) / 10;
}
