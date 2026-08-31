export interface Todo {
  id: string;
  name: string;
  description: string | null;
  dateAdded: string;
  targetDate: string | null;
  done: boolean;
  completedAt: string | null;
  labelId: string | null;
}

export type SortField = "dateAdded" | "targetDate" | "completedAt";
