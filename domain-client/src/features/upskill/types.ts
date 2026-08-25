export interface Topic {
  id: string;
  name: string;
  targetDate: string | null;
  dateAdded: string;
  subtopicCount: number;
  doneCount: number;
}

export interface Resource {
  id: string;
  subtopicId: string;
  label: string | null;
  url: string;
}

export interface Subtopic {
  id: string;
  topicId: string;
  name: string;
  targetDate: string | null;
  done: boolean;
  dateAdded: string;
  resources: Resource[];
}

export interface ResourceInput {
  label: string | null;
  url: string;
}
