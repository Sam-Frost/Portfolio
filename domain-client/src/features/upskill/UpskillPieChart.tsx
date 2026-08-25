const SIZE = 104;
const STROKE = 14;
const RADIUS = (SIZE - STROKE) / 2;
const CIRCUMFERENCE = 2 * Math.PI * RADIUS;

interface UpskillPieChartProps {
  done: number;
  total: number;
}

// Self-contained inline-SVG donut — covered vs. pending subtopics — so this
// small app doesn't need a charting dependency for one chart. --green marks
// covered work; --ring-track (already reserved for progress-ring remainder)
// marks pending.
export function UpskillPieChart({ done, total }: UpskillPieChartProps) {
  const pending = Math.max(total - done, 0);
  const doneFraction = total > 0 ? done / total : 0;
  const doneLength = CIRCUMFERENCE * doneFraction;

  return (
    <div className="flex items-center gap-5">
      <div className="relative shrink-0" style={{ width: SIZE, height: SIZE }}>
        <svg width={SIZE} height={SIZE} viewBox={`0 0 ${SIZE} ${SIZE}`} className="-rotate-90">
          <circle
            cx={SIZE / 2}
            cy={SIZE / 2}
            r={RADIUS}
            fill="none"
            stroke="var(--ring-track)"
            strokeWidth={STROKE}
          />
          {total > 0 && (
            <circle
              cx={SIZE / 2}
              cy={SIZE / 2}
              r={RADIUS}
              fill="none"
              stroke="var(--green)"
              strokeWidth={STROKE}
              strokeDasharray={`${doneLength} ${CIRCUMFERENCE - doneLength}`}
              strokeLinecap={doneFraction > 0 && doneFraction < 1 ? "round" : "butt"}
            />
          )}
        </svg>
        <div className="absolute inset-0 flex flex-col items-center justify-center">
          <span className="text-[length:var(--text-caption)] text-(--fg) font-medium">
            {done}/{total}
          </span>
          <span className="text-[length:var(--text-pill)] text-(--text-faint)">done</span>
        </div>
      </div>

      <div className="flex flex-col gap-2 text-[length:var(--text-pill)] text-(--text-muted)">
        <div className="flex items-center gap-1.5">
          <span className="size-2 rounded-full shrink-0" style={{ backgroundColor: "var(--green)" }} />
          Covered ({done})
        </div>
        <div className="flex items-center gap-1.5">
          <span className="size-2 rounded-full shrink-0" style={{ backgroundColor: "var(--ring-track)" }} />
          Pending ({pending})
        </div>
      </div>
    </div>
  );
}
