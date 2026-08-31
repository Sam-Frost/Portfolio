import { useState } from "react";
import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";
import { Dialog, DialogButton, Field, TextField, VisibilityToggle, slugify } from "./components";
import type { BlogInput, CmsBlog } from "./types";

interface Props {
  initial?: CmsBlog;
  onClose: () => void;
  onSubmit: (input: BlogInput) => Promise<void>;
}

export function BlogFormDialog({ initial, onClose, onSubmit }: Props) {
  const [title, setTitle] = useState(initial?.title ?? "");
  const [slug, setSlug] = useState(initial?.slug ?? "");
  const [slugTouched, setSlugTouched] = useState(Boolean(initial));
  const [readTime, setReadTime] = useState(initial?.readTime ?? "");
  const [genre, setGenre] = useState(initial?.genre ?? "");
  const [date, setDate] = useState(initial?.date ?? "");
  const [body, setBody] = useState(initial?.body ?? "");
  const [visible, setVisible] = useState(initial?.visible ?? true);
  const [tab, setTab] = useState<"write" | "preview">("write");
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const effectiveSlug = slugTouched ? slug : slugify(title);

  async function save() {
    if (!title.trim()) {
      setError("Title is required.");
      return;
    }
    if (!body.trim()) {
      setError("Body is required.");
      return;
    }
    setSaving(true);
    setError(null);
    try {
      await onSubmit({ title: title.trim(), slug: effectiveSlug, readTime, genre, date, body, visible });
      onClose();
    } catch (e) {
      setError(e instanceof Error ? e.message : "Couldn't save.");
      setSaving(false);
    }
  }

  const tabClass = (active: boolean) =>
    `px-2.5 py-1 text-[length:var(--text-pill)] rounded-md cursor-pointer transition-colors ${
      active ? "bg-(--card-alt) text-(--fg)" : "text-(--text-muted) hover:text-(--fg)"
    }`;

  return (
    <Dialog
      title={initial ? "Edit post" : "New post"}
      onClose={onClose}
      wide
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

      <div className="flex flex-col sm:flex-row gap-x-3">
        <div className="flex-1">
          <Field label="Title">
            <TextField value={title} onChange={setTitle} autoFocus />
          </Field>
        </div>
        <div className="flex-1">
          <Field label="Slug">
            <TextField
              value={effectiveSlug}
              onChange={(v) => {
                setSlugTouched(true);
                setSlug(v);
              }}
            />
          </Field>
        </div>
      </div>
      <div className="flex flex-col sm:flex-row gap-x-3">
        <div className="flex-1">
          <Field label="Read time">
            <TextField value={readTime} onChange={setReadTime} placeholder="5 min read" />
          </Field>
        </div>
        <div className="flex-1">
          <Field label="Genre">
            <TextField value={genre} onChange={setGenre} placeholder="Backend" />
          </Field>
        </div>
        <div className="flex-1">
          <Field label="Date">
            <TextField value={date} onChange={setDate} placeholder="Dec 2024" />
          </Field>
        </div>
      </div>

      <div className="mb-1 flex items-center gap-1">
        <button type="button" className={tabClass(tab === "write")} onClick={() => setTab("write")}>
          Write
        </button>
        <button type="button" className={tabClass(tab === "preview")} onClick={() => setTab("preview")}>
          Preview
        </button>
        <span className="ml-auto text-[length:var(--text-pill)] text-(--text-faint)">Markdown</span>
      </div>

      {tab === "write" ? (
        <textarea
          value={body}
          rows={16}
          onChange={(e) => setBody(e.target.value)}
          placeholder="# Heading&#10;&#10;Write your post in markdown…"
          className="w-full rounded-lg bg-(--card-alt) border-(--line) border-[0.5px] border-solid px-2.5 py-1.5 font-mono text-[length:var(--text-pill)] leading-relaxed text-(--fg) resize-y focus:outline-none"
        />
      ) : (
        <div className="prose prose-invert prose-sm max-w-none min-h-[16rem] rounded-lg bg-(--card-alt) border-(--line) border-[0.5px] border-solid px-3 py-2 overflow-y-auto">
          <ReactMarkdown remarkPlugins={[remarkGfm]}>{body || "_Nothing to preview yet._"}</ReactMarkdown>
        </div>
      )}

      <div className="mt-3">
        <VisibilityToggle visible={visible} onChange={setVisible} />
      </div>
    </Dialog>
  );
}
