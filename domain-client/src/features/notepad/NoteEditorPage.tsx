import { useEffect, useRef, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { Archive, ArchiveRestore, ArrowLeft, Lock, LockOpen, Star } from "lucide-react";
import { fetchNote, updateNote } from "./api";
import { RichTextEditor } from "../../components/domain/RichTextEditor";
import type { Note } from "./types";

type SaveStatus = "saved" | "saving" | "error";

const AUTOSAVE_DELAY_MS = 700;

const STATUS_COPY: Record<SaveStatus, { label: string; dotClass: string }> = {
  saved: { label: "All changes saved", dotClass: "bg-(--green)" },
  saving: { label: "Saving...", dotClass: "bg-(--label-orange)" },
  error: { label: "Couldn't save", dotClass: "bg-(--label-red)" },
};

export function NoteEditorPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();

  const [note, setNote] = useState<Note | null>(null);
  const [title, setTitle] = useState("");
  const [loading, setLoading] = useState(true);
  const [loadError, setLoadError] = useState<string | null>(null);
  const [status, setStatus] = useState<SaveStatus>("saved");

  const pendingRef = useRef<{ title?: string; contentHtml?: string }>({});
  const timeoutRef = useRef<number | null>(null);

  useEffect(() => {
    if (!id) return;
    setLoading(true);
    fetchNote(id)
      .then((n) => {
        setNote(n);
        setTitle(n.title);
      })
      .catch((err) => setLoadError(err instanceof Error ? err.message : "Couldn't load note."))
      .finally(() => setLoading(false));
  }, [id]);

  function flush() {
    if (timeoutRef.current) window.clearTimeout(timeoutRef.current);
    timeoutRef.current = null;
    if (!id || Object.keys(pendingRef.current).length === 0) return;
    const patch = pendingRef.current;
    pendingRef.current = {};
    updateNote(id, patch).catch(() => {});
  }

  // Best-effort save of whatever's still pending when the editor unmounts
  // (navigating back to the list, or away entirely) so a quick exit right
  // after typing doesn't lose the last debounce window.
  useEffect(() => {
    return () => flush();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [id]);

  const locked = note?.locked ?? false;

  function scheduleSave(patch: { title?: string; contentHtml?: string }) {
    if (!id || locked) return;
    pendingRef.current = { ...pendingRef.current, ...patch };
    setStatus("saving");
    if (timeoutRef.current) window.clearTimeout(timeoutRef.current);
    timeoutRef.current = window.setTimeout(() => {
      const toSend = pendingRef.current;
      pendingRef.current = {};
      updateNote(id, toSend)
        .then(() => setStatus("saved"))
        .catch(() => setStatus("error"));
    }, AUTOSAVE_DELAY_MS);
  }

  function handleTitleChange(value: string) {
    setTitle(value);
    scheduleSave({ title: value });
  }

  function handleContentChange(html: string) {
    scheduleSave({ contentHtml: html });
  }

  function handleTogglePin() {
    if (!id || !note) return;
    const nextPinned = !note.pinned;
    setNote({ ...note, pinned: nextPinned });
    updateNote(id, { pinned: nextPinned })
      .then((updated) => setNote(updated))
      .catch(() => setNote((prev) => (prev ? { ...prev, pinned: !nextPinned } : prev)));
  }

  function handleToggleLock() {
    if (!id || !note) return;
    const nextLocked = !note.locked;
    // Locking freezes the note as-is: fold any still-pending edits into the
    // same request, since the server rejects content writes once locked.
    if (timeoutRef.current) window.clearTimeout(timeoutRef.current);
    timeoutRef.current = null;
    const patch = { ...pendingRef.current, locked: nextLocked };
    pendingRef.current = {};
    setNote({ ...note, locked: nextLocked });
    setStatus("saving");
    updateNote(id, patch)
      .then((updated) => {
        setNote(updated);
        setTitle(updated.title);
        setStatus("saved");
      })
      .catch(() => {
        setNote((prev) => (prev ? { ...prev, locked: !nextLocked } : prev));
        setStatus("error");
      });
  }

  function handleToggleArchive() {
    if (!id || !note) return;
    const nextArchived = !note.archived;
    updateNote(id, { archived: nextArchived })
      .then(() => navigate("/notepad"))
      .catch(() => setNote((prev) => (prev ? { ...prev, archived: !nextArchived } : prev)));
  }

  if (loading) {
    return <div className="text-(--text-faint) text-[length:var(--text-caption)]">Loading note...</div>;
  }

  if (loadError || !note) {
    return (
      <div className="text-(--text-faint) text-[length:var(--text-caption)]">
        {loadError ?? "Note not found."}
      </div>
    );
  }

  const { label, dotClass } = STATUS_COPY[status];

  return (
    <div className="h-full flex flex-col">
      <div className="shrink-0 flex items-center justify-between mb-4">
        <button
          onClick={() => navigate("/notepad")}
          className="flex items-center gap-1.5 text-[length:var(--text-caption)] text-(--text-muted) hover:text-(--fg) transition-colors cursor-pointer"
        >
          <ArrowLeft size={14} />
          All notes
        </button>

        <div className="flex items-center gap-3">
          <div className="flex items-center gap-1.5 text-[length:var(--text-pill)] text-(--text-faint)">
            {locked ? (
              <>
                <Lock size={12} className="shrink-0" />
                View only
              </>
            ) : (
              <>
                <span className={`size-2 rounded-full shrink-0 ${dotClass}`} />
                {label}
              </>
            )}
          </div>

          <button
            onClick={handleToggleLock}
            aria-label={locked ? "Unlock note" : "Lock note"}
            aria-pressed={locked}
            className={`transition-colors cursor-pointer ${
              locked ? "text-(--label-orange)" : "text-(--text-faint) hover:text-(--fg)"
            }`}
          >
            {locked ? <Lock size={15} /> : <LockOpen size={15} />}
          </button>

          <button
            onClick={handleTogglePin}
            aria-label={note.pinned ? "Unpin note" : "Pin note"}
            aria-pressed={note.pinned}
            className={`transition-colors cursor-pointer ${
              note.pinned ? "text-(--label-yellow)" : "text-(--text-faint) hover:text-(--label-yellow)"
            }`}
          >
            <Star size={15} className={note.pinned ? "fill-(--label-yellow)" : ""} />
          </button>

          <button
            onClick={handleToggleArchive}
            aria-label={note.archived ? "Unarchive note" : "Archive note"}
            className="text-(--text-faint) hover:text-(--fg) transition-colors cursor-pointer"
          >
            {note.archived ? <ArchiveRestore size={15} /> : <Archive size={15} />}
          </button>
        </div>
      </div>

      <input
        value={title}
        onChange={(e) => handleTitleChange(e.target.value)}
        readOnly={locked}
        placeholder="Untitled"
        className="shrink-0 mb-3 bg-transparent outline-none text-xl font-space font-semibold text-(--fg) placeholder:text-(--text-faint) read-only:cursor-default"
      />

      <RichTextEditor
        key={note.id}
        initialContentHtml={note.contentHtml}
        onChange={handleContentChange}
        readOnly={locked}
      />
    </div>
  );
}
