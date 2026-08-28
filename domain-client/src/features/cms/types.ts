// Mirrors server/internal/cms model JSON. Every list item carries the
// CMS-only `visible` / `order` / `updatedAt` bookkeeping fields.

export interface CmsProject {
  id: string;
  title: string;
  slug: string;
  description: string;
  stack: string[];
  github: string;
  liveLink: string;
  visible: boolean;
  order: number;
  updatedAt: string;
}

export interface CmsExperience {
  id: string;
  logo: string;
  position: string;
  company: string;
  description: string;
  details: string[];
  techStack: string[];
  startDate: string;
  endDate: string;
  visible: boolean;
  order: number;
  updatedAt: string;
}

export interface CmsBlog {
  id: string;
  title: string;
  slug: string;
  readTime: string;
  genre: string;
  date: string;
  body: string;
  visible: boolean;
  order: number;
  updatedAt: string;
}

export interface CmsSummary {
  domain: string;
  imageSubText: string;
  heroHighlightText: string;
  heroName: string;
  heroSubText: string;
  heroDetails: string;
  updatedAt: string;
}

export type CmsSection = "summary" | "projects" | "experiences" | "blogs";

export interface CmsStatus {
  hasUnpublishedChanges: boolean;
  changedSections: CmsSection[];
  neverPublished: boolean;
  lastPublishedAt: string | null;
  lastPublishVersion: number;
  lastPublishStatus: string;
  lastPublishError: string | null;
}

export interface CmsPublication {
  id: string;
  version: number;
  publishedAt: string;
  status: string;
  error: string | null;
}

export interface ProjectInput {
  title: string;
  slug: string;
  description: string;
  stack: string[];
  github: string;
  liveLink: string;
  visible: boolean;
}

export interface ExperienceInput {
  logo: string;
  position: string;
  company: string;
  description: string;
  details: string[];
  techStack: string[];
  startDate: string;
  endDate: string;
  visible: boolean;
}

export interface BlogInput {
  title: string;
  slug: string;
  readTime: string;
  genre: string;
  date: string;
  body: string;
  visible: boolean;
}
