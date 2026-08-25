import { Link, useRouteError } from "react-router-dom";

function messageFor(error: unknown): string {
  if (error instanceof Error) return error.message;
  return "Something went wrong.";
}

// errorElement for the router's top-level routes: catches any error thrown
// while rendering a route (or its loader/action, once those exist) that the
// route itself didn't handle, so a bug in one page shows this instead of a
// blank screen.
export function DomainErrorPage() {
  const error = useRouteError();

  return (
    <div className="min-h-screen flex items-center justify-center bg-(--bg) px-6">
      <div className="w-full max-w-sm text-center">
        <div className="bg-(--card) border-(--line) border-[0.5px] border-solid rounded-xl p-8">
          <p className="text-[length:var(--text-caption)] text-red-400 mb-1.5">Error</p>
          <h1 className="text-xl font-space font-semibold text-(--fg) mb-1.5">Something broke</h1>
          <p className="text-[length:var(--text-caption)] text-(--text-muted) mb-6 break-words">
            {messageFor(error)}
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
