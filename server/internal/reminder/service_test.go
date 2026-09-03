package reminder

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Sam-Frost/portfolio/internal/apperr"
)

func mustInvalid(t *testing.T, err error) {
	t.Helper()
	var appErr *apperr.Error
	if !errors.As(err, &appErr) || appErr.Kind != apperr.KindInvalidInput {
		t.Fatalf("err = %v, want apperr.InvalidInput", err)
	}
}

func TestCreate_OnceRequiresFutureFireAt(t *testing.T) {
	svc := NewService(NewMemoryRepository())
	ctx := context.Background()

	_, err := svc.Create(ctx, "todo1", CreateInput{Kind: KindOnce})
	mustInvalid(t, err)

	past := time.Now().Add(-time.Hour).UTC().Format(time.RFC3339)
	_, err = svc.Create(ctx, "todo1", CreateInput{Kind: KindOnce, FireAt: &past})
	mustInvalid(t, err)

	future := time.Now().Add(time.Hour).UTC().Format(time.RFC3339)
	rem, err := svc.Create(ctx, "todo1", CreateInput{Kind: KindOnce, FireAt: &future})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if rem.Kind != KindOnce || rem.IntervalSeconds != nil {
		t.Fatalf("unexpected reminder: %+v", rem)
	}
}

func TestCreate_RepeatNeedsMinInterval(t *testing.T) {
	svc := NewService(NewMemoryRepository())
	ctx := context.Background()

	short := 30
	_, err := svc.Create(ctx, "todo1", CreateInput{Kind: KindRepeat, IntervalSeconds: &short})
	mustInvalid(t, err)

	ok := 900
	before := time.Now().UTC()
	rem, err := svc.Create(ctx, "todo1", CreateInput{Kind: KindRepeat, IntervalSeconds: &ok})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if rem.FireAt.Before(before.Add(890*time.Second)) || rem.FireAt.After(time.Now().Add(910*time.Second)) {
		t.Fatalf("FireAt not ~now+interval: %v", rem.FireAt)
	}
}

func TestCreate_RejectsUnknownKind(t *testing.T) {
	svc := NewService(NewMemoryRepository())
	_, err := svc.Create(context.Background(), "todo1", CreateInput{Kind: "weekly"})
	mustInvalid(t, err)
}

func TestSettle_RepeatCatchesUpPastMissedTicks(t *testing.T) {
	repo := NewMemoryRepository()
	svc := NewService(repo)
	ctx := context.Background()

	interval := 600 // 10 min
	d := Due{Reminder: Reminder{
		ID:              "r1",
		TodoID:          "t1",
		Kind:            KindRepeat,
		IntervalSeconds: &interval,
		FireAt:          time.Date(2026, 9, 3, 12, 0, 0, 0, time.UTC),
	}}
	repo.reminders["r1"] = d.Reminder

	// The scheduler was down for 35 minutes; next fire should land after now,
	// not queue up 3 back-to-back fires.
	now := time.Date(2026, 9, 3, 12, 35, 0, 0, time.UTC)
	if err := svc.Settle(ctx, d, now); err != nil {
		t.Fatalf("Settle: %v", err)
	}
	got := repo.reminders["r1"].FireAt
	want := time.Date(2026, 9, 3, 12, 40, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Fatalf("FireAt = %v, want %v", got, want)
	}
}

func TestSettle_OnceIsDeleted(t *testing.T) {
	repo := NewMemoryRepository()
	svc := NewService(repo)

	d := Due{Reminder: Reminder{ID: "r1", TodoID: "t1", Kind: KindOnce, FireAt: time.Now()}}
	repo.reminders["r1"] = d.Reminder

	if err := svc.Settle(context.Background(), d, time.Now()); err != nil {
		t.Fatalf("Settle: %v", err)
	}
	if _, ok := repo.reminders["r1"]; ok {
		t.Fatal("one-time reminder should have been deleted after firing")
	}
}

func TestDueBefore_SkipsDoneTodos(t *testing.T) {
	repo := NewMemoryRepository()
	svc := NewService(repo)
	ctx := context.Background()

	repo.SetTodo("active", "Ship it", false)
	repo.SetTodo("finished", "Done deal", true)

	past := time.Now().Add(-time.Hour)
	repo.reminders["r-active"] = Reminder{ID: "r-active", TodoID: "active", Kind: KindOnce, FireAt: past}
	repo.reminders["r-finished"] = Reminder{ID: "r-finished", TodoID: "finished", Kind: KindOnce, FireAt: past}

	due, err := svc.DueBefore(ctx, time.Now())
	if err != nil {
		t.Fatalf("DueBefore: %v", err)
	}
	if len(due) != 1 || due[0].TodoID != "active" {
		t.Fatalf("due = %+v, want just the active todo's reminder", due)
	}
}
