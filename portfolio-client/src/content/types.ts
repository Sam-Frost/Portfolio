// Shape of the content.json the CMS publishes (server/internal/cms.Content).
// The public site fetches this at runtime; see client.ts + ContentContext.

export interface SiteProject {
  id: string;
  title: string;
  slug: string;
  description: string;
  stack: string[];
  github: string;
  liveLink: string;
  visible: boolean;
  order: number;
}

export interface SiteExperience {
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
}

export interface SiteBlog {
  id: string;
  title: string;
  slug: string;
  readTime: string;
  genre: string;
  date: string;
  body: string;
  visible: boolean;
  order: number;
}

export interface SiteSummary {
  domain: string;
  imageSubText: string;
  heroHighlightText: string;
  heroName: string;
  heroSubText: string;
  heroDetails: string;
}

export interface SiteContent {
  version: number;
  publishedAt: string;
  summary: SiteSummary;
  projects: SiteProject[];
  experiences: SiteExperience[];
  blogs: SiteBlog[];
}

// visibleSorted filters out hidden items and orders by the CMS `order`
// field — the public site's view of any content list.
export function visibleSorted<T extends { visible: boolean; order: number }>(items: T[]): T[] {
  return items.filter((i) => i.visible).sort((a, b) => a.order - b.order);
}
