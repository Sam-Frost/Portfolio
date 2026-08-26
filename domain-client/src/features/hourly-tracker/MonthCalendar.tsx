import { ChevronLeft, ChevronRight } from "lucide-react";
import { daysInMonth, firstWeekdayOfMonth, formatMinutes, monthDayKey } from "./dateUtils";

const WEEKDAY_LABELS = ["S", "M", "T", "W", "T", "F", "S"];
const MONTH_LABELS = [
  "January",
  "February",
  "March",
  "April",
  "May",
  "June",
  "July",
  "August",
  "September",
  "October",
  "November",
  "December",
];

interface MonthCalendarProps {
  year: number;
  monthIndex: number;
  todayKey: string;
  selectedKey: string | null;
  workedMinutesByDate: Record<string, number>;
  onSelect: (dateKey: string) => void;
  onPrevMonth: () => void;
  onNextMonth: () => void;
}

// Small, self-contained month grid — no shared calendar component exists
// in this codebase yet to reuse. Each day cell is shaded by worked minutes
// that day (a single-hue --gold ramp against --card-alt, sequential/
// magnitude encoding) so the grid doubles as a lightweight heatmap.
export function MonthCalendar({
  year,
  monthIndex,
  todayKey,
  selectedKey,
  workedMinutesByDate,
  onSelect,
  onPrevMonth,
  onNextMonth,
}: MonthCalendarProps) {
  const totalDays = daysInMonth(year, monthIndex);
  const leadingBlanks = firstWeekdayOfMonth(year, monthIndex);
  const cells: (number | null)[] = [
    ...Array<null>(leadingBlanks).fill(null),
    ...Array.from({ length: totalDays }, (_, i) => i + 1),
  ];
  while (cells.length % 7 !== 0) cells.push(null);

  const maxWorked = Math.max(1, ...Object.values(workedMinutesByDate));

  return (
    <div className="bg-(--card) border-(--line) border-[0.5px] border-solid rounded-xl p-4">
      <div className="flex items-center justify-between mb-3">
        <button
          type="button"
          onClick={onPrevMonth}
          aria-label="Previous month"
          className="flex items-center justify-center size-7 rounded-lg text-(--text-muted) hover:text-(--fg) hover:bg-(--card-alt) transition-colors cursor-pointer"
        >
          <ChevronLeft size={15} />
        </button>
        <span className="text-[length:var(--text-caption)] text-(--fg) font-medium">
          {MONTH_LABELS[monthIndex]} {year}
        </span>
        <button
          type="button"
          onClick={onNextMonth}
          aria-label="Next month"
          className="flex items-center justify-center size-7 rounded-lg text-(--text-muted) hover:text-(--fg) hover:bg-(--card-alt) transition-colors cursor-pointer"
        >
          <ChevronRight size={15} />
        </button>
      </div>

      <div className="grid grid-cols-7 gap-1 mb-1">
        {WEEKDAY_LABELS.map((label, i) => (
          <div key={i} className="text-center text-[length:var(--text-pill)] text-(--text-faint)">
            {label}
          </div>
        ))}
      </div>

      <div className="grid grid-cols-7 gap-1">
        {cells.map((day, i) => {
          if (day === null) return <div key={`blank-${i}`} />;

          const dateKey = monthDayKey(year, monthIndex, day);
          const worked = workedMinutesByDate[dateKey] ?? 0;
          const intensity = worked > 0 ? Math.min(1, worked / maxWorked) : 0;
          const isToday = dateKey === todayKey;
          const isSelected = dateKey === selectedKey;

          return (
            <button
              key={dateKey}
              type="button"
              onClick={() => onSelect(dateKey)}
              title={worked > 0 ? formatMinutes(worked) : undefined}
              style={{
                backgroundColor:
                  worked > 0
                    ? `color-mix(in oklab, var(--gold) ${Math.round(20 + intensity * 60)}%, var(--card-alt))`
                    : "var(--card-alt)",
              }}
              className={`aspect-square rounded-lg flex items-center justify-center text-[length:var(--text-pill)] transition-shadow cursor-pointer ${
                isSelected ? "ring-1 ring-(--fg)" : ""
              } ${isToday ? "text-(--fg) font-medium" : "text-(--text-muted)"}`}
            >
              {day}
            </button>
          );
        })}
      </div>
    </div>
  );
}
