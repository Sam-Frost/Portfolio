import { useEffect, useState } from "react";
import { Outlet, useMatches } from "react-router-dom";
import { LogOut } from "lucide-react";
import { DomainSidebar } from "../components/DomainSidebar";
import { LeaveDomainDialog } from "../components/LeaveDomainDialog";
import { PUBLIC_SITE_URL } from "../config";
import { clearToken, consumeJustLoggedIn } from "../features/auth/token";
import { fetchSettings } from "../features/settings/api";
import type { Settings } from "../features/settings/types";
import { TimeLeftClock } from "../features/time-left-clock/TimeLeftClock";
import { TimeLeftClockModal } from "../features/time-left-clock/TimeLeftClockModal";

type RouteHandle = { title: string; subtitle?: string };

export function DomainLayout() {
  const matches = useMatches();
  const handle = [...matches].reverse().find((match) => match.handle)?.handle as RouteHandle | undefined;
  const [showLeaveDialog, setShowLeaveDialog] = useState(false);
  const [settings, setSettings] = useState<Settings | null>(null);
  // Consumed once per login (see markJustLoggedIn/consumeJustLoggedIn) so the
  // countdown popup only appears right after signing in, not on every
  // navigation or page reload within the same session.
  const [showLoginModal, setShowLoginModal] = useState(() => consumeJustLoggedIn());

  useEffect(() => {
    fetchSettings()
      .then(setSettings)
      .catch(() => {
        // The top bar clock is a nice-to-have; a failed fetch just means it
        // stays hidden rather than blocking the rest of the domain area.
      });
  }, []);

  const goalDate = settings?.timeLeftClock.goalDate;

  return (
    <div className="h-screen flex bg-(--bg) overflow-hidden">
      <DomainSidebar />
      <div className="flex-1 flex flex-col min-w-0">
        <div className="shrink-0 border-b-[0.5px] border-(--line-soft) px-6 py-4 grid grid-cols-[1fr_auto_1fr] items-center gap-4">
          <div className="min-w-0">
            <span className="font-space font-semibold text-(--fg) truncate block">{handle?.title ?? "Domain"}</span>
            {handle?.subtitle && (
              <p className="text-[length:var(--text-caption)] text-(--text-muted) truncate">{handle.subtitle}</p>
            )}
          </div>

          <div className="flex justify-center">
            {settings && goalDate && <TimeLeftClock goalDate={goalDate} format={settings.timeLeftClock.format} />}
          </div>

          <button
            className="justify-self-end flex items-center gap-1.5 text-[length:var(--text-caption)] text-(--text-muted) hover:text-(--fg) cursor-pointer"
            onClick={() => setShowLeaveDialog(true)}
          >
            <LogOut size={14} />
            Leave Domain
          </button>
        </div>
        {/* min-h-0 lets a page (e.g. Todos) opt into its own internal scroll
            region instead of this whole area scrolling. */}
        <div className="flex-1 min-h-0 overflow-hidden px-6 py-6 w-full max-w-(--maxw)">
          <Outlet />
        </div>
      </div>
      {showLeaveDialog && (
        <LeaveDomainDialog
          onCancel={() => setShowLeaveDialog(false)}
          onConfirm={() => { clearToken(); window.location.href = PUBLIC_SITE_URL; }}
        />
      )}
      {showLoginModal && settings && goalDate && (
        <TimeLeftClockModal
          goalDate={goalDate}
          format={settings.timeLeftClock.format}
          onClose={() => setShowLoginModal(false)}
        />
      )}
    </div>
  );
}
