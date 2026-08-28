import { useEffect, type KeyboardEvent, type ReactNode } from "react";
import { Eye, EyeOff, X } from "lucide-react";

// Shared form primitives for the CMS editors. Styling matches the existing
// domain dialogs (LeaveDomainDialog, EditTopicDialog): CSS-variable colors,
// 0.5px borders, rounded-lg/xl.

const inputClass =
  "w-full rounded-lg bg-(--card-alt) border-(--line) border-[0.5px] border-solid px-2.5 py-1.5 text-[length:var(--text-caption)] text-(--fg) focus:outline-none";

export function Field({ label, children }: { label: string; children: ReactNode }) {
  return (
    <label className="block mb-3">
      <span className="block text-[length:var(--text-pill)] text-(--text-muted) mb-1">{label}</span>
      {children}
    </label>
  );
}

export function TextField({
  value,
  onChange,
  placeholder,
  autoFocus,
}: {
  value: string;
  onChange: (v: string) => void;
  placeholder?: string;
  autoFocus?: boolean;
}) {
  return (
    <input
      autoFocus={autoFocus}
      value={value}
      placeholder={placeholder}
      onChange={(e) => onChange(e.target.value)}
      className={inputClass}
    />
  );
}

export function TextArea({
  value,
  onChange,
  rows = 3,
  placeholder,
  mono,
}: {
  value: string;
  onChange: (v: string) => void;
  rows?: number;
  placeholder?: string;
  mono?: boolean;
}) {
  return (
    <textarea
      value={value}
      rows={rows}
      placeholder={placeholder}
      onChange={(e) => onChange(e.target.value)}
      className={`${inputClass} resize-y leading-relaxed ${mono ? "font-mono text-[length:var(--text-pill)]" : ""}`}
    />
  );
}

// TagInput edits a string[] as removable chips + a text box that splits on
// Enter or comma. Used for project stack / experience tech stack.
export function TagInput({ values, onChange }: { values: string[]; onChange: (v: string[]) => void }) {
  function add(raw: string) {
    const parts = raw
      .split(",")
      .map((s) => s.trim())
      .filter(Boolean);
    if (parts.length === 0) return;
    onChange([...values, ...parts.filter((p) => !values.includes(p))]);
  }

  function handleKeyDown(e: KeyboardEvent<HTMLInputElement>) {
    const target = e.currentTarget;
    if (e.key === "Enter" || e.key === ",") {
      e.preventDefault();
      add(target.value);
      target.value = "";
    } else if (e.key === "Backspace" && target.value === "" && values.length > 0) {
      onChange(values.slice(0, -1));
    }
  }

  return (
    <div className={`${inputClass} flex flex-wrap gap-1.5`}>
      {values.map((v) => (
        <span
          key={v}
          className="inline-flex items-center gap-1 rounded-md bg-(--line-soft) px-1.5 py-0.5 text-[length:var(--text-pill)] text-(--text-muted)"
        >
          {v}
          <button type="button" onClick={() => onChange(values.filter((x) => x !== v))} className="cursor-pointer hover:text-(--fg)">
            <X size={10} />
          </button>
        </span>
      ))}
      <input
        onKeyDown={handleKeyDown}
        onBlur={(e) => {
          add(e.target.value);
          e.target.value = "";
        }}
        placeholder={values.length === 0 ? "type and press Enter" : ""}
        className="flex-1 min-w-24 bg-transparent text-[length:var(--text-pill)] text-(--fg) placeholder:text-(--text-faint) focus:outline-none"
      />
    </div>
  );
}

// LinesInput edits a string[] as one <textarea> line per item — used for an
// experience's bullet-point details.
export function LinesInput({ values, onChange, rows = 5 }: { values: string[]; onChange: (v: string[]) => void; rows?: number }) {
  return (
    <textarea
      value={values.join("\n")}
      rows={rows}
      placeholder="One bullet point per line"
      onChange={(e) => onChange(e.target.value.split("\n"))}
      className={`${inputClass} resize-y leading-relaxed`}
    />
  );
}

export function VisibilityToggle({ visible, onChange }: { visible: boolean; onChange: (v: boolean) => void }) {
  return (
    <button
      type="button"
      onClick={() => onChange(!visible)}
      className={`inline-flex items-center gap-1.5 rounded-md px-2 py-1 text-[length:var(--text-pill)] transition-colors cursor-pointer ${
        visible ? "text-(--fg) hover:bg-(--card-alt)" : "text-(--text-faint) hover:bg-(--card-alt)"
      }`}
      aria-pressed={visible}
    >
      {visible ? <Eye size={13} /> : <EyeOff size={13} />}
      {visible ? "Visible" : "Hidden"}
    </button>
  );
}

// Dialog is the shared modal shell (backdrop click + Escape to close),
// matching EditTopicDialog. `wide` widens it for the blog editor.
export function Dialog({
  title,
  onClose,
  children,
  footer,
  wide,
}: {
  title: string;
  onClose: () => void;
  children: ReactNode;
  footer: ReactNode;
  wide?: boolean;
}) {
  useEffect(() => {
    function onKey(e: globalThis.KeyboardEvent) {
      if (e.key === "Escape") onClose();
    }
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [onClose]);

  return (
    <div className="fixed inset-0 z-50 flex items-start justify-center overflow-y-auto bg-black/50 px-4 py-10" onClick={onClose} role="presentation">
      <div
        className={`w-full ${wide ? "max-w-3xl" : "max-w-md"} rounded-xl border-(--line) border-[0.5px] border-solid bg-(--card) p-4 shadow-lg`}
        onClick={(e) => e.stopPropagation()}
        role="dialog"
        aria-modal="true"
        aria-label={title}
      >
        <div className="flex items-center justify-between mb-3">
          <h2 className="text-[length:var(--text-caption)] text-(--fg)">{title}</h2>
          <button type="button" onClick={onClose} aria-label="Close" className="text-(--text-faint) hover:text-(--fg) cursor-pointer">
            <X size={14} />
          </button>
        </div>
        {children}
        <div className="flex items-center justify-end gap-1.5 mt-4">{footer}</div>
      </div>
    </div>
  );
}

export function DialogButton({
  onClick,
  children,
  primary,
  disabled,
}: {
  onClick: () => void;
  children: ReactNode;
  primary?: boolean;
  disabled?: boolean;
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      disabled={disabled}
      className={`rounded-md px-2.5 py-1 text-[length:var(--text-pill)] transition-colors ${
        disabled ? "opacity-50 cursor-not-allowed" : "cursor-pointer"
      } ${primary ? "bg-(--fg) text-(--bg)" : "text-(--text-muted) hover:text-(--fg) hover:bg-(--card-alt)"}`}
    >
      {children}
    </button>
  );
}

export function ErrorBanner({ message, onDismiss }: { message: string; onDismiss: () => void }) {
  return (
    <div className="mb-3 flex items-center justify-between gap-2 rounded-lg border-(--line) border-[0.5px] border-solid bg-(--card) px-3 py-2 text-[length:var(--text-pill)] text-red-400">
      <span>{message}</span>
      <button onClick={onDismiss} aria-label="Dismiss error" className="shrink-0 text-(--text-faint) hover:text-(--fg) cursor-pointer">
        <X size={12} />
      </button>
    </div>
  );
}

export function slugify(s: string): string {
  return s
    .toLowerCase()
    .trim()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-+|-+$/g, "");
}
