import { useRef, useState, type KeyboardEvent } from "react";
import { LabelPicker } from "../labels/LabelPicker";
import type { Label } from "../labels/types";
import type { Todo } from "./types";

interface EditTodoDialogProps {
  todo: Todo;
  labels: Label[];
  onClose: () => void;
  onSave: (input: { name: string; description: string | null; targetDate: string | null; labelId: string | null }) => void;
}

export function EditTodoDialog({ todo, labels, onClose, onSave }: EditTodoDialogProps) {
  const [name, setName] = useState(todo.name);
  const [description, setDescription] = useState(todo.description ?? "");
  const [targetDate, setTargetDate] = useState(todo.targetDate ?? "");
  const [labelId, setLabelId] = useState<string | null>(todo.labelId);
  const nameInputRef = useRef<HTMLInputElement>(null);

  function handleSave() {
    const trimmed = name.trim();
    if (!trimmed) {
      nameInputRef.current?.focus();
      return;
    }
    onSave({
      name: trimmed,
      description: description.trim() || null,
      targetDate: targetDate || null,
      labelId,
    });
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
        className="w-full max-w-sm max-h-[90vh] overflow-y-auto themed-scrollbar rounded-xl border-(--line) border-[0.5px] border-solid bg-(--card) p-4 shadow-lg"
        onClick={(e) => e.stopPropagation()}
        role="dialog"
        aria-modal="true"
        aria-label="Edit todo"
      >
        <h2 className="text-[length:var(--text-caption)] text-(--fg) mb-3">Edit todo</h2>

        <label className="block text-[length:var(--text-pill)] text-(--text-muted) mb-1">Name</label>
        <input
          ref={nameInputRef}
          autoFocus
          value={name}
          onChange={(e) => setName(e.target.value)}
          className="w-full mb-3 rounded-lg bg-(--card-alt) border-(--line) border-[0.5px] border-solid px-2.5 py-1.5 text-[length:var(--text-caption)] text-(--fg) focus:outline-none"
        />

        <label className="block text-[length:var(--text-pill)] text-(--text-muted) mb-1">Description</label>
        <textarea
          value={description}
          onChange={(e) => setDescription(e.target.value)}
          rows={3}
          className="w-full mb-3 resize-none rounded-lg bg-(--card-alt) border-(--line) border-[0.5px] border-solid px-2.5 py-1.5 text-[length:var(--text-caption)] text-(--fg) focus:outline-none"
        />

        <label className="block text-[length:var(--text-pill)] text-(--text-muted) mb-1">Target date</label>
        <input
          type="date"
          value={targetDate}
          onChange={(e) => setTargetDate(e.target.value)}
          className="w-full mb-4 rounded-lg bg-(--card-alt) border-(--line) border-[0.5px] border-solid px-2 py-1.5 text-[length:var(--text-pill)] text-(--fg) focus:outline-none"
        />

        <label className="block text-[length:var(--text-pill)] text-(--text-muted) mb-1">Label</label>
        <div className="mb-4">
          <LabelPicker labels={labels} selectedId={labelId} onChange={setLabelId} alwaysVisible />
        </div>

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
