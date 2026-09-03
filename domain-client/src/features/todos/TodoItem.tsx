import { Bell, Check, Undo2 } from "lucide-react";
import { useState } from "react";
import { LABEL_COLOR_VAR } from "../labels/colors";
import type { Label } from "../labels/types";
import { EditTodoDialog } from "./EditTodoDialog";
import { ReminderDialog } from "./reminders/ReminderDialog";
import type { Todo } from "./types";

interface TodoItemProps {
  todo: Todo;
  labels: Label[];
  onMarkDone: (id: string) => void;
  onUndo: (id: string) => void;
  onUpdate: (
    id: string,
    input: { name: string; description: string | null; targetDate: string | null; labelId: string | null },
  ) => void;
}

function formatDate(value: string) {
  return new Date(value).toLocaleDateString(undefined, { year: "numeric", month: "short", day: "numeric" });
}

function formatDateShort(value: string) {
  return new Date(value).toLocaleDateString(undefined, { month: "short", day: "numeric" });
}

export function TodoItem({ todo, labels, onMarkDone, onUndo, onUpdate }: TodoItemProps) {
  const [editing, setEditing] = useState(false);
  const [remindering, setRemindering] = useState(false);
  const label = labels.find((l) => l.id === todo.labelId) ?? null;

  return (
    <div
      onDoubleClick={() => setEditing(true)}
      className="group relative flex flex-col gap-1.5 sm:flex-row sm:items-start sm:gap-2 bg-(--card) border-(--line) border-[0.5px] border-solid rounded-lg px-3 py-2 sm:py-1.5"
    >
      <div className="flex-1 min-w-0">
        <span
          className={`block text-[length:var(--text-caption)] sm:text-[length:var(--text-pill)] sm:truncate ${
            todo.done ? "text-(--text-faint) line-through" : "text-(--fg)"
          }`}
        >
          {todo.targetDate && (
            <span className="text-(--text-faint)">{formatDateShort(todo.targetDate)} — </span>
          )}
          {todo.name}
        </span>
      </div>

      {/* On mobile this is a full-width row beneath the name, its contents
          (label + action) grouped at the left edge under the title;
          `sm:contents` dissolves it on desktop so the label and the Done
          control sit inline at the end of the title row as before. */}
      <div className="flex items-center gap-2 shrink-0 sm:contents">
        {label && (
          <span className="flex items-center gap-1 shrink-0 max-w-40 sm:max-w-24 rounded-md bg-(--card-alt) border-(--line) border-[0.5px] border-solid px-1.5 py-0.5 text-[length:var(--text-pill)] text-(--text-muted)">
            <span className="size-1.5 rounded-full shrink-0" style={{ backgroundColor: LABEL_COLOR_VAR[label.color] }} />
            <span className="truncate">{label.name}</span>
          </span>
        )}

        {!todo.done && (
          <button
            onClick={() => setRemindering(true)}
            aria-label="Reminders"
            title="Reminders"
            className="shrink-0 flex items-center justify-center size-6 sm:size-5 rounded-md text-(--text-faint) hover:text-(--fg) hover:bg-(--card-alt) transition-colors cursor-pointer"
          >
            <Bell size={12} />
          </button>
        )}

        {todo.done ? (
          <span className="flex items-center gap-1 shrink-0 sm:ml-auto">
            <span className="flex items-center gap-1 text-[length:var(--text-pill)] text-(--green)">
              <Check size={12} strokeWidth={3} />
              Done
              {todo.completedAt && (
                <span className="text-(--text-faint)">· {formatDateShort(todo.completedAt)}</span>
              )}
            </span>
            <button
              onClick={() => onUndo(todo.id)}
              aria-label="Undo completion"
              title="Mark as not done"
              className="flex items-center justify-center size-6 sm:size-5 rounded-md text-(--text-faint) hover:text-(--fg) hover:bg-(--card-alt) transition-colors cursor-pointer"
            >
              <Undo2 size={12} />
            </button>
          </span>
        ) : (
          <button
            onClick={() => onMarkDone(todo.id)}
            className="shrink-0 sm:ml-auto rounded-md bg-(--card-alt) border-(--line) border-[0.5px] border-solid px-2.5 py-1 sm:px-2 sm:py-0.5 text-[length:var(--text-pill)] text-(--text-muted) hover:text-(--fg) hover:border-(--fg) transition-colors cursor-pointer"
          >
            Done
          </button>
        )}
      </div>

      {/* Hovering anywhere on the card reveals this — delay-0 in the base
          state keeps mouse-leave hiding instant, group-hover:delay-1000
          means the 1s hold only applies on the way in. */}
      <div className="pointer-events-none absolute left-3 top-full z-10 mt-1 w-max max-w-64 whitespace-pre-wrap rounded-md border-(--line) border-[0.5px] border-solid bg-(--card-alt) px-2.5 py-1.5 text-[length:var(--text-pill)] text-(--text-muted) opacity-0 shadow-lg transition-opacity delay-0 duration-150 group-hover:opacity-100 group-hover:delay-1000">
        <p className="font-medium text-(--fg)">{todo.name}</p>
        <p className="whitespace-nowrap">Added: {formatDate(todo.dateAdded)}</p>
        {todo.targetDate && <p className="whitespace-nowrap">Due: {formatDate(todo.targetDate)}</p>}
        {todo.done && todo.completedAt && (
          <p className="whitespace-nowrap">Done: {formatDate(todo.completedAt)}</p>
        )}
        {todo.description && (
          <p className="mt-1 border-t-(--line) border-t-[0.5px] border-solid pt-1">{todo.description}</p>
        )}
      </div>

      {editing && (
        <EditTodoDialog
          todo={todo}
          labels={labels}
          onClose={() => setEditing(false)}
          onSave={(input) => {
            onUpdate(todo.id, input);
            setEditing(false);
          }}
        />
      )}

      {remindering && <ReminderDialog todo={todo} onClose={() => setRemindering(false)} />}
    </div>
  );
}
