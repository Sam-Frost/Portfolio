import { useState } from "react";
import { Outlet, useMatches } from "react-router-dom";
import { LogOut } from "lucide-react";
import { DomainSidebar } from "../components/DomainSidebar";
import { LeaveDomainDialog } from "../components/LeaveDomainDialog";
import { PUBLIC_SITE_URL } from "../config";
import { clearToken } from "../features/auth/token";

type RouteHandle = { title: string; subtitle?: string };

export function DomainLayout() {
  const matches = useMatches();
  const handle = [...matches].reverse().find((match) => match.handle)?.handle as RouteHandle | undefined;
  const [showLeaveDialog, setShowLeaveDialog] = useState(false);

  return (
    <div className="h-screen flex bg-(--bg) overflow-hidden">
      <DomainSidebar />
      <div className="flex-1 flex flex-col min-w-0">
        <div className="shrink-0 border-b-[0.5px] border-(--line-soft) px-6 py-4 flex items-center justify-between">
          <div>
            <span className="font-space font-semibold text-(--fg)">{handle?.title ?? "Domain"}</span>
            {handle?.subtitle && (
              <p className="text-[length:var(--text-caption)] text-(--text-muted)">{handle.subtitle}</p>
            )}
          </div>
          <button
            className="flex items-center gap-1.5 text-[length:var(--text-caption)] text-(--text-muted) hover:text-(--fg) cursor-pointer"
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
    </div>
  );
}
