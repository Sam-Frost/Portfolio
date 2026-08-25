import { useRef, useState, type KeyboardEvent } from "react";
import type { Topic } from "./types";

interface EditTopicDialogProps {
  topic: Topic;
  onClose: () => void;
  onSave: (input: { name: string; targetDate: string | null }) => void;
}

export function EditTopicDialog({ topic, onClose, onSave }: EditTopicDialogProps) {
  const [name, setName] = useState(topic.name);
  const [targetDate, setTargetDate] = useState(topic.targetDate ?? "");
  const nameInputRef = useRef<HTMLInputElement>(null);

  function handleSave() {
    const trimmed = name.trim();
    if (!trimmed) {
      nameInputRef.current?.focus();
      return;
    }
    onSave({ name: trimmed, targetDate: targetDate || null });
  }

  function handleKeyDown(e: KeyboardEvent<HTMLDivElement>) {
    if (e.key === "Escape") onClose();
  }

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 px-4"
      onClick={onClose}
      onKeyDown={handleKeyDown}
      role="presentation"
    >
      <div
        className="w-full max-w-sm rounded-xl border-(--line) border-[0.5px] border-solid bg-(--card) p-4 shadow-lg"
        onClick={(e) => e.stopPropagation()}
        role="dialog"
        aria-modal="true"
        aria-label="Edit topic"
      >
        <h2 className="text-[length:var(--text-caption)] text-(--fg) mb-3">Edit topic</h2>

        <label className="block text-[length:var(--text-pill)] text-(--text-muted) mb-1">Name</label>
        <input
          ref={nameInputRef}
          autoFocus
          value={name}
          onChange={(e) => setName(e.target.value)}
          className="w-full mb-3 rounded-lg bg-(--card-alt) border-(--line) border-[0.5px] border-solid px-2.5 py-1.5 text-[length:var(--text-caption)] text-(--fg) focus:outline-none"
        />

        <label className="block text-[length:var(--text-pill)] text-(--text-muted) mb-1">Target date</label>
        <input
          type="date"
          value={targetDate}
          onChange={(e) => setTargetDate(e.target.value)}
          className="w-full mb-4 rounded-lg bg-(--card-alt) border-(--line) border-[0.5px] border-solid px-2 py-1.5 text-[length:var(--text-pill)] text-(--fg) focus:outline-none"
        />

        <div className="flex items-center justify-end gap-1.5">
          <button
            type="button"
            onClick={onClose}
            className="rounded-md px-2.5 py-1 text-[length:var(--text-pill)] text-(--text-muted) hover:text-(--fg) hover:bg-(--card-alt) transition-colors cursor-pointer"
          >
            Cancel
          </button>
          <button
            type="button"
            onClick={handleSave}
            className="rounded-md bg-(--fg) text-(--bg) px-2.5 py-1 text-[length:var(--text-pill)] cursor-pointer"
          >
            Save
          </button>
        </div>
      </div>
    </div>
  );
}
