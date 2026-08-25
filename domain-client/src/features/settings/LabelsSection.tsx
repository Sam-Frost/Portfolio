import { Check, Pencil, Trash2, X } from "lucide-react";
import { useEffect, useState } from "react";
import { createLabel, deleteLabel, fetchLabels, updateLabel } from "../labels/api";
import { LABEL_COLOR_VAR } from "../labels/colors";
import { LABEL_COLORS, type Label, type LabelColor } from "../labels/types";

function ColorSwatchPicker({ value, onChange }: { value: LabelColor; onChange: (color: LabelColor) => void }) {
  return (
    <div className="flex items-center gap-1.5">
      {LABEL_COLORS.map((color) => (
        <button
          key={color}
          type="button"
          onClick={() => onChange(color)}
          aria-label={color}
          aria-pressed={value === color}
          className={`size-5 rounded-full cursor-pointer transition-shadow ${
            value === color ? "ring-2 ring-offset-2 ring-offset-(--card) ring-(--fg)" : ""
          }`}
          style={{ backgroundColor: LABEL_COLOR_VAR[color] }}
        />
      ))}
    </div>
  );
}

function LabelRow({
  label,
  onSave,
  onDelete,
}: {
  label: Label;
  onSave: (input: { name: string; color: LabelColor }) => Promise<void>;
  onDelete: () => void;
}) {
  const [editing, setEditing] = useState(false);
  const [name, setName] = useState(label.name);
  const [color, setColor] = useState<LabelColor>(label.color);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  if (!editing) {
    return (
      <div className="flex items-center gap-2 rounded-lg bg-(--card-alt) border-(--line) border-[0.5px] border-solid px-3 py-2">
        <span className="size-2.5 rounded-full shrink-0" style={{ backgroundColor: LABEL_COLOR_VAR[label.color] }} />
        <span className="flex-1 min-w-0 truncate text-[length:var(--text-caption)] text-(--fg)">{label.name}</span>
        <button
          type="button"
          onClick={() => setEditing(true)}
          aria-label="Edit label"
          className="text-(--text-faint) hover:text-(--fg) transition-colors cursor-pointer"
        >
          <Pencil size={13} />
        </button>
        <button
          type="button"
          onClick={onDelete}
          aria-label="Delete label"
          className="text-(--text-faint) hover:text-red-400 transition-colors cursor-pointer"
        >
          <Trash2 size={13} />
        </button>
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-2 rounded-lg bg-(--card-alt) border-(--line) border-[0.5px] border-solid px-3 py-2.5">
      <div className="flex items-center gap-2">
        <input
          autoFocus
          value={name}
          onChange={(e) => setName(e.target.value)}
          className="flex-1 min-w-0 rounded-md border-(--line) border-[0.5px] border-solid bg-(--bg) px-2 py-1 text-[length:var(--text-caption)] text-(--fg) outline-none focus:border-(--line-strong)"
        />
        <button
          type="button"
          disabled={saving}
          onClick={async () => {
            const trimmed = name.trim();
            if (!trimmed) return;
            setSaving(true);
            setError(null);
            try {
              await onSave({ name: trimmed, color });
              setEditing(false);
            } catch (err) {
              setError(err instanceof Error ? err.message : "Couldn't save label.");
            } finally {
              setSaving(false);
            }
          }}
          aria-label="Save label"
          className="text-(--green) hover:opacity-80 transition-opacity cursor-pointer disabled:opacity-50"
        >
          <Check size={15} />
        </button>
        <button
          type="button"
          onClick={() => {
            setName(label.name);
            setColor(label.color);
            setError(null);
            setEditing(false);
          }}
          aria-label="Cancel edit"
          className="text-(--text-faint) hover:text-(--fg) transition-colors cursor-pointer"
        >
          <X size={15} />
        </button>
      </div>
      <ColorSwatchPicker value={color} onChange={setColor} />
      {error && <p className="text-[length:var(--text-pill)] text-red-400">{error}</p>}
    </div>
  );
}

export function LabelsSection() {
  const [labels, setLabels] = useState<Label[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const [newName, setNewName] = useState("");
  const [newColor, setNewColor] = useState<LabelColor>(LABEL_COLORS[0]);
  const [adding, setAdding] = useState(false);

  useEffect(() => {
    fetchLabels()
      .then(setLabels)
      .catch((err) => setError(err instanceof Error ? err.message : "Couldn't load labels."))
      .finally(() => setLoading(false));
  }, []);

  async function handleAdd() {
    const trimmed = newName.trim();
    if (!trimmed) return;
    setAdding(true);
    setError(null);
    try {
      const label = await createLabel({ name: trimmed, color: newColor });
      setLabels((prev) => [...prev, label].sort((a, b) => a.name.localeCompare(b.name)));
      setNewName("");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Couldn't add label.");
    } finally {
      setAdding(false);
    }
  }

  async function handleSave(id: string, input: { name: string; color: LabelColor }) {
    const updated = await updateLabel(id, input);
    setLabels((prev) => prev.map((l) => (l.id === id ? updated : l)).sort((a, b) => a.name.localeCompare(b.name)));
  }

  function handleDelete(id: string) {
    const original = labels;
    setLabels((prev) => prev.filter((l) => l.id !== id));
    deleteLabel(id).catch((err) => {
      setLabels(original);
      setError(err instanceof Error ? err.message : "Couldn't delete label.");
    });
  }

  return (
    <div className="flex flex-col gap-3 max-w-sm">
      {loading && <p className="text-[length:var(--text-caption)] text-(--text-faint)">Loading labels...</p>}

      {!loading && (
        <>
          {labels.length > 0 && (
            <div className="flex flex-col gap-1.5">
              {labels.map((label) => (
                <LabelRow
                  key={label.id}
                  label={label}
                  onSave={(input) => handleSave(label.id, input)}
                  onDelete={() => handleDelete(label.id)}
                />
              ))}
            </div>
          )}

          <div className="flex flex-col gap-2 rounded-lg border-(--line) border-[0.5px] border-solid px-3 py-2.5">
            <span className="text-[length:var(--text-pill)] text-(--text-muted)">Add a label</span>
            <div className="flex items-center gap-2">
              <input
                value={newName}
                onChange={(e) => setNewName(e.target.value)}
                onKeyDown={(e) => {
                  if (e.key === "Enter") handleAdd();
                }}
                placeholder="Label name"
                className="flex-1 min-w-0 rounded-md border-(--line) border-[0.5px] border-solid bg-(--bg) px-2 py-1 text-[length:var(--text-caption)] text-(--fg) placeholder:text-(--text-faint) outline-none focus:border-(--line-strong)"
              />
              <button
                type="button"
                disabled={adding || !newName.trim()}
                onClick={handleAdd}
                className="rounded-md bg-(--fg) text-(--bg) px-3 py-1 text-[length:var(--text-pill)] cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed"
              >
                Add
              </button>
            </div>
            <ColorSwatchPicker value={newColor} onChange={setNewColor} />
          </div>

          {error && (
            <div className="flex items-center justify-between gap-2 rounded-lg border-(--line) border-[0.5px] border-solid bg-(--card-alt) px-3 py-2 text-[length:var(--text-pill)] text-red-400">
              <span>{error}</span>
              <button
                onClick={() => setError(null)}
                aria-label="Dismiss error"
                className="shrink-0 text-(--text-faint) hover:text-(--fg) transition-colors cursor-pointer"
              >
                <X size={12} />
              </button>
            </div>
          )}
        </>
      )}
    </div>
  );
}
