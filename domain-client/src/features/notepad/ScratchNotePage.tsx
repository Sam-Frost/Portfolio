import { useEffect, useRef, useState } from "react";
import { useNavigate } from "react-router-dom";
import { ArrowLeft } from "lucide-react";
import { fetchScratchNote, updateNote } from "./api";
import { RichTextEditor } from "../../components/domain/RichTextEditor";
import type { Note } from "./types";

type SaveStatus = "saved" | "saving" | "error";

const AUTOSAVE_DELAY_MS = 700;

const STATUS_COPY: Record<SaveStatus, { label: string; dotClass: string }> = {
  saved: { label: "All changes saved", dotClass: "bg-(--green)" },
  saving: { label: "Saving...", dotClass: "bg-(--label-orange)" },
  error: { label: "Couldn't save", dotClass: "bg-(--label-red)" },
};

// The "Random Notepad": one always-there, title-less scratch buffer, kept
// deliberately apart from the organized notes list. It's a single server-side
// note (fetchScratchNote get-or-creates it) edited through the same debounced
// autosave path as NoteEditorPage — minus the title, pin, and archive
// affordances, none of which make sense for a scratch pad.
export function ScratchNotePage() {
  const navigate = useNavigate();

  const [note, setNote] = useState<Note | null>(null);
  const [loading, setLoading] = useState(true);
  const [loadError, setLoadError] = useState<string | null>(null);
  const [status, setStatus] = useState<SaveStatus>("saved");

  const idRef = useRef<string | null>(null);
  const pendingRef = useRef<string | null>(null);
  const timeoutRef = useRef<number | null>(null);

  useEffect(() => {
    setLoading(true);
    fetchScratchNote()
      .then((n) => {
        setNote(n);
        idRef.current = n.id;
      })
      .catch((err) => setLoadError(err instanceof Error ? err.message : "Couldn't open the notepad."))
      .finally(() => setLoading(false));
  }, []);

  function flush() {
    if (timeoutRef.current) window.clearTimeout(timeoutRef.current);
    timeoutRef.current = null;
    const id = idRef.current;
    if (!id || pendingRef.current === null) return;
    const html = pendingRef.current;
    pendingRef.current = null;
    updateNote(id, { contentHtml: html }).catch(() => {});
  }

  // Best-effort save of whatever's still in the debounce window when the
  // page unmounts (navigating back to the list, or away entirely).
  useEffect(() => {
    return () => flush();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  function handleContentChange(html: string) {
    const id = idRef.current;
    if (!id) return;
    pendingRef.current = html;
    setStatus("saving");
    if (timeoutRef.current) window.clearTimeout(timeoutRef.current);
    timeoutRef.current = window.setTimeout(() => {
      const toSend = pendingRef.current ?? "";
      pendingRef.current = null;
      updateNote(id, { contentHtml: toSend })
        .then(() => setStatus("saved"))
        .catch(() => setStatus("error"));
    }, AUTOSAVE_DELAY_MS);
  }

  if (loading) {
    return <div className="text-(--text-faint) text-[length:var(--text-caption)]">Loading notepad...</div>;
  }

  if (loadError || !note) {
    return (
      <div className="text-(--text-faint) text-[length:var(--text-caption)]">
        {loadError ?? "Couldn't open the notepad."}
      </div>
    );
  }

  const { label, dotClass } = STATUS_COPY[status];

  return (
    <div className="min-h-full lg:h-full flex flex-col">
      <div className="shrink-0 flex items-center justify-between gap-3 flex-wrap mb-4">
        <button
          onClick={() => navigate("/notepad")}
          className="flex items-center gap-1.5 text-[length:var(--text-caption)] text-(--text-muted) hover:text-(--fg) transition-colors cursor-pointer"
        >
          <ArrowLeft size={14} />
          All notes
        </button>

        <div className="flex items-center gap-1.5 text-[length:var(--text-pill)] text-(--text-faint)">
          <span className={`size-2 rounded-full shrink-0 ${dotClass}`} />
          {label}
        </div>
      </div>

      <span className="shrink-0 mb-3 text-xl font-space font-semibold text-(--text-faint)">Random Notepad</span>

      <RichTextEditor key={note.id} initialContentHtml={note.contentHtml} onChange={handleContentChange} />
    </div>
  );
}
