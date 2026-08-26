import { formatMinutes } from "./dateUtils";
import type { WorkSession } from "./types";

interface DayDetailPanelProps {
  dateKey: string | null;
  sessions: WorkSession[];
}

function formatTimeIST(iso: string): string {
  return new Date(iso).toLocaleTimeString("en-IN", {
    timeZone: "Asia/Kolkata",
    hour: "2-digit",
    minute: "2-digit",
  });
}

const STATUS_LABEL: Record<WorkSession["status"], string> = {
  running: "In progress",
  completed: "Completed",
  cancelled: "Cancelled",
};

// Sessions for whichever day is selected on the calendar. A session that
// crosses midnight IST is passed in twice — once per day it touches (see
// dateUtils.sessionTouchesDay) — by the page, so it shows up here on both.
export function DayDetailPanel({ dateKey, sessions }: DayDetailPanelProps) {
  if (!dateKey) {
    return (
      <div className="bg-(--card) border-(--line) border-[0.5px] border-solid rounded-xl p-4 text-[length:var(--text-pill)] text-(--text-faint)">
        Pick a day on the calendar to see its sessions.
      </div>
    );
  }

  const total = sessions.reduce((sum, s) => sum + (s.actualMinutes ?? 0), 0);

  return (
    <div className="bg-(--card) border-(--line) border-[0.5px] border-solid rounded-xl p-4">
      <div className="flex items-center justify-between mb-3">
        <h2 className="text-[length:var(--text-caption)] text-(--fg) font-medium">{dateKey}</h2>
        {total > 0 && <span className="text-[length:var(--text-pill)] text-(--text-faint)">{formatMinutes(total)} total</span>}
      </div>

      {sessions.length === 0 ? (
        <p className="text-[length:var(--text-pill)] text-(--text-faint)">No sessions on this day.</p>
      ) : (
        <div className="flex flex-col gap-2">
          {sessions.map((s) => (
            <div key={s.id} className="rounded-lg bg-(--card-alt) px-3 py-2">
              <div className="flex items-center justify-between gap-2 mb-1">
                <span className="text-[length:var(--text-pill)] text-(--text-muted)">
                  {formatTimeIST(s.startedAt)} – {s.endedAt ? formatTimeIST(s.endedAt) : "now"}
                </span>
                <span
                  className={`text-[length:var(--text-pill)] ${
                    s.status === "cancelled" ? "text-(--text-faint)" : "text-(--green-fg)"
                  }`}
                >
                  {STATUS_LABEL[s.status]}
                </span>
              </div>
              <p className="text-[length:var(--text-pill)] text-(--text-faint) mb-1">
                Planned {formatMinutes(s.plannedMinutes)}
                {s.actualMinutes !== null ? ` · Actual ${formatMinutes(s.actualMinutes)}` : ""}
              </p>
              {s.note && <p className="text-[length:var(--text-caption)] text-(--fg)">{s.note}</p>}
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
