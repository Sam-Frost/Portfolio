import { useEffect } from "react";
import type { TimeLeftFormat } from "../settings/types";
import { formatCountdown } from "./formatCountdown";
import { useCountdown } from "./useCountdown";

interface TimeLeftClockModalProps {
  goalDate: string;
  format: TimeLeftFormat;
  onClose: () => void;
}

const AUTO_CLOSE_MS = 3000;

// One-shot popup shown centered on screen right after login, then replaced
// by the persistent TimeLeftClock in the top bar.
export function TimeLeftClockModal({ goalDate, format, onClose }: TimeLeftClockModalProps) {
  const msRemaining = useCountdown(goalDate);

  useEffect(() => {
    const id = setTimeout(onClose, AUTO_CLOSE_MS);
    return () => clearTimeout(id);
  }, [onClose]);

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 px-4" role="presentation">
      <div
        className="rounded-xl border-(--line) border-[0.5px] border-solid bg-(--card) px-10 py-8 shadow-lg text-center"
        role="dialog"
        aria-modal="true"
        aria-label="Time left"
      >
        <p className="text-[length:var(--text-caption)] text-(--text-muted) mb-2">Time left</p>
        <p className="font-space text-3xl font-semibold text-(--fg) tabular-nums">
          {msRemaining === null ? "--" : msRemaining <= 0 ? "Goal reached" : formatCountdown(msRemaining, format)}
        </p>
      </div>
    </div>
  );
}
