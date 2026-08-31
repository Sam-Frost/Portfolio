import { useCallback, useEffect, useState } from "react";
import { fetchCmsStatus } from "./api";
import { PublishBar } from "./PublishBar";
import { ProjectsTab } from "./ProjectsTab";
import { ExperienceTab } from "./ExperienceTab";
import { WritingsTab } from "./WritingsTab";
import { SummaryTab } from "./SummaryTab";
import type { CmsStatus } from "./types";

const TABS = [
  { key: "projects", label: "Projects" },
  { key: "experiences", label: "Experience" },
  { key: "blogs", label: "Writings" },
  { key: "summary", label: "Summary" },
] as const;

type TabKey = (typeof TABS)[number]["key"];

export function CmsPage() {
  const [tab, setTab] = useState<TabKey>("projects");
  const [status, setStatus] = useState<CmsStatus | null>(null);

  const refreshStatus = useCallback(() => {
    fetchCmsStatus()
      .then(setStatus)
      .catch(() => {
        /* the publish bar just shows "checking…" if this fails */
      });
  }, []);

  useEffect(refreshStatus, [refreshStatus]);

  return (
    <div className="min-h-full lg:h-full flex flex-col">
      <div className="shrink-0">
        <PublishBar status={status} onPublished={refreshStatus} />

        <div className="mb-4 flex items-center gap-1 border-b-[0.5px] border-(--line-soft) overflow-x-auto themed-scrollbar">
          {TABS.map((t) => (
            <button
              key={t.key}
              type="button"
              onClick={() => setTab(t.key)}
              className={`-mb-px shrink-0 whitespace-nowrap border-b-2 px-3 py-2 text-[length:var(--text-caption)] transition-colors cursor-pointer ${
                tab === t.key
                  ? "border-(--fg) text-(--fg)"
                  : "border-transparent text-(--text-muted) hover:text-(--fg)"
              }`}
            >
              {t.label}
            </button>
          ))}
        </div>
      </div>

      <div className="flex-1 lg:min-h-0 lg:overflow-y-auto themed-scrollbar pb-4">
        {tab === "projects" && <ProjectsTab onChanged={refreshStatus} />}
        {tab === "experiences" && <ExperienceTab onChanged={refreshStatus} />}
        {tab === "blogs" && <WritingsTab onChanged={refreshStatus} />}
        {tab === "summary" && <SummaryTab onChanged={refreshStatus} />}
      </div>
    </div>
  );
}
