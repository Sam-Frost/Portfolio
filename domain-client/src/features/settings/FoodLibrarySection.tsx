import { useEffect, useState } from "react";
import { X } from "lucide-react";
import { createFood, deleteFood, fetchFoods, updateFood } from "../fitness/api";
import { FoodLibrary } from "../fitness/FoodLibrary";
import type { Food } from "../fitness/types";

type FoodInput = { name: string; unit: string; proteinPerUnit: number };

// The fitness food library is a single shared list (server/internal/fitness
// — fitness_foods has no cycle_id). It's edited here and reused by every
// cycle's Protein tab.
export function FoodLibrarySection() {
  const [foods, setFoods] = useState<Food[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    fetchFoods()
      .then(setFoods)
      .catch((err) => setError(err instanceof Error ? err.message : "Couldn't load foods."))
      .finally(() => setLoading(false));
  }, []);

  function handleAdd(input: FoodInput) {
    createFood(input)
      .then((food) => setFoods((prev) => [...prev, food].sort((a, b) => a.name.localeCompare(b.name))))
      .catch((err) => setError(err instanceof Error ? err.message : "Couldn't add food."));
  }

  function handleUpdate(id: string, input: FoodInput) {
    const original = foods;
    setFoods((prev) => prev.map((f) => (f.id === id ? { ...f, ...input } : f)));
    updateFood(id, input)
      .then((updated) =>
        setFoods((prev) => prev.map((f) => (f.id === id ? updated : f)).sort((a, b) => a.name.localeCompare(b.name))),
      )
      .catch((err) => {
        setFoods(original);
        setError(err instanceof Error ? err.message : "Couldn't update food.");
      });
  }

  function handleDelete(id: string) {
    const original = foods;
    setFoods((prev) => prev.filter((f) => f.id !== id));
    deleteFood(id).catch((err) => {
      setFoods(original);
      setError(err instanceof Error ? err.message : "Couldn't delete food.");
    });
  }

  return (
    <div className="flex flex-col gap-3">
      <p className="text-[length:var(--text-caption)] text-(--text-faint)">
        Shared across every fitness cycle. Deleting a food also removes every protein log that used it.
      </p>

      {loading ? (
        <p className="text-[length:var(--text-caption)] text-(--text-faint)">Loading foods...</p>
      ) : (
        <FoodLibrary foods={foods} onAdd={handleAdd} onUpdate={handleUpdate} onDelete={handleDelete} bare />
      )}

      {error && (
        <div className="flex items-center justify-between gap-2 rounded-lg border-(--line) border-[0.5px] border-solid bg-(--card-alt) px-3 py-2 text-[length:var(--text-pill)] text-red-400">
          <span>{error}</span>
          <button
            onClick={() => setError(null)}
            aria-label="Dismiss error"
            className="shrink-0 text-(--text-faint) hover:text-(--fg) transition-colors cursor-pointer"
          >
            <X size={12} />
          </button>
        </div>
      )}
    </div>
  );
}
