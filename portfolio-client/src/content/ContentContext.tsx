import { createContext, useContext, useEffect, useState, type ReactNode } from "react";
import { loadContent } from "./client";
import { fallbackContent } from "./fallback";
import type { SiteContent } from "./types";

interface ContentValue {
  content: SiteContent;
  // false until the content.json fetch has resolved (success or fallback);
  // lets a deep-linked page (e.g. BlogPage) wait before deciding a slug is
  // genuinely missing rather than just not-loaded-yet.
  loaded: boolean;
}

const ContentContext = createContext<ContentValue>({ content: fallbackContent, loaded: false });

// useContent returns the site content. It's the compiled-in fallback on
// first render, then swaps to the fetched content.json once it resolves —
// components just re-render with the newer data, no loading state needed.
export function useContent(): SiteContent {
  return useContext(ContentContext).content;
}

export function useContentLoaded(): boolean {
  return useContext(ContentContext).loaded;
}

export function ContentProvider({ children }: { children: ReactNode }) {
  const [value, setValue] = useState<ContentValue>({ content: fallbackContent, loaded: false });

  useEffect(() => {
    let cancelled = false;
    loadContent().then((content) => {
      if (!cancelled) setValue({ content, loaded: true });
    });
    return () => {
      cancelled = true;
    };
  }, []);

  return <ContentContext.Provider value={value}>{children}</ContentContext.Provider>;
}
