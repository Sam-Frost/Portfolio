import { useRef, useState, type KeyboardEvent } from "react";
import { CalendarPlus, Link as LinkIcon, X } from "lucide-react";
import type { ResourceInput } from "./types";

interface ResourceDraft {
  label: string;
  url: string;
}

interface AddSubtopicFormProps {
  onAdd: (input: { name: string; targetDate: string | null; resources: ResourceInput[] }) => void;
}

export function AddSubtopicForm({ onAdd }: AddSubtopicFormProps) {
  const [name, setName] = useState("");
  const [targetDate, setTargetDate] = useState("");
  const [showDate, setShowDate] = useState(false);
  const [resources, setResources] = useState<ResourceDraft[]>([]);
  const nameInputRef = useRef<HTMLInputElement>(null);

  function addResourceRow() {
    setResources((prev) => [...prev, { label: "", url: "" }]);
  }

  function updateResourceRow(index: number, field: "label" | "url", value: string) {
    setResources((prev) => prev.map((r, i) => (i === index ? { ...r, [field]: value } : r)));
  }

  function removeResourceRow(index: number) {
    setResources((prev) => prev.filter((_, i) => i !== index));
  }

  function submit() {
    const trimmed = name.trim();
    if (!trimmed) return;

    const cleanResources: ResourceInput[] = resources
      .map((r) => ({ label: r.label.trim() || null, url: r.url.trim() }))
      .filter((r) => r.url !== "");

    onAdd({ name: trimmed, targetDate: targetDate || null, resources: cleanResources });

    setName("");
    setTargetDate("");
    setShowDate(false);
    setResources([]);
    nameInputRef.current?.focus();
  }

  function handleNameKeyDown(e: KeyboardEvent<HTMLInputElement>) {
    if (e.key === "Enter") {
      e.preventDefault();
      submit();
    }
  }

  return (
    <div className="bg-(--card) border-(--line) border-[0.5px] border-solid rounded-xl px-4 py-3 mb-4">
      <input
        ref={nameInputRef}
        value={name}
        onChange={(e) => setName(e.target.value)}
        onKeyDown={handleNameKeyDown}
        placeholder="Add a subtopic..."
        className="w-full bg-transparent text-[length:var(--text-caption)] text-(--fg) placeholder:text-(--text-faint) focus:outline-none"
      />

      {resources.length > 0 && (
        <div className="flex flex-col gap-1.5 mt-2">
          {resources.map((r, i) => (
            <div key={i} className="flex items-center gap-1.5">
              <input
                value={r.label}
                onChange={(e) => updateResourceRow(i, "label", e.target.value)}
                placeholder="Label (optional)"
                className="w-28 shrink-0 rounded-md bg-(--card-alt) border-(--line) border-[0.5px] border-solid px-2 py-1 text-[length:var(--text-pill)] text-(--fg) placeholder:text-(--text-faint) focus:outline-none"
              />
              <input
                value={r.url}
                onChange={(e) => updateResourceRow(i, "url", e.target.value)}
                placeholder="https://..."
                className="flex-1 min-w-0 rounded-md bg-(--card-alt) border-(--line) border-[0.5px] border-solid px-2 py-1 text-[length:var(--text-pill)] text-(--fg) placeholder:text-(--text-faint) focus:outline-none"
              />
              <button
                type="button"
                onClick={() => removeResourceRow(i)}
                aria-label="Remove resource"
                className="shrink-0 flex items-center justify-center size-6 rounded-md text-(--text-faint) hover:text-(--fg) hover:bg-(--card-alt) transition-colors cursor-pointer"
              >
                <X size={12} />
              </button>
            </div>
          ))}
        </div>
      )}

      <div className="flex items-center justify-between gap-2 mt-2">
        <div className="flex items-center gap-1">
          <button
            type="button"
            onClick={() => setShowDate((v) => !v)}
            aria-label="Set target date"
            className="flex items-center justify-center size-6 rounded-md text-(--text-faint) hover:text-(--fg) hover:bg-(--card-alt) transition-colors cursor-pointer"
          >
            <CalendarPlus size={13} />
          </button>

          {showDate && (
            <input
              type="date"
              value={targetDate}
              onChange={(e) => setTargetDate(e.target.value)}
              className="bg-(--card-alt) border-(--line) border-[0.5px] border-solid rounded-lg px-2 py-1 text-[length:var(--text-pill)] text-(--fg) focus:outline-none"
            />
          )}

          <button
            type="button"
            onClick={addResourceRow}
            className="flex items-center gap-1 rounded-md px-2 py-1 text-[length:var(--text-pill)] text-(--text-faint) hover:text-(--fg) hover:bg-(--card-alt) transition-colors cursor-pointer"
          >
            <LinkIcon size={13} />
            Add resource
          </button>
        </div>

        <button
          type="button"
          onClick={submit}
          className="shrink-0 rounded-md bg-(--fg) text-(--bg) px-2.5 py-1 text-[length:var(--text-pill)] cursor-pointer"
        >
          Add subtopic
        </button>
      </div>
    </div>
  );
}
