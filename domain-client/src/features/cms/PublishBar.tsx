import { useState } from "react";
import { UploadCloud } from "lucide-react";
import { publish } from "./api";
import type { CmsStatus } from "./types";

const SECTION_LABELS: Record<string, string> = {
  summary: "Summary",
  projects: "Projects",
  experiences: "Experience",
  blogs: "Writings",
};

function relativeTime(iso: string): string {
  const diff = Date.now() - new Date(iso).getTime();
  const mins = Math.round(diff / 60000);
  if (mins < 1) return "just now";
  if (mins < 60) return `${mins}m ago`;
  const hrs = Math.round(mins / 60);
  if (hrs < 24) return `${hrs}h ago`;
  return `${Math.round(hrs / 24)}d ago`;
}

interface Props {
  status: CmsStatus | null;
  onPublished: () => void;
}

export function PublishBar({ status, onPublished }: Props) {
  const [publishing, setPublishing] = useState(false);
  const [message, setMessage] = useState<string | null>(null);

  async function doPublish() {
    if (!window.confirm("Publish all current content to the live site?")) return;
    setPublishing(true);
    setMessage(null);
    try {
      const pub = await publish();
      setMessage(`Published v${pub.version} to the live site.`);
      onPublished();
    } catch (e) {
      setMessage(e instanceof Error ? e.message : "Publish failed.");
    } finally {
      setPublishing(false);
    }
  }

  const dirty = status?.hasUnpublishedChanges ?? false;
  const changed = (status?.changedSections ?? []).map((s) => SECTION_LABELS[s] ?? s);

  return (
    <div className="mb-4 flex flex-wrap items-center gap-x-4 gap-y-2 rounded-xl border-(--line) border-[0.5px] border-solid bg-(--card) px-4 py-3">
      <div className="min-w-0 flex-1">
        {!status ? (
          <span className="text-[length:var(--text-pill)] text-(--text-faint)">Checking publish status…</span>
        ) : status.neverPublished ? (
          <span className="text-[length:var(--text-caption)] text-(--gold)">
            Not published yet — publish to push the initial content live.
          </span>
        ) : dirty ? (
          <span className="text-[length:var(--text-caption)] text-(--gold)">
            Unpublished changes{changed.length > 0 ? ` in ${changed.join(", ")}` : ""}.
          </span>
        ) : (
          <span className="text-[length:var(--text-caption)] text-(--text-muted)">All changes published.</span>
        )}

        {status && !status.neverPublished && status.lastPublishedAt && (
          <span className="ml-2 text-[length:var(--text-pill)] text-(--text-faint)">
            Last: v{status.lastPublishVersion} · {relativeTime(status.lastPublishedAt)}
            {status.lastPublishStatus === "failed" ? " · previous attempt failed" : ""}
          </span>
        )}

        {message && <div className="mt-1 text-[length:var(--text-pill)] text-(--text-muted)">{message}</div>}
      </div>

      <button
        type="button"
        onClick={doPublish}
        disabled={publishing}
        className={`inline-flex items-center gap-1.5 rounded-lg px-3 py-1.5 text-[length:var(--text-pill)] ${
          publishing
            ? "opacity-60 cursor-not-allowed bg-(--card-alt) text-(--text-muted)"
            : dirty || status?.neverPublished
              ? "bg-(--fg) text-(--bg) cursor-pointer"
              : "bg-(--card-alt) text-(--text-muted) cursor-pointer"
        }`}
      >
        <UploadCloud size={13} />
        {publishing ? "Publishing…" : "Publish"}
      </button>
    </div>
  );
}
