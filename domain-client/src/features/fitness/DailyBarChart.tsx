import { formatAxisDate, formatDayKey } from "./dateUtils";

export interface DailyPoint {
  date: string;
  value: number;
}

interface DailyBarChartProps {
  title: string;
  points: DailyPoint[];
  unit?: string;
  // Optional horizontal reference line (e.g. a protein target).
  target?: number | null;
  emptyLabel?: string;
}

// Hand-rolled bar chart — one bar per logged day, single --gold hue, per-bar
// hover tooltip — following features/hourly-tracker/DailyHoursChart.tsx.
// A left gutter carries the y-scale (max / mid / 0) and a row beneath the
// plot carries ~4 evenly-spaced date ticks. When `target` is given, a
// dashed --green rule is drawn across the plot.
export function DailyBarChart({ title, points, unit, target, emptyLabel }: DailyBarChartProps) {
  const hasTarget = target != null && target > 0;
  const max = Math.max(1, ...points.map((p) => p.value), hasTarget ? target : 0);
  const hasData = points.length > 0;
  const ticks = axisTickIndexes(points.length);

  return (
    <div className="bg-(--card) border-(--line) border-[0.5px] border-solid rounded-xl p-4">
      <div className="flex items-center justify-between mb-3">
        <h2 className="text-[length:var(--text-pill)] font-medium uppercase tracking-wide text-(--text-faint)">
          {title}
        </h2>
        {hasData && (
          <span className="text-[length:var(--text-pill)] text-(--text-faint)">
            {points.length} {points.length === 1 ? "day" : "days"}
          </span>
        )}
      </div>

      {!hasData ? (
        <p className="text-[length:var(--text-pill)] text-(--text-faint) py-8 text-center">
          {emptyLabel ?? "Nothing logged yet."}
        </p>
      ) : (
        <div className="flex gap-2">
          {/* y-axis scale */}
          <div className="flex flex-col justify-between h-32 shrink-0 text-right text-[9px] leading-none text-(--text-faint) py-[1px]">
            <span>
              {round1(max)}
              {unit ? ` ${unit}` : ""}
            </span>
            <span>{round1(max / 2)}</span>
            <span>0</span>
          </div>

          <div className="flex-1 min-w-0">
            <div className="relative h-32 border-l border-b border-(--line)">
              {hasTarget && (
                <div
                  className="absolute left-0 right-0 border-t border-dashed border-(--green) z-10"
                  style={{ bottom: `${(target / max) * 100}%` }}
                >
                  <span className="absolute -top-2.5 right-0 bg-(--card) px-1 text-[9px] leading-none text-(--green)">
                    target {round1(target)}
                  </span>
                </div>
              )}
              <div className="flex items-end gap-[3px] h-full px-0.5">
                {points.map((p) => {
                  const heightPct = p.value > 0 ? Math.max(3, (p.value / max) * 100) : 0;
                  return (
                    <div
                      key={p.date}
                      className="group relative flex-1 min-w-0 h-full flex flex-col items-center justify-end"
                    >
                      {p.value > 0 && (
                        <div
                          className="w-full rounded-t-[3px] bg-(--gold) transition-opacity group-hover:opacity-80"
                          style={{ height: `${heightPct}%` }}
                        />
                      )}
                      <span className="pointer-events-none absolute bottom-full mb-1.5 left-1/2 -translate-x-1/2 z-20 whitespace-nowrap rounded-md border-(--line) border-[0.5px] border-solid bg-(--card-alt) px-2 py-1 text-[length:var(--text-pill)] text-(--text-muted) opacity-0 shadow-lg transition-opacity duration-150 group-hover:opacity-100">
                        {formatDayKey(p.date)}: {round1(p.value)}
                        {unit ? ` ${unit}` : ""}
                      </span>
                    </div>
                  );
                })}
              </div>
            </div>

            {/* x-axis date ticks */}
            <div className="relative h-3 mt-1 text-[9px] leading-none text-(--text-faint)">
              {ticks.map((i, ti) => (
                <span
                  key={i}
                  className={`absolute whitespace-nowrap ${edgeAnchor(ti, ticks.length)}`}
                  style={{ left: `${((i + 0.5) / points.length) * 100}%` }}
                >
                  {formatAxisDate(points[i].date)}
                </span>
              ))}
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

// Keep the first/last axis labels inside the plot's own width (they sit near
// left:0%/100%): left-align the first, right-align the last, center the rest —
// otherwise they poke past the card edge and add a stray horizontal scrollbar.
function edgeAnchor(index: number, count: number): string {
  if (count <= 1) return "-translate-x-1/2";
  if (index === 0) return "translate-x-0";
  if (index === count - 1) return "-translate-x-full";
  return "-translate-x-1/2";
}

// Up to 4 roughly-even indexes into a series of length n (always includes
// first and last when n > 1).
function axisTickIndexes(n: number): number[] {
  if (n <= 1) return n === 1 ? [0] : [];
  const count = Math.min(4, n);
  const out = new Set<number>();
  for (let k = 0; k < count; k++) out.add(Math.round((k / (count - 1)) * (n - 1)));
  return [...out].sort((a, b) => a - b);
}

function round1(n: number): number {
  return Math.round(n * 10) / 10;
}
