export type SessionStatus = "running" | "completed" | "cancelled";

export interface WorkSession {
  id: string;
  plannedMinutes: number;
  startedAt: string;
  endedAt: string | null;
  status: SessionStatus;
  note: string | null;
  actualMinutes: number | null;
}

// One IST calendar day's worked minutes — see server/internal/worksession's
// Service.DailySummary for how a session crossing midnight IST is split
// between the two days it touched.
export interface DailySummary {
  date: string;
  workedMinutes: number;
}
