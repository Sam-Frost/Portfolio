import { apiRequest } from "../../lib/apiClient";
import type { ResourceInput, Resource, Subtopic, Topic } from "./types";

export function fetchTopics(): Promise<Topic[]> {
  return apiRequest<Topic[]>("/api/upskill/topics");
}

export function fetchTopic(id: string): Promise<Topic> {
  return apiRequest<Topic>(`/api/upskill/topics/${id}`);
}

export function createTopic(input: { name: string; targetDate: string | null }): Promise<Topic> {
  return apiRequest<Topic>("/api/upskill/topics", {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export function updateTopic(id: string, input: { name: string; targetDate: string | null }): Promise<Topic> {
  return apiRequest<Topic>(`/api/upskill/topics/${id}`, {
    method: "PATCH",
    // Partial update where a JSON null means "leave unchanged" — the edit
    // dialog always submits a full set of values, so a cleared targetDate
    // must be sent as "" to actually take effect (matches todos/api.ts).
    body: JSON.stringify({ name: input.name, targetDate: input.targetDate ?? "" }),
  });
}

export function deleteTopic(id: string): Promise<void> {
  return apiRequest<void>(`/api/upskill/topics/${id}`, { method: "DELETE" });
}

export function fetchSubtopics(topicId: string): Promise<Subtopic[]> {
  return apiRequest<Subtopic[]>(`/api/upskill/topics/${topicId}/subtopics`);
}

export function createSubtopic(
  topicId: string,
  input: { name: string; targetDate: string | null; resources: ResourceInput[] },
): Promise<Subtopic> {
  return apiRequest<Subtopic>(`/api/upskill/topics/${topicId}/subtopics`, {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export function updateSubtopic(
  id: string,
  input: { name?: string; targetDate?: string | null; done?: boolean },
): Promise<Subtopic> {
  return apiRequest<Subtopic>(`/api/upskill/subtopics/${id}`, {
    method: "PATCH",
    body: JSON.stringify(input),
  });
}

export function deleteSubtopic(id: string): Promise<void> {
  return apiRequest<void>(`/api/upskill/subtopics/${id}`, { method: "DELETE" });
}

export function addResource(subtopicId: string, input: ResourceInput): Promise<Resource> {
  return apiRequest<Resource>(`/api/upskill/subtopics/${subtopicId}/resources`, {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export function deleteResource(id: string): Promise<void> {
  return apiRequest<void>(`/api/upskill/resources/${id}`, { method: "DELETE" });
}
