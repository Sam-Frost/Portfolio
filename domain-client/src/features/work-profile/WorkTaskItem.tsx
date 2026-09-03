import { useState } from "react";
import { Check, Trash2, Undo2 } from "lucide-react";
import { JiraConfirmDialog } from "./JiraConfirmDialog";
import type { WorkTask } from "./types";

interface Props {
  task: WorkTask;
  onMarkDone: (id: string) => void;
  onUndo: (id: string) => void;
  onDelete: (id: string) => void;
}

function formatDateShort(value: string) {
  return new Date(value).toLocaleDateString(undefined, { month: "short", day: "numeric" });
}

function isOverdue(task: WorkTask): boolean {
  if (task.done || !task.targetDate) return false;
  const today = new Date();
  today.setHours(0, 0, 0, 0);
  return new Date(task.targetDate).getTime() < today.getTime();
}

export function WorkTaskItem({ task, onMarkDone, onUndo, onDelete }: Props) {
  const [confirming, setConfirming] = useState(false);

  return (
    <div className="group relative flex items-center gap-2 bg-(--card) border-(--line) border-[0.5px] border-solid rounded-lg px-3 py-2">
      <div className="flex-1 min-w-0">
        <span
          className={`block text-[length:var(--text-caption)] sm:text-[length:var(--text-pill)] ${
            task.done ? "text-(--text-faint) line-through" : "text-(--fg)"
          }`}
        >
          {task.targetDate && (
            <span className={isOverdue(task) ? "text-red-400" : "text-(--text-faint)"}>
              {formatDateShort(task.targetDate)} —{" "}
            </span>
          )}
          {task.name}
        </span>
        {task.description && (
          <span className="block truncate text-[length:var(--text-pill)] text-(--text-faint)">
            {task.description}
          </span>
        )}
      </div>

      {task.done ? (
        <>
          <span className="flex items-center gap-1 shrink-0 text-[length:var(--text-pill)] text-(--green)">
            <Check size={12} strokeWidth={3} />
            Done
            {task.completedAt && (
              <span className="text-(--text-faint)">· {formatDateShort(task.completedAt)}</span>
            )}
          </span>
          <button
            onClick={() => onUndo(task.id)}
            aria-label="Undo completion"
            className="shrink-0 flex items-center justify-center size-6 sm:size-5 rounded-md text-(--text-faint) hover:text-(--fg) hover:bg-(--card-alt) transition-colors cursor-pointer"
          >
            <Undo2 size={12} />
          </button>
        </>
      ) : (
        <button
          onClick={() => setConfirming(true)}
          className="shrink-0 rounded-md bg-(--card-alt) border-(--line) border-[0.5px] border-solid px-2.5 py-1 sm:px-2 sm:py-0.5 text-[length:var(--text-pill)] text-(--text-muted) hover:text-(--fg) hover:border-(--fg) transition-colors cursor-pointer"
        >
          Done
        </button>
      )}

      <button
        onClick={() => onDelete(task.id)}
        aria-label="Delete task"
        className="shrink-0 flex items-center justify-center size-6 sm:size-5 rounded-md text-(--text-faint) hover:text-red-400 hover:bg-(--card-alt) transition-colors cursor-pointer opacity-0 group-hover:opacity-100 focus-visible:opacity-100"
      >
        <Trash2 size={12} />
      </button>

      {confirming && (
        <JiraConfirmDialog
          taskName={task.name}
          onCancel={() => setConfirming(false)}
          onConfirm={() => {
            setConfirming(false);
            onMarkDone(task.id);
          }}
        />
      )}
    </div>
  );
}
