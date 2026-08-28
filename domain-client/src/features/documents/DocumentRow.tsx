import { Download, FileText, FolderInput, Pencil, Trash2 } from "lucide-react";
import { LabelPicker } from "../labels/LabelPicker";
import { formatBytes } from "./formatBytes";
import { RowMenu } from "./RowMenu";
import type { DocumentItem, DocumentLabel } from "./types";

interface DocumentRowProps {
  document: DocumentItem;
  labels: DocumentLabel[];
  showFolderHint: boolean;
  folderName: string | null;
  onDownload: () => void;
  onRename: () => void;
  onMove: () => void;
  onDelete: () => void;
  onLabelChange: (labelId: string | null) => void;
}

function shortDate(iso: string): string {
  return new Date(iso).toLocaleDateString(undefined, { month: "short", day: "numeric", year: "numeric" });
}

export function DocumentRow({
  document,
  labels,
  showFolderHint,
  folderName,
  onDownload,
  onRename,
  onMove,
  onDelete,
  onLabelChange,
}: DocumentRowProps) {
  return (
    <div className="group flex items-center gap-2.5 rounded-lg border-(--line) border-[0.5px] border-solid bg-(--card) px-3 py-2.5">
      <FileText size={16} className="shrink-0 text-(--text-muted)" />

      <button onClick={onDownload} className="flex-1 min-w-0 text-left cursor-pointer" title="Download">
        <span className="block truncate text-[length:var(--text-caption)] text-(--fg)">{document.name}</span>
        <span className="block text-[length:var(--text-pill)] text-(--text-faint)">
          {formatBytes(document.sizeBytes)} · {shortDate(document.uploadedAt ?? document.createdAt)}
          {showFolderHint && folderName ? ` · ${folderName}` : ""}
        </span>
      </button>

      <div onClick={(e) => e.stopPropagation()} role="presentation">
        <LabelPicker labels={labels} selectedId={document.labelId} onChange={onLabelChange} />
      </div>

      <div className="opacity-0 group-hover:opacity-100 transition-opacity">
        <RowMenu
          actions={[
            { label: "Download", icon: <Download size={13} />, onSelect: onDownload },
            { label: "Rename", icon: <Pencil size={13} />, onSelect: onRename },
            { label: "Move", icon: <FolderInput size={13} />, onSelect: onMove },
            { label: "Delete", icon: <Trash2 size={13} />, onSelect: onDelete, destructive: true },
          ]}
        />
      </div>
    </div>
  );
}
