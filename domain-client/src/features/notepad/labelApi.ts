import { apiRequest } from "../../lib/apiClient";
import type { LabelColor } from "../labels/types";
import type { NoteLabel } from "./types";

export function fetchNoteLabels(): Promise<NoteLabel[]> {
  return apiRequest<NoteLabel[]>("/api/notepad-labels");
}

export function createNoteLabel(input: { name: string; color: LabelColor }): Promise<NoteLabel> {
  return apiRequest<NoteLabel>("/api/notepad-labels", {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export function updateNoteLabel(id: string, input: { name: string; color: LabelColor }): Promise<NoteLabel> {
  return apiRequest<NoteLabel>(`/api/notepad-labels/${id}`, {
    method: "PATCH",
    body: JSON.stringify(input),
  });
}

export function deleteNoteLabel(id: string): Promise<void> {
  return apiRequest<void>(`/api/notepad-labels/${id}`, { method: "DELETE" });
}
