import type { LabelColor } from "../labels/types";

export interface NoteSummary {
  id: string;
  title: string;
  pinned: boolean;
  archived: boolean;
  locked: boolean;
  labelId: string | null;
  createdAt: string;
  updatedAt: string;
}

export interface Note extends NoteSummary {
  contentHtml: string;
}

// Notepad has its own label set, independent of the todo labels.
export interface NoteLabel {
  id: string;
  name: string;
  color: LabelColor;
}
