export const MAX_IMAGE_BYTES = 5 * 1024 * 1024; // 5MB

// Reads an image file as a data URL and inserts it into editor at the
// current cursor position via execCommand, so a picked/pasted/dropped image
// becomes part of the note's contentHtml like any other rich text. No
// upload endpoint and no expiring URLs to keep valid on every future load —
// the image travels with the HTML itself.
export function insertImageFile(file: File, editor: HTMLElement, onChange: () => void) {
  if (!file.type.startsWith("image/")) return;
  if (file.size > MAX_IMAGE_BYTES) {
    window.alert("That image is too large (max 5MB).");
    return;
  }

  const reader = new FileReader();
  reader.onload = () => {
    const dataUrl = reader.result;
    if (typeof dataUrl !== "string") return;
    insertImageDataUrl(dataUrl, editor);
    onChange();
  };
  reader.readAsDataURL(file);
}

// Inserts an <img src=dataUrl> at the cursor and returns the inserted
// element, without notifying the caller's onChange — used directly by
// insertImageFile above, and by the diagram dialog (DiagramDialog.tsx),
// which needs to tag the freshly-inserted image with its Excalidraw scene
// data before the caller persists the note.
export function insertImageDataUrl(dataUrl: string, editor: HTMLElement): HTMLImageElement | null {
  editor.focus();
  document.execCommand("insertImage", false, dataUrl);
  // Images are draggable by default, which lets a stray drag on the image
  // body (rather than the resize handle) yank it out of the note via the
  // browser's native drag-to-move instead of resizing it. Every image this
  // module has already touched carries draggable="false", so the one image
  // still missing it is the one execCommand just inserted.
  const inserted = editor.querySelector<HTMLImageElement>('img:not([draggable="false"])');
  if (inserted) {
    inserted.draggable = false;
    const figure = wrapImageInFigure(inserted);
    const caption = figure.querySelector<HTMLElement>("figcaption");
    if (caption) placeCaretIn(caption);
  }
  return inserted;
}

// Wraps a freshly-inserted <img> in a <figure> that centers the image (see
// the [&_figure]/[&_figcaption] rules in RichTextEditor) and carries an
// initially-empty <figcaption> beneath it, so every embedded image/diagram
// gets a centered layout and an optional caption without the user building
// that structure by hand. Idempotent — an <img> that is already the direct
// child of a <figure> is returned untouched.
export function wrapImageInFigure(img: HTMLImageElement): HTMLElement {
  if (img.parentElement?.tagName === "FIGURE") return img.parentElement;

  const doc = img.ownerDocument;
  const figure = doc.createElement("figure");
  // Left empty: RichTextEditor shows a CSS placeholder on an empty
  // figcaption while editing and hides it entirely in read-only views, so
  // an untouched caption costs nothing and adds no stored markup beyond the
  // bare tag.
  const caption = doc.createElement("figcaption");

  // If execCommand dropped the new image inside an existing figure's
  // caption (the caret was parked there after a previous insert), lift the
  // new figure out to sit after that one rather than nesting inside it.
  const enclosingFigure = img.closest("figure");
  if (enclosingFigure) {
    img.remove();
    enclosingFigure.after(figure);
  } else {
    img.replaceWith(figure);
  }
  figure.append(img, caption);
  return figure;
}

// Collapses the caret to the start of el and points the document selection
// at it, so right after inserting an image the user can type its caption
// straight away.
export function placeCaretIn(el: HTMLElement) {
  const selection = el.ownerDocument.getSelection();
  if (!selection) return;
  const range = el.ownerDocument.createRange();
  range.selectNodeContents(el);
  range.collapse(true);
  selection.removeAllRanges();
  selection.addRange(range);
}
