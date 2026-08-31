import { useEffect, useRef, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { ArrowLeft, Lock } from "lucide-react";
import { fetchDiaryEntry, upsertDiaryEntry } from "./api";
import { RichTextEditor } from "../../components/domain/RichTextEditor";
import { ApiError } from "../../lib/apiClient";
import { isDiaryDateLocked } from "./dateUtils";
import type { DiaryEntry } from "./types";

type SaveStatus = "saved" | "saving" | "error";

const AUTOSAVE_DELAY_MS = 700;

const STATUS_COPY: Record<SaveStatus, { label: string; dotClass: string }> = {
  saved: { label: "All changes saved", dotClass: "bg-(--green)" },
  saving: { label: "Saving...", dotClass: "bg-(--label-orange)" },
  error: { label: "Couldn't save", dotClass: "bg-(--label-red)" },
};

// date is "YYYY-MM-DD"; built as a local Date at noon (not midnight) so
// formatting can't roll over to the adjacent day under a negative UTC
// offset browser timezone.
function formatDateLabel(date: string) {
  const [y, m, d] = date.split("-").map(Number);
  return new Date(y, m - 1, d, 12).toLocaleDateString(undefined, {
    weekday: "long",
    year: "numeric",
    month: "long",
    day: "numeric",
  });
}

export function DiaryEntryPage() {
  const { date } = useParams<{ date: string }>();
  const navigate = useNavigate();

  const [entry, setEntry] = useState<DiaryEntry | null>(null);
  const [exists, setExists] = useState(false);
  const [loading, setLoading] = useState(true);
  const [loadError, setLoadError] = useState<string | null>(null);
  const [status, setStatus] = useState<SaveStatus>("saved");
  // Flipped true if a save is rejected with 409 mid-session (the grace
  // window closed while the page was open) — the client-side lock check
  // below is only a pre-check to avoid a wasted round trip; this is what
  // makes the UI actually respect the server's authoritative rejection.
  const [rejectedAsLocked, setRejectedAsLocked] = useState(false);

  const pendingContentRef = useRef<string | null>(null);
  const timeoutRef = useRef<number | null>(null);

  useEffect(() => {
    if (!date) return;
    setLoading(true);
    setLoadError(null);
    setRejectedAsLocked(false);
    fetchDiaryEntry(date)
      .then((e) => {
        setEntry(e);
        setExists(true);
      })
      .catch((err) => {
        if (err instanceof ApiError && err.status === 404) {
          setEntry(null);
          setExists(false);
        } else {
          setLoadError(err instanceof Error ? err.message : "Couldn't load this entry.");
        }
      })
      .finally(() => setLoading(false));
  }, [date]);

  function flush() {
    if (timeoutRef.current) window.clearTimeout(timeoutRef.current);
    timeoutRef.current = null;
    if (!date || pendingContentRef.current === null) return;
    const content = pendingContentRef.current;
    pendingContentRef.current = null;
    upsertDiaryEntry(date, content).catch(() => {});
  }

  // Best-effort save of whatever's still pending when leaving the page, same
  // as notepad's NoteEditorPage.
  useEffect(() => {
    return () => flush();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [date]);

  function handleContentChange(html: string) {
    if (!date) return;
    pendingContentRef.current = html;
    setStatus("saving");
    if (timeoutRef.current) window.clearTimeout(timeoutRef.current);
    timeoutRef.current = window.setTimeout(() => {
      const content = pendingContentRef.current ?? "";
      pendingContentRef.current = null;
      upsertDiaryEntry(date, content)
        .then((e) => {
          setEntry(e);
          setExists(true);
          setStatus("saved");
        })
        .catch((err) => {
          setStatus("error");
          if (err instanceof ApiError && err.status === 409) {
            setRejectedAsLocked(true);
          }
        });
    }, AUTOSAVE_DELAY_MS);
  }

  if (!date) return null;

  if (loading) {
    return <div className="text-(--text-faint) text-[length:var(--text-caption)]">Loading entry...</div>;
  }

  if (loadError && !entry) {
    return <div className="text-(--text-faint) text-[length:var(--text-caption)]">{loadError}</div>;
  }

  // Trust the server's own Locked flag once we have one; otherwise (a day
  // with no entry yet) fall back to the same pure client-side check the
  // calendar grid uses. rejectedAsLocked forces this true the moment a save
  // is actually rejected, even if the pre-check said it was still open.
  const locked = rejectedAsLocked || (entry ? entry.locked : isDiaryDateLocked(date));

  const { label, dotClass } = STATUS_COPY[status];

  return (
    <div className="min-h-full lg:h-full flex flex-col">
      <div className="shrink-0 flex items-center justify-between gap-3 flex-wrap mb-4">
        <button
          onClick={() => navigate("/diary")}
          className="flex items-center gap-1.5 text-[length:var(--text-caption)] text-(--text-muted) hover:text-(--fg) transition-colors cursor-pointer"
        >
          <ArrowLeft size={14} />
          Calendar
        </button>

        {locked ? (
          <div className="flex items-center gap-1.5 text-[length:var(--text-pill)] text-(--text-faint)">
            <Lock size={12} />
            Read-only — edit window closed
          </div>
        ) : (
          <div className="flex items-center gap-1.5 text-[length:var(--text-pill)] text-(--text-faint)">
            <span className={`size-2 rounded-full shrink-0 ${dotClass}`} />
            {label}
          </div>
        )}
      </div>

      <h1 className="shrink-0 mb-3 text-xl font-space font-semibold text-(--fg)">{formatDateLabel(date)}</h1>

      {rejectedAsLocked && (
        <div className="shrink-0 mb-3 rounded-lg border-(--line) border-[0.5px] border-solid bg-(--card) px-3 py-2 text-[length:var(--text-pill)] text-(--label-red)">
          This day's edit window closed while you were writing — your last change wasn't saved.
        </div>
      )}

      {!exists && locked ? (
        <div className="flex-1 min-h-[40vh] lg:min-h-0 flex items-center justify-center rounded-lg border-(--line) border-dashed border-[1px] text-(--text-faint) text-[length:var(--text-caption)] text-center px-6">
          No entry was written for this day, and its edit window has closed.
        </div>
      ) : (
        <RichTextEditor key={date} initialContentHtml={entry?.content ?? ""} onChange={handleContentChange} readOnly={locked} />
      )}
    </div>
  );
}
