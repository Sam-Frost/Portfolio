import { useEffect, useState } from "react";
import { Trash2 } from "lucide-react";
import { Link } from "react-router-dom";
import { createProteinLog, deleteProteinLog, fetchFoods, fetchProteinLogs } from "./api";
import { LogProteinForm } from "./LogProteinForm";
import { DailyBarChart } from "./DailyBarChart";
import { LineChart } from "./LineChart";
import { formatDayKey, todayISTKey } from "./dateUtils";
import type { Cycle, Food, ProteinLog } from "./types";

interface ProteinTabProps {
  cycle: Cycle;
  onError: (message: string) => void;
}

export function ProteinTab({ cycle, onError }: ProteinTabProps) {
  const [foods, setFoods] = useState<Food[]>([]);
  const [logs, setLogs] = useState<ProteinLog[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    setLoading(true);
    Promise.all([fetchFoods(), fetchProteinLogs(cycle.id)])
      .then(([f, l]) => {
        setFoods(f);
        setLogs(l);
      })
      .catch((err) => onError(err instanceof Error ? err.message : "Couldn't load protein data."))
      .finally(() => setLoading(false));
  }, [cycle.id, onError]);

  const foodById = new Map(foods.map((f) => [f.id, f]));

  const proteinByDate = new Map<string, number>();
  for (const l of logs) proteinByDate.set(l.logDate, (proteinByDate.get(l.logDate) ?? 0) + l.protein);
  const dailyTotals = [...proteinByDate.entries()]
    .sort(([a], [b]) => a.localeCompare(b))
    .map(([date, protein]) => ({ date, value: Math.round(protein * 10) / 10 }));

  const todayKey = todayISTKey();
  const todayTotal = dailyTotals.find((d) => d.date === todayKey)?.value ?? 0;

  function handleLogProtein(input: { foodId: string; date: string; quantity: number }) {
    createProteinLog(cycle.id, input)
      .then((log) => setLogs((prev) => [log, ...prev]))
      .catch((err) => onError(err instanceof Error ? err.message : "Couldn't log protein."));
  }

  function handleDeleteProteinLog(id: string) {
    const original = logs;
    setLogs((prev) => prev.filter((l) => l.id !== id));
    deleteProteinLog(id).catch((err) => {
      setLogs(original);
      onError(err instanceof Error ? err.message : "Couldn't delete entry.");
    });
  }

  const target = cycle.proteinTarget;
  const todayEntries = logs
    .filter((l) => l.logDate === todayKey)
    .sort((a, b) => b.createdAt.localeCompare(a.createdAt));

  if (loading) {
    return (
      <div className="bg-(--card) border-(--line) border-[0.5px] border-solid rounded-xl p-8 text-center text-(--text-faint) text-[length:var(--text-caption)]">
        Loading protein data...
      </div>
    );
  }

  const todayPct = target != null && target > 0 ? Math.min(todayTotal / target, 1) : 0;
  const targetHit = target != null && todayTotal >= target;

  return (
    <div className="flex flex-col gap-4">
      <div className="grid gap-4 lg:grid-cols-2 lg:items-start">
        {/* Left: today's protein, logging, and today's intake */}
        <div className="flex flex-col gap-4">
          <div className="rounded-lg bg-(--card-alt) px-3 py-2.5">
            <div className="flex items-baseline justify-between gap-2">
              <span className="text-[length:var(--text-pill)] text-(--text-faint) uppercase tracking-wide">Today's protein</span>
              <span className="text-[length:var(--text-caption)] text-(--fg) font-medium">
                {round1(todayTotal)} g
                {target != null ? (
                  <span className="text-(--text-faint)">
                    {" "}
                    / {target} · {Math.round((todayTotal / target) * 100)}%
                  </span>
                ) : (
                  <span className="text-(--text-faint)"> · no target set on the cycle</span>
                )}
              </span>
            </div>
            {target != null && (
              <div className="mt-2 h-2 rounded-full bg-(--ring-track) overflow-hidden">
                <div
                  className="h-full rounded-full transition-[width]"
                  style={{ width: `${todayPct * 100}%`, backgroundColor: targetHit ? "var(--green)" : "var(--gold)" }}
                />
              </div>
            )}
          </div>

          {foods.length === 0 ? (
            <div className="bg-(--card) border-(--line) border-[0.5px] border-solid rounded-xl p-4 text-[length:var(--text-pill)] text-(--text-faint)">
              No foods in your library yet — add them in{" "}
              <Link to="/settings" className="text-(--fg) underline hover:text-(--gold) transition-colors">
                Settings
              </Link>
              .
            </div>
          ) : (
            <LogProteinForm foods={foods} onSubmit={handleLogProtein} />
          )}

          {todayEntries.length > 0 && (
            <div className="bg-(--card) border-(--line) border-[0.5px] border-solid rounded-xl p-4">
              <h2 className="text-[length:var(--text-pill)] font-medium uppercase tracking-wide text-(--text-faint) mb-3">
                Today's intake
              </h2>
              <div className="flex flex-col gap-1">
                {todayEntries.map((l) => (
                  <div key={l.id} className="group flex items-center justify-between gap-2 rounded-lg px-2 py-1.5 hover:bg-(--card-alt)">
                    <span className="text-[length:var(--text-pill)] text-(--fg) truncate">
                      {foodById.get(l.foodId)?.name ?? "Food"}{" "}
                      <span className="text-(--text-faint)">
                        × {round1(l.quantity)} {foodById.get(l.foodId)?.unit ?? ""}
                      </span>
                    </span>
                    <div className="flex items-center gap-2 shrink-0">
                      <span className="text-[length:var(--text-pill)] text-(--green)">{round1(l.protein)} g</span>
                      <button
                        onClick={() => handleDeleteProteinLog(l.id)}
                        aria-label="Delete entry"
                        className="flex items-center justify-center size-6 rounded-md text-(--text-faint) opacity-100 lg:opacity-0 lg:group-hover:opacity-100 hover:text-(--fg) transition-opacity cursor-pointer"
                      >
                        <Trash2 size={12} />
                      </button>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>

        {/* Right: intake vs target */}
        <DailyBarChart
          title="Intake vs target"
          points={dailyTotals}
          unit="g"
          target={target}
          emptyLabel="No protein logged yet."
        />
      </div>

      <LineChart
        title="Intake over time"
        points={dailyTotals}
        unit="g"
        target={target}
        emptyLabel="No protein logged yet."
      />

      {logs.length > 0 && (
        <div className="bg-(--card) border-(--line) border-[0.5px] border-solid rounded-xl p-4">
          <h2 className="text-[length:var(--text-pill)] font-medium uppercase tracking-wide text-(--text-faint) mb-3">
            All entries
          </h2>
          <div className="flex flex-col gap-1 max-h-72 overflow-y-auto themed-scrollbar">
            {logs.map((l) => (
              <div key={l.id} className="group flex items-center justify-between gap-2 rounded-lg px-2 py-1.5 hover:bg-(--card-alt)">
                <span className="text-[length:var(--text-pill)] text-(--text-muted) truncate">
                  {formatDayKey(l.logDate)} · {foodById.get(l.foodId)?.name ?? "Food"} × {round1(l.quantity)}
                </span>
                <div className="flex items-center gap-2 shrink-0">
                  <span className="text-[length:var(--text-pill)] text-(--fg)">{round1(l.protein)} g</span>
                  <button
                    onClick={() => handleDeleteProteinLog(l.id)}
                    aria-label="Delete entry"
                    className="flex items-center justify-center size-6 rounded-md text-(--text-faint) opacity-100 lg:opacity-0 lg:group-hover:opacity-100 hover:text-(--fg) transition-opacity cursor-pointer"
                  >
                    <Trash2 size={12} />
                  </button>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}

function round1(n: number): number {
  return Math.round(n * 10) / 10;
}
