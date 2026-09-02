import {
  // KeyRound,
  FolderLock,
  ListTodo,
  // CalendarCheck, // unused while Consistency / Daily Work Tracker are disabled — see below
  LayoutTemplate,
  // Sparkles,
  BookOpen,
  PenTool,
  Clock,
  Dumbbell,
  StickyNote,
  // Ban, // unused while Bad Habit Tracker is disabled — see below
  Settings,
  GraduationCap,
  type LucideIcon,
} from "lucide-react";

export const SECTION_GROUPS = ["Productivity", "Personal", "System"] as const;
export type SectionGroup = (typeof SECTION_GROUPS)[number];

export type Section =
  | { label: string; icon: LucideIcon; enabled: true; path: string; group: SectionGroup }
  | { label: string; icon: LucideIcon; enabled: false; group: SectionGroup };

// The set of gated feature areas. Shared between DomainSidebar (nav)
// and SettingsPage (per-section settings headings) so the two stay in sync.
// `group` controls which sidebar section (see SECTION_GROUPS) an item sits in.
export const sections: Section[] = [
  { label: "Todos", icon: ListTodo, enabled: true, path: "todos", group: "Productivity" },
  { label: "Notepad", icon: StickyNote, enabled: true, path: "notepad", group: "Productivity" },
  { label: "Sessions", icon: Clock, enabled: true, path: "hourly-tracker", group: "Productivity" },
  // Disabled for now — uncomment to bring back.
  // { label: "Daily Work Tracker", icon: CalendarCheck, enabled: false, group: "Productivity" },
  // { label: "Bad Habit Tracker", icon: Ban, enabled: false, group: "Productivity" },

  { label: "Personal Diary", icon: BookOpen, enabled: true, path: "diary", group: "Personal" },
  { label: "Upskill", icon: GraduationCap, enabled: true, path: "upskill", group: "Personal" },
  { label: "Fitness", icon: Dumbbell, enabled: true, path: "fitness", group: "Personal" },
  { label: "Drawing Board", icon: PenTool, enabled: true, path: "drawing-board", group: "Personal" },

  // { label: "Claude Skills", icon: Sparkles, enabled: false, group: "System" },
  { label: "CMS", icon: LayoutTemplate, enabled: true, path: "cms", group: "System" },
  // { label: "Credential Manager", icon: KeyRound, enabled: true, path: "credentials", group: "System" },
  { label: "Documents", icon: FolderLock, enabled: true, path: "documents", group: "System" },
  { label: "Settings", icon: Settings, enabled: true, path: "settings", group: "System" },
];
