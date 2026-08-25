export function CredentialManagerPage() {
  return (
    <div>
      <h1 className="text-xl font-space font-semibold text-(--fg) mb-1.5">Credential Manager</h1>
      <p className="text-[length:var(--text-caption)] text-(--text-muted) mb-6">Store and manage your credentials.</p>
      <div className="bg-(--card) border-(--line) border-[0.5px] border-solid rounded-xl p-8 text-center text-(--text-faint) text-[length:var(--text-caption)]">
        No credentials yet.
      </div>
    </div>
  );
}
