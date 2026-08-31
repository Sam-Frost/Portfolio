import { useEffect, useState } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import { ArrowLeft, Archive, RotateCcw, X } from "lucide-react";
import { ApiError } from "../../lib/apiClient";
import { activateCycle, archiveCycle, fetchCycle, updateCycle } from "./api";
import { ConfirmDeleteDialog } from "./ConfirmDeleteDialog";
import { EditCycleDialog } from "./EditCycleDialog";
import { ExerciseTab } from "./ExerciseTab";
import { WeightTab } from "./WeightTab";
import { ProteinTab } from "./ProteinTab";
import type { Cycle, CycleInput } from "./types";

type Tab = "exercise" | "weight" | "protein";
const TABS: { id: Tab; label: string }[] = [
  { id: "exercise", label: "Exercise" },
  { id: "weight", label: "Weight" },
  { id: "protein", label: "Protein" },
];

export function FitnessCyclePage() {
  const { cycleId } = useParams<{ cycleId: string }>();
  const navigate = useNavigate();
  const [cycle, setCycle] = useState<Cycle | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [tab, setTab] = useState<Tab>("exercise");
  const [editing, setEditing] = useState(false);
  const [confirmingArchive, setConfirmingArchive] = useState(false);

  useEffect(() => {
    if (!cycleId) return;
    setLoading(true);
    fetchCycle(cycleId)
      .then(setCycle)
      .catch((err) => {
        if (err instanceof ApiError && err.status === 404) {
          navigate("/fitness", { replace: true });
          return;
        }
        setError(err instanceof Error ? err.message : "Couldn't load this cycle.");
      })
      .finally(() => setLoading(false));
  }, [cycleId, navigate]);

  function handleUpdate(input: CycleInput) {
    if (!cycle) return;
    const original = cycle;
    setCycle({ ...cycle, ...input });
    setEditing(false);
    updateCycle(cycle.id, input)
      .then(setCycle)
      .catch((err) => {
        setCycle(original);
        setError(err instanceof Error ? err.message : "Couldn't update cycle.");
      });
  }

  function handleArchive() {
    if (!cycle) return;
    const original = cycle;
    setConfirmingArchive(false);
    setCycle({ ...cycle, status: "archived" });
    archiveCycle(cycle.id)
      .then(setCycle)
      .catch((err) => {
        setCycle(original);
        setError(err instanceof Error ? err.message : "Couldn't archive cycle.");
      });
  }

  function handleActivate() {
    if (!cycle) return;
    const original = cycle;
    setCycle({ ...cycle, status: "active", archivedAt: null });
    activateCycle(cycle.id)
      .then(setCycle)
      .catch((err) => {
        setCycle(original);
        setError(err instanceof Error ? err.message : "Couldn't activate cycle.");
      });
  }

  if (loading) {
    return (
      <div className="bg-(--card) border-(--line) border-[0.5px] border-solid rounded-xl p-8 text-center text-(--text-faint) text-[length:var(--text-caption)]">
        Loading...
      </div>
    );
  }
  if (!cycle) return null;

  return (
    <div className="h-full flex flex-col">
      <div className="shrink-0">
        <Link
          to="/fitness"
          className="inline-flex items-center gap-1.5 mb-3 text-[length:var(--text-pill)] text-(--text-muted) hover:text-(--fg) transition-colors"
        >
          <ArrowLeft size={13} />
          All cycles
        </Link>

        <div className="flex items-start justify-between gap-4 mb-4">
          <div
            onDoubleClick={() => setEditing(true)}
            role="button"
            tabIndex={0}
            title="Double-click to edit"
          >
            <div className="flex items-center gap-2">
              <h1 className="font-space font-semibold text-(--fg) text-[length:var(--text-caption)]">{cycle.name}</h1>
              {cycle.status === "archived" && (
                <span className="rounded-full bg-(--card-alt) text-(--text-faint) px-1.5 py-0.5 text-[9px] uppercase tracking-wide">
                  Archived
                </span>
              )}
            </div>
            <p className="text-[length:var(--text-pill)] text-(--text-faint) mt-0.5">Started {cycle.startDate}</p>
          </div>

          {cycle.status === "active" ? (
            <button
              onClick={() => setConfirmingArchive(true)}
              className="shrink-0 inline-flex items-center gap-1.5 rounded-md border-(--line) border-[0.5px] border-solid px-2.5 py-1 text-[length:var(--text-pill)] text-(--text-muted) hover:text-(--fg) hover:bg-(--card-alt) transition-colors cursor-pointer"
            >
              <Archive size={12} />
              Archive
            </button>
          ) : (
            <button
              onClick={handleActivate}
              className="shrink-0 inline-flex items-center gap-1.5 rounded-md border-(--line) border-[0.5px] border-solid px-2.5 py-1 text-[length:var(--text-pill)] text-(--text-muted) hover:text-(--fg) hover:bg-(--card-alt) transition-colors cursor-pointer"
            >
              <RotateCcw size={12} />
              Make active
            </button>
          )}
        </div>

        <div className="flex items-center gap-1 mb-4 border-b-[0.5px] border-(--line)">
          {TABS.map((t) => (
            <button
              key={t.id}
              onClick={() => setTab(t.id)}
              className={`-mb-[0.5px] border-b-2 px-3 py-2 text-[length:var(--text-pill)] transition-colors cursor-pointer ${
                tab === t.id
                  ? "border-(--fg) text-(--fg)"
                  : "border-transparent text-(--text-muted) hover:text-(--fg)"
              }`}
            >
              {t.label}
            </button>
          ))}
        </div>

        {error && (
          <div className="mb-3 flex items-center justify-between gap-2 rounded-lg border-(--line) border-[0.5px] border-solid bg-(--card) px-3 py-2 text-[length:var(--text-pill)] text-red-400">
            <span>{error}</span>
            <button onClick={() => setError(null)} aria-label="Dismiss error" className="shrink-0 text-(--text-faint) hover:text-(--fg) transition-colors cursor-pointer">
              <X size={12} />
            </button>
          </div>
        )}
      </div>

      <div className="flex-1 min-h-0 overflow-y-auto overflow-x-hidden themed-scrollbar pb-2">
        {tab === "exercise" && <ExerciseTab cycleId={cycle.id} onError={setError} />}
        {tab === "weight" && <WeightTab cycle={cycle} onError={setError} />}
        {tab === "protein" && <ProteinTab cycle={cycle} onError={setError} />}
      </div>

      {editing && (
        <EditCycleDialog cycle={cycle} onClose={() => setEditing(false)} onSave={handleUpdate} />
      )}

      {confirmingArchive && (
        <ConfirmDeleteDialog
          title="Archive this cycle?"
          message={`"${cycle.name}" will stop being the active cycle. Its data stays, and you can make it active again later.`}
          confirmLabel="Archive"
          onCancel={() => setConfirmingArchive(false)}
          onConfirm={handleArchive}
        />
      )}
    </div>
  );
}
