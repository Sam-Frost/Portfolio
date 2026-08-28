import { ChevronRight, House } from "lucide-react";
import type { Folder } from "./types";

interface BreadcrumbProps {
  folders: Folder[];
  currentFolderId: string | null;
  onNavigate: (folderId: string | null) => void;
}

// Walks parentId links from the current folder up to the root.
function trail(folders: Folder[], currentId: string | null): Folder[] {
  const byId = new Map(folders.map((f) => [f.id, f]));
  const out: Folder[] = [];
  let id = currentId;
  while (id) {
    const f = byId.get(id);
    if (!f) break;
    out.unshift(f);
    id = f.parentId;
  }
  return out;
}

export function Breadcrumb({ folders, currentFolderId, onNavigate }: BreadcrumbProps) {
  const path = trail(folders, currentFolderId);

  return (
    <nav className="flex items-center gap-1 text-[length:var(--text-caption)] min-w-0">
      <button
        onClick={() => onNavigate(null)}
        className={`flex items-center gap-1 rounded-md px-1.5 py-0.5 transition-colors cursor-pointer ${
          currentFolderId === null ? "text-(--fg)" : "text-(--text-muted) hover:text-(--fg)"
        }`}
      >
        <House size={13} />
        Documents
      </button>

      {path.map((folder, i) => (
        <div key={folder.id} className="flex items-center gap-1 min-w-0">
          <ChevronRight size={12} className="shrink-0 text-(--text-faint)" />
          <button
            onClick={() => onNavigate(folder.id)}
            className={`truncate rounded-md px-1.5 py-0.5 transition-colors cursor-pointer ${
              i === path.length - 1 ? "text-(--fg)" : "text-(--text-muted) hover:text-(--fg)"
            }`}
          >
            {folder.name}
          </button>
        </div>
      ))}
    </nav>
  );
}
