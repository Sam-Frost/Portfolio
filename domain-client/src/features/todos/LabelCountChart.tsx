import { LABEL_COLOR_VAR } from "../labels/colors";
import type { Label } from "../labels/types";
import type { Todo } from "./types";

interface LabelCountChartProps {
  todos: Todo[];
  labels: Label[];
  tab: "active" | "completed";
}

function countByStatus(todos: Todo[], labelId: string | null) {
  const matches = todos.filter((t) => t.labelId === labelId);
  return { active: matches.filter((t) => !t.done).length, completed: matches.filter((t) => t.done).length };
}

export function LabelCountChart({ todos, labels, tab }: LabelCountChartProps) {
  const rows = [
    ...labels.map((label) => ({
      key: label.id,
      name: label.name,
      color: LABEL_COLOR_VAR[label.color],
      ...countByStatus(todos, label.id),
    })),
    {
      key: "none",
      name: "No label",
      color: "var(--text-faint)",
      ...countByStatus(todos, null),
    },
  ]
    .map((row) => ({ ...row, total: row.active + row.completed }))
    .filter((row) => row.total > 0)
    .sort((a, b) => b.total - a.total);

  const max = Math.max(...rows.map((row) => row.total), 1);

  return (
    <div className="shrink-0 bg-(--card) border-(--line) border-[0.5px] border-solid rounded-xl p-4">
      <h2 className="text-[length:var(--text-pill)] font-medium uppercase tracking-wide text-(--text-faint) mb-3">
        Label count
      </h2>

      {rows.length === 0 ? (
        <p className="text-[length:var(--text-pill)] text-(--text-faint)">No labels used yet.</p>
      ) : (
        <div className="flex flex-col gap-2.5">
          {rows.map((row) => {
            const trackPct = (row.total / max) * 100;
            const activePct = (row.active / row.total) * trackPct;
            const completedPct = (row.completed / row.total) * trackPct;

            return (
              <div key={row.key} className="flex items-center gap-2">
                <span className="w-20 shrink-0 truncate text-[length:var(--text-pill)] text-(--text-muted)">
                  {row.name}
                </span>
                <div className="flex-1 h-1.5 rounded-full bg-(--card-alt) overflow-hidden flex gap-[2px]">
                  {row.active > 0 && (
                    <div
                      className={`h-full ${row.completed > 0 ? "rounded-l-full" : "rounded-full"}`}
                      style={{ width: `${activePct}%`, backgroundColor: row.color }}
                    />
                  )}
                  {row.completed > 0 && (
                    <div
                      className={`h-full opacity-35 ${row.active > 0 ? "rounded-r-full" : "rounded-full"}`}
                      style={{ width: `${completedPct}%`, backgroundColor: row.color }}
                    />
                  )}
                </div>
                <span className="w-6 shrink-0 text-right text-[length:var(--text-pill)] text-(--fg)">
                  {row[tab]}
                </span>
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
}
