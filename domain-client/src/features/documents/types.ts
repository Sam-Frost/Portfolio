import type { LabelColor } from "../labels/types";

export interface Folder {
  id: string;
  parentId: string | null;
  name: string;
  createdAt: string;
}

export type DocumentStatus = "pending" | "ready";

// Named DocumentItem to avoid shadowing the DOM `Document` global.
export interface DocumentItem {
  id: string;
  folderId: string | null;
  labelId: string | null;
  name: string;
  contentType: string;
  sizeBytes: number;
  status: DocumentStatus;
  createdAt: string;
  uploadedAt: string | null;
}

// The create-document response: the pending row plus the URL to PUT bytes to.
export interface CreatedDocument {
  document: DocumentItem;
  uploadUrl: string;
}

// Document Storage has its own label set, independent of the todo labels.
export interface DocumentLabel {
  id: string;
  name: string;
  color: LabelColor;
}
