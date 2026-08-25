import { apiRequest } from "../../lib/apiClient";
import type { Settings } from "./types";

export function fetchSettings(): Promise<Settings> {
  return apiRequest<Settings>("/api/settings");
}

export function updateSettings(patch: Partial<Settings>): Promise<Settings> {
  return apiRequest<Settings>("/api/settings", {
    method: "PATCH",
    body: JSON.stringify(patch),
  });
}
