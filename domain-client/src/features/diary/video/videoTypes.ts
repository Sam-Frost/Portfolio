// Mirrors server/internal/diary.Video's JSON shape.
export interface DiaryVideo {
  id: string;
  entryDate: string; // "YYYY-MM-DD" (IST calendar date)
  title: string | null;
  contentType: string; // "video/mp4" | "video/webm"
  sizeBytes: number;
  durationSeconds: number | null;
  status: "pending" | "ready";
  createdAt: string;
  uploadedAt: string | null;
}

// Response of POST /api/diary/videos/{date}.
export interface CreatedVideo {
  video: DiaryVideo;
  uploadId: string;
}

// One uploaded multipart part — collected as the recording streams up and
// handed back to the complete call. Matches blobstore.Part.
export interface UploadedPart {
  number: number;
  etag: string;
}
