package todo

import (
	"context"
	"errors"
	"testing"

	"github.com/Sam-Frost/portfolio/internal/apperr"
)

func assertNotFound(t *testing.T, err error) {
	t.Helper()
	var appErr *apperr.Error
	if !errors.As(err, &appErr) || appErr.Kind != apperr.KindNotFound {
		t.Fatalf("err = %v, want apperr.NotFound", err)
	}
}

func mustCreate(t *testing.T, repo *MemoryRepository, todo Todo) Todo {
	t.Helper()
	created, err := repo.Create(context.Background(), todo)
	if err != nil {
		t.Fatalf("Create(%+v): %v", todo, err)
	}
	return created
}

func TestMemoryRepository_ListSortsByDateAddedDesc(t *testing.T) {
	repo := NewMemoryRepository()
	first := mustCreate(t, repo, Todo{Name: "first"})
	second := mustCreate(t, repo, Todo{Name: "second"})

	got, err := repo.List(context.Background(), SortByDateAdded, SortDesc, nil)
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	if len(got) != 2 || got[0].ID != second.ID || got[1].ID != first.ID {
		t.Fatalf("List(dateAdded, desc) = %+v, want [second, first]", got)
	}
}

func TestMemoryRepository_ListSortsByTargetDateNullsLast(t *testing.T) {
	repo := NewMemoryRepository()
	late := "2026-09-01"
	early := "2026-08-01"

	withEarly := mustCreate(t, repo, Todo{Name: "early", TargetDate: &early})
	withLate := mustCreate(t, repo, Todo{Name: "late", TargetDate: &late})
	noDate := mustCreate(t, repo, Todo{Name: "no date"})

	asc, err := repo.List(context.Background(), SortByTargetDate, SortAsc, nil)
	if err != nil {
		t.Fatalf("List asc: %v", err)
	}
	assertOrder(t, "asc", asc, withEarly.ID, withLate.ID, noDate.ID)

	desc, err := repo.List(context.Background(), SortByTargetDate, SortDesc, nil)
	if err != nil {
		t.Fatalf("List desc: %v", err)
	}
	// Nulls stay last even when order flips.
	assertOrder(t, "desc", desc, withLate.ID, withEarly.ID, noDate.ID)
}

func TestMemoryRepository_ListSortsByCompletedAtNullsLast(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()
	done := true

	firstDone := mustCreate(t, repo, Todo{Name: "first done"})
	secondDone := mustCreate(t, repo, Todo{Name: "second done"})
	notDone := mustCreate(t, repo, Todo{Name: "not done"})

	// Completion order is firstDone then secondDone, so secondDone carries the
	// later completedAt.
	if _, err := repo.Update(ctx, firstDone.ID, UpdateInput{Done: &done}); err != nil {
		t.Fatalf("Update firstDone: %v", err)
	}
	if _, err := repo.Update(ctx, secondDone.ID, UpdateInput{Done: &done}); err != nil {
		t.Fatalf("Update secondDone: %v", err)
	}

	asc, err := repo.List(ctx, SortByCompletedAt, SortAsc, nil)
	if err != nil {
		t.Fatalf("List asc: %v", err)
	}
	assertOrder(t, "asc", asc, firstDone.ID, secondDone.ID, notDone.ID)

	desc, err := repo.List(ctx, SortByCompletedAt, SortDesc, nil)
	if err != nil {
		t.Fatalf("List desc: %v", err)
	}
	// The never-completed todo stays last even when order flips.
	assertOrder(t, "desc", desc, secondDone.ID, firstDone.ID, notDone.ID)
}

func assertOrder(t *testing.T, label string, got []Todo, wantIDs ...string) {
	t.Helper()
	if len(got) != len(wantIDs) {
		t.Fatalf("%s: got %d todos, want %d", label, len(got), len(wantIDs))
	}
	for i, id := range wantIDs {
		if got[i].ID != id {
			t.Fatalf("%s: position %d = %q, want %q", label, i, got[i].ID, id)
		}
	}
}

func TestMemoryRepository_UpdateAppliesOnlyProvidedFields(t *testing.T) {
	repo := NewMemoryRepository()
	original := mustCreate(t, repo, Todo{Name: "original"})

	done := true
	updated, err := repo.Update(context.Background(), original.ID, UpdateInput{Done: &done})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}

	if updated.Name != "original" {
		t.Errorf("Name = %q, want unchanged %q", updated.Name, "original")
	}
	if !updated.Done {
		t.Errorf("Done = false, want true")
	}
}

func TestMemoryRepository_UpdateStampsAndClearsCompletedAt(t *testing.T) {
	repo := NewMemoryRepository()
	todo := mustCreate(t, repo, Todo{Name: "task"})
	if todo.CompletedAt != nil {
		t.Fatalf("CompletedAt = %v on a fresh todo, want nil", todo.CompletedAt)
	}

	done := true
	completed, err := repo.Update(context.Background(), todo.ID, UpdateInput{Done: &done})
	if err != nil {
		t.Fatalf("Update(done=true): %v", err)
	}
	if completed.CompletedAt == nil {
		t.Fatal("CompletedAt = nil after marking done, want a timestamp")
	}
	first := *completed.CompletedAt

	notDone := false
	undone, err := repo.Update(context.Background(), todo.ID, UpdateInput{Done: &notDone})
	if err != nil {
		t.Fatalf("Update(done=false): %v", err)
	}
	if undone.CompletedAt != nil {
		t.Errorf("CompletedAt = %v after undo, want nil", undone.CompletedAt)
	}

	redone, err := repo.Update(context.Background(), todo.ID, UpdateInput{Done: &done})
	if err != nil {
		t.Fatalf("Update(done=true) again: %v", err)
	}
	if redone.CompletedAt == nil {
		t.Fatal("CompletedAt = nil after redo, want a fresh timestamp")
	}
	if redone.CompletedAt.Before(first) {
		t.Errorf("redo CompletedAt %v is before original %v", *redone.CompletedAt, first)
	}
}

func TestMemoryRepository_UpdateWithEmptyTargetDateClearsIt(t *testing.T) {
	repo := NewMemoryRepository()
	due := "2026-09-01"
	original := mustCreate(t, repo, Todo{Name: "original", TargetDate: &due})

	empty := ""
	updated, err := repo.Update(context.Background(), original.ID, UpdateInput{TargetDate: &empty})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if updated.TargetDate != nil {
		t.Errorf("TargetDate = %v, want nil", updated.TargetDate)
	}
}

func TestMemoryRepository_UpdateUnknownIDReturnsNotFound(t *testing.T) {
	repo := NewMemoryRepository()

	_, err := repo.Update(context.Background(), "missing", UpdateInput{})
	assertNotFound(t, err)
}

func TestMemoryRepository_ListFiltersByLabelID(t *testing.T) {
	repo := NewMemoryRepository()
	work := "work-label"
	other := "other-label"

	withWork := mustCreate(t, repo, Todo{Name: "labeled", LabelID: &work})
	mustCreate(t, repo, Todo{Name: "other label", LabelID: &other})
	mustCreate(t, repo, Todo{Name: "no label"})

	got, err := repo.List(context.Background(), SortByDateAdded, SortDesc, &work)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(got) != 1 || got[0].ID != withWork.ID {
		t.Fatalf("List(labelID=%q) = %+v, want only %+v", work, got, withWork)
	}
}

func TestMemoryRepository_DeleteUnknownIDReturnsNotFound(t *testing.T) {
	repo := NewMemoryRepository()

	err := repo.Delete(context.Background(), "missing")
	assertNotFound(t, err)
}

func TestMemoryRepository_CountActiveExcludesDone(t *testing.T) {
	repo := NewMemoryRepository()
	mustCreate(t, repo, Todo{Name: "active one"})
	mustCreate(t, repo, Todo{Name: "active two"})
	done := mustCreate(t, repo, Todo{Name: "done one"})

	trueVal := true
	if _, err := repo.Update(context.Background(), done.ID, UpdateInput{Done: &trueVal}); err != nil {
		t.Fatalf("Update: %v", err)
	}

	count, err := repo.CountActive(context.Background())
	if err != nil {
		t.Fatalf("CountActive: %v", err)
	}
	if count != 2 {
		t.Errorf("CountActive() = %d, want 2", count)
	}
}
