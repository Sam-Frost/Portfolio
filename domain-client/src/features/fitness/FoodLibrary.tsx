import { useState, type FormEvent } from "react";
import { Pencil, Trash2 } from "lucide-react";
import { ConfirmDeleteDialog } from "./ConfirmDeleteDialog";
import { EditFoodDialog } from "./EditFoodDialog";
import type { Food } from "./types";

interface FoodLibraryProps {
  foods: Food[];
  onAdd: (input: { name: string; unit: string; proteinPerUnit: number }) => void;
  onUpdate: (id: string, input: { name: string; unit: string; proteinPerUnit: number }) => void;
  onDelete: (id: string) => void;
  // When embedded in a page that already provides a titled card (Settings),
  // drop this component's own card chrome + heading.
  bare?: boolean;
}

export function FoodLibrary({ foods, onAdd, onUpdate, onDelete, bare = false }: FoodLibraryProps) {
  const [name, setName] = useState("");
  const [unit, setUnit] = useState("");
  const [proteinPerUnit, setProteinPerUnit] = useState("");
  const [editing, setEditing] = useState<Food | null>(null);
  const [deleting, setDeleting] = useState<Food | null>(null);

  function submit(e: FormEvent) {
    e.preventDefault();
    const n = Number(proteinPerUnit);
    if (!name.trim() || !unit.trim() || !Number.isFinite(n) || n <= 0) return;
    onAdd({ name: name.trim(), unit: unit.trim(), proteinPerUnit: n });
    setName("");
    setUnit("");
    setProteinPerUnit("");
  }

  const field =
    "rounded-lg bg-(--card-alt) border-(--line) border-[0.5px] border-solid px-2.5 py-1.5 text-[length:var(--text-pill)] text-(--fg) placeholder:text-(--text-faint) focus:outline-none";

  return (
    <div
      className={
        bare
          ? ""
          : "bg-(--card) border-(--line) border-[0.5px] border-solid rounded-xl p-4"
      }
    >
      {!bare && (
        <h2 className="text-[length:var(--text-pill)] font-medium uppercase tracking-wide text-(--text-faint) mb-3">
          Food library
        </h2>
      )}

      <form onSubmit={submit} className="flex flex-wrap items-end gap-2 mb-3">
        <label className="flex flex-col gap-1 flex-1 min-w-32">
          <span className="text-[length:var(--text-pill)] text-(--text-muted)">Food</span>
          <input value={name} onChange={(e) => setName(e.target.value)} placeholder="e.g. Milk" className={field} />
        </label>
        <label className="flex flex-col gap-1">
          <span className="text-[length:var(--text-pill)] text-(--text-muted)">Unit</span>
          <input value={unit} onChange={(e) => setUnit(e.target.value)} placeholder="glass" className={`${field} w-24`} />
        </label>
        <label className="flex flex-col gap-1">
          <span className="text-[length:var(--text-pill)] text-(--text-muted)">Protein/unit (g)</span>
          <input type="number" step={0.1} min={0} value={proteinPerUnit} onChange={(e) => setProteinPerUnit(e.target.value)} placeholder="8" className={`${field} w-24`} />
        </label>
        <button type="submit" className="rounded-md bg-(--fg) text-(--bg) px-3 py-1.5 text-[length:var(--text-pill)] cursor-pointer">
          Add
        </button>
      </form>

      {foods.length === 0 ? (
        <p className="text-[length:var(--text-pill)] text-(--text-faint) py-2">
          No foods yet — add the things you eat and how much protein each has.
        </p>
      ) : (
        <div className="flex flex-col gap-1">
          {foods.map((food) => (
            <div key={food.id} className="group flex items-center justify-between gap-2 rounded-lg px-2 py-1.5 hover:bg-(--card-alt)">
              <span className="text-[length:var(--text-pill)] text-(--fg) truncate">
                {food.name}{" "}
                <span className="text-(--text-faint)">
                  · {round1(food.proteinPerUnit)} g / {food.unit}
                </span>
              </span>
              <div className="flex items-center gap-0.5 shrink-0">
                <button
                  onClick={() => setEditing(food)}
                  aria-label="Edit food"
                  className="flex items-center justify-center size-6 rounded-md text-(--text-faint) opacity-100 lg:opacity-0 lg:group-hover:opacity-100 hover:text-(--fg) transition-opacity cursor-pointer"
                >
                  <Pencil size={12} />
                </button>
                <button
                  onClick={() => setDeleting(food)}
                  aria-label="Delete food"
                  className="flex items-center justify-center size-6 rounded-md text-(--text-faint) opacity-100 lg:opacity-0 lg:group-hover:opacity-100 hover:text-(--fg) transition-opacity cursor-pointer"
                >
                  <Trash2 size={12} />
                </button>
              </div>
            </div>
          ))}
        </div>
      )}

      {editing && (
        <EditFoodDialog
          food={editing}
          onClose={() => setEditing(null)}
          onSave={(input) => {
            onUpdate(editing.id, input);
            setEditing(null);
          }}
        />
      )}

      {deleting && (
        <ConfirmDeleteDialog
          title="Delete food?"
          message={`"${deleting.name}" and every protein log that used it will be deleted.`}
          onCancel={() => setDeleting(null)}
          onConfirm={() => {
            onDelete(deleting.id);
            setDeleting(null);
          }}
        />
      )}
    </div>
  );
}

function round1(n: number): number {
  return Math.round(n * 10) / 10;
}
