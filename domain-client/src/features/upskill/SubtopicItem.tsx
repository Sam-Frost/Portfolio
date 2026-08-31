import { useState } from "react";
import { Check, ExternalLink, Info, Trash2 } from "lucide-react";
import { ConfirmDeleteDialog } from "./ConfirmDeleteDialog";
import { EditSubtopicDialog } from "./EditSubtopicDialog";
import type { ResourceInput, Subtopic } from "./types";

interface SubtopicItemProps {
  subtopic: Subtopic;
  onMarkDone: (id: string) => void;
  onUpdate: (id: string, input: { name: string; targetDate: string | null }) => void;
  onDelete: (id: string) => void;
  onAddResource: (subtopicId: string, input: ResourceInput) => void;
  onRemoveResource: (subtopicId: string, resourceId: string) => void;
}

function formatDate(value: string) {
  return new Date(value).toLocaleDateString(undefined, { year: "numeric", month: "short", day: "numeric" });
}

export function SubtopicItem({
  subtopic,
  onMarkDone,
  onUpdate,
  onDelete,
  onAddResource,
  onRemoveResource,
}: SubtopicItemProps) {
  const [editing, setEditing] = useState(false);
  const [confirmingDelete, setConfirmingDelete] = useState(false);

  return (
    <div
      onDoubleClick={() => setEditing(true)}
      className="group relative flex items-center gap-2 bg-(--card) border-(--line) border-[0.5px] border-solid rounded-lg px-3 py-1.5"
    >
      <div className="flex-1 min-w-0 flex items-center gap-1.5">
        <span
          className={`text-[length:var(--text-pill)] truncate ${
            subtopic.done ? "text-(--text-faint) line-through" : "text-(--fg)"
          }`}
        >
          {subtopic.name}
        </span>
        <Info size={11} className="shrink-0 text-(--text-faint)" />
      </div>

      {subtopic.resources.map((res) => (
        <a
          key={res.id}
          href={res.url}
          target="_blank"
          rel="noopener noreferrer"
          onClick={(e) => e.stopPropagation()}
          className="shrink-0 flex items-center gap-1 max-w-24 rounded-md bg-(--card-alt) border-(--line) border-[0.5px] border-solid px-1.5 py-0.5 text-[length:var(--text-pill)] text-(--text-muted) hover:text-(--fg) transition-colors"
        >
          <ExternalLink size={10} className="shrink-0" />
          <span className="truncate">{res.label ?? "Visit"}</span>
        </a>
      ))}

      {subtopic.done ? (
        <span className="flex items-center gap-1 shrink-0 text-[length:var(--text-pill)] text-(--green)">
          <Check size={12} strokeWidth={3} />
          Done
        </span>
      ) : (
        <button
          onClick={() => onMarkDone(subtopic.id)}
          className="shrink-0 rounded-md bg-(--card-alt) border-(--line) border-[0.5px] border-solid px-2 py-0.5 text-[length:var(--text-pill)] text-(--text-muted) hover:text-(--fg) hover:border-(--fg) transition-colors cursor-pointer"
        >
          Done
        </button>
      )}

      <button
        onClick={() => setConfirmingDelete(true)}
        aria-label="Delete subtopic"
        className="shrink-0 flex items-center justify-center size-6 rounded-md text-(--text-faint) opacity-100 lg:opacity-0 lg:group-hover:opacity-100 hover:text-(--fg) hover:bg-(--card-alt) transition-opacity cursor-pointer"
      >
        <Trash2 size={13} />
      </button>

      <div className="pointer-events-none absolute left-3 top-full z-10 mt-1 w-max max-w-64 whitespace-pre-wrap rounded-md border-(--line) border-[0.5px] border-solid bg-(--card-alt) px-2.5 py-1.5 text-[length:var(--text-pill)] text-(--text-muted) opacity-0 shadow-lg transition-opacity delay-0 duration-150 group-hover:opacity-100 group-hover:delay-1000">
        <p className="whitespace-nowrap">Added: {formatDate(subtopic.dateAdded)}</p>
        {subtopic.targetDate && <p className="whitespace-nowrap">Due: {formatDate(subtopic.targetDate)}</p>}
      </div>

      {editing && (
        <EditSubtopicDialog
          subtopic={subtopic}
          onClose={() => setEditing(false)}
          onSave={(input) => {
            onUpdate(subtopic.id, input);
            setEditing(false);
          }}
          onAddResource={(input) => onAddResource(subtopic.id, input)}
          onRemoveResource={(resourceId) => onRemoveResource(subtopic.id, resourceId)}
        />
      )}

      {confirmingDelete && (
        <ConfirmDeleteDialog
          title="Delete subtopic?"
          message={`"${subtopic.name}" and its resources will be deleted.`}
          onCancel={() => setConfirmingDelete(false)}
          onConfirm={() => {
            onDelete(subtopic.id);
            setConfirmingDelete(false);
          }}
        />
      )}
    </div>
  );
}
