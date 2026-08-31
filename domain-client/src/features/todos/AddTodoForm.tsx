import { useRef, useState, type KeyboardEvent } from "react";
import { CalendarPlus } from "lucide-react";
import { LabelPicker } from "../labels/LabelPicker";
import type { Label } from "../labels/types";

interface AddTodoFormProps {
  labels: Label[];
  // Controlled by TodosPage so it can track the active label filter — picking
  // a label to view todos by also preselects it here. Deliberately not reset
  // after submit, so adding several todos with the same label in a row doesn't
  // require reselecting it each time.
  labelId: string | null;
  onLabelChange: (id: string | null) => void;
  onAdd: (input: { name: string; description: string | null; targetDate: string | null; labelId: string | null }) => void;
}

export function AddTodoForm({ labels, labelId, onLabelChange, onAdd }: AddTodoFormProps) {
  const [text, setText] = useState("");
  const [targetDate, setTargetDate] = useState("");
  const [showDate, setShowDate] = useState(false);
  const [confirming, setConfirming] = useState(false);
  const textareaRef = useRef<HTMLTextAreaElement>(null);
  const yesButtonRef = useRef<HTMLButtonElement>(null);

  function resize() {
    const el = textareaRef.current;
    if (!el) return;
    el.style.height = "auto";
    el.style.height = `${el.scrollHeight}px`;
  }

  function submit() {
    const trimmed = text.trim();
    if (!trimmed) return;

    const [firstLine, ...rest] = trimmed.split("\n");
    const description = rest.join("\n").trim();

    onAdd({
      name: firstLine.trim(),
      description: description || null,
      targetDate: targetDate || null,
      labelId,
    });

    setText("");
    setTargetDate("");
    setShowDate(false);
    setConfirming(false);
    requestAnimationFrame(resize);
  }

  function cancelConfirm() {
    setConfirming(false);
    textareaRef.current?.focus();
  }

  function handleKeyDown(e: KeyboardEvent<HTMLTextAreaElement>) {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      if (!text.trim()) return;

      // Multi-line entries are easy to submit by accident with a stray
      // Enter while typing a description — confirm first. The Yes button
      // gets focus, so a second Enter (native button behavior) submits it.
      if (text.includes("\n")) {
        setConfirming(true);
        requestAnimationFrame(() => yesButtonRef.current?.focus());
      } else {
        submit();
      }
    } else if (e.key === "Escape" && confirming) {
      cancelConfirm();
    }
  }

  return (
    <div className="group bg-(--card) border-(--line) border-[0.5px] border-solid rounded-xl px-4 py-3 mb-4">
      <textarea
        ref={textareaRef}
        value={text}
        onChange={(e) => {
          setText(e.target.value);
          resize();
        }}
        onFocus={() => setConfirming(false)}
        onKeyDown={handleKeyDown}
        placeholder="Add a todo..."
        rows={1}
        className="w-full resize-none bg-transparent text-[length:var(--text-caption)] text-(--fg) placeholder:text-(--text-faint) focus:outline-none"
      />

      {confirming && (
        <div className="flex items-center justify-between gap-2 mt-2 rounded-lg bg-(--card-alt) border-(--line) border-[0.5px] border-solid px-3 py-2">
          <span className="text-[length:var(--text-pill)] text-(--text-muted)">Add this todo?</span>
          <div className="flex items-center gap-1.5">
            <button
              type="button"
              onClick={cancelConfirm}
              className="rounded-md px-2.5 py-1 text-[length:var(--text-pill)] text-(--text-muted) hover:text-(--fg) hover:bg-(--card) transition-colors cursor-pointer"
            >
              No
            </button>
            <button
              ref={yesButtonRef}
              type="button"
              onClick={submit}
              className="rounded-md bg-(--fg) text-(--bg) px-2.5 py-1 text-[length:var(--text-pill)] cursor-pointer"
            >
              Yes
            </button>
          </div>
        </div>
      )}

      <div className="flex items-center gap-1 mt-1">
        <LabelPicker labels={labels} selectedId={labelId} onChange={onLabelChange} />

        <button
          type="button"
          onClick={() => setShowDate((v) => !v)}
          aria-label="Set target date"
          className="flex items-center justify-center size-6 rounded-md text-(--text-faint) opacity-0 group-focus-within:opacity-60 hover:opacity-100 transition-opacity cursor-pointer"
        >
          <CalendarPlus size={13} />
        </button>

        {showDate && (
          <input
            type="date"
            value={targetDate}
            onChange={(e) => setTargetDate(e.target.value)}
            className="bg-(--card-alt) border-(--line) border-[0.5px] border-solid rounded-lg px-2 py-1 text-[length:var(--text-pill)] text-(--fg) focus:outline-none"
          />
        )}
      </div>
    </div>
  );
}
