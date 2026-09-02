import { Video } from "lucide-react";
import { isDiaryDateLocked } from "./dateUtils";

interface CalendarGridProps {
  year: number;
  month: number; // 0-11
  entryDates: Set<string>;
  videoDates?: Set<string>; // days with at least one recorded clip
  todayDate: string; // "YYYY-MM-DD", IST
  loading?: boolean;
  onSelectDate: (date: string) => void;
}

const WEEKDAY_LABELS = ["Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"];

function pad(n: number) {
  return String(n).padStart(2, "0");
}

// A small, self-contained month grid — no shared calendar component exists
// elsewhere in the codebase yet. Days with an existing entry render
// distinctly from empty ones; today gets a ring so it's easy to find.
export function CalendarGrid({ year, month, entryDates, videoDates, todayDate, loading, onSelectDate }: CalendarGridProps) {
  const firstOfMonth = new Date(year, month, 1);
  const daysInMonth = new Date(year, month + 1, 0).getDate();
  const leadingBlanks = firstOfMonth.getDay(); // 0 = Sunday

  const cells: Array<{ day: number; date: string } | null> = [];
  for (let i = 0; i < leadingBlanks; i++) cells.push(null);
  for (let day = 1; day <= daysInMonth; day++) {
    cells.push({ day, date: `${year}-${pad(month + 1)}-${pad(day)}` });
  }

  return (
    <div className={`transition-opacity ${loading ? "opacity-60 pointer-events-none" : ""}`}>
      <div className="grid grid-cols-7 gap-1.5 mb-1.5">
        {WEEKDAY_LABELS.map((wd) => (
          <div
            key={wd}
            className="text-center text-[length:var(--text-pill)] text-(--text-faint) uppercase tracking-wide py-1"
          >
            {wd}
          </div>
        ))}
      </div>

      <div className="grid grid-cols-7 gap-1.5">
        {cells.map((cell, i) => {
          if (!cell) return <div key={`blank-${i}`} />;

          const hasEntry = entryDates.has(cell.date);
          const hasVideo = videoDates?.has(cell.date) ?? false;
          const hasContent = hasEntry || hasVideo;
          const isToday = cell.date === todayDate;
          const locked = isDiaryDateLocked(cell.date);

          const contentLabel = hasEntry && hasVideo ? "has an entry and a video" : hasEntry ? "has an entry" : hasVideo ? "has a video" : null;

          return (
            <button
              key={cell.date}
              type="button"
              onClick={() => onSelectDate(cell.date)}
              aria-label={contentLabel ? `${cell.date}, ${contentLabel}` : cell.date}
              className={`relative aspect-square rounded-lg border-[0.5px] border-solid text-[length:var(--text-caption)] flex flex-col items-center justify-center gap-1 transition-colors cursor-pointer ${
                hasContent
                  ? "bg-(--green-bg) border-(--green) text-(--green-fg) hover:border-(--green)"
                  : "bg-(--card) border-(--line) text-(--text-muted) hover:border-(--line-strong) hover:text-(--fg)"
              } ${isToday ? "ring-1 ring-(--gold) ring-offset-1 ring-offset-(--bg)" : ""}`}
            >
              <span>{cell.day}</span>
              {hasContent && (
                <span className="flex items-center gap-0.5">
                  {hasEntry && <span className="size-1 rounded-full bg-(--green)" />}
                  {hasVideo && <Video size={9} className="text-(--green)" />}
                </span>
              )}
              {!hasContent && locked && (
                <span className="absolute top-1.5 right-1.5 size-1 rounded-full bg-(--text-faint)" title="No entry — edit window closed" />
              )}
            </button>
          );
        })}
      </div>
    </div>
  );
}
