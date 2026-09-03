import { useEffect, useState, type KeyboardEvent } from "react";
import { Bell, Repeat, Trash2 } from "lucide-react";
import type { Todo } from "../types";
import { createReminder, deleteReminder, fetchReminders } from "./api";
import type { Reminder, ReminderKind } from "./types";

interface Props {
  todo: Todo;
  onClose: () => void;
  /** Bubbled up so the list can refresh its "has reminders" indicator. */
  onChange?: () => void;
}

const QUICK_OFFSETS: { label: string; minutes: number }[] = [
  { label: "+15 min", minutes: 15 },
  { label: "+1 hour", minutes: 60 },
  { label: "+3 hours", minutes: 180 },
  { label: "+1 day", minutes: 60 * 24 },
  { label: "+2 days", minutes: 60 * 24 * 2 },
];

const REPEAT_UNITS: { label: string; seconds: number }[] = [
  { label: "minutes", seconds: 60 },
  { label: "hours", seconds: 3600 },
  { label: "days", seconds: 86400 },
];

function toLocalInputValue(d: Date): string {
  const pad = (n: number) => String(n).padStart(2, "0");
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`;
}

function describe(r: Reminder): string {
  const when = new Date(r.fireAt).toLocaleString(undefined, {
    month: "short",
    day: "numeric",
    hour: "numeric",
    minute: "2-digit",
  });
  if (r.kind === "once") return `One-time · ${when}`;
  const secs = r.intervalSeconds ?? 0;
  const unit =
    secs % 86400 === 0 ? `${secs / 86400} day` : secs % 3600 === 0 ? `${secs / 3600} hour` : `${secs / 60} min`;
  const plural = unit.startsWith("1 ") ? unit : `${unit}s`;
  return `Every ${plural} · next ${when}`;
}

export function ReminderDialog({ todo, onClose, onChange }: Props) {
  const [reminders, setReminders] = useState<Reminder[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);

  const [kind, setKind] = useState<ReminderKind>("once");
  const [fireAt, setFireAt] = useState(() => toLocalInputValue(new Date(Date.now() + 15 * 60_000)));
  const [repeatN, setRepeatN] = useState(2);
  const [repeatUnit, setRepeatUnit] = useState(REPEAT_UNITS[1]); // hours

  function refresh() {
    fetchReminders(todo.id)
      .then(setReminders)
      .catch((e) => setError(e instanceof Error ? e.message : "Couldn't load reminders."))
      .finally(() => setLoading(false));
  }

  useEffect(refresh, [todo.id]);

  function applyOffset(minutes: number) {
    setFireAt(toLocalInputValue(new Date(Date.now() + minutes * 60_000)));
  }

  async function add() {
    setSaving(true);
    setError(null);
    try {
      if (kind === "once") {
        const iso = fireAt ? new Date(fireAt).toISOString() : "";
        if (!iso || new Date(iso).getTime() <= Date.now()) {
          setError("Pick a time in the future.");
          return;
        }
        await createReminder(todo.id, { kind: "once", fireAt: iso });
      } else {
        const intervalSeconds = Math.round(repeatN * repeatUnit.seconds);
        if (intervalSeconds < 60) {
          setError("Repeat interval must be at least 1 minute.");
          return;
        }
        await createReminder(todo.id, { kind: "repeat", intervalSeconds });
      }
      refresh();
      onChange?.();
    } catch (e) {
      setError(e instanceof Error ? e.message : "Couldn't add the reminder.");
    } finally {
      setSaving(false);
    }
  }

  async function remove(id: string) {
    try {
      await deleteReminder(id);
      setReminders((prev) => prev.filter((r) => r.id !== id));
      onChange?.();
    } catch (e) {
      setError(e instanceof Error ? e.message : "Couldn't delete the reminder.");
    }
  }

  const tabCls = (active: boolean) =>
    `flex-1 flex items-center justify-center gap-1.5 rounded-lg px-2.5 py-1.5 text-[length:var(--text-pill)] transition-colors cursor-pointer ${
      active ? "bg-(--fg) text-(--bg)" : "bg-(--card-alt) text-(--text-muted) hover:text-(--fg)"
    }`;
  const fieldCls =
    "rounded-lg bg-(--card-alt) border-(--line) border-[0.5px] border-solid px-2.5 py-1.5 text-[length:var(--text-caption)] text-(--fg) focus:outline-none";

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 px-4"
      onClick={onClose}
      onKeyDown={(e: KeyboardEvent<HTMLDivElement>) => e.key === "Escape" && onClose()}
      role="presentation"
    >
      <div
        className="w-full max-w-sm max-h-[90vh] overflow-y-auto themed-scrollbar rounded-xl border-(--line) border-[0.5px] border-solid bg-(--card) p-4 shadow-lg"
        onClick={(e) => e.stopPropagation()}
        role="dialog"
        aria-modal="true"
        aria-label="Reminders"
      >
        <h2 className="flex items-center gap-1.5 text-[length:var(--text-caption)] text-(--fg) mb-1">
          <Bell size={14} /> Reminders
        </h2>
        <p className="text-[length:var(--text-pill)] text-(--text-faint) mb-3 truncate">{todo.name}</p>

        {loading ? (
          <p className="text-[length:var(--text-pill)] text-(--text-faint) mb-3">Loading…</p>
        ) : reminders.length > 0 ? (
          <ul className="flex flex-col gap-1.5 mb-4">
            {reminders.map((r) => (
              <li
                key={r.id}
                className="flex items-center gap-2 rounded-lg bg-(--card-alt) border-(--line) border-[0.5px] border-solid px-2.5 py-1.5"
              >
                {r.kind === "repeat" ? <Repeat size={12} /> : <Bell size={12} />}
                <span className="flex-1 min-w-0 truncate text-[length:var(--text-pill)] text-(--fg)">
                  {describe(r)}
                </span>
                <button
                  type="button"
                  onClick={() => remove(r.id)}
                  aria-label="Delete reminder"
                  className="text-(--text-faint) hover:text-red-400 transition-colors cursor-pointer"
                >
                  <Trash2 size={13} />
                </button>
              </li>
            ))}
          </ul>
        ) : (
          <p className="text-[length:var(--text-pill)] text-(--text-faint) mb-4">No reminders yet.</p>
        )}

        <div className="flex gap-1.5 mb-3">
          <button type="button" className={tabCls(kind === "once")} onClick={() => setKind("once")}>
            <Bell size={12} /> One-time
          </button>
          <button type="button" className={tabCls(kind === "repeat")} onClick={() => setKind("repeat")}>
            <Repeat size={12} /> Repeating
          </button>
        </div>

        {kind === "once" ? (
          <div className="flex flex-col gap-2 mb-4">
            <div className="flex flex-wrap gap-1.5">
              {QUICK_OFFSETS.map((o) => (
                <button
                  key={o.label}
                  type="button"
                  onClick={() => applyOffset(o.minutes)}
                  className="rounded-md bg-(--card-alt) border-(--line) border-[0.5px] border-solid px-2 py-1 text-[length:var(--text-pill)] text-(--text-muted) hover:text-(--fg) hover:border-(--fg) transition-colors cursor-pointer"
                >
                  {o.label}
                </button>
              ))}
            </div>
            <input
              type="datetime-local"
              value={fireAt}
              onChange={(e) => setFireAt(e.target.value)}
              className={fieldCls}
            />
          </div>
        ) : (
          <div className="flex items-center gap-2 mb-4">
            <span className="text-[length:var(--text-pill)] text-(--text-muted)">Every</span>
            <input
              type="number"
              min={1}
              value={repeatN}
              onChange={(e) => setRepeatN(Math.max(1, Number(e.target.value) || 1))}
              className={`${fieldCls} w-16`}
            />
            <select
              value={repeatUnit.label}
              onChange={(e) => setRepeatUnit(REPEAT_UNITS.find((u) => u.label === e.target.value)!)}
              className={fieldCls}
            >
              {REPEAT_UNITS.map((u) => (
                <option key={u.label} value={u.label}>
                  {u.label}
                </option>
              ))}
            </select>
          </div>
        )}

        {error && <p className="text-[length:var(--text-pill)] text-red-400 mb-3">{error}</p>}

        <div className="flex items-center justify-end gap-1.5">
          <button
            type="button"
            onClick={onClose}
            className="rounded-md px-2.5 py-1 text-[length:var(--text-pill)] text-(--text-muted) hover:text-(--fg) hover:bg-(--card-alt) transition-colors cursor-pointer"
          >
            Close
          </button>
          <button
            type="button"
            onClick={add}
            disabled={saving}
            className={`rounded-md bg-(--fg) text-(--bg) px-2.5 py-1 text-[length:var(--text-pill)] ${
              saving ? "opacity-60 cursor-not-allowed" : "cursor-pointer"
            }`}
          >
            Add reminder
          </button>
        </div>
      </div>
    </div>
  );
}
