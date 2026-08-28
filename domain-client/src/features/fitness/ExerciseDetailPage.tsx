import { useEffect, useState } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import { ArrowLeft, Trash2, X } from "lucide-react";
import { ApiError } from "../../lib/apiClient";
import { deleteExerciseLog, fetchExercise, fetchExerciseLogs, upsertExerciseLog } from "./api";
import { GoalDonut } from "./GoalDonut";
import { DailyBarChart } from "./DailyBarChart";
import { LogValueForm } from "./LogValueForm";
import { daysUntil, formatDayKey } from "./dateUtils";
import type { Exercise, ExerciseLog } from "./types";

export function ExerciseDetailPage() {
  const { exerciseId } = useParams<{ exerciseId: string }>();
  const navigate = useNavigate();
  const [exercise, setExercise] = useState<Exercise | null>(null);
  const [logs, setLogs] = useState<ExerciseLog[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!exerciseId) return;
    setLoading(true);
    Promise.all([fetchExercise(exerciseId), fetchExerciseLogs(exerciseId)])
      .then(([ex, ls]) => {
        setExercise(ex);
        setLogs(ls);
      })
      .catch((err) => {
        if (err instanceof ApiError && err.status === 404) {
          navigate("/fitness", { replace: true });
          return;
        }
        setError(err instanceof Error ? err.message : "Couldn't load this exercise.");
      })
      .finally(() => setLoading(false));
  }, [exerciseId, navigate]);

  function recalcTotal(next: ExerciseLog[]) {
    setExercise((prev) => (prev ? { ...prev, totalLogged: next.reduce((s, l) => s + l.quantity, 0) } : prev));
  }

  function handleLog(date: string, quantity: number) {
    if (!exerciseId) return;
    upsertExerciseLog(exerciseId, date, quantity)
      .then((saved) => {
        setLogs((prev) => {
          const next = [...prev.filter((l) => l.logDate !== saved.logDate), saved].sort((a, b) =>
            a.logDate.localeCompare(b.logDate),
          );
          recalcTotal(next);
          return next;
        });
      })
      .catch((err) => setError(err instanceof Error ? err.message : "Couldn't save log."));
  }

  function handleDelete(id: string) {
    const original = logs;
    const next = logs.filter((l) => l.id !== id);
    setLogs(next);
    recalcTotal(next);
    deleteExerciseLog(id).catch((err) => {
      setLogs(original);
      recalcTotal(original);
      setError(err instanceof Error ? err.message : "Couldn't delete log.");
    });
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

  return (
    <div className="h-full flex flex-col">
      <div className="shrink-0">
        <Link
          to={`/fitness/${exercise.cycleId}`}
          className="inline-flex items-center gap-1.5 mb-3 text-[length:var(--text-pill)] text-(--text-muted) hover:text-(--fg) transition-colors"
        >
          <ArrowLeft size={13} />
          Back to cycle
        </Link>

        <div className="mb-4">
          <h1 className="font-space font-semibold text-(--fg) text-[length:var(--text-caption)]">{exercise.name}</h1>
          {exercise.goalDate && (
            <p className="text-[length:var(--text-pill)] text-(--text-faint) mt-0.5">
              Goal date {exercise.goalDate}
              {days != null && ` · ${days >= 0 ? `${days} days to go` : `${-days} days past`}`}
            </p>
          )}
        </div>

        <div className="bg-(--card) border-(--line) border-[0.5px] border-solid rounded-xl px-4 py-3 mb-4">
          <GoalDonut current={exercise.totalLogged} goal={exercise.goalQuantity} unit={exercise.unit} />
        </div>

        <LogValueForm
          valueLabel="Quantity"
          unit={exercise.unit}
          step={1}
          submitLabel="Log day"
          onSubmit={handleLog}
        />

        {error && (
          <div className="mb-3 flex items-center justify-between gap-2 rounded-lg border-(--line) border-[0.5px] border-solid bg-(--card) px-3 py-2 text-[length:var(--text-pill)] text-red-400">
            <span>{error}</span>
            <button onClick={() => setError(null)} aria-label="Dismiss error" className="shrink-0 text-(--text-faint) hover:text-(--fg) transition-colors cursor-pointer">
              <X size={12} />
            </button>
          </div>
        )}
      </div>

      <div className="flex-1 min-h-0 overflow-y-auto themed-scrollbar flex flex-col gap-4 pb-2">
        <DailyBarChart
          title="Daily performance"
          points={logs.map((l) => ({ date: l.logDate, value: l.quantity }))}
          unit={exercise.unit ?? undefined}
          emptyLabel="No days logged yet."
        />

        <div className="bg-(--card) border-(--line) border-[0.5px] border-solid rounded-xl p-4">
          <h2 className="text-[length:var(--text-pill)] font-medium uppercase tracking-wide text-(--text-faint) mb-3">
            Logged days
          </h2>
          {logs.length === 0 ? (
            <p className="text-[length:var(--text-pill)] text-(--text-faint) py-4 text-center">Nothing logged yet.</p>
          ) : (
            <div className="flex flex-col gap-1 max-h-72 overflow-y-auto themed-scrollbar">
              {[...logs].reverse().map((l) => (
                <div key={l.id} className="group flex items-center justify-between gap-2 rounded-lg px-2 py-1.5 hover:bg-(--card-alt)">
                  <span className="text-[length:var(--text-pill)] text-(--text-muted)">{formatDayKey(l.logDate)}</span>
                  <div className="flex items-center gap-2">
                    <span className="text-[length:var(--text-pill)] text-(--fg)">
                      {round1(l.quantity)}
                      {exercise.unit ? ` ${exercise.unit}` : ""}
                    </span>
                    <button
                      onClick={() => handleDelete(l.id)}
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
  );
}

function round1(n: number): number {
  return Math.round(n * 10) / 10;
}
