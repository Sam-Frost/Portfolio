import { useEffect, useState } from "react";
import { Plus } from "lucide-react";
import { createExperience, deleteExperience, fetchExperiences, updateExperience } from "./api";
import { ConfirmDialog } from "../../components/domain/ConfirmDialog";
import { ErrorBanner } from "./components";
import { ItemRow } from "./ItemRow";
import { ExperienceFormDialog } from "./ExperienceFormDialog";
import type { CmsExperience, ExperienceInput } from "./types";

type DialogState = null | { mode: "new" } | { mode: "edit"; experience: CmsExperience };

export function ExperienceTab({ onChanged }: { onChanged: () => void }) {
  const [items, setItems] = useState<CmsExperience[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [dialog, setDialog] = useState<DialogState>(null);
  const [pendingDelete, setPendingDelete] = useState<CmsExperience | null>(null);

  useEffect(() => {
    fetchExperiences()
      .then(setItems)
      .catch((e) => setError(e instanceof Error ? e.message : "Couldn't load experience."))
      .finally(() => setLoading(false));
  }, []);

  function reload() {
    fetchExperiences().then(setItems).catch(() => {});
    onChanged();
  }

  async function handleCreate(input: ExperienceInput) {
    const created = await createExperience(input);
    setItems((prev) => [...prev, created]);
    onChanged();
  }

  async function handleEdit(id: string, input: ExperienceInput) {
    const updated = await updateExperience(id, input);
    setItems((prev) => prev.map((x) => (x.id === id ? updated : x)));
    onChanged();
  }

  function toggleVisible(e: CmsExperience, visible: boolean) {
    setItems((prev) => prev.map((x) => (x.id === e.id ? { ...x, visible } : x)));
    updateExperience(e.id, { visible }).then(onChanged).catch(reload);
  }

  function move(index: number, dir: -1 | 1) {
    const a = items[index];
    const b = items[index + dir];
    if (!a || !b) return;
    const reordered = [...items];
    reordered[index] = b;
    reordered[index + dir] = a;
    setItems(reordered);
    Promise.all([updateExperience(a.id, { order: b.order }), updateExperience(b.id, { order: a.order })])
      .then(onChanged)
      .catch(reload);
  }

  async function confirmDelete() {
    const e = pendingDelete;
    if (!e) return;
    await deleteExperience(e.id);
    setItems((prev) => prev.filter((x) => x.id !== e.id));
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
        <Plus size={13} /> Add experience
      </button>

      {error && <ErrorBanner message={error} onDismiss={() => setError(null)} />}

      {loading ? (
        <p className="text-[length:var(--text-caption)] text-(--text-faint)">Loading…</p>
      ) : items.length === 0 ? (
        <p className="text-[length:var(--text-caption)] text-(--text-faint)">No experience entries yet.</p>
      ) : (
        <div className="flex flex-col gap-1.5">
          {items.map((e, i) => (
            <ItemRow
              key={e.id}
              title={`${e.position} · ${e.company}`}
              subtitle={`${e.startDate} – ${e.endDate}`}
              visible={e.visible}
              isFirst={i === 0}
              isLast={i === items.length - 1}
              onMoveUp={() => move(i, -1)}
              onMoveDown={() => move(i, 1)}
              onToggleVisible={(v) => toggleVisible(e, v)}
              onEdit={() => setDialog({ mode: "edit", experience: e })}
              onDelete={() => setPendingDelete(e)}
            />
          ))}
        </div>
      )}

      {dialog?.mode === "new" && <ExperienceFormDialog onClose={() => setDialog(null)} onSubmit={handleCreate} />}
      {dialog?.mode === "edit" && (
        <ExperienceFormDialog
          initial={dialog.experience}
          onClose={() => setDialog(null)}
          onSubmit={(input) => handleEdit(dialog.experience.id, input)}
        />
      )}

      {pendingDelete && (
        <ConfirmDialog
          title="Delete experience?"
          body={
            <>
              <strong className="text-(--fg)">
                {pendingDelete.position} · {pendingDelete.company}
              </strong>{" "}
              will be removed. This can't be undone.
            </>
          }
          onCancel={() => setPendingDelete(null)}
          onConfirm={confirmDelete}
        />
      )}
    </div>
  );
}
