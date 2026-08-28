package fitness

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

func mustCreateCycle(t *testing.T, repo *MemoryRepository) Cycle {
	t.Helper()
	c, err := repo.CreateCycle(context.Background(), Cycle{Name: "Cut", StartDate: "2026-08-01"})
	if err != nil {
		t.Fatalf("CreateCycle: %v", err)
	}
	return c
}

func TestMemoryRepository_DeleteCycleCascades(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()
	cycle := mustCreateCycle(t, repo)

	ex, err := repo.CreateExercise(ctx, Exercise{CycleID: cycle.ID, Name: "Pushups"})
	if err != nil {
		t.Fatalf("CreateExercise: %v", err)
	}
	if _, err := repo.UpsertExerciseLog(ctx, ex.ID, "2026-08-02", 10); err != nil {
		t.Fatalf("UpsertExerciseLog: %v", err)
	}
	if _, err := repo.UpsertWeightLog(ctx, cycle.ID, "2026-08-02", 80); err != nil {
		t.Fatalf("UpsertWeightLog: %v", err)
	}
	food, err := repo.CreateFood(ctx, Food{Name: "Milk", Unit: "glass", ProteinPerUnit: 8})
	if err != nil {
		t.Fatalf("CreateFood: %v", err)
	}
	if _, err := repo.CreateProteinLog(ctx, ProteinLog{CycleID: cycle.ID, FoodID: food.ID, LogDate: "2026-08-02", Quantity: 1, Protein: 8}); err != nil {
		t.Fatalf("CreateProteinLog: %v", err)
	}

	if err := repo.DeleteCycle(ctx, cycle.ID); err != nil {
		t.Fatalf("DeleteCycle: %v", err)
	}

	if len(repo.exercises) != 0 || len(repo.exerciseLogs) != 0 || len(repo.weightLogs) != 0 ||
		len(repo.proteinLogs) != 0 {
		t.Errorf("cycle children remain after delete: ex=%d exlogs=%d wlogs=%d plogs=%d",
			len(repo.exercises), len(repo.exerciseLogs), len(repo.weightLogs), len(repo.proteinLogs))
	}
	// The shared food library is not owned by the cycle — it stays.
	if len(repo.foods) != 1 {
		t.Errorf("shared foods = %d, want 1 (unaffected by cycle delete)", len(repo.foods))
	}
}

func TestMemoryRepository_UpsertWeightLogReplacesSameDay(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()
	cycle := mustCreateCycle(t, repo)

	if _, err := repo.UpsertWeightLog(ctx, cycle.ID, "2026-08-02", 80.5); err != nil {
		t.Fatalf("first UpsertWeightLog: %v", err)
	}
	if _, err := repo.UpsertWeightLog(ctx, cycle.ID, "2026-08-02", 79.8); err != nil {
		t.Fatalf("second UpsertWeightLog: %v", err)
	}

	logs, err := repo.ListWeightLogs(ctx, cycle.ID)
	if err != nil {
		t.Fatalf("ListWeightLogs: %v", err)
	}
	if len(logs) != 1 || logs[0].Weight != 79.8 {
		t.Fatalf("logs = %+v, want one entry at 79.8", logs)
	}
}

func TestMemoryRepository_ProteinDailyTotalsGroupsByDate(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()
	cycle := mustCreateCycle(t, repo)
	food, _ := repo.CreateFood(ctx, Food{Name: "Milk", Unit: "glass", ProteinPerUnit: 8})

	for _, l := range []ProteinLog{
		{CycleID: cycle.ID, FoodID: food.ID, LogDate: "2026-08-02", Quantity: 1, Protein: 8},
		{CycleID: cycle.ID, FoodID: food.ID, LogDate: "2026-08-02", Quantity: 2, Protein: 16},
		{CycleID: cycle.ID, FoodID: food.ID, LogDate: "2026-08-03", Quantity: 1, Protein: 8},
	} {
		if _, err := repo.CreateProteinLog(ctx, l); err != nil {
			t.Fatalf("CreateProteinLog: %v", err)
		}
	}

	totals, err := repo.ProteinDailyTotals(ctx, cycle.ID)
	if err != nil {
		t.Fatalf("ProteinDailyTotals: %v", err)
	}
	if len(totals) != 2 || totals[0].Date != "2026-08-02" || totals[0].Protein != 24 || totals[1].Protein != 8 {
		t.Fatalf("totals = %+v, want [{08-02 24} {08-03 8}]", totals)
	}
}

func TestMemoryRepository_ActivateCycleArchivesOthersAndComputesStats(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()
	first := mustCreateCycle(t, repo) // active
	second, err := repo.CreateCycle(ctx, Cycle{Name: "Bulk", StartDate: "2026-10-01"})
	if err != nil {
		t.Fatalf("CreateCycle: %v", err)
	} // now second is active, first archived

	if _, err := repo.CreateExercise(ctx, Exercise{CycleID: first.ID, Name: "Squats"}); err != nil {
		t.Fatalf("CreateExercise: %v", err)
	}
	if _, err := repo.UpsertWeightLog(ctx, first.ID, "2026-08-02", 80); err != nil {
		t.Fatalf("UpsertWeightLog: %v", err)
	}
	if _, err := repo.UpsertWeightLog(ctx, first.ID, "2026-08-05", 79.2); err != nil {
		t.Fatalf("UpsertWeightLog: %v", err)
	}

	reactivated, err := repo.ActivateCycle(ctx, first.ID)
	if err != nil {
		t.Fatalf("ActivateCycle: %v", err)
	}
	if reactivated.Status != StatusActive || reactivated.ArchivedAt != nil {
		t.Errorf("reactivated cycle = %q archivedAt=%v, want active with no archivedAt", reactivated.Status, reactivated.ArchivedAt)
	}
	if reactivated.ExerciseCount != 1 {
		t.Errorf("ExerciseCount = %d, want 1", reactivated.ExerciseCount)
	}
	if reactivated.LatestWeight == nil || *reactivated.LatestWeight != 79.2 {
		t.Errorf("LatestWeight = %v, want 79.2 (most recent by date)", reactivated.LatestWeight)
	}

	gotSecond, err := repo.GetCycle(ctx, second.ID)
	if err != nil {
		t.Fatalf("GetCycle(second): %v", err)
	}
	if gotSecond.Status != StatusArchived {
		t.Errorf("second cycle status = %q, want archived", gotSecond.Status)
	}
}

func TestMemoryRepository_DeleteExerciseLogUnknownIDReturnsNotFound(t *testing.T) {
	repo := NewMemoryRepository()
	assertNotFound(t, repo.DeleteExerciseLog(context.Background(), "missing"))
}

func TestMemoryRepository_CreateExerciseUnknownCycleReturnsInvalidInput(t *testing.T) {
	repo := NewMemoryRepository()
	_, err := repo.CreateExercise(context.Background(), Exercise{CycleID: "missing", Name: "x"})
	var appErr *apperr.Error
	if !errors.As(err, &appErr) || appErr.Kind != apperr.KindInvalidInput {
		t.Fatalf("err = %v, want apperr.InvalidInput", err)
	}
}
