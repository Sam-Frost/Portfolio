export type ReminderKind = "once" | "repeat";

export interface Reminder {
  id: string;
  todoId: string;
  kind: ReminderKind;
  /** RFC3339. For a repeating reminder this is the next occurrence. */
  fireAt: string;
  intervalSeconds: number | null;
  createdAt: string;
}

export interface CreateReminderInput {
  kind: ReminderKind;
  /** RFC3339; required for "once". */
  fireAt?: string;
  /** seconds, >= 60; required for "repeat". */
  intervalSeconds?: number;
}
