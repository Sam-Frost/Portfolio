import { Check } from "lucide-react";
import type { Goal } from "./types";

interface GoalChecklistProps {
  goals: Goal[];
  onToggle: (index: number) => void;
  disabled?: boolean;
}

// The goal bullets set at start, each toggleable done/not-done — shown in
// both the complete and cancel dialogs so progress can be recorded either
// way a session ends.
export function GoalChecklist({ goals, onToggle, disabled }: GoalChecklistProps) {
  return (
    <ul className="flex flex-col gap-1.5">
      {goals.map((goal, i) => (
        <li key={i}>
          <button
            type="button"
            onClick={() => onToggle(i)}
            disabled={disabled}
            aria-pressed={goal.done}
            className="flex w-full items-center gap-2 text-left cursor-pointer disabled:cursor-not-allowed group"
          >
            <span
              className={`shrink-0 flex h-4 w-4 items-center justify-center rounded border-[0.5px] border-solid transition-colors ${
                goal.done
                  ? "bg-(--green) border-(--green) text-(--bg)"
                  : "border-(--line-strong) text-transparent group-hover:border-(--text-muted)"
              }`}
            >
              <Check size={11} strokeWidth={3} />
            </span>
            <span
              className={`text-[length:var(--text-caption)] ${
                goal.done ? "text-(--text-faint) line-through" : "text-(--fg)"
              }`}
            >
              {goal.text}
            </span>
          </button>
        </li>
      ))}
    </ul>
  );
}
