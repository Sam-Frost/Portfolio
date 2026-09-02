// Crash-recovery store for in-flight video recordings. Every chunk the
// MediaRecorder emits is written here the instant it arrives, so a tab
// crash / refresh / power loss during recording loses at most the few
// seconds not yet flushed. On the next visit, getActiveRecording() surfaces
// an unfinished upload so it can be resumed (flush the remaining chunks as
// the final part, then complete) or discarded.
//
// IndexedDB stores Blobs directly, so the buffered tail is kept as-is.

const DB_NAME = "diary-video-recordings";
const STORE = "recordings";
const DB_VERSION = 1;

export interface RecordingRecord {
  videoId: string;
  date: string; // "YYYY-MM-DD"
  mimeType: string;
  createdAt: number; // epoch ms
  // Multipart progress: parts already uploaded + confirmed, and the number
  // to assign the next one.
  nextPartNumber: number;
  parts: Array<{ number: number; etag: string }>;
  // Recorded bytes not yet uploaded as a part.
  chunks: Blob[];
  // true once the recorder has stopped — a resume should flush + complete
  // rather than expect more footage.
  stopped: boolean;
}

function openDB(): Promise<IDBDatabase> {
  return new Promise((resolve, reject) => {
    const req = indexedDB.open(DB_NAME, DB_VERSION);
    req.onupgradeneeded = () => {
      if (!req.result.objectStoreNames.contains(STORE)) {
        req.result.createObjectStore(STORE, { keyPath: "videoId" });
      }
    };
    req.onsuccess = () => resolve(req.result);
    req.onerror = () => reject(req.error);
  });
}

function tx<T>(mode: IDBTransactionMode, run: (store: IDBObjectStore) => IDBRequest<T>): Promise<T> {
  return openDB().then(
    (db) =>
      new Promise<T>((resolve, reject) => {
        const t = db.transaction(STORE, mode);
        const req = run(t.objectStore(STORE));
        t.oncomplete = () => {
          resolve(req.result);
          db.close();
        };
        t.onerror = () => {
          reject(t.error);
          db.close();
        };
      }),
  );
}

export function saveRecording(rec: RecordingRecord): Promise<IDBValidKey> {
  return tx("readwrite", (s) => s.put(rec));
}

export function deleteRecording(videoId: string): Promise<undefined> {
  return tx("readwrite", (s) => s.delete(videoId) as IDBRequest<undefined>);
}

// The most recent recording that never finished uploading, if any.
export async function getActiveRecording(): Promise<RecordingRecord | null> {
  const all = await tx<RecordingRecord[]>("readonly", (s) => s.getAll() as IDBRequest<RecordingRecord[]>);
  const unfinished = all
    .filter((r) => !r.stopped || r.chunks.length > 0 || r.parts.length === 0)
    .sort((a, b) => b.createdAt - a.createdAt);
  // A record that is `stopped` with 0 chunks and >=1 part was mid-`complete`
  // when the page died — still worth resuming (complete is idempotent).
  return unfinished[0] ?? all.filter((r) => r.stopped).sort((a, b) => b.createdAt - a.createdAt)[0] ?? null;
}

// Best-effort: clear everything older than a day so abandoned records don't
// pile up (the server also sweeps orphaned S3 uploads via a lifecycle rule).
export async function pruneStaleRecordings(): Promise<void> {
  try {
    const all = await tx<RecordingRecord[]>("readonly", (s) => s.getAll() as IDBRequest<RecordingRecord[]>);
    const cutoff = Date.now() - 24 * 60 * 60 * 1000;
    await Promise.all(all.filter((r) => r.createdAt < cutoff).map((r) => deleteRecording(r.videoId)));
  } catch {
    // ignore — pruning is not essential
  }
}
