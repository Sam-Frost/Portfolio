import { useState, type KeyboardEvent } from "react";
import { ChevronRight, Folder as FolderIcon, Home } from "lucide-react";
import type { Folder } from "./types";

interface MoveDialogProps {
  folders: Folder[];
  // Folder ids the target may not be (the item itself + its subtree, when
  // moving a folder). Empty when moving a document.
  disabledIds?: Set<string>;
  currentParentId: string | null;
  title?: string;
  onCancel: () => void;
  onConfirm: (destinationId: string | null) => Promise<void> | void;
}

interface TreeNode extends Folder {
  depth: number;
}

function flattenTree(folders: Folder[]): TreeNode[] {
  const byParent = new Map<string | null, Folder[]>();
  for (const f of folders) {
    const key = f.parentId;
    byParent.set(key, [...(byParent.get(key) ?? []), f]);
  }
  const out: TreeNode[] = [];
  const walk = (parentId: string | null, depth: number) => {
    for (const f of byParent.get(parentId) ?? []) {
      out.push({ ...f, depth });
      walk(f.id, depth + 1);
    }
  };
  walk(null, 0);
  return out;
}

export function MoveDialog({
  folders,
  disabledIds,
  currentParentId,
  title = "Move to",
  onCancel,
  onConfirm,
}: MoveDialogProps) {
  const nodes = flattenTree(folders);
  const [selected, setSelected] = useState<string | null>(currentParentId);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function confirm() {
    if (busy) return;
    setBusy(true);
    setError(null);
    try {
      await onConfirm(selected);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Couldn't move.");
      setBusy(false);
    }
  }

  function handleKeyDown(e: KeyboardEvent<HTMLDivElement>) {
    if (e.key === "Escape") onCancel();
  }

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 px-4"
      onClick={onCancel}
      onKeyDown={handleKeyDown}
      role="presentation"
    >
      <div
        className="w-full max-w-sm rounded-xl border-(--line) border-[0.5px] border-solid bg-(--card) p-4 shadow-lg"
        onClick={(e) => e.stopPropagation()}
        role="dialog"
        aria-modal="true"
        aria-label={title}
      >
        <h2 className="text-[length:var(--text-caption)] text-(--fg) mb-3">{title}</h2>

        <div className="max-h-64 overflow-y-auto themed-scrollbar rounded-lg border-(--line) border-[0.5px] border-solid p-1">
          <button
            type="button"
            onClick={() => setSelected(null)}
            className={`flex w-full items-center gap-1.5 rounded-md px-2 py-1.5 text-left text-[length:var(--text-pill)] transition-colors cursor-pointer ${
              selected === null ? "bg-(--card-alt) text-(--fg)" : "text-(--text-muted) hover:text-(--fg) hover:bg-(--card-alt)"
            }`}
          >
            <Home size={13} className="shrink-0" />
            Root
          </button>

          {nodes.map((node) => {
            const disabled = disabledIds?.has(node.id) ?? false;
            return (
              <button
                key={node.id}
                type="button"
                disabled={disabled}
                onClick={() => setSelected(node.id)}
                style={{ paddingLeft: `${node.depth * 14 + 8}px` }}
                className={`flex w-full items-center gap-1.5 rounded-md py-1.5 pr-2 text-left text-[length:var(--text-pill)] transition-colors ${
                  disabled
                    ? "text-(--text-faint) cursor-not-allowed"
                    : selected === node.id
                      ? "bg-(--card-alt) text-(--fg) cursor-pointer"
                      : "text-(--text-muted) hover:text-(--fg) hover:bg-(--card-alt) cursor-pointer"
                }`}
              >
                {node.depth > 0 && <ChevronRight size={11} className="shrink-0 opacity-40" />}
                <FolderIcon size={13} className="shrink-0" />
                <span className="truncate">{node.name}</span>
              </button>
            );
          })}
        </div>

        {error && <p className="mt-2 text-[length:var(--text-pill)] text-red-400">{error}</p>}

        <div className="mt-4 flex items-center justify-end gap-1.5">
          <button
            type="button"
            onClick={onCancel}
            className="rounded-md px-2.5 py-1 text-[length:var(--text-pill)] text-(--text-muted) hover:text-(--fg) hover:bg-(--card-alt) transition-colors cursor-pointer"
          >
            Cancel
          </button>
          <button
            type="button"
            disabled={busy}
            onClick={confirm}
            className="rounded-md bg-(--fg) text-(--bg) px-2.5 py-1 text-[length:var(--text-pill)] cursor-pointer disabled:opacity-50"
          >
            Move here
          </button>
        </div>
      </div>
    </div>
  );
}
