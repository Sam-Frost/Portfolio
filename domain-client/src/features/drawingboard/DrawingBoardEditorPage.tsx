import { useCallback, useEffect, useRef, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { Excalidraw } from "@excalidraw/excalidraw";
import "@excalidraw/excalidraw/index.css";
import type { ExcalidrawElement } from "@excalidraw/excalidraw/element/types";
import type { AppState, BinaryFiles } from "@excalidraw/excalidraw/types";
import { ArrowLeft, Trash2 } from "lucide-react";
import { deleteDrawingBoard, fetchDrawingBoard, updateDrawingBoard } from "./api";
import { DeleteBoardDialog } from "./DeleteBoardDialog";
import type { DrawingBoard, DrawingBoardScene } from "./types";

type SaveStatus = "saved" | "saving" | "error";

const AUTOSAVE_DELAY_MS = 700;

const STATUS_COPY: Record<SaveStatus, { label: string; dotClass: string }> = {
  saved: { label: "All changes saved", dotClass: "bg-(--green)" },
  saving: { label: "Saving...", dotClass: "bg-(--label-orange)" },
  error: { label: "Couldn't save", dotClass: "bg-(--label-red)" },
};

// AppState carries a lot of purely-runtime UI state alongside the fields
// worth persisting (e.g. `collaborators` is a Map, which JSON.stringify
// silently turns into "{}") — this is the subset that actually matters for
// restoring the viewport next time the board is opened.
function toStoredAppState(appState: AppState): DrawingBoardScene["appState"] {
  return {
    viewBackgroundColor: appState.viewBackgroundColor,
    scrollX: appState.scrollX,
    scrollY: appState.scrollY,
    zoom: { value: appState.zoom.value },
  };
}

export function DrawingBoardEditorPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();

  const [board, setBoard] = useState<DrawingBoard | null>(null);
  const [name, setName] = useState("");
  const [loading, setLoading] = useState(true);
  const [loadError, setLoadError] = useState<string | null>(null);
  const [status, setStatus] = useState<SaveStatus>("saved");
  const [confirmingDelete, setConfirmingDelete] = useState(false);

  const pendingRef = useRef<{ name?: string; sceneData?: DrawingBoardScene }>({});
  const timeoutRef = useRef<number | null>(null);

  useEffect(() => {
    if (!id) return;
    setLoading(true);
    fetchDrawingBoard(id)
      .then((b) => {
        setBoard(b);
        setName(b.name);
      })
      .catch((err) => setLoadError(err instanceof Error ? err.message : "Couldn't load board."))
      .finally(() => setLoading(false));
  }, [id]);

  function flush() {
    if (timeoutRef.current) window.clearTimeout(timeoutRef.current);
    timeoutRef.current = null;
    if (!id || Object.keys(pendingRef.current).length === 0) return;
    const patch = pendingRef.current;
    pendingRef.current = {};
    updateDrawingBoard(id, patch).catch(() => {});
  }

  // Best-effort save of whatever's still pending when the editor unmounts
  // (navigating back to the list, or away entirely) so a quick exit right
  // after drawing/renaming doesn't lose the last debounce window.
  useEffect(() => {
    return () => flush();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [id]);

  // Scene changes (onChange, fired continuously while drawing) and name
  // edits share one debounce timer and merge into one pending patch — like
  // notepad's autosave — so a rename landing mid-drag can't silently cancel
  // the scene save that was about to fire, or vice versa.
  function scheduleSave(patch: { name?: string; sceneData?: DrawingBoardScene }) {
    if (!id) return;
    pendingRef.current = { ...pendingRef.current, ...patch };
    setStatus("saving");
    if (timeoutRef.current) window.clearTimeout(timeoutRef.current);
    timeoutRef.current = window.setTimeout(() => {
      const toSend = pendingRef.current;
      pendingRef.current = {};
      updateDrawingBoard(id, toSend)
        .then(() => setStatus("saved"))
        .catch(() => setStatus("error"));
    }, AUTOSAVE_DELAY_MS);
  }

  // Stabilized with useCallback: Excalidraw re-syncs against a changed
  // onChange reference, so an inline function here (recreated on every
  // status/name re-render) causes it to immediately re-fire onChange with
  // the current scene — which schedules another save, triggers another
  // status re-render, and loops forever. id/status/name never change
  // mid-editing-session in a way that needs a fresh closure.
  const handleChange = useCallback(
    (elements: readonly ExcalidrawElement[], appState: AppState, files: BinaryFiles) => {
      scheduleSave({ sceneData: { elements, appState: toStoredAppState(appState), files } });
    },
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [id],
  );

  function handleNameChange(value: string) {
    setName(value);
    scheduleSave({ name: value });
  }

  function handleDelete() {
    if (!id) return;
    deleteDrawingBoard(id)
      .then(() => navigate("/drawing-board"))
      .catch((err) => setLoadError(err instanceof Error ? err.message : "Couldn't delete board."));
  }

  if (loading) {
    return (
      <div className="min-h-full lg:h-full flex items-center justify-center text-(--text-faint) text-[length:var(--text-caption)]">
        Loading board...
      </div>
    );
  }

  if (loadError || !board) {
    return (
      <div className="min-h-full lg:h-full flex flex-col items-center justify-center gap-3 text-(--text-faint) text-[length:var(--text-caption)]">
        <span>{loadError ?? "Board not found."}</span>
        <button
          onClick={() => navigate("/drawing-board")}
          className="rounded-lg border-(--line) border-[0.5px] border-solid px-3 py-1.5 text-(--text-muted) hover:text-(--fg) transition-colors cursor-pointer"
        >
          Back to boards
        </button>
      </div>
    );
  }

  const statusCopy = STATUS_COPY[status];

  return (
    <div className="min-h-full lg:h-full flex flex-col gap-3">
      <div className="shrink-0 flex items-center gap-2">
        <button
          onClick={() => navigate("/drawing-board")}
          aria-label="Back to boards"
          className="shrink-0 text-(--text-faint) hover:text-(--fg) transition-colors cursor-pointer"
        >
          <ArrowLeft size={16} />
        </button>
        <input
          value={name}
          onChange={(e) => handleNameChange(e.target.value)}
          className="flex-1 min-w-0 bg-transparent text-[length:var(--text-caption)] text-(--fg) outline-none"
          aria-label="Board name"
        />
        <span className="hidden sm:flex items-center gap-1.5 shrink-0 text-[length:var(--text-pill)] text-(--text-faint)">
          <span className={`size-1.5 rounded-full ${statusCopy.dotClass}`} />
          {statusCopy.label}
        </span>
        <button
          onClick={() => setConfirmingDelete(true)}
          aria-label="Delete board"
          className="shrink-0 text-(--text-faint) hover:text-(--label-red) transition-colors cursor-pointer"
        >
          <Trash2 size={15} />
        </button>
      </div>

      <div className="relative flex-1 min-h-[70vh] lg:min-h-0 rounded-xl overflow-hidden border-(--line) border-[0.5px] border-solid">
        <Excalidraw
          initialData={{
            elements: board.sceneData.elements,
            // Persisted appState round-trips through plain JSON, which
            // loses AppState's branded `zoom.value` (NormalizedZoomValue)
            // type — Excalidraw normalizes it back on load regardless.
            appState: board.sceneData.appState as Partial<AppState>,
            files: board.sceneData.files,
          }}
          onChange={handleChange}
          theme="dark"
        />
      </div>

      {confirmingDelete && (
        <DeleteBoardDialog
          name={name}
          onCancel={() => setConfirmingDelete(false)}
          onConfirm={handleDelete}
        />
      )}
    </div>
  );
}
