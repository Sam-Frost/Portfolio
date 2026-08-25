import { apiRequest } from "../../lib/apiClient";
import type { Label, LabelColor } from "./types";

export function fetchLabels(): Promise<Label[]> {
  return apiRequest<Label[]>("/api/labels");
}

export function createLabel(input: { name: string; color: LabelColor }): Promise<Label> {
  return apiRequest<Label>("/api/labels", {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export function updateLabel(id: string, input: { name: string; color: LabelColor }): Promise<Label> {
  return apiRequest<Label>(`/api/labels/${id}`, {
    method: "PATCH",
    body: JSON.stringify(input),
  });
}

export function deleteLabel(id: string): Promise<void> {
  return apiRequest<void>(`/api/labels/${id}`, { method: "DELETE" });
}
