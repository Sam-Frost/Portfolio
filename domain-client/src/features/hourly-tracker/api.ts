import { apiRequest } from "../../lib/apiClient";
import type { DailySummary, WorkSession } from "./types";

// undefined (not just null) on 204 — see apiRequest — when no session is
// currently running.
export function fetchCurrentSession(): Promise<WorkSession | undefined> {
  return apiRequest("/api/work-sessions/current");
}

export function startSession(plannedMinutes: number): Promise<WorkSession> {
  return apiRequest("/api/work-sessions", {
    method: "POST",
    body: JSON.stringify({ plannedMinutes }),
  });
}

export function completeSession(id: string, note: string): Promise<WorkSession> {
  return apiRequest(`/api/work-sessions/${id}/complete`, {
    method: "POST",
    body: JSON.stringify({ note }),
  });
}

export function cancelSession(id: string, note?: string): Promise<WorkSession> {
  return apiRequest(`/api/work-sessions/${id}/cancel`, {
    method: "POST",
    body: JSON.stringify({ note: note ?? null }),
  });
}

// from/to are "YYYY-MM-DD" IST calendar dates, both inclusive.
export function fetchSessionsInRange(from: string, to: string): Promise<WorkSession[]> {
  return apiRequest(`/api/work-sessions?from=${from}&to=${to}`);
}

export function fetchDailySummary(from: string, to: string): Promise<DailySummary[]> {
  return apiRequest(`/api/work-sessions/daily-summary?from=${from}&to=${to}`);
}
