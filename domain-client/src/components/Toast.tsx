import { useEffect, useState } from "react";

interface ToastProps {
  message: string;
  onDone: () => void;
  duration?: number;
}

// Mounted per toast via a changing `key` from the caller, so each new toast
// restarts this animation from scratch instead of re-triggering an update.
export function Toast({ message, onDone, duration = 2200 }: ToastProps) {
  const [visible, setVisible] = useState(false);

  useEffect(() => {
    const raf = requestAnimationFrame(() => setVisible(true));
    const hideTimer = setTimeout(() => setVisible(false), duration);
    const doneTimer = setTimeout(onDone, duration + 300);

    return () => {
      cancelAnimationFrame(raf);
      clearTimeout(hideTimer);
      clearTimeout(doneTimer);
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  return (
    <div
      role="status"
      aria-live="polite"
      className={`absolute top-4 left-1/2 z-50 -translate-x-1/2 transition-all duration-300 ease-out ${
        visible ? "opacity-100 translate-y-0" : "opacity-0 -translate-y-4"
      }`}
    >
      <div className="rounded-lg border-(--line) border-[0.5px] border-solid bg-(--card) px-4 py-2 text-[length:var(--text-pill)] text-(--fg) shadow-lg">
        {message}
      </div>
    </div>
  );
}
