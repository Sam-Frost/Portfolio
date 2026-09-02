import { useEffect, useState } from "react";
import { fetchSettings, updateSettings } from "./api";
import { LabelsSection } from "./LabelsSection";
import { DocumentLabelsSection } from "./DocumentLabelsSection";
import { NotepadLabelsSection } from "./NotepadLabelsSection";
import { FoodLibrarySection } from "./FoodLibrarySection";
import type { Settings, TimeLeftFormat } from "./types";

// Only sections that actually have a setting are rendered. As new
// per-section settings are added, give them their own <section> block below
// rather than a generic "coming soon" placeholder.

// datetime-local inputs work in "local wall clock" strings with no timezone;
// these convert to/from the RFC3339 UTC strings the backend stores.
function toDatetimeLocalValue(iso: string | null): string {
  if (!iso) return "";
  const d = new Date(iso);
  const pad = (n: number) => String(n).padStart(2, "0");
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`;
}

function fromDatetimeLocalValue(value: string): string | null {
  return value === "" ? null : new Date(value).toISOString();
}

export function SettingsPage() {
  const [settings, setSettings] = useState<Settings | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [savingClock, setSavingClock] = useState(false);

  useEffect(() => {
    fetchSettings()
      .then(setSettings)
      .catch((err) => setError(err instanceof Error ? err.message : "Couldn't load settings."))
      .finally(() => setLoading(false));
  }, []);

  function handleTimeLeftClockChange(patch: Partial<Settings["timeLeftClock"]>) {
    setSettings((prev) => {
      if (!prev) return prev;
      const timeLeftClock = { ...prev.timeLeftClock, ...patch };
      setSavingClock(true);
      updateSettings({ timeLeftClock })
        .catch((err) => setError(err instanceof Error ? err.message : "Couldn't save settings."))
        .finally(() => setSavingClock(false));
      return { ...prev, timeLeftClock };
    });
  }

  return (
    <div className="h-full overflow-y-auto themed-scrollbar pr-2">
      <h1 className="text-xl font-space font-semibold text-(--fg) mb-1.5">Settings</h1>
      <p className="text-[length:var(--text-caption)] text-(--text-muted) mb-6">
        Configuration for each domain section.
      </p>

      {loading && (
        <div className="bg-(--card) border-(--line) border-[0.5px] border-solid rounded-xl p-8 text-center text-(--text-faint) text-[length:var(--text-caption)]">
          Loading settings...
        </div>
      )}

      {!loading && error && (
        <div className="mb-4 flex items-center justify-between gap-2 rounded-lg border-(--line) border-[0.5px] border-solid bg-(--card) px-3 py-2 text-[length:var(--text-pill)] text-red-400">
          <span>{error}</span>
          <button
            onClick={() => setError(null)}
            aria-label="Dismiss error"
            className="shrink-0 text-(--text-faint) hover:text-(--fg) transition-colors cursor-pointer"
          >
            ×
          </button>
        </div>
      )}

      {!loading && (
        <div className="flex flex-col gap-4">
          {/* Labels has its own /api/labels CRUD, independent of the generic
              /api/settings resource above — it renders even if that fetch
              fails, unlike the sections below which display server state. */}
          <section className="bg-(--card) border-(--line) border-[0.5px] border-solid rounded-xl p-5">
            <h2 className="text-sm font-space font-medium text-(--fg) mb-3">Todos — Labels</h2>
            <LabelsSection />
          </section>

          {/* Documents has its own label set (server/internal/documentlabel),
              managed here just like the todo labels above. */}
          <section className="bg-(--card) border-(--line) border-[0.5px] border-solid rounded-xl p-5">
            <h2 className="text-sm font-space font-medium text-(--fg) mb-3">Documents — Labels</h2>
            <DocumentLabelsSection />
          </section>

          {/* Notepad has its own label set (server/internal/notepadlabel),
              managed here just like the todo labels above. */}
          <section className="bg-(--card) border-(--line) border-[0.5px] border-solid rounded-xl p-5">
            <h2 className="text-sm font-space font-medium text-(--fg) mb-3">Notepad — Labels</h2>
            <NotepadLabelsSection />
          </section>

          {/* The fitness food library has its own /api/fitness/foods CRUD and
              is shared across every cycle — managed here, not per-cycle. */}
          <section className="bg-(--card) border-(--line) border-[0.5px] border-solid rounded-xl p-5">
            <h2 className="text-sm font-space font-medium text-(--fg) mb-3">Fitness — Food library</h2>
            <FoodLibrarySection />
          </section>

          {settings && (
            <section className="bg-(--card) border-(--line) border-[0.5px] border-solid rounded-xl p-5">
              <h2 className="text-sm font-space font-medium text-(--fg) mb-3">Time Left Clock</h2>
              <p className="text-[length:var(--text-caption)] text-(--text-faint) mb-3">
                Shown as a popup on login, then permanently in the top bar.
              </p>
              <div className="flex flex-col gap-4 max-w-xs">
                <label className="flex flex-col gap-1.5">
                  <span className="text-[length:var(--text-caption)] text-(--text-muted)">Goal date & time</span>
                  <input
                    type="datetime-local"
                    value={toDatetimeLocalValue(settings.timeLeftClock.goalDate)}
                    onChange={(e) =>
                      handleTimeLeftClockChange({ goalDate: fromDatetimeLocalValue(e.target.value) })
                    }
                    className="rounded-lg border-(--line) border-[0.5px] border-solid bg-(--bg) px-3 py-2 text-[length:var(--text-caption)] text-(--fg) outline-none focus:border-(--line-strong)"
                  />
                </label>

                <label className="flex flex-col gap-1.5">
                  <span className="text-[length:var(--text-caption)] text-(--text-muted)">Countdown format</span>
                  <select
                    value={settings.timeLeftClock.format}
                    onChange={(e) => handleTimeLeftClockChange({ format: e.target.value as TimeLeftFormat })}
                    className="rounded-lg border-(--line) border-[0.5px] border-solid bg-(--bg) px-3 py-2 text-[length:var(--text-caption)] text-(--fg) outline-none focus:border-(--line-strong)"
                  >
                    <option value="weeks_days_time">Weeks + days + time</option>
                    <option value="days_time">Days + time</option>
                  </select>
                </label>

                {savingClock && (
                  <span className="text-[length:var(--text-pill)] text-(--text-faint)">Saving...</span>
                )}
              </div>
            </section>
          )}
        </div>
      )}
    </div>
  );
}
