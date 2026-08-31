import { Check } from "lucide-react";
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

const CATEGORY_LABEL: Record<WorkSession["category"], string> = {
  professional: "Professional",
  personal: "Personal",
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
          {sessions.map((s) => {
            const goals = s.goals ?? [];
            return (
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
                  {CATEGORY_LABEL[s.category]} · Planned {formatMinutes(s.plannedMinutes)}
                  {s.actualMinutes !== null ? ` · Actual ${formatMinutes(s.actualMinutes)}` : ""}
                </p>

                {goals.length > 0 && (
                  <ul className="flex flex-col gap-0.5 my-1.5">
                    {goals.map((g, i) => (
                      <li key={i} className="flex items-center gap-1.5 text-[length:var(--text-pill)]">
                        <span
                          className={`shrink-0 flex h-3 w-3 items-center justify-center rounded-[3px] border-[0.5px] border-solid ${
                            g.done ? "bg-(--green) border-(--green) text-(--bg)" : "border-(--line-strong) text-transparent"
                          }`}
                        >
                          <Check size={9} strokeWidth={3} />
                        </span>
                        <span className={g.done ? "text-(--text-faint) line-through" : "text-(--text-muted)"}>
                          {g.text}
                        </span>
                      </li>
                    ))}
                  </ul>
                )}

                {s.startNote && (
                  <p className="text-[length:var(--text-pill)] text-(--text-faint) mb-0.5">
                    <span className="text-(--text-muted)">Start:</span> {s.startNote}
                  </p>
                )}
                {s.note && <p className="text-[length:var(--text-caption)] text-(--fg)">{s.note}</p>}
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
}
