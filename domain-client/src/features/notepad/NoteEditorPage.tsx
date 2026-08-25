import { useEffect, useRef, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { ArrowLeft } from "lucide-react";
import { fetchNote, updateNote } from "./api";
import { NotepadToolbar } from "./NotepadToolbar";
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

  const editorRef = useRef<HTMLDivElement | null>(null);
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

  // Runs after the fetch above commits `note` and the contentEditable div
  // has mounted — setting innerHTML directly in the fetch callback above
  // would silently no-op, since that div doesn't exist yet while `loading`
  // is still true.
  useEffect(() => {
    if (note && editorRef.current) {
      editorRef.current.innerHTML = note.contentHtml;
    }
  }, [note]);

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

  function scheduleSave(patch: { title?: string; contentHtml?: string }) {
    if (!id) return;
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

  function handleContentChange() {
    scheduleSave({ contentHtml: editorRef.current?.innerHTML ?? "" });
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

        <div className="flex items-center gap-1.5 text-[length:var(--text-pill)] text-(--text-faint)">
          <span className={`size-2 rounded-full shrink-0 ${dotClass}`} />
          {label}
        </div>
      </div>

      <input
        value={title}
        onChange={(e) => handleTitleChange(e.target.value)}
        placeholder="Untitled"
        className="shrink-0 mb-3 bg-transparent outline-none text-xl font-space font-semibold text-(--fg) placeholder:text-(--text-faint)"
      />

      <div className="shrink-0">
        <NotepadToolbar editorRef={editorRef} onChange={handleContentChange} />
      </div>

      <div
        ref={editorRef}
        contentEditable
        suppressContentEditableWarning
        onInput={handleContentChange}
        className="flex-1 min-h-0 overflow-y-auto themed-scrollbar rounded-lg border-(--line) border-[0.5px] border-solid bg-(--card) px-4 py-3 text-[length:var(--text-caption)] text-(--fg) outline-none [&_ul]:list-disc [&_ul]:pl-5 [&_ol]:list-decimal [&_ol]:pl-5 [&_h1]:text-xl [&_h1]:font-semibold [&_h2]:text-lg [&_h2]:font-semibold [&_h3]:text-base [&_h3]:font-semibold"
      />
    </div>
  );
}
