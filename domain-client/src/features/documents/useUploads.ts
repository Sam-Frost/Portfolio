import { useState } from "react";
import { completeUpload, createDocument, uploadToUrl } from "./api";
import type { DocumentItem } from "./types";

export interface Upload {
  id: string; // client-side id (also the document id once created)
  name: string;
  progress: number; // 0..1
  status: "uploading" | "done" | "error";
  error?: string;
}

let counter = 0;

// useUploads runs the 3-step browser-direct upload for each file
// (create pending row -> PUT bytes to the presigned URL -> complete) and
// tracks per-file progress. onComplete fires once per finished document so
// the page can slot it into its list.
export function useUploads(onComplete: (doc: DocumentItem) => void) {
  const [uploads, setUploads] = useState<Upload[]>([]);

  function patch(id: string, next: Partial<Upload>) {
    setUploads((prev) => prev.map((u) => (u.id === id ? { ...u, ...next } : u)));
  }

  function clearFinished() {
    setUploads((prev) => prev.filter((u) => u.status === "uploading"));
  }

  async function startUpload(file: File, folderId: string | null) {
    const tempId = `tmp-${counter++}`;
    setUploads((prev) => [...prev, { id: tempId, name: file.name, progress: 0, status: "uploading" }]);

    try {
      const { document, uploadUrl } = await createDocument({
        name: file.name,
        folderId,
        contentType: file.type || "application/octet-stream",
        sizeBytes: file.size,
      });
      patch(tempId, { id: document.id });

      await uploadToUrl(uploadUrl, file, (fraction) => patch(document.id, { progress: fraction }));

      const ready = await completeUpload(document.id);
      patch(document.id, { progress: 1, status: "done" });
      onComplete(ready);
    } catch (err) {
      patch(tempId, { status: "error", error: err instanceof Error ? err.message : "Upload failed." });
    }
  }

  function startUploads(files: FileList | File[], folderId: string | null) {
    for (const file of Array.from(files)) void startUpload(file, folderId);
  }

  return { uploads, startUploads, clearFinished };
}
