import { ArrowDown, ArrowUp, Pencil, Trash2 } from "lucide-react";
import { VisibilityToggle } from "./components";

interface ItemRowProps {
  title: string;
  subtitle?: string;
  visible: boolean;
  isFirst: boolean;
  isLast: boolean;
  onMoveUp: () => void;
  onMoveDown: () => void;
  onToggleVisible: (v: boolean) => void;
  onEdit: () => void;
  onDelete: () => void;
}

// One row in a CMS list (projects / experience / writings). Reorder is
// two adjacent PATCHes driven by the up/down arrows (see swapOrder in the
// tabs); drag-and-drop can come later.
export function ItemRow({
  title,
  subtitle,
  visible,
  isFirst,
  isLast,
  onMoveUp,
  onMoveDown,
  onToggleVisible,
  onEdit,
  onDelete,
}: ItemRowProps) {
  return (
    <div
      className={`flex items-center gap-3 rounded-xl border-(--line) border-[0.5px] border-solid bg-(--card) px-3 py-2.5 ${
        visible ? "" : "opacity-60"
      }`}
    >
      <div className="flex flex-col">
        <button
          type="button"
          onClick={onMoveUp}
          disabled={isFirst}
          aria-label="Move up"
          className="text-(--text-faint) hover:text-(--fg) disabled:opacity-25 disabled:hover:text-(--text-faint) cursor-pointer disabled:cursor-not-allowed"
        >
          <ArrowUp size={12} />
        </button>
        <button
          type="button"
          onClick={onMoveDown}
          disabled={isLast}
          aria-label="Move down"
          className="text-(--text-faint) hover:text-(--fg) disabled:opacity-25 disabled:hover:text-(--text-faint) cursor-pointer disabled:cursor-not-allowed"
        >
          <ArrowDown size={12} />
        </button>
      </div>

      <div className="min-w-0 flex-1">
        <div className="truncate text-[length:var(--text-caption)] text-(--fg)">{title}</div>
        {subtitle && <div className="truncate text-[length:var(--text-pill)] text-(--text-muted)">{subtitle}</div>}
      </div>

      <VisibilityToggle visible={visible} onChange={onToggleVisible} />

      <button
        type="button"
        onClick={onEdit}
        aria-label="Edit"
        className="text-(--text-faint) hover:text-(--fg) cursor-pointer"
      >
        <Pencil size={13} />
      </button>
      <button
        type="button"
        onClick={onDelete}
        aria-label="Delete"
        className="text-(--text-faint) hover:text-red-400 cursor-pointer"
      >
        <Trash2 size={13} />
      </button>
    </div>
  );
}
