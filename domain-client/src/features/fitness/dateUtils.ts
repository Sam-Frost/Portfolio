// IST (India Standard Time): fixed +05:30, no DST — every calendar-day
// field in the fitness feature is an IST day, matching the backend's
// DateLayout convention and features/hourly-tracker/dateUtils.ts.
const IST_OFFSET_MS = 5.5 * 60 * 60 * 1000;
const DAY_MS = 24 * 60 * 60 * 1000;

export function pad2(n: number): string {
  return String(n).padStart(2, "0");
}

// "YYYY-MM-DD" for the IST calendar day an ISO instant (or now) falls on.
export function istDateKey(iso?: string): string {
  const base = iso ? new Date(iso).getTime() : Date.now();
  const d = new Date(base + IST_OFFSET_MS);
  return `${d.getUTCFullYear()}-${pad2(d.getUTCMonth() + 1)}-${pad2(d.getUTCDate())}`;
}

export function todayISTKey(): string {
  return istDateKey();
}

// Whole days from today (IST) until dateKey. Negative if the date has
// passed, 0 if it's today.
export function daysUntil(dateKey: string): number {
  const [y, m, d] = dateKey.split("-").map(Number);
  const target = Date.UTC(y, m - 1, d);
  const [ty, tm, td] = todayISTKey().split("-").map(Number);
  const today = Date.UTC(ty, tm - 1, td);
  return Math.round((target - today) / DAY_MS);
}

// "Mon 12 Aug" style short label for a "YYYY-MM-DD" key.
export function formatDayKey(dateKey: string): string {
  const [y, m, d] = dateKey.split("-").map(Number);
  return new Date(Date.UTC(y, m - 1, d)).toLocaleDateString(undefined, {
    weekday: "short",
    day: "numeric",
    month: "short",
    timeZone: "UTC",
  });
}

// "12 Aug" — compact form for chart axis ticks.
export function formatAxisDate(dateKey: string): string {
  const [y, m, d] = dateKey.split("-").map(Number);
  return new Date(Date.UTC(y, m - 1, d)).toLocaleDateString(undefined, {
    day: "numeric",
    month: "short",
    timeZone: "UTC",
  });
}
