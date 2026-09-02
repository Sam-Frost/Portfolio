import { apiRequest } from "../../lib/apiClient";
import type { DrawingBoard, DrawingBoardScene, DrawingBoardSummary } from "./types";

export function fetchDrawingBoards(): Promise<DrawingBoardSummary[]> {
  return apiRequest<DrawingBoardSummary[]>("/api/drawing-boards");
}

export function fetchDrawingBoard(id: string): Promise<DrawingBoard> {
  return apiRequest<DrawingBoard>(`/api/drawing-boards/${id}`);
}

export function createDrawingBoard(name?: string): Promise<DrawingBoard> {
  return apiRequest<DrawingBoard>("/api/drawing-boards", {
    method: "POST",
    body: JSON.stringify({ name: name ?? null }),
  });
}

// Partial update — only the fields present are changed, matching the
// backend's UpdateInput convention. Autosave calls this with whichever of
// name/sceneData just changed.
export function updateDrawingBoard(
  id: string,
  patch: { name?: string; sceneData?: DrawingBoardScene },
): Promise<DrawingBoard> {
  return apiRequest<DrawingBoard>(`/api/drawing-boards/${id}`, {
    method: "PATCH",
    body: JSON.stringify(patch),
  });
}

export function deleteDrawingBoard(id: string): Promise<void> {
  return apiRequest<void>(`/api/drawing-boards/${id}`, { method: "DELETE" });
}
