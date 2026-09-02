import { useCallback, useEffect, useRef, useState } from "react";
import { ApiError } from "../../../lib/apiClient";
import { createVideoUpload, deleteVideo } from "./videoApi";
import {
  deleteRecording,
  getActiveRecording,
  pruneStaleRecordings,
  saveRecording,
  type RecordingRecord,
} from "./idb";
import { drain, finalize } from "./uploader";
import type { DiaryVideo } from "./videoTypes";

export type RecorderPhase =
  | "idle" // no camera yet
  | "preview" // camera live, not recording
  | "recording"
  | "finalizing" // recorder stopped, uploading the tail + completing
  | "done"
  | "error";

// MediaRecorder writes WebM on Chrome/Firefox and MP4 on Safari; recent
// Chrome supports MP4 too. Fragmented MP4 is natively seekable, so we
// prefer it — no post-processing needed to stream it back while / right
// after recording.
const MIME_CANDIDATES = [
  "video/mp4;codecs=avc1.42E01E,mp4a.40.2",
  "video/mp4",
  "video/webm;codecs=vp9,opus",
  "video/webm;codecs=vp8,opus",
  "video/webm",
];

function pickMimeType(): string {
  if (typeof MediaRecorder === "undefined") return "";
  for (const c of MIME_CANDIDATES) {
    if (MediaRecorder.isTypeSupported(c)) return c;
  }
  return "";
}

function baseContentType(mimeType: string): "video/mp4" | "video/webm" {
  return mimeType.startsWith("video/mp4") ? "video/mp4" : "video/webm";
}

function sleep(ms: number): Promise<void> {
  return new Promise((r) => setTimeout(r, ms));
}

function errMessage(err: unknown, fallback: string): string {
  if (err instanceof ApiError || err instanceof Error) return err.message;
  return fallback;
}

interface UseVideoRecorder {
  phase: RecorderPhase;
  error: string | null;
  elapsedSec: number;
  /** 0–1, rough: bytes confirmed in S3 / bytes recorded. */
  uploadProgress: number;
  /** A recording from a previous session that never finished uploading. */
  resumable: RecordingRecord | null;
  result: DiaryVideo | null;
  /** Live camera stream for the <video> preview element. */
  previewStream: MediaStream | null;

  beginPreview: () => Promise<void>;
  startRecording: () => Promise<void>;
  stopRecording: () => void;
  /** Abort: stop the recorder, tell the server to discard the upload. */
  cancel: () => Promise<void>;
  resumeUpload: () => Promise<void>;
  discardResumable: () => Promise<void>;
}

export function useVideoRecorder(date: string): UseVideoRecorder {
  const [phase, setPhase] = useState<RecorderPhase>("idle");
  const [error, setError] = useState<string | null>(null);
  const [elapsedSec, setElapsedSec] = useState(0);
  const [uploadProgress, setUploadProgress] = useState(0);
  const [resumable, setResumable] = useState<RecordingRecord | null>(null);
  const [result, setResult] = useState<DiaryVideo | null>(null);
  const [previewStream, setPreviewStream] = useState<MediaStream | null>(null);

  const streamRef = useRef<MediaStream | null>(null);
  const recorderRef = useRef<MediaRecorder | null>(null);
  const recordRef = useRef<RecordingRecord | null>(null);
  const drainingRef = useRef(false);
  const startedAtRef = useRef(0);
  const timerRef = useRef<number | null>(null);
  const recordedBytesRef = useRef(0);

  useEffect(() => {
    void pruneStaleRecordings();
    void getActiveRecording().then((rec) => {
      if (rec) setResumable(rec);
    });
  }, []);

  const stopStream = useCallback(() => {
    streamRef.current?.getTracks().forEach((t) => t.stop());
    streamRef.current = null;
    setPreviewStream(null);
  }, []);

  const stopTimer = useCallback(() => {
    if (timerRef.current !== null) window.clearInterval(timerRef.current);
    timerRef.current = null;
  }, []);

  useEffect(() => {
    return () => {
      stopTimer();
      streamRef.current?.getTracks().forEach((t) => t.stop());
      try {
        if (recorderRef.current?.state === "recording") recorderRef.current.stop();
      } catch {
        // recorder already gone
      }
    };
  }, [stopTimer]);

  const beginPreview = useCallback(async () => {
    setError(null);
    try {
      const stream = await navigator.mediaDevices.getUserMedia({
        video: { width: { ideal: 1280 }, height: { ideal: 720 } },
        audio: true,
      });
      streamRef.current = stream;
      setPreviewStream(stream);
      setPhase("preview");
    } catch (err) {
      setError(
        err instanceof DOMException && (err.name === "NotAllowedError" || err.name === "SecurityError")
          ? "Camera / microphone access was blocked. Allow it in your browser and try again."
          : errMessage(err, "Couldn't start the camera."),
      );
      setPhase("error");
    }
  }, []);

  const handleChunk = useCallback(async (data: Blob) => {
    const rec = recordRef.current;
    if (!rec || data.size === 0) return;
    rec.chunks.push(data);
    recordedBytesRef.current += data.size;
    // Durability first: the chunk is safe on disk before we try to upload.
    try {
      await saveRecording(rec);
    } catch {
      // IndexedDB unavailable (private window, quota) — recording still
      // works, just without crash recovery.
    }
    if (drainingRef.current) return;
    drainingRef.current = true;
    try {
      await drain(rec, false);
      const confirmed = rec.parts.length; // parts are ~8 MiB each
      setUploadProgress(
        recordedBytesRef.current > 0
          ? Math.min(0.99, (confirmed * 8 * 1024 * 1024) / recordedBytesRef.current)
          : 0,
      );
    } catch (err) {
      setError(errMessage(err, "Upload stalled — it will retry."));
    } finally {
      drainingRef.current = false;
    }
  }, []);

  const startRecording = useCallback(async () => {
    const stream = streamRef.current;
    if (!stream) return;
    setError(null);

    const mimeType = pickMimeType();
    let created;
    try {
      created = await createVideoUpload(date, baseContentType(mimeType));
    } catch (err) {
      setError(
        err instanceof ApiError && err.status === 409
          ? "This day's edit window has closed — you can no longer add a recording."
          : errMessage(err, "Couldn't start the recording."),
      );
      setPhase("error");
      return;
    }

    const rec: RecordingRecord = {
      videoId: created.video.id,
      date,
      mimeType,
      createdAt: Date.now(),
      nextPartNumber: 1,
      parts: [],
      chunks: [],
      stopped: false,
    };
    recordRef.current = rec;
    recordedBytesRef.current = 0;
    void saveRecording(rec);

    let recorder: MediaRecorder;
    try {
      recorder = new MediaRecorder(stream, mimeType ? { mimeType } : undefined);
    } catch (err) {
      // Couldn't start the recorder — discard the upload we just opened.
      void deleteVideo(rec.videoId).catch(() => {});
      void deleteRecording(rec.videoId).catch(() => {});
      recordRef.current = null;
      setError(errMessage(err, "This browser can't record video."));
      setPhase("error");
      return;
    }
    recorderRef.current = recorder;
    recorder.ondataavailable = (e: BlobEvent) => void handleChunk(e.data);
    recorder.onstop = async () => {
      setPhase("finalizing");
      stopTimer();
      try {
        while (drainingRef.current) await sleep(100);
        const durationSec = Math.max(1, Math.round((Date.now() - startedAtRef.current) / 1000));
        const video = await finalize(recordRef.current as RecordingRecord, durationSec);
        stopStream();
        setResult(video);
        setUploadProgress(1);
        setPhase("done");
      } catch (err) {
        setError(errMessage(err, "Couldn't finish uploading the recording."));
        setPhase("error");
      }
    };

    // A 4s timeslice: dataavailable fires every 4s so a crash loses at most
    // that much, and parts stream up during recording.
    recorder.start(4000);
    startedAtRef.current = Date.now();
    setElapsedSec(0);
    setUploadProgress(0);
    setPhase("recording");
    timerRef.current = window.setInterval(() => {
      setElapsedSec(Math.round((Date.now() - startedAtRef.current) / 1000));
    }, 1000);
  }, [date, handleChunk, stopStream, stopTimer]);

  const stopRecording = useCallback(() => {
    const recorder = recorderRef.current;
    if (recorder && recorder.state !== "inactive") {
      recorder.stop();
    }
  }, []);

  const cancel = useCallback(async () => {
    stopTimer();
    try {
      if (recorderRef.current && recorderRef.current.state !== "inactive") {
        recorderRef.current.onstop = null;
        recorderRef.current.stop();
      }
    } catch {
      // already stopped
    }
    stopStream();
    const rec = recordRef.current;
    recordRef.current = null;
    if (rec) {
      try {
        await deleteVideo(rec.videoId);
      } catch {
        // best effort — the server sweeps orphaned uploads
      }
      await deleteRecording(rec.videoId).catch(() => {});
    }
    setPhase("idle");
  }, [stopStream, stopTimer]);

  const resumeUpload = useCallback(async () => {
    const rec = resumable;
    if (!rec) return;
    setPhase("finalizing");
    setError(null);
    try {
      // Duration isn't known for a recording recovered after a crash.
      const video = await finalize(rec, null);
      setResult(video);
      setUploadProgress(1);
      setResumable(null);
      setPhase("done");
    } catch (err) {
      setError(errMessage(err, "Couldn't finish the earlier recording."));
      setPhase("error");
    }
  }, [resumable]);

  const discardResumable = useCallback(async () => {
    const rec = resumable;
    if (!rec) return;
    setResumable(null);
    try {
      await deleteVideo(rec.videoId);
    } catch {
      // best effort
    }
    await deleteRecording(rec.videoId).catch(() => {});
  }, [resumable]);

  return {
    phase,
    error,
    elapsedSec,
    uploadProgress,
    resumable,
    result,
    previewStream,
    beginPreview,
    startRecording,
    stopRecording,
    cancel,
    resumeUpload,
    discardResumable,
  };
}
