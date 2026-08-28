import { apiRequest } from "../../lib/apiClient";
import type { LabelColor } from "../labels/types";
import type { DocumentLabel } from "./types";

export function fetchDocumentLabels(): Promise<DocumentLabel[]> {
  return apiRequest<DocumentLabel[]>("/api/document-labels");
}

export function createDocumentLabel(input: { name: string; color: LabelColor }): Promise<DocumentLabel> {
  return apiRequest<DocumentLabel>("/api/document-labels", {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export function updateDocumentLabel(
  id: string,
  input: { name: string; color: LabelColor },
): Promise<DocumentLabel> {
  return apiRequest<DocumentLabel>(`/api/document-labels/${id}`, {
    method: "PATCH",
    body: JSON.stringify(input),
  });
}

export function deleteDocumentLabel(id: string): Promise<void> {
  return apiRequest<void>(`/api/document-labels/${id}`, { method: "DELETE" });
}
