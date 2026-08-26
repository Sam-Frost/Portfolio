import { LABEL_COLOR_VAR } from "../labels/colors";
import type { Label } from "../labels/types";
import type { Todo } from "./types";

interface LabelDistributionPieProps {
  todos: Todo[];
  labels: Label[];
  tab: "active" | "completed";
}

// Anything past this many slices blurs together, so the smallest ones fold
// into a single "Other" wedge instead (see dataviz anti-patterns: pie charts
// only read at a glance up to ~6 segments).
const MAX_SLICES = 6;

export function LabelDistributionPie({ todos, labels, tab }: LabelDistributionPieProps) {
  const scopedTodos = todos.filter((t) => t.done === (tab === "completed"));

  const rows = [
    ...labels.map((label) => ({
      key: label.id,
      name: label.name,
      color: LABEL_COLOR_VAR[label.color],
      count: scopedTodos.filter((t) => t.labelId === label.id).length,
    })),
    {
      key: "none",
      name: "No label",
      color: "var(--text-faint)",
      count: scopedTodos.filter((t) => t.labelId === null).length,
    },
  ]
    .filter((row) => row.count > 0)
    .sort((a, b) => b.count - a.count);

  const total = rows.reduce((sum, row) => sum + row.count, 0);

  const visible = rows.length > MAX_SLICES ? rows.slice(0, MAX_SLICES - 1) : rows;
  const rest = rows.length > MAX_SLICES ? rows.slice(MAX_SLICES - 1) : [];
  const slices = visible.map((row) => ({ ...row, pct: total ? (row.count / total) * 100 : 0 }));
  if (rest.length > 0) {
    const restCount = rest.reduce((sum, row) => sum + row.count, 0);
    slices.push({
      key: "other",
      name: "Other",
      color: "var(--text-muted)",
      count: restCount,
      pct: total ? (restCount / total) * 100 : 0,
    });
  }

  let cursor = 0;
  const gradient = slices
    .map((slice) => {
      const start = cursor;
      cursor += slice.pct;
      return `${slice.color} ${start}% ${cursor}%`;
    })
    .join(", ");

  return (
    <div className="shrink-0 bg-(--card) border-(--line) border-[0.5px] border-solid rounded-xl p-4">
      <h2 className="text-[length:var(--text-pill)] font-medium uppercase tracking-wide text-(--text-faint) mb-3">
        Label distribution
      </h2>

      {total === 0 ? (
        <p className="text-[length:var(--text-pill)] text-(--text-faint)">
          {tab === "active" ? "No active todos." : "No completed todos yet."}
        </p>
      ) : (
        <div className="flex flex-col items-center gap-4">
          <div className="relative size-24 shrink-0" role="img" aria-label="Todo distribution by label">
            <div className="absolute inset-0 rounded-full" style={{ background: `conic-gradient(${gradient})` }} />
            <div className="absolute inset-[19px] rounded-full bg-(--card) flex items-center justify-center">
              <span className="text-[length:var(--text-pill)] text-(--text-faint)">{total}</span>
            </div>
          </div>

          <div className="flex flex-col gap-1.5 w-full">
            {slices.map((slice) => (
              <div key={slice.key} className="flex items-center gap-1.5">
                <span className="size-2 rounded-full shrink-0" style={{ backgroundColor: slice.color }} />
                <span className="flex-1 min-w-0 truncate text-[length:var(--text-pill)] text-(--text-muted)">
                  {slice.name}
                </span>
                <span className="shrink-0 text-[length:var(--text-pill)] text-(--fg)">
                  {Math.round(slice.pct)}%
                </span>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
