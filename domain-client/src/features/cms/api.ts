import { apiRequest } from "../../lib/apiClient";
import type {
  BlogInput,
  CmsBlog,
  CmsExperience,
  CmsProject,
  CmsPublication,
  CmsStatus,
  CmsSummary,
  ExperienceInput,
  ProjectInput,
} from "./types";

// ─────────────────────────────────────────────
// status / publish
// ─────────────────────────────────────────────

export function fetchCmsStatus(): Promise<CmsStatus> {
  return apiRequest<CmsStatus>("/api/cms/status");
}

export function publish(): Promise<CmsPublication> {
  return apiRequest<CmsPublication>("/api/cms/publish", { method: "POST" });
}

export function fetchPublications(): Promise<CmsPublication[]> {
  return apiRequest<CmsPublication[]>("/api/cms/publications");
}

// ─────────────────────────────────────────────
// projects
// ─────────────────────────────────────────────

export function fetchProjects(): Promise<CmsProject[]> {
  return apiRequest<CmsProject[]>("/api/cms/projects");
}

export function createProject(input: ProjectInput): Promise<CmsProject> {
  return apiRequest<CmsProject>("/api/cms/projects", { method: "POST", body: JSON.stringify(input) });
}

export function updateProject(id: string, input: Partial<ProjectInput> & { order?: number }): Promise<CmsProject> {
  return apiRequest<CmsProject>(`/api/cms/projects/${id}`, { method: "PATCH", body: JSON.stringify(input) });
}

export function deleteProject(id: string): Promise<void> {
  return apiRequest<void>(`/api/cms/projects/${id}`, { method: "DELETE" });
}

// ─────────────────────────────────────────────
// experiences
// ─────────────────────────────────────────────

export function fetchExperiences(): Promise<CmsExperience[]> {
  return apiRequest<CmsExperience[]>("/api/cms/experiences");
}

export function createExperience(input: ExperienceInput): Promise<CmsExperience> {
  return apiRequest<CmsExperience>("/api/cms/experiences", { method: "POST", body: JSON.stringify(input) });
}

export function updateExperience(
  id: string,
  input: Partial<ExperienceInput> & { order?: number },
): Promise<CmsExperience> {
  return apiRequest<CmsExperience>(`/api/cms/experiences/${id}`, { method: "PATCH", body: JSON.stringify(input) });
}

export function deleteExperience(id: string): Promise<void> {
  return apiRequest<void>(`/api/cms/experiences/${id}`, { method: "DELETE" });
}

// ─────────────────────────────────────────────
// blogs
// ─────────────────────────────────────────────

export function fetchBlogs(): Promise<CmsBlog[]> {
  return apiRequest<CmsBlog[]>("/api/cms/blogs");
}

export function createBlog(input: BlogInput): Promise<CmsBlog> {
  return apiRequest<CmsBlog>("/api/cms/blogs", { method: "POST", body: JSON.stringify(input) });
}

export function updateBlog(id: string, input: Partial<BlogInput> & { order?: number }): Promise<CmsBlog> {
  return apiRequest<CmsBlog>(`/api/cms/blogs/${id}`, { method: "PATCH", body: JSON.stringify(input) });
}

export function deleteBlog(id: string): Promise<void> {
  return apiRequest<void>(`/api/cms/blogs/${id}`, { method: "DELETE" });
}

// ─────────────────────────────────────────────
// summary
// ─────────────────────────────────────────────

export function fetchSummary(): Promise<CmsSummary> {
  return apiRequest<CmsSummary>("/api/cms/summary");
}

export function updateSummary(input: Partial<Omit<CmsSummary, "updatedAt">>): Promise<CmsSummary> {
  return apiRequest<CmsSummary>("/api/cms/summary", { method: "PATCH", body: JSON.stringify(input) });
}
