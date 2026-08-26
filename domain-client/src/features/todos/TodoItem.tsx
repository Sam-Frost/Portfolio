import { Check } from "lucide-react";
import { useState } from "react";
import { LABEL_COLOR_VAR } from "../labels/colors";
import type { Label } from "../labels/types";
import { EditTodoDialog } from "./EditTodoDialog";
import type { Todo } from "./types";

interface TodoItemProps {
  todo: Todo;
  labels: Label[];
  onMarkDone: (id: string) => void;
  onUpdate: (
    id: string,
    input: { name: string; description: string | null; targetDate: string | null; labelId: string | null },
  ) => void;
}

function formatDate(value: string) {
  return new Date(value).toLocaleDateString(undefined, { year: "numeric", month: "short", day: "numeric" });
}

export function TodoItem({ todo, labels, onMarkDone, onUpdate }: TodoItemProps) {
  const [editing, setEditing] = useState(false);
  const label = labels.find((l) => l.id === todo.labelId) ?? null;

  return (
    <div
      onDoubleClick={() => setEditing(true)}
      className="group relative flex items-start gap-2 bg-(--card) border-(--line) border-[0.5px] border-solid rounded-lg px-3 py-1.5"
    >
      <div className="flex-1 min-w-0">
        <span
          className={`text-[length:var(--text-pill)] break-words ${
            todo.done ? "text-(--text-faint) line-through" : "text-(--fg)"
          }`}
        >
          {todo.name}
        </span>
      </div>

      {label && (
        <span className="flex items-center gap-1 shrink-0 max-w-24 rounded-md bg-(--card-alt) border-(--line) border-[0.5px] border-solid px-1.5 py-0.5 text-[length:var(--text-pill)] text-(--text-muted)">
          <span className="size-1.5 rounded-full shrink-0" style={{ backgroundColor: LABEL_COLOR_VAR[label.color] }} />
          <span className="truncate">{label.name}</span>
        </span>
      )}

      {todo.done ? (
        <span className="flex items-center gap-1 shrink-0 text-[length:var(--text-pill)] text-(--green)">
          <Check size={12} strokeWidth={3} />
          Done
        </span>
      ) : (
        <button
          onClick={() => onMarkDone(todo.id)}
          className="shrink-0 rounded-md bg-(--card-alt) border-(--line) border-[0.5px] border-solid px-2 py-0.5 text-[length:var(--text-pill)] text-(--text-muted) hover:text-(--fg) hover:border-(--fg) transition-colors cursor-pointer"
        >
          Done
        </button>
      )}

      {/* Hovering anywhere on the card reveals this — delay-0 in the base
          state keeps mouse-leave hiding instant, group-hover:delay-1000
          means the 1s hold only applies on the way in. */}
      <div className="pointer-events-none absolute left-3 top-full z-10 mt-1 w-max max-w-64 whitespace-pre-wrap rounded-md border-(--line) border-[0.5px] border-solid bg-(--card-alt) px-2.5 py-1.5 text-[length:var(--text-pill)] text-(--text-muted) opacity-0 shadow-lg transition-opacity delay-0 duration-150 group-hover:opacity-100 group-hover:delay-1000">
        <p className="whitespace-nowrap">Added: {formatDate(todo.dateAdded)}</p>
        {todo.targetDate && <p className="whitespace-nowrap">Due: {formatDate(todo.targetDate)}</p>}
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
    </div>
  );
}
