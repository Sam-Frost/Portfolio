import { useEffect, useRef, useState } from "react";
import { useNavigate } from "react-router-dom";
import {
  Archive,
  ArchiveRestore,
  Lock,
  LockOpen,
  NotebookPen,
  Plus,
  Search,
  Star,
  Tag,
  TagX,
  Trash2,
} from "lucide-react";
import { createNote, deleteNote, fetchNotes, updateNote } from "./api";
import { fetchNoteLabels } from "./labelApi";
import { DeleteNoteDialog } from "./DeleteNoteDialog";
import { LABEL_COLOR_VAR } from "../labels/colors";
import type { NoteLabel, NoteSummary } from "./types";

type View = "active" | "archive";

// Sentinel labelFilter value for "notes with no label assigned", distinct
// from null (no filter, i.e. "All labels"). Real label ids are 32-char hex
// (see internal/id.New on the backend), so this can never collide.
const NO_LABEL_FILTER = "none";

function formatDate(value: string) {
  return new Date(value).toLocaleDateString(undefined, { year: "numeric", month: "short", day: "numeric" });
}

export function NotepadPage() {
  const navigate = useNavigate();
  const [view, setView] = useState<View>("active");
  const [notes, setNotes] = useState<NoteSummary[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [creating, setCreating] = useState(false);
  const [pendingDelete, setPendingDelete] = useState<NoteSummary | null>(null);
  const [search, setSearch] = useState("");
  const [pinnedOnly, setPinnedOnly] = useState(false);
  const [labels, setLabels] = useState<NoteLabel[]>([]);
  const [labelFilter, setLabelFilter] = useState<string | null>(null);
  const [showLabelMenu, setShowLabelMenu] = useState(false);
  const labelMenuRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    setLoading(true);
    setError(null);
    fetchNotes(view === "archive")
      .then(setNotes)
      .catch((err) => setError(err instanceof Error ? err.message : "Couldn't load notes."))
      .finally(() => setLoading(false));
  }, [view]);

  useEffect(() => {
    fetchNoteLabels()
      .then(setLabels)
      .catch((err) => setError(err instanceof Error ? err.message : "Couldn't load labels."));
  }, []);

  useEffect(() => {
    if (!showLabelMenu) return;

    function handlePointerDown(e: MouseEvent) {
      if (labelMenuRef.current && !labelMenuRef.current.contains(e.target as Node)) {
        setShowLabelMenu(false);
      }
    }

    document.addEventListener("mousedown", handlePointerDown);
    return () => document.removeEventListener("mousedown", handlePointerDown);
  }, [showLabelMenu]);

  function selectLabelFilter(next: string | null) {
    setLabelFilter(next);
    setShowLabelMenu(false);
  }

  const query = search.trim().toLowerCase();
  const visibleNotes = notes.filter((n) => {
    if (query && !n.title.toLowerCase().includes(query)) return false;
    if (view === "active" && pinnedOnly && !n.pinned) return false;
    if (labelFilter === NO_LABEL_FILTER && n.labelId !== null) return false;
    if (labelFilter && labelFilter !== NO_LABEL_FILTER && n.labelId !== labelFilter) return false;
    return true;
  });
  const activeLabel = labelFilter && labelFilter !== NO_LABEL_FILTER ? labels.find((l) => l.id === labelFilter) ?? null : null;
  const isNoLabelFilter = labelFilter === NO_LABEL_FILTER;

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
    setPendingDelete(null);
    deleteNote(id).catch((err) => {
      setNotes(previous);
      setError(err instanceof Error ? err.message : "Couldn't delete note.");
    });
  }

  function handleTogglePin(note: NoteSummary) {
    const nextPinned = !note.pinned;
    setNotes((prev) =>
      prev
        .map((n) => (n.id === note.id ? { ...n, pinned: nextPinned } : n))
        .sort((a, b) => {
          if (a.pinned !== b.pinned) return a.pinned ? -1 : 1;
          return new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime();
        }),
    );
    updateNote(note.id, { pinned: nextPinned }).catch((err) => {
      setNotes((prev) => prev.map((n) => (n.id === note.id ? { ...n, pinned: note.pinned } : n)));
      setError(err instanceof Error ? err.message : "Couldn't update note.");
    });
  }

  function handleToggleLock(note: NoteSummary) {
    const nextLocked = !note.locked;
    setNotes((prev) => prev.map((n) => (n.id === note.id ? { ...n, locked: nextLocked } : n)));
    updateNote(note.id, { locked: nextLocked }).catch((err) => {
      setNotes((prev) => prev.map((n) => (n.id === note.id ? { ...n, locked: note.locked } : n)));
      setError(err instanceof Error ? err.message : "Couldn't update note.");
    });
  }

  function handleToggleArchive(note: NoteSummary) {
    const nextArchived = view === "active";
    const previous = notes;
    setNotes((prev) => prev.filter((n) => n.id !== note.id));
    updateNote(note.id, { archived: nextArchived }).catch((err) => {
      setNotes(previous);
      setError(err instanceof Error ? err.message : "Couldn't update note.");
    });
  }

  const tabClass = (active: boolean) =>
    `rounded-lg px-3 py-1.5 text-[length:var(--text-caption)] transition-colors cursor-pointer ${
      active ? "bg-(--card-alt) text-(--fg)" : "text-(--text-muted) hover:text-(--fg)"
    }`;

  return (
    <div className="min-h-full lg:h-full flex flex-col">
      <div className="shrink-0 flex items-center justify-between gap-2 flex-wrap mb-3">
        <div className="flex items-center gap-1">
          <button onClick={() => setView("active")} className={tabClass(view === "active")}>
            Notes
          </button>
          <button onClick={() => setView("archive")} className={tabClass(view === "archive")}>
            Archive
          </button>
        </div>
        <div className="flex items-center gap-2">
          <button
            onClick={() => navigate("/notepad/scratch")}
            title="One always-there scratch pad, separate from your notes"
            className="flex items-center gap-1.5 rounded-lg border-(--line) border-[0.5px] border-solid px-3 py-1.5 text-[length:var(--text-caption)] text-(--text-muted) hover:text-(--fg) transition-colors cursor-pointer"
          >
            <NotebookPen size={14} />
            Random Notepad
          </button>
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
      </div>

      <div className="shrink-0 flex flex-wrap items-center gap-2 mb-3">
        <div className="relative flex-1 min-w-[180px]">
          <Search size={13} className="absolute left-2.5 top-1/2 -translate-y-1/2 text-(--text-faint)" />
          <input
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder={view === "archive" ? "Search archived notes" : "Search notes"}
            className="w-full rounded-lg border-(--line) border-[0.5px] border-solid bg-(--card) pl-8 pr-3 py-1.5 text-[length:var(--text-caption)] text-(--fg) placeholder:text-(--text-faint) outline-none focus:border-(--line-strong) transition-colors"
          />
        </div>

        {view === "active" && (
          <button
            onClick={() => setPinnedOnly((v) => !v)}
            aria-pressed={pinnedOnly}
            className={`flex items-center gap-1.5 rounded-lg border-[0.5px] border-solid px-3 py-1.5 text-[length:var(--text-caption)] transition-colors cursor-pointer ${
              pinnedOnly
                ? "border-(--label-yellow) text-(--label-yellow)"
                : "border-(--line) text-(--text-muted) hover:text-(--fg)"
            }`}
          >
            <Star size={13} className={pinnedOnly ? "fill-(--label-yellow)" : ""} />
            Pinned only
          </button>
        )}

        <div className="relative" ref={labelMenuRef}>
          <button
            onClick={() => setShowLabelMenu((v) => !v)}
            aria-label="Filter by label"
            className={`flex items-center gap-1.5 rounded-lg border-[0.5px] border-solid px-3 py-1.5 text-[length:var(--text-caption)] transition-colors cursor-pointer ${
              showLabelMenu || activeLabel || isNoLabelFilter
                ? "border-(--line-strong) text-(--fg)"
                : "border-(--line) text-(--text-muted) hover:text-(--fg)"
            }`}
          >
            {activeLabel ? (
              <>
                <span
                  className="size-2 rounded-full shrink-0"
                  style={{ backgroundColor: LABEL_COLOR_VAR[activeLabel.color] }}
                />
                {activeLabel.name}
              </>
            ) : isNoLabelFilter ? (
              <>
                <TagX size={13} />
                No label
              </>
            ) : (
              <>
                <Tag size={13} />
                Label
              </>
            )}
          </button>

          {showLabelMenu && (
            <div className="absolute right-0 z-10 mt-2 w-52 max-h-56 overflow-y-auto rounded-lg border-(--line) border-[0.5px] border-solid bg-(--card) p-1.5 shadow-lg themed-scrollbar">
              <button
                onClick={() => selectLabelFilter(null)}
                className={`flex w-full items-center rounded-lg px-3 py-1.5 text-left text-[length:var(--text-pill)] transition-colors cursor-pointer ${
                  !labelFilter
                    ? "bg-(--card-alt) text-(--fg)"
                    : "text-(--text-muted) hover:text-(--fg) hover:bg-(--card-alt)"
                }`}
              >
                All labels
              </button>
              <button
                onClick={() => selectLabelFilter(NO_LABEL_FILTER)}
                className={`flex w-full items-center gap-1.5 rounded-lg px-3 py-1.5 text-left text-[length:var(--text-pill)] transition-colors cursor-pointer ${
                  isNoLabelFilter
                    ? "bg-(--card-alt) text-(--fg)"
                    : "text-(--text-muted) hover:text-(--fg) hover:bg-(--card-alt)"
                }`}
              >
                <TagX size={12} className="shrink-0" />
                <span className="truncate flex-1">No label</span>
              </button>
              {labels.map((l) => (
                <button
                  key={l.id}
                  onClick={() => selectLabelFilter(l.id)}
                  className={`flex w-full items-center gap-1.5 rounded-lg px-3 py-1.5 text-left text-[length:var(--text-pill)] transition-colors cursor-pointer ${
                    labelFilter === l.id
                      ? "bg-(--card-alt) text-(--fg)"
                      : "text-(--text-muted) hover:text-(--fg) hover:bg-(--card-alt)"
                  }`}
                >
                  <span
                    className="size-2 rounded-full shrink-0"
                    style={{ backgroundColor: LABEL_COLOR_VAR[l.color] }}
                  />
                  <span className="truncate flex-1">{l.name}</span>
                </button>
              ))}
              {labels.length === 0 && (
                <p className="px-3 py-1.5 text-[length:var(--text-pill)] text-(--text-faint)">No labels yet.</p>
              )}
            </div>
          )}
        </div>
      </div>

      {error && (
        <div className="shrink-0 mb-3 flex items-center justify-between gap-2 rounded-lg border-(--line) border-[0.5px] border-solid bg-(--card) px-3 py-2 text-[length:var(--text-pill)] text-red-400">
          <span>{error}</span>
          <button onClick={() => setError(null)} className="shrink-0 text-(--text-faint) hover:text-(--fg) transition-colors cursor-pointer">
            Dismiss
          </button>
        </div>
      )}

      <div className="flex-1 lg:min-h-0 lg:overflow-y-auto themed-scrollbar">
        {loading && (
          <div className="bg-(--card) border-(--line) border-[0.5px] border-solid rounded-xl p-8 text-center text-(--text-faint) text-[length:var(--text-caption)]">
            Loading notes...
          </div>
        )}

        {!loading && visibleNotes.length === 0 && (
          <div className="bg-(--card) border-(--line) border-[0.5px] border-solid rounded-xl p-8 text-center text-(--text-faint) text-[length:var(--text-caption)]">
            {notes.length === 0
              ? view === "archive"
                ? "No archived notes."
                : "No notes yet."
              : "No notes match your filters."}
          </div>
        )}

        {!loading && visibleNotes.length > 0 && (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-2 pb-2">
            {visibleNotes.map((note) => {
              const noteLabel = labels.find((l) => l.id === note.labelId) ?? null;
              return (
              <div
                key={note.id}
                onClick={() => navigate(`/notepad/${note.id}`)}
                className="group relative flex flex-col gap-1 bg-(--card) border-(--line) border-[0.5px] border-solid rounded-lg px-3 py-2.5 cursor-pointer hover:border-(--line-strong) transition-colors"
              >
                <span className="flex items-center gap-1.5 pr-14 text-[length:var(--text-caption)] text-(--fg)">
                  {note.pinned && view === "active" && (
                    <Star size={12} className="shrink-0 fill-(--label-yellow) text-(--label-yellow)" />
                  )}
                  {note.locked && <Lock size={12} className="shrink-0 text-(--label-orange)" />}
                  <span className="truncate">{note.title}</span>
                </span>
                <div className="flex items-center gap-2">
                  <span className="text-[length:var(--text-pill)] text-(--text-faint)">{formatDate(note.createdAt)}</span>
                  {noteLabel && (
                    <span className="flex items-center gap-1 min-w-0 max-w-24 rounded-md bg-(--card-alt) px-1.5 py-0.5 text-[length:var(--text-pill)] text-(--text-muted)">
                      <span
                        className="size-1.5 rounded-full shrink-0"
                        style={{ backgroundColor: LABEL_COLOR_VAR[noteLabel.color] }}
                      />
                      <span className="truncate">{noteLabel.name}</span>
                    </span>
                  )}
                </div>

                <div className="absolute top-2 right-2 flex items-center gap-1.5 opacity-100 lg:opacity-0 lg:group-hover:opacity-100 transition-opacity">
                  {view === "active" && (
                    <button
                      onClick={(e) => {
                        e.stopPropagation();
                        handleTogglePin(note);
                      }}
                      aria-label={note.pinned ? "Unpin note" : "Pin note"}
                      className={`transition-colors cursor-pointer ${
                        note.pinned
                          ? "text-(--label-yellow)"
                          : "text-(--text-faint) hover:text-(--label-yellow)"
                      }`}
                    >
                      <Star size={13} className={note.pinned ? "fill-(--label-yellow)" : ""} />
                    </button>
                  )}
                  <button
                    onClick={(e) => {
                      e.stopPropagation();
                      handleToggleLock(note);
                    }}
                    aria-label={note.locked ? "Unlock note" : "Lock note"}
                    className={`transition-colors cursor-pointer ${
                      note.locked
                        ? "text-(--label-orange)"
                        : "text-(--text-faint) hover:text-(--fg)"
                    }`}
                  >
                    {note.locked ? <Lock size={13} /> : <LockOpen size={13} />}
                  </button>
                  <button
                    onClick={(e) => {
                      e.stopPropagation();
                      handleToggleArchive(note);
                    }}
                    aria-label={view === "archive" ? "Unarchive note" : "Archive note"}
                    className="text-(--text-faint) hover:text-(--fg) transition-colors cursor-pointer"
                  >
                    {view === "archive" ? <ArchiveRestore size={13} /> : <Archive size={13} />}
                  </button>
                  <button
                    onClick={(e) => {
                      e.stopPropagation();
                      setPendingDelete(note);
                    }}
                    aria-label="Delete note"
                    className="text-(--text-faint) hover:text-(--label-red) transition-colors cursor-pointer"
                  >
                    <Trash2 size={13} />
                  </button>
                </div>
              </div>
              );
            })}
          </div>
        )}
      </div>

      {pendingDelete && (
        <DeleteNoteDialog
          title={pendingDelete.title}
          onCancel={() => setPendingDelete(null)}
          onConfirm={() => handleDelete(pendingDelete.id)}
        />
      )}
    </div>
  );
}
