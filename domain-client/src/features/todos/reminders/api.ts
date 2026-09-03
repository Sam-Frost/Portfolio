import { apiRequest } from "../../../lib/apiClient";
import type { CreateReminderInput, Reminder } from "./types";

export function fetchReminders(todoId: string): Promise<Reminder[]> {
  return apiRequest<Reminder[]>(`/api/todos/${todoId}/reminders`);
}

export function createReminder(todoId: string, input: CreateReminderInput): Promise<Reminder> {
  return apiRequest<Reminder>(`/api/todos/${todoId}/reminders`, {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export function deleteReminder(id: string): Promise<void> {
  return apiRequest<void>(`/api/reminders/${id}`, { method: "DELETE" });
}
