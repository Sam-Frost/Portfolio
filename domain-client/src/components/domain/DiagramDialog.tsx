import { useCallback, useRef, useState } from "react";
import { Excalidraw, exportToBlob } from "@excalidraw/excalidraw";
import "@excalidraw/excalidraw/index.css";
import type { ExcalidrawElement, ExcalidrawFrameLikeElement } from "@excalidraw/excalidraw/element/types";
import type { AppState, BinaryFiles } from "@excalidraw/excalidraw/types";
import { X, Check } from "lucide-react";
import type { DrawingBoardScene, DrawingBoardSceneAppState } from "../../features/drawingboard/types";

interface DiagramDialogProps {
  // Present when reopening an already-embedded diagram to edit it; absent
  // for a brand new one.
  initialScene?: DrawingBoardScene;
  onCancel: () => void;
  // scene is the raw Excalidraw data (kept around so the diagram can be
  // reopened and edited later); pngDataUrl is a rendered snapshot to embed
  // as the note's visible <img>.
  onSave: (scene: DrawingBoardScene, pngDataUrl: string) => void;
}

// Mirrors drawingboard's DrawingBoardEditorPage: AppState carries a lot of
// purely-runtime UI state (e.g. `collaborators` is a Map, which
// JSON.stringify silently turns into "{}") — this is the subset worth
// persisting to restore the viewport.
function toStoredAppState(appState: AppState): DrawingBoardSceneAppState {
  return {
    viewBackgroundColor: appState.viewBackgroundColor,
    scrollX: appState.scrollX,
    scrollY: appState.scrollY,
    zoom: { value: appState.zoom.value },
  };
}

function isFrameLike(el: ExcalidrawElement): el is ExcalidrawFrameLikeElement {
  return el.type === "frame" || el.type === "magicframe";
}

// Decides which slice of the canvas becomes the embedded PNG. The full
// scene is always stored for editing regardless of what this returns —
// this only narrows the rasterized snapshot:
//   - exactly one frame  -> crop to that frame (Excalidraw's own region tool)
//   - else a selection   -> just the selected elements (+ their bound text /
//                           frame children), so you can lasso the bit you want
//   - else               -> the whole drawing (previous behaviour)
function resolveExportScope(
  elements: readonly ExcalidrawElement[],
  appState: AppState,
): { elements: readonly ExcalidrawElement[]; exportingFrame: ExcalidrawFrameLikeElement | null; label: string } {
  const frames = elements.filter(isFrameLike);
  if (frames.length === 1) {
    return { elements, exportingFrame: frames[0], label: "Frame" };
  }

  const selectedIds = new Set(
    Object.entries(appState.selectedElementIds ?? {})
      .filter(([, on]) => on)
      .map(([id]) => id),
  );
  if (selectedIds.size > 0) {
    const scoped = elements.filter((el) => {
      if (selectedIds.has(el.id)) return true;
      const containerId = "containerId" in el ? el.containerId : null;
      if (containerId != null && selectedIds.has(containerId)) return true;
      if (el.frameId != null && selectedIds.has(el.frameId)) return true;
      return false;
    });
    if (scoped.length > 0) {
      return { elements: scoped, exportingFrame: null, label: "Selection" };
    }
  }

  return { elements, exportingFrame: null, label: "Whole drawing" };
}

function blobToDataUrl(blob: Blob): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onload = () => resolve(reader.result as string);
    reader.onerror = () => reject(reader.error);
    reader.readAsDataURL(blob);
  });
}

// A full-screen Excalidraw canvas for drawing a diagram to embed in a note.
// Saving rasterizes the scene to a PNG (what actually shows up in the
// note's contentHtml) while also handing back the raw scene data, which the
// caller stores alongside the image so the diagram can be reopened here and
// edited rather than just replaced.
export function DiagramDialog({ initialScene, onCancel, onSave }: DiagramDialogProps) {
  const sceneRef = useRef<{ elements: readonly ExcalidrawElement[]; appState: AppState; files: BinaryFiles } | null>(
    null,
  );
  const [saving, setSaving] = useState(false);
  // What the current canvas state would rasterize to on save — shown next to
  // the Insert button so it's clear a selection/frame narrows the snapshot.
  const [exportLabel, setExportLabel] = useState("Whole drawing");

  // Stabilized with useCallback like drawingboard's handleChange — an
  // inline function recreated on every render makes Excalidraw treat
  // onChange as "changed" and re-sync, which fires onChange again in a loop.
  const handleChange = useCallback((elements: readonly ExcalidrawElement[], appState: AppState, files: BinaryFiles) => {
    sceneRef.current = { elements, appState, files };
    setExportLabel(elements.length === 0 ? "Whole drawing" : resolveExportScope(elements, appState).label);
  }, []);

  async function handleSave() {
    const scene = sceneRef.current;
    if (!scene || scene.elements.length === 0) {
      onCancel();
      return;
    }
    setSaving(true);
    try {
      const { elements: exportElements, exportingFrame } = resolveExportScope(scene.elements, scene.appState);
      const blob = await exportToBlob({
        elements: exportElements,
        exportingFrame,
        // The canvas is always themed dark here (see `theme` below), so a
        // stroke that reads as white on-screen is stored as near-black and
        // only inverted for display. Export with the same dark filter so it
        // stays white — otherwise it bakes out as grey/black and vanishes
        // against the note's dark card. exportBackground:false keeps the
        // PNG transparent so it blends with the note either way.
        appState: { ...scene.appState, exportWithDarkMode: true, exportBackground: false },
        files: scene.files,
        mimeType: "image/png",
      });
      const dataUrl = await blobToDataUrl(blob);
      onSave({ elements: scene.elements, appState: toStoredAppState(scene.appState), files: scene.files }, dataUrl);
    } finally {
      setSaving(false);
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex flex-col bg-(--bg)">
      <div className="shrink-0 flex items-center justify-between gap-3 px-4 py-2.5 border-b-(--line) border-b-[0.5px] border-solid">
        <span className="text-[length:var(--text-caption)] text-(--fg)">Diagram</span>
        <div className="flex items-center gap-1.5">
          <span
            className="hidden sm:inline text-[length:var(--text-pill)] text-(--text-muted)"
            title="Select elements, or add a frame (F), to embed just part of the drawing. The full drawing is always kept for editing."
          >
            Embeds: {exportLabel}
          </span>
          <button
            type="button"
            onClick={onCancel}
            className="flex items-center gap-1.5 rounded-md px-2.5 py-1 text-[length:var(--text-pill)] text-(--text-muted) hover:text-(--fg) hover:bg-(--card-alt) transition-colors cursor-pointer"
          >
            <X size={13} />
            Cancel
          </button>
          <button
            type="button"
            onClick={handleSave}
            disabled={saving}
            className="flex items-center gap-1.5 rounded-md bg-(--fg) text-(--bg) px-2.5 py-1 text-[length:var(--text-pill)] cursor-pointer disabled:opacity-60"
          >
            <Check size={13} />
            {saving ? "Saving..." : initialScene ? "Save" : "Insert"}
          </button>
        </div>
      </div>

      <div className="diagram-dialog__canvas relative flex-1 min-h-0">
        <Excalidraw
          initialData={
            initialScene
              ? { elements: initialScene.elements, appState: initialScene.appState as Partial<AppState>, files: initialScene.files }
              // A brand new diagram otherwise defaults to Excalidraw's own
              // white canvas regardless of the dark `theme` below — that's
              // just the UI chrome, not the drawing surface — which then
              // bakes a white background into the exported PNG. Transparent
              // instead, so both the canvas and the embedded image blend
              // with the note.
              : { appState: { viewBackgroundColor: "transparent" } }
          }
          onChange={handleChange}
          theme="dark"
        />
      </div>
    </div>
  );
}
