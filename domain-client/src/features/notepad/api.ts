import { apiRequest } from "../../lib/apiClient";
import type { Note, NoteSummary } from "./types";

export function fetchNotes(archived = false): Promise<NoteSummary[]> {
  return apiRequest<NoteSummary[]>(`/api/notes${archived ? "?archived=true" : ""}`);
}

export function fetchNote(id: string): Promise<Note> {
  return apiRequest<Note>(`/api/notes/${id}`);
}

// The singleton "Random Notepad" scratch buffer — one always-there,
// title-less note the server get-or-creates on first open. It never shows up
// in fetchNotes(); once you have it, edit it with updateNote() like any note.
export function fetchScratchNote(): Promise<Note> {
  return apiRequest<Note>("/api/notes/scratch");
}

export function createNote(title?: string): Promise<Note> {
  return apiRequest<Note>("/api/notes", {
    method: "POST",
    body: JSON.stringify({ title: title ?? null }),
  });
}

// Partial update — only the fields present are changed, matching the
// backend's UpdateInput convention. Autosave calls this with whichever of
// title/contentHtml just changed.
export function updateNote(
  id: string,
  patch: {
    title?: string;
    contentHtml?: string;
    pinned?: boolean;
    archived?: boolean;
    locked?: boolean;
    // "" clears the label, matching the backend's UpdateInput convention.
    labelId?: string;
  },
): Promise<Note> {
  return apiRequest<Note>(`/api/notes/${id}`, {
    method: "PATCH",
    body: JSON.stringify(patch),
  });
}

export function deleteNote(id: string): Promise<void> {
  return apiRequest<void>(`/api/notes/${id}`, { method: "DELETE" });
}
