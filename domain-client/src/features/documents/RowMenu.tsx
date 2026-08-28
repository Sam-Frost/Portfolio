import { useEffect, useRef, useState } from "react";
import { MoreHorizontal } from "lucide-react";

export interface RowMenuAction {
  label: string;
  icon: React.ReactNode;
  onSelect: () => void;
  destructive?: boolean;
}

// The "⋯" overflow menu shared by folder and document rows.
export function RowMenu({ actions }: { actions: RowMenuAction[] }) {
  const [open, setOpen] = useState(false);
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!open) return;
    function onDown(e: MouseEvent) {
      if (ref.current && !ref.current.contains(e.target as Node)) setOpen(false);
    }
    document.addEventListener("mousedown", onDown);
    return () => document.removeEventListener("mousedown", onDown);
  }, [open]);

  return (
    <div className="relative" ref={ref}>
      <button
        onClick={() => setOpen((v) => !v)}
        aria-label="More actions"
        className={`flex items-center justify-center size-7 rounded-lg transition-colors cursor-pointer ${
          open ? "bg-(--card-alt) text-(--fg)" : "text-(--text-faint) hover:text-(--fg) hover:bg-(--card-alt)"
        }`}
      >
        <MoreHorizontal size={15} />
      </button>

      {open && (
        <div className="absolute right-0 z-20 mt-1 w-40 rounded-lg border-(--line) border-[0.5px] border-solid bg-(--card) p-1 shadow-lg">
          {actions.map((action) => (
            <button
              key={action.label}
              onClick={() => {
                setOpen(false);
                action.onSelect();
              }}
              className={`flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-left text-[length:var(--text-pill)] transition-colors cursor-pointer ${
                action.destructive
                  ? "text-red-400 hover:bg-(--card-alt)"
                  : "text-(--text-muted) hover:text-(--fg) hover:bg-(--card-alt)"
              }`}
            >
              {action.icon}
              {action.label}
            </button>
          ))}
        </div>
      )}
    </div>
  );
}
