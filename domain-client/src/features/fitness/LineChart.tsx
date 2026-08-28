import { formatAxisDate, formatDayKey } from "./dateUtils";

export interface LinePoint {
  date: string;
  value: number;
}

interface LineChartProps {
  title: string;
  points: LinePoint[];
  unit?: string;
  // Optional dashed horizontal reference line (e.g. target weight).
  target?: number | null;
  emptyLabel?: string;
}

// A wide viewBox close to the rendered aspect ratio, so `preserveAspectRatio
// ="none"` stretching stays imperceptible; strokes are kept crisp with
// vector-effect below.
const VW = 600;
const VH = 150;
const PAD_X = 12;
const PAD_Y = 16;

// Hand-rolled inline-SVG line chart — a value over time — kept dependency-free
// in the same spirit as features/upskill/UpskillPieChart.tsx. Points are
// evenly spaced along x (one per logged day); y is scaled to the data range
// (plus the target line, if any) with a little padding. A left gutter shows
// the y min/max and a row beneath shows ~4 date ticks.
//
// The line + target rule live in a `preserveAspectRatio="none"` SVG (so the
// wide viewBox stretches to fill any width). Point markers can't live there
// — that stretch would squash them into ellipses — so they're HTML dots
// positioned by percentage in an overlay, each with a hover tooltip showing
// that day's value.
export function LineChart({ title, points, unit, target, emptyLabel }: LineChartProps) {
  const hasData = points.length > 0;
  const hasTarget = target != null && target > 0;

  const values = points.map((p) => p.value);
  const lo = Math.min(...values, hasTarget ? target : Infinity);
  const hi = Math.max(...values, hasTarget ? target : -Infinity);
  const span = hi - lo || 1;
  const yMin = lo - span * 0.15;
  const yMax = hi + span * 0.15;

  const x = (i: number) =>
    points.length <= 1 ? VW / 2 : PAD_X + (i / (points.length - 1)) * (VW - 2 * PAD_X);
  const y = (v: number) => PAD_Y + (1 - (v - yMin) / (yMax - yMin)) * (VH - 2 * PAD_Y);

  const linePath = points.map((p, i) => `${i === 0 ? "M" : "L"} ${x(i).toFixed(1)} ${y(p.value).toFixed(1)}`).join(" ");
  const ticks = axisTickIndexes(points.length);

  return (
    <div className="bg-(--card) border-(--line) border-[0.5px] border-solid rounded-xl p-4">
      <div className="flex items-center justify-between mb-3">
        <h2 className="text-[length:var(--text-pill)] font-medium uppercase tracking-wide text-(--text-faint)">
          {title}
        </h2>
        {hasData && (
          <span className="text-[length:var(--text-pill)] text-(--text-faint)">
            {round1(points[0].value)} → {round1(points[points.length - 1].value)}
            {unit ? ` ${unit}` : ""}
          </span>
        )}
      </div>

      {!hasData ? (
        <p className="text-[length:var(--text-pill)] text-(--text-faint) py-8 text-center">
          {emptyLabel ?? "Nothing logged yet."}
        </p>
      ) : (
        <div className="flex gap-2">
          <div className="flex flex-col justify-between h-36 shrink-0 text-right text-[9px] leading-none text-(--text-faint) py-[6px]">
            <span>
              {round1(yMax)}
              {unit ? ` ${unit}` : ""}
            </span>
            {hasTarget && <span className="text-(--green)">{round1(target)}</span>}
            <span>{round1(yMin)}</span>
          </div>

          <div className="flex-1 min-w-0">
            <div className="relative h-36">
              <svg viewBox={`0 0 ${VW} ${VH}`} className="absolute inset-0 h-full w-full" preserveAspectRatio="none">
                {hasTarget && (
                  <line
                    x1={PAD_X}
                    x2={VW - PAD_X}
                    y1={y(target)}
                    y2={y(target)}
                    stroke="var(--green)"
                    strokeWidth={1}
                    strokeDasharray="4 3"
                    vectorEffect="non-scaling-stroke"
                  />
                )}
                <path
                  d={linePath}
                  fill="none"
                  stroke="var(--gold)"
                  strokeWidth={2}
                  strokeLinejoin="round"
                  strokeLinecap="round"
                  vectorEffect="non-scaling-stroke"
                />
              </svg>

              {/* Round HTML dots + hover tooltip — kept out of the stretched SVG. */}
              <div className="absolute inset-0">
                {points.map((p, i) => (
                  <div
                    key={p.date}
                    className="group absolute flex size-4 -translate-x-1/2 -translate-y-1/2 items-center justify-center"
                    style={{ left: `${(x(i) / VW) * 100}%`, top: `${(y(p.value) / VH) * 100}%` }}
                  >
                    <div className="size-1.5 rounded-full bg-(--gold) ring-2 ring-(--card) transition-transform group-hover:scale-150" />
                    <span className="pointer-events-none absolute bottom-full left-1/2 mb-1.5 -translate-x-1/2 z-20 whitespace-nowrap rounded-md border-(--line) border-[0.5px] border-solid bg-(--card-alt) px-2 py-1 text-[length:var(--text-pill)] text-(--text-muted) opacity-0 shadow-lg transition-opacity duration-150 group-hover:opacity-100">
                      {formatDayKey(p.date)}: {round1(p.value)}
                      {unit ? ` ${unit}` : ""}
                    </span>
                  </div>
                ))}
              </div>
            </div>

            <div className="relative h-3 mt-1 text-[9px] leading-none text-(--text-faint)">
              {ticks.map((i) => (
                <span
                  key={i}
                  className="absolute -translate-x-1/2 whitespace-nowrap"
                  style={{ left: `${points.length <= 1 ? 50 : (i / (points.length - 1)) * 100}%` }}
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
