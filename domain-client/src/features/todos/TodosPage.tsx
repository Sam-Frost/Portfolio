import { useEffect, useRef, useState } from "react";
import { ArrowDownWideNarrow, ArrowUpWideNarrow, SlidersHorizontal, Tag, TagX, X } from "lucide-react";
import { Toast } from "../../components/Toast";
import { fetchLabels } from "../labels/api";
import { LABEL_COLOR_VAR } from "../labels/colors";
import type { Label } from "../labels/types";
import { createTodo, fetchTodos, setTodoDone, updateTodo } from "./api";
import { AddTodoForm } from "./AddTodoForm";
import { DueTodayPanel } from "./DueTodayPanel";
import { LabelCountChart } from "./LabelCountChart";
import { LabelDistributionPie } from "./LabelDistributionPie";
import { TodoItem } from "./TodoItem";
import type { SortField, Todo } from "./types";

type Tab = "active" | "completed";

// `tab` restricts an option to one list — "Done date" only makes sense for
// completed todos (active ones have no completedAt).
const SORT_OPTIONS: { field: SortField; label: string; tab?: Tab }[] = [
  { field: "dateAdded", label: "Date added" },
  { field: "targetDate", label: "Target date" },
  { field: "completedAt", label: "Done date", tab: "completed" },
];

// Sentinel labelFilter value for "todos with no label assigned", distinct
// from null (no filter, i.e. "All labels"). Real label ids are 32-char hex
// (see internal/id.New on the backend), so this can never collide.
const NO_LABEL_FILTER = "none";

function labelCounts(todos: Todo[], labelId: string | null) {
  const matches = todos.filter((t) => t.labelId === labelId);
  return { active: matches.filter((t) => !t.done).length, completed: matches.filter((t) => t.done).length };
}

export function TodosPage() {
  const [todos, setTodos] = useState<Todo[]>([]);
  const [labels, setLabels] = useState<Label[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [sortField, setSortField] = useState<SortField>("dateAdded");
  const [ascending, setAscending] = useState(false);
  const [showSortMenu, setShowSortMenu] = useState(false);
  const [labelFilter, setLabelFilter] = useState<string | null>(null);
  // Label preselected in the add-todo bar. Kept in sync with labelFilter:
  // choosing a label to view by also preselects it for new todos. "All
  // labels" (null) and "No label" both mean "no specific label" here.
  const [addLabelId, setAddLabelId] = useState<string | null>(null);
  const [showLabelMenu, setShowLabelMenu] = useState(false);
  const [tab, setTab] = useState<Tab>("active");
  const [toast, setToast] = useState<{ key: number; message: string } | null>(null);
  const sortMenuRef = useRef<HTMLDivElement>(null);
  const labelMenuRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!showSortMenu) return;

    function handlePointerDown(e: MouseEvent) {
      if (sortMenuRef.current && !sortMenuRef.current.contains(e.target as Node)) {
        setShowSortMenu(false);
      }
    }

    document.addEventListener("mousedown", handlePointerDown);
    return () => document.removeEventListener("mousedown", handlePointerDown);
  }, [showSortMenu]);

  useEffect(() => {
    if (!showLabelMenu) return;

    function handlePointerDown(e: MouseEvent) {
      if (labelMenuRef.current && !labelMenuRef.current.contains(e.target as Node)) {
        setShowLabelMenu(false);
      }
    }

    document.addEventListener("mousedown", handlePointerDown);
    return () => document.removeEventListener("mousedown", handlePointerDown);
  }, [showLabelMenu]);

  useEffect(() => {
    fetchLabels()
      .then(setLabels)
      .catch((err) => setError(err instanceof Error ? err.message : "Couldn't load labels."));
  }, []);

  useEffect(() => {
    setLoading(true);
    // Label filtering happens client-side (see labelFilteredTodos below) so
    // the full list stays around to compute per-label counts in the menu.
    fetchTodos({ sortBy: sortField, order: ascending ? "asc" : "desc" })
      .then(setTodos)
      .catch((err) => setError(err instanceof Error ? err.message : "Couldn't load todos."))
      .finally(() => setLoading(false));
  }, [sortField, ascending]);

  function selectTab(next: Tab) {
    setTab(next);
    // "Done date" is a completed-only sort; fall back when leaving that tab.
    if (next !== "completed" && sortField === "completedAt") {
      setSortField("dateAdded");
    }
  }

  function selectLabelFilter(next: string | null) {
    setLabelFilter(next);
    setAddLabelId(next === null || next === NO_LABEL_FILTER ? null : next);
    setShowLabelMenu(false);
  }

  function handleCreate(input: {
    name: string;
    description: string | null;
    targetDate: string | null;
    labelId: string | null;
  }) {
    createTodo(input)
      .then((todo) => {
        setTodos((prev) => [todo, ...prev]);
        setToast((prev) => ({ key: (prev?.key ?? 0) + 1, message: "Todo added" }));
      })
      .catch((err) => setError(err instanceof Error ? err.message : "Couldn't add todo."));
  }

  function setDone(id: string, done: boolean) {
    const original = todos.find((t) => t.id === id);
    // Optimistically flip; the server owns completedAt (stamped on completion,
    // cleared on undo), so reconcile with its response once it lands.
    setTodos((prev) =>
      prev.map((t) => (t.id === id ? { ...t, done, completedAt: done ? t.completedAt : null } : t)),
    );
    setTodoDone(id, done)
      .then((updated) => setTodos((prev) => prev.map((t) => (t.id === id ? updated : t))))
      .catch((err) => {
        if (original) setTodos((prev) => prev.map((t) => (t.id === id ? original : t)));
        setError(err instanceof Error ? err.message : "Couldn't update todo.");
      });
  }

  function handleMarkDone(id: string) {
    setDone(id, true);
  }

  function handleUndo(id: string) {
    setDone(id, false);
  }

  function handleUpdate(
    id: string,
    input: { name: string; description: string | null; targetDate: string | null; labelId: string | null },
  ) {
    const original = todos.find((t) => t.id === id);
    setTodos((prev) => prev.map((t) => (t.id === id ? { ...t, ...input } : t)));
    updateTodo(id, input).catch((err) => {
      if (original) {
        setTodos((prev) => prev.map((t) => (t.id === id ? original : t)));
      }
      setError(err instanceof Error ? err.message : "Couldn't update todo.");
    });
  }

  const labelFilteredTodos =
    labelFilter === null
      ? todos
      : labelFilter === NO_LABEL_FILTER
        ? todos.filter((t) => t.labelId === null)
        : todos.filter((t) => t.labelId === labelFilter);
  const activeTodos = labelFilteredTodos.filter((t) => !t.done);
  const completedTodos = labelFilteredTodos.filter((t) => t.done);
  const visibleTodos = tab === "active" ? activeTodos : completedTodos;
  const activeLabel = labelFilter && labelFilter !== NO_LABEL_FILTER ? labels.find((l) => l.id === labelFilter) ?? null : null;
  const isNoLabelFilter = labelFilter === NO_LABEL_FILTER;
  const noLabelCounts = labelCounts(todos, null);

  return (
    <div className="min-h-full lg:h-full flex flex-col">
      {toast && <Toast key={toast.key} message={toast.message} onDone={() => setToast(null)} />}

      <div className="flex-1 min-h-0 flex flex-col lg:flex-row gap-4">
        <div className="flex-1 min-w-0 flex flex-col lg:min-h-0">
          <div className="shrink-0">
            <AddTodoForm
              labels={labels}
              labelId={addLabelId}
              onLabelChange={setAddLabelId}
              onAdd={handleCreate}
            />

            {error && (
              <div className="mb-3 flex items-center justify-between gap-2 rounded-lg border-(--line) border-[0.5px] border-solid bg-(--card) px-3 py-2 text-[length:var(--text-pill)] text-red-400">
                <span>{error}</span>
                <button
                  onClick={() => setError(null)}
                  aria-label="Dismiss error"
                  className="shrink-0 text-(--text-faint) hover:text-(--fg) transition-colors cursor-pointer"
                >
                  <X size={12} />
                </button>
              </div>
            )}

            <div className="flex items-center justify-between gap-2 flex-wrap mb-3">
              <div className="flex items-center gap-1 rounded-lg bg-(--card-alt) p-1">
                <button
                  onClick={() => selectTab("active")}
                  className={`rounded-md px-3 py-1 text-[length:var(--text-pill)] transition-colors cursor-pointer ${
                    tab === "active" ? "bg-(--card) text-(--fg)" : "text-(--text-muted) hover:text-(--fg)"
                  }`}
                >
                  Active ({activeTodos.length})
                </button>
                <button
                  onClick={() => selectTab("completed")}
                  className={`rounded-md px-3 py-1 text-[length:var(--text-pill)] transition-colors cursor-pointer ${
                    tab === "completed" ? "bg-(--card) text-(--fg)" : "text-(--text-muted) hover:text-(--fg)"
                  }`}
                >
                  Completed ({completedTodos.length})
                </button>
              </div>

              <div className="flex items-center gap-1.5">
                <div className="relative" ref={labelMenuRef}>
                  <button
                    onClick={() => setShowLabelMenu((v) => !v)}
                    aria-label="Filter by label"
                    className={`flex items-center gap-1.5 h-7 rounded-lg px-2 text-[length:var(--text-pill)] transition-colors cursor-pointer ${
                      showLabelMenu || activeLabel || isNoLabelFilter
                        ? "bg-(--card-alt) text-(--fg)"
                        : "text-(--text-muted) hover:text-(--fg) hover:bg-(--card-alt)"
                    }`}
                  >
                    {activeLabel ? (
                      <>
                        <span
                          className="size-2 rounded-full shrink-0"
                          style={{ backgroundColor: LABEL_COLOR_VAR[activeLabel.color] }}
                        />
                        {activeLabel.name}
                      </>
                    ) : isNoLabelFilter ? (
                      <>
                        <TagX size={15} />
                        No label
                      </>
                    ) : (
                      <Tag size={15} />
                    )}
                  </button>

                  {showLabelMenu && (
                    <div className="absolute right-0 z-10 mt-2 w-52 max-h-56 overflow-y-auto rounded-lg border-(--line) border-[0.5px] border-solid bg-(--card) p-1.5 shadow-lg themed-scrollbar">
                      <button
                        onClick={() => selectLabelFilter(null)}
                        className={`flex w-full items-center rounded-lg px-3 py-1.5 text-left text-[length:var(--text-pill)] transition-colors cursor-pointer ${
                          !labelFilter
                            ? "bg-(--card-alt) text-(--fg)"
                            : "text-(--text-muted) hover:text-(--fg) hover:bg-(--card-alt)"
                        }`}
                      >
                        All labels
                      </button>
                      <button
                        onClick={() => selectLabelFilter(NO_LABEL_FILTER)}
                        className={`flex w-full items-center gap-1.5 rounded-lg px-3 py-1.5 text-left text-[length:var(--text-pill)] transition-colors cursor-pointer ${
                          isNoLabelFilter
                            ? "bg-(--card-alt) text-(--fg)"
                            : "text-(--text-muted) hover:text-(--fg) hover:bg-(--card-alt)"
                        }`}
                      >
                        <TagX size={12} className="shrink-0" />
                        <span className="truncate flex-1">No label</span>
                        <span className="shrink-0 text-(--text-faint)">
                          {tab === "active" ? noLabelCounts.active : noLabelCounts.completed}
                        </span>
                      </button>
                      {labels.map((l) => {
                        const counts = labelCounts(todos, l.id);
                        return (
                          <button
                            key={l.id}
                            onClick={() => selectLabelFilter(l.id)}
                            className={`flex w-full items-center gap-1.5 rounded-lg px-3 py-1.5 text-left text-[length:var(--text-pill)] transition-colors cursor-pointer ${
                              labelFilter === l.id
                                ? "bg-(--card-alt) text-(--fg)"
                                : "text-(--text-muted) hover:text-(--fg) hover:bg-(--card-alt)"
                            }`}
                          >
                            <span
                              className="size-2 rounded-full shrink-0"
                              style={{ backgroundColor: LABEL_COLOR_VAR[l.color] }}
                            />
                            <span className="truncate flex-1">{l.name}</span>
                            <span className="shrink-0 text-(--text-faint)">
                              {tab === "active" ? counts.active : counts.completed}
                            </span>
                          </button>
                        );
                      })}
                      {labels.length === 0 && (
                        <p className="px-3 py-1.5 text-[length:var(--text-pill)] text-(--text-faint)">No labels yet.</p>
                      )}
                    </div>
                  )}
                </div>

                <div className="relative" ref={sortMenuRef}>
                  <button
                    onClick={() => setShowSortMenu((v) => !v)}
                    aria-label="Sort options"
                    className={`flex items-center justify-center size-7 rounded-lg transition-colors cursor-pointer ${
                      showSortMenu ? "bg-(--card-alt) text-(--fg)" : "text-(--text-muted) hover:text-(--fg) hover:bg-(--card-alt)"
                    }`}
                  >
                    <SlidersHorizontal size={15} />
                  </button>

                  {showSortMenu && (
                    <div className="absolute right-0 z-10 mt-2 flex items-center gap-1.5 rounded-lg border-(--line) border-[0.5px] border-solid bg-(--card) p-1.5 shadow-lg">
                      {SORT_OPTIONS.filter((o) => !o.tab || o.tab === tab).map((option) => (
                        <button
                          key={option.field}
                          onClick={() => setSortField(option.field)}
                          className={`whitespace-nowrap rounded-lg px-3 py-1.5 text-[length:var(--text-pill)] transition-colors cursor-pointer ${
                            sortField === option.field
                              ? "bg-(--card-alt) text-(--fg)"
                              : "text-(--text-muted) hover:text-(--fg) hover:bg-(--card-alt)"
                          }`}
                        >
                          {option.label}
                        </button>
                      ))}
                      <button
                        onClick={() => setAscending((v) => !v)}
                        aria-label={ascending ? "Sort descending" : "Sort ascending"}
                        className="flex items-center justify-center size-7 rounded-lg text-(--text-muted) hover:text-(--fg) hover:bg-(--card-alt) transition-colors cursor-pointer"
                      >
                        {ascending ? <ArrowUpWideNarrow size={15} /> : <ArrowDownWideNarrow size={15} />}
                      </button>
                    </div>
                  )}
                </div>
              </div>
            </div>
          </div>

          <div className="flex-1 lg:min-h-0 lg:overflow-y-auto themed-scrollbar lg:pr-2">
            {loading && (
              <div className="bg-(--card) border-(--line) border-[0.5px] border-solid rounded-xl p-8 text-center text-(--text-faint) text-[length:var(--text-caption)]">
                Loading todos...
              </div>
            )}

            {!loading && visibleTodos.length === 0 && (
              <div className="bg-(--card) border-(--line) border-[0.5px] border-solid rounded-xl p-8 text-center text-(--text-faint) text-[length:var(--text-caption)]">
                {tab === "active" ? "No active todos." : "No completed todos yet."}
              </div>
            )}

            {!loading && visibleTodos.length > 0 && (
              <div className="grid grid-cols-1 md:grid-cols-2 gap-1.5 pb-2">
                {visibleTodos.map((todo) => (
                  <TodoItem
                    key={todo.id}
                    todo={todo}
                    labels={labels}
                    onMarkDone={handleMarkDone}
                    onUndo={handleUndo}
                    onUpdate={handleUpdate}
                  />
                ))}
              </div>
            )}
          </div>
        </div>

        <div className="lg:w-64 shrink-0 flex flex-col gap-4 lg:min-h-0 lg:overflow-y-auto themed-scrollbar">
          <LabelCountChart todos={todos} labels={labels} tab={tab} />
          <LabelDistributionPie todos={todos} labels={labels} tab={tab} />
          <DueTodayPanel todos={todos} labels={labels} />
        </div>
      </div>
    </div>
  );
}
