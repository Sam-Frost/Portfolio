import { useEffect, useState } from "react";
import { Outlet, useLocation, useMatches } from "react-router-dom";
import { LogOut, Menu } from "lucide-react";
import { DomainSidebar } from "../components/DomainSidebar";
import { LeaveDomainDialog } from "../components/LeaveDomainDialog";
import { PUBLIC_SITE_URL } from "../config";
import { clearToken, consumeJustLoggedIn } from "../features/auth/token";
import { fetchSettings } from "../features/settings/api";
import { HourlyTrackerProvider } from "../features/hourly-tracker/HourlyTrackerProvider";
import { SpotifyPlayerProvider } from "../features/spotify/SpotifyPlayerProvider";
import type { Settings } from "../features/settings/types";
import { TimeLeftClock } from "../features/time-left-clock/TimeLeftClock";
import { TimeLeftClockModal } from "../features/time-left-clock/TimeLeftClockModal";

type RouteHandle = { title: string; subtitle?: string; fullWidth?: boolean };

export function DomainLayout() {
  const matches = useMatches();
  const handle = [...matches].reverse().find((match) => match.handle)?.handle as RouteHandle | undefined;
  const [showLeaveDialog, setShowLeaveDialog] = useState(false);
  const [navOpen, setNavOpen] = useState(false);
  const [settings, setSettings] = useState<Settings | null>(null);
  const location = useLocation();
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

  // Close the mobile nav drawer whenever the route changes (covers programmatic
  // navigation, not just sidebar link clicks).
  useEffect(() => {
    setNavOpen(false);
  }, [location.pathname]);

  const goalDate = settings?.timeLeftClock.goalDate;

  return (
    <SpotifyPlayerProvider>
      <HourlyTrackerProvider>
        <div className="h-dvh flex bg-(--bg) overflow-hidden">
          <DomainSidebar open={navOpen} onClose={() => setNavOpen(false)} />
          <div className="flex-1 flex flex-col min-w-0">
            <div className="shrink-0 border-b-[0.5px] border-(--line-soft) px-4 sm:px-6 py-3 sm:py-4 grid grid-cols-[1fr_auto] sm:grid-cols-[1fr_auto_1fr] items-center gap-3 sm:gap-4">
              <div className="min-w-0 flex items-center gap-2 sm:gap-3">
                <button
                  type="button"
                  className="lg:hidden -ml-1 shrink-0 rounded-md p-1.5 text-(--text-muted) hover:text-(--fg) hover:bg-(--card-alt) transition-colors cursor-pointer"
                  onClick={() => setNavOpen(true)}
                  aria-label="Open navigation menu"
                >
                  <Menu size={18} />
                </button>
                <div className="min-w-0">
                  <span className="font-space font-semibold text-(--fg) truncate block">{handle?.title ?? "Domain"}</span>
                  {handle?.subtitle && (
                    <p className="text-[length:var(--text-caption)] text-(--text-muted) truncate">{handle.subtitle}</p>
                  )}
                </div>
              </div>

              <div className="hidden sm:flex justify-center">
                {settings && goalDate && <TimeLeftClock goalDate={goalDate} format={settings.timeLeftClock.format} />}
              </div>

              <button
                className="justify-self-end flex items-center gap-1.5 text-[length:var(--text-caption)] text-(--text-muted) hover:text-(--fg) cursor-pointer"
                onClick={() => setShowLeaveDialog(true)}
              >
                <LogOut size={14} />
                <span className="hidden sm:inline">Leave Domain</span>
              </button>
            </div>
            {/* min-h-0 lets a page (e.g. Todos) opt into its own internal scroll
                region instead of this whole area scrolling. relative scopes
                page-level overlays (e.g. Toast) to this non-sidebar content
                area instead of the full viewport. */}
            <div
              className={`relative flex-1 min-h-0 overflow-y-auto lg:overflow-hidden px-4 sm:px-6 py-4 sm:py-6 w-full ${
                handle?.fullWidth ? "" : "max-w-(--maxw) mx-auto"
              }`}
            >
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
      </HourlyTrackerProvider>
    </SpotifyPlayerProvider>
  );
}
