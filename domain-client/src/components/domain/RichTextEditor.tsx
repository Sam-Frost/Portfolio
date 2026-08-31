import { useEffect, useRef } from "react";
import { RichTextToolbar } from "./RichTextToolbar";

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
}

// The rich-text surface shared by every domain feature that edits HTML
// content (notepad, diary, ...): a document.execCommand-driven toolbar
// (RichTextToolbar) plus a contentEditable div. Originally lived inline in
// notepad's NoteEditorPage; extracted here so diary can reuse the same
// editor rather than reimplementing one.
export function RichTextEditor({ initialContentHtml, onChange, readOnly = false, className }: RichTextEditorProps) {
  const editorRef = useRef<HTMLDivElement | null>(null);

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

  return (
    <div className="flex flex-col lg:flex-1 lg:min-h-0">
      {!readOnly && (
        <div className="shrink-0">
          <RichTextToolbar editorRef={editorRef} onChange={handleInput} />
        </div>
      )}

      <div
        ref={editorRef}
        contentEditable={!readOnly}
        suppressContentEditableWarning
        onInput={readOnly ? undefined : handleInput}
        className={`lg:min-h-0 lg:flex-1 lg:overflow-y-auto themed-scrollbar rounded-lg border-(--line) border-[0.5px] border-solid bg-(--card) px-4 py-3 text-[length:var(--text-caption)] text-(--fg) outline-none [&_ul]:list-disc [&_ul]:pl-5 [&_ol]:list-decimal [&_ol]:pl-5 [&_h1]:text-xl [&_h1]:font-semibold [&_h2]:text-lg [&_h2]:font-semibold [&_h3]:text-base [&_h3]:font-semibold ${
          readOnly ? "cursor-default" : "min-h-[50vh]"
        } ${className ?? ""}`}
      />
    </div>
  );
}
