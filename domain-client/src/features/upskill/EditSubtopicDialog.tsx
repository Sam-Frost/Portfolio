import { useRef, useState, type KeyboardEvent } from "react";
import { Plus, X } from "lucide-react";
import type { ResourceInput, Subtopic } from "./types";

interface EditSubtopicDialogProps {
  subtopic: Subtopic;
  onClose: () => void;
  onSave: (input: { name: string; targetDate: string | null }) => void;
  onAddResource: (input: ResourceInput) => void;
  onRemoveResource: (resourceId: string) => void;
}

export function EditSubtopicDialog({ subtopic, onClose, onSave, onAddResource, onRemoveResource }: EditSubtopicDialogProps) {
  const [name, setName] = useState(subtopic.name);
  const [targetDate, setTargetDate] = useState(subtopic.targetDate ?? "");
  const [newLabel, setNewLabel] = useState("");
  const [newUrl, setNewUrl] = useState("");
  const nameInputRef = useRef<HTMLInputElement>(null);

  function handleSave() {
    const trimmed = name.trim();
    if (!trimmed) {
      nameInputRef.current?.focus();
      return;
    }
    onSave({ name: trimmed, targetDate: targetDate || null });
  }

  function handleAddResource() {
    const url = newUrl.trim();
    if (!url) return;
    onAddResource({ label: newLabel.trim() || null, url });
    setNewLabel("");
    setNewUrl("");
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
        aria-label="Edit subtopic"
      >
        <h2 className="text-[length:var(--text-caption)] text-(--fg) mb-3">Edit subtopic</h2>

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

        <label className="block text-[length:var(--text-pill)] text-(--text-muted) mb-1">Resources</label>
        <div className="flex flex-col gap-1.5 mb-2">
          {subtopic.resources.map((res) => (
            <div
              key={res.id}
              className="flex items-center gap-1.5 rounded-lg bg-(--card-alt) border-(--line) border-[0.5px] border-solid px-2.5 py-1.5"
            >
              <span className="flex-1 min-w-0 truncate text-[length:var(--text-pill)] text-(--fg)">
                {res.label ?? res.url}
              </span>
              <button
                type="button"
                onClick={() => onRemoveResource(res.id)}
                aria-label="Remove resource"
                className="shrink-0 flex items-center justify-center size-5 rounded-md text-(--text-faint) hover:text-(--fg) cursor-pointer"
              >
                <X size={11} />
              </button>
            </div>
          ))}
          {subtopic.resources.length === 0 && (
            <p className="text-[length:var(--text-pill)] text-(--text-faint)">No resources yet.</p>
          )}
        </div>

        <div className="flex items-center gap-1.5 mb-4">
          <input
            value={newLabel}
            onChange={(e) => setNewLabel(e.target.value)}
            placeholder="Label (optional)"
            className="w-24 shrink-0 rounded-md bg-(--card-alt) border-(--line) border-[0.5px] border-solid px-2 py-1 text-[length:var(--text-pill)] text-(--fg) placeholder:text-(--text-faint) focus:outline-none"
          />
          <input
            value={newUrl}
            onChange={(e) => setNewUrl(e.target.value)}
            placeholder="https://..."
            className="flex-1 min-w-0 rounded-md bg-(--card-alt) border-(--line) border-[0.5px] border-solid px-2 py-1 text-[length:var(--text-pill)] text-(--fg) placeholder:text-(--text-faint) focus:outline-none"
          />
          <button
            type="button"
            onClick={handleAddResource}
            aria-label="Add resource"
            className="shrink-0 flex items-center justify-center size-6 rounded-md text-(--text-faint) hover:text-(--fg) hover:bg-(--card-alt) transition-colors cursor-pointer"
          >
            <Plus size={13} />
          </button>
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
