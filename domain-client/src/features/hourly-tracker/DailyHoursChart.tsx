import { formatMinutes } from "./dateUtils";
import type { DailySummary } from "./types";

interface DailyHoursChartProps {
  summaries: DailySummary[];
  monthLabel: string;
}

// Hand-rolled bar chart (no charting dependency, matching
// features/todos/LabelCountChart's approach) — one thin bar per day of the
// visible month, split into a --gold professional segment and a --green
// personal segment, with a per-bar hover tooltip breaking the two out.
export function DailyHoursChart({ summaries, monthLabel }: DailyHoursChartProps) {
  const max = Math.max(1, ...summaries.map((s) => s.workedMinutes));
  const hasData = summaries.some((s) => s.workedMinutes > 0);

  return (
    <div className="bg-(--card) border-(--line) border-[0.5px] border-solid rounded-xl p-4">
      <div className="flex items-center justify-between mb-3">
        <h2 className="text-[length:var(--text-pill)] font-medium uppercase tracking-wide text-(--text-faint)">
          Hours worked — {monthLabel}
        </h2>
        {hasData && (
          <div className="flex items-center gap-3 text-[length:var(--text-pill)] text-(--text-faint)">
            <span className="flex items-center gap-1">
              <span className="inline-block h-2 w-2 rounded-[2px] bg-(--gold)" /> Professional
            </span>
            <span className="flex items-center gap-1">
              <span className="inline-block h-2 w-2 rounded-[2px] bg-(--green)" /> Personal
            </span>
          </div>
        )}
      </div>

      {!hasData ? (
        <p className="text-[length:var(--text-pill)] text-(--text-faint) py-8 text-center">
          No sessions logged this month yet.
        </p>
      ) : (
        <div className="flex items-end gap-[3px] h-32 mb-4">
          {summaries.map((s) => {
            const totalPct = s.workedMinutes > 0 ? Math.max(4, (s.workedMinutes / max) * 100) : 0;
            // Each segment's height as a share of the (already scaled) bar,
            // so a short day still shows both categories in proportion.
            const proPct = s.workedMinutes > 0 ? (s.professionalMinutes / s.workedMinutes) * 100 : 0;
            const personalPct = 100 - proPct;
            const day = Number(s.date.slice(-2));

            return (
              <div key={s.date} className="group relative flex-1 min-w-0 h-full flex flex-col items-center justify-end">
                {s.workedMinutes > 0 && (
                  <div
                    className="w-full overflow-hidden rounded-t-[4px] transition-opacity group-hover:opacity-80"
                    style={{ height: `${totalPct}%` }}
                  >
                    <div className="w-full bg-(--green)" style={{ height: `${personalPct}%` }} />
                    <div className="w-full bg-(--gold)" style={{ height: `${proPct}%` }} />
                  </div>
                )}
                {/* Absolutely positioned so the day labels sit *below* the
                    baseline without stealing height from the bar's flex
                    column — otherwise every 5th day's bar (and only those)
                    gets pushed up off the shared baseline. */}
                {day % 5 === 0 && (
                  <span className="pointer-events-none absolute top-full mt-1 left-1/2 -translate-x-1/2 text-[9px] leading-none text-(--text-faint)">
                    {day}
                  </span>
                )}
                <span className="pointer-events-none absolute bottom-full mb-1.5 left-1/2 -translate-x-1/2 z-10 whitespace-nowrap rounded-md border-(--line) border-[0.5px] border-solid bg-(--card-alt) px-2 py-1 text-[length:var(--text-pill)] text-(--text-muted) opacity-0 shadow-lg transition-opacity duration-150 group-hover:opacity-100">
                  <span className="block text-(--fg)">{s.date}: {formatMinutes(s.workedMinutes)}</span>
                  {s.professionalMinutes > 0 && (
                    <span className="block">Professional: {formatMinutes(s.professionalMinutes)}</span>
                  )}
                  {s.personalMinutes > 0 && (
                    <span className="block">Personal: {formatMinutes(s.personalMinutes)}</span>
                  )}
                </span>
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
}
