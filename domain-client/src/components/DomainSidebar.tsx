import { useState } from "react";
import { NavLink } from "react-router-dom";
import { ChevronDown } from "lucide-react";
import { sections, SECTION_GROUPS, type SectionGroup } from "../data/domainSections";

const COLLAPSIBLE_GROUPS: SectionGroup[] = ["System"];

export function DomainSidebar() {
  const [collapsed, setCollapsed] = useState<Partial<Record<SectionGroup, boolean>>>({});

  return (
    <aside className="w-56 shrink-0 border-r-[0.5px] border-(--line) bg-(--card) px-3 py-6 flex flex-col gap-5">
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
                      <span className="pointer-events-none absolute left-full top-1/2 ml-2 -translate-y-1/2 z-10 whitespace-nowrap rounded-md border-(--line) border-[0.5px] border-solid bg-(--card-alt) px-2.5 py-1 text-[length:var(--text-pill)] text-(--text-muted) opacity-0 shadow-lg transition-opacity duration-150 group-hover:opacity-100">
                        Coming soon
                      </span>
                    </div>
                  )
                )}
          </div>
        );
      })}
    </aside>
  );
}
