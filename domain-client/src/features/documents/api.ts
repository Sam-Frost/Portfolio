import { apiRequest } from "../../lib/apiClient";
import type { CreatedDocument, DocumentItem, Folder } from "./types";

// --- folders ---

export function fetchFolders(): Promise<Folder[]> {
  return apiRequest<Folder[]>("/api/documents/folders");
}

export function createFolder(input: { name: string; parentId: string | null }): Promise<Folder> {
  return apiRequest<Folder>("/api/documents/folders", {
    method: "POST",
    body: JSON.stringify(input),
  });
}

// PATCH is a partial update: send "" for parentId to move to the root.
export function updateFolder(
  id: string,
  input: { name?: string; parentId?: string | null },
): Promise<Folder> {
  return apiRequest<Folder>(`/api/documents/folders/${id}`, {
    method: "PATCH",
    body: JSON.stringify({
      ...(input.name !== undefined ? { name: input.name } : {}),
      ...(input.parentId !== undefined ? { parentId: input.parentId ?? "" } : {}),
    }),
  });
}

export function deleteFolder(id: string): Promise<void> {
  return apiRequest<void>(`/api/documents/folders/${id}`, { method: "DELETE" });
}

// --- documents ---

export function fetchDocuments(params: {
  folderId: string | null;
  labelId?: string | null;
  q?: string;
}): Promise<DocumentItem[]> {
  const query = new URLSearchParams();
  if (params.q) query.set("q", params.q);
  else if (params.folderId) query.set("folderId", params.folderId);
  if (params.labelId) query.set("labelId", params.labelId);
  const qs = query.toString();
  return apiRequest<DocumentItem[]>(`/api/documents${qs ? `?${qs}` : ""}`);
}

export function createDocument(input: {
  name: string;
  folderId: string | null;
  contentType: string;
  sizeBytes: number;
}): Promise<CreatedDocument> {
  return apiRequest<CreatedDocument>("/api/documents", {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export function completeUpload(id: string): Promise<DocumentItem> {
  return apiRequest<DocumentItem>(`/api/documents/${id}/complete`, { method: "POST" });
}

export function getDownloadUrl(id: string): Promise<{ url: string }> {
  return apiRequest<{ url: string }>(`/api/documents/${id}/download`);
}

export function updateDocument(
  id: string,
  input: { name?: string; folderId?: string | null; labelId?: string | null },
): Promise<DocumentItem> {
  return apiRequest<DocumentItem>(`/api/documents/${id}`, {
    method: "PATCH",
    body: JSON.stringify({
      ...(input.name !== undefined ? { name: input.name } : {}),
      ...(input.folderId !== undefined ? { folderId: input.folderId ?? "" } : {}),
      ...(input.labelId !== undefined ? { labelId: input.labelId ?? "" } : {}),
    }),
  });
}

export function deleteDocument(id: string): Promise<void> {
  return apiRequest<void>(`/api/documents/${id}`, { method: "DELETE" });
}

// uploadToUrl PUTs the raw file bytes straight to the blob store (S3 or the
// local-disk store), bypassing apiRequest — the URL is pre-signed, so it
// carries no bearer token and must not get the JSON Content-Type header.
// XMLHttpRequest (not fetch) is used for the upload-progress events.
export function uploadToUrl(url: string, file: File, onProgress?: (fraction: number) => void): Promise<void> {
  return new Promise((resolve, reject) => {
    const xhr = new XMLHttpRequest();
    xhr.open("PUT", url);
    xhr.setRequestHeader("Content-Type", file.type || "application/octet-stream");

    xhr.upload.onprogress = (e) => {
      if (e.lengthComputable && onProgress) onProgress(e.loaded / e.total);
    };
    xhr.onload = () => {
      if (xhr.status >= 200 && xhr.status < 300) resolve();
      else reject(new Error(`Upload failed (${xhr.status})`));
    };
    xhr.onerror = () => reject(new Error("Upload failed. Check your connection and try again."));
    xhr.send(file);
  });
}
