import { useEffect } from "react";
import { Navigate, Outlet } from "react-router-dom";
import { getToken } from "./token";

// Gate for every gated route: the real enforcement is server-side (every /api/*
// route but /api/auth/login requires a valid bearer JWT); this just keeps
// someone with no token out of the dashboard shell client-side too.
export function RequireAuth() {
  // The browser can restore this page from bfcache (e.g. hitting Back after
  // "Leave Domain") without re-running React, so a cleared token wouldn't
  // otherwise be re-checked until some other navigation happened. Force a
  // real reload so the token check below actually re-runs.
  useEffect(() => {
    const onPageShow = (event: PageTransitionEvent) => {
      if (event.persisted) window.location.reload();
    };
    window.addEventListener("pageshow", onPageShow);
    return () => window.removeEventListener("pageshow", onPageShow);
  }, []);

  return getToken() ? <Outlet /> : <Navigate to="/" replace />;
}
