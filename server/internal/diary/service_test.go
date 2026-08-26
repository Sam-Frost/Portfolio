package diary

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Sam-Frost/portfolio/internal/apperr"
)

func assertKind(t *testing.T, err error, want apperr.Kind) {
	t.Helper()
	var appErr *apperr.Error
	if !errors.As(err, &appErr) || appErr.Kind != want {
		t.Fatalf("err = %v, want apperr kind %v", err, want)
	}
}

func TestIsLocked(t *testing.T) {
	// 2026-08-20's day is over at IST midnight starting 2026-08-21, +24h
	// lands the lock instant at 2026-08-22T00:00:00+05:30 — matches the
	// spec's own worked example exactly.
	lockAt := time.Date(2026, 8, 22, 0, 0, 0, 0, IST)

	cases := []struct {
		name string
		now  time.Time
		want bool
	}{
		{"just before lock instant", lockAt.Add(-time.Second), false},
		{"exactly at lock instant", lockAt, true},
		{"well after lock instant", lockAt.Add(48 * time.Hour), true},
		{"same day, still open", time.Date(2026, 8, 20, 23, 0, 0, 0, IST), false},
		{"next day, still within grace window", time.Date(2026, 8, 21, 12, 0, 0, 0, IST), false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := IsLocked("2026-08-20", tc.now.UTC()); got != tc.want {
				t.Errorf("IsLocked(%v) = %v, want %v", tc.now, got, tc.want)
			}
		})
	}
}

func TestService_GetByDateRejectsMalformedDate(t *testing.T) {
	svc := NewService(NewMemoryRepository())

	_, err := svc.GetByDate(context.Background(), "20-08-2026")
	assertKind(t, err, apperr.KindInvalidInput)
}

func TestService_UpsertRejectsMalformedDate(t *testing.T) {
	svc := NewService(NewMemoryRepository())

	_, err := svc.Upsert(context.Background(), "20-08-2026", "<p>hi</p>")
	assertKind(t, err, apperr.KindInvalidInput)
}

func TestService_UpsertCreatesThenUpdatesSameDay(t *testing.T) {
	svc := NewService(NewMemoryRepository())
	today := time.Now().In(IST).Format(EntryDateLayout)

	created, err := svc.Upsert(context.Background(), today, "<p>first</p>")
	if err != nil {
		t.Fatalf("Upsert: %v", err)
	}
	if created.Locked {
		t.Errorf("today's freshly-created entry should not be locked")
	}

	updated, err := svc.Upsert(context.Background(), today, "<p>second</p>")
	if err != nil {
		t.Fatalf("second Upsert: %v", err)
	}
	if updated.ID != created.ID {
		t.Errorf("second Upsert made a new entry (ID %q != %q) instead of updating the day's entry", updated.ID, created.ID)
	}
	if updated.Content != "<p>second</p>" {
		t.Errorf("Content = %q, want %q", updated.Content, "<p>second</p>")
	}
}

func TestService_UpsertRejectsWhenLocked(t *testing.T) {
	svc := NewService(NewMemoryRepository())

	// Comfortably more than 24h past IST midnight for this date's day-end.
	longAgo := "2020-01-01"
	_, err := svc.Upsert(context.Background(), longAgo, "<p>too late</p>")
	assertKind(t, err, apperr.KindConflict)
}

func TestService_GetByDateReportsLockedForOldEntry(t *testing.T) {
	repo := NewMemoryRepository()
	svc := NewService(repo)

	// Seed directly through the repo (bypassing the service's own lock
	// check) so we can assert Get still reports it as locked.
	if _, err := repo.Upsert(context.Background(), "2020-01-01", "<p>old</p>"); err != nil {
		t.Fatalf("seed Upsert: %v", err)
	}

	e, err := svc.GetByDate(context.Background(), "2020-01-01")
	if err != nil {
		t.Fatalf("GetByDate: %v", err)
	}
	if !e.Locked {
		t.Errorf("Locked = false, want true for an entry from 2020")
	}
}

func TestService_ListDatesRejectsToBeforeFrom(t *testing.T) {
	svc := NewService(NewMemoryRepository())

	_, err := svc.ListDates(context.Background(), "2026-08-20", "2026-08-01")
	assertKind(t, err, apperr.KindInvalidInput)
}

func TestService_ListDatesFiltersToRange(t *testing.T) {
	svc := NewService(NewMemoryRepository())

	for _, d := range []string{"2026-08-01", "2026-08-15", "2026-09-01"} {
		if _, err := svc.repo.Upsert(context.Background(), d, "<p>x</p>"); err != nil {
			t.Fatalf("seed Upsert(%s): %v", d, err)
		}
	}

	dates, err := svc.ListDates(context.Background(), "2026-08-01", "2026-08-31")
	if err != nil {
		t.Fatalf("ListDates: %v", err)
	}
	if len(dates) != 2 || dates[0] != "2026-08-01" || dates[1] != "2026-08-15" {
		t.Fatalf("ListDates = %v, want [2026-08-01 2026-08-15]", dates)
	}
}
