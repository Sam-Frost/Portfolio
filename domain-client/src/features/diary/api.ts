import { apiRequest } from "../../lib/apiClient";
import type { DiaryEntry } from "./types";

export function fetchDiaryEntry(date: string): Promise<DiaryEntry> {
  return apiRequest<DiaryEntry>(`/api/diary/entries/${date}`);
}

// Upsert — the backend creates the day's entry if none exists yet,
// otherwise updates it in place ("one entry per day", edited over time).
// Rejected with a 409 once that date's 24-hour edit window has closed.
export function upsertDiaryEntry(date: string, content: string): Promise<DiaryEntry> {
  return apiRequest<DiaryEntry>(`/api/diary/entries/${date}`, {
    method: "PUT",
    body: JSON.stringify({ content }),
  });
}

// Which dates in [from, to] have an entry — for the calendar view. Content
// isn't included, so this stays cheap even for a full month.
export function fetchDiaryDates(from: string, to: string): Promise<string[]> {
  const params = new URLSearchParams({ from, to });
  return apiRequest<string[]>(`/api/diary/entries?${params.toString()}`);
}
