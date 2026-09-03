import { useEffect, useState } from "react";
import { CalendarClock, CalendarX2 } from "lucide-react";
import { fetchOverview } from "./api";
import type { WorkOverview, WorkTaskWithTab } from "./types";

function Row({ task }: { task: WorkTaskWithTab }) {
  return (
    <li className="flex items-center gap-2 rounded-lg bg-(--card) border-(--line) border-[0.5px] border-solid px-3 py-2">
      <span className="flex-1 min-w-0 truncate text-[length:var(--text-caption)] sm:text-[length:var(--text-pill)] text-(--fg)">
        {task.name}
      </span>
      <span className="shrink-0 rounded-md bg-(--card-alt) border-(--line) border-[0.5px] border-solid px-1.5 py-0.5 text-[length:var(--text-pill)] text-(--text-muted)">
        {task.tabName}
      </span>
      {task.targetDate && (
        <span className="shrink-0 text-[length:var(--text-pill)] text-(--text-faint)">
          {new Date(task.targetDate).toLocaleDateString(undefined, { month: "short", day: "numeric" })}
        </span>
      )}
    </li>
  );
}

function Group({
  title,
  icon,
  tasks,
  empty,
}: {
  title: string;
  icon: React.ReactNode;
  tasks: WorkTaskWithTab[];
  empty: string;
}) {
  return (
    <section>
      <h3 className="flex items-center gap-1.5 text-[length:var(--text-pill)] font-medium uppercase tracking-wide text-(--text-faint) mb-2">
        {icon}
        {title} ({tasks.length})
      </h3>
      {tasks.length === 0 ? (
        <p className="text-[length:var(--text-pill)] text-(--text-faint)">{empty}</p>
      ) : (
        <ul className="flex flex-col gap-1.5">
          {tasks.map((t) => (
            <Row key={t.id} task={t} />
          ))}
        </ul>
      )}
    </section>
  );
}

export function OverviewPanel({ refreshKey }: { refreshKey: number }) {
  const [data, setData] = useState<WorkOverview | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    fetchOverview()
      .then(setData)
      .catch((e) => setError(e instanceof Error ? e.message : "Couldn't load the overview."));
  }, [refreshKey]);

  if (error) return <p className="text-[length:var(--text-pill)] text-red-400">{error}</p>;
  if (!data) return <p className="text-[length:var(--text-pill)] text-(--text-faint)">Loading…</p>;

  return (
    <div className="flex flex-col gap-6 max-w-xl">
      <Group
        title="Due today"
        icon={<CalendarClock size={12} />}
        tasks={data.dueToday}
        empty="Nothing due today."
      />
      <Group
        title="Past due date"
        icon={<CalendarX2 size={12} />}
        tasks={data.overdue}
        empty="Nothing overdue."
      />
    </div>
  );
}
