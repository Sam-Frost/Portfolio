import { useEffect, useRef, useState } from "react";
import { Bell, BellOff, Mail, Smartphone } from "lucide-react";
import type { NotificationSettings } from "../settings/types";
import {
  currentPermission,
  disablePushOnThisDevice,
  enablePushOnThisDevice,
  isStandalone,
  isSubscribedOnThisDevice,
} from "./push";
import { sendTestNotification } from "./api";

interface Props {
  value: NotificationSettings;
  onChange: (patch: Partial<NotificationSettings>) => void;
  saving: boolean;
}

const inputCls =
  "rounded-lg border-(--line) border-[0.5px] border-solid bg-(--bg) px-3 py-2 text-[length:var(--text-caption)] text-(--fg) outline-none focus:border-(--line-strong)";

function Toggle({
  checked,
  onChange,
  label,
  icon,
}: {
  checked: boolean;
  onChange: (next: boolean) => void;
  label: string;
  icon: React.ReactNode;
}) {
  return (
    <label className="flex items-center gap-2.5 cursor-pointer">
      <input
        type="checkbox"
        checked={checked}
        onChange={(e) => onChange(e.target.checked)}
        className="size-4 accent-(--fg) cursor-pointer"
      />
      <span className="flex items-center gap-1.5 text-[length:var(--text-caption)] text-(--fg)">
        {icon}
        {label}
      </span>
    </label>
  );
}

export function NotificationSettingsSection({ value, onChange, saving }: Props) {
  const [email, setEmail] = useState(value.recipientEmail ?? "");
  const [deviceState, setDeviceState] = useState<"loading" | "on" | "off" | "unsupported">("loading");
  const [busy, setBusy] = useState(false);
  const [msg, setMsg] = useState<string | null>(null);
  const [testMsg, setTestMsg] = useState<string | null>(null);
  const lastSavedEmail = useRef(value.recipientEmail ?? "");

  useEffect(() => {
    if (currentPermission() === "unsupported") {
      setDeviceState("unsupported");
      return;
    }
    isSubscribedOnThisDevice().then((sub) => setDeviceState(sub ? "on" : "off"));
  }, []);

  function commitEmail() {
    const trimmed = email.trim();
    if (trimmed === lastSavedEmail.current) return;
    lastSavedEmail.current = trimmed;
    onChange({ recipientEmail: trimmed === "" ? null : trimmed });
  }

  async function toggleDevice() {
    setBusy(true);
    setMsg(null);
    try {
      if (deviceState === "on") {
        await disablePushOnThisDevice();
        setDeviceState("off");
      } else {
        await enablePushOnThisDevice();
        setDeviceState("on");
        setMsg("Notifications enabled on this device.");
      }
    } catch (err) {
      setMsg(err instanceof Error ? err.message : "Couldn't change notification setting.");
    } finally {
      setBusy(false);
    }
  }

  async function runTest() {
    setTestMsg(null);
    try {
      await sendTestNotification();
      setTestMsg("Test sent — check this device and your email.");
    } catch (err) {
      setTestMsg(err instanceof Error ? err.message : "Couldn't send a test.");
    }
  }

  return (
    <div className="flex flex-col gap-5 max-w-md">
      <p className="text-[length:var(--text-caption)] text-(--text-faint)">
        A morning summary of todos due or overdue, plus per-todo reminders, delivered to this
        device and by email.
      </p>

      <label className="flex flex-col gap-1.5">
        <span className="text-[length:var(--text-caption)] text-(--text-muted)">Recipient email</span>
        <input
          type="email"
          inputMode="email"
          placeholder="you@example.com"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          onBlur={commitEmail}
          onKeyDown={(e) => e.key === "Enter" && commitEmail()}
          className={inputCls}
        />
      </label>

      <label className="flex flex-col gap-1.5 max-w-[10rem]">
        <span className="text-[length:var(--text-caption)] text-(--text-muted)">Morning digest time</span>
        <input
          type="time"
          value={value.morningTime}
          onChange={(e) => onChange({ morningTime: e.target.value })}
          className={inputCls}
        />
      </label>

      <div className="flex flex-col gap-2">
        <span className="text-[length:var(--text-caption)] text-(--text-muted)">Channels</span>
        <Toggle
          checked={value.emailEnabled}
          onChange={(next) => onChange({ emailEnabled: next })}
          label="Email"
          icon={<Mail size={14} />}
        />
        <Toggle
          checked={value.pushEnabled}
          onChange={(next) => onChange({ pushEnabled: next })}
          label="Push notifications"
          icon={<Smartphone size={14} />}
        />
      </div>

      <div className="flex flex-col gap-2 border-t-[0.5px] border-(--line-soft) pt-4">
        <span className="text-[length:var(--text-caption)] text-(--text-muted)">This device</span>
        {deviceState === "unsupported" && (
          <p className="text-[length:var(--text-pill)] text-(--text-faint)">
            This browser can't receive push notifications. On iPhone, open this app from your
            Home Screen.
          </p>
        )}
        {deviceState !== "unsupported" && (
          <>
            {!isStandalone() && (
              <p className="text-[length:var(--text-pill)] text-(--text-faint)">
                On iPhone: add this app to your Home Screen and open it from there before
                enabling.
              </p>
            )}
            <button
              type="button"
              disabled={busy || deviceState === "loading"}
              onClick={toggleDevice}
              className={`flex w-fit items-center gap-1.5 rounded-lg px-3 py-2 text-[length:var(--text-caption)] transition-colors ${
                deviceState === "on"
                  ? "bg-(--card-alt) text-(--text-muted) hover:text-(--fg)"
                  : "bg-(--fg) text-(--bg)"
              } ${busy || deviceState === "loading" ? "opacity-60 cursor-not-allowed" : "cursor-pointer"}`}
            >
              {deviceState === "on" ? <BellOff size={14} /> : <Bell size={14} />}
              {deviceState === "on" ? "Disable on this device" : "Enable notifications on this device"}
            </button>
          </>
        )}
        {msg && <p className="text-[length:var(--text-pill)] text-(--text-faint)">{msg}</p>}
      </div>

      <div className="flex items-center gap-3">
        <button
          type="button"
          onClick={runTest}
          className="w-fit rounded-lg border-(--line) border-[0.5px] border-solid bg-(--card-alt) px-3 py-2 text-[length:var(--text-caption)] text-(--fg) hover:border-(--line-strong) cursor-pointer"
        >
          Send test notification
        </button>
        {saving && <span className="text-[length:var(--text-pill)] text-(--text-faint)">Saving…</span>}
      </div>
      {testMsg && <p className="text-[length:var(--text-pill)] text-(--text-faint)">{testMsg}</p>}
    </div>
  );
}
