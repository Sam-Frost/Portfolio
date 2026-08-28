import { useState, type FormEvent } from "react";
import { todayISTKey } from "./dateUtils";

interface LogValueFormProps {
  valueLabel: string;
  unit?: string | null;
  step?: number;
  submitLabel?: string;
  onSubmit: (date: string, value: number) => void;
}

// Shared "pick a date, enter one number" logger used by the weight tab and
// the exercise detail page. Date defaults to today (IST); an existing entry
// for the chosen date is overwritten server-side (PUT upsert).
export function LogValueForm({ valueLabel, unit, step = 0.1, submitLabel = "Log", onSubmit }: LogValueFormProps) {
  const [date, setDate] = useState(todayISTKey());
  const [value, setValue] = useState("");

  function submit(e: FormEvent) {
    e.preventDefault();
    const n = Number(value);
    if (!Number.isFinite(n) || n <= 0) return;
    onSubmit(date, n);
    setValue("");
    setDate(todayISTKey());
  }

  return (
    <form
      onSubmit={submit}
      className="bg-(--card) border-(--line) border-[0.5px] border-solid rounded-xl px-4 py-3 mb-4 flex flex-wrap items-end gap-3"
    >
      <label className="flex flex-col gap-1">
        <span className="text-[length:var(--text-pill)] text-(--text-muted)">Date</span>
        <input
          type="date"
          value={date}
          max={todayISTKey()}
          onChange={(e) => setDate(e.target.value)}
          className="rounded-lg bg-(--card-alt) border-(--line) border-[0.5px] border-solid px-2 py-1.5 text-[length:var(--text-pill)] text-(--fg) focus:outline-none"
        />
      </label>

      <label className="flex flex-col gap-1">
        <span className="text-[length:var(--text-pill)] text-(--text-muted)">
          {valueLabel}
          {unit ? ` (${unit})` : ""}
        </span>
        <input
          type="number"
          inputMode="decimal"
          step={step}
          min={0}
          value={value}
          onChange={(e) => setValue(e.target.value)}
          placeholder="0"
          className="w-28 rounded-lg bg-(--card-alt) border-(--line) border-[0.5px] border-solid px-2.5 py-1.5 text-[length:var(--text-caption)] text-(--fg) placeholder:text-(--text-faint) focus:outline-none"
        />
      </label>

      <button
        type="submit"
        className="rounded-md bg-(--fg) text-(--bg) px-3 py-1.5 text-[length:var(--text-pill)] cursor-pointer"
      >
        {submitLabel}
      </button>
    </form>
  );
}
