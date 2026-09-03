export type TimeLeftFormat = "weeks_days_time" | "days_time";

export interface NotificationSettings {
  recipientEmail: string | null;
  /** Local (IST) "HH:MM" the daily overdue-todo digest is sent at. */
  morningTime: string;
  emailEnabled: boolean;
  pushEnabled: boolean;
}

export interface Settings {
  dailyWorkTracker: {
    totalWorkHoursRequired: number | null;
  };
  timeLeftClock: {
    goalDate: string | null;
    format: TimeLeftFormat;
  };
  notifications: NotificationSettings;
}
