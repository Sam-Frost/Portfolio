import { useEffect, useRef, useState } from "react";
import { LayoutGrid, Plus, X } from "lucide-react";
import {
  createTab,
  createTask,
  deleteTab,
  deleteTask,
  fetchTabs,
  fetchTasks,
  renameTab,
  updateTask,
} from "./api";
import { WorkTaskItem } from "./WorkTaskItem";
import { OverviewPanel } from "./OverviewPanel";
import type { WorkTab, WorkTask } from "./types";

const OVERVIEW = "__overview__";

export function WorkProfilePage() {
  const [tabs, setTabs] = useState<WorkTab[]>([]);
  const [active, setActive] = useState<string>(OVERVIEW);
  const [tasks, setTasks] = useState<WorkTask[]>([]);
  const [loadingTasks, setLoadingTasks] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [overviewKey, setOverviewKey] = useState(0);

  const [addingTab, setAddingTab] = useState(false);
  const [newTabName, setNewTabName] = useState("");
  const [renamingId, setRenamingId] = useState<string | null>(null);

  const [taskName, setTaskName] = useState("");
  const [taskDate, setTaskDate] = useState("");
  const newTabRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    fetchTabs()
      .then(setTabs)
      .catch((e) => setError(e instanceof Error ? e.message : "Couldn't load tabs."));
  }, []);

  useEffect(() => {
    if (active === OVERVIEW) return;
    setLoadingTasks(true);
    fetchTasks(active)
      .then(setTasks)
      .catch((e) => setError(e instanceof Error ? e.message : "Couldn't load tasks."))
      .finally(() => setLoadingTasks(false));
  }, [active]);

  useEffect(() => {
    if (addingTab) newTabRef.current?.focus();
  }, [addingTab]);

  function bumpOverview() {
    setOverviewKey((k) => k + 1);
  }

  async function submitNewTab() {
    const name = newTabName.trim();
    if (!name) {
      setAddingTab(false);
      return;
    }
    try {
      const tab = await createTab(name);
      setTabs((prev) => [...prev, tab]);
      setNewTabName("");
      setAddingTab(false);
      setActive(tab.id);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Couldn't create the tab.");
    }
  }

  async function submitRename(id: string, name: string) {
    setRenamingId(null);
    const trimmed = name.trim();
    const current = tabs.find((t) => t.id === id);
    if (!trimmed || trimmed === current?.name) return;
    try {
      const updated = await renameTab(id, trimmed);
      setTabs((prev) => prev.map((t) => (t.id === id ? updated : t)));
    } catch (e) {
      setError(e instanceof Error ? e.message : "Couldn't rename the tab.");
    }
  }

  async function removeTab(id: string) {
    if (!confirm("Delete this tab and all its tasks?")) return;
    try {
      await deleteTab(id);
      setTabs((prev) => prev.filter((t) => t.id !== id));
      if (active === id) setActive(OVERVIEW);
      bumpOverview();
    } catch (e) {
      setError(e instanceof Error ? e.message : "Couldn't delete the tab.");
    }
  }

  async function addTask() {
    const name = taskName.trim();
    if (!name || active === OVERVIEW) return;
    try {
      const task = await createTask(active, {
        name,
        description: null,
        targetDate: taskDate || null,
      });
      setTasks((prev) => [...prev, task]);
      setTaskName("");
      setTaskDate("");
      bumpOverview();
    } catch (e) {
      setError(e instanceof Error ? e.message : "Couldn't add the task.");
    }
  }

  function patchTask(id: string, patch: Parameters<typeof updateTask>[1], optimistic: Partial<WorkTask>) {
    const original = tasks.find((t) => t.id === id);
    setTasks((prev) => prev.map((t) => (t.id === id ? { ...t, ...optimistic } : t)));
    updateTask(id, patch)
      .then((updated) => setTasks((prev) => prev.map((t) => (t.id === id ? updated : t))))
      .catch((e) => {
        if (original) setTasks((prev) => prev.map((t) => (t.id === id ? original : t)));
        setError(e instanceof Error ? e.message : "Couldn't update the task.");
      })
      .finally(bumpOverview);
  }

  function markDone(id: string) {
    patchTask(id, { done: true, jiraAcknowledged: true }, { done: true });
  }

  function undo(id: string) {
    patchTask(id, { done: false }, { done: false, completedAt: null });
  }

  async function removeTask(id: string) {
    const original = tasks;
    setTasks((prev) => prev.filter((t) => t.id !== id));
    try {
      await deleteTask(id);
      bumpOverview();
    } catch (e) {
      setTasks(original);
      setError(e instanceof Error ? e.message : "Couldn't delete the task.");
    }
  }

  const activeTasks = tasks.filter((t) => !t.done);
  const doneTasks = tasks.filter((t) => t.done);

  const tabBtn = (isActive: boolean) =>
    `shrink-0 rounded-lg px-3 py-1.5 text-[length:var(--text-pill)] transition-colors cursor-pointer ${
      isActive ? "bg-(--card-alt) text-(--fg)" : "text-(--text-muted) hover:text-(--fg) hover:bg-(--card-alt)"
    }`;
  const inputCls =
    "rounded-lg bg-(--card-alt) border-(--line) border-[0.5px] border-solid px-2.5 py-1.5 text-[length:var(--text-caption)] text-(--fg) focus:outline-none";

  return (
    <div className="min-h-full lg:h-full flex flex-col">
      {error && (
        <div className="mb-3 flex items-center justify-between gap-2 rounded-lg border-(--line) border-[0.5px] border-solid bg-(--card) px-3 py-2 text-[length:var(--text-pill)] text-red-400">
          <span>{error}</span>
          <button onClick={() => setError(null)} aria-label="Dismiss error" className="shrink-0 cursor-pointer">
            <X size={12} />
          </button>
        </div>
      )}

      {/* Tab bar */}
      <div className="shrink-0 flex items-center gap-1 overflow-x-auto themed-scrollbar pb-2 mb-3 border-b-[0.5px] border-(--line-soft)">
        <button className={tabBtn(active === OVERVIEW)} onClick={() => setActive(OVERVIEW)}>
          <span className="flex items-center gap-1.5">
            <LayoutGrid size={13} /> Overview
          </span>
        </button>

        {tabs.map((tab) => (
          <div key={tab.id} className="group/tab relative shrink-0 flex items-center">
            {renamingId === tab.id ? (
              <input
                autoFocus
                defaultValue={tab.name}
                onBlur={(e) => submitRename(tab.id, e.target.value)}
                onKeyDown={(e) => {
                  if (e.key === "Enter") submitRename(tab.id, e.currentTarget.value);
                  if (e.key === "Escape") setRenamingId(null);
                }}
                className={`${inputCls} w-32 py-1`}
              />
            ) : (
              <button
                className={tabBtn(active === tab.id)}
                onClick={() => setActive(tab.id)}
                onDoubleClick={() => setRenamingId(tab.id)}
              >
                {tab.name}
              </button>
            )}
            <button
              onClick={() => removeTab(tab.id)}
              aria-label={`Delete ${tab.name}`}
              className="ml-0.5 flex items-center justify-center size-5 rounded text-(--text-faint) hover:text-red-400 transition-colors cursor-pointer opacity-0 group-hover/tab:opacity-100 focus-visible:opacity-100"
            >
              <X size={12} />
            </button>
          </div>
        ))}

        {addingTab ? (
          <input
            ref={newTabRef}
            value={newTabName}
            onChange={(e) => setNewTabName(e.target.value)}
            onBlur={submitNewTab}
            onKeyDown={(e) => {
              if (e.key === "Enter") submitNewTab();
              if (e.key === "Escape") {
                setAddingTab(false);
                setNewTabName("");
              }
            }}
            placeholder="Tab name"
            className={`${inputCls} w-32 py-1`}
          />
        ) : (
          <button
            onClick={() => setAddingTab(true)}
            aria-label="Add tab"
            className="shrink-0 flex items-center justify-center size-7 rounded-lg text-(--text-muted) hover:text-(--fg) hover:bg-(--card-alt) transition-colors cursor-pointer"
          >
            <Plus size={15} />
          </button>
        )}
      </div>

      {/* Body */}
      <div className="flex-1 lg:min-h-0 lg:overflow-y-auto themed-scrollbar lg:pr-2">
        {active === OVERVIEW ? (
          <OverviewPanel refreshKey={overviewKey} />
        ) : (
          <div className="flex flex-col gap-3 max-w-2xl">
            <div className="flex items-center gap-2">
              <input
                value={taskName}
                onChange={(e) => setTaskName(e.target.value)}
                onKeyDown={(e) => e.key === "Enter" && addTask()}
                placeholder="New task"
                className={`${inputCls} flex-1`}
              />
              <input
                type="date"
                value={taskDate}
                onChange={(e) => setTaskDate(e.target.value)}
                className={`${inputCls} shrink-0`}
              />
              <button
                onClick={addTask}
                className="shrink-0 rounded-lg bg-(--fg) text-(--bg) px-3 py-1.5 text-[length:var(--text-pill)] cursor-pointer"
              >
                Add
              </button>
            </div>

            {loadingTasks ? (
              <p className="text-[length:var(--text-pill)] text-(--text-faint)">Loading…</p>
            ) : tasks.length === 0 ? (
              <p className="text-[length:var(--text-pill)] text-(--text-faint)">No tasks in this tab yet.</p>
            ) : (
              <>
                <div className="flex flex-col gap-1.5">
                  {activeTasks.map((task) => (
                    <WorkTaskItem
                      key={task.id}
                      task={task}
                      onMarkDone={markDone}
                      onUndo={undo}
                      onDelete={removeTask}
                    />
                  ))}
                </div>
                {doneTasks.length > 0 && (
                  <details className="mt-2">
                    <summary className="cursor-pointer text-[length:var(--text-pill)] text-(--text-faint) mb-1.5">
                      Completed ({doneTasks.length})
                    </summary>
                    <div className="flex flex-col gap-1.5">
                      {doneTasks.map((task) => (
                        <WorkTaskItem
                          key={task.id}
                          task={task}
                          onMarkDone={markDone}
                          onUndo={undo}
                          onDelete={removeTask}
                        />
                      ))}
                    </div>
                  </details>
                )}
              </>
            )}
          </div>
        )}
      </div>
    </div>
  );
}
