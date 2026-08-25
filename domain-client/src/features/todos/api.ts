import { apiRequest } from "../../lib/apiClient";
import type { Todo } from "./types";

export function fetchTodos(params?: {
  sortBy: "dateAdded" | "targetDate";
  order: "asc" | "desc";
  labelId?: string | null;
}): Promise<Todo[]> {
  if (!params) return apiRequest<Todo[]>("/api/todos");

  const query = new URLSearchParams({ sortBy: params.sortBy, order: params.order });
  if (params.labelId) query.set("labelId", params.labelId);
  return apiRequest<Todo[]>(`/api/todos?${query.toString()}`);
}

export function createTodo(input: {
  name: string;
  description: string | null;
  targetDate: string | null;
  labelId: string | null;
}): Promise<Todo> {
  return apiRequest<Todo>("/api/todos", {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export function setTodoDone(id: string, done: boolean): Promise<Todo> {
  return apiRequest<Todo>(`/api/todos/${id}`, {
    method: "PATCH",
    body: JSON.stringify({ done }),
  });
}

export function updateTodo(
  id: string,
  input: { name: string; description: string | null; targetDate: string | null; labelId: string | null },
): Promise<Todo> {
  return apiRequest<Todo>(`/api/todos/${id}`, {
    method: "PATCH",
    // PATCH is a partial update where a JSON null (or omitted field) means
    // "leave unchanged" — but the edit dialog always submits a full set of
    // values, so a cleared description/targetDate/labelId must be sent as
    // "" to actually take effect rather than being silently ignored.
    body: JSON.stringify({
      name: input.name,
      description: input.description ?? "",
      targetDate: input.targetDate ?? "",
      labelId: input.labelId ?? "",
    }),
  });
}

export function deleteTodo(id: string): Promise<void> {
  return apiRequest<void>(`/api/todos/${id}`, { method: "DELETE" });
}
