import { useState, type FormEvent } from "react";
import { useNavigate } from "react-router-dom";
import { Check, Plus, Trash2, X } from "lucide-react";
import { ConfirmDeleteDialog } from "./ConfirmDeleteDialog";
import { EditExerciseDialog } from "./EditExerciseDialog";
import { daysUntil } from "./dateUtils";
import type { Exercise, ExerciseInput } from "./types";

interface ExerciseCardProps {
  exercise: Exercise;
  onUpdate: (id: string, input: ExerciseInput) => void;
  onDelete: (id: string) => void;
  // Upserts today's log for this exercise (see ExerciseTab.handleQuickLog).
  onQuickLog: (id: string, quantity: number) => void;
}

export function ExerciseCard({ exercise, onUpdate, onDelete, onQuickLog }: ExerciseCardProps) {
  const navigate = useNavigate();
  const [editing, setEditing] = useState(false);
  const [confirming, setConfirming] = useState(false);
  const [quickOpen, setQuickOpen] = useState(false);
  const [quickValue, setQuickValue] = useState("");

  const hasGoal = exercise.goalQuantity != null && exercise.goalQuantity > 0;
  const progress = hasGoal ? Math.min(exercise.totalLogged / (exercise.goalQuantity as number), 1) : 0;
  const pct = hasGoal ? Math.round(progress * 100) : null;
  const days = exercise.goalDate ? daysUntil(exercise.goalDate) : null;
  const overdue = days != null && days < 0 && progress < 1;

  function submitQuick(e: FormEvent) {
    e.preventDefault();
    e.stopPropagation();
    const n = Number(quickValue);
    if (!Number.isFinite(n) || n <= 0) return;
    onQuickLog(exercise.id, n);
    setQuickValue("");
    setQuickOpen(false);
  }

  return (
    <div
      onClick={() => navigate(`/fitness/exercises/${exercise.id}`)}
      onDoubleClick={(e) => {
        e.stopPropagation();
        setEditing(true);
      }}
      role="button"
      tabIndex={0}
      className="group relative flex flex-col gap-2 bg-(--card) border-(--line) border-[0.5px] border-solid rounded-xl px-4 py-3 cursor-pointer hover:border-(--line-strong) transition-colors"
    >
      <div className="flex items-start justify-between gap-2">
        <span className="min-w-0 flex-1 truncate text-[length:var(--text-caption)] text-(--fg) font-medium">{exercise.name}</span>
        <div className="flex items-center gap-0.5 shrink-0">
          <button
            onClick={(e) => {
              e.stopPropagation();
              setQuickOpen((v) => !v);
            }}
            aria-label="Log today"
            className="flex items-center justify-center size-6 rounded-md text-(--text-faint) opacity-0 group-hover:opacity-100 hover:text-(--fg) hover:bg-(--card-alt) transition-opacity cursor-pointer"
          >
            <Plus size={13} />
          </button>
          <button
            onClick={(e) => {
              e.stopPropagation();
              setConfirming(true);
            }}
            aria-label="Delete exercise"
            className="flex items-center justify-center size-6 rounded-md text-(--text-faint) opacity-0 group-hover:opacity-100 hover:text-(--fg) hover:bg-(--card-alt) transition-opacity cursor-pointer"
          >
            <Trash2 size={13} />
          </button>
        </div>
      </div>

      <div className="text-[length:var(--text-pill)] text-(--text-faint)">
        {hasGoal
          ? `${round1(exercise.totalLogged)} / ${round1(exercise.goalQuantity as number)}${exercise.unit ? ` ${exercise.unit}` : ""}`
          : `${round1(exercise.totalLogged)}${exercise.unit ? ` ${exercise.unit}` : ""} logged`}
        {days != null && (
          <span className={overdue ? "text-red-400" : undefined}>
            {" · "}
            {days > 0 ? `${days}d to goal` : days === 0 ? "goal is today" : `${-days}d overdue`}
          </span>
        )}
      </div>

      {hasGoal && (
        <div className="flex items-center gap-2 mt-1">
          <div className="flex-1 h-1.5 rounded-full bg-(--ring-track) overflow-hidden">
            <div className="h-full rounded-full bg-(--green)" style={{ width: `${progress * 100}%` }} />
          </div>
          <span className="shrink-0 text-[length:var(--text-pill)] text-(--text-faint)">{pct}%</span>
        </div>
      )}

      {quickOpen && (
        <form
          onClick={(e) => e.stopPropagation()}
          onSubmit={submitQuick}
          className="mt-1 flex items-center gap-1.5"
        >
          <input
            autoFocus
            type="number"
            inputMode="decimal"
            step={1}
            min={0}
            value={quickValue}
            onChange={(e) => setQuickValue(e.target.value)}
            placeholder={`today's ${exercise.unit ?? "count"}`}
            className="flex-1 min-w-0 rounded-md bg-(--card-alt) border-(--line) border-[0.5px] border-solid px-2 py-1 text-[length:var(--text-pill)] text-(--fg) placeholder:text-(--text-faint) focus:outline-none"
          />
          <button type="submit" aria-label="Save" className="flex items-center justify-center size-6 rounded-md bg-(--fg) text-(--bg) cursor-pointer">
            <Check size={12} />
          </button>
          <button
            type="button"
            aria-label="Cancel"
            onClick={() => {
              setQuickOpen(false);
              setQuickValue("");
            }}
            className="flex items-center justify-center size-6 rounded-md text-(--text-faint) hover:text-(--fg) cursor-pointer"
          >
            <X size={12} />
          </button>
        </form>
      )}

      {editing && (
        <div onClick={(e) => e.stopPropagation()}>
          <EditExerciseDialog
            exercise={exercise}
            onClose={() => setEditing(false)}
            onSave={(input) => {
              onUpdate(exercise.id, input);
              setEditing(false);
            }}
          />
        </div>
      )}

      {confirming && (
        <div onClick={(e) => e.stopPropagation()}>
          <ConfirmDeleteDialog
            title="Delete exercise?"
            message={`"${exercise.name}" and all its daily logs will be deleted.`}
            onCancel={() => setConfirming(false)}
            onConfirm={() => {
              onDelete(exercise.id);
              setConfirming(false);
            }}
          />
        </div>
      )}
    </div>
  );
}

function round1(n: number): number {
  return Math.round(n * 10) / 10;
}
