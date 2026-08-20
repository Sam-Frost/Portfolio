import { Outlet } from "react-router-dom";
import { DomainSidebar } from "../components/DomainSidebar";
import { PUBLIC_SITE_URL } from "../config";

export function DomainLayout() {
  return (
    <div className="min-h-screen flex bg-(--bg)">
      <DomainSidebar />
      <div className="flex-1 flex flex-col">
        <div className="border-b-[0.5px] border-(--line-soft) px-6 py-4 flex items-center justify-between">
          <span className="font-space text-(--fg)">Domain</span>
          <button
            className="text-[length:var(--text-caption)] text-(--text-muted) hover:text-(--fg) cursor-pointer"
            onClick={() => { window.location.href = PUBLIC_SITE_URL; }}
          >
            Exit
          </button>
        </div>
        <div className="flex-1 px-6 py-6 w-full max-w-(--maxw)">
          <Outlet />
        </div>
      </div>
    </div>
  );
}
