import { apiRequest } from "../../lib/apiClient";
import type { WorkOverview, WorkTab, WorkTask } from "./types";

// --- tabs ---

export function fetchTabs(): Promise<WorkTab[]> {
  return apiRequest<WorkTab[]>("/api/work/tabs");
}

export function createTab(name: string): Promise<WorkTab> {
  return apiRequest<WorkTab>("/api/work/tabs", { method: "POST", body: JSON.stringify({ name }) });
}

export function renameTab(id: string, name: string): Promise<WorkTab> {
  return apiRequest<WorkTab>(`/api/work/tabs/${id}`, { method: "PATCH", body: JSON.stringify({ name }) });
}

export function deleteTab(id: string): Promise<void> {
  return apiRequest<void>(`/api/work/tabs/${id}`, { method: "DELETE" });
}

// --- tasks ---

export function fetchTasks(tabId: string): Promise<WorkTask[]> {
  return apiRequest<WorkTask[]>(`/api/work/tabs/${tabId}/tasks`);
}

export function createTask(
  tabId: string,
  input: { name: string; description: string | null; targetDate: string | null },
): Promise<WorkTask> {
  return apiRequest<WorkTask>(`/api/work/tabs/${tabId}/tasks`, {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export function updateTask(
  id: string,
  input: Partial<{
    name: string;
    description: string | null;
    targetDate: string | null;
    done: boolean;
    jiraAcknowledged: boolean;
  }>,
): Promise<WorkTask> {
  return apiRequest<WorkTask>(`/api/work/tasks/${id}`, { method: "PATCH", body: JSON.stringify(input) });
}

export function deleteTask(id: string): Promise<void> {
  return apiRequest<void>(`/api/work/tasks/${id}`, { method: "DELETE" });
}

export function fetchOverview(): Promise<WorkOverview> {
  return apiRequest<WorkOverview>("/api/work/overview");
}
