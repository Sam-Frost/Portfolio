import { useEffect, useState } from "react";

// Ticks once a second while goalDateIso is set, returning ms remaining
// (clamped to >= 0) or null when there's no goal date configured.
export function useCountdown(goalDateIso: string | null): number | null {
  const [now, setNow] = useState(() => Date.now());

  useEffect(() => {
    if (!goalDateIso) return;
    const id = setInterval(() => setNow(Date.now()), 1000);
    return () => clearInterval(id);
  }, [goalDateIso]);

  if (!goalDateIso) return null;
  return Math.max(0, new Date(goalDateIso).getTime() - now);
}
