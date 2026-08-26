import { formatMinutes } from "./dateUtils";
import type { DailySummary } from "./types";

interface DailyHoursChartProps {
  summaries: DailySummary[];
  monthLabel: string;
}

// Hand-rolled bar chart (no charting dependency, matching
// features/todos/LabelCountChart's approach) — one thin, rounded-top bar
// per day of the visible month, a single --gold hue since this is one
// series, with a per-bar hover tooltip.
export function DailyHoursChart({ summaries, monthLabel }: DailyHoursChartProps) {
  const max = Math.max(1, ...summaries.map((s) => s.workedMinutes));
  const hasData = summaries.some((s) => s.workedMinutes > 0);

  return (
    <div className="bg-(--card) border-(--line) border-[0.5px] border-solid rounded-xl p-4">
      <div className="flex items-center justify-between mb-3">
        <h2 className="text-[length:var(--text-pill)] font-medium uppercase tracking-wide text-(--text-faint)">
          Hours worked — {monthLabel}
        </h2>
        {hasData && <span className="text-[length:var(--text-pill)] text-(--text-faint)">{formatMinutes(max)} max/day</span>}
      </div>

      {!hasData ? (
        <p className="text-[length:var(--text-pill)] text-(--text-faint) py-8 text-center">
          No sessions logged this month yet.
        </p>
      ) : (
        <div className="flex items-end gap-[3px] h-32">
          {summaries.map((s) => {
            const heightPct = s.workedMinutes > 0 ? Math.max(4, (s.workedMinutes / max) * 100) : 0;
            const day = Number(s.date.slice(-2));

            return (
              <div key={s.date} className="group relative flex-1 min-w-0 h-full flex flex-col items-center justify-end">
                {s.workedMinutes > 0 && (
                  <div
                    className="w-full rounded-t-[4px] bg-(--gold) transition-opacity group-hover:opacity-80"
                    style={{ height: `${heightPct}%` }}
                  />
                )}
                {day % 5 === 0 && <span className="mt-1 text-[9px] leading-none text-(--text-faint)">{day}</span>}
                <span className="pointer-events-none absolute bottom-full mb-1.5 left-1/2 -translate-x-1/2 z-10 whitespace-nowrap rounded-md border-(--line) border-[0.5px] border-solid bg-(--card-alt) px-2 py-1 text-[length:var(--text-pill)] text-(--text-muted) opacity-0 shadow-lg transition-opacity duration-150 group-hover:opacity-100">
                  {s.date}: {formatMinutes(s.workedMinutes)}
                </span>
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
}
