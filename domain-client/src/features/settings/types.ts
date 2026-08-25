export type TimeLeftFormat = "weeks_days_time" | "days_time";

export interface Settings {
  dailyWorkTracker: {
    totalWorkHoursRequired: number | null;
  };
  timeLeftClock: {
    goalDate: string | null;
    format: TimeLeftFormat;
  };
}
