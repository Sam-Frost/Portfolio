import { useEffect, useState } from "react";
import { sections } from "../../data/domainSections";
import { fetchSettings, updateSettings } from "./api";
import { LabelsSection } from "./LabelsSection";
import type { Settings } from "./types";

const settingsSections = sections.filter((section) => section.label !== "Settings");

export function SettingsPage() {
  const [settings, setSettings] = useState<Settings | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [savingHours, setSavingHours] = useState(false);

  useEffect(() => {
    fetchSettings()
      .then(setSettings)
      .catch((err) => setError(err instanceof Error ? err.message : "Couldn't load settings."))
      .finally(() => setLoading(false));
  }, []);

  function handleTotalHoursChange(value: string) {
    const totalWorkHoursRequired = value === "" ? null : Number(value);
    setSettings((prev) =>
      prev ? { ...prev, dailyWorkTracker: { ...prev.dailyWorkTracker, totalWorkHoursRequired } } : prev
    );
    setSavingHours(true);
    updateSettings({ dailyWorkTracker: { totalWorkHoursRequired } })
      .catch((err) => setError(err instanceof Error ? err.message : "Couldn't save settings."))
      .finally(() => setSavingHours(false));
  }

  return (
    <div className="h-full overflow-y-auto themed-scrollbar">
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

          {settings &&
            settingsSections.map((section) => (
            <section
              key={section.label}
              className="bg-(--card) border-(--line) border-[0.5px] border-solid rounded-xl p-5"
            >
              <h2 className="text-sm font-space font-medium text-(--fg) mb-3">{section.label}</h2>

              {section.label === "Daily Work Tracker" ? (
                <label className="flex flex-col gap-1.5 max-w-xs">
                  <span className="text-[length:var(--text-caption)] text-(--text-muted)">
                    Total work hours required (per day)
                  </span>
                  <input
                    type="number"
                    min={0}
                    step={0.5}
                    value={settings?.dailyWorkTracker.totalWorkHoursRequired ?? ""}
                    onChange={(e) => handleTotalHoursChange(e.target.value)}
                    className="rounded-lg border-(--line) border-[0.5px] border-solid bg-(--bg) px-3 py-2 text-[length:var(--text-caption)] text-(--fg) outline-none focus:border-(--line-strong)"
                  />
                  {savingHours && (
                    <span className="text-[length:var(--text-pill)] text-(--text-faint)">Saving...</span>
                  )}
                </label>
              ) : (
                <p className="text-[length:var(--text-caption)] text-(--text-faint)">No settings yet.</p>
              )}
            </section>
          ))}
        </div>
      )}
    </div>
  );
}
