// Mirrors server/internal/diary.Entry's JSON shape.
export interface DiaryEntry {
  id: string;
  entryDate: string; // "YYYY-MM-DD" (IST calendar date)
  content: string; // HTML — written with the same rich-text editor as notepad
  createdAt: string;
  updatedAt: string;
  // Computed server-side fresh on every response (see diary.IsLocked) —
  // never trust a stale cached value of this across a day boundary.
  locked: boolean;
}
