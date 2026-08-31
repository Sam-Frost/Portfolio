import { useEffect, useState } from "react";
import { Plus } from "lucide-react";
import { createBlog, deleteBlog, fetchBlogs, updateBlog } from "./api";
import { ConfirmDialog } from "../../components/domain/ConfirmDialog";
import { ErrorBanner } from "./components";
import { ItemRow } from "./ItemRow";
import { BlogFormDialog } from "./BlogFormDialog";
import type { BlogInput, CmsBlog } from "./types";

type DialogState = null | { mode: "new" } | { mode: "edit"; blog: CmsBlog };

export function WritingsTab({ onChanged }: { onChanged: () => void }) {
  const [items, setItems] = useState<CmsBlog[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [dialog, setDialog] = useState<DialogState>(null);
  const [pendingDelete, setPendingDelete] = useState<CmsBlog | null>(null);

  useEffect(() => {
    fetchBlogs()
      .then(setItems)
      .catch((e) => setError(e instanceof Error ? e.message : "Couldn't load posts."))
      .finally(() => setLoading(false));
  }, []);

  function reload() {
    fetchBlogs().then(setItems).catch(() => {});
    onChanged();
  }

  async function handleCreate(input: BlogInput) {
    const created = await createBlog(input);
    setItems((prev) => [...prev, created]);
    onChanged();
  }

  async function handleEdit(id: string, input: BlogInput) {
    const updated = await updateBlog(id, input);
    setItems((prev) => prev.map((x) => (x.id === id ? updated : x)));
    onChanged();
  }

  function toggleVisible(b: CmsBlog, visible: boolean) {
    setItems((prev) => prev.map((x) => (x.id === b.id ? { ...x, visible } : x)));
    updateBlog(b.id, { visible }).then(onChanged).catch(reload);
  }

  function move(index: number, dir: -1 | 1) {
    const a = items[index];
    const b = items[index + dir];
    if (!a || !b) return;
    const reordered = [...items];
    reordered[index] = b;
    reordered[index + dir] = a;
    setItems(reordered);
    Promise.all([updateBlog(a.id, { order: b.order }), updateBlog(b.id, { order: a.order })])
      .then(onChanged)
      .catch(reload);
  }

  async function confirmDelete() {
    const b = pendingDelete;
    if (!b) return;
    await deleteBlog(b.id);
    setItems((prev) => prev.filter((x) => x.id !== b.id));
    setPendingDelete(null);
    onChanged();
  }

  return (
    <div>
      <button
        type="button"
        onClick={() => setDialog({ mode: "new" })}
        className="mb-3 inline-flex items-center gap-1.5 rounded-lg border-(--line) border-[0.5px] border-solid bg-(--card) px-3 py-1.5 text-[length:var(--text-pill)] text-(--text-muted) hover:text-(--fg) cursor-pointer"
      >
        <Plus size={13} /> Add post
      </button>

      {error && <ErrorBanner message={error} onDismiss={() => setError(null)} />}

      {loading ? (
        <p className="text-[length:var(--text-caption)] text-(--text-faint)">Loading…</p>
      ) : items.length === 0 ? (
        <p className="text-[length:var(--text-caption)] text-(--text-faint)">No posts yet.</p>
      ) : (
        <div className="flex flex-col gap-1.5">
          {items.map((b, i) => (
            <ItemRow
              key={b.id}
              title={b.title}
              subtitle={[b.readTime, b.genre, b.date].filter(Boolean).join(" · ")}
              visible={b.visible}
              isFirst={i === 0}
              isLast={i === items.length - 1}
              onMoveUp={() => move(i, -1)}
              onMoveDown={() => move(i, 1)}
              onToggleVisible={(v) => toggleVisible(b, v)}
              onEdit={() => setDialog({ mode: "edit", blog: b })}
              onDelete={() => setPendingDelete(b)}
            />
          ))}
        </div>
      )}

      {dialog?.mode === "new" && <BlogFormDialog onClose={() => setDialog(null)} onSubmit={handleCreate} />}
      {dialog?.mode === "edit" && (
        <BlogFormDialog
          initial={dialog.blog}
          onClose={() => setDialog(null)}
          onSubmit={(input) => handleEdit(dialog.blog.id, input)}
        />
      )}

      {pendingDelete && (
        <ConfirmDialog
          title="Delete post?"
          body={
            <>
              <strong className="text-(--fg)">{pendingDelete.title}</strong> will be removed. This can't be undone.
            </>
          }
          onCancel={() => setPendingDelete(null)}
          onConfirm={confirmDelete}
        />
      )}
    </div>
  );
}
