package reminder

import (
	"context"
	"time"

	"github.com/Sam-Frost/portfolio/internal/apperr"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListByTodo(ctx context.Context, todoID string) ([]Reminder, error) {
	return s.repo.ListByTodo(ctx, todoID)
}

func (s *Service) Create(ctx context.Context, todoID string, input CreateInput) (Reminder, error) {
	if todoID == "" {
		return Reminder{}, apperr.InvalidInput("todoId is required")
	}

	rem := Reminder{TodoID: todoID, Kind: input.Kind}
	now := time.Now().UTC()

	switch input.Kind {
	case KindOnce:
		if input.FireAt == nil || *input.FireAt == "" {
			return Reminder{}, apperr.InvalidInput("fireAt is required for a one-time reminder")
		}
		t, err := time.Parse(time.RFC3339, *input.FireAt)
		if err != nil {
			return Reminder{}, apperr.InvalidInput("fireAt must be an RFC3339 timestamp")
		}
		if !t.After(now) {
			return Reminder{}, apperr.InvalidInput("fireAt must be in the future")
		}
		rem.FireAt = t.UTC()

	case KindRepeat:
		if input.IntervalSeconds == nil {
			return Reminder{}, apperr.InvalidInput("intervalSeconds is required for a repeating reminder")
		}
		if *input.IntervalSeconds < MinIntervalSeconds {
			return Reminder{}, apperr.InvalidInput("intervalSeconds must be at least 60")
		}
		rem.IntervalSeconds = input.IntervalSeconds
		rem.FireAt = now.Add(time.Duration(*input.IntervalSeconds) * time.Second)

	default:
		return Reminder{}, apperr.InvalidInput(`kind must be "once" or "repeat"`)
	}

	return s.repo.Create(ctx, rem)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	if id == "" {
		return apperr.InvalidInput("id is required")
	}
	return s.repo.Delete(ctx, id)
}

// DueBefore / Fired / Reschedule are the scheduler's surface.

func (s *Service) DueBefore(ctx context.Context, t time.Time) ([]Due, error) {
	return s.repo.DueBefore(ctx, t)
}

// Settle advances a reminder after it has fired: a repeating one is pushed
// to its next occurrence (catching up past any missed ticks so it doesn't
// fire in a burst), a one-time one is deleted.
func (s *Service) Settle(ctx context.Context, d Due, now time.Time) error {
	if d.Kind != KindRepeat || d.IntervalSeconds == nil {
		return s.repo.Delete(ctx, d.ID)
	}
	interval := time.Duration(*d.IntervalSeconds) * time.Second
	next := d.FireAt.Add(interval)
	for !next.After(now) {
		next = next.Add(interval)
	}
	return s.repo.Reschedule(ctx, d.ID, next)
}
