import { useEffect, useRef, useState, type PointerEvent as ReactPointerEvent } from "react";
import { X } from "lucide-react";
import { formatClock } from "./dateUtils";
import type { FinishPayload, WorkSession } from "./types";
import { CancelSessionDialog } from "./CancelSessionDialog";

const POSITION_KEY = "hourlyTracker.tilePosition";
// Keep in sync with the tile's rendered size (w-44 + padding) below — used
// only to clamp the draggable position within the viewport.
const TILE_WIDTH = 176;
const TILE_HEIGHT = 84;
const MARGIN = 16;

interface Position {
  top: number;
  left: number;
}

function defaultPosition(): Position {
  return {
    top: window.innerHeight - TILE_HEIGHT - MARGIN,
    left: window.innerWidth - TILE_WIDTH - MARGIN,
  };
}

function clamp(position: Position): Position {
  const maxLeft = Math.max(0, window.innerWidth - TILE_WIDTH);
  const maxTop = Math.max(0, window.innerHeight - TILE_HEIGHT);
  return {
    left: Math.min(Math.max(position.left, 0), maxLeft),
    top: Math.min(Math.max(position.top, 0), maxTop),
  };
}

function loadPosition(): Position {
  try {
    const raw = localStorage.getItem(POSITION_KEY);
    if (!raw) return defaultPosition();
    const parsed = JSON.parse(raw) as Partial<Position>;
    if (typeof parsed.top !== "number" || typeof parsed.left !== "number") return defaultPosition();
    return clamp({ top: parsed.top, left: parsed.left });
  } catch {
    return defaultPosition();
  }
}

function savePosition(position: Position): void {
  try {
    localStorage.setItem(POSITION_KEY, JSON.stringify(position));
  } catch {
    // Best-effort — a failed write just means the tile resets to the
    // default corner next reload.
  }
}

interface FloatingTimerTileProps {
  session: WorkSession;
  remainingSeconds: number;
  onCancel: (payload: FinishPayload) => Promise<void>;
}

// Fixed-position, draggable across every page in the domain area (rendered
// by HourlyTrackerProvider, not any one route) — its dragged screen
// position persists in localStorage so it doesn't reset on navigation or
// reload.
export function FloatingTimerTile({ session, remainingSeconds, onCancel }: FloatingTimerTileProps) {
  const [position, setPosition] = useState<Position>(loadPosition);
  const [showCancelConfirm, setShowCancelConfirm] = useState(false);
  const tileRef = useRef<HTMLDivElement>(null);
  const dragState = useRef<{ pointerId: number; offsetX: number; offsetY: number } | null>(null);

  // Re-clamp if the viewport shrinks (e.g. window resize) so the tile can't
  // end up stranded off-screen.
  useEffect(() => {
    function handleResize() {
      setPosition((prev) => clamp(prev));
    }
    window.addEventListener("resize", handleResize);
    return () => window.removeEventListener("resize", handleResize);
  }, []);

  function handlePointerDown(e: ReactPointerEvent<HTMLDivElement>) {
    // Don't start a drag from the cancel button.
    if ((e.target as HTMLElement).closest("button")) return;

    const rect = tileRef.current?.getBoundingClientRect();
    if (!rect) return;
    dragState.current = { pointerId: e.pointerId, offsetX: e.clientX - rect.left, offsetY: e.clientY - rect.top };
    tileRef.current?.setPointerCapture(e.pointerId);
  }

  function handlePointerMove(e: ReactPointerEvent<HTMLDivElement>) {
    if (!dragState.current || dragState.current.pointerId !== e.pointerId) return;
    setPosition(clamp({ left: e.clientX - dragState.current.offsetX, top: e.clientY - dragState.current.offsetY }));
  }

  function endDrag(e: ReactPointerEvent<HTMLDivElement>) {
    if (!dragState.current || dragState.current.pointerId !== e.pointerId) return;
    dragState.current = null;
    setPosition((prev) => {
      savePosition(prev);
      return prev;
    });
  }

  return (
    <>
      <div
        ref={tileRef}
        style={{ top: position.top, left: position.left }}
        className="fixed z-40 w-44 select-none touch-none rounded-xl border-(--line) border-[0.5px] border-solid bg-(--card) p-3 shadow-lg cursor-grab active:cursor-grabbing"
        onPointerDown={handlePointerDown}
        onPointerMove={handlePointerMove}
        onPointerUp={endDrag}
        onPointerCancel={endDrag}
      >
        <div className="flex items-start justify-between gap-2">
          <div className="min-w-0">
            <p className="text-[length:var(--text-pill)] uppercase tracking-wide text-(--text-faint)">
              Sessions
            </p>
            <p className="font-space text-xl text-(--fg) tabular-nums">{formatClock(remainingSeconds)}</p>
          </div>
          <button
            type="button"
            onClick={() => setShowCancelConfirm(true)}
            aria-label="Cancel session"
            className="shrink-0 text-(--text-faint) hover:text-(--fg) transition-colors cursor-pointer"
          >
            <X size={14} />
          </button>
        </div>
      </div>

      {showCancelConfirm && (
        <CancelSessionDialog
          session={session}
          elapsedSeconds={session.plannedMinutes * 60 - remainingSeconds}
          onCancel={() => setShowCancelConfirm(false)}
          onConfirm={async (payload) => {
            await onCancel(payload);
            setShowCancelConfirm(false);
          }}
        />
      )}
    </>
  );
}
