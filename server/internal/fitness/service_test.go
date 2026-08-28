package fitness

import (
	"context"
	"errors"
	"testing"

	"github.com/Sam-Frost/portfolio/internal/apperr"
)

func assertInvalidInput(t *testing.T, err error) {
	t.Helper()
	var appErr *apperr.Error
	if !errors.As(err, &appErr) || appErr.Kind != apperr.KindInvalidInput {
		t.Fatalf("err = %v, want apperr.InvalidInput", err)
	}
}

func newSvc() *Service { return NewService(NewMemoryRepository()) }

func mustCycle(t *testing.T, s *Service) Cycle {
	t.Helper()
	c, err := s.CreateCycle(context.Background(), CreateCycleInput{Name: "Cut", StartDate: "2026-08-01"})
	if err != nil {
		t.Fatalf("CreateCycle: %v", err)
	}
	return c
}

func TestCreateCycle_RejectsBlankName(t *testing.T) {
	_, err := newSvc().CreateCycle(context.Background(), CreateCycleInput{Name: "  ", StartDate: "2026-08-01"})
	assertInvalidInput(t, err)
}

func TestCreateCycle_RejectsBadStartDate(t *testing.T) {
	_, err := newSvc().CreateCycle(context.Background(), CreateCycleInput{Name: "Cut", StartDate: "08/01/2026"})
	assertInvalidInput(t, err)
}

func TestCreateCycle_RejectsNonPositiveWeight(t *testing.T) {
	_, err := newSvc().CreateCycle(context.Background(), CreateCycleInput{
		Name: "Cut", StartDate: "2026-08-01", WeightStart: new(0.0),
	})
	assertInvalidInput(t, err)
}

func TestCreateCycle_ArchivesPreviousActiveCycle(t *testing.T) {
	s := newSvc()
	ctx := context.Background()
	first := mustCycle(t, s)

	second, err := s.CreateCycle(ctx, CreateCycleInput{Name: "Bulk", StartDate: "2026-10-01"})
	if err != nil {
		t.Fatalf("second CreateCycle: %v", err)
	}

	got, err := s.GetCycle(ctx, first.ID)
	if err != nil {
		t.Fatalf("GetCycle(first): %v", err)
	}
	if got.Status != StatusArchived || got.ArchivedAt == nil {
		t.Errorf("first cycle status = %q archivedAt=%v, want archived with a timestamp", got.Status, got.ArchivedAt)
	}

	active, ok, err := s.ActiveCycle(ctx)
	if err != nil || !ok {
		t.Fatalf("ActiveCycle: ok=%v err=%v", ok, err)
	}
	if active.ID != second.ID {
		t.Errorf("active cycle = %q, want %q", active.ID, second.ID)
	}
}

func TestCreateExercise_RejectsNonPositiveGoalQuantity(t *testing.T) {
	s := newSvc()
	c := mustCycle(t, s)
	_, err := s.CreateExercise(context.Background(), c.ID, CreateExerciseInput{Name: "Pushups", GoalQuantity: new(-5.0)})
	assertInvalidInput(t, err)
}

func TestUpsertExerciseLog_ReplacesSameDayEntry(t *testing.T) {
	s := newSvc()
	ctx := context.Background()
	c := mustCycle(t, s)
	ex, err := s.CreateExercise(ctx, c.ID, CreateExerciseInput{Name: "Pushups"})
	if err != nil {
		t.Fatalf("CreateExercise: %v", err)
	}

	if _, err := s.UpsertExerciseLog(ctx, ex.ID, UpsertExerciseLogInput{Date: "2026-08-10", Quantity: 20}); err != nil {
		t.Fatalf("first upsert: %v", err)
	}
	if _, err := s.UpsertExerciseLog(ctx, ex.ID, UpsertExerciseLogInput{Date: "2026-08-10", Quantity: 35}); err != nil {
		t.Fatalf("second upsert: %v", err)
	}

	logs, err := s.ListExerciseLogs(ctx, ex.ID)
	if err != nil {
		t.Fatalf("ListExerciseLogs: %v", err)
	}
	if len(logs) != 1 || logs[0].Quantity != 35 {
		t.Fatalf("logs = %+v, want one entry of quantity 35", logs)
	}

	got, err := s.GetExercise(ctx, ex.ID)
	if err != nil {
		t.Fatalf("GetExercise: %v", err)
	}
	if got.TotalLogged != 35 {
		t.Errorf("TotalLogged = %v, want 35", got.TotalLogged)
	}
}

func TestCreateProteinLog_SnapshotsProteinFromFood(t *testing.T) {
	s := newSvc()
	ctx := context.Background()
	c := mustCycle(t, s)

	food, err := s.CreateFood(ctx, CreateFoodInput{Name: "Milk", Unit: "glass", ProteinPerUnit: 8})
	if err != nil {
		t.Fatalf("CreateFood: %v", err)
	}

	log, err := s.CreateProteinLog(ctx, c.ID, CreateProteinLogInput{FoodID: food.ID, Date: "2026-08-10", Quantity: 2})
	if err != nil {
		t.Fatalf("CreateProteinLog: %v", err)
	}
	if log.Protein != 16 {
		t.Errorf("Protein = %v, want 16 (2 glasses * 8g)", log.Protein)
	}

	// Editing the food's protein content must not rewrite the existing log.
	if _, err := s.UpdateFood(ctx, food.ID, UpdateFoodInput{ProteinPerUnit: new(10.0)}); err != nil {
		t.Fatalf("UpdateFood: %v", err)
	}
	totals, err := s.ProteinDailyTotals(ctx, c.ID)
	if err != nil {
		t.Fatalf("ProteinDailyTotals: %v", err)
	}
	if len(totals) != 1 || totals[0].Protein != 16 {
		t.Fatalf("totals = %+v, want a single day at 16g", totals)
	}
}

func TestCreateProteinLog_AcceptsSharedFoodAcrossCycles(t *testing.T) {
	s := newSvc()
	ctx := context.Background()
	mustCycle(t, s)
	food, err := s.CreateFood(ctx, CreateFoodInput{Name: "Milk", Unit: "glass", ProteinPerUnit: 8})
	if err != nil {
		t.Fatalf("CreateFood: %v", err)
	}

	// The food library is shared, so a food added while one cycle was
	// active is usable from a later cycle without re-creating it.
	second, err := s.CreateCycle(ctx, CreateCycleInput{Name: "Bulk", StartDate: "2026-10-01"})
	if err != nil {
		t.Fatalf("second CreateCycle: %v", err)
	}

	if _, err := s.CreateProteinLog(ctx, second.ID, CreateProteinLogInput{FoodID: food.ID, Date: "2026-10-02", Quantity: 1}); err != nil {
		t.Fatalf("CreateProteinLog with shared food: %v", err)
	}
}
