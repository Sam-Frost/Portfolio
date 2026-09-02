import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { PenTool, Plus, Trash2 } from "lucide-react";
import { createDrawingBoard, deleteDrawingBoard, fetchDrawingBoards } from "./api";
import { DeleteBoardDialog } from "./DeleteBoardDialog";
import type { DrawingBoardSummary } from "./types";

function formatDate(value: string) {
  return new Date(value).toLocaleDateString(undefined, { year: "numeric", month: "short", day: "numeric" });
}

export function DrawingBoardListPage() {
  const navigate = useNavigate();
  const [boards, setBoards] = useState<DrawingBoardSummary[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [creating, setCreating] = useState(false);
  const [pendingDelete, setPendingDelete] = useState<DrawingBoardSummary | null>(null);

  useEffect(() => {
    setLoading(true);
    fetchDrawingBoards()
      .then(setBoards)
      .catch((err) => setError(err instanceof Error ? err.message : "Couldn't load boards."))
      .finally(() => setLoading(false));
  }, []);

  function handleCreate() {
    setCreating(true);
    createDrawingBoard()
      .then((board) => navigate(`/drawing-board/${board.id}`))
      .catch((err) => setError(err instanceof Error ? err.message : "Couldn't create board."))
      .finally(() => setCreating(false));
  }

  function handleDelete(id: string) {
    const previous = boards;
    setBoards((prev) => prev.filter((b) => b.id !== id));
    setPendingDelete(null);
    deleteDrawingBoard(id).catch((err) => {
      setBoards(previous);
      setError(err instanceof Error ? err.message : "Couldn't delete board.");
    });
  }

  return (
    <div className="min-h-full lg:h-full flex flex-col">
      <div className="shrink-0 flex items-center justify-between gap-2 mb-3">
        <span className="text-[length:var(--text-caption)] text-(--text-muted)">
          {boards.length === 1 ? "1 board" : `${boards.length} boards`}
        </span>
        <button
          onClick={handleCreate}
          disabled={creating}
          className={`flex items-center gap-1.5 rounded-lg bg-(--fg) text-(--bg) px-3 py-1.5 text-[length:var(--text-caption)] transition-opacity ${
            creating ? "opacity-60 cursor-not-allowed" : "cursor-pointer"
          }`}
        >
          <Plus size={14} />
          New board
        </button>
      </div>

      {error && (
        <div className="shrink-0 mb-3 flex items-center justify-between gap-2 rounded-lg border-(--line) border-[0.5px] border-solid bg-(--card) px-3 py-2 text-[length:var(--text-pill)] text-red-400">
          <span>{error}</span>
          <button onClick={() => setError(null)} className="shrink-0 text-(--text-faint) hover:text-(--fg) transition-colors cursor-pointer">
            Dismiss
          </button>
        </div>
      )}

      <div className="flex-1 lg:min-h-0 lg:overflow-y-auto themed-scrollbar">
        {loading && (
          <div className="bg-(--card) border-(--line) border-[0.5px] border-solid rounded-xl p-8 text-center text-(--text-faint) text-[length:var(--text-caption)]">
            Loading boards...
          </div>
        )}

        {!loading && boards.length === 0 && (
          <div className="bg-(--card) border-(--line) border-[0.5px] border-solid rounded-xl p-8 text-center text-(--text-faint) text-[length:var(--text-caption)]">
            No boards yet.
          </div>
        )}

        {!loading && boards.length > 0 && (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-2 pb-2">
            {boards.map((board) => (
              <div
                key={board.id}
                onClick={() => navigate(`/drawing-board/${board.id}`)}
                className="group relative flex flex-col gap-1 bg-(--card) border-(--line) border-[0.5px] border-solid rounded-lg px-3 py-2.5 cursor-pointer hover:border-(--line-strong) transition-colors"
              >
                <span className="flex items-center gap-1.5 pr-8 text-[length:var(--text-caption)] text-(--fg)">
                  <PenTool size={12} className="shrink-0 text-(--text-faint)" />
                  <span className="truncate">{board.name}</span>
                </span>
                <span className="text-[length:var(--text-pill)] text-(--text-faint)">
                  Edited {formatDate(board.updatedAt)}
                </span>

                <div className="absolute top-2 right-2 flex items-center gap-1.5 opacity-100 lg:opacity-0 lg:group-hover:opacity-100 transition-opacity">
                  <button
                    onClick={(e) => {
                      e.stopPropagation();
                      setPendingDelete(board);
                    }}
                    aria-label="Delete board"
                    className="text-(--text-faint) hover:text-(--label-red) transition-colors cursor-pointer"
                  >
                    <Trash2 size={13} />
                  </button>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      {pendingDelete && (
        <DeleteBoardDialog
          name={pendingDelete.name}
          onCancel={() => setPendingDelete(null)}
          onConfirm={() => handleDelete(pendingDelete.id)}
        />
      )}
    </div>
  );
}
