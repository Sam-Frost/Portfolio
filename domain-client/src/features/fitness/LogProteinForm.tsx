import { useState, type FormEvent } from "react";
import { todayISTKey } from "./dateUtils";
import type { Food } from "./types";

interface LogProteinFormProps {
  foods: Food[];
  onSubmit: (input: { foodId: string; date: string; quantity: number }) => void;
}

export function LogProteinForm({ foods, onSubmit }: LogProteinFormProps) {
  const [foodId, setFoodId] = useState("");
  const [quantity, setQuantity] = useState("");
  const [date, setDate] = useState(todayISTKey());

  const food = foods.find((f) => f.id === foodId);
  const qty = Number(quantity);
  const computed = food && Number.isFinite(qty) && qty > 0 ? Math.round(qty * food.proteinPerUnit * 10) / 10 : null;

  function submit(e: FormEvent) {
    e.preventDefault();
    if (!foodId || !Number.isFinite(qty) || qty <= 0) return;
    onSubmit({ foodId, date, quantity: qty });
    setQuantity("");
    setDate(todayISTKey());
  }

  const field =
    "rounded-lg bg-(--card-alt) border-(--line) border-[0.5px] border-solid px-2.5 py-1.5 text-[length:var(--text-pill)] text-(--fg) focus:outline-none";

  return (
    <form onSubmit={submit} className="bg-(--card) border-(--line) border-[0.5px] border-solid rounded-xl p-4 flex flex-wrap items-end gap-3">
      <label className="flex flex-col gap-1 flex-1 min-w-40">
        <span className="text-[length:var(--text-pill)] text-(--text-muted)">Food</span>
        <select value={foodId} onChange={(e) => setFoodId(e.target.value)} className={field} disabled={foods.length === 0}>
          <option value="">{foods.length === 0 ? "Add a food first" : "Select a food"}</option>
          {foods.map((f) => (
            <option key={f.id} value={f.id}>
              {f.name} ({f.unit})
            </option>
          ))}
        </select>
      </label>
      <label className="flex flex-col gap-1">
        <span className="text-[length:var(--text-pill)] text-(--text-muted)">Quantity{food ? ` (${food.unit})` : ""}</span>
        <input type="number" inputMode="decimal" step={0.1} min={0} value={quantity} onChange={(e) => setQuantity(e.target.value)} placeholder="1" className={`${field} w-24`} />
      </label>
      <label className="flex flex-col gap-1">
        <span className="text-[length:var(--text-pill)] text-(--text-muted)">Date</span>
        <input type="date" value={date} max={todayISTKey()} onChange={(e) => setDate(e.target.value)} className={field} />
      </label>
      <div className="flex items-center gap-3">
        {computed != null && <span className="text-[length:var(--text-pill)] text-(--green)">= {computed} g protein</span>}
        <button type="submit" className="rounded-md bg-(--fg) text-(--bg) px-3 py-1.5 text-[length:var(--text-pill)] cursor-pointer">
          Log
        </button>
      </div>
    </form>
  );
}
