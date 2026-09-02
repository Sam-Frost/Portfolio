import { apiRequest } from "../../../lib/apiClient";
import type { CreatedVideo, DiaryVideo, UploadedPart } from "./videoTypes";

// List a day's video log clips, oldest first. Includes still-pending
// uploads so a resumable recording stays visible.
export function listDayVideos(date: string): Promise<DiaryVideo[]> {
  return apiRequest<DiaryVideo[]>(`/api/diary/videos/${date}`);
}

// How many ready clips each date in [from, to] has — for the calendar, so a
// day with only a video (no written entry) is still marked.
export function fetchVideoDateCounts(from: string, to: string): Promise<Record<string, number>> {
  const params = new URLSearchParams({ from, to });
  return apiRequest<Record<string, number>>(`/api/diary/videos?${params.toString()}`);
}

// Open a multipart upload for a new clip on `date`. Rejected with 409 once
// the day's 24-hour edit window has closed.
export function createVideoUpload(date: string, contentType: string, title?: string): Promise<CreatedVideo> {
  return apiRequest<CreatedVideo>(`/api/diary/videos/${date}`, {
    method: "POST",
    body: JSON.stringify({ contentType, title: title ?? null }),
  });
}

// Presigned URL to PUT one part's bytes to. partNumber is 1-based.
export function getUploadPartUrl(videoId: string, partNumber: number): Promise<string> {
  return apiRequest<{ url: string }>(`/api/diary/videos/${videoId}/parts`, {
    method: "POST",
    body: JSON.stringify({ partNumber }),
  }).then((r) => r.url);
}

// Finalize the upload: the ETags collected from each part + the measured
// clip duration. Returns the now-ready row.
export function completeVideoUpload(
  videoId: string,
  parts: UploadedPart[],
  durationSeconds: number | null,
): Promise<DiaryVideo> {
  return apiRequest<DiaryVideo>(`/api/diary/videos/${videoId}/complete`, {
    method: "POST",
    body: JSON.stringify({ parts, durationSeconds }),
  });
}

// Short-lived URL to stream the clip from (the store serves Range requests,
// so <video> seeking works).
export function getPlaybackUrl(videoId: string): Promise<string> {
  return apiRequest<{ url: string }>(`/api/diary/videos/${videoId}/play`).then((r) => r.url);
}

export function deleteVideo(videoId: string): Promise<void> {
  return apiRequest<void>(`/api/diary/videos/${videoId}`, { method: "DELETE" });
}
