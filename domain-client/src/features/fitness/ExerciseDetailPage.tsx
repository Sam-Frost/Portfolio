import { useCallback, useState } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import { ArrowLeft, X } from "lucide-react";
import { ExerciseDetail } from "./ExerciseDetail";
import type { Exercise } from "./types";

// Standalone route (/fitness/exercises/:exerciseId) — a thin frame (back link,
// error banner, scroll) around the shared <ExerciseDetail>, which the Exercise
// tab also renders inline.
export function ExerciseDetailPage() {
  const { exerciseId } = useParams<{ exerciseId: string }>();
  const navigate = useNavigate();
  const [error, setError] = useState<string | null>(null);
  const [cycleId, setCycleId] = useState<string | null>(null);

  const handleLoaded = useCallback((exercise: Exercise) => setCycleId(exercise.cycleId), []);
  const handleGone = useCallback(() => navigate("/fitness", { replace: true }), [navigate]);
  const backTo = cycleId ? `/fitness/${cycleId}` : "/fitness";

  if (!exerciseId) return null;

  return (
    <div className="min-h-full lg:h-full flex flex-col">
      <div className="shrink-0">
        <Link
          to={backTo}
          className="inline-flex items-center gap-1.5 mb-3 text-[length:var(--text-pill)] text-(--text-muted) hover:text-(--fg) transition-colors"
        >
          <ArrowLeft size={13} />
          Back to cycle
        </Link>

        {error && (
          <div className="mb-3 flex items-center justify-between gap-2 rounded-lg border-(--line) border-[0.5px] border-solid bg-(--card) px-3 py-2 text-[length:var(--text-pill)] text-red-400">
            <span>{error}</span>
            <button onClick={() => setError(null)} aria-label="Dismiss error" className="shrink-0 text-(--text-faint) hover:text-(--fg) transition-colors cursor-pointer">
              <X size={12} />
            </button>
          </div>
        )}
      </div>

      <div className="flex-1 lg:min-h-0 lg:overflow-y-auto overflow-x-hidden themed-scrollbar pb-2">
        <ExerciseDetail
          key={exerciseId}
          exerciseId={exerciseId}
          onError={setError}
          onLoaded={handleLoaded}
          onNotFound={handleGone}
          onExerciseDelete={() => navigate(backTo, { replace: true })}
        />
      </div>
    </div>
  );
}
