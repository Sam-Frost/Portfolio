import { createContext, useContext, useEffect, useRef, useState, type ReactNode } from "react";
import * as api from "./api";
import { formatClock } from "./dateUtils";
import type { WorkSession } from "./types";
import { FloatingTimerTile } from "./FloatingTimerTile";
import { CompleteSessionDialog } from "./CompleteSessionDialog";

interface HourlyTrackerContextValue {
  // null while the initial rehydrate is in flight or once it's resolved to
  // "nothing running".
  session: WorkSession | null;
  remainingSeconds: number;
  start: (plannedMinutes: number) => Promise<void>;
}

const HourlyTrackerContext = createContext<HourlyTrackerContextValue | null>(null);

export function useHourlyTracker(): HourlyTrackerContextValue {
  const ctx = useContext(HourlyTrackerContext);
  if (!ctx) throw new Error("useHourlyTracker must be used within HourlyTrackerProvider");
  return ctx;
}

function remainingSecondsFor(session: WorkSession): number {
  const deadline = new Date(session.startedAt).getTime() + session.plannedMinutes * 60_000;
  return Math.max(0, Math.round((deadline - Date.now()) / 1000));
}

// Mounted once in DomainLayout (see SpotifyPlayerProvider for the same
// "state that survives navigation across the whole gated app" pattern):
// holds the single running session (rehydrated from GET /current on
// mount, so the floating timer survives a page reload), a live countdown,
// and renders the floating timer tile / the "time's up" note prompt on top
// of whatever page is showing.
export function HourlyTrackerProvider({ children }: { children: ReactNode }) {
  const [session, setSession] = useState<WorkSession | null>(null);
  const [remainingSeconds, setRemainingSeconds] = useState(0);
  const [showCompleteDialog, setShowCompleteDialog] = useState(false);
  const originalTitleRef = useRef<string | null>(null);
  // Guards against re-showing the complete prompt for the same session on
  // every tick once remainingSeconds is pinned at 0.
  const promptedSessionIdRef = useRef<string | null>(null);

  useEffect(() => {
    api
      .fetchCurrentSession()
      .then((s) => setSession(s ?? null))
      .catch(() => {
        // Best-effort rehydrate — worst case the floating timer just
        // doesn't reappear until the next successful fetch/navigation.
      });
  }, []);

  // Live countdown, ticking every second while a session is running, and
  // — in the same tick, so it can never see a stale/uninitialized
  // remainingSeconds from a separate effect — the prompt-for-a-note check
  // the moment it reaches zero. That includes firing immediately on
  // rehydrate if the planned duration already elapsed while the app was
  // closed, but NOT firing on every rehydrate merely because
  // remainingSeconds's initial state (0) hasn't been recomputed yet for a
  // session that still has time left.
  useEffect(() => {
    if (!session) {
      setRemainingSeconds(0);
      return;
    }

    const current = session;
    function tick() {
      const remaining = remainingSecondsFor(current);
      setRemainingSeconds(remaining);
      if (remaining <= 0 && promptedSessionIdRef.current !== current.id) {
        promptedSessionIdRef.current = current.id;
        setShowCompleteDialog(true);
      }
    }

    tick();
    const interval = window.setInterval(tick, 1000);
    return () => window.clearInterval(interval);
  }, [session]);

  // Mirror the countdown into the tab title while a session is running,
  // restoring the page's original title otherwise (including on unmount).
  useEffect(() => {
    if (originalTitleRef.current === null) originalTitleRef.current = document.title;
    const original = originalTitleRef.current;

    if (!session || remainingSeconds <= 0) {
      document.title = original;
      return;
    }

    document.title = `${formatClock(remainingSeconds)} — Hourly Tracker`;
    return () => {
      document.title = original;
    };
  }, [session, remainingSeconds]);

  async function start(plannedMinutes: number) {
    const created = await api.startSession(plannedMinutes);
    promptedSessionIdRef.current = null;
    setSession(created);
  }

  async function complete(note: string) {
    if (!session) return;
    await api.completeSession(session.id, note);
    setSession(null);
    setShowCompleteDialog(false);
  }

  async function cancel(note?: string) {
    if (!session) return;
    await api.cancelSession(session.id, note);
    setSession(null);
    setShowCompleteDialog(false);
  }

  const value: HourlyTrackerContextValue = { session, remainingSeconds, start };

  return (
    <HourlyTrackerContext.Provider value={value}>
      {children}
      {session && !showCompleteDialog && (
        <FloatingTimerTile session={session} remainingSeconds={remainingSeconds} onCancel={cancel} />
      )}
      {session && showCompleteDialog && <CompleteSessionDialog session={session} onSubmit={complete} />}
    </HourlyTrackerContext.Provider>
  );
}
