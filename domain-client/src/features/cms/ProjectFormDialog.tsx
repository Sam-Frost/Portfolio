import { useState } from "react";
import { Dialog, DialogButton, Field, TagInput, TextArea, TextField, VisibilityToggle, slugify } from "./components";
import type { CmsProject, ProjectInput } from "./types";

interface Props {
  initial?: CmsProject;
  onClose: () => void;
  onSubmit: (input: ProjectInput) => Promise<void>;
}

export function ProjectFormDialog({ initial, onClose, onSubmit }: Props) {
  const [title, setTitle] = useState(initial?.title ?? "");
  const [slug, setSlug] = useState(initial?.slug ?? "");
  const [slugTouched, setSlugTouched] = useState(Boolean(initial));
  const [description, setDescription] = useState(initial?.description ?? "");
  const [stack, setStack] = useState<string[]>(initial?.stack ?? []);
  const [github, setGithub] = useState(initial?.github ?? "");
  const [liveLink, setLiveLink] = useState(initial?.liveLink ?? "");
  const [visible, setVisible] = useState(initial?.visible ?? true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const effectiveSlug = slugTouched ? slug : slugify(title);

  async function save() {
    if (!title.trim()) {
      setError("Title is required.");
      return;
    }
    setSaving(true);
    setError(null);
    try {
      await onSubmit({ title: title.trim(), slug: effectiveSlug, description, stack, github, liveLink, visible });
      onClose();
    } catch (e) {
      setError(e instanceof Error ? e.message : "Couldn't save.");
      setSaving(false);
    }
  }

  return (
    <Dialog
      title={initial ? "Edit project" : "New project"}
      onClose={onClose}
      footer={
        <>
          <DialogButton onClick={onClose}>Cancel</DialogButton>
          <DialogButton onClick={save} primary disabled={saving}>
            {saving ? "Saving…" : "Save"}
          </DialogButton>
        </>
      }
    >
      {error && <p className="mb-3 text-[length:var(--text-pill)] text-red-400">{error}</p>}

      <Field label="Title">
        <TextField value={title} onChange={setTitle} autoFocus />
      </Field>
      <Field label="Slug">
        <TextField
          value={effectiveSlug}
          onChange={(v) => {
            setSlugTouched(true);
            setSlug(v);
          }}
          placeholder="auto-generated from title"
        />
      </Field>
      <Field label="Description">
        <TextArea value={description} onChange={setDescription} rows={2} />
      </Field>
      <Field label="Stack">
        <TagInput values={stack} onChange={setStack} />
      </Field>
      <Field label="GitHub URL">
        <TextField value={github} onChange={setGithub} placeholder="https://github.com/…" />
      </Field>
      <Field label="Live URL">
        <TextField value={liveLink} onChange={setLiveLink} placeholder="https://…" />
      </Field>

      <VisibilityToggle visible={visible} onChange={setVisible} />
    </Dialog>
  );
}
