// IST-aware date helpers for the diary calendar. A diary entry's identity
// and edit-lock deadline are both defined in terms of the IST calendar day
// (see server/internal/diary/model.go's IsLocked) — never the browser's own
// timezone, which may not be IST.

const IST_TIME_ZONE = "Asia/Kolkata";

// Formats `date` (an absolute instant) as its IST calendar date,
// "YYYY-MM-DD" — matches the backend's diary.EntryDateLayout.
export function toISTDateString(date: Date): string {
  const parts = new Intl.DateTimeFormat("en-CA", {
    timeZone: IST_TIME_ZONE,
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
  }).formatToParts(date);
  const lookup = Object.fromEntries(parts.map((p) => [p.type, p.value])) as Record<string, string>;
  return `${lookup.year}-${lookup.month}-${lookup.day}`;
}

export function todayIST(): string {
  return toISTDateString(new Date());
}

// Mirrors the backend's diary.IsLocked exactly: an entry becomes uneditable
// once the calendar day it belongs to (IST) has been over for 24 hours —
// i.e. at IST midnight two days after entryDate (2026-08-20 locks at
// 2026-08-22T00:00:00+05:30). This is a pure function of the date and the
// current instant, no server round trip needed — used to disable the
// editor early; the server enforces the same rule as the real guard.
export function isDiaryDateLocked(entryDate: string, now: Date = new Date()): boolean {
  const [year, month, day] = entryDate.split("-").map(Number);
  // IST is a fixed UTC+5:30 offset (no DST), so "midnight IST on day D+2"
  // is exactly this UTC instant — Date.UTC normalizes the out-of-range
  // hour/minute/day arguments for us.
  const lockInstantMs = Date.UTC(year, month - 1, day + 2, -5, -30);
  return now.getTime() >= lockInstantMs;
}
