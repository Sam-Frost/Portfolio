export type SessionStatus = "running" | "completed" | "cancelled";

// Deliberately just these two — the daily bar chart colours them separately.
export type SessionCategory = "professional" | "personal";

// One checklist bullet: text is set when the session starts, done is ticked
// (or not) when it ends.
export interface Goal {
  text: string;
  done: boolean;
}

export interface WorkSession {
  id: string;
  plannedMinutes: number;
  category: SessionCategory;
  startedAt: string;
  endedAt: string | null;
  status: SessionStatus;
  goals: Goal[] | null;
  startNote: string | null;
  note: string | null;
  actualMinutes: number | null;
}

// What the complete/cancel endpoints accept: the goals with their done
// flags ticked and a closing remark.
export interface FinishPayload {
  goals: Goal[];
  note: string;
}

// One IST calendar day's worked minutes — split by category (the two
// always sum to workedMinutes). See server/internal/worksession's
// Service.DailySummary for how a session crossing midnight IST is split
// between the two days it touched.
export interface DailySummary {
  date: string;
  workedMinutes: number;
  professionalMinutes: number;
  personalMinutes: number;
}
