import { useEffect, useState } from "react";
import { createExercise, deleteExercise, fetchExercise, fetchExercises, updateExercise, upsertExerciseLog } from "./api";
import { AddExerciseForm } from "./AddExerciseForm";
import { ExerciseCard } from "./ExerciseCard";
import { todayISTKey } from "./dateUtils";
import type { Exercise, ExerciseInput } from "./types";

interface ExerciseTabProps {
  cycleId: string;
  onError: (message: string) => void;
}

export function ExerciseTab({ cycleId, onError }: ExerciseTabProps) {
  const [exercises, setExercises] = useState<Exercise[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    setLoading(true);
    fetchExercises(cycleId)
      .then(setExercises)
      .catch((err) => onError(err instanceof Error ? err.message : "Couldn't load exercises."))
      .finally(() => setLoading(false));
  }, [cycleId, onError]);

  function handleAdd(input: ExerciseInput) {
    createExercise(cycleId, input)
      .then((exercise) => setExercises((prev) => [exercise, ...prev]))
      .catch((err) => onError(err instanceof Error ? err.message : "Couldn't add exercise."));
  }

  function handleUpdate(id: string, input: ExerciseInput) {
    const original = exercises;
    setExercises((prev) => prev.map((e) => (e.id === id ? { ...e, ...input } : e)));
    updateExercise(id, input)
      .then((updated) => setExercises((prev) => prev.map((e) => (e.id === id ? updated : e))))
      .catch((err) => {
        setExercises(original);
        onError(err instanceof Error ? err.message : "Couldn't update exercise.");
      });
  }

  function handleDelete(id: string) {
    const original = exercises;
    setExercises((prev) => prev.filter((e) => e.id !== id));
    deleteExercise(id).catch((err) => {
      setExercises(original);
      onError(err instanceof Error ? err.message : "Couldn't delete exercise.");
    });
  }

  // Quick-log upserts today's entry, so the new totalLogged isn't a simple
  // add — re-fetch the one exercise for an accurate figure.
  function handleQuickLog(id: string, quantity: number) {
    upsertExerciseLog(id, todayISTKey(), quantity)
      .then(() => fetchExercise(id))
      .then((updated) => setExercises((prev) => prev.map((e) => (e.id === id ? updated : e))))
      .catch((err) => onError(err instanceof Error ? err.message : "Couldn't log today."));
  }

  return (
    <div>
      <AddExerciseForm onAdd={handleAdd} />

      {loading ? (
        <div className="bg-(--card) border-(--line) border-[0.5px] border-solid rounded-xl p-8 text-center text-(--text-faint) text-[length:var(--text-caption)]">
          Loading exercises...
        </div>
      ) : exercises.length === 0 ? (
        <div className="bg-(--card) border-(--line) border-[0.5px] border-solid rounded-xl p-8 text-center text-(--text-faint) text-[length:var(--text-caption)]">
          No exercises yet. Add one above, then open it to log your daily reps.
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-1.5 pb-2">
          {exercises.map((exercise) => (
            <ExerciseCard
              key={exercise.id}
              exercise={exercise}
              onUpdate={handleUpdate}
              onDelete={handleDelete}
              onQuickLog={handleQuickLog}
            />
          ))}
        </div>
      )}
    </div>
  );
}
