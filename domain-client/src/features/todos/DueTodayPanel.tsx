import type { ReactNode } from "react";
import { LABEL_COLOR_VAR } from "../labels/colors";
import type { Label } from "../labels/types";
import { getTodayIST } from "./dueToday";
import type { Todo } from "./types";

interface DueTodayPanelProps {
  todos: Todo[];
  labels: Label[];
}

function formatDateShort(value: string) {
  return new Date(value).toLocaleDateString(undefined, { month: "short", day: "numeric" });
}

function TodoRow({ todo, label, trailing }: { todo: Todo; label: Label | null; trailing?: ReactNode }) {
  return (
    <div className="flex items-center gap-1.5 rounded-lg bg-(--card-alt) border-(--line) border-[0.5px] border-solid px-2.5 py-1.5">
      {label && (
        <span className="size-1.5 rounded-full shrink-0" style={{ backgroundColor: LABEL_COLOR_VAR[label.color] }} />
      )}
      <span className="flex-1 min-w-0 truncate text-[length:var(--text-pill)] text-(--fg)">{todo.name}</span>
      {trailing}
    </div>
  );
}

export function DueTodayPanel({ todos, labels }: DueTodayPanelProps) {
  const today = getTodayIST();
  const labelFor = (todo: Todo) => labels.find((l) => l.id === todo.labelId) ?? null;
  // Only active todos — a completed todo isn't something left to act on, so
  // it's already surfaced (struck through) in the Completed tab.
  const active = todos.filter((t) => !t.done);
  const dueToday = active.filter((t) => t.targetDate === today);
  // targetDate is a plain "YYYY-MM-DD" string, so a lexical compare against
  // today is a calendar-date compare. Oldest (most overdue) first.
  const pastDue = active
    .filter((t) => t.targetDate && t.targetDate < today)
    .sort((a, b) => (a.targetDate! < b.targetDate! ? -1 : 1));

  return (
    <div className="shrink-0 bg-(--card) border-(--line) border-[0.5px] border-solid rounded-xl p-4">
      <h2 className="text-[length:var(--text-pill)] font-medium uppercase tracking-wide text-(--text-faint) mb-3">
        Due today
      </h2>

      {dueToday.length === 0 ? (
        <p className="text-[length:var(--text-pill)] text-(--text-faint)">Nothing due today.</p>
      ) : (
        <div className="flex flex-col gap-1.5">
          {dueToday.map((todo) => (
            <TodoRow key={todo.id} todo={todo} label={labelFor(todo)} />
          ))}
        </div>
      )}

      <h2 className="text-[length:var(--text-pill)] font-medium uppercase tracking-wide text-(--text-faint) mt-4 mb-3">
        Past due
      </h2>

      {pastDue.length === 0 ? (
        <p className="text-[length:var(--text-pill)] text-(--text-faint)">Nothing past due.</p>
      ) : (
        <div className="flex flex-col gap-1.5">
          {pastDue.map((todo) => (
            <TodoRow
              key={todo.id}
              todo={todo}
              label={labelFor(todo)}
              trailing={
                <span className="shrink-0 text-[length:var(--text-pill)] text-(--label-red)">
                  {formatDateShort(todo.targetDate!)}
                </span>
              }
            />
          ))}
        </div>
      )}
    </div>
  );
}
