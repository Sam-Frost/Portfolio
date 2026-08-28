import { fallbackContent } from "./fallback";
import type { SiteContent } from "./types";

// Where the published content.json lives. Same-origin by default (served
// from the site's own CloudFront origin); override with VITE_CONTENT_URL
// for local dev pointed at a running server or a staging bucket.
const CONTENT_URL = import.meta.env.VITE_CONTENT_URL ?? "/content.json";

function looksValid(data: unknown): data is SiteContent {
  if (typeof data !== "object" || data === null) return false;
  const c = data as Record<string, unknown>;
  return (
    typeof c.summary === "object" &&
    c.summary !== null &&
    Array.isArray(c.projects) &&
    Array.isArray(c.experiences) &&
    Array.isArray(c.blogs)
  );
}

// loadContent fetches the published content, falling back to the compiled-in
// copy on any failure (network error, non-200, malformed body) so the site
// always renders.
export async function loadContent(): Promise<SiteContent> {
  try {
    const res = await fetch(CONTENT_URL, { cache: "no-cache" });
    if (!res.ok) return fallbackContent;
    const data: unknown = await res.json();
    return looksValid(data) ? data : fallbackContent;
  } catch {
    return fallbackContent;
  }
}
