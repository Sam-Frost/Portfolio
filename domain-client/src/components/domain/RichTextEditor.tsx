import { useEffect, useRef, useState } from "react";
import { RichTextToolbar } from "./RichTextToolbar";
import { insertImageFile, insertImageDataUrl, wrapImageInFigure } from "./insertImage";
import { DiagramDialog } from "./DiagramDialog";
import { DIAGRAM_SCENE_ATTR, encodeScene, decodeScene } from "./diagramImage";
import type { DrawingBoardScene } from "../../features/drawingboard/types";

const MIN_IMAGE_WIDTH = 60;

// The eight drag handles drawn around a hovered image, MS-Word style. Each
// carries the outward direction it pulls in (fx/fy in -1..1) — the width
// delta while dragging is the pointer movement projected onto that
// direction, so every handle (corner or edge) resizes the image about its
// centre while its own aspect ratio is preserved (height stays "auto").
const RESIZE_HANDLES = [
  { key: "nw", fx: -1, fy: -1, cursor: "nwse-resize", cx: 0, cy: 0 },
  { key: "n", fx: 0, fy: -1, cursor: "ns-resize", cx: 0.5, cy: 0 },
  { key: "ne", fx: 1, fy: -1, cursor: "nesw-resize", cx: 1, cy: 0 },
  { key: "e", fx: 1, fy: 0, cursor: "ew-resize", cx: 1, cy: 0.5 },
  { key: "se", fx: 1, fy: 1, cursor: "nwse-resize", cx: 1, cy: 1 },
  { key: "s", fx: 0, fy: 1, cursor: "ns-resize", cx: 0.5, cy: 1 },
  { key: "sw", fx: -1, fy: 1, cursor: "nesw-resize", cx: 0, cy: 1 },
  { key: "w", fx: -1, fy: 0, cursor: "ew-resize", cx: 0, cy: 0.5 },
] as const;

interface OverlayRect {
  x: number;
  y: number;
  width: number;
  height: number;
}

interface RichTextEditorProps {
  // Content the editor is seeded with on mount. This is an *uncontrolled*
  // editor (like notepad's original contentEditable div) — changing this
  // prop after mount does nothing, since re-syncing on every change would
  // stomp the user's cursor position while they type. Callers that need to
  // load different content into a fresh editor (switching notes, switching
  // diary days) should force a remount with a `key` prop tied to the
  // resource's id/date.
  initialContentHtml: string;
  onChange: (html: string) => void;
  // Renders the content without the toolbar or contentEditable, for a day/
  // note whose edit window has closed (diary) — view-only, same markup.
  readOnly?: boolean;
  className?: string;
  // Notepad-only: shows the "insert diagram" toolbar button and lets
  // double-clicking an embedded diagram reopen it in Excalidraw for
  // editing. Left off for surfaces (diary) that don't want the extra
  // weight of pulling in Excalidraw.
  allowDiagrams?: boolean;
}

// The rich-text surface shared by every domain feature that edits HTML
// content (notepad, diary, ...): a document.execCommand-driven toolbar
// (RichTextToolbar) plus a contentEditable div. Originally lived inline in
// notepad's NoteEditorPage; extracted here so diary can reuse the same
// editor rather than reimplementing one.
export function RichTextEditor({
  initialContentHtml,
  onChange,
  readOnly = false,
  className,
  allowDiagrams = false,
}: RichTextEditorProps) {
  const editorRef = useRef<HTMLDivElement | null>(null);
  // The image currently under the pointer (or being dragged), and the
  // screen position of the resize handle drawn over its bottom-right
  // corner. Kept as an overlay — position: fixed, computed from
  // getBoundingClientRect — rather than markup saved into the image itself,
  // so contentHtml stays plain <img> tags with no editor-only cruft.
  const activeImageRef = useRef<HTMLImageElement | null>(null);
  const draggingRef = useRef(false);
  // The pointer is currently over the selection box / one of its handles,
  // and a pending "hide the box" timer. Together these give the overlay a
  // bit of hover-intent so it doesn't flicker as the pointer crosses
  // between the image and the handles sitting on its edges (see
  // scheduleHide).
  const overOverlayRef = useRef(false);
  const hideTimerRef = useRef<number | null>(null);
  // Screen-space rect of the hovered image, used to draw the selection box
  // and its handles as a position:fixed overlay — computed from
  // getBoundingClientRect rather than saved into the markup, so contentHtml
  // stays plain <img> tags with no editor-only cruft.
  const [overlayRect, setOverlayRect] = useState<OverlayRect | null>(null);

  // Diagram dialog state: "new" for a blank diagram, or the <img> being
  // re-opened for editing (its data-excalidraw-scene attribute supplies the
  // dialog's initial scene). null means the dialog is closed.
  const [diagramTarget, setDiagramTarget] = useState<"new" | HTMLImageElement | null>(null);
  // The caret position at the moment "Insert diagram" was clicked. The
  // dialog steals focus (and with it, the browser's one document-wide
  // Selection) for as long as it's open, so without this a new diagram
  // would land wherever the selection happens to collapse to instead of
  // where the user was actually typing.
  const savedRangeRef = useRef<Range | null>(null);

  // Mount-only: seeds the div's innerHTML once. See the initialContentHtml
  // doc comment above for why this doesn't re-run on prop changes.
  useEffect(() => {
    if (editorRef.current) {
      editorRef.current.innerHTML = initialContentHtml;
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  function handleInput() {
    onChange(editorRef.current?.innerHTML ?? "");
  }

  // Pasted/dropped images are inserted as data URLs (see insertImage.ts)
  // instead of falling through to the browser's default paste/drop
  // handling, which would either drop the image or leave a bare file path.
  function handlePaste(e: React.ClipboardEvent<HTMLDivElement>) {
    const file = Array.from(e.clipboardData.items)
      .find((item) => item.type.startsWith("image/"))
      ?.getAsFile();
    if (!file || !editorRef.current) return;
    e.preventDefault();
    insertImageFile(file, editorRef.current, handleInput);
  }

  function handleDrop(e: React.DragEvent<HTMLDivElement>) {
    const file = Array.from(e.dataTransfer.files).find((f) => f.type.startsWith("image/"));
    if (!file || !editorRef.current) return;
    e.preventDefault();

    // Drop the cursor where the image was released, rather than wherever it
    // last happened to be, so the image lands under the pointer.
    const range = document.caretRangeFromPoint?.(e.clientX, e.clientY);
    if (range) {
      const selection = window.getSelection();
      selection?.removeAllRanges();
      selection?.addRange(range);
    }
    insertImageFile(file, editorRef.current, handleInput);
  }

  function updateOverlay(img: HTMLImageElement) {
    const rect = img.getBoundingClientRect();
    const editor = editorRef.current;
    // The overlay is a position:fixed layer, so unlike the image it isn't
    // clipped by the editor's own scroll box. Once the image scrolls
    // *entirely* out of the editor's viewport, drop the overlay rather than
    // let it hang over the toolbar and the rest of the page. A large image
    // that merely overflows the visible area (common for inserted photos,
    // which take the full editor width) still gets a box — the border is
    // clamped and off-screen handles are hidden at render. Skipped mid-drag
    // so growing the image past the editor edge doesn't yank the handles
    // out from under the pointer.
    if (editor && !draggingRef.current) {
      const b = editor.getBoundingClientRect();
      const inView = rect.bottom > b.top + 4 && rect.top < b.bottom - 4;
      if (!inView) {
        setOverlayRect(null);
        return;
      }
    }
    setOverlayRect((prev) =>
      prev && prev.x === rect.left && prev.y === rect.top && prev.width === rect.width && prev.height === rect.height
        ? prev
        : { x: rect.left, y: rect.top, width: rect.width, height: rect.height },
    );
  }

  function cancelHide() {
    if (hideTimerRef.current !== null) {
      window.clearTimeout(hideTimerRef.current);
      hideTimerRef.current = null;
    }
  }

  // The selection box is a position:fixed overlay sitting on top of the
  // image with its handles straddling the edges, so the pointer constantly
  // crosses between the image and the handles — and near the editor's edges
  // it briefly leaves the editor box entirely. Hiding on the first
  // mouseleave made it flicker and jump; instead defer the hide and cancel
  // it the moment the pointer lands back on the image or the overlay.
  function scheduleHide() {
    if (draggingRef.current) return;
    cancelHide();
    hideTimerRef.current = window.setTimeout(() => {
      hideTimerRef.current = null;
      if (draggingRef.current || overOverlayRef.current) return;
      activeImageRef.current = null;
      setOverlayRect(null);
    }, 150);
  }

  function handleEditorMouseOver(e: React.MouseEvent<HTMLDivElement>) {
    if (e.target instanceof HTMLImageElement) {
      cancelHide();
      activeImageRef.current = e.target;
      updateOverlay(e.target);
    } else {
      // Pointer moved off the image onto surrounding text — retire the box
      // (deferred, so crossing onto a handle keeps it; see scheduleHide).
      scheduleHide();
    }
  }

  function handleEditorMouseLeave() {
    scheduleHide();
  }

  // Keep the handle glued to the image corner as the page or the editor's
  // own scroll region moves under it — capture:true catches scroll on the
  // inner editor div, which doesn't bubble.
  useEffect(() => {
    if (readOnly) return;
    // rAF-throttle so the fixed overlay is repainted in the same frame the
    // browser scrolls the image, instead of lagging a frame or more behind
    // it (which looked like the box "flying" across the page on fast
    // scrolls).
    let raf = 0;
    const reposition = () => {
      if (raf) return;
      raf = window.requestAnimationFrame(() => {
        raf = 0;
        if (activeImageRef.current) updateOverlay(activeImageRef.current);
      });
    };
    window.addEventListener("scroll", reposition, true);
    window.addEventListener("resize", reposition);
    return () => {
      if (raf) window.cancelAnimationFrame(raf);
      window.removeEventListener("scroll", reposition, true);
      window.removeEventListener("resize", reposition);
      cancelHide();
    };
  }, [readOnly]);

  // Drag-resizes the hovered image from the grabbed handle, width only —
  // height stays "auto" so the image's own aspect ratio is preserved. The
  // width delta is the pointer movement projected onto the handle's outward
  // direction (fx/fy), so a corner tracks the diagonal and an edge tracks
  // its axis, both growing/shrinking the image about its centre.
  function handleResizeStart(e: React.PointerEvent, fx: number, fy: number) {
    e.preventDefault();
    e.stopPropagation();
    const img = activeImageRef.current;
    if (!img) return;

    const startX = e.clientX;
    const startY = e.clientY;
    const startWidth = img.getBoundingClientRect().width;
    const len = Math.hypot(fx, fy) || 1;
    draggingRef.current = true;

    const onMove = (ev: PointerEvent) => {
      const delta = ((ev.clientX - startX) * fx + (ev.clientY - startY) * fy) / len;
      const width = Math.max(MIN_IMAGE_WIDTH, startWidth + delta);
      img.style.width = `${width}px`;
      img.style.height = "auto";
      updateOverlay(img);
    };
    const onUp = () => {
      draggingRef.current = false;
      window.removeEventListener("pointermove", onMove);
      window.removeEventListener("pointerup", onUp);
      handleInput();
      // Pointer may have ended the drag away from both image and handle
      // (e.g. released past the editor edge) with no mouseleave to follow.
      scheduleHide();
    };
    window.addEventListener("pointermove", onMove);
    window.addEventListener("pointerup", onUp);
  }

  // Double-clicking an embedded diagram reopens it in Excalidraw rather
  // than falling through to contentEditable's default double-click (word
  // selection).
  function handleEditorDoubleClick(e: React.MouseEvent<HTMLDivElement>) {
    if (!allowDiagrams) return;
    const target = e.target;
    if (target instanceof HTMLImageElement && target.hasAttribute(DIAGRAM_SCENE_ATTR)) {
      e.preventDefault();
      setDiagramTarget(target);
    }
  }

  function openNewDiagram() {
    const editor = editorRef.current;
    const selection = window.getSelection();
    // Only worth saving if the caret is actually inside this editor —
    // otherwise leave it null and let insertImageDataUrl fall back to
    // wherever focus() + execCommand land by default.
    if (editor && selection && selection.rangeCount > 0) {
      const range = selection.getRangeAt(0);
      if (editor.contains(range.commonAncestorContainer)) {
        savedRangeRef.current = range.cloneRange();
      }
    }
    setDiagramTarget("new");
  }

  function handleDiagramSave(scene: DrawingBoardScene, pngDataUrl: string) {
    if (!editorRef.current) return;
    const encoded = encodeScene(scene);

    if (diagramTarget instanceof HTMLImageElement) {
      diagramTarget.src = pngDataUrl;
      diagramTarget.setAttribute(DIAGRAM_SCENE_ATTR, encoded);
      // Diagrams embedded before figure-wrapping existed are bare <img>s —
      // give them the centered figure + caption on their next edit too.
      wrapImageInFigure(diagramTarget);
    } else {
      // Restore the caret to where it was when the dialog was opened —
      // the dialog's own focus/selection in the meantime otherwise leaves
      // execCommand to insert wherever the selection now collapses to.
      if (savedRangeRef.current) {
        const selection = window.getSelection();
        selection?.removeAllRanges();
        selection?.addRange(savedRangeRef.current);
      }
      const img = insertImageDataUrl(pngDataUrl, editorRef.current);
      if (img) {
        img.setAttribute(DIAGRAM_SCENE_ATTR, encoded);
        img.title = "Double-click to edit diagram";
      }
    }
    savedRangeRef.current = null;
    handleInput();
    setDiagramTarget(null);
  }

  const editingScene =
    diagramTarget instanceof HTMLImageElement
      ? (decodeScene(diagramTarget.getAttribute(DIAGRAM_SCENE_ATTR) ?? "") ?? undefined)
      : undefined;

  return (
    <div className="flex flex-col lg:flex-1 lg:min-h-0">
      {!readOnly && (
        <div className="shrink-0">
          <RichTextToolbar
            editorRef={editorRef}
            onChange={handleInput}
            onInsertDiagram={allowDiagrams ? openNewDiagram : undefined}
          />
        </div>
      )}

      <div
        ref={editorRef}
        contentEditable={!readOnly}
        suppressContentEditableWarning
        onInput={readOnly ? undefined : handleInput}
        onPaste={readOnly ? undefined : handlePaste}
        onDrop={readOnly ? undefined : handleDrop}
        onDragOver={readOnly ? undefined : (e) => e.preventDefault()}
        onMouseOver={readOnly ? undefined : handleEditorMouseOver}
        onMouseLeave={readOnly ? undefined : handleEditorMouseLeave}
        onDoubleClick={readOnly ? undefined : handleEditorDoubleClick}
        className={`lg:min-h-0 lg:flex-1 lg:overflow-y-auto themed-scrollbar rounded-lg border-(--line) border-[0.5px] border-solid bg-(--card) px-4 py-3 text-[length:var(--text-caption)] text-(--fg) outline-none [&_ul]:list-disc [&_ul]:pl-5 [&_ol]:list-decimal [&_ol]:pl-5 [&_h1]:text-xl [&_h1]:font-semibold [&_h2]:text-lg [&_h2]:font-semibold [&_h3]:text-base [&_h3]:font-semibold [&_img]:max-w-full [&_img]:h-auto [&_img]:rounded-lg [&_img]:my-2 [&_img]:mx-auto [&_img]:[-webkit-user-drag:none] [&_img[data-excalidraw-scene]]:cursor-pointer [&_figure]:my-4 [&_figure]:mx-auto [&_figure]:text-center [&_figure_img]:my-0 [&_figcaption]:mt-2 [&_figcaption]:text-center [&_figcaption]:text-[length:var(--text-pill)] [&_figcaption]:text-(--text-muted) [&_figcaption]:italic [&_figcaption]:outline-none ${
          readOnly
            ? "cursor-default [&_figcaption:empty]:hidden"
            : "min-h-[50vh] [&_figcaption:empty]:before:content-['Add_a_caption'] [&_figcaption:empty]:before:text-(--text-faint) [&_figcaption:empty]:before:not-italic"
        } ${className ?? ""}`}
      />

      {!readOnly && overlayRect && (() => {
        // Clamp the visible border to the editor's own box so a tall image
        // that overflows the scroll area doesn't draw its frame over the
        // toolbar / page. Handles stay pinned to the true image corners
        // (so a resize drag still tracks the real edges) and any handle
        // that lands outside the editor viewport is dropped until it
        // scrolls into view.
        const editorBox = editorRef.current?.getBoundingClientRect();
        const clamp = (lo: number, v: number, hi: number) => Math.max(lo, Math.min(v, hi));
        const box = editorBox
          ? {
              left: clamp(editorBox.left, overlayRect.x, editorBox.right),
              top: clamp(editorBox.top, overlayRect.y, editorBox.bottom),
              right: clamp(editorBox.left, overlayRect.x + overlayRect.width, editorBox.right),
              bottom: clamp(editorBox.top, overlayRect.y + overlayRect.height, editorBox.bottom),
            }
          : { left: overlayRect.x, top: overlayRect.y, right: overlayRect.x + overlayRect.width, bottom: overlayRect.y + overlayRect.height };

        return (
          <>
            {/* Click-through frame (pointer-events:none) so hovering and
                caret placement on the image underneath keep working. */}
            <div
              style={{ left: box.left, top: box.top, width: box.right - box.left, height: box.bottom - box.top }}
              className="pointer-events-none fixed z-50 rounded-lg border border-(--fg)/60"
            />
            {RESIZE_HANDLES.map((h) => {
              const hx = overlayRect.x + h.cx * overlayRect.width;
              const hy = overlayRect.y + h.cy * overlayRect.height;
              if (editorBox && (hy < editorBox.top || hy > editorBox.bottom || hx < editorBox.left - 8 || hx > editorBox.right + 8)) {
                return null;
              }
              return (
                <div
                  key={h.key}
                  onPointerDown={(e) => handleResizeStart(e, h.fx, h.fy)}
                  onMouseEnter={() => {
                    overOverlayRef.current = true;
                    cancelHide();
                  }}
                  onMouseLeave={() => {
                    overOverlayRef.current = false;
                    scheduleHide();
                  }}
                  style={{ left: hx, top: hy, cursor: h.cursor }}
                  className="pointer-events-auto fixed z-50 -ml-[5px] -mt-[5px] size-2.5 rounded-[3px] border border-(--fg) bg-(--card) shadow-sm"
                />
              );
            })}
          </>
        );
      })()}

      {diagramTarget && (
        <DiagramDialog initialScene={editingScene} onCancel={() => setDiagramTarget(null)} onSave={handleDiagramSave} />
      )}
    </div>
  );
}
