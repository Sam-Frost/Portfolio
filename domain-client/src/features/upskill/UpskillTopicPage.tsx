import { useEffect, useState } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import { ArrowLeft, X } from "lucide-react";
import { ApiError } from "../../lib/apiClient";
import {
  addResource,
  createSubtopic,
  deleteResource,
  deleteSubtopic,
  fetchSubtopics,
  fetchTopic,
  updateSubtopic,
} from "./api";
import { AddSubtopicForm } from "./AddSubtopicForm";
import { SubtopicItem } from "./SubtopicItem";
import { UpskillPieChart } from "./UpskillPieChart";
import type { ResourceInput, Subtopic, Topic } from "./types";

function formatDate(value: string) {
  return new Date(value).toLocaleDateString(undefined, { year: "numeric", month: "short", day: "numeric" });
}

// Oldest first, by the date each subtopic was added.
const byDateAdded = (a: Subtopic, b: Subtopic) => a.dateAdded.localeCompare(b.dateAdded);

export function UpskillTopicPage() {
  const { topicId } = useParams<{ topicId: string }>();
  const navigate = useNavigate();
  const [topic, setTopic] = useState<Topic | null>(null);
  const [subtopics, setSubtopics] = useState<Subtopic[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!topicId) return;
    setLoading(true);
    Promise.all([fetchTopic(topicId), fetchSubtopics(topicId)])
      .then(([fetchedTopic, fetchedSubtopics]) => {
        setTopic(fetchedTopic);
        setSubtopics([...fetchedSubtopics].sort(byDateAdded));
      })
      .catch((err) => {
        if (err instanceof ApiError && err.status === 404) {
          navigate("/upskill", { replace: true });
          return;
        }
        setError(err instanceof Error ? err.message : "Couldn't load this topic.");
      })
      .finally(() => setLoading(false));
  }, [topicId, navigate]);

  function handleCreateSubtopic(input: { name: string; targetDate: string | null; resources: ResourceInput[] }) {
    if (!topicId) return;
    createSubtopic(topicId, input)
      .then((subtopic) => {
        setSubtopics((prev) => [...prev, subtopic].sort(byDateAdded));
        setTopic((prev) => (prev ? { ...prev, subtopicCount: prev.subtopicCount + 1 } : prev));
      })
      .catch((err) => setError(err instanceof Error ? err.message : "Couldn't add subtopic."));
  }

  function handleMarkDone(id: string) {
    setSubtopics((prev) => prev.map((s) => (s.id === id ? { ...s, done: true } : s)));
    setTopic((prev) => (prev ? { ...prev, doneCount: prev.doneCount + 1 } : prev));
    updateSubtopic(id, { done: true }).catch((err) => {
      setSubtopics((prev) => prev.map((s) => (s.id === id ? { ...s, done: false } : s)));
      setTopic((prev) => (prev ? { ...prev, doneCount: prev.doneCount - 1 } : prev));
      setError(err instanceof Error ? err.message : "Couldn't update subtopic.");
    });
  }

  function handleUpdateSubtopic(id: string, input: { name: string; targetDate: string | null }) {
    const original = subtopics.find((s) => s.id === id);
    setSubtopics((prev) => prev.map((s) => (s.id === id ? { ...s, ...input } : s)));
    updateSubtopic(id, input).catch((err) => {
      if (original) setSubtopics((prev) => prev.map((s) => (s.id === id ? original : s)));
      setError(err instanceof Error ? err.message : "Couldn't update subtopic.");
    });
  }

  function handleDeleteSubtopic(id: string) {
    const original = subtopics;
    const wasDone = subtopics.find((s) => s.id === id)?.done ?? false;
    setSubtopics((prev) => prev.filter((s) => s.id !== id));
    setTopic((prev) =>
      prev ? { ...prev, subtopicCount: prev.subtopicCount - 1, doneCount: prev.doneCount - (wasDone ? 1 : 0) } : prev,
    );
    deleteSubtopic(id).catch((err) => {
      setSubtopics(original);
      setTopic((prev) =>
        prev ? { ...prev, subtopicCount: prev.subtopicCount + 1, doneCount: prev.doneCount + (wasDone ? 1 : 0) } : prev,
      );
      setError(err instanceof Error ? err.message : "Couldn't delete subtopic.");
    });
  }

  function handleAddResource(subtopicId: string, input: ResourceInput) {
    addResource(subtopicId, input)
      .then((resource) => {
        setSubtopics((prev) =>
          prev.map((s) => (s.id === subtopicId ? { ...s, resources: [...s.resources, resource] } : s)),
        );
      })
      .catch((err) => setError(err instanceof Error ? err.message : "Couldn't add resource."));
  }

  function handleRemoveResource(subtopicId: string, resourceId: string) {
    const original = subtopics;
    setSubtopics((prev) =>
      prev.map((s) => (s.id === subtopicId ? { ...s, resources: s.resources.filter((r) => r.id !== resourceId) } : s)),
    );
    deleteResource(resourceId).catch((err) => {
      setSubtopics(original);
      setError(err instanceof Error ? err.message : "Couldn't remove resource.");
    });
  }

  if (loading) {
    return (
      <div className="bg-(--card) border-(--line) border-[0.5px] border-solid rounded-xl p-8 text-center text-(--text-faint) text-[length:var(--text-caption)]">
        Loading...
      </div>
    );
  }

  if (!topic) return null;

  return (
    <div className="h-full flex flex-col">
      <div className="shrink-0">
        <Link
          to="/upskill"
          className="inline-flex items-center gap-1.5 mb-3 text-[length:var(--text-pill)] text-(--text-muted) hover:text-(--fg) transition-colors"
        >
          <ArrowLeft size={13} />
          All topics
        </Link>

        <div className="flex items-start justify-between gap-4 mb-4">
          <div>
            <h1 className="font-space font-semibold text-(--fg) text-[length:var(--text-caption)]">{topic.name}</h1>
            {topic.targetDate && (
              <p className="text-[length:var(--text-pill)] text-(--text-faint) mt-0.5">Due {formatDate(topic.targetDate)}</p>
            )}
          </div>
        </div>

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
      </div>

      <div className="flex-1 min-h-0 flex gap-4">
        <div className="flex-1 min-w-0 overflow-y-auto themed-scrollbar">
          {subtopics.length === 0 && (
            <div className="bg-(--card) border-(--line) border-[0.5px] border-solid rounded-xl p-8 text-center text-(--text-faint) text-[length:var(--text-caption)]">
              No subtopics yet. Add one on the right.
            </div>
          )}

          {subtopics.length > 0 && (
            <div className="grid grid-cols-1 gap-1.5 pb-2">
              {subtopics.map((subtopic) => (
                <SubtopicItem
                  key={subtopic.id}
                  subtopic={subtopic}
                  onMarkDone={handleMarkDone}
                  onUpdate={handleUpdateSubtopic}
                  onDelete={handleDeleteSubtopic}
                  onAddResource={handleAddResource}
                  onRemoveResource={handleRemoveResource}
                />
              ))}
            </div>
          )}
        </div>

        <div className="shrink-0 w-80 flex flex-col overflow-y-auto themed-scrollbar">
          <AddSubtopicForm onAdd={handleCreateSubtopic} />

          <div className="bg-(--card) border-(--line) border-[0.5px] border-solid rounded-xl px-4 py-3">
            <UpskillPieChart done={topic.doneCount} total={topic.subtopicCount} />
          </div>
        </div>
      </div>
    </div>
  );
}
