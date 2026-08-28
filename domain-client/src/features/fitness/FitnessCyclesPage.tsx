import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { X } from "lucide-react";
import { activateCycle, createCycle, deleteCycle, fetchCycles } from "./api";
import { StartCycleForm } from "./StartCycleForm";
import { ConfirmDeleteDialog } from "./ConfirmDeleteDialog";
import type { Cycle, CycleInput } from "./types";

function formatDate(value: string) {
  const [y, m, d] = value.split("-").map(Number);
  return new Date(Date.UTC(y, m - 1, d)).toLocaleDateString(undefined, {
    year: "numeric",
    month: "short",
    day: "numeric",
    timeZone: "UTC",
  });
}

function round1(n: number): number {
  return Math.round(n * 10) / 10;
}

// One-line scannable summary: exercise count, and current -> target weight
// when both are known.
function summary(cycle: Cycle): string {
  const parts: string[] = [];
  parts.push(`${cycle.exerciseCount} ${cycle.exerciseCount === 1 ? "exercise" : "exercises"}`);
  const current = cycle.latestWeight ?? cycle.weightStart;
  if (current != null && cycle.weightTarget != null) {
    parts.push(`${round1(current)} → ${cycle.weightTarget} kg`);
  } else if (current != null) {
    parts.push(`${round1(current)} kg`);
  }
  return parts.join(" · ");
}

function CycleRow({
  cycle,
  onOpen,
  onDelete,
  onActivate,
}: {
  cycle: Cycle;
  onOpen: () => void;
  onDelete: () => void;
  onActivate: () => void;
}) {
  const [confirming, setConfirming] = useState(false);
  return (
    <div
      onClick={onOpen}
      role="button"
      tabIndex={0}
      className="group relative flex items-center justify-between gap-3 bg-(--card) border-(--line) border-[0.5px] border-solid rounded-xl px-4 py-3 cursor-pointer hover:border-(--line-strong) transition-colors"
    >
      <div className="min-w-0">
        <div className="flex items-center gap-2">
          <span className="truncate text-[length:var(--text-caption)] text-(--fg) font-medium">{cycle.name}</span>
          {cycle.status === "active" ? (
            <span className="shrink-0 rounded-full bg-(--green) text-(--bg) px-1.5 py-0.5 text-[9px] font-medium uppercase tracking-wide">
              Active
            </span>
          ) : (
            <span className="shrink-0 rounded-full bg-(--card-alt) text-(--text-faint) px-1.5 py-0.5 text-[9px] uppercase tracking-wide">
              Archived
            </span>
          )}
        </div>
        <span className="text-[length:var(--text-pill)] text-(--text-faint)">
          Started {formatDate(cycle.startDate)} · {summary(cycle)}
        </span>
      </div>

      <div className="flex items-center gap-1 shrink-0">
        {cycle.status === "archived" && (
          <button
            onClick={(e) => {
              e.stopPropagation();
              onActivate();
            }}
            className="rounded-md border-(--line) border-[0.5px] border-solid px-2 py-1 text-[9px] uppercase tracking-wide text-(--text-muted) opacity-0 group-hover:opacity-100 hover:text-(--fg) hover:bg-(--card-alt) transition-opacity cursor-pointer"
          >
            Make active
          </button>
        )}
        <button
          onClick={(e) => {
            e.stopPropagation();
            setConfirming(true);
          }}
          aria-label="Delete cycle"
          className="flex items-center justify-center size-6 rounded-md text-(--text-faint) opacity-0 group-hover:opacity-100 hover:text-(--fg) hover:bg-(--card-alt) transition-opacity cursor-pointer"
        >
          <X size={13} />
        </button>
      </div>

      {confirming && (
        <div onClick={(e) => e.stopPropagation()}>
          <ConfirmDeleteDialog
            title="Delete cycle?"
            message={`"${cycle.name}" and all its exercises, weigh-ins, foods and protein logs will be deleted.`}
            onCancel={() => setConfirming(false)}
            onConfirm={() => {
              onDelete();
              setConfirming(false);
            }}
          />
        </div>
      )}
    </div>
  );
}

// After activate/create the previously-active cycle becomes archived — apply
// that flip locally so the list doesn't briefly show two "Active" badges.
function withActivated(cycles: Cycle[], activeId: string): Cycle[] {
  return cycles.map((c) => {
    if (c.id === activeId) return { ...c, status: "active" as const, archivedAt: null };
    if (c.status === "active") return { ...c, status: "archived" as const };
    return c;
  });
}

export function FitnessCyclesPage() {
  const navigate = useNavigate();
  const [cycles, setCycles] = useState<Cycle[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    fetchCycles()
      .then(setCycles)
      .catch((err) => setError(err instanceof Error ? err.message : "Couldn't load cycles."))
      .finally(() => setLoading(false));
  }, []);

  function handleStart(input: CycleInput) {
    createCycle(input)
      .then((cycle) => {
        setCycles((prev) => [cycle, ...withActivated(prev, cycle.id)]);
        navigate(cycle.id);
      })
      .catch((err) => setError(err instanceof Error ? err.message : "Couldn't start cycle."));
  }

  function handleDelete(id: string) {
    const original = cycles;
    setCycles((prev) => prev.filter((c) => c.id !== id));
    deleteCycle(id).catch((err) => {
      setCycles(original);
      setError(err instanceof Error ? err.message : "Couldn't delete cycle.");
    });
  }

  function handleActivate(id: string) {
    const original = cycles;
    setCycles((prev) => withActivated(prev, id));
    activateCycle(id)
      .then((updated) => setCycles((prev) => prev.map((c) => (c.id === id ? updated : c))))
      .catch((err) => {
        setCycles(original);
        setError(err instanceof Error ? err.message : "Couldn't activate cycle.");
      });
  }

  return (
    <div className="h-full flex flex-col">
      <div className="shrink-0">
        <StartCycleForm onStart={handleStart} />

        {error && (
          <div className="mb-3 flex items-center justify-between gap-2 rounded-lg border-(--line) border-[0.5px] border-solid bg-(--card) px-3 py-2 text-[length:var(--text-pill)] text-red-400">
            <span>{error}</span>
            <button onClick={() => setError(null)} aria-label="Dismiss error" className="shrink-0 text-(--text-faint) hover:text-(--fg) transition-colors cursor-pointer">
              <X size={12} />
            </button>
          </div>
        )}
      </div>

      <div className="flex-1 min-h-0 overflow-y-auto themed-scrollbar">
        {loading && (
          <div className="bg-(--card) border-(--line) border-[0.5px] border-solid rounded-xl p-8 text-center text-(--text-faint) text-[length:var(--text-caption)]">
            Loading cycles...
          </div>
        )}

        {!loading && cycles.length === 0 && (
          <div className="bg-(--card) border-(--line) border-[0.5px] border-solid rounded-xl p-8 text-center text-(--text-faint) text-[length:var(--text-caption)]">
            No fitness cycle yet. Start one above to begin tracking exercise, weight and protein.
          </div>
        )}

        {!loading && cycles.length > 0 && (
          <div className="flex flex-col gap-1.5 pb-2">
            {cycles.map((cycle) => (
              <CycleRow
                key={cycle.id}
                cycle={cycle}
                onOpen={() => navigate(cycle.id)}
                onDelete={() => handleDelete(cycle.id)}
                onActivate={() => handleActivate(cycle.id)}
              />
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
