import { useSpotifyPlayer } from "./SpotifyPlayerProvider";

export function DeviceChips() {
  const { devices, thisDeviceId, transferTo } = useSpotifyPlayer();

  if (devices.length === 0) {
    return <p className="text-[10px] text-(--text-faint)">No devices found — open Spotify somewhere first.</p>;
  }

  return (
    <div className="flex flex-wrap gap-1.5">
      {devices.map((d) => (
        <button
          key={d.id}
          type="button"
          onClick={() => transferTo(d.id)}
          className={`rounded-full border-[0.5px] border-solid px-2 py-1 text-[9px] transition-colors ${
            d.isActive
              ? "border-(--green) text-(--green) bg-(--green-bg)"
              : "border-(--line-strong) text-(--text-muted) hover:text-(--fg)"
          }`}
        >
          {d.id === thisDeviceId ? "This browser" : d.name}
        </button>
      ))}
    </div>
  );
}
