import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { Plus, Trash2 } from "lucide-react";
import { createNote, deleteNote, fetchNotes } from "./api";
import type { NoteSummary } from "./types";

function formatDate(value: string) {
  return new Date(value).toLocaleDateString(undefined, { year: "numeric", month: "short", day: "numeric" });
}

export function NotepadPage() {
  const navigate = useNavigate();
  const [notes, setNotes] = useState<NoteSummary[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [creating, setCreating] = useState(false);

  useEffect(() => {
    fetchNotes()
      .then(setNotes)
      .catch((err) => setError(err instanceof Error ? err.message : "Couldn't load notes."))
      .finally(() => setLoading(false));
  }, []);

  function handleCreate() {
    setCreating(true);
    createNote()
      .then((note) => navigate(`/notepad/${note.id}`))
      .catch((err) => setError(err instanceof Error ? err.message : "Couldn't create note."))
      .finally(() => setCreating(false));
  }

  function handleDelete(id: string) {
    const previous = notes;
    setNotes((prev) => prev.filter((n) => n.id !== id));
    deleteNote(id).catch((err) => {
      setNotes(previous);
      setError(err instanceof Error ? err.message : "Couldn't delete note.");
    });
  }

  return (
    <div className="h-full flex flex-col">
      <div className="shrink-0 flex items-center justify-between mb-3">
        <h1 className="text-xl font-space font-semibold text-(--fg)">Notepad</h1>
        <button
          onClick={handleCreate}
          disabled={creating}
          className={`flex items-center gap-1.5 rounded-lg bg-(--fg) text-(--bg) px-3 py-1.5 text-[length:var(--text-caption)] transition-opacity ${
            creating ? "opacity-60 cursor-not-allowed" : "cursor-pointer"
          }`}
        >
          <Plus size={14} />
          New note
        </button>
      </div>

      {error && (
        <div className="shrink-0 mb-3 flex items-center justify-between gap-2 rounded-lg border-(--line) border-[0.5px] border-solid bg-(--card) px-3 py-2 text-[length:var(--text-pill)] text-red-400">
          <span>{error}</span>
          <button onClick={() => setError(null)} className="shrink-0 text-(--text-faint) hover:text-(--fg) transition-colors cursor-pointer">
            Dismiss
          </button>
        </div>
      )}

      <div className="flex-1 min-h-0 overflow-y-auto themed-scrollbar">
        {loading && (
          <div className="bg-(--card) border-(--line) border-[0.5px] border-solid rounded-xl p-8 text-center text-(--text-faint) text-[length:var(--text-caption)]">
            Loading notes...
          </div>
        )}

        {!loading && notes.length === 0 && (
          <div className="bg-(--card) border-(--line) border-[0.5px] border-solid rounded-xl p-8 text-center text-(--text-faint) text-[length:var(--text-caption)]">
            No notes yet.
          </div>
        )}

        {!loading && notes.length > 0 && (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-2 pb-2">
            {notes.map((note) => (
              <div
                key={note.id}
                onClick={() => navigate(`/notepad/${note.id}`)}
                className="group relative flex flex-col gap-1 bg-(--card) border-(--line) border-[0.5px] border-solid rounded-lg px-3 py-2.5 cursor-pointer hover:border-(--line-strong) transition-colors"
              >
                <span className="truncate pr-5 text-[length:var(--text-caption)] text-(--fg)">{note.title}</span>
                <span className="text-[length:var(--text-pill)] text-(--text-faint)">{formatDate(note.createdAt)}</span>

                <button
                  onClick={(e) => {
                    e.stopPropagation();
                    handleDelete(note.id);
                  }}
                  aria-label="Delete note"
                  className="absolute top-2 right-2 opacity-0 group-hover:opacity-100 text-(--text-faint) hover:text-(--label-red) transition-colors cursor-pointer"
                >
                  <Trash2 size={13} />
                </button>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
