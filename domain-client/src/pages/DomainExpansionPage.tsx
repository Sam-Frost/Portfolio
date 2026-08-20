import { useState } from "react";
import { useNavigate } from "react-router-dom";
import Button from "../components/Button";
import { PUBLIC_SITE_URL } from "../config";

export function DomainExpansionPage() {
  const navigate = useNavigate();
  const [password, setPassword] = useState("");

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    navigate("/dashboard");
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-(--bg) px-6">
      <div className="w-full max-w-sm">
        <button
          className="mb-6 text-[length:var(--text-caption)] text-(--text-muted) hover:text-(--fg) cursor-pointer"
          onClick={() => { window.location.href = PUBLIC_SITE_URL; }}
        >
          ← Back to site
        </button>

        <div className="bg-(--card) border-(--line) border-[0.5px] border-solid rounded-xl p-8">
          <div className="h-10 w-10 rounded-lg bg-(--card-alt) border-(--line) border-[0.5px] border-solid flex items-center justify-center mb-5">
            <LockIcon />
          </div>

          <h1 className="text-xl font-space font-semibold text-(--fg) mb-1.5">Domain Access</h1>
          <p className="text-[length:var(--text-caption)] text-(--text-muted) mb-6">Enter the password to enter domain.</p>

          <form className="flex flex-col gap-4" onSubmit={handleSubmit}>
            <Field label="Password" type="password" value={password} onChange={setPassword} placeholder="••••••••" />

            <Button styling="w-full py-2.5 mt-2" text="Access" />
          </form>
        </div>

        <p className="mt-4 text-center text-[length:var(--text-pill)] text-(--text-faint)">Restricted area · authorized access only</p>
      </div>
    </div>
  );
}

function Field({ label, type, value, onChange, placeholder }: {
  label: string
  type: string
  value: string
  onChange: (value: string) => void
  placeholder?: string
}) {
  return (
    <label className="flex flex-col gap-1.5">
      <span className="text-[length:var(--text-caption)] text-(--text-muted)">{label}</span>
      <input
        type={type}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        placeholder={placeholder}
        className="bg-(--card-alt) border-(--line) border-[0.5px] border-solid rounded-lg px-3 py-2 text-sm text-(--fg) placeholder:text-(--text-faint) outline-none focus:border-(--gold) transition-colors"
      />
    </label>
  );
}

function LockIcon() {
  return (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="var(--gold)" strokeWidth={2} strokeLinecap="round" strokeLinejoin="round">
      <rect width="18" height="11" x="3" y="11" rx="2" />
      <path d="M7 11V7a5 5 0 0 1 10 0v4" />
    </svg>
  );
}
