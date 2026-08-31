export interface NoteSummary {
  id: string;
  title: string;
  pinned: boolean;
  archived: boolean;
  locked: boolean;
  createdAt: string;
  updatedAt: string;
}

export interface Note extends NoteSummary {
  contentHtml: string;
}
