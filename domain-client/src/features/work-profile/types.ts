export interface WorkTab {
  id: string;
  name: string;
  position: number;
  createdAt: string;
}

export interface WorkTask {
  id: string;
  tabId: string;
  name: string;
  description: string | null;
  targetDate: string | null;
  done: boolean;
  completedAt: string | null;
  jiraAcknowledged: boolean;
  createdAt: string;
}

export interface WorkTaskWithTab extends WorkTask {
  tabName: string;
}

export interface WorkOverview {
  dueToday: WorkTaskWithTab[];
  overdue: WorkTaskWithTab[];
}
