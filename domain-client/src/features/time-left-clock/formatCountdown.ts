import type { TimeLeftFormat } from "../settings/types";

const pad = (n: number) => String(n).padStart(2, "0");

// msRemaining must already be clamped to >= 0 by the caller.
export function formatCountdown(msRemaining: number, format: TimeLeftFormat): string {
  const totalSeconds = Math.floor(msRemaining / 1000);
  const seconds = totalSeconds % 60;
  const totalMinutes = Math.floor(totalSeconds / 60);
  const minutes = totalMinutes % 60;
  const totalHours = Math.floor(totalMinutes / 60);
  const hours = totalHours % 24;
  const totalDays = Math.floor(totalHours / 24);

  const time = `${pad(hours)}:${pad(minutes)}:${pad(seconds)}`;

  if (format === "weeks_days_time") {
    const weeks = Math.floor(totalDays / 7);
    const days = totalDays % 7;
    return `${weeks}w ${days}d ${time}`;
  }

  return `${totalDays}d ${time}`;
}
