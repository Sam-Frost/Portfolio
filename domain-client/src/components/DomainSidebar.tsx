import { NavLink } from "react-router-dom";
import { KeyRound, FolderLock, ListTodo, CalendarCheck, LayoutTemplate, type LucideIcon } from "lucide-react";

type Section =
  | { label: string; icon: LucideIcon; enabled: true; path: string }
  | { label: string; icon: LucideIcon; enabled: false };

const sections: Section[] = [
  { label: "Credential Manager", icon: KeyRound, enabled: true, path: "credentials" },
  { label: "Document Storage", icon: FolderLock, enabled: false },
  { label: "Todos", icon: ListTodo, enabled: false },
  { label: "Daily Work Tracker", icon: CalendarCheck, enabled: false },
  { label: "CMS", icon: LayoutTemplate, enabled: false },
];

export function DomainSidebar() {
  return (
    <aside className="w-56 shrink-0 border-r-[0.5px] border-(--line) bg-(--card) px-3 py-6 flex flex-col gap-1">
      {sections.map((section) =>
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
    </aside>
  );
}
