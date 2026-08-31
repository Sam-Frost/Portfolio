import { useEffect, useRef, useState } from "react";
import { FolderPlus, Search, Tag, TagX, Upload, X } from "lucide-react";
import { Toast } from "../../components/Toast";
import { LABEL_COLOR_VAR } from "../labels/colors";
import {
  createFolder,
  deleteDocument,
  deleteFolder,
  fetchDocuments,
  fetchFolders,
  getDownloadUrl,
  updateDocument,
  updateFolder,
} from "./api";
import { fetchDocumentLabels } from "./labelApi";
import { Breadcrumb } from "./Breadcrumb";
import { ConfirmDialog } from "../../components/domain/ConfirmDialog";
import { DocumentRow } from "./DocumentRow";
import { FolderRow } from "./FolderRow";
import { MoveDialog } from "./MoveDialog";
import { PromptDialog } from "./PromptDialog";
import { UploadProgressList } from "./UploadProgressList";
import { useUploads } from "./useUploads";
import type { DocumentItem, DocumentLabel, Folder } from "./types";

// Sentinel labelFilter value for "documents with no label", distinct from
// null ("All labels"). Real ids are 32-char hex, so this can't collide.
const NO_LABEL_FILTER = "none";

type Dialog =
  | { type: "newFolder" }
  | { type: "renameFolder"; folder: Folder }
  | { type: "renameDoc"; doc: DocumentItem }
  | { type: "moveFolder"; folder: Folder }
  | { type: "moveDoc"; doc: DocumentItem }
  | { type: "deleteFolder"; folder: Folder }
  | { type: "deleteDoc"; doc: DocumentItem }
  | null;

function subtreeIds(folders: Folder[], rootId: string): Set<string> {
  const out = new Set<string>([rootId]);
  let grew = true;
  while (grew) {
    grew = false;
    for (const f of folders) {
      if (f.parentId && out.has(f.parentId) && !out.has(f.id)) {
        out.add(f.id);
        grew = true;
      }
    }
  }
  return out;
}

export function DocumentsPage() {
  const [folders, setFolders] = useState<Folder[]>([]);
  const [labels, setLabels] = useState<DocumentLabel[]>([]);
  const [documents, setDocuments] = useState<DocumentItem[]>([]);
  const [currentFolderId, setCurrentFolderId] = useState<string | null>(null);
  const [labelFilter, setLabelFilter] = useState<string | null>(null);
  const [query, setQuery] = useState("");
  const [activeQuery, setActiveQuery] = useState("");
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [dialog, setDialog] = useState<Dialog>(null);
  const [toast, setToast] = useState<{ key: number; message: string } | null>(null);
  const [showLabelMenu, setShowLabelMenu] = useState(false);
  const labelMenuRef = useRef<HTMLDivElement>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);
  const [dragging, setDragging] = useState(false);

  function notify(message: string) {
    setToast((prev) => ({ key: (prev?.key ?? 0) + 1, message }));
  }

  const { uploads, startUploads, clearFinished } = useUploads((doc) => {
    notify(`Uploaded ${doc.name}`);
    setDocuments((prev) => {
      // Only slot it into the current view; otherwise the next fetch picks it up.
      if (!activeQuery && doc.folderId === currentFolderId && !prev.some((d) => d.id === doc.id)) {
        return [doc, ...prev];
      }
      return prev;
    });
  });

  useEffect(() => {
    fetchFolders()
      .then(setFolders)
      .catch((err) => setError(err instanceof Error ? err.message : "Couldn't load folders."));
    fetchDocumentLabels()
      .then(setLabels)
      .catch((err) => setError(err instanceof Error ? err.message : "Couldn't load labels."));
  }, []);

  // Debounce the search box.
  useEffect(() => {
    const t = setTimeout(() => setActiveQuery(query.trim()), 250);
    return () => clearTimeout(t);
  }, [query]);

  useEffect(() => {
    setLoading(true);
    // A search or a label filter spans every folder; plain browsing shows
    // the current folder's contents.
    const spanning = labelFilter !== null || activeQuery !== "";
    const realLabelId = labelFilter && labelFilter !== NO_LABEL_FILTER ? labelFilter : null;
    fetchDocuments({ folderId: spanning ? null : currentFolderId, labelId: realLabelId, q: activeQuery })
      .then(setDocuments)
      .catch((err) => setError(err instanceof Error ? err.message : "Couldn't load documents."))
      .finally(() => setLoading(false));
  }, [currentFolderId, labelFilter, activeQuery]);

  useEffect(() => {
    if (!showLabelMenu) return;
    function onDown(e: MouseEvent) {
      if (labelMenuRef.current && !labelMenuRef.current.contains(e.target as Node)) setShowLabelMenu(false);
    }
    document.addEventListener("mousedown", onDown);
    return () => document.removeEventListener("mousedown", onDown);
  }, [showLabelMenu]);

  const searching = activeQuery !== "";
  // A search or a label filter shows a flat, folder-spanning result list.
  const spanning = searching || labelFilter !== null;
  const childFolders = spanning ? [] : folders.filter((f) => f.parentId === currentFolderId);
  const visibleDocuments =
    labelFilter === NO_LABEL_FILTER ? documents.filter((d) => d.labelId === null) : documents;
  const activeLabel = labelFilter && labelFilter !== NO_LABEL_FILTER ? labels.find((l) => l.id === labelFilter) : null;
  const folderNameById = new Map(folders.map((f) => [f.id, f.name]));

  async function handleDownload(doc: DocumentItem) {
    try {
      const { url } = await getDownloadUrl(doc.id);
      window.location.href = url;
    } catch (err) {
      setError(err instanceof Error ? err.message : "Couldn't start the download.");
    }
  }

  function handleLabelChange(doc: DocumentItem, labelId: string | null) {
    const original = doc;
    setDocuments((prev) => prev.map((d) => (d.id === doc.id ? { ...d, labelId } : d)));
    updateDocument(doc.id, { labelId }).catch((err) => {
      setDocuments((prev) => prev.map((d) => (d.id === doc.id ? original : d)));
      setError(err instanceof Error ? err.message : "Couldn't update the label.");
    });
  }

  function reloadDocuments() {
    const realLabelId = labelFilter && labelFilter !== NO_LABEL_FILTER ? labelFilter : null;
    fetchDocuments({ folderId: spanning ? null : currentFolderId, labelId: realLabelId, q: activeQuery })
      .then(setDocuments)
      .catch(() => {});
  }

  return (
    <div
      className="h-full flex flex-col"
      onDragOver={(e) => {
        e.preventDefault();
        setDragging(true);
      }}
      onDragLeave={(e) => {
        if (e.currentTarget === e.target) setDragging(false);
      }}
      onDrop={(e) => {
        e.preventDefault();
        setDragging(false);
        if (e.dataTransfer.files.length > 0) startUploads(e.dataTransfer.files, currentFolderId);
      }}
    >
      {toast && <Toast key={toast.key} message={toast.message} onDone={() => setToast(null)} />}

      <input
        ref={fileInputRef}
        type="file"
        multiple
        hidden
        onChange={(e) => {
          if (e.target.files?.length) startUploads(e.target.files, currentFolderId);
          e.target.value = "";
        }}
      />

      <div className="shrink-0">
        <div className="mb-3 flex items-center justify-between gap-3">
          <Breadcrumb folders={folders} currentFolderId={currentFolderId} onNavigate={setCurrentFolderId} />

          <div className="flex items-center gap-1.5 shrink-0">
            <button
              onClick={() => setDialog({ type: "newFolder" })}
              className="flex items-center gap-1.5 h-7 rounded-lg px-2 text-[length:var(--text-pill)] text-(--text-muted) hover:text-(--fg) hover:bg-(--card-alt) transition-colors cursor-pointer"
            >
              <FolderPlus size={15} />
              New folder
            </button>
            <button
              onClick={() => fileInputRef.current?.click()}
              className="flex items-center gap-1.5 h-7 rounded-lg bg-(--fg) text-(--bg) px-2.5 text-[length:var(--text-pill)] cursor-pointer"
            >
              <Upload size={15} />
              Upload
            </button>
          </div>
        </div>

        <div className="mb-3 flex items-center gap-1.5">
          <div className="relative flex-1">
            <Search size={14} className="pointer-events-none absolute left-2.5 top-1/2 -translate-y-1/2 text-(--text-faint)" />
            <input
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              placeholder="Search documents by name"
              className="w-full rounded-lg border-(--line) border-[0.5px] border-solid bg-(--bg) pl-8 pr-3 py-1.5 text-[length:var(--text-caption)] text-(--fg) placeholder:text-(--text-faint) outline-none focus:border-(--line-strong)"
            />
          </div>

          <div className="relative" ref={labelMenuRef}>
            <button
              onClick={() => setShowLabelMenu((v) => !v)}
              aria-label="Filter by label"
              className={`flex items-center gap-1.5 h-8 rounded-lg px-2.5 text-[length:var(--text-pill)] transition-colors cursor-pointer ${
                showLabelMenu || activeLabel || labelFilter === NO_LABEL_FILTER
                  ? "bg-(--card-alt) text-(--fg)"
                  : "text-(--text-muted) hover:text-(--fg) hover:bg-(--card-alt)"
              }`}
            >
              {activeLabel ? (
                <>
                  <span className="size-2 rounded-full shrink-0" style={{ backgroundColor: LABEL_COLOR_VAR[activeLabel.color] }} />
                  {activeLabel.name}
                </>
              ) : labelFilter === NO_LABEL_FILTER ? (
                <>
                  <TagX size={15} />
                  No label
                </>
              ) : (
                <Tag size={15} />
              )}
            </button>

            {showLabelMenu && (
              <div className="absolute right-0 z-10 mt-2 w-52 max-h-56 overflow-y-auto rounded-lg border-(--line) border-[0.5px] border-solid bg-(--card) p-1.5 shadow-lg themed-scrollbar">
                <button
                  onClick={() => {
                    setLabelFilter(null);
                    setShowLabelMenu(false);
                  }}
                  className={`flex w-full items-center rounded-lg px-3 py-1.5 text-left text-[length:var(--text-pill)] transition-colors cursor-pointer ${
                    !labelFilter ? "bg-(--card-alt) text-(--fg)" : "text-(--text-muted) hover:text-(--fg) hover:bg-(--card-alt)"
                  }`}
                >
                  All labels
                </button>
                <button
                  onClick={() => {
                    setLabelFilter(NO_LABEL_FILTER);
                    setShowLabelMenu(false);
                  }}
                  className={`flex w-full items-center gap-1.5 rounded-lg px-3 py-1.5 text-left text-[length:var(--text-pill)] transition-colors cursor-pointer ${
                    labelFilter === NO_LABEL_FILTER
                      ? "bg-(--card-alt) text-(--fg)"
                      : "text-(--text-muted) hover:text-(--fg) hover:bg-(--card-alt)"
                  }`}
                >
                  <TagX size={12} className="shrink-0" />
                  <span className="truncate">No label</span>
                </button>
                {labels.map((l) => (
                  <button
                    key={l.id}
                    onClick={() => {
                      setLabelFilter(l.id);
                      setShowLabelMenu(false);
                    }}
                    className={`flex w-full items-center gap-1.5 rounded-lg px-3 py-1.5 text-left text-[length:var(--text-pill)] transition-colors cursor-pointer ${
                      labelFilter === l.id
                        ? "bg-(--card-alt) text-(--fg)"
                        : "text-(--text-muted) hover:text-(--fg) hover:bg-(--card-alt)"
                    }`}
                  >
                    <span className="size-2 rounded-full shrink-0" style={{ backgroundColor: LABEL_COLOR_VAR[l.color] }} />
                    <span className="truncate">{l.name}</span>
                  </button>
                ))}
                {labels.length === 0 && (
                  <p className="px-3 py-1.5 text-[length:var(--text-pill)] text-(--text-faint)">
                    No labels yet — add them in Settings.
                  </p>
                )}
              </div>
            )}
          </div>
        </div>

        {error && (
          <div className="mb-3 flex items-center justify-between gap-2 rounded-lg border-(--line) border-[0.5px] border-solid bg-(--card) px-3 py-2 text-[length:var(--text-pill)] text-red-400">
            <span>{error}</span>
            <button onClick={() => setError(null)} aria-label="Dismiss error" className="shrink-0 text-(--text-faint) hover:text-(--fg) cursor-pointer">
              <X size={12} />
            </button>
          </div>
        )}

        <UploadProgressList uploads={uploads} onDismiss={clearFinished} />
      </div>

      <div className="relative flex-1 min-h-0 overflow-y-auto themed-scrollbar pr-1">
        {dragging && (
          <div className="pointer-events-none absolute inset-0 z-10 flex items-center justify-center rounded-xl border-2 border-dashed border-(--line-strong) bg-(--bg)/70 text-[length:var(--text-caption)] text-(--text-muted)">
            Drop files to upload{currentFolderId ? " into this folder" : ""}
          </div>
        )}

        {loading ? (
          <div className="rounded-xl border-(--line) border-[0.5px] border-solid bg-(--card) p-8 text-center text-[length:var(--text-caption)] text-(--text-faint)">
            Loading…
          </div>
        ) : childFolders.length === 0 && visibleDocuments.length === 0 ? (
          <div className="rounded-xl border-(--line) border-[0.5px] border-solid bg-(--card) p-8 text-center text-[length:var(--text-caption)] text-(--text-faint)">
            {searching
              ? "No documents match your search."
              : spanning
                ? "No documents with this label."
                : "This folder is empty. Upload a file or create a folder."}
          </div>
        ) : (
          <div className="flex flex-col gap-1.5 pb-2">
            {childFolders.map((folder) => (
              <FolderRow
                key={folder.id}
                folder={folder}
                onOpen={() => setCurrentFolderId(folder.id)}
                onRename={() => setDialog({ type: "renameFolder", folder })}
                onMove={() => setDialog({ type: "moveFolder", folder })}
                onDelete={() => setDialog({ type: "deleteFolder", folder })}
              />
            ))}
            {visibleDocuments.map((doc) => (
              <DocumentRow
                key={doc.id}
                document={doc}
                labels={labels}
                showFolderHint={spanning}
                folderName={doc.folderId ? folderNameById.get(doc.folderId) ?? null : "Documents"}
                onDownload={() => handleDownload(doc)}
                onRename={() => setDialog({ type: "renameDoc", doc })}
                onMove={() => setDialog({ type: "moveDoc", doc })}
                onDelete={() => setDialog({ type: "deleteDoc", doc })}
                onLabelChange={(labelId) => handleLabelChange(doc, labelId)}
              />
            ))}
          </div>
        )}
      </div>

      {dialog?.type === "newFolder" && (
        <PromptDialog
          title="New folder"
          label="Folder name"
          confirmLabel="Create"
          onCancel={() => setDialog(null)}
          onConfirm={async (name) => {
            const folder = await createFolder({ name, parentId: currentFolderId });
            setFolders((prev) => [...prev, folder]);
            setDialog(null);
          }}
        />
      )}

      {dialog?.type === "renameFolder" && (
        <PromptDialog
          title="Rename folder"
          label="Folder name"
          initialValue={dialog.folder.name}
          onCancel={() => setDialog(null)}
          onConfirm={async (name) => {
            const updated = await updateFolder(dialog.folder.id, { name });
            setFolders((prev) => prev.map((f) => (f.id === updated.id ? updated : f)));
            setDialog(null);
          }}
        />
      )}

      {dialog?.type === "renameDoc" && (
        <PromptDialog
          title="Rename document"
          label="Name"
          initialValue={dialog.doc.name}
          onCancel={() => setDialog(null)}
          onConfirm={async (name) => {
            const updated = await updateDocument(dialog.doc.id, { name });
            setDocuments((prev) => prev.map((d) => (d.id === updated.id ? updated : d)));
            setDialog(null);
          }}
        />
      )}

      {dialog?.type === "moveFolder" && (
        <MoveDialog
          folders={folders}
          disabledIds={subtreeIds(folders, dialog.folder.id)}
          currentParentId={dialog.folder.parentId}
          title={`Move "${dialog.folder.name}"`}
          onCancel={() => setDialog(null)}
          onConfirm={async (destinationId) => {
            const updated = await updateFolder(dialog.folder.id, { parentId: destinationId });
            setFolders((prev) => prev.map((f) => (f.id === updated.id ? updated : f)));
            setDialog(null);
          }}
        />
      )}

      {dialog?.type === "moveDoc" && (
        <MoveDialog
          folders={folders}
          currentParentId={dialog.doc.folderId}
          title={`Move "${dialog.doc.name}"`}
          onCancel={() => setDialog(null)}
          onConfirm={async (destinationId) => {
            await updateDocument(dialog.doc.id, { folderId: destinationId });
            setDocuments((prev) => prev.filter((d) => d.id !== dialog.doc.id));
            setDialog(null);
          }}
        />
      )}

      {dialog?.type === "deleteFolder" && (
        <ConfirmDialog
          title="Delete folder?"
          body={
            <>
              <strong className="text-(--fg)">{dialog.folder.name}</strong> and everything inside it — subfolders and
              documents — will be permanently deleted.
            </>
          }
          onCancel={() => setDialog(null)}
          onConfirm={async () => {
            const gone = subtreeIds(folders, dialog.folder.id);
            await deleteFolder(dialog.folder.id);
            setFolders((prev) => prev.filter((f) => !gone.has(f.id)));
            if (currentFolderId && gone.has(currentFolderId)) setCurrentFolderId(dialog.folder.parentId);
            else reloadDocuments();
            setDialog(null);
            notify("Folder deleted");
          }}
        />
      )}

      {dialog?.type === "deleteDoc" && (
        <ConfirmDialog
          title="Delete document?"
          body={
            <>
              <strong className="text-(--fg)">{dialog.doc.name}</strong> will be permanently deleted.
            </>
          }
          onCancel={() => setDialog(null)}
          onConfirm={async () => {
            await deleteDocument(dialog.doc.id);
            setDocuments((prev) => prev.filter((d) => d.id !== dialog.doc.id));
            setDialog(null);
            notify("Document deleted");
          }}
        />
      )}
    </div>
  );
}
