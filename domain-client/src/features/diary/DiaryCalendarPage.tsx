import { useEffect, useMemo, useState } from "react";
import { useNavigate } from "react-router-dom";
import { ChevronLeft, ChevronRight, CalendarDays } from "lucide-react";
import { fetchDiaryDates } from "./api";
import { CalendarGrid } from "./CalendarGrid";
import { todayIST } from "./dateUtils";

const MONTH_LABEL_FORMAT = new Intl.DateTimeFormat(undefined, { month: "long", year: "numeric" });

function pad(n: number) {
  return String(n).padStart(2, "0");
}

function monthRange(year: number, month: number) {
  const from = `${year}-${pad(month + 1)}-01`;
  const lastDay = new Date(year, month + 1, 0).getDate();
  const to = `${year}-${pad(month + 1)}-${pad(lastDay)}`;
  return { from, to };
}

export function DiaryCalendarPage() {
  const navigate = useNavigate();
  const today = todayIST();

  const [viewYear, viewMonth0] = useMemo(() => {
    const [y, m] = today.split("-").map(Number);
    return [y, m - 1] as const;
  }, [today]);

  const [year, setYear] = useState(viewYear);
  const [month, setMonth] = useState(viewMonth0); // 0-11

  const [entryDates, setEntryDates] = useState<Set<string>>(new Set());
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const { from, to } = monthRange(year, month);
    setLoading(true);
    fetchDiaryDates(from, to)
      .then((dates) => setEntryDates(new Set(dates)))
      .catch((err) => setError(err instanceof Error ? err.message : "Couldn't load diary entries."))
      .finally(() => setLoading(false));
  }, [year, month]);

  function goToPrevMonth() {
    if (month === 0) {
      setYear((y) => y - 1);
      setMonth(11);
    } else {
      setMonth((m) => m - 1);
    }
  }

  function goToNextMonth() {
    if (month === 11) {
      setYear((y) => y + 1);
      setMonth(0);
    } else {
      setMonth((m) => m + 1);
    }
  }

  const label = MONTH_LABEL_FORMAT.format(new Date(year, month, 1));

  return (
    <div className="h-full flex flex-col">
      <div className="shrink-0 flex items-center justify-between mb-4">
        <h1 className="text-xl font-space font-semibold text-(--fg)">Personal Diary</h1>
        <button
          onClick={() => navigate(`/diary/${today}`)}
          className="flex items-center gap-1.5 rounded-lg bg-(--fg) text-(--bg) px-3 py-1.5 text-[length:var(--text-caption)] cursor-pointer transition-opacity hover:opacity-90"
        >
          <CalendarDays size={14} />
          Today's entry
        </button>
      </div>

      <div className="shrink-0 flex items-center justify-between mb-3">
        <button
          type="button"
          onClick={goToPrevMonth}
          aria-label="Previous month"
          className="flex items-center justify-center size-7 rounded-lg text-(--text-muted) hover:text-(--fg) hover:bg-(--card-alt) transition-colors cursor-pointer"
        >
          <ChevronLeft size={16} />
        </button>
        <span className="text-[length:var(--text-caption)] text-(--fg) font-medium">{label}</span>
        <button
          type="button"
          onClick={goToNextMonth}
          aria-label="Next month"
          className="flex items-center justify-center size-7 rounded-lg text-(--text-muted) hover:text-(--fg) hover:bg-(--card-alt) transition-colors cursor-pointer"
        >
          <ChevronRight size={16} />
        </button>
      </div>

      {error && (
        <div className="shrink-0 mb-3 flex items-center justify-between gap-2 rounded-lg border-(--line) border-[0.5px] border-solid bg-(--card) px-3 py-2 text-[length:var(--text-pill)] text-(--label-red)">
          <span>{error}</span>
          <button onClick={() => setError(null)} className="shrink-0 text-(--text-faint) hover:text-(--fg) transition-colors cursor-pointer">
            Dismiss
          </button>
        </div>
      )}

      <div className="flex-1 min-h-0 overflow-y-auto themed-scrollbar">
        <CalendarGrid
          year={year}
          month={month}
          entryDates={entryDates}
          todayDate={today}
          loading={loading}
          onSelectDate={(date) => navigate(`/diary/${date}`)}
        />
      </div>
    </div>
  );
}
