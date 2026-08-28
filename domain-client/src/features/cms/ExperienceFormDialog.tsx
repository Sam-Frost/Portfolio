import { useState } from "react";
import { Dialog, DialogButton, Field, LinesInput, TagInput, TextArea, TextField, VisibilityToggle } from "./components";
import type { CmsExperience, ExperienceInput } from "./types";

interface Props {
  initial?: CmsExperience;
  onClose: () => void;
  onSubmit: (input: ExperienceInput) => Promise<void>;
}

export function ExperienceFormDialog({ initial, onClose, onSubmit }: Props) {
  const [logo, setLogo] = useState(initial?.logo ?? "");
  const [position, setPosition] = useState(initial?.position ?? "");
  const [company, setCompany] = useState(initial?.company ?? "");
  const [description, setDescription] = useState(initial?.description ?? "");
  const [details, setDetails] = useState<string[]>(initial?.details ?? []);
  const [techStack, setTechStack] = useState<string[]>(initial?.techStack ?? []);
  const [startDate, setStartDate] = useState(initial?.startDate ?? "");
  const [endDate, setEndDate] = useState(initial?.endDate ?? "");
  const [visible, setVisible] = useState(initial?.visible ?? true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function save() {
    if (!position.trim() || !company.trim()) {
      setError("Position and company are required.");
      return;
    }
    setSaving(true);
    setError(null);
    try {
      await onSubmit({
        logo,
        position: position.trim(),
        company: company.trim(),
        description,
        details: details.filter((d) => d.trim()),
        techStack,
        startDate,
        endDate,
        visible,
      });
      onClose();
    } catch (e) {
      setError(e instanceof Error ? e.message : "Couldn't save.");
      setSaving(false);
    }
  }

  return (
    <Dialog
      title={initial ? "Edit experience" : "New experience"}
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

      <div className="flex gap-3">
        <div className="w-20 shrink-0">
          <Field label="Logo">
            <TextField value={logo} onChange={setLogo} placeholder="SL" />
          </Field>
        </div>
        <div className="flex-1">
          <Field label="Position">
            <TextField value={position} onChange={setPosition} autoFocus />
          </Field>
        </div>
      </div>
      <Field label="Company">
        <TextField value={company} onChange={setCompany} />
      </Field>
      <div className="flex gap-3">
        <div className="flex-1">
          <Field label="Start date">
            <TextField value={startDate} onChange={setStartDate} placeholder="June 2024" />
          </Field>
        </div>
        <div className="flex-1">
          <Field label="End date">
            <TextField value={endDate} onChange={setEndDate} placeholder="Present" />
          </Field>
        </div>
      </div>
      <Field label="Description">
        <TextArea value={description} onChange={setDescription} rows={2} />
      </Field>
      <Field label="Details (one bullet per line)">
        <LinesInput values={details} onChange={setDetails} />
      </Field>
      <Field label="Tech stack">
        <TagInput values={techStack} onChange={setTechStack} />
      </Field>

      <VisibilityToggle visible={visible} onChange={setVisible} />
    </Dialog>
  );
}
