import { apiRequest } from "../../lib/apiClient";
import type { DailySummary, FinishPayload, SessionCategory, WorkSession } from "./types";

// undefined (not just null) on 204 — see apiRequest — when no session is
// currently running.
export function fetchCurrentSession(): Promise<WorkSession | undefined> {
  return apiRequest("/api/work-sessions/current");
}

export interface StartSessionInput {
  plannedMinutes: number;
  category: SessionCategory;
  goals: string[];
  startNote: string;
}

export function startSession(input: StartSessionInput): Promise<WorkSession> {
  return apiRequest("/api/work-sessions", {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export function completeSession(id: string, payload: FinishPayload): Promise<WorkSession> {
  return apiRequest(`/api/work-sessions/${id}/complete`, {
    method: "POST",
    body: JSON.stringify(payload),
  });
}

export function cancelSession(id: string, payload: FinishPayload): Promise<WorkSession> {
  return apiRequest(`/api/work-sessions/${id}/cancel`, {
    method: "POST",
    body: JSON.stringify(payload),
  });
}

// from/to are "YYYY-MM-DD" IST calendar dates, both inclusive.
export function fetchSessionsInRange(from: string, to: string): Promise<WorkSession[]> {
  return apiRequest(`/api/work-sessions?from=${from}&to=${to}`);
}

export function fetchDailySummary(from: string, to: string): Promise<DailySummary[]> {
  return apiRequest(`/api/work-sessions/daily-summary?from=${from}&to=${to}`);
}
