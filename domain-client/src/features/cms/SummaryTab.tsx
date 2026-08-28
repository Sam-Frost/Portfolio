import { useEffect, useState } from "react";
import { fetchSummary, updateSummary } from "./api";
import { ErrorBanner, Field, TextArea, TextField } from "./components";
import type { CmsSummary } from "./types";

const FIELDS: { key: keyof Omit<CmsSummary, "updatedAt">; label: string; area?: boolean }[] = [
  { key: "heroName", label: "Name" },
  { key: "heroHighlightText", label: "Highlight pill (e.g. “Open to backend roles”)" },
  { key: "heroSubText", label: "Subtitle (e.g. “Backend engineer · distributed systems”)" },
  { key: "heroDetails", label: "Bio paragraph", area: true },
  { key: "imageSubText", label: "Text under the avatar" },
  { key: "domain", label: "Domain label" },
];

export function SummaryTab({ onChanged }: { onChanged: () => void }) {
  const [summary, setSummary] = useState<CmsSummary | null>(null);
  const [draft, setDraft] = useState<Omit<CmsSummary, "updatedAt"> | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);
  const [saved, setSaved] = useState(false);

  useEffect(() => {
    fetchSummary()
      .then((s) => {
        setSummary(s);
        setDraft(s);
      })
      .catch((e) => setError(e instanceof Error ? e.message : "Couldn't load summary."));
  }, []);

  if (!draft || !summary) {
    return error ? (
      <ErrorBanner message={error} onDismiss={() => setError(null)} />
    ) : (
      <p className="text-[length:var(--text-caption)] text-(--text-faint)">Loading…</p>
    );
  }

  const dirty = FIELDS.some(({ key }) => draft[key] !== summary[key]);

  async function save() {
    setSaving(true);
    setError(null);
    try {
      const updated = await updateSummary(draft!);
      setSummary(updated);
      setDraft(updated);
      setSaved(true);
      setTimeout(() => setSaved(false), 1500);
      onChanged();
    } catch (e) {
      setError(e instanceof Error ? e.message : "Couldn't save.");
    } finally {
      setSaving(false);
    }
  }

  return (
    <div className="max-w-xl">
      {error && <ErrorBanner message={error} onDismiss={() => setError(null)} />}

      {FIELDS.map(({ key, label, area }) => (
        <Field key={key} label={label}>
          {area ? (
            <TextArea value={draft[key]} onChange={(v) => setDraft({ ...draft, [key]: v })} rows={4} />
          ) : (
            <TextField value={draft[key]} onChange={(v) => setDraft({ ...draft, [key]: v })} />
          )}
        </Field>
      ))}

      <button
        type="button"
        onClick={save}
        disabled={!dirty || saving}
        className={`mt-1 rounded-md px-3 py-1.5 text-[length:var(--text-pill)] ${
          !dirty || saving ? "opacity-50 cursor-not-allowed bg-(--card-alt) text-(--text-muted)" : "bg-(--fg) text-(--bg) cursor-pointer"
        }`}
      >
        {saving ? "Saving…" : saved ? "Saved" : "Save summary"}
      </button>
    </div>
  );
}
