const SIZE = 104;
const STROKE = 14;
const RADIUS = (SIZE - STROKE) / 2;
const CIRCUMFERENCE = 2 * Math.PI * RADIUS;

interface GoalDonutProps {
  current: number;
  goal: number | null;
  unit?: string | null;
}

// Self-contained inline-SVG donut — progress toward a goal quantity — same
// approach as features/upskill/UpskillPieChart.tsx so the app keeps its
// no-charting-dependency stance. --green marks progress; --ring-track marks
// what's left.
export function GoalDonut({ current, goal, unit }: GoalDonutProps) {
  const hasGoal = goal != null && goal > 0;
  const fraction = hasGoal ? Math.min(current / goal, 1) : 0;
  const doneLength = CIRCUMFERENCE * fraction;
  const pct = hasGoal ? Math.round((current / goal) * 100) : 0;
  const remaining = hasGoal ? Math.max(goal - current, 0) : 0;

  return (
    <div className="flex items-center gap-5">
      <div className="relative shrink-0" style={{ width: SIZE, height: SIZE }}>
        <svg width={SIZE} height={SIZE} viewBox={`0 0 ${SIZE} ${SIZE}`} className="-rotate-90">
          <circle cx={SIZE / 2} cy={SIZE / 2} r={RADIUS} fill="none" stroke="var(--ring-track)" strokeWidth={STROKE} />
          {hasGoal && (
            <circle
              cx={SIZE / 2}
              cy={SIZE / 2}
              r={RADIUS}
              fill="none"
              stroke="var(--green)"
              strokeWidth={STROKE}
              strokeDasharray={`${doneLength} ${CIRCUMFERENCE - doneLength}`}
              strokeLinecap={fraction > 0 && fraction < 1 ? "round" : "butt"}
            />
          )}
        </svg>
        <div className="absolute inset-0 flex flex-col items-center justify-center">
          <span className="text-[length:var(--text-caption)] text-(--fg) font-medium">
            {hasGoal ? `${pct}%` : "—"}
          </span>
          <span className="text-[length:var(--text-pill)] text-(--text-faint)">to goal</span>
        </div>
      </div>

      <div className="flex flex-col gap-1.5 text-[length:var(--text-pill)] text-(--text-muted)">
        <div>
          Logged: <span className="text-(--fg)">{round1(current)}</span>
          {unit ? ` ${unit}` : ""}
        </div>
        {hasGoal ? (
          <>
            <div>
              Goal: <span className="text-(--fg)">{round1(goal)}</span>
              {unit ? ` ${unit}` : ""}
            </div>
            <div>
              Remaining: <span className="text-(--fg)">{round1(remaining)}</span>
              {unit ? ` ${unit}` : ""}
            </div>
          </>
        ) : (
          <div className="text-(--text-faint)">No goal set</div>
        )}
      </div>
    </div>
  );
}

function round1(n: number): number {
  return Math.round(n * 10) / 10;
}
