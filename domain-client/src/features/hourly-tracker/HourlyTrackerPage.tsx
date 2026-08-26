import { useEffect, useState } from "react";
import { fetchDailySummary, fetchSessionsInRange } from "./api";
import { DailyHoursChart } from "./DailyHoursChart";
import { DayDetailPanel } from "./DayDetailPanel";
import { daysInMonth, istDateKey, monthDayKey, pad2, sessionTouchesDay } from "./dateUtils";
import { MonthCalendar } from "./MonthCalendar";
import { StartSessionForm } from "./StartSessionForm";
import { useHourlyTracker } from "./HourlyTrackerProvider";
import type { DailySummary, WorkSession } from "./types";

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

function monthRange(year: number, monthIndex: number): { from: string; to: string } {
  const from = `${year}-${pad2(monthIndex + 1)}-01`;
  const to = `${year}-${pad2(monthIndex + 1)}-${pad2(daysInMonth(year, monthIndex))}`;
  return { from, to };
}

export function HourlyTrackerPage() {
  const { session, remainingSeconds, start } = useHourlyTracker();

  const todayKey = istDateKey(new Date().toISOString());
  const [todayYear, todayMonth] = todayKey.split("-").map(Number);

  const [viewYear, setViewYear] = useState(todayYear);
  const [viewMonth, setViewMonth] = useState(todayMonth - 1);
  const [sessions, setSessions] = useState<WorkSession[]>([]);
  const [summaries, setSummaries] = useState<DailySummary[]>([]);
  const [selectedDate, setSelectedDate] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Refetches on session too — not just the visible month — so the
  // calendar/chart pick up a session the moment it finishes, without
  // requiring a manual reload.
  useEffect(() => {
    const { from, to } = monthRange(viewYear, viewMonth);

    setLoading(true);
    Promise.all([fetchSessionsInRange(from, to), fetchDailySummary(from, to)])
      .then(([s, d]) => {
        setSessions(s);
        setSummaries(d);
      })
      .catch((err) => setError(err instanceof Error ? err.message : "Couldn't load sessions."))
      .finally(() => setLoading(false));
  }, [viewYear, viewMonth, session]);

  const monthLabel = `${MONTH_LABELS[viewMonth]} ${viewYear}`;

  const workedMinutesByDate: Record<string, number> = {};
  for (const s of summaries) workedMinutesByDate[s.date] = s.workedMinutes;

  const lastDay = daysInMonth(viewYear, viewMonth);
  const denseSummaries: DailySummary[] = Array.from({ length: lastDay }, (_, i) => {
    const key = monthDayKey(viewYear, viewMonth, i + 1);
    return { date: key, workedMinutes: workedMinutesByDate[key] ?? 0 };
  });

  const selectedSessions = selectedDate ? sessions.filter((s) => sessionTouchesDay(s, selectedDate)) : [];

  function goToPrevMonth() {
    setSelectedDate(null);
    if (viewMonth === 0) {
      setViewYear((y) => y - 1);
      setViewMonth(11);
    } else {
      setViewMonth((m) => m - 1);
    }
  }

  function goToNextMonth() {
    setSelectedDate(null);
    if (viewMonth === 11) {
      setViewYear((y) => y + 1);
      setViewMonth(0);
    } else {
      setViewMonth((m) => m + 1);
    }
  }

  return (
    <div className="h-full overflow-y-auto themed-scrollbar pr-2">
      <div className="flex flex-col gap-4 pb-4">
        {error && (
          <div className="rounded-lg border-(--line) border-[0.5px] border-solid bg-(--card) px-3 py-2 text-[length:var(--text-pill)] text-red-400">
            {error}
          </div>
        )}

        {!session ? (
          <StartSessionForm onStart={start} />
        ) : (
          <div className="bg-(--card) border-(--line) border-[0.5px] border-solid rounded-xl p-4 text-[length:var(--text-pill)] text-(--text-muted)">
            A session is already running — see the floating timer ({Math.floor(remainingSeconds / 60)}m remaining).
          </div>
        )}

        <div className="grid grid-cols-1 lg:grid-cols-[minmax(0,1fr)_18rem] gap-4">
          <div className="flex flex-col gap-4 min-w-0">
            <MonthCalendar
              year={viewYear}
              monthIndex={viewMonth}
              todayKey={todayKey}
              selectedKey={selectedDate}
              workedMinutesByDate={workedMinutesByDate}
              onSelect={setSelectedDate}
              onPrevMonth={goToPrevMonth}
              onNextMonth={goToNextMonth}
            />
            <DailyHoursChart summaries={denseSummaries} monthLabel={monthLabel} />
          </div>

          <DayDetailPanel dateKey={selectedDate} sessions={selectedSessions} />
        </div>

        {loading && <p className="text-[length:var(--text-pill)] text-(--text-faint)">Loading…</p>}
      </div>
    </div>
  );
}
