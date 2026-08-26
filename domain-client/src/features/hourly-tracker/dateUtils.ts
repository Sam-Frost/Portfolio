// IST (India Standard Time) has no DST and a fixed +05:30 offset, so every
// conversion here is plain arithmetic against that offset rather than a
// timezone library — matching the backend's own IST day-bucketing (see
// server/internal/worksession's ParseISTDayStart/splitByISTDay), so a
// session's day here always agrees with which day the server attributed it
// to, regardless of the viewer's own machine timezone.
const IST_OFFSET_MS = 5.5 * 60 * 60 * 1000;
const DAY_MS = 24 * 60 * 60 * 1000;

export function pad2(n: number): string {
  return String(n).padStart(2, "0");
}

// "YYYY-MM-DD" for the IST calendar day an ISO instant falls on.
export function istDateKey(iso: string): string {
  const d = new Date(new Date(iso).getTime() + IST_OFFSET_MS);
  return `${d.getUTCFullYear()}-${pad2(d.getUTCMonth() + 1)}-${pad2(d.getUTCDate())}`;
}

// UTC instant (ms since epoch) of 00:00 IST on the given "YYYY-MM-DD" key.
function istDayStartMs(dateKey: string): number {
  const [y, m, d] = dateKey.split("-").map(Number);
  return Date.UTC(y, m - 1, d) - IST_OFFSET_MS;
}

// Whether a session's [startedAt, endedAt ?? now) span overlaps the IST
// calendar day identified by dateKey — a cross-midnight session touches
// (and so should appear under) both days it spans.
export function sessionTouchesDay(session: { startedAt: string; endedAt: string | null }, dateKey: string): boolean {
  const dayStart = istDayStartMs(dateKey);
  const dayEnd = dayStart + DAY_MS;
  const start = new Date(session.startedAt).getTime();
  const end = session.endedAt ? new Date(session.endedAt).getTime() : Date.now();
  return start < dayEnd && end > dayStart;
}

export function daysInMonth(year: number, monthIndex: number): number {
  return new Date(Date.UTC(year, monthIndex + 1, 0)).getUTCDate();
}

// 0 (Sunday) .. 6 (Saturday) that the 1st of the given month falls on.
export function firstWeekdayOfMonth(year: number, monthIndex: number): number {
  return new Date(Date.UTC(year, monthIndex, 1)).getUTCDay();
}

export function monthDayKey(year: number, monthIndex: number, day: number): string {
  return `${year}-${pad2(monthIndex + 1)}-${pad2(day)}`;
}

export function formatMinutes(totalMinutes: number): string {
  const h = Math.floor(totalMinutes / 60);
  const m = totalMinutes % 60;
  if (h === 0) return `${m}m`;
  if (m === 0) return `${h}h`;
  return `${h}h ${m}m`;
}

export function formatClock(totalSeconds: number): string {
  const m = Math.floor(totalSeconds / 60);
  const s = totalSeconds % 60;
  return `${m}:${pad2(s)}`;
}
