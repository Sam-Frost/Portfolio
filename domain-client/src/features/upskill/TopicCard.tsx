import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { Trash2 } from "lucide-react";
import { ConfirmDeleteDialog } from "./ConfirmDeleteDialog";
import { EditTopicDialog } from "./EditTopicDialog";
import type { Topic } from "./types";

interface TopicCardProps {
  topic: Topic;
  onUpdate: (id: string, input: { name: string; targetDate: string | null }) => void;
  onDelete: (id: string) => void;
}

function formatDate(value: string) {
  return new Date(value).toLocaleDateString(undefined, { year: "numeric", month: "short", day: "numeric" });
}

export function TopicCard({ topic, onUpdate, onDelete }: TopicCardProps) {
  const navigate = useNavigate();
  const [editing, setEditing] = useState(false);
  const [confirmingDelete, setConfirmingDelete] = useState(false);
  const progress = topic.subtopicCount > 0 ? topic.doneCount / topic.subtopicCount : 0;

  return (
    <div
      onClick={() => navigate(topic.id)}
      onDoubleClick={(e) => {
        e.stopPropagation();
        setEditing(true);
      }}
      role="button"
      tabIndex={0}
      className="group relative flex flex-col gap-2 bg-(--card) border-(--line) border-[0.5px] border-solid rounded-xl px-4 py-3 cursor-pointer hover:border-(--line-strong) transition-colors"
    >
      <div className="flex items-start justify-between gap-2">
        <span className="min-w-0 flex-1 truncate text-[length:var(--text-caption)] text-(--fg) font-medium">{topic.name}</span>
        <button
          onClick={(e) => {
            e.stopPropagation();
            setConfirmingDelete(true);
          }}
          aria-label="Delete topic"
          className="shrink-0 flex items-center justify-center size-6 rounded-md text-(--text-faint) opacity-0 group-hover:opacity-100 hover:text-(--fg) hover:bg-(--card-alt) transition-opacity cursor-pointer"
        >
          <Trash2 size={13} />
        </button>
      </div>

      {topic.targetDate && (
        <span className="text-[length:var(--text-pill)] text-(--text-faint)">Due {formatDate(topic.targetDate)}</span>
      )}

      <div className="flex items-center gap-2 mt-1">
        <div className="flex-1 h-1.5 rounded-full bg-(--ring-track) overflow-hidden">
          <div className="h-full rounded-full bg-(--green)" style={{ width: `${progress * 100}%` }} />
        </div>
        <span className="shrink-0 text-[length:var(--text-pill)] text-(--text-faint)">
          {topic.doneCount}/{topic.subtopicCount}
        </span>
      </div>

      {/* These dialogs render as DOM children of this card (no portal), so a
          click anywhere in them — including the overlay background — would
          otherwise bubble up into the card's onClick and trigger navigate(). */}
      {editing && (
        <div onClick={(e) => e.stopPropagation()}>
          <EditTopicDialog
            topic={topic}
            onClose={() => setEditing(false)}
            onSave={(input) => {
              onUpdate(topic.id, input);
              setEditing(false);
            }}
          />
        </div>
      )}

      {confirmingDelete && (
        <div onClick={(e) => e.stopPropagation()}>
          <ConfirmDeleteDialog
            title="Delete topic?"
            message={`"${topic.name}" and all its subtopics and resources will be deleted.`}
            onCancel={() => setConfirmingDelete(false)}
            onConfirm={() => {
              onDelete(topic.id);
              setConfirmingDelete(false);
            }}
          />
        </div>
      )}
    </div>
  );
}
