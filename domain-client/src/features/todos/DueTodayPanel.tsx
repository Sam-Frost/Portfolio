import { LABEL_COLOR_VAR } from "../labels/colors";
import type { Label } from "../labels/types";
import { getTodayIST } from "./dueToday";
import type { Todo } from "./types";

interface DueTodayPanelProps {
  todos: Todo[];
  labels: Label[];
}

export function DueTodayPanel({ todos, labels }: DueTodayPanelProps) {
  const today = getTodayIST();
  // Only active todos — a completed todo due today isn't something left to
  // act on, so it's already surfaced (struck through) in the Completed tab.
  const dueToday = todos.filter((t) => !t.done && t.targetDate === today);

  return (
    <div className="shrink-0 bg-(--card) border-(--line) border-[0.5px] border-solid rounded-xl p-4">
      <h2 className="text-[length:var(--text-pill)] font-medium uppercase tracking-wide text-(--text-faint) mb-3">
        Due today
      </h2>

      {dueToday.length === 0 ? (
        <p className="text-[length:var(--text-pill)] text-(--text-faint)">Nothing due today.</p>
      ) : (
        <div className="flex flex-col gap-1.5">
          {dueToday.map((todo) => {
            const label = labels.find((l) => l.id === todo.labelId) ?? null;
            return (
              <div
                key={todo.id}
                className="flex items-center gap-1.5 rounded-lg bg-(--card-alt) border-(--line) border-[0.5px] border-solid px-2.5 py-1.5"
              >
                {label && (
                  <span
                    className="size-1.5 rounded-full shrink-0"
                    style={{ backgroundColor: LABEL_COLOR_VAR[label.color] }}
                  />
                )}
                <span className="flex-1 min-w-0 truncate text-[length:var(--text-pill)] text-(--fg)">{todo.name}</span>
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
}
