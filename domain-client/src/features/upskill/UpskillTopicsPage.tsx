import { useEffect, useState } from "react";
import { X } from "lucide-react";
import { createTopic, deleteTopic, fetchTopics, updateTopic } from "./api";
import { AddTopicForm } from "./AddTopicForm";
import { TopicCard } from "./TopicCard";
import type { Topic } from "./types";

export function UpskillTopicsPage() {
  const [topics, setTopics] = useState<Topic[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    fetchTopics()
      .then(setTopics)
      .catch((err) => setError(err instanceof Error ? err.message : "Couldn't load topics."))
      .finally(() => setLoading(false));
  }, []);

  function handleCreate(input: { name: string; targetDate: string | null }) {
    createTopic(input)
      .then((topic) => setTopics((prev) => [topic, ...prev]))
      .catch((err) => setError(err instanceof Error ? err.message : "Couldn't add topic."));
  }

  function handleUpdate(id: string, input: { name: string; targetDate: string | null }) {
    const original = topics.find((t) => t.id === id);
    setTopics((prev) => prev.map((t) => (t.id === id ? { ...t, ...input } : t)));
    updateTopic(id, input).catch((err) => {
      if (original) setTopics((prev) => prev.map((t) => (t.id === id ? original : t)));
      setError(err instanceof Error ? err.message : "Couldn't update topic.");
    });
  }

  function handleDelete(id: string) {
    const original = topics;
    setTopics((prev) => prev.filter((t) => t.id !== id));
    deleteTopic(id).catch((err) => {
      setTopics(original);
      setError(err instanceof Error ? err.message : "Couldn't delete topic.");
    });
  }

  return (
    <div className="h-full flex flex-col">
      <div className="shrink-0">
        <AddTopicForm onAdd={handleCreate} />

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

      <div className="flex-1 min-h-0 overflow-y-auto themed-scrollbar">
        {loading && (
          <div className="bg-(--card) border-(--line) border-[0.5px] border-solid rounded-xl p-8 text-center text-(--text-faint) text-[length:var(--text-caption)]">
            Loading topics...
          </div>
        )}

        {!loading && topics.length === 0 && (
          <div className="bg-(--card) border-(--line) border-[0.5px] border-solid rounded-xl p-8 text-center text-(--text-faint) text-[length:var(--text-caption)]">
            No topics yet. Add one above to start tracking what you're upskilling in.
          </div>
        )}

        {!loading && topics.length > 0 && (
          <div className="grid grid-cols-1 md:grid-cols-2 gap-1.5 pb-2">
            {topics.map((topic) => (
              <TopicCard key={topic.id} topic={topic} onUpdate={handleUpdate} onDelete={handleDelete} />
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
