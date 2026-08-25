import { Hourglass } from "lucide-react";
import type { TimeLeftFormat } from "../settings/types";
import { formatCountdown } from "./formatCountdown";
import { useCountdown } from "./useCountdown";

interface TimeLeftClockProps {
  goalDate: string;
  format: TimeLeftFormat;
}

// Persistent pill shown centered in the domain top bar once a goal date is
// configured in Settings.
export function TimeLeftClock({ goalDate, format }: TimeLeftClockProps) {
  const msRemaining = useCountdown(goalDate);
  if (msRemaining === null) return null;

  return (
    <div className="flex items-center gap-1.5 rounded-full border-(--line) border-[0.5px] border-solid bg-(--card) px-3 py-1 text-[length:var(--text-caption)] text-(--fg) tabular-nums">
      <Hourglass size={12} className="text-(--text-faint)" />
      {msRemaining <= 0 ? "Goal reached" : formatCountdown(msRemaining, format)}
    </div>
  );
}
