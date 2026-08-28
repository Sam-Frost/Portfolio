const UNITS = ["B", "KB", "MB", "GB", "TB"];

// formatBytes renders a byte count as a short human string, e.g. "2.4 MB".
export function formatBytes(bytes: number): string {
  if (bytes <= 0) return "0 B";
  const exp = Math.min(Math.floor(Math.log(bytes) / Math.log(1024)), UNITS.length - 1);
  const value = bytes / 1024 ** exp;
  return `${value >= 10 || exp === 0 ? Math.round(value) : value.toFixed(1)} ${UNITS[exp]}`;
}
