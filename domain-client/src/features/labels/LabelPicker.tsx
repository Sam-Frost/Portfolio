import { useEffect, useRef, useState } from "react";
import { Tag } from "lucide-react";
import { LABEL_COLOR_VAR } from "./colors";
import type { Label } from "./types";

interface LabelPickerProps {
  labels: Label[];
  selectedId: string | null;
  onChange: (id: string | null) => void;
  // Keeps the trigger visible even with no label selected — used in
  // contexts (edit dialog) that aren't the minimal always-hover-to-reveal
  // add-todo form.
  alwaysVisible?: boolean;
}

export function LabelPicker({ labels, selectedId, onChange, alwaysVisible = false }: LabelPickerProps) {
  const [open, setOpen] = useState(false);
  const ref = useRef<HTMLDivElement>(null);
  const selected = labels.find((l) => l.id === selectedId) ?? null;

  useEffect(() => {
    if (!open) return;

    function handlePointerDown(e: MouseEvent) {
      if (ref.current && !ref.current.contains(e.target as Node)) {
        setOpen(false);
      }
    }

    document.addEventListener("mousedown", handlePointerDown);
    return () => document.removeEventListener("mousedown", handlePointerDown);
  }, [open]);

  return (
    <div className="relative" ref={ref}>
      <button
        type="button"
        onClick={() => setOpen((v) => !v)}
        aria-label="Set label"
        className={`flex items-center justify-center gap-1.5 h-6 rounded-md px-1.5 text-[length:var(--text-pill)] text-(--text-faint) transition-opacity cursor-pointer ${
          selected || alwaysVisible ? "opacity-100" : "opacity-0 group-focus-within:opacity-60 hover:opacity-100"
        }`}
      >
        {selected ? (
          <>
            <span
              className="size-2 rounded-full shrink-0"
              style={{ backgroundColor: LABEL_COLOR_VAR[selected.color] }}
            />
            <span className="text-(--text-muted)">{selected.name}</span>
          </>
        ) : (
          <Tag size={13} />
        )}
      </button>

      {open && (
        <div className="absolute left-0 top-full z-10 mt-1 w-40 max-h-56 overflow-y-auto rounded-lg border-(--line) border-[0.5px] border-solid bg-(--card) p-1 shadow-lg themed-scrollbar">
          <button
            type="button"
            onClick={() => {
              onChange(null);
              setOpen(false);
            }}
            className={`flex w-full items-center rounded-md px-2 py-1.5 text-left text-[length:var(--text-pill)] transition-colors cursor-pointer ${
              !selectedId ? "bg-(--card-alt) text-(--fg)" : "text-(--text-muted) hover:text-(--fg) hover:bg-(--card-alt)"
            }`}
          >
            No label
          </button>

          {labels.map((label) => (
            <button
              key={label.id}
              type="button"
              onClick={() => {
                onChange(label.id);
                setOpen(false);
              }}
              className={`flex w-full items-center gap-1.5 rounded-md px-2 py-1.5 text-left text-[length:var(--text-pill)] transition-colors cursor-pointer ${
                selectedId === label.id
                  ? "bg-(--card-alt) text-(--fg)"
                  : "text-(--text-muted) hover:text-(--fg) hover:bg-(--card-alt)"
              }`}
            >
              <span
                className="size-2 rounded-full shrink-0"
                style={{ backgroundColor: LABEL_COLOR_VAR[label.color] }}
              />
              <span className="truncate">{label.name}</span>
            </button>
          ))}

          {labels.length === 0 && (
            <p className="px-2 py-1.5 text-[length:var(--text-pill)] text-(--text-faint)">No labels yet.</p>
          )}
        </div>
      )}
    </div>
  );
}
