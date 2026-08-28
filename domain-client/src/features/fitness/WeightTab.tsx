import { useEffect, useState } from "react";
import { Trash2 } from "lucide-react";
import { deleteWeightLog, fetchWeightLogs, upsertWeightLog } from "./api";
import { LogValueForm } from "./LogValueForm";
import { LineChart } from "./LineChart";
import { formatDayKey } from "./dateUtils";
import type { Cycle, WeightLog } from "./types";

interface WeightTabProps {
  cycle: Cycle;
  onError: (message: string) => void;
}

// Start → target progress in a single box: a track with `start` on the left and
// `target` on the right, and a cursor sliding left-to-right as weight comes off.
// The cursor carries the weight lost; the weight still to go sits below.
function ProgressBox({
  start,
  target,
  latest,
}: {
  start: number | null;
  target: number | null;
  latest: number | null;
}) {
  const hasRange = start != null && target != null && start !== target;
  const lost = start != null && latest != null ? round1(start - latest) : null;
  const remaining = latest != null && target != null ? round1(latest - target) : null;
  const pct =
    hasRange && latest != null
      ? Math.min(100, Math.max(0, ((start - latest) / (start - target)) * 100))
      : 0;

  return (
    <div className="bg-(--card) border-(--line) border-[0.5px] border-solid rounded-xl p-4 flex flex-col gap-2">
      <div className="flex items-start justify-between">
        <div className="flex flex-col gap-0.5">
          <span className="text-[length:var(--text-pill)] text-(--text-faint) uppercase tracking-wide">Start</span>
          <span className="text-[length:var(--text-caption)] text-(--fg) font-medium">
            {start != null ? `${round1(start)} kg` : "—"}
          </span>
        </div>
        <div className="flex flex-col items-end gap-0.5">
          <span className="text-[length:var(--text-pill)] text-(--text-faint) uppercase tracking-wide">Target</span>
          <span className="text-[length:var(--text-caption)] text-(--fg) font-medium">
            {target != null ? `${round1(target)} kg` : "—"}
          </span>
        </div>
      </div>

      <div className="relative mx-1 mt-7 mb-1 h-1.5 rounded-full bg-(--card-alt)">
        <div
          className="absolute inset-y-0 left-0 rounded-full bg-(--green) transition-[width] duration-500 ease-out"
          style={{ width: `${pct}%` }}
        />
        <div
          className="absolute top-1/2 flex -translate-x-1/2 -translate-y-1/2 flex-col items-center gap-1.5 transition-[left] duration-500 ease-out"
          style={{ left: `${pct}%` }}
        >
          <span className="-mt-0.5 whitespace-nowrap rounded-md bg-(--green) px-1.5 py-0.5 text-[length:var(--text-pill)] font-medium text-(--bg)">
            {lost != null && lost > 0 ? `${lost} kg lost` : "0 kg lost"}
          </span>
          <span className="size-3 rounded-full border-2 border-(--card) bg-(--green)" />
        </div>
      </div>

      <p className="text-center text-[length:var(--text-pill)] text-(--text-muted)">
        {remaining == null
          ? "—"
          : remaining > 0
            ? `${remaining} kg remaining`
            : "Target reached 🎉"}
      </p>
    </div>
  );
}

export function WeightTab({ cycle, onError }: WeightTabProps) {
  const [logs, setLogs] = useState<WeightLog[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    setLoading(true);
    fetchWeightLogs(cycle.id)
      .then(setLogs)
      .catch((err) => onError(err instanceof Error ? err.message : "Couldn't load weigh-ins."))
      .finally(() => setLoading(false));
  }, [cycle.id, onError]);

  function handleLog(date: string, weight: number) {
    upsertWeightLog(cycle.id, date, weight)
      .then((saved) => {
        setLogs((prev) => {
          const rest = prev.filter((l) => l.logDate !== saved.logDate);
          return [...rest, saved].sort((a, b) => a.logDate.localeCompare(b.logDate));
        });
      })
      .catch((err) => onError(err instanceof Error ? err.message : "Couldn't save weigh-in."));
  }

  function handleDelete(id: string) {
    const original = logs;
    setLogs((prev) => prev.filter((l) => l.id !== id));
    deleteWeightLog(id).catch((err) => {
      setLogs(original);
      onError(err instanceof Error ? err.message : "Couldn't delete weigh-in.");
    });
  }

  const latest = logs.length > 0 ? logs[logs.length - 1].weight : cycle.weightStart;

  return (
    <div className="flex flex-col gap-4">
      <div className="grid gap-4 lg:grid-cols-[minmax(0,1fr)_300px]">
        <div className="flex flex-col gap-4">
          <ProgressBox start={cycle.weightStart} target={cycle.weightTarget} latest={latest} />
          <LogValueForm valueLabel="Weight" unit="kg" onSubmit={handleLog} submitLabel="Log weight" />
        </div>

        <div className="bg-(--card) border-(--line) border-[0.5px] border-solid rounded-xl p-4">
          <h2 className="text-[length:var(--text-pill)] font-medium uppercase tracking-wide text-(--text-faint) mb-3">
            Weigh-ins
          </h2>
          {loading ? (
            <p className="text-[length:var(--text-pill)] text-(--text-faint) py-4 text-center">Loading...</p>
          ) : logs.length === 0 ? (
            <p className="text-[length:var(--text-pill)] text-(--text-faint) py-4 text-center">Nothing logged yet.</p>
          ) : (
            <div className="flex flex-col gap-1 max-h-64 overflow-y-auto themed-scrollbar">
              {[...logs].reverse().map((l) => (
                <div key={l.id} className="group flex items-center justify-between gap-2 rounded-lg px-2 py-1.5 hover:bg-(--card-alt)">
                  <span className="text-[length:var(--text-pill)] text-(--text-muted)">{formatDayKey(l.logDate)}</span>
                  <span className="text-[length:var(--text-pill)] text-(--fg)">{round1(l.weight)} kg</span>
                  <button
                    onClick={() => handleDelete(l.id)}
                    aria-label="Delete weigh-in"
                    className="flex items-center justify-center size-6 rounded-md text-(--text-faint) opacity-0 group-hover:opacity-100 hover:text-(--fg) transition-opacity cursor-pointer"
                  >
                    <Trash2 size={12} />
                  </button>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>

      <LineChart
        title="Weight over time"
        points={logs.map((l) => ({ date: l.logDate, value: l.weight }))}
        unit="kg"
        target={cycle.weightTarget}
        emptyLabel="No weigh-ins logged yet."
      />
    </div>
  );
}

function round1(n: number): number {
  return Math.round(n * 10) / 10;
}
