package diary

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

func validateDate(field, date string) error {
	if _, err := time.ParseInLocation(EntryDateLayout, date, IST); err != nil {
		return apperr.InvalidInput(field + " must be in YYYY-MM-DD format")
	}
	return nil
}

// withLocked stamps e.Locked fresh from the current instant — Locked is
// never persisted, so every read computes it rather than trusting a value
// that could go stale between the row being written and being served.
func withLocked(e Entry) Entry {
	e.Locked = IsLocked(e.EntryDate, time.Now().UTC())
	return e
}

func (s *Service) GetByDate(ctx context.Context, date string) (Entry, error) {
	if err := validateDate("date", date); err != nil {
		return Entry{}, err
	}

	e, err := s.repo.GetByDate(ctx, date)
	if err != nil {
		return Entry{}, err
	}
	return withLocked(e), nil
}

// Upsert is the real edit-lock guard: the server is the source of truth,
// not just the frontend disabling its editor. A date that already has an
// entry is updated in place ("one entry per day", edited over time), not
// rejected as a duplicate — until its 24-hour grace window (see IsLocked)
// closes, at which point every write is rejected regardless of whether an
// entry exists yet.
func (s *Service) Upsert(ctx context.Context, date string, content string) (Entry, error) {
	if err := validateDate("date", date); err != nil {
		return Entry{}, err
	}
	if IsLocked(date, time.Now().UTC()) {
		return Entry{}, apperr.Conflict("this entry's edit window has closed")
	}

	e, err := s.repo.Upsert(ctx, date, content)
	if err != nil {
		return Entry{}, err
	}
	return withLocked(e), nil
}

func (s *Service) ListDates(ctx context.Context, from, to string) ([]string, error) {
	if err := validateDate("from", from); err != nil {
		return nil, err
	}
	if err := validateDate("to", to); err != nil {
		return nil, err
	}
	if to < from {
		return nil, apperr.InvalidInput("to must not be before from")
	}

	return s.repo.ListDates(ctx, from, to)
}
