import { useEffect, useState } from "react";
import { Pencil, Trash2 } from "lucide-react";
import { ApiError } from "../../lib/apiClient";
import {
  deleteExercise,
  deleteExerciseLog,
  fetchExercise,
  fetchExerciseLogs,
  updateExercise,
  upsertExerciseLog,
} from "./api";
import { ConfirmDeleteDialog } from "./ConfirmDeleteDialog";
import { EditExerciseDialog } from "./EditExerciseDialog";
import { GoalDonut } from "./GoalDonut";
import { DailyBarChart } from "./DailyBarChart";
import { LogValueForm } from "./LogValueForm";
import { daysUntil, formatDayKey } from "./dateUtils";
import type { Exercise, ExerciseInput, ExerciseLog } from "./types";

interface ExerciseDetailProps {
  exerciseId: string;
  onError: (message: string) => void;
  // Fired whenever this exercise's stored fields or running total change, so a
  // parent list/dropdown can stay in sync without refetching.
  onExerciseChange?: (exercise: Exercise) => void;
  // Fired after the exercise itself is deleted.
  onExerciseDelete?: (id: string) => void;
  // Fired once the exercise is first loaded (used by the standalone route page
  // to resolve its "back to cycle" link).
  onLoaded?: (exercise: Exercise) => void;
  // Called for a hard 404 (deleted elsewhere) — the standalone page navigates
  // away; the in-tab view drops the selection.
  onNotFound?: () => void;
}

// Everything about a single exercise: goal progress, a day logger, the daily
// chart and the editable log history — plus edit/delete of the exercise. Shared
// by the Exercise tab (picked from a dropdown) and the standalone detail route.
export function ExerciseDetail({
  exerciseId,
  onError,
  onExerciseChange,
  onExerciseDelete,
  onLoaded,
  onNotFound,
}: ExerciseDetailProps) {
  const [exercise, setExercise] = useState<Exercise | null>(null);
  const [logs, setLogs] = useState<ExerciseLog[]>([]);
  const [loading, setLoading] = useState(true);
  const [editing, setEditing] = useState(false);
  const [confirmingDelete, setConfirmingDelete] = useState(false);

  useEffect(() => {
    let active = true;
    setLoading(true);
    setEditing(false);
    setConfirmingDelete(false);
    Promise.all([fetchExercise(exerciseId), fetchExerciseLogs(exerciseId)])
      .then(([ex, ls]) => {
        if (!active) return;
        setExercise(ex);
        setLogs(ls);
        onLoaded?.(ex);
      })
      .catch((err) => {
        if (!active) return;
        if (err instanceof ApiError && err.status === 404) {
          onNotFound?.();
          return;
        }
        onError(err instanceof Error ? err.message : "Couldn't load this exercise.");
      })
      .finally(() => {
        if (active) setLoading(false);
      });
    return () => {
      active = false;
    };
  }, [exerciseId, onError, onLoaded, onNotFound]);

  // Keep logs + the exercise's running total in step, and let the parent know.
  function applyLogs(next: ExerciseLog[]) {
    setLogs(next);
    if (!exercise) return;
    const updated = { ...exercise, totalLogged: next.reduce((s, l) => s + l.quantity, 0) };
    setExercise(updated);
    onExerciseChange?.(updated);
  }

  function handleLog(date: string, quantity: number) {
    upsertExerciseLog(exerciseId, date, quantity)
      .then((saved) => {
        applyLogs(
          [...logs.filter((l) => l.logDate !== saved.logDate), saved].sort((a, b) =>
            a.logDate.localeCompare(b.logDate),
          ),
        );
      })
      .catch((err) => onError(err instanceof Error ? err.message : "Couldn't save log."));
  }

  function handleDeleteLog(id: string) {
    const original = logs;
    applyLogs(logs.filter((l) => l.id !== id));
    deleteExerciseLog(id).catch((err) => {
      applyLogs(original);
      onError(err instanceof Error ? err.message : "Couldn't delete log.");
    });
  }

  function handleEdit(input: ExerciseInput) {
    if (!exercise) return;
    const original = exercise;
    const optimistic = { ...exercise, ...input };
    setExercise(optimistic);
    onExerciseChange?.(optimistic);
    setEditing(false);
    updateExercise(exercise.id, input)
      .then((updated) => {
        setExercise(updated);
        onExerciseChange?.(updated);
      })
      .catch((err) => {
        setExercise(original);
        onExerciseChange?.(original);
        onError(err instanceof Error ? err.message : "Couldn't update exercise.");
      });
  }

  function handleDeleteExercise() {
    if (!exercise) return;
    setConfirmingDelete(false);
    deleteExercise(exercise.id)
      .then(() => onExerciseDelete?.(exercise.id))
      .catch((err) => onError(err instanceof Error ? err.message : "Couldn't delete exercise."));
  }

  if (loading) {
    return (
      <div className="bg-(--card) border-(--line) border-[0.5px] border-solid rounded-xl p-8 text-center text-(--text-faint) text-[length:var(--text-caption)]">
        Loading...
      </div>
    );
  }
  if (!exercise) return null;

  const days = exercise.goalDate ? daysUntil(exercise.goalDate) : null;

  const goalNote =
    days == null ? null : days >= 0 ? `${days}d to goal` : `${-days}d past goal`;

  return (
    <div className="flex flex-col gap-3">
      <div className="flex items-center justify-between gap-3">
        <div className="flex items-baseline gap-2 min-w-0">
          <h2 className="font-space font-semibold text-(--fg) text-[length:var(--text-caption)] truncate">
            {exercise.name}
          </h2>
          {exercise.goalDate && (
            <span className="shrink-0 text-[length:var(--text-pill)] text-(--text-faint)">
              {exercise.goalDate}
              {goalNote && ` · ${goalNote}`}
            </span>
          )}
        </div>
        <div className="flex items-center gap-0.5 shrink-0">
          <button
            onClick={() => setEditing(true)}
            aria-label="Edit exercise"
            className="flex items-center justify-center size-7 rounded-md text-(--text-faint) hover:text-(--fg) hover:bg-(--card-alt) transition-colors cursor-pointer"
          >
            <Pencil size={13} />
          </button>
          <button
            onClick={() => setConfirmingDelete(true)}
            aria-label="Delete exercise"
            className="flex items-center justify-center size-7 rounded-md text-(--text-faint) hover:text-(--fg) hover:bg-(--card-alt) transition-colors cursor-pointer"
          >
            <Trash2 size={13} />
          </button>
        </div>
      </div>

      {/* Two columns so goal + logging + chart + history all sit on one screen
          without the tab needing to scroll (collapses to one column < lg). */}
      <div className="grid gap-3 lg:grid-cols-2 lg:items-start">
        <div className="flex flex-col gap-3">
          <div className="bg-(--card) border-(--line) border-[0.5px] border-solid rounded-xl px-4 py-3">
            <GoalDonut current={exercise.totalLogged} goal={exercise.goalQuantity} unit={exercise.unit} />
          </div>

          <LogValueForm
            valueLabel="Quantity"
            unit={exercise.unit}
            step={1}
            submitLabel="Log day"
            onSubmit={handleLog}
          />
        </div>

        <div className="flex flex-col gap-3">
          <DailyBarChart
            title="Daily performance"
            points={logs.map((l) => ({ date: l.logDate, value: l.quantity }))}
            unit={exercise.unit ?? undefined}
            emptyLabel="No days logged yet."
          />

          <div className="bg-(--card) border-(--line) border-[0.5px] border-solid rounded-xl p-4">
            <h2 className="text-[length:var(--text-pill)] font-medium uppercase tracking-wide text-(--text-faint) mb-2">
              Logged days
            </h2>
            {logs.length === 0 ? (
              <p className="text-[length:var(--text-pill)] text-(--text-faint) py-3 text-center">Nothing logged yet.</p>
            ) : (
              <div className="flex flex-col gap-1 max-h-44 overflow-y-auto themed-scrollbar">
                {[...logs].reverse().map((l) => (
                  <div key={l.id} className="group flex items-center justify-between gap-2 rounded-lg px-2 py-1 hover:bg-(--card-alt)">
                    <span className="text-[length:var(--text-pill)] text-(--text-muted)">{formatDayKey(l.logDate)}</span>
                    <div className="flex items-center gap-2">
                      <span className="text-[length:var(--text-pill)] text-(--fg)">
                        {round1(l.quantity)}
                        {exercise.unit ? ` ${exercise.unit}` : ""}
                      </span>
                      <button
                        onClick={() => handleDeleteLog(l.id)}
                        aria-label="Delete log"
                        className="flex items-center justify-center size-6 rounded-md text-(--text-faint) opacity-0 group-hover:opacity-100 hover:text-(--fg) transition-opacity cursor-pointer"
                      >
                        <Trash2 size={12} />
                      </button>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>
      </div>

      {editing && (
        <EditExerciseDialog exercise={exercise} onClose={() => setEditing(false)} onSave={handleEdit} />
      )}

      {confirmingDelete && (
        <ConfirmDeleteDialog
          title="Delete exercise?"
          message={`"${exercise.name}" and all its daily logs will be deleted.`}
          onCancel={() => setConfirmingDelete(false)}
          onConfirm={handleDeleteExercise}
        />
      )}
    </div>
  );
}

function round1(n: number): number {
  return Math.round(n * 10) / 10;
}
