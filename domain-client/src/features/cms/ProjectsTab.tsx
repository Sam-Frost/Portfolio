import { useEffect, useState } from "react";
import { Plus } from "lucide-react";
import { createProject, deleteProject, fetchProjects, updateProject } from "./api";
import { ErrorBanner } from "./components";
import { ItemRow } from "./ItemRow";
import { ProjectFormDialog } from "./ProjectFormDialog";
import type { CmsProject, ProjectInput } from "./types";

type DialogState = null | { mode: "new" } | { mode: "edit"; project: CmsProject };

export function ProjectsTab({ onChanged }: { onChanged: () => void }) {
  const [projects, setProjects] = useState<CmsProject[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [dialog, setDialog] = useState<DialogState>(null);

  useEffect(() => {
    fetchProjects()
      .then(setProjects)
      .catch((e) => setError(e instanceof Error ? e.message : "Couldn't load projects."))
      .finally(() => setLoading(false));
  }, []);

  function reload() {
    fetchProjects().then(setProjects).catch(() => {});
    onChanged();
  }

  async function handleCreate(input: ProjectInput) {
    const created = await createProject(input);
    setProjects((prev) => [...prev, created]);
    onChanged();
  }

  async function handleEdit(id: string, input: ProjectInput) {
    const updated = await updateProject(id, input);
    setProjects((prev) => prev.map((p) => (p.id === id ? updated : p)));
    onChanged();
  }

  function toggleVisible(p: CmsProject, visible: boolean) {
    setProjects((prev) => prev.map((x) => (x.id === p.id ? { ...x, visible } : x)));
    updateProject(p.id, { visible }).then(onChanged).catch(reload);
  }

  function move(index: number, dir: -1 | 1) {
    const a = projects[index];
    const b = projects[index + dir];
    if (!a || !b) return;
    const reordered = [...projects];
    reordered[index] = b;
    reordered[index + dir] = a;
    setProjects(reordered);
    Promise.all([updateProject(a.id, { order: b.order }), updateProject(b.id, { order: a.order })])
      .then(onChanged)
      .catch(reload);
  }

  function remove(p: CmsProject) {
    if (!window.confirm(`Delete "${p.title}"? This can't be undone.`)) return;
    setProjects((prev) => prev.filter((x) => x.id !== p.id));
    deleteProject(p.id).then(onChanged).catch(reload);
  }

  return (
    <div>
      <button
        type="button"
        onClick={() => setDialog({ mode: "new" })}
        className="mb-3 inline-flex items-center gap-1.5 rounded-lg border-(--line) border-[0.5px] border-solid bg-(--card) px-3 py-1.5 text-[length:var(--text-pill)] text-(--text-muted) hover:text-(--fg) cursor-pointer"
      >
        <Plus size={13} /> Add project
      </button>

      {error && <ErrorBanner message={error} onDismiss={() => setError(null)} />}

      {loading ? (
        <p className="text-[length:var(--text-caption)] text-(--text-faint)">Loading…</p>
      ) : projects.length === 0 ? (
        <p className="text-[length:var(--text-caption)] text-(--text-faint)">No projects yet.</p>
      ) : (
        <div className="flex flex-col gap-1.5">
          {projects.map((p, i) => (
            <ItemRow
              key={p.id}
              title={p.title}
              subtitle={p.stack.join(" · ")}
              visible={p.visible}
              isFirst={i === 0}
              isLast={i === projects.length - 1}
              onMoveUp={() => move(i, -1)}
              onMoveDown={() => move(i, 1)}
              onToggleVisible={(v) => toggleVisible(p, v)}
              onEdit={() => setDialog({ mode: "edit", project: p })}
              onDelete={() => remove(p)}
            />
          ))}
        </div>
      )}

      {dialog?.mode === "new" && (
        <ProjectFormDialog onClose={() => setDialog(null)} onSubmit={handleCreate} />
      )}
      {dialog?.mode === "edit" && (
        <ProjectFormDialog
          initial={dialog.project}
          onClose={() => setDialog(null)}
          onSubmit={(input) => handleEdit(dialog.project.id, input)}
        />
      )}
    </div>
  );
}
