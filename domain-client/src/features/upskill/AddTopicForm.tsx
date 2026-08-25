import { useRef, useState, type KeyboardEvent } from "react";
import { CalendarPlus } from "lucide-react";

interface AddTopicFormProps {
  onAdd: (input: { name: string; targetDate: string | null }) => void;
}

export function AddTopicForm({ onAdd }: AddTopicFormProps) {
  const [name, setName] = useState("");
  const [targetDate, setTargetDate] = useState("");
  const [showDate, setShowDate] = useState(false);
  const inputRef = useRef<HTMLInputElement>(null);

  function submit() {
    const trimmed = name.trim();
    if (!trimmed) return;

    onAdd({ name: trimmed, targetDate: targetDate || null });

    setName("");
    setTargetDate("");
    setShowDate(false);
    inputRef.current?.focus();
  }

  function handleKeyDown(e: KeyboardEvent<HTMLInputElement>) {
    if (e.key === "Enter") {
      e.preventDefault();
      submit();
    }
  }

  return (
    <div className="group bg-(--card) border-(--line) border-[0.5px] border-solid rounded-xl px-4 py-3 mb-4">
      <input
        ref={inputRef}
        value={name}
        onChange={(e) => setName(e.target.value)}
        onKeyDown={handleKeyDown}
        placeholder="Add a topic..."
        className="w-full bg-transparent text-[length:var(--text-caption)] text-(--fg) placeholder:text-(--text-faint) focus:outline-none"
      />

      <div className="flex items-center gap-1 mt-1">
        <button
          type="button"
          onClick={() => setShowDate((v) => !v)}
          aria-label="Set target date"
          className="flex items-center justify-center size-6 rounded-md text-(--text-faint) opacity-0 group-focus-within:opacity-60 hover:opacity-100 transition-opacity cursor-pointer"
        >
          <CalendarPlus size={13} />
        </button>

        {showDate && (
          <input
            type="date"
            value={targetDate}
            onChange={(e) => setTargetDate(e.target.value)}
            className="bg-(--card-alt) border-(--line) border-[0.5px] border-solid rounded-lg px-2 py-1 text-[length:var(--text-pill)] text-(--fg) focus:outline-none"
          />
        )}
      </div>
    </div>
  );
}
