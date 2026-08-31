import { useState } from "react";
import { NavLink } from "react-router-dom";
import { ChevronDown } from "lucide-react";
import { sections, SECTION_GROUPS, type SectionGroup } from "../data/domainSections";
import { SpotifyWidget } from "../features/spotify/SpotifyWidget";

const COLLAPSIBLE_GROUPS: SectionGroup[] = ["System"];

interface DomainSidebarProps {
  /** Drawer open state — only relevant below the `lg` breakpoint, where the
      sidebar is an overlay. At `lg+` the sidebar is always visible. */
  open?: boolean;
  onClose?: () => void;
}

export function DomainSidebar({ open = false, onClose }: DomainSidebarProps) {
  const [collapsed, setCollapsed] = useState<Partial<Record<SectionGroup, boolean>>>({ System: true });

  return (
    <>
      {/* Backdrop — mobile/tablet drawer only. */}
      <div
        className={`fixed inset-0 z-40 bg-black/50 transition-opacity lg:hidden ${
          open ? "opacity-100" : "pointer-events-none opacity-0"
        }`}
        onClick={onClose}
        aria-hidden="true"
      />

      <aside
        className={`fixed inset-y-0 left-0 z-50 flex w-56 shrink-0 flex-col gap-5 border-r-[0.5px] border-(--line) bg-(--card) px-3 py-6 transition-transform duration-200 ease-out lg:static lg:z-auto lg:translate-x-0 ${
          open ? "translate-x-0" : "-translate-x-full"
        }`}
      >
        <div className="flex-1 min-h-0 overflow-y-auto overflow-x-hidden themed-scrollbar flex flex-col gap-5">
          {SECTION_GROUPS.map((group) => {
            const collapsible = COLLAPSIBLE_GROUPS.includes(group);
            const isCollapsed = collapsible && collapsed[group];

            return (
              <div key={group} className="flex flex-col gap-1">
                {collapsible ? (
                  <button
                    type="button"
                    onClick={() => setCollapsed((prev) => ({ ...prev, [group]: !prev[group] }))}
                    aria-expanded={!isCollapsed}
                    className="flex items-center justify-between px-3 mb-1 text-[length:var(--text-pill)] font-medium uppercase tracking-wide text-(--text-faint) hover:text-(--text-muted) transition-colors cursor-pointer"
                  >
                    {group}
                    <ChevronDown size={12} className={`transition-transform ${isCollapsed ? "-rotate-90" : ""}`} />
                  </button>
                ) : (
                  <span className="px-3 mb-1 text-[length:var(--text-pill)] font-medium uppercase tracking-wide text-(--text-faint)">
                    {group}
                  </span>
                )}

                {!isCollapsed &&
                  sections
                    .filter((section) => section.group === group)
                    .map((section) =>
                      section.enabled ? (
                        <NavLink
                          key={section.label}
                          to={section.path}
                          onClick={onClose}
                          className={({ isActive }) =>
                            `flex items-center gap-2.5 rounded-lg px-3 py-2 text-[length:var(--text-caption)] transition-colors ${
                              isActive
                                ? "bg-(--card-alt) text-(--fg)"
                                : "text-(--text-muted) hover:text-(--fg) hover:bg-(--card-alt)"
                            }`
                          }
                        >
                          <section.icon size={16} />
                          {section.label}
                        </NavLink>
                      ) : (
                        <div key={section.label} className="group relative">
                          <div className="flex items-center gap-2.5 rounded-lg px-3 py-2 text-[length:var(--text-caption)] text-(--text-faint) cursor-not-allowed select-none">
                            <section.icon size={16} />
                            {section.label}
                          </div>
                          <span className="pointer-events-none absolute left-3 top-full mt-1 z-10 whitespace-nowrap rounded-md border-(--line) border-[0.5px] border-solid bg-(--card-alt) px-2.5 py-1 text-[length:var(--text-pill)] text-(--text-muted) opacity-0 shadow-lg transition-opacity duration-150 group-hover:opacity-100">
                            Coming soon
                          </span>
                        </div>
                      )
                    )}
              </div>
            );
          })}
        </div>

        <div className="shrink-0">
          <SpotifyWidget />
        </div>
      </aside>
    </>
  );
}
