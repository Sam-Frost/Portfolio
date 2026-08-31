import { FolderClosed, FolderInput, Pencil, Trash2 } from "lucide-react";
import { RowMenu } from "./RowMenu";
import type { Folder } from "./types";

interface FolderRowProps {
  folder: Folder;
  onOpen: () => void;
  onRename: () => void;
  onMove: () => void;
  onDelete: () => void;
}

export function FolderRow({ folder, onOpen, onRename, onMove, onDelete }: FolderRowProps) {
  return (
    <div className="group flex items-center gap-2.5 rounded-lg border-(--line) border-[0.5px] border-solid bg-(--card) px-3 py-2.5">
      <button
        onClick={onOpen}
        className="flex flex-1 min-w-0 items-center gap-2.5 text-left cursor-pointer"
      >
        <FolderClosed size={16} className="shrink-0 text-(--text-muted)" />
        <span className="truncate text-[length:var(--text-caption)] text-(--fg)">{folder.name}</span>
      </button>

      <div className="opacity-100 lg:opacity-0 lg:group-hover:opacity-100 transition-opacity">
        <RowMenu
          actions={[
            { label: "Rename", icon: <Pencil size={13} />, onSelect: onRename },
            { label: "Move", icon: <FolderInput size={13} />, onSelect: onMove },
            { label: "Delete", icon: <Trash2 size={13} />, onSelect: onDelete, destructive: true },
          ]}
        />
      </div>
    </div>
  );
}
