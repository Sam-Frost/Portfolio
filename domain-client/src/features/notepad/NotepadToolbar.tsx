import type { RefObject } from "react";
import { Bold, Italic, Underline, List, ListOrdered } from "lucide-react";

interface NotepadToolbarProps {
  editorRef: RefObject<HTMLDivElement | null>;
  onChange: () => void;
}

const HEADING_OPTIONS = [
  { value: "P", label: "Paragraph" },
  { value: "H1", label: "Heading 1" },
  { value: "H2", label: "Heading 2" },
  { value: "H3", label: "Heading 3" },
];

// document.execCommand's legacy fontSize scale (1-7); mapped to a handful of
// readable labels instead of exposing all seven steps.
const FONT_SIZE_OPTIONS = [
  { value: "2", label: "Small" },
  { value: "3", label: "Normal" },
  { value: "5", label: "Large" },
  { value: "7", label: "Huge" },
];

export function NotepadToolbar({ editorRef, onChange }: NotepadToolbarProps) {
  function focusEditor() {
    editorRef.current?.focus();
  }

  function runCommand(command: string, value?: string) {
    focusEditor();
    document.execCommand(command, false, value);
    onChange();
  }

  function iconButton(command: string, label: string, Icon: typeof Bold) {
    return (
      <button
        type="button"
        aria-label={label}
        title={label}
        // mousedown (not click) + preventDefault keeps the editor's text
        // selection intact — a click would first steal focus and collapse it.
        onMouseDown={(e) => {
          e.preventDefault();
          runCommand(command);
        }}
        className="flex items-center justify-center size-7 rounded-lg text-(--text-muted) hover:text-(--fg) hover:bg-(--card-alt) transition-colors cursor-pointer"
      >
        <Icon size={15} />
      </button>
    );
  }

  return (
    <div className="flex items-center gap-1 flex-wrap rounded-lg border-(--line) border-[0.5px] border-solid bg-(--card) px-1.5 py-1 mb-3">
      <select
        aria-label="Text style"
        onMouseDown={focusEditor}
        onChange={(e) => runCommand("formatBlock", e.target.value)}
        defaultValue="P"
        className="h-7 rounded-md bg-(--card-alt) px-2 text-[length:var(--text-pill)] text-(--text-muted) hover:text-(--fg) transition-colors cursor-pointer outline-none"
      >
        {HEADING_OPTIONS.map((opt) => (
          <option key={opt.value} value={opt.value}>
            {opt.label}
          </option>
        ))}
      </select>

      <select
        aria-label="Font size"
        onMouseDown={focusEditor}
        onChange={(e) => runCommand("fontSize", e.target.value)}
        defaultValue="3"
        className="h-7 rounded-md bg-(--card-alt) px-2 text-[length:var(--text-pill)] text-(--text-muted) hover:text-(--fg) transition-colors cursor-pointer outline-none"
      >
        {FONT_SIZE_OPTIONS.map((opt) => (
          <option key={opt.value} value={opt.value}>
            {opt.label}
          </option>
        ))}
      </select>

      <span className="w-px h-5 bg-(--line) mx-0.5" />

      {iconButton("bold", "Bold", Bold)}
      {iconButton("italic", "Italic", Italic)}
      {iconButton("underline", "Underline", Underline)}

      <span className="w-px h-5 bg-(--line) mx-0.5" />

      {iconButton("insertUnorderedList", "Bullet list", List)}
      {iconButton("insertOrderedList", "Numbered list", ListOrdered)}
    </div>
  );
}
