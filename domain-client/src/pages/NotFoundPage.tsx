import { Link } from "react-router-dom";

// Catch-all route (`path: "*"`) for any URL under the domain area that
// doesn't match a real route — kept outside RequireAuth/DomainLayout so a
// bad link never forces a login redirect first.
export function NotFoundPage() {
  return (
    <div className="min-h-screen flex items-center justify-center bg-(--bg) px-6">
      <div className="w-full max-w-sm text-center">
        <div className="bg-(--card) border-(--line) border-[0.5px] border-solid rounded-xl p-8">
          <p className="text-[length:var(--text-caption)] text-(--gold) mb-1.5">404</p>
          <h1 className="text-xl font-space font-semibold text-(--fg) mb-1.5">Page not found</h1>
          <p className="text-[length:var(--text-caption)] text-(--text-muted) mb-6">
            There's nothing at this address in the domain area.
          </p>
          <Link
            to="/"
            className="inline-block w-full rounded-lg bg-(--fg) text-(--bg) py-2.5 text-[length:var(--text-caption)]"
          >
            Back to domain
          </Link>
        </div>
      </div>
    </div>
  );
}
