import { completeVideoUpload, getUploadPartUrl } from "./videoApi";
import { deleteRecording, saveRecording, type RecordingRecord } from "./idb";
import type { DiaryVideo } from "./videoTypes";

// S3 requires every multipart part except the last to be at least 5 MiB.
// We cut a part once the buffered tail reaches PART_TARGET (comfortably
// above the minimum); the remainder at stop becomes the final any-size part.
const PART_TARGET = 8 * 1024 * 1024;
const MAX_PART_ATTEMPTS = 4;

export function bufferedBytes(chunks: Blob[]): number {
  return chunks.reduce((n, c) => n + c.size, 0);
}

function sleep(ms: number): Promise<void> {
  return new Promise((r) => setTimeout(r, ms));
}

// Uploads `blob` as rec's next part, retrying transient failures. Mutates
// rec (parts, nextPartNumber) and persists it on success so a crash right
// after can't re-upload the same part under a new number.
async function uploadPart(rec: RecordingRecord, blob: Blob): Promise<void> {
  const partNumber = rec.nextPartNumber;
  let lastErr: unknown;
  for (let attempt = 1; attempt <= MAX_PART_ATTEMPTS; attempt++) {
    try {
      const url = await getUploadPartUrl(rec.videoId, partNumber);
      const res = await fetch(url, { method: "PUT", body: blob });
      if (!res.ok) throw new Error(`part ${partNumber}: HTTP ${res.status}`);
      const etag = res.headers.get("ETag");
      if (!etag) {
        throw new Error("part response had no ETag header — the blob store's CORS config must expose ETag");
      }
      rec.parts.push({ number: partNumber, etag: etag.replace(/"/g, "") });
      rec.nextPartNumber = partNumber + 1;
      await saveRecording(rec);
      return;
    } catch (err) {
      lastErr = err;
      if (attempt < MAX_PART_ATTEMPTS) await sleep(500 * attempt);
    }
  }
  throw lastErr instanceof Error ? lastErr : new Error("part upload failed");
}

// Drain full parts out of the buffered tail. Called after each recorded
// chunk while recording. `force` (used at stop) also uploads a
// smaller-than-target remainder — but never an empty or sub-5MiB blob
// unless it's the only part.
export async function drain(rec: RecordingRecord, force: boolean): Promise<void> {
  while (bufferedBytes(rec.chunks) >= PART_TARGET) {
    const blob = new Blob(rec.chunks, { type: rec.mimeType });
    rec.chunks = [];
    await uploadPart(rec, blob);
  }
  if (force && bufferedBytes(rec.chunks) > 0) {
    // At stop the remainder becomes the final part — S3 allows the last
    // part to be smaller than the 5 MiB minimum.
    const blob = new Blob(rec.chunks, { type: rec.mimeType });
    rec.chunks = [];
    await uploadPart(rec, blob);
  }
}

// Finish an upload: flush whatever's left, then tell the server to assemble
// the parts. Clears the crash-recovery record on success.
export async function finalize(rec: RecordingRecord, durationSeconds: number | null): Promise<DiaryVideo> {
  rec.stopped = true;
  await saveRecording(rec);
  await drain(rec, true);
  if (rec.parts.length === 0) {
    throw new Error("nothing was recorded");
  }
  const video = await completeVideoUpload(rec.videoId, rec.parts, durationSeconds);
  await deleteRecording(rec.videoId);
  return video;
}
